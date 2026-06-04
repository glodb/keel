package cache

import (
	"context"
	"sync"

	"github.com/glodb/keel/settings/cachesettings/rediscache"
	"github.com/glodb/keel/settings/configmanager"
)

type LoadedCache struct {
	cache Cache
}

var (
	overrideMu    sync.RWMutex
	overrideCache Cache // user-supplied override; takes precedence when non-nil
)

var getCache = sync.OnceValue(func() Cache {
	if configmanager.GetInstance().CacheType == "redis" {
		return rediscache.GetInstance()
	}
	// Default to redis cache when no specific type is configured.
	return rediscache.GetInstance()
})

// current returns the active cache: the user override if one has been set,
// otherwise the lazily-initialised default.
func current() Cache {
	overrideMu.RLock()
	c := overrideCache
	overrideMu.RUnlock()
	if c != nil {
		return c
	}
	return getCache()
}

// SetCache overrides the global cache implementation for all callers. Useful
// for injecting a custom backend or a fake in tests. Passing nil reverts to
// the default implementation.
func SetCache(cache Cache) {
	overrideMu.Lock()
	overrideCache = cache
	overrideMu.Unlock()
}

func GetInstance() *LoadedCache {
	return &LoadedCache{
		cache: current(),
	}
}

func GetCache() Cache {
	return current()
}

func GetCacheContext() context.Context {
	cache := current()
	if cache == nil {
		return context.Background()
	}
	return cache.GetCacheContext()
}

// SetCache overrides the global cache implementation. It delegates to the
// package-level SetCache so the change is visible to every caller, not just
// this wrapper instance.
func (l *LoadedCache) SetCache(cache Cache) {
	l.cache = cache
	SetCache(cache)
}

// Fallback implementations when Redis is not available
func (l *LoadedCache) GetCacheContext() context.Context {
	if l.cache != nil {
		return l.cache.GetCacheContext()
	}
	return context.Background()
}
