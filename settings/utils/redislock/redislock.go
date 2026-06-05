package redislock

import (
	"context"
	"fmt"
	"time"

	"github.com/glodb/keel/internal/utils"
	"github.com/glodb/keel/settings/cachesettings/cache"
	"github.com/glodb/keel/settings/logger"
)

const (
	// DefaultLockTimeout is the default lock expiration time
	DefaultLockTimeout = 10 * time.Second
	// DefaultRetryDelay is the delay between lock acquisition retries
	DefaultRetryDelay = 50 * time.Millisecond
	// DefaultMaxRetries is the maximum number of lock acquisition attempts
	DefaultMaxRetries = 20
)

// RedisLock represents a distributed lock using Redis
type RedisLock struct {
	key        string
	value      string
	timeout    time.Duration
	ctx        context.Context
	acquired   bool
	maxRetries int
	retryDelay time.Duration
}

// NewRedisLock creates a new Redis distributed lock
func NewRedisLock(key string) *RedisLock {
	return &RedisLock{
		key:        fmt.Sprintf("lock:%s", key),
		value:      utils.GetInstance().GenerateXID(), // Unique value to identify this lock
		timeout:    DefaultLockTimeout,
		ctx:        cache.GetCacheContext(),
		acquired:   false,
		maxRetries: DefaultMaxRetries,
		retryDelay: DefaultRetryDelay,
	}
}

// WithTimeout sets a custom timeout for the lock
func (l *RedisLock) WithTimeout(timeout time.Duration) *RedisLock {
	l.timeout = timeout
	return l
}

// WithMaxRetries sets the maximum number of retry attempts
func (l *RedisLock) WithMaxRetries(maxRetries int) *RedisLock {
	l.maxRetries = maxRetries
	return l
}

// WithRetryDelay sets the delay between retry attempts
func (l *RedisLock) WithRetryDelay(delay time.Duration) *RedisLock {
	l.retryDelay = delay
	return l
}

// TryLock attempts to acquire the lock without retrying
func (l *RedisLock) TryLock() (bool, error) {
	// Use SetNX (Set if Not eXists) with expiration
	// This is atomic in Redis
	err := cache.GetCache().SetEx(l.ctx, l.key, []byte(l.value), int(l.timeout.Seconds()))
	if err != nil {
		// Check if error is because key already exists
		exists, existsErr := cache.GetCache().Exists(l.ctx, l.key)
		if existsErr == nil && exists {
			return false, nil // Lock already held by someone else
		}
		logger.Log().Error("Failed to acquire lock",
			logger.StringField("key", l.key),
			logger.StringField("error", err.Error()))
		return false, err
	}

	l.acquired = true
	return true, nil
}

// Lock acquires the lock with retries
func (l *RedisLock) Lock() error {
	for attempt := 0; attempt < l.maxRetries; attempt++ {
		acquired, err := l.TryLock()
		if err != nil {
			return err
		}
		if acquired {
			logger.Log().Info("Lock acquired",
				logger.StringField("key", l.key),
				logger.IntField("attempt", attempt+1))
			return nil
		}

		// Wait before retrying
		if attempt < l.maxRetries-1 {
			time.Sleep(l.retryDelay)
		}
	}

	return fmt.Errorf("failed to acquire lock after %d attempts", l.maxRetries)
}

// Unlock releases the lock
// Only releases if the lock value matches (prevents releasing someone else's lock)
func (l *RedisLock) Unlock() error {
	if !l.acquired {
		return nil // Nothing to unlock
	}

	// Get current value to verify we own the lock
	currentValue, err := cache.GetCache().Get(l.ctx, l.key)
	if err != nil {
		// Lock might have already expired
		logger.Log().Warn("Lock not found during unlock",
			logger.StringField("key", l.key))
		l.acquired = false
		return nil
	}

	// Verify we own the lock
	if string(currentValue) != l.value {
		logger.Log().Warn("Lock value mismatch - not owned by this process",
			logger.StringField("key", l.key))
		return fmt.Errorf("lock not owned by this process")
	}

	// Delete the lock
	err = cache.GetCache().Del(l.ctx, l.key)
	if err != nil {
		logger.Log().Error("Failed to release lock",
			logger.StringField("key", l.key),
			logger.StringField("error", err.Error()))
		return err
	}

	l.acquired = false
	logger.Log().Info("Lock released", logger.StringField("key", l.key))
	return nil
}

// IsAcquired returns whether the lock is currently held
func (l *RedisLock) IsAcquired() bool {
	return l.acquired
}

// Extend extends the lock timeout (useful for long-running operations)
func (l *RedisLock) Extend(additionalTime time.Duration) error {
	if !l.acquired {
		return fmt.Errorf("cannot extend lock that is not acquired")
	}

	// Get current value to verify we own the lock
	currentValue, err := cache.GetCache().Get(l.ctx, l.key)
	if err != nil {
		return fmt.Errorf("failed to verify lock ownership: %w", err)
	}

	if string(currentValue) != l.value {
		return fmt.Errorf("lock not owned by this process")
	}

	// Set new expiration
	err = cache.GetCache().SetEx(l.ctx, l.key, []byte(l.value), int((l.timeout + additionalTime).Seconds()))
	if err != nil {
		return fmt.Errorf("failed to extend lock: %w", err)
	}

	l.timeout += additionalTime
	logger.Log().Info("Lock extended",
		logger.StringField("key", l.key),
		logger.StringField("newTimeout", l.timeout.String()))
	return nil
}

// ExecuteWithLock executes a function while holding the lock
// Automatically acquires and releases the lock
func ExecuteWithLock(key string, fn func() error) error {
	lock := NewRedisLock(key)

	// Acquire lock
	if err := lock.Lock(); err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	// Ensure lock is released
	defer func() {
		if err := lock.Unlock(); err != nil {
			logger.Log().Error("Failed to unlock after execution",
				logger.StringField("key", key),
				logger.StringField("error", err.Error()))
		}
	}()

	// Execute function
	return fn()
}
