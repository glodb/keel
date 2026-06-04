package circuitbreaker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/glodb/keel/settings/errors"
	"github.com/glodb/keel/settings/logger"
)

// State represents the current state of the circuit breaker
type State int

const (
	// StateClosed means the circuit breaker allows requests to pass through
	StateClosed State = iota
	// StateOpen means the circuit breaker blocks requests and returns errors immediately
	StateOpen
	// StateHalfOpen means the circuit breaker allows a limited number of test requests
	StateHalfOpen
)

// String returns the string representation of the state
func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// Config holds the configuration for a circuit breaker
type Config struct {
	// Name identifies the circuit breaker for logging and metrics
	Name string `json:"name"`

	// MaxRequests is the maximum number of requests allowed to pass through
	// when the circuit breaker is half-open
	MaxRequests uint32 `json:"max_requests"`

	// Interval is the cyclic period to clear the internal counts
	Interval time.Duration `json:"interval"`

	// Timeout is the period of the open state,
	// after which the state becomes half-open
	Timeout time.Duration `json:"timeout"`

	// ReadyToTrip returns true when the circuit breaker should open
	// Default: 5 consecutive failures
	ReadyToTrip func(counts Counts) bool `json:"-"`

	// OnStateChange is called whenever the state changes
	OnStateChange func(name string, from State, to State) `json:"-"`
}

// Counts holds the statistics for the circuit breaker
type Counts struct {
	Requests             uint32 `json:"requests"`
	TotalSuccesses       uint32 `json:"total_successes"`
	TotalFailures        uint32 `json:"total_failures"`
	ConsecutiveSuccesses uint32 `json:"consecutive_successes"`
	ConsecutiveFailures  uint32 `json:"consecutive_failures"`
}

// Reset clears all counts
func (c *Counts) Reset() {
	c.Requests = 0
	c.TotalSuccesses = 0
	c.TotalFailures = 0
	c.ConsecutiveSuccesses = 0
	c.ConsecutiveFailures = 0
}

// OnRequest increments the request count
func (c *Counts) OnRequest() {
	c.Requests++
}

// OnSuccess increments success counts and resets consecutive failures
func (c *Counts) OnSuccess() {
	c.TotalSuccesses++
	c.ConsecutiveSuccesses++
	c.ConsecutiveFailures = 0
}

// OnFailure increments failure counts and resets consecutive successes
func (c *Counts) OnFailure() {
	c.TotalFailures++
	c.ConsecutiveFailures++
	c.ConsecutiveSuccesses = 0
}

// FailureRate returns the failure rate as a percentage
func (c *Counts) FailureRate() float64 {
	if c.Requests == 0 {
		return 0.0
	}
	return float64(c.TotalFailures) / float64(c.Requests) * 100.0
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	config  Config
	mutex   sync.RWMutex
	state   State
	counts  Counts
	expiry  time.Time
	logger  *logger.Logger
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration
func NewCircuitBreaker(config Config) *CircuitBreaker {
	cb := &CircuitBreaker{
		config:  config,
		state:   StateClosed,
		logger:  logger.Log(),
	}

	// Set default values
	if cb.config.MaxRequests == 0 {
		cb.config.MaxRequests = 1
	}
	if cb.config.Interval == 0 {
		cb.config.Interval = 60 * time.Second
	}
	if cb.config.Timeout == 0 {
		cb.config.Timeout = 60 * time.Second
	}
	if cb.config.ReadyToTrip == nil {
		cb.config.ReadyToTrip = func(counts Counts) bool {
			return counts.ConsecutiveFailures > 5
		}
	}

	cb.logger.Info("Circuit breaker created",
		logger.StringField("name", config.Name),
		logger.StringField("max_requests", fmt.Sprintf("%d", config.MaxRequests)),
		logger.StringField("interval", config.Interval.String()),
		logger.StringField("timeout", config.Timeout.String()))

	return cb
}

// Execute runs the given function if the circuit breaker allows it
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	// Check if we can proceed
	if err := cb.beforeRequest(); err != nil {
		return err
	}

	// Execute the function and handle the result
	err := fn()
	cb.afterRequest(err == nil)

	return err
}

// ExecuteWithFallback runs the given function if the circuit breaker allows it,
// otherwise runs the fallback function
func (cb *CircuitBreaker) ExecuteWithFallback(ctx context.Context, fn func() error, fallback func() error) error {
	if err := cb.beforeRequest(); err != nil {
		cb.logger.Debug("Circuit breaker is open, executing fallback",
			logger.StringField("circuit_breaker", cb.config.Name),
			logger.ErrorField("error", err))

		if fallback != nil {
			return fallback()
		}
		return err
	}

	err := fn()
	cb.afterRequest(err == nil)
	return err
}

// beforeRequest checks if a request can be made
func (cb *CircuitBreaker) beforeRequest() error {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, counts := cb.currentState(now)

	if state == StateOpen {
		return errors.NewAppError(
			errors.ErrCodeCircuitBreakerOpen,
			fmt.Sprintf("Circuit breaker '%s' is open", cb.config.Name),
			nil,
		).WithMetadata("circuit_breaker", cb.config.Name).
			WithMetadata("state", state.String()).
			WithMetadata("failure_rate", fmt.Sprintf("%.2f%%", counts.FailureRate()))
	}

	if state == StateHalfOpen && counts.Requests >= cb.config.MaxRequests {
		return errors.NewAppError(
			errors.ErrCodeCircuitBreakerOpen,
			fmt.Sprintf("Circuit breaker '%s' is half-open with max requests reached", cb.config.Name),
			nil,
		).WithMetadata("circuit_breaker", cb.config.Name).
			WithMetadata("state", state.String()).
			WithMetadata("requests", counts.Requests).
			WithMetadata("max_requests", cb.config.MaxRequests)
	}

	cb.counts.OnRequest()
	return nil
}

// afterRequest handles the result of a request
func (cb *CircuitBreaker) afterRequest(success bool) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, _ := cb.currentState(now)

	if success {
		cb.counts.OnSuccess()

		if state == StateHalfOpen && cb.counts.ConsecutiveSuccesses >= cb.config.MaxRequests {
			cb.setState(StateClosed, now)
		}
	} else {
		cb.counts.OnFailure()

		if cb.config.ReadyToTrip(cb.counts) {
			cb.setState(StateOpen, now)
		}
	}
}

// currentState returns the current state of the circuit breaker
func (cb *CircuitBreaker) currentState(now time.Time) (State, Counts) {
	switch cb.state {
	case StateClosed:
		if !cb.expiry.IsZero() && cb.expiry.Before(now) {
			cb.counts.Reset()
			cb.expiry = now.Add(cb.config.Interval)
		}
	case StateOpen:
		if cb.expiry.Before(now) {
			cb.setState(StateHalfOpen, now)
		}
	}
	return cb.state, cb.counts
}

// setState changes the state of the circuit breaker
func (cb *CircuitBreaker) setState(state State, now time.Time) {
	if cb.state == state {
		return
	}

	prevState := cb.state
	cb.state = state

	cb.logger.Info("Circuit breaker state changed",
		logger.StringField("circuit_breaker", cb.config.Name),
		logger.StringField("from_state", prevState.String()),
		logger.StringField("to_state", state.String()),
		logger.AnyField("counts", cb.counts))

	switch state {
	case StateClosed:
		cb.counts.Reset()
		cb.expiry = now.Add(cb.config.Interval)
	case StateOpen:
		cb.expiry = now.Add(cb.config.Timeout)
	case StateHalfOpen:
		cb.counts.Reset()
		cb.expiry = time.Time{} // No expiry for half-open state
	}

	if cb.config.OnStateChange != nil {
		cb.config.OnStateChange(cb.config.Name, prevState, state)
	}

	// Record metrics
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() State {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	now := time.Now()
	state, _ := cb.currentState(now)
	return state
}

// GetCounts returns the current counts
func (cb *CircuitBreaker) GetCounts() Counts {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.counts
}

// Reset resets the circuit breaker to its initial state
func (cb *CircuitBreaker) Reset() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.state = StateClosed
	cb.counts.Reset()
	cb.expiry = time.Now().Add(cb.config.Interval)

	cb.logger.Info("Circuit breaker reset",
		logger.StringField("circuit_breaker", cb.config.Name))
}

// GetName returns the name of the circuit breaker
func (cb *CircuitBreaker) GetName() string {
	return cb.config.Name
}

// IsHealthy returns true if the circuit breaker is in a healthy state
func (cb *CircuitBreaker) IsHealthy() bool {
	return cb.GetState() != StateOpen
}

// GetMetrics returns metrics for the circuit breaker
func (cb *CircuitBreaker) GetMetrics() map[string]interface{} {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	now := time.Now()
	state, counts := cb.currentState(now)

	return map[string]interface{}{
		"name":                  cb.config.Name,
		"state":                 state.String(),
		"requests":              counts.Requests,
		"total_successes":       counts.TotalSuccesses,
		"total_failures":        counts.TotalFailures,
		"consecutive_successes": counts.ConsecutiveSuccesses,
		"consecutive_failures":  counts.ConsecutiveFailures,
		"failure_rate":          counts.FailureRate(),
		"expiry":                cb.expiry,
		"max_requests":          cb.config.MaxRequests,
		"interval":              cb.config.Interval.String(),
		"timeout":               cb.config.Timeout.String(),
	}
}

// CircuitBreakerManager manages multiple circuit breakers
type CircuitBreakerManager struct {
	breakers sync.Map // map[string]*CircuitBreaker
	logger   *logger.Logger
}

var (
	managerInstance *CircuitBreakerManager
	managerOnce     sync.Once
)

// GetManager returns the singleton circuit breaker manager
func GetManager() *CircuitBreakerManager {
	managerOnce.Do(func() {
		managerInstance = &CircuitBreakerManager{
			logger: logger.Log(),
		}
	})
	return managerInstance
}

// GetOrCreate gets an existing circuit breaker or creates a new one
func (cbm *CircuitBreakerManager) GetOrCreate(name string, config Config) *CircuitBreaker {
	if cb, ok := cbm.breakers.Load(name); ok {
		return cb.(*CircuitBreaker)
	}

	config.Name = name
	cb := NewCircuitBreaker(config)
	cbm.breakers.Store(name, cb)

	cbm.logger.Info("Circuit breaker registered",
		logger.StringField("name", name))

	return cb
}

// Get retrieves a circuit breaker by name
func (cbm *CircuitBreakerManager) Get(name string) (*CircuitBreaker, bool) {
	if cb, ok := cbm.breakers.Load(name); ok {
		return cb.(*CircuitBreaker), true
	}
	return nil, false
}

// GetAll returns all circuit breakers
func (cbm *CircuitBreakerManager) GetAll() map[string]*CircuitBreaker {
	result := make(map[string]*CircuitBreaker)
	cbm.breakers.Range(func(key, value interface{}) bool {
		result[key.(string)] = value.(*CircuitBreaker)
		return true
	})
	return result
}

// ResetAll resets all circuit breakers
func (cbm *CircuitBreakerManager) ResetAll() {
	cbm.breakers.Range(func(key, value interface{}) bool {
		value.(*CircuitBreaker).Reset()
		return true
	})
	cbm.logger.Info("All circuit breakers reset")
}

// GetHealthStatus returns the health status of all circuit breakers
func (cbm *CircuitBreakerManager) GetHealthStatus() map[string]bool {
	result := make(map[string]bool)
	cbm.breakers.Range(func(key, value interface{}) bool {
		result[key.(string)] = value.(*CircuitBreaker).IsHealthy()
		return true
	})
	return result
}

// GetAllMetrics returns metrics for all circuit breakers
func (cbm *CircuitBreakerManager) GetAllMetrics() map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})
	cbm.breakers.Range(func(key, value interface{}) bool {
		result[key.(string)] = value.(*CircuitBreaker).GetMetrics()
		return true
	})
	return result
}
