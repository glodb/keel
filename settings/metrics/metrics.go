package metrics

import (
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Metrics provides Prometheus metrics collection
type Metrics struct {
	// HTTP Metrics
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	httpRequestsInFlight *prometheus.GaugeVec

	// Database Metrics
	dbOperationsTotal   *prometheus.CounterVec
	dbOperationDuration *prometheus.HistogramVec
	dbOperationsFailed  *prometheus.CounterVec
	dbConnectionsActive *prometheus.GaugeVec

	// Business Metrics
	userRegistrationsTotal *prometheus.CounterVec
	userLoginsTotal        *prometheus.CounterVec
	userLoginsFailed       *prometheus.CounterVec
	activeSessions         *prometheus.GaugeVec

	// System Metrics
	goroutinesCount  *prometheus.GaugeVec
	memoryUsageBytes *prometheus.GaugeVec
	errorsTotal      *prometheus.CounterVec

	// API Metrics
	apiValidCallsTotal   *prometheus.CounterVec
	apiInvalidCallsTotal *prometheus.CounterVec
	apiValidationErrors  *prometheus.CounterVec
	apiCallsTotal        *prometheus.CounterVec
	apiCallDuration      *prometheus.HistogramVec

	// Rate Limiting Metrics
	rateLimitHits     *prometheus.CounterVec
	rateLimitExceeded *prometheus.CounterVec

	// Cache Metrics
	cacheHits   *prometheus.CounterVec
	cacheMisses *prometheus.CounterVec
	cacheSize   *prometheus.GaugeVec

	// Session Metrics
	sessionCreatedTotal   *prometheus.CounterVec
	sessionPerPlatform    *prometheus.GaugeVec
	sessionPerOsType      *prometheus.GaugeVec
	sessionPerOsVersion   *prometheus.GaugeVec
	sessionPerDeviceModel *prometheus.GaugeVec
	sessionPerDeviceId    *prometheus.GaugeVec
	sessionPerLanguage    *prometheus.GaugeVec
	sessionPerCountry     *prometheus.GaugeVec
	sessionPerCity        *prometheus.GaugeVec

	// Session Access Metrics
	sessionAccessTotal     *prometheus.CounterVec
	sessionAccessPerSecond *prometheus.GaugeVec

	// Version Metrics
	clientVersionAccess    *prometheus.CounterVec
	clientVersionPerSecond *prometheus.GaugeVec

	// Authentication Metrics
	authenticatedRequests   *prometheus.CounterVec
	unauthenticatedRequests *prometheus.CounterVec
	userAccessPerMinute     *prometheus.CounterVec
	userAccessPerHour       *prometheus.CounterVec
	userAccessPerDay        *prometheus.CounterVec

	registry *prometheus.Registry
	logger   *zap.Logger
	mu       sync.RWMutex
}

var (
	instance *Metrics
	once     sync.Once
)

// GetInstance returns the singleton metrics instance
func GetInstance() *Metrics {
	once.Do(func() {
		instance = newMetrics()
	})
	return instance
}

// newMetrics creates a new metrics instance
func newMetrics() *Metrics {
	m := &Metrics{
		registry: prometheus.NewRegistry(),
		logger:   zap.L(),
	}

	m.initializeMetrics()
	m.registerMetrics()

	return m
}

// initializeMetrics initializes all Prometheus metrics
func (m *Metrics) initializeMetrics() {
	// HTTP Metrics
	m.httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status", "service"},
	)

	m.httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status", "service"},
	)

	m.httpRequestsInFlight = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being processed",
		},
		[]string{"service"},
	)

	// Database Metrics
	m.dbOperationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_operations_total",
			Help: "Total number of database operations",
		},
		[]string{"operation", "collection", "status", "service"},
	)

	m.dbOperationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_operation_duration_seconds",
			Help:    "Database operation duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "collection", "service"},
	)

	m.dbOperationsFailed = prometheus.NewCounterVec(

		prometheus.CounterOpts{
			Name: "db_operations_failed",
			Help: "Total number of failed database operations",
		},
		[]string{"operation", "collection", "service"},
	)

	m.dbConnectionsActive = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "db_connections_active",
			Help: "Number of active database connections",
		},
		[]string{"database", "service"},
	)

	// Business Metrics
	m.userRegistrationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_registrations_total",
			Help: "Total number of user registration attempts",
		},
		[]string{"status", "service"},
	)

	m.userLoginsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_logins_total",
			Help: "Total number of user login attempts",
		},
		[]string{"status", "service"},
	)

	m.userLoginsFailed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_logins_failed",
			Help: "Total number of failed user logins",
		},
		[]string{"reason", "service"},
	)

	m.activeSessions = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "active_sessions",
			Help: "Number of active user sessions",
		},
		[]string{"service"},
	)

	// System Metrics
	m.goroutinesCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "goroutines_count",
			Help: "Number of goroutines",
		},
		[]string{"service"},
	)

	m.memoryUsageBytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "memory_usage_bytes",
			Help: "Memory usage in bytes",
		},
		[]string{"type", "service"},
	)

	m.errorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "errors_total",
			Help: "Total number of errors",
		},
		[]string{"type", "component", "service"},
	)

	// API Metrics
	m.apiValidCallsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_valid_calls_total",
			Help: "Total number of valid API calls",
		},
		[]string{"endpoint", "method", "service"},
	)

	m.apiInvalidCallsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_invalid_calls_total",
			Help: "Total number of invalid API calls",
		},
		[]string{"endpoint", "method", "service"},
	)

	m.apiValidationErrors = prometheus.NewCounterVec(

		prometheus.CounterOpts{
			Name: "api_validation_errors",
			Help: "Total number of validation errors",
		},
		[]string{"endpoint", "method", "errors", "service"},
	)

	m.apiCallsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_calls_total",
			Help: "Total number of API calls",
		},
		[]string{"endpoint", "method", "status", "service"},
	)

	m.apiCallDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_call_duration_seconds",
			Help:    "API call duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint", "method", "service"},
	)

	// Rate Limiting Metrics
	m.rateLimitHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_hits",
			Help: "Number of rate limit hits",
		},
		[]string{"client_id", "path", "service"},
	)

	m.rateLimitExceeded = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_exceeded",
			Help: "Number of rate limit violations",
		},
		[]string{"client_id", "path", "service"},
	)

	// Cache Metrics
	m.cacheHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_hits",
			Help: "Number of cache hits",
		},
		[]string{"cache_operation", "service"},
	)

	m.cacheMisses = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_misses",
			Help: "Number of cache misses",
		},
		[]string{"cache_operation", "service"},
	)

	m.cacheSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cache_size",
			Help: "Current cache size",
		},
		[]string{"cache_operation", "service"},
	)

	// Session Metrics
	m.sessionCreatedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "session_created_total",
			Help: "Total number of session creations",
		},
		[]string{"service"},
	)

	m.sessionPerPlatform = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "session_per_platform",
			Help: "Number of sessions per platform",
		},
		[]string{"platform", "service"},
	)

	m.sessionPerOsType = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "session_per_os_type",
			Help: "Number of sessions per os type",
		},
		[]string{"os_type", "service"},
	)

	m.sessionPerOsVersion = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "session_per_os_version",
			Help: "Number of sessions per os version",
		},
		[]string{"os_version", "service"},
	)

	m.sessionPerDeviceModel = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "session_per_device_model",
			Help: "Number of sessions per device model",
		},
		[]string{"device_model", "service"},
	)

	m.sessionPerDeviceId = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "session_per_device_id",
			Help: "Number of sessions per device id",
		},
		[]string{"device_id", "service"},
	)

	m.sessionPerLanguage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "session_per_language",
			Help: "Number of sessions per language",
		},
		[]string{"language", "service"},
	)

	m.sessionPerCountry = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "session_per_country",
			Help: "Number of sessions per country",
		},
		[]string{"country", "service"},
	)

	m.sessionPerCity = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "session_per_city",
			Help: "Number of sessions per city",
		},
		[]string{"city", "service"},
	)

	// Session Access Metrics
	m.sessionAccessTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "session_access_total",
			Help: "Total number of session accesses",
		},
		[]string{"service"},
	)

	m.sessionAccessPerSecond = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "session_access_per_second",
			Help: "Number of session accesses per second",
		},
		[]string{"service"},
	)

	// Version Metrics
	m.clientVersionAccess = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "client_version_access_total",
			Help: "Total number of client version accesses",
		},
		[]string{"version", "service"},
	)

	m.clientVersionPerSecond = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "client_version_per_second",
			Help: "Number of client version accesses per second",
		},
		[]string{"version", "service"},
	)

	// Authentication Metrics
	m.authenticatedRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "authenticated_requests_total",
			Help: "Total number of authenticated requests",
		},
		[]string{"service"},
	)

	m.unauthenticatedRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "unauthenticated_requests_total",
			Help: "Total number of unauthenticated requests",
		},
		[]string{"service"},
	)

	m.userAccessPerMinute = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_access_per_minute",
			Help: "Number of user accesses per minute",
		},
		[]string{"service"},
	)

	m.userAccessPerHour = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_access_per_hour",
			Help: "Number of user accesses per hour",
		},
		[]string{"service"},
	)

	m.userAccessPerDay = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_access_per_day",
			Help: "Number of user accesses per day",
		},
		[]string{"service"},
	)

}

// registerMetrics registers all metrics with the Prometheus registry
func (m *Metrics) registerMetrics() {
	metrics := []prometheus.Collector{
		m.httpRequestsTotal,
		m.httpRequestDuration,
		m.httpRequestsInFlight,
		m.dbOperationsTotal,
		m.dbOperationDuration,
		m.dbOperationsFailed,
		m.dbConnectionsActive,
		m.userRegistrationsTotal,
		m.userLoginsTotal,
		m.userLoginsFailed,
		m.activeSessions,
		m.goroutinesCount,
		m.memoryUsageBytes,
		m.errorsTotal,
		m.apiValidCallsTotal,
		m.apiInvalidCallsTotal,
		m.apiValidationErrors,
		m.apiCallsTotal,
		m.apiCallDuration,
		m.rateLimitHits,
		m.rateLimitExceeded,
		m.cacheHits,
		m.cacheMisses,
		m.cacheSize,
		m.sessionCreatedTotal,
		m.sessionPerPlatform,
		m.sessionPerOsType,
		m.sessionPerOsVersion,
		m.sessionPerDeviceModel,
		m.sessionPerDeviceId,
		m.sessionPerLanguage,
		m.sessionPerCountry,
		m.sessionPerCity,
		m.sessionAccessTotal,
		m.sessionAccessPerSecond,
		m.clientVersionAccess,
		m.clientVersionPerSecond,
		m.authenticatedRequests,
		m.unauthenticatedRequests,
		m.userAccessPerMinute,
		m.userAccessPerHour,
		m.userAccessPerDay,
	}

	for _, metric := range metrics {
		if err := m.registry.Register(metric); err != nil {
			m.logger.Error("Failed to register metric", zap.Error(err))
		}
	}
}

// GetMetricsHandler returns the Prometheus metrics handler
func (m *Metrics) GetMetricsHandler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

// RecordHTTPRequest records an HTTP request
func (m *Metrics) RecordHTTPRequest(method, path, status, service string, duration time.Duration) {
	m.httpRequestsTotal.WithLabelValues(method, path, status, service).Inc()
	m.httpRequestDuration.WithLabelValues(method, path, status, service).Observe(duration.Seconds())
}

// RecordHTTPRequestInFlight records an in-flight HTTP request
func (m *Metrics) RecordHTTPRequestInFlight(service string, count float64) {
	m.httpRequestsInFlight.WithLabelValues(service).Set(count)
}

// IncrementHTTPRequestInFlight increments the in-flight request count
func (m *Metrics) IncrementHTTPRequestInFlight(service string) {
	m.httpRequestsInFlight.WithLabelValues(service).Inc()
}

// DecrementHTTPRequestInFlight decrements the in-flight request count
func (m *Metrics) DecrementHTTPRequestInFlight(service string) {
	m.httpRequestsInFlight.WithLabelValues(service).Dec()
}

// RecordDBOperation records a database operation
func (m *Metrics) RecordDBOperation(operation, collection, status, service string, duration time.Duration) {
	m.dbOperationsTotal.WithLabelValues(operation, collection, status, service).Inc()
	m.dbOperationDuration.WithLabelValues(operation, collection, service).Observe(duration.Seconds())
	if status != "success" {
		m.dbOperationsFailed.WithLabelValues(operation, collection, service).Inc()
	}
}

// RecordDBOperationFailed records a failed database operation
func (m *Metrics) RecordDBOperationFailed(operation, collection, service string) {
	m.dbOperationsFailed.WithLabelValues(operation, collection, service).Inc()
}

// RecordDBConnections records active database connections
func (m *Metrics) RecordDBConnections(database, service string, count float64) {
	m.dbConnectionsActive.WithLabelValues(database, service).Set(count)
}

// RecordUserRegistration records a user registration attempt
func (m *Metrics) RecordUserRegistration(status, service string) {
	m.userRegistrationsTotal.WithLabelValues(status, service).Inc()
}

// RecordUserLogin records a user login attempt
func (m *Metrics) RecordUserLogin(status, service string) {
	m.userLoginsTotal.WithLabelValues(status, service).Inc()
}

// RecordUserLoginFailed records a failed user login
func (m *Metrics) RecordUserLoginFailed(reason, service string) {
	m.userLoginsFailed.WithLabelValues(reason, service).Inc()
}

// RecordActiveSessions records the number of active sessions
func (m *Metrics) RecordActiveSessions(service string, count float64) {
	m.activeSessions.WithLabelValues(service).Set(count)
}

// RecordGoroutines records the number of goroutines
func (m *Metrics) RecordGoroutines(service string, count float64) {
	m.goroutinesCount.WithLabelValues(service).Set(count)
}

// RecordMemoryUsage records memory usage
func (m *Metrics) RecordMemoryUsage(memoryType, service string, bytes float64) {
	m.memoryUsageBytes.WithLabelValues(memoryType, service).Set(bytes)
}

// RecordError records an error
func (m *Metrics) RecordError(errorType, component, service string) {
	m.errorsTotal.WithLabelValues(errorType, component, service).Inc()
}

// RecordValidAPICall records a valid API call
func (m *Metrics) RecordValidAPICall(endpoint, method, service string, duration time.Duration) {
	m.apiValidCallsTotal.WithLabelValues(endpoint, method, service).Inc()
	m.apiCallDuration.WithLabelValues(endpoint, method, service).Observe(duration.Seconds())
}

// RecordInvalidAPICall records an invalid API call
func (m *Metrics) RecordInvalidAPICall(endpoint, method, service string, duration time.Duration) {
	m.apiInvalidCallsTotal.WithLabelValues(endpoint, method, service).Inc()
	m.apiCallDuration.WithLabelValues(endpoint, method, service).Observe(duration.Seconds())
}

// RecordValidationErrors records a validation error
func (m *Metrics) RecordValidationErrors(endpoint, method, errors, service string) {
	m.apiValidationErrors.WithLabelValues(endpoint, method, errors, service).Inc()
}

// RecordAPICall records an API call
func (m *Metrics) RecordAPICall(endpoint, method, status, service string, duration time.Duration) {
	m.apiCallsTotal.WithLabelValues(endpoint, method, status, service).Inc()
	m.apiCallDuration.WithLabelValues(endpoint, method, service).Observe(duration.Seconds())
}

// RecordRateLimitHit records a rate limit hit
func (m *Metrics) RecordRateLimitHit(clientID, path, service string) {
	m.rateLimitHits.WithLabelValues(clientID, path, service).Inc()
}

// RecordRateLimitExceeded records a rate limit violation
func (m *Metrics) RecordRateLimitExceeded(clientID, path, service string) {
	m.rateLimitExceeded.WithLabelValues(clientID, path, service).Inc()
}

// RecordCacheHit records a cache hit
func (m *Metrics) RecordCacheHit(cacheOperation, service string) {
	m.cacheHits.WithLabelValues(cacheOperation, service).Inc()
}

// RecordCacheMiss records a cache miss
func (m *Metrics) RecordCacheMiss(cacheOperation, service string) {
	m.cacheMisses.WithLabelValues(cacheOperation, service).Inc()
}

// RecordCacheSize records cache size
func (m *Metrics) RecordCacheSize(cacheOperation, service string, size float64) {
	m.cacheSize.WithLabelValues(cacheOperation, service).Set(size)
}

// RecordSessionCreated records a session creation
func (m *Metrics) RecordSessionCreated(service, platform, osType, osVersion, deviceModel, deviceId, language, country, city string) {
	m.sessionCreatedTotal.WithLabelValues(service).Inc()
	m.sessionPerPlatform.WithLabelValues(platform, service).Inc()
	m.sessionPerOsType.WithLabelValues(osType, service).Inc()
	m.sessionPerOsVersion.WithLabelValues(osVersion, service).Inc()
	m.sessionPerDeviceModel.WithLabelValues(deviceModel, service).Inc()
	m.sessionPerDeviceId.WithLabelValues(deviceId, service).Inc()
	m.sessionPerLanguage.WithLabelValues(language, service).Inc()
	m.sessionPerCountry.WithLabelValues(country, service).Inc()
	m.sessionPerCity.WithLabelValues(city, service).Inc()
}

// RecordSessionAccess records a session access
func (m *Metrics) RecordSessionAccess(service string) {
	m.sessionAccessTotal.WithLabelValues(service).Inc()
	m.sessionAccessPerSecond.WithLabelValues(service).Inc()
}

// RecordClientVersionAccess records a client version access
func (m *Metrics) RecordClientVersionAccess(version, service string) {
	m.clientVersionAccess.WithLabelValues(version, service).Inc()
	m.clientVersionPerSecond.WithLabelValues(version, service).Inc()
}

// RecordAuthenticatedRequest records an authenticated request
func (m *Metrics) RecordAuthenticatedRequest(service string) {
	m.authenticatedRequests.WithLabelValues(service).Inc()
}

// RecordUnauthenticatedRequest records an unauthenticated request
func (m *Metrics) RecordUnauthenticatedRequest(service string) {
	m.unauthenticatedRequests.WithLabelValues(service).Inc()
}

// RecordUserAccess records user access with time-based tracking
func (m *Metrics) RecordUserAccess(email, service string) {
	// For minute, hour, day tracking, we'll use the counter
	// In a real implementation, you might want to use a more sophisticated approach
	// with sliding windows or time-series data
	m.userAccessPerMinute.WithLabelValues(service).Inc()
	m.userAccessPerHour.WithLabelValues(service).Inc()
	m.userAccessPerDay.WithLabelValues(service).Inc()
}

// GetMetrics returns all registered metrics
func (m *Metrics) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"http_requests_total":       m.httpRequestsTotal,
		"http_request_duration":     m.httpRequestDuration,
		"http_requests_in_flight":   m.httpRequestsInFlight,
		"db_operations_total":       m.dbOperationsTotal,
		"db_operation_duration":     m.dbOperationDuration,
		"db_operations_failed":      m.dbOperationsFailed,
		"db_connections_active":     m.dbConnectionsActive,
		"user_registrations_total":  m.userRegistrationsTotal,
		"user_logins_total":         m.userLoginsTotal,
		"user_logins_failed":        m.userLoginsFailed,
		"active_sessions":           m.activeSessions,
		"goroutines_count":          m.goroutinesCount,
		"memory_usage_bytes":        m.memoryUsageBytes,
		"errors_total":              m.errorsTotal,
		"api_valid_calls_total":     m.apiValidCallsTotal,
		"api_invalid_calls_total":   m.apiInvalidCallsTotal,
		"api_validation_errors":     m.apiValidationErrors,
		"api_calls_total":           m.apiCallsTotal,
		"api_call_duration":         m.apiCallDuration,
		"rate_limit_hits":           m.rateLimitHits,
		"rate_limit_exceeded":       m.rateLimitExceeded,
		"cache_hits":                m.cacheHits,
		"cache_misses":              m.cacheMisses,
		"cache_size":                m.cacheSize,
		"session_created_total":     m.sessionCreatedTotal,
		"session_per_platform":      m.sessionPerPlatform,
		"session_per_os_type":       m.sessionPerOsType,
		"session_per_os_version":    m.sessionPerOsVersion,
		"session_per_device_model":  m.sessionPerDeviceModel,
		"session_per_device_id":     m.sessionPerDeviceId,
		"session_per_language":      m.sessionPerLanguage,
		"session_per_country":       m.sessionPerCountry,
		"session_per_city":          m.sessionPerCity,
		"session_access_total":      m.sessionAccessTotal,
		"session_access_per_second": m.sessionAccessPerSecond,
		"client_version_access":     m.clientVersionAccess,
		"client_version_per_second": m.clientVersionPerSecond,
		"authenticated_requests":    m.authenticatedRequests,
		"unauthenticated_requests":  m.unauthenticatedRequests,
		"user_access_per_minute":    m.userAccessPerMinute,
		"user_access_per_hour":      m.userAccessPerHour,
		"user_access_per_day":       m.userAccessPerDay,
	}
}
