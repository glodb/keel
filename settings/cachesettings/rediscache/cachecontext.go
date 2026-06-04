package rediscache

import (
	"context"
	"time"

	"github.com/glodb/keel/settings/configmanager"
	"github.com/glodb/keel/settings/logger"
)

// GetCacheContext returns a context with timeout for cache operations
// Returns both context and cancel function for proper resource management
func (cache *RedisCache) GetContext() (context.Context, context.CancelFunc) {
	timeout := time.Duration(configmanager.GetInstance().CacheContextTimeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	logger.Log().Debug("Cache context created",
		logger.DurationField("timeout", timeout))

	return ctx, cancel
}

// GetCacheContext returns a context without timeout for cache operations
// this context is actually used on semaphore operations where we don't need to cancel the context
func (cache *RedisCache) GetCacheContext() context.Context {
	ctx := context.Background()

	return ctx
}

// GetCacheContextWithDeadline returns a context with a specific deadline
func (cache *RedisCache) GetCacheContextWithDeadline(deadline time.Time) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithDeadline(context.Background(), deadline)

	logger.Log().Debug("Cache context created with deadline",
		logger.TimeField("deadline", deadline))

	return ctx, cancel
}

// GetCacheContextWithValue returns a context with timeout and additional values
func (cache *RedisCache) GetCacheContextWithValue(key, value interface{}) (context.Context, context.CancelFunc) {
	timeout := time.Duration(configmanager.GetInstance().CacheContextTimeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	ctx = context.WithValue(ctx, key, value)

	logger.Log().Debug("Cache context created with value",
		logger.DurationField("timeout", timeout),
		logger.AnyField("key", key))

	return ctx, cancel
}

// IsContextExpired checks if a context has expired
func IsContextExpired(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

// GetContextRemainingTime returns the remaining time before context expires
func (cache *RedisCache) GetContextRemainingTime(ctx context.Context) (time.Duration, bool) {
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		return remaining, remaining > 0
	}
	return 0, false
}
