package health

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/glodb/keel/database/baseconnections"

	"github.com/glodb/keel/settings/cachesettings/cache"
	"github.com/glodb/keel/settings/circuitbreaker"
	"github.com/glodb/keel/settings/errors"
	"github.com/glodb/keel/settings/logger"
	"github.com/glodb/keel/settings/metrics"
)

// Status represents the health status
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusUnhealthy Status = "unhealthy"
	StatusUnknown   Status = "unknown"
)

// CheckType represents different types of health checks
type CheckType string

const (
	CheckTypeReadiness CheckType = "readiness"
	CheckTypeLiveness  CheckType = "liveness"
	CheckTypeStartup   CheckType = "startup"
)

// Check represents a single health check
type Check struct {
	Name        string                                    `json:"name"`
	Type        CheckType                                 `json:"type"`
	Status      Status                                    `json:"status"`
	Message     string                                    `json:"message,omitempty"`
	Error       string                                    `json:"error,omitempty"`
	Duration    time.Duration                             `json:"duration"`
	Timestamp   time.Time                                 `json:"timestamp"`
	Metadata    map[string]interface{}                    `json:"metadata,omitempty"`
	CheckFunc   func(ctx context.Context) (Status, error) `json:"-"`
	Timeout     time.Duration                             `json:"timeout"`
	Critical    bool                                      `json:"critical"`
	Interval    time.Duration                             `json:"interval"`
	LastChecked time.Time                                 `json:"last_checked"`
}

// HealthReport represents the overall health status of the service
type HealthReport struct {
	Service         string                    `json:"service"`
	Version         string                    `json:"version"`
	Status          Status                    `json:"status"`
	Timestamp       time.Time                 `json:"timestamp"`
	Uptime          time.Duration             `json:"uptime"`
	Checks          map[string]Check          `json:"checks"`
	System          SystemInfo                `json:"system"`
	Dependencies    map[string]DependencyInfo `json:"dependencies"`
	CircuitBreakers map[string]bool           `json:"circuit_breakers"`
}

// SystemInfo provides system-level information
type SystemInfo struct {
	Goroutines    int     `json:"goroutines"`
	MemoryUsedMB  float64 `json:"memory_used_mb"`
	MemoryTotalMB float64 `json:"memory_total_mb"`
	CPUCores      int     `json:"cpu_cores"`
	GCPauseMs     float64 `json:"gc_pause_ms"`
	GCCycles      uint32  `json:"gc_cycles"`
}

// DependencyInfo provides information about external dependencies
type DependencyInfo struct {
	Name      string        `json:"name"`
	Type      string        `json:"type"`
	Status    Status        `json:"status"`
	Latency   time.Duration `json:"latency"`
	LastCheck time.Time     `json:"last_check"`
	Error     string        `json:"error,omitempty"`
}

// HealthChecker manages health checks for the service
type HealthChecker struct {
	serviceName    string
	serviceVersion string
	startTime      time.Time
	checks         map[string]*Check
	mutex          sync.RWMutex
	logger         *logger.Logger
	metrics        *metrics.Metrics
	ticker         *time.Ticker
	stopChan       chan struct{}
	running        bool
}

var (
	defaultChecker *HealthChecker
	checkerOnce    sync.Once
)

// GetDefaultChecker returns the default health checker instance
func GetDefaultChecker() *HealthChecker {
	checkerOnce.Do(func() {
		defaultChecker = NewHealthChecker("keel-backend", "1.0.0")
	})
	return defaultChecker
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(serviceName, serviceVersion string) *HealthChecker {
	hc := &HealthChecker{
		serviceName:    serviceName,
		serviceVersion: serviceVersion,
		startTime:      time.Now(),
		checks:         make(map[string]*Check),
		logger:         logger.Log(),
		metrics:        metrics.GetInstance(),
		stopChan:       make(chan struct{}),
	}

	// Register default system checks
	hc.registerDefaultChecks()

	return hc
}

// registerDefaultChecks registers the default health checks
func (hc *HealthChecker) registerDefaultChecks() {
	// Memory usage check
	hc.RegisterCheck("memory", CheckTypeLiveness, func(ctx context.Context) (Status, error) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		memoryUsedMB := float64(m.Alloc) / 1024 / 1024
		memoryLimit := 1024.0 // 1GB limit, make configurable

		if memoryUsedMB > memoryLimit {
			return StatusUnhealthy, fmt.Errorf("memory usage %.2fMB exceeds limit %.2fMB", memoryUsedMB, memoryLimit)
		} else if memoryUsedMB > memoryLimit*0.8 {
			return StatusDegraded, fmt.Errorf("memory usage %.2fMB is approaching limit %.2fMB", memoryUsedMB, memoryLimit)
		}

		return StatusHealthy, nil
	}, 30*time.Second, true)

	// Goroutine count check
	hc.RegisterCheck("goroutines", CheckTypeLiveness, func(ctx context.Context) (Status, error) {
		goroutineCount := runtime.NumGoroutine()
		goroutineLimit := 10000 // Make configurable

		if goroutineCount > goroutineLimit {
			return StatusUnhealthy, fmt.Errorf("goroutine count %d exceeds limit %d", goroutineCount, goroutineLimit)
		} else if goroutineCount > goroutineLimit/2 {
			return StatusDegraded, fmt.Errorf("goroutine count %d is high", goroutineCount)
		}

		return StatusHealthy, nil
	}, 30*time.Second, false)

	// Database connectivity checks
	hc.RegisterCheck("database", CheckTypeReadiness, func(ctx context.Context) (Status, error) {
		pool := baseconnections.GetConnectionPool()
		connections := pool.GetAllConnections()

		if len(connections) == 0 {
			return StatusUnhealthy, fmt.Errorf("no database connections available")
		}

		healthyCount := 0
		var lastError error

		for dbType, _ := range connections {
			healthy, err := pool.IsConnectionHealthy(ctx, dbType)
			if healthy && err == nil {
				healthyCount++
			} else {
				lastError = err
			}
		}

		if healthyCount == 0 {
			return StatusUnhealthy, fmt.Errorf("all database connections are unhealthy: %v", lastError)
		} else if healthyCount < len(connections) {
			return StatusDegraded, fmt.Errorf("some database connections are unhealthy: %d/%d healthy", healthyCount, len(connections))
		}

		return StatusHealthy, nil
	}, 15*time.Second, true)

	// Cache connectivity check
	hc.RegisterCheck("cache", CheckTypeReadiness, func(ctx context.Context) (Status, error) {
		cacheInstance := cache.GetCache()

		// Test cache connectivity with a simple ping
		testKey := "health_check_ping"
		testValue := time.Now().Unix()

		// Try to set a value
		if err := cacheInstance.SetInt(ctx, testKey, int(testValue)); err != nil {
			return StatusUnhealthy, fmt.Errorf("cache write failed: %v", err)
		}

		// Try to get the value back
		retrievedValue, err := cacheInstance.GetInt(ctx, testKey)
		if err != nil {
			return StatusUnhealthy, fmt.Errorf("cache read failed: %v", err)
		}

		if retrievedValue != testValue {
			return StatusDegraded, fmt.Errorf("cache value mismatch: expected %d, got %d", testValue, retrievedValue)
		}

		// Clean up
		_ = cacheInstance.Del(ctx, testKey)

		return StatusHealthy, nil
	}, 30*time.Second, true)

	// Circuit breaker status check
	hc.RegisterCheck("circuit_breakers", CheckTypeLiveness, func(ctx context.Context) (Status, error) {
		manager := circuitbreaker.GetManager()
		healthStatus := manager.GetHealthStatus()

		if len(healthStatus) == 0 {
			return StatusHealthy, nil // No circuit breakers configured
		}

		unhealthyCount := 0
		for name, healthy := range healthStatus {
			if !healthy {
				unhealthyCount++
				hc.logger.Warn("Circuit breaker is open",
					logger.StringField("circuit_breaker", name))
			}
		}

		if unhealthyCount == len(healthStatus) {
			return StatusUnhealthy, fmt.Errorf("all circuit breakers are open")
		} else if unhealthyCount > 0 {
			return StatusDegraded, fmt.Errorf("%d/%d circuit breakers are open", unhealthyCount, len(healthStatus))
		}

		return StatusHealthy, nil
	}, 20*time.Second, false)
}

// RegisterCheck registers a new health check
func (hc *HealthChecker) RegisterCheck(name string, checkType CheckType, checkFunc func(ctx context.Context) (Status, error), timeout time.Duration, critical bool) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	check := &Check{
		Name:      name,
		Type:      checkType,
		CheckFunc: checkFunc,
		Timeout:   timeout,
		Critical:  critical,
		Interval:  30 * time.Second, // Default interval
		Metadata:  make(map[string]interface{}),
	}

	hc.checks[name] = check

	hc.logger.Info("Health check registered",
		logger.StringField("check_name", name),
		logger.StringField("check_type", string(checkType)),
		logger.BoolField("critical", critical))
}

// RemoveCheck removes a health check
func (hc *HealthChecker) RemoveCheck(name string) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()

	delete(hc.checks, name)

	hc.logger.Info("Health check removed",
		logger.StringField("check_name", name))
}

// RunCheck executes a specific health check
func (hc *HealthChecker) RunCheck(ctx context.Context, name string) (*Check, error) {
	hc.mutex.RLock()
	check, exists := hc.checks[name]
	hc.mutex.RUnlock()

	if !exists {
		return nil, errors.NewAppError(errors.ErrCodeRecordNotFound, fmt.Sprintf("health check '%s' not found", name), nil)
	}

	return hc.executeCheck(ctx, check)
}

// executeCheck executes a single health check
func (hc *HealthChecker) executeCheck(ctx context.Context, check *Check) (*Check, error) {
	start := time.Now()

	// Create timeout context
	checkCtx, cancel := context.WithTimeout(ctx, check.Timeout)
	defer cancel()

	// Execute the check
	status, err := check.CheckFunc(checkCtx)
	duration := time.Since(start)

	// Update check result
	result := *check // Create a copy
	result.Status = status
	result.Duration = duration
	result.Timestamp = time.Now()
	result.LastChecked = time.Now()

	if err != nil {
		result.Error = err.Error()
		result.Message = err.Error()
	} else {
		result.Error = ""
		result.Message = "Check passed"
	}

	// Add performance metadata
	result.Metadata["duration_ms"] = float64(duration.Nanoseconds()) / 1000000.0
	result.Metadata["timeout_ms"] = float64(check.Timeout.Nanoseconds()) / 1000000.0

	// Record metrics
	statusStr := string(status)
	hc.metrics.RecordAPICall("health_check", "execute", statusStr, hc.serviceName, duration)

	if status == StatusUnhealthy || status == StatusDegraded {
		hc.metrics.RecordError("health_check_failed", check.Name, "health_checker")
	}

	hc.logger.Debug("Health check executed",
		logger.StringField("check_name", check.Name),
		logger.StringField("status", statusStr),
		logger.DurationField("duration", duration),
		logger.ErrorField("error", err))

	return &result, nil
}

// GetHealthReport generates a comprehensive health report
func (hc *HealthChecker) GetHealthReport(ctx context.Context) (*HealthReport, error) {
	hc.mutex.RLock()
	checks := make(map[string]*Check, len(hc.checks))
	for name, check := range hc.checks {
		checks[name] = check
	}
	hc.mutex.RUnlock()

	// Execute all health checks concurrently
	checkResults := make(map[string]Check)
	var wg sync.WaitGroup
	var resultMutex sync.Mutex

	for name, check := range checks {
		wg.Add(1)
		go func(checkName string, checkInstance *Check) {
			defer wg.Done()

			result, err := hc.executeCheck(ctx, checkInstance)
			if err != nil {
				// Create a failed check result
				result = &Check{
					Name:      checkName,
					Type:      checkInstance.Type,
					Status:    StatusUnhealthy,
					Error:     err.Error(),
					Message:   err.Error(),
					Duration:  0,
					Timestamp: time.Now(),
					Critical:  checkInstance.Critical,
				}
			}

			resultMutex.Lock()
			checkResults[checkName] = *result
			resultMutex.Unlock()
		}(name, check)
	}

	wg.Wait()

	// Determine overall status
	overallStatus := hc.calculateOverallStatus(checkResults)

	// Collect system information
	systemInfo := hc.getSystemInfo()

	// Collect dependency information
	dependencies := hc.getDependencyInfo(ctx)

	// Collect circuit breaker status
	circuitBreakers := circuitbreaker.GetManager().GetHealthStatus()

	report := &HealthReport{
		Service:         hc.serviceName,
		Version:         hc.serviceVersion,
		Status:          overallStatus,
		Timestamp:       time.Now(),
		Uptime:          time.Since(hc.startTime),
		Checks:          checkResults,
		System:          systemInfo,
		Dependencies:    dependencies,
		CircuitBreakers: circuitBreakers,
	}

	return report, nil
}

// calculateOverallStatus determines the overall health status based on individual checks
func (hc *HealthChecker) calculateOverallStatus(checks map[string]Check) Status {
	if len(checks) == 0 {
		return StatusUnknown
	}

	hasUnhealthy := false
	hasDegraded := false

	for _, check := range checks {
		switch check.Status {
		case StatusUnhealthy:
			if check.Critical {
				return StatusUnhealthy // Critical check failed, service is unhealthy
			}
			hasUnhealthy = true
		case StatusDegraded:
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		return StatusDegraded // Non-critical checks failed, service is degraded
	} else if hasDegraded {
		return StatusDegraded
	}

	return StatusHealthy
}

// getSystemInfo collects system-level information
func (hc *HealthChecker) getSystemInfo() SystemInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return SystemInfo{
		Goroutines:    runtime.NumGoroutine(),
		MemoryUsedMB:  float64(m.Alloc) / 1024 / 1024,
		MemoryTotalMB: float64(m.Sys) / 1024 / 1024,
		CPUCores:      runtime.NumCPU(),
		GCPauseMs:     float64(m.PauseNs[(m.NumGC+255)%256]) / 1000000.0,
		GCCycles:      m.NumGC,
	}
}

// getDependencyInfo collects information about external dependencies
func (hc *HealthChecker) getDependencyInfo(ctx context.Context) map[string]DependencyInfo {
	dependencies := make(map[string]DependencyInfo)

	// Database dependencies
	pool := baseconnections.GetConnectionPool()
	connections := pool.GetAllConnections()

	for dbType, _ := range connections {
		start := time.Now()
		healthy, err := pool.IsConnectionHealthy(ctx, dbType)
		latency := time.Since(start)

		status := StatusHealthy
		errorMsg := ""
		if !healthy || err != nil {
			status = StatusUnhealthy
			if err != nil {
				errorMsg = err.Error()
			}
		}

		dependencies[fmt.Sprintf("database_%s", dbType.String())] = DependencyInfo{
			Name:      fmt.Sprintf("Database (%s)", dbType.String()),
			Type:      "database",
			Status:    status,
			Latency:   latency,
			LastCheck: time.Now(),
			Error:     errorMsg,
		}
	}

	// Cache dependency
	start := time.Now()
	cacheInstance := cache.GetCache()
	testKey := "health_dependency_check"
	cacheErr := cacheInstance.SetString(ctx, testKey, "test")
	cacheLatency := time.Since(start)

	cacheStatus := StatusHealthy
	cacheError := ""
	if cacheErr != nil {
		cacheStatus = StatusUnhealthy
		cacheError = cacheErr.Error()
	} else {
		_ = cacheInstance.Del(ctx, testKey) // Clean up
	}

	dependencies["cache"] = DependencyInfo{
		Name:      "Redis Cache",
		Type:      "cache",
		Status:    cacheStatus,
		Latency:   cacheLatency,
		LastCheck: time.Now(),
		Error:     cacheError,
	}

	return dependencies
}

// StartPeriodicChecks starts periodic health checks
func (hc *HealthChecker) StartPeriodicChecks(interval time.Duration) {
	if hc.running {
		return
	}

	hc.running = true
	hc.ticker = time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-hc.ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				report, err := hc.GetHealthReport(ctx)
				cancel()

				if err != nil {
					hc.logger.Error("Failed to generate health report",
						logger.ErrorField("error", err))
				} else {
					hc.logger.Debug("Health check completed",
						logger.StringField("status", string(report.Status)),
						logger.IntField("checks_count", len(report.Checks)))

					// Record overall health status
					hc.metrics.RecordAPICall("health_status", "periodic", string(report.Status), hc.serviceName, 0)
				}

			case <-hc.stopChan:
				return
			}
		}
	}()

	hc.logger.Info("Periodic health checks started",
		logger.DurationField("interval", interval))
}

// StopPeriodicChecks stops periodic health checks
func (hc *HealthChecker) StopPeriodicChecks() {
	if !hc.running {
		return
	}

	hc.running = false
	if hc.ticker != nil {
		hc.ticker.Stop()
	}
	close(hc.stopChan)

	hc.logger.Info("Periodic health checks stopped")
}

// IsHealthy returns true if the service is healthy
func (hc *HealthChecker) IsHealthy(ctx context.Context) bool {
	report, err := hc.GetHealthReport(ctx)
	if err != nil {
		return false
	}
	return report.Status == StatusHealthy
}

// IsReady returns true if the service is ready to serve requests
func (hc *HealthChecker) IsReady(ctx context.Context) bool {
	report, err := hc.GetHealthReport(ctx)
	if err != nil {
		return false
	}
	return report.Status == StatusHealthy || report.Status == StatusDegraded
}

// GetCheckNames returns all registered check names
func (hc *HealthChecker) GetCheckNames() []string {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()

	names := make([]string, 0, len(hc.checks))
	for name := range hc.checks {
		names = append(names, name)
	}
	return names
}
