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

var getCache = sync.OnceValue(func() Cache {
	instance := &LoadedCache{}

	if configmanager.GetInstance().CacheType == "redis" {
		instance.cache = rediscache.GetInstance()
	}

	//By default load redis cache but can be overridden by the user
	if instance.cache == nil {
		instance.cache = rediscache.GetInstance()
	}

	return instance.cache
})

func GetInstance() *LoadedCache {
	return &LoadedCache{
		cache: getCache(),
	}
}

func GetCache() Cache {
	return getCache()
}

func GetCacheContext() context.Context {
	cache := getCache()
	if cache == nil {
		return context.Background()
	}
	return cache.GetCacheContext()
}

func (l *LoadedCache) SetCache(cache Cache) {
	l.cache = cache
}

// Fallback implementations when Redis is not available
func (l *LoadedCache) GetCacheContext() context.Context {
	if l.cache != nil {
		return l.cache.GetCacheContext()
	}
	return context.Background()
}
