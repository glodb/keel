package rediscache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/glodb/keel/app/models/cachemodels"
	"github.com/glodb/keel/settings/configmanager"
	"github.com/glodb/keel/settings/logger"
	"github.com/glodb/keel/settings/metrics"

	"github.com/gomodule/redigo/redis"
	"golang.org/x/sync/semaphore"
)

type RedisCache struct {
	pool      *redis.Pool
	semaphore *semaphore.Weighted
}

var getInstance = sync.OnceValue(func() *RedisCache {
	instance := &RedisCache{}
	pool, err := instance.newPool()
	if err != nil {
		logger.Log().Error("Failed to create Redis pool", logger.ErrorField("error", err))
		return nil
	}
	instance.pool = pool
	logger.Log().Debug("RedisCache initialized", logger.IntField("maxConnections", configmanager.GetInstance().Redis.RedisMaxConnections))
	instance.semaphore = semaphore.NewWeighted(int64(configmanager.GetInstance().Redis.RedisMaxConnections))
	return instance
})

func GetInstance() *RedisCache {
	return getInstance()
}

// refresh Pool is used to refresh the redis pool if the pool is not working
func (cache *RedisCache) refreshPool() error {
	var err error
	cache.pool, err = cache.newPool()
	return err
}

func (cache *RedisCache) GetConnection(ctx context.Context) redis.Conn {
	cache.semaphore.Acquire(ctx, 1)
	c := cache.pool.Get()
	return c
}

func (cache *RedisCache) ReleaseConnection(conn redis.Conn) {
	conn.Close()
	cache.semaphore.Release(1)
}

func (cache *RedisCache) newPool() (*redis.Pool, error) {
	var redErr error
	pool := redis.Pool{
		MaxIdle:     configmanager.GetInstance().Redis.RedisMaxIdleConnections,
		MaxActive:   configmanager.GetInstance().Redis.RedisMaxConnections, // max number of connections
		IdleTimeout: 240 * time.Second,
		Wait:        true, // Wait for connection to be available
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial(configmanager.GetInstance().Redis.RedisCon, configmanager.GetInstance().Redis.RedisAddress)
			if err != nil {
				redErr = err
				logger.Log().Error("Error dialing Redis", logger.ErrorField("error", err))
				return nil, err
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}

	return &pool, redErr
}

func (cache *RedisCache) Set(ctx context.Context, key string, value []byte) error {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {
		logger.Log().Debug("Error acquiring semaphore", logger.ErrorField("error", err))
		return err
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()

	_, err := c.Do("SET", configmanager.GetInstance().DeploymentEnv+key, value)
	if err != nil {
		logger.Log().Debug("Error setting key", logger.StringField("key", key), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_set", configmanager.GetInstance().ServiceLBName)
	} else {
		metrics.GetInstance().RecordCacheHit("redis_set", configmanager.GetInstance().ServiceLBName)
	}
	return err
}

// AcquireLock acquires a distributed lock using SETNX command
func (cache *RedisCache) AcquireCacheLock(ctx context.Context, lockKey string, expirationMilli int64) (bool, error) {
	// return false, errors.New("simulated lock acquisition failure")
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {
		logger.Log().Debug("Error acquiring semaphore", logger.ErrorField("error", err))
		return false, err
	}
	defer cache.semaphore.Release(1)
	c := cache.pool.Get()
	defer c.Close()
	for i := 0; i < configmanager.GetInstance().RedisRetries; i++ {
		// Try to set the lock key with expiration
		result, err := redis.String(c.Do("SET", configmanager.GetInstance().DeploymentEnv+lockKey, "1", "NX", "PX", int(expirationMilli)))
		if err != nil {
			return false, err
		}
		if result == "OK" {
			return true, nil
		}
		// Sleep before retrying
		time.Sleep(time.Duration(configmanager.GetInstance().RedisRetryInterval))
	}
	return false, fmt.Errorf("failed to acquire lock after %d retries", configmanager.GetInstance().RedisRetries)
}

func (cache *RedisCache) AcquireCacheLock2(lockKey string, expirationMilli int64) (bool, error) {
	return false, errors.New("simulated lock acquisition failure")
}

// ReleaseLock releases the distributed lock by deleting the lock key, del do same thing but just for more readable code
func (cache *RedisCache) ReleaseLock(ctx context.Context, lockKey string) error {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {
		logger.Log().Debug("Error acquiring semaphore", logger.ErrorField("error", err))
		return err
	}
	defer cache.semaphore.Release(1)
	c := cache.pool.Get()
	defer c.Close()
	_, err := c.Do("DEL", configmanager.GetInstance().DeploymentEnv+lockKey)

	if err != nil {
		metrics.GetInstance().RecordCacheMiss("redis_release_lock", configmanager.GetInstance().ServiceLBName)
	} else {
		metrics.GetInstance().RecordCacheHit("redis_release_lock", configmanager.GetInstance().ServiceLBName)
	}
	return err
}

func (cache *RedisCache) SetEx(ctx context.Context, key string, value []byte, expiryTime int) error {
	cache.semaphore.Acquire(ctx, 1)
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	_, err := c.Do("SETEX", configmanager.GetInstance().DeploymentEnv+key, expiryTime, value)
	if err != nil {
		logger.Log().Debug("Error setting key", logger.StringField("key", key), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_set_ex", configmanager.GetInstance().ServiceLBName)
	} else {
		metrics.GetInstance().RecordCacheHit("redis_set_ex", configmanager.GetInstance().ServiceLBName)
	}

	return err
}

func (cache *RedisCache) SUnion(ctx context.Context, keys ...interface{}) []string {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {
		return []string{}
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Strings(c.Do("SUNION", keys...))
	if err != nil {
		logger.Log().Debug("Error getting union from sets", logger.AnyField("keys", keys), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_sunion", configmanager.GetInstance().ServiceLBName)
		return []string{}
	}
	if data == nil {
		logger.Log().Debug("No data found in union", logger.AnyField("keys", keys))
		return []string{}
	}
	metrics.GetInstance().RecordCacheHit("redis_sunion", configmanager.GetInstance().ServiceLBName)
	return data
}

func (cache *RedisCache) GetInt(ctx context.Context, key string) (int64, error) {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {
		logger.Log().Debug("Error acquiring semaphore", logger.ErrorField("error", err))
		return -1, err
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	var data int64
	dataint, err := c.Do("GET", configmanager.GetInstance().DeploymentEnv+key)
	if err != nil {
		logger.Log().Debug("Error getting int", logger.StringField("key", key), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_get_int", configmanager.GetInstance().ServiceLBName)
		return -1, err
	}
	metrics.GetInstance().RecordCacheHit("redis_get_int", configmanager.GetInstance().ServiceLBName)
	if dataint != nil {
		data, err = redis.Int64(dataint, err)
	}
	return data, err
}

func (cache *RedisCache) SetInt(ctx context.Context, key string, value int) error {
	cache.semaphore.Acquire(ctx, 1)
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)

	defer c.Close()
	_, err := c.Do("SET", configmanager.GetInstance().DeploymentEnv+key, value)
	if err != nil {
		logger.Log().Debug("Error setting int", logger.StringField("key", key), logger.IntField("value", value), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_set_int", configmanager.GetInstance().ServiceLBName)
	} else {
		metrics.GetInstance().RecordCacheHit("redis_set_int", configmanager.GetInstance().ServiceLBName)
	}
	return err
}

func (cache *RedisCache) Increment(ctx context.Context, key string) (int64, error) {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {

		return -1, err
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)

	defer c.Close()
	val, err := c.Do("INCR", configmanager.GetInstance().DeploymentEnv+key)
	if err != nil {
		metrics.GetInstance().RecordCacheMiss("redis_increment", configmanager.GetInstance().ServiceLBName)
		return -1, err
	}
	metrics.GetInstance().RecordCacheHit("redis_increment", configmanager.GetInstance().ServiceLBName)
	return val.(int64), err
}

func (cache *RedisCache) Decrement(ctx context.Context, key string) (int64, error) {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {

		return -1, err
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)

	defer c.Close()
	val, err := c.Do("DECR", configmanager.GetInstance().DeploymentEnv+key)
	if err != nil {
		logger.Log().Debug("Error decrementing", logger.StringField("key", key), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_decrement", configmanager.GetInstance().ServiceLBName)
		return -1, err
	}
	metrics.GetInstance().RecordCacheHit("redis_decrement", configmanager.GetInstance().ServiceLBName)
	return val.(int64), err
}

func (cache *RedisCache) SetString(ctx context.Context, key string, value string) error {
	cache.semaphore.Acquire(ctx, 1)

	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	_, err := c.Do("SET", configmanager.GetInstance().DeploymentEnv+key, value)
	if err != nil {
		logger.Log().Debug("Error setting string", logger.StringField("key", key), logger.StringField("value", value), logger.ErrorField("error", err))
		v := string(value)
		if len(v) > 15 {
			v = v[0:12] + "..."
		}
		metrics.GetInstance().RecordCacheMiss("redis_set_string", configmanager.GetInstance().ServiceLBName)
	} else {
		metrics.GetInstance().RecordCacheHit("redis_set_string", configmanager.GetInstance().ServiceLBName)
	}
	return err
}

func (cache *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {

		return nil, err
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	var data []byte
	dataint, err := c.Do("GET", configmanager.GetInstance().DeploymentEnv+key)
	if err != nil {
		logger.Log().Debug("Error getting", logger.StringField("key", key), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_get", configmanager.GetInstance().ServiceLBName)
		return []byte{}, err
	}
	if dataint != nil {
		data, err = redis.Bytes(dataint, err)
	}
	metrics.GetInstance().RecordCacheHit("redis_get", configmanager.GetInstance().ServiceLBName)
	return data, err
}

func (cache *RedisCache) GetKeys(ctx context.Context, pattern string) []string {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {

		return nil
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()

	data, err := redis.Strings(c.Do("Keys", pattern))
	if err != nil {
		logger.Log().Debug("Error getting keys", logger.StringField("pattern", pattern), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_get_keys", configmanager.GetInstance().ServiceLBName)
		return []string{}
	}
	metrics.GetInstance().RecordCacheHit("redis_get_keys", configmanager.GetInstance().ServiceLBName)
	return data
}
func (cache *RedisCache) GetString(ctx context.Context, key string) (string, error) {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {
		logger.Log().Debug("Error acquiring semaphore", logger.ErrorField("error", err))
		return "", err
	}
	defer cache.semaphore.Release(1)

	c := cache.pool.Get()
	defer c.Close()

	var data string
	dataint, err := c.Do("GET", configmanager.GetInstance().DeploymentEnv+key)
	if err != nil {
		logger.Log().Debug("Error getting string", logger.StringField("key", key), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_get_string", configmanager.GetInstance().ServiceLBName)
		return "", err
	}
	if dataint != nil {
		data, err = redis.String(dataint, err)
		if err != nil {
			logger.Log().Debug("Error converting to string", logger.ErrorField("error", err))
		}
	}
	metrics.GetInstance().RecordCacheHit("redis_get_string", configmanager.GetInstance().ServiceLBName)
	return data, err
}

func (cache *RedisCache) Del(ctx context.Context, key string) error {
	cache.semaphore.Acquire(ctx, 1)
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	_, err := c.Do("DEL", configmanager.GetInstance().DeploymentEnv+key)
	if err != nil {
		logger.Log().Debug("Error deleting", logger.StringField("key", key), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_del", configmanager.GetInstance().ServiceLBName)
	} else {
		metrics.GetInstance().RecordCacheHit("redis_del", configmanager.GetInstance().ServiceLBName)
	}
	return err
}

func (cache *RedisCache) DelMany(ctx context.Context, keys []string) error {
	cache.semaphore.Acquire(ctx, 1)
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()

	newKeys := make([]string, 0)
	for _, key := range keys {
		newKeys = append(newKeys, configmanager.GetInstance().DeploymentEnv+key)
	}

	_, err := c.Do("DEL", redis.Args{}.AddFlat(newKeys)...)
	if err != nil {
		logger.Log().Debug("Error deleting many", logger.AnyField("keys", keys), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_del_many", configmanager.GetInstance().ServiceLBName)
	} else {
		metrics.GetInstance().RecordCacheHit("redis_del_many", configmanager.GetInstance().ServiceLBName)
	}
	return err
}

func (cache *RedisCache) Append(ctx context.Context, key string, value interface{}) error {
	cache.semaphore.Acquire(ctx, 1)
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	_, err := c.Do("APPEND", configmanager.GetInstance().DeploymentEnv+key, value)
	if err != nil {
		logger.Log().Debug("Error appending", logger.StringField("key", key), logger.AnyField("value", value), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_append", configmanager.GetInstance().ServiceLBName)
	} else {
		metrics.GetInstance().RecordCacheHit("redis_append", configmanager.GetInstance().ServiceLBName)
	}
	return err
}

func (cache *RedisCache) SAdd(ctx context.Context, key string, values ...interface{}) error {
	if len(values) < 1 {
		return errors.New("not enough parameters")
	}

	if err := cache.semaphore.Acquire(ctx, 1); err != nil {
		logger.Log().Debug("Error acquiring semaphore", logger.ErrorField("error", err))
		return err
	}
	if ctx.Err() != nil {
		logger.Log().Debug("Context error", logger.ErrorField("error", ctx.Err()))
		return ctx.Err()
	}
	defer cache.semaphore.Release(1)

	c := cache.pool.Get()
	defer c.Close()

	keys := []interface{}{configmanager.GetInstance().DeploymentEnv + key}
	keys = append(keys, values...)

	_, err := c.Do("SADD", keys...)
	if err != nil {
		logger.Log().Debug("Error adding to set", logger.StringField("key", key), logger.AnyField("values", values), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_sadd", configmanager.GetInstance().ServiceLBName)
		return err
	}
	metrics.GetInstance().RecordCacheHit("redis_sadd", configmanager.GetInstance().ServiceLBName)

	return nil
}

func (cache *RedisCache) SPop(ctx context.Context, key string, count int) []string {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {
		return []string{}
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Strings(c.Do("SPOP", configmanager.GetInstance().DeploymentEnv+key, count))
	if err != nil {
		logger.Log().Debug("Error popping from set", logger.StringField("key", key), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_spop", configmanager.GetInstance().ServiceLBName)
	} else {
		metrics.GetInstance().RecordCacheHit("redis_spop", configmanager.GetInstance().ServiceLBName)
	}
	return data
}

func (cache *RedisCache) SPopTransaction(ctx context.Context, key string, count int, lockKey string) []string {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {
		return []string{}
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()

	c.Send("MULTI")

	c.Send("SPOP", configmanager.GetInstance().DeploymentEnv+key, count)
	c.Send("DEL", configmanager.GetInstance().DeploymentEnv+lockKey)

	reply, err := c.Do("EXEC")
	if err != nil {
		logger.Log().Debug("Error popping from set", logger.StringField("key", key), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_spop_transaction", configmanager.GetInstance().ServiceLBName)
		return []string{}
	}

	// Process the reply
	data, err := redis.Strings(reply.([]interface{})[0], nil)
	if err != nil {
		logger.Log().Debug("Error popping from set", logger.StringField("key", key), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_spop_transaction", configmanager.GetInstance().ServiceLBName)
		return []string{}
	}
	metrics.GetInstance().RecordCacheHit("redis_spop_transaction", configmanager.GetInstance().ServiceLBName)
	return data
}

func (cache *RedisCache) SAddTransaction(ctx context.Context, key string, values []interface{}, lockKey string) error {
	if len(values) < 1 {
		return errors.New("not enough parameters")
	}
	cache.semaphore.Acquire(ctx, 1)
	conn := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer conn.Close()

	conn.Send("MULTI")

	// Your update logic goes here
	conn.Send("SADD", append([]interface{}{configmanager.GetInstance().DeploymentEnv + key}, values...)...)

	// Release the lock after the update
	conn.Send("DEL", configmanager.GetInstance().DeploymentEnv+lockKey)

	// Execute the transaction
	_, err := conn.Do("EXEC")
	if err != nil {
		metrics.GetInstance().RecordCacheMiss("redis_sadd_transaction", configmanager.GetInstance().ServiceLBName)
		return err
	}
	metrics.GetInstance().RecordCacheHit("redis_sadd_transaction", configmanager.GetInstance().ServiceLBName)
	return err
}

func (cache *RedisCache) SMembers(ctx context.Context, key string) []string {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {

		return nil
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Strings(c.Do("SMEMBERS", configmanager.GetInstance().DeploymentEnv+key))
	if err != nil {
		logger.Log().Debug("Error getting members from set", logger.StringField("key", key), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_smembers", configmanager.GetInstance().ServiceLBName)
	} else {
		metrics.GetInstance().RecordCacheHit("redis_smembers", configmanager.GetInstance().ServiceLBName)
	}
	return data
}

func (cache *RedisCache) SRem(ctx context.Context, key string, value ...interface{}) error {
	if len(value) < 1 {
		return errors.New("Not enough parameters")
	}

	if err := cache.semaphore.Acquire(ctx, 1); err != nil {
		return err
	}
	defer cache.semaphore.Release(1)

	c := cache.pool.Get()
	defer c.Close()

	_, err := c.Do("SREM", append([]interface{}{configmanager.GetInstance().DeploymentEnv + key}, value...)...)
	if err != nil {
		logger.Log().Debug("Error removing from set", logger.AnyField("value", value), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_srem", configmanager.GetInstance().ServiceLBName)
		return err
	}
	metrics.GetInstance().RecordCacheHit("redis_srem", configmanager.GetInstance().ServiceLBName)

	return nil
}

func (cache *RedisCache) SCard(ctx context.Context, key string) (int64, error) {
	if key == "" {
		return 0, errors.New("Key cannot be empty")
	}

	if err := cache.semaphore.Acquire(ctx, 1); err != nil {
		return 0, err
	}
	defer cache.semaphore.Release(1)

	c := cache.pool.Get()
	defer c.Close()

	// Send the SCARD command to Redis
	length, err := redis.Int64(c.Do("SCARD", configmanager.GetInstance().DeploymentEnv+key))
	if err != nil {
		logger.Log().Debug("Error getting card from set", logger.StringField("key", key), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_scard", configmanager.GetInstance().ServiceLBName)
		return 0, err
	}
	metrics.GetInstance().RecordCacheHit("redis_scard", configmanager.GetInstance().ServiceLBName)

	return length, nil
}

func (cache *RedisCache) SetISMember(ctx context.Context, key string, member string) bool {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {

		return false
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	val, err := redis.Bool(c.Do("SISMEMBER", configmanager.GetInstance().DeploymentEnv+key, member))
	if err != nil {
		logger.Log().Debug("Error checking if member is in set", logger.StringField("key", key), logger.StringField("member", member), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_sismember", configmanager.GetInstance().ServiceLBName)
		return false
	}
	metrics.GetInstance().RecordCacheHit("redis_sismember", configmanager.GetInstance().ServiceLBName)
	return val
}

func (cache *RedisCache) SetMISMember(ctx context.Context, key string, values ...interface{}) []int64 {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {

		return nil
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()

	keys := []interface{}{configmanager.GetInstance().DeploymentEnv + key}
	keys = append(keys, values...)

	val, err := redis.Int64s(c.Do("SMISMEMBER", keys...))
	if err != nil {
		logger.Log().Debug("Error checking if members are in set", logger.StringField("key", key), logger.AnyField("values", values), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_smismember", configmanager.GetInstance().ServiceLBName)
		return nil
	}
	metrics.GetInstance().RecordCacheHit("redis_smismember", configmanager.GetInstance().ServiceLBName)
	return val
}

func (cache *RedisCache) Expire(ctx context.Context, key string, expirationTime int) error {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {

		return err
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	_, err := redis.Bool(c.Do("EXPIRE", configmanager.GetInstance().DeploymentEnv+key, expirationTime))
	if err != nil {
		logger.Log().Debug("Error expiring key", logger.StringField("key", key), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_expire", configmanager.GetInstance().ServiceLBName)
		return err
	}
	metrics.GetInstance().RecordCacheHit("redis_expire", configmanager.GetInstance().ServiceLBName)
	return nil
}

// TTL returns the remaining time to live of a key in seconds
// Returns -2 if the key does not exist, -1 if the key exists but has no expiry
func (cache *RedisCache) TTL(ctx context.Context, key string) int64 {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {
		return -2
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	ttl, err := redis.Int64(c.Do("TTL", configmanager.GetInstance().DeploymentEnv+key))
	if err != nil {
		logger.Log().Debug("Error getting TTL", logger.StringField("key", key), logger.ErrorField("error", err))
		return -2
	}
	return ttl
}

func (cache *RedisCache) HashMultiSet(ctx context.Context, key string, args map[string]interface{}) {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	_, err := c.Do("HMSET", redis.Args{configmanager.GetInstance().DeploymentEnv + key}.AddFlat(args)...)
	if err != nil {
		logger.Log().Debug("Error setting hash", logger.StringField("key", key), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_hmset", configmanager.GetInstance().ServiceLBName)
	} else {
		metrics.GetInstance().RecordCacheHit("redis_hmset", configmanager.GetInstance().ServiceLBName)
	}
}

func (cache *RedisCache) HashMultiSetString(ctx context.Context, key string, args map[string]string) {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	_, err := c.Do("HMSET", redis.Args{configmanager.GetInstance().DeploymentEnv + key}.AddFlat(args)...)
	if err != nil {
		logger.Log().Debug("Error setting hash", logger.StringField("key", key), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_hmset", configmanager.GetInstance().ServiceLBName)
	} else {
		metrics.GetInstance().RecordCacheHit("redis_hmset", configmanager.GetInstance().ServiceLBName)
	}
}

func (cache *RedisCache) HashMultiSetInt(ctx context.Context, key string, args map[string]int) {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	_, err := c.Do("HMSET", redis.Args{configmanager.GetInstance().DeploymentEnv + key}.AddFlat(args)...)
	if err != nil {
		logger.Log().Debug("Error setting hash", logger.StringField("key", key), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_hmset", configmanager.GetInstance().ServiceLBName)
	} else {
		metrics.GetInstance().RecordCacheHit("redis_hmset", configmanager.GetInstance().ServiceLBName)
	}
}

// HashSet first index should be key
func (cache *RedisCache) HashSet(ctx context.Context, key string, args []interface{}) {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	keys := []interface{}{configmanager.GetInstance().DeploymentEnv + key}
	keys = append(keys, args...)

	_, err := c.Do("HSET", keys...)
	if err != nil {
		logger.Log().Debug("Error setting hash", logger.StringField("key", key), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_hset", configmanager.GetInstance().ServiceLBName)
	} else {
		metrics.GetInstance().RecordCacheHit("redis_hset", configmanager.GetInstance().ServiceLBName)
	}
}

func (cache *RedisCache) LPush(ctx context.Context, key string, args []byte) {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {
		return
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	_, err := c.Do("LPUSH", configmanager.GetInstance().DeploymentEnv+key, args)
	if err != nil {
		logger.Log().Debug("Error pushing to list", logger.StringField("key", key), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_lpush", configmanager.GetInstance().ServiceLBName)
	} else {
		metrics.GetInstance().RecordCacheHit("redis_lpush", configmanager.GetInstance().ServiceLBName)
	}
}

func (cache *RedisCache) LTrim(ctx context.Context, key string, start int, end int) {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {
		return
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	_, err := c.Do("LTRIM", configmanager.GetInstance().DeploymentEnv+key, start, end)
	if err != nil {
		logger.Log().Debug("Error trimming list", logger.StringField("key", key), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_ltrim", configmanager.GetInstance().ServiceLBName)
	} else {
		metrics.GetInstance().RecordCacheHit("redis_ltrim", configmanager.GetInstance().ServiceLBName)
	}
}

func (cache *RedisCache) LRange(ctx context.Context, key string, start int, end int) []string {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {
		return nil
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Strings(c.Do("LRANGE", configmanager.GetInstance().DeploymentEnv+key, start, end))
	if err != nil {
		logger.Log().Debug("Error getting list range", logger.StringField("key", key), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_lrange", configmanager.GetInstance().ServiceLBName)
	} else {
		metrics.GetInstance().RecordCacheHit("redis_lrange", configmanager.GetInstance().ServiceLBName)
	}
	return data
}

func (cache *RedisCache) HashGet(ctx context.Context, key string, field string) string {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.String(c.Do("HGET", configmanager.GetInstance().DeploymentEnv+key, field))
	if err != nil {
		logger.Log().Debug("Error getting hash", logger.StringField("key", key), logger.StringField("field", field), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_hget", configmanager.GetInstance().ServiceLBName)
	} else {
		metrics.GetInstance().RecordCacheHit("redis_hget", configmanager.GetInstance().ServiceLBName)
	}
	return data
}

func (cache *RedisCache) HashGetNoPrint(ctx context.Context, key string, field string) (string, error) {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.String(c.Do("HGET", configmanager.GetInstance().DeploymentEnv+key, field))
	if err != nil {
		logger.Log().Debug("Error getting hash", logger.StringField("key", key), logger.StringField("field", field), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_hget", configmanager.GetInstance().ServiceLBName)
	} else {
		metrics.GetInstance().RecordCacheHit("redis_hget", configmanager.GetInstance().ServiceLBName)
	}
	return data, err
}

func (cache *RedisCache) HashGetBytes(ctx context.Context, key string, field string) []byte {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Bytes(c.Do("HGET", configmanager.GetInstance().DeploymentEnv+key, field))
	if err != nil {
		logger.Log().Debug("Error getting hash", logger.StringField("key", key), logger.StringField("field", field), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_hget", configmanager.GetInstance().ServiceLBName)
	} else {
		metrics.GetInstance().RecordCacheHit("redis_hget", configmanager.GetInstance().ServiceLBName)
	}
	return data
}
func (cache *RedisCache) HashDel(ctx context.Context, key string, field string) error {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	_, err := c.Do("HDEL", configmanager.GetInstance().DeploymentEnv+key, field)
	if err != nil {
		logger.Log().Debug("Error deleting hash", logger.StringField("key", key), logger.StringField("field", field), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_hdel", configmanager.GetInstance().ServiceLBName)
	} else {
		metrics.GetInstance().RecordCacheHit("redis_hdel", configmanager.GetInstance().ServiceLBName)
	}
	return err
}

func (cache *RedisCache) HashLen(ctx context.Context, key string) (int64, error) {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	length, err := redis.Int64(c.Do("HLEN", configmanager.GetInstance().DeploymentEnv+key))
	if err != nil {
		logger.Log().Debug("Error getting hash length", logger.StringField("key", key), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_hlen", configmanager.GetInstance().ServiceLBName)
	} else {
		metrics.GetInstance().RecordCacheHit("redis_hlen", configmanager.GetInstance().ServiceLBName)
	}
	return length, err
}

func (cache *RedisCache) HashGetAll(ctx context.Context, key string) map[string]string {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.StringMap(c.Do("HGETALL", configmanager.GetInstance().DeploymentEnv+key))
	if err != nil {
		logger.Log().Debug("Error getting hash all", logger.StringField("key", key), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_hgetall", configmanager.GetInstance().ServiceLBName)
	} else {
		metrics.GetInstance().RecordCacheHit("redis_hgetall", configmanager.GetInstance().ServiceLBName)
	}
	return data
}

func (cache *RedisCache) HashGetInt(ctx context.Context, key string, field string) int {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Int(c.Do("HGET", configmanager.GetInstance().DeploymentEnv+key, field))
	if err != nil {
		logger.Log().Debug("Error getting hash int", logger.StringField("key", key), logger.StringField("field", field), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_hget", configmanager.GetInstance().ServiceLBName)
	} else {
		metrics.GetInstance().RecordCacheHit("redis_hget", configmanager.GetInstance().ServiceLBName)
	}
	return data
}

func (cache *RedisCache) HashGetBool(ctx context.Context, key string, field string) bool {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Bool(c.Do("HGET", configmanager.GetInstance().DeploymentEnv+key, field))
	if err != nil {
		logger.Log().Debug("Error getting hash bool", logger.StringField("key", key), logger.StringField("field", field), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_hget", configmanager.GetInstance().ServiceLBName)
	} else {
		metrics.GetInstance().RecordCacheHit("redis_hget", configmanager.GetInstance().ServiceLBName)
	}
	return data
}

func (cache *RedisCache) HashGetInt64(ctx context.Context, key string, field string) int64 {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Int64(c.Do("HGET", configmanager.GetInstance().DeploymentEnv+key, field))
	if err != nil {
		logger.Log().Debug("Error getting hash int64", logger.StringField("key", key), logger.StringField("field", field), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_hget", configmanager.GetInstance().ServiceLBName)
	} else {
		metrics.GetInstance().RecordCacheHit("redis_hget", configmanager.GetInstance().ServiceLBName)
	}
	return data
}

func (cache *RedisCache) HashGetFloat64(ctx context.Context, key string, field string) float64 {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Float64(c.Do("HGET", configmanager.GetInstance().DeploymentEnv+key, field))
	if err != nil {
		logger.Log().Debug("Error getting hash float64", logger.StringField("key", key), logger.StringField("field", field), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_hget", configmanager.GetInstance().ServiceLBName)
	} else {
		metrics.GetInstance().RecordCacheHit("redis_hget", configmanager.GetInstance().ServiceLBName)
	}
	return data
}

func (cache *RedisCache) HashGetInt64WithError(ctx context.Context, key string, field string) (int64, error) {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Int64(c.Do("HGET", configmanager.GetInstance().DeploymentEnv+key, field))
	if err != nil {
		logger.Log().Debug("Error getting hash int64", logger.StringField("key", key), logger.StringField("field", field), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_hget", configmanager.GetInstance().ServiceLBName)
	} else {
		metrics.GetInstance().RecordCacheHit("redis_hget", configmanager.GetInstance().ServiceLBName)
	}
	return data, err
}

func (cache *RedisCache) HashMGet(ctx context.Context, key string, field []interface{}) []string {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Strings(c.Do("HMGET", append([]interface{}{configmanager.GetInstance().DeploymentEnv + key}, field...)...))
	if err != nil {
		logger.Log().Debug("Error getting hash", logger.StringField("key", key), logger.AnyField("field", field), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_hmget", configmanager.GetInstance().ServiceLBName)
	} else {
		metrics.GetInstance().RecordCacheHit("redis_hmget", configmanager.GetInstance().ServiceLBName)
	}
	return data
}

func (cache *RedisCache) HashMGetInts(ctx context.Context, key string, field []interface{}) []int {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Ints(c.Do("HMGET", append([]interface{}{configmanager.GetInstance().DeploymentEnv + key}, field...)...))
	if err != nil {
		metrics.GetInstance().RecordCacheMiss("redis_hmget", configmanager.GetInstance().ServiceLBName)
	} else {
		metrics.GetInstance().RecordCacheHit("redis_hmget", configmanager.GetInstance().ServiceLBName)
	}
	return data
}

func (cache *RedisCache) HashIncrementBy(ctx context.Context, key string, field string, val int64) int64 {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	value, err := redis.Int64(c.Do("HINCRBY", configmanager.GetInstance().DeploymentEnv+key, field, val))
	if err != nil {
		metrics.GetInstance().RecordCacheMiss("redis_hincrby", configmanager.GetInstance().ServiceLBName)
	} else {
		metrics.GetInstance().RecordCacheHit("redis_hincrby", configmanager.GetInstance().ServiceLBName)
	}
	return value
}
func (cache *RedisCache) HashIncrementByFloat(ctx context.Context, key string, field string, val float64) int64 {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	value, err := redis.Int64(c.Do("HINCRBYFLOAT", configmanager.GetInstance().DeploymentEnv+key, field, val))
	if err != nil {
		metrics.GetInstance().RecordCacheMiss("redis_hincrbyfloat", configmanager.GetInstance().ServiceLBName)
	} else {
		metrics.GetInstance().RecordCacheHit("redis_hincrbyfloat", configmanager.GetInstance().ServiceLBName)
	}
	return value
}

func (cache *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {
		return false, err
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Bool(c.Do("EXISTS", configmanager.GetInstance().DeploymentEnv+key))
	if err != nil {
		metrics.GetInstance().RecordCacheMiss("redis_exists", configmanager.GetInstance().ServiceLBName)
	} else {
		metrics.GetInstance().RecordCacheHit("redis_exists", configmanager.GetInstance().ServiceLBName)
		return false, err
	}
	return data, nil
}

func (cache *RedisCache) ZAdd(ctx context.Context, key string, seq int64, value interface{}) error {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	_, err := c.Do("ZADD", configmanager.GetInstance().DeploymentEnv+key, seq, value)
	if err != nil {
		logger.Log().Debug("Error adding to  set", logger.StringField("key", key), logger.ErrorField("error", err))
		metrics.GetInstance().RecordCacheMiss("redis_zadd", configmanager.GetInstance().ServiceLBName)
	} else {
		metrics.GetInstance().RecordCacheHit("redis_zadd", configmanager.GetInstance().ServiceLBName)
	}
	return err
}
func (cache *RedisCache) ZRange(ctx context.Context, key string, start int, end int) []string {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Strings(c.Do("ZRANGE", configmanager.GetInstance().DeploymentEnv+key, start, end))
	if err != nil {
		metrics.GetInstance().RecordCacheMiss("redis_zrange", configmanager.GetInstance().ServiceLBName)
	}
	return data
}
func (cache *RedisCache) ZRangeWithScore(ctx context.Context, key string, start int, end int) []string {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Strings(c.Do("ZRANGEBYSCORE", configmanager.GetInstance().DeploymentEnv+key, start, end))
	if err != nil {
		metrics.GetInstance().RecordCacheMiss("redis_zrange_with_score", configmanager.GetInstance().ServiceLBName)
	}
	return data

}
func (cache *RedisCache) ZRevRange(ctx context.Context, key string, start int, end int) []string {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Strings(c.Do("ZREVRANGE", configmanager.GetInstance().DeploymentEnv+key, start, end))
	if err != nil {
		metrics.GetInstance().RecordCacheMiss("redis_zrevrange", configmanager.GetInstance().ServiceLBName)
	}
	return data
}
func (cache *RedisCache) ZRevRangeWithScore(ctx context.Context, key string, start int, end int) []string {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {

	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Strings(c.Do("ZREVRANGEBYSCORE", configmanager.GetInstance().DeploymentEnv+key, start, end))
	if err != nil {
		metrics.GetInstance().RecordCacheMiss("redis_zrevrange_with_score", configmanager.GetInstance().ServiceLBName)
	}
	return data
}
func (cache *RedisCache) ZRem(ctx context.Context, key string, value ...interface{}) error {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	_, err := c.Do("ZREM", configmanager.GetInstance().DeploymentEnv+key, value)
	if err != nil {
		metrics.GetInstance().RecordCacheMiss("redis_zrem", configmanager.GetInstance().ServiceLBName)
	}
	return err
}
func (cache *RedisCache) ZRemRangeByScore(ctx context.Context, key string, start int, end int) error {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	_, err := c.Do("ZREMRANGEBYSCORE", configmanager.GetInstance().DeploymentEnv+key, start, end)
	if err != nil {
		metrics.GetInstance().RecordCacheMiss("redis_zremrangebyscore", configmanager.GetInstance().ServiceLBName)
	}
	return err
}
func (cache *RedisCache) ZCount(ctx context.Context, key string, start int, end int) (int64, error) {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Int64(c.Do("ZCOUNT", configmanager.GetInstance().DeploymentEnv+key, start, end))
	if err != nil {
		metrics.GetInstance().RecordCacheMiss("redis_zcount", configmanager.GetInstance().ServiceLBName)
	}
	return data, err
}
func (cache *RedisCache) ZCard(ctx context.Context, key string) (int64, error) {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Int64(c.Do("ZCARD", configmanager.GetInstance().DeploymentEnv+key))
	if err != nil {
		metrics.GetInstance().RecordCacheMiss("redis_zcard", configmanager.GetInstance().ServiceLBName)
	}
	return data, err
}
func (cache *RedisCache) ZScore(ctx context.Context, key string, member string) (float64, error) {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Float64(c.Do("ZSCORE", configmanager.GetInstance().DeploymentEnv+key, member))
	if err != nil {
		metrics.GetInstance().RecordCacheMiss("redis_zscore", configmanager.GetInstance().ServiceLBName)
	}
	return data, err
}
func (cache *RedisCache) ZRangeByScoreWithLimit(ctx context.Context, key string, start int64, end int64, limit int) []string {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Strings(c.Do("ZRANGEBYSCORE", configmanager.GetInstance().DeploymentEnv+key, start, end, "LIMIT", 0, limit))
	if err != nil {
		metrics.GetInstance().RecordCacheMiss("redis_zrangebyscore_with_limit", configmanager.GetInstance().ServiceLBName)
	}
	return data
}
func (cache *RedisCache) ZRevRangeByScoreWithLimit(ctx context.Context, key string, start int64, end int64, limit int) []string {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()
	data, err := redis.Strings(c.Do("ZREVRANGEBYSCORE", configmanager.GetInstance().DeploymentEnv+key, start, end, "LIMIT", 0, limit))
	if err != nil {
		metrics.GetInstance().RecordCacheMiss("redis_zrevrangebyscore_with_limit", configmanager.GetInstance().ServiceLBName)
	}
	return data
}
func (cache *RedisCache) ZPaginate(ctx context.Context, key string, page int, limit int) ([]string, error) {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	start := (page - 1) * limit
	end := page * limit
	defer c.Close()
	data, err := redis.Strings(c.Do("ZRANGEBYSCORE", configmanager.GetInstance().DeploymentEnv+key, start, end, "LIMIT", 0, limit))
	if err != nil {
		metrics.GetInstance().RecordCacheMiss("redis_zrangebyscore_with_limit", configmanager.GetInstance().ServiceLBName)
	}
	return data, err
}

func (cache *RedisCache) ZMultiAdd(ctx context.Context, key string, items []cachemodels.ZMultiAddItem) error {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()

	fullKey := configmanager.GetInstance().DeploymentEnv + key
	args := make([]interface{}, 0, len(items)*2+1)
	args = append(args, fullKey)
	for _, item := range items {
		args = append(args, item.Seq, item.Value)
	}

	_, err := c.Do("ZADD", args...)
	if err != nil {
		metrics.GetInstance().RecordCacheMiss("redis_zmultiadd", configmanager.GetInstance().ServiceLBName)
	}
	return err
}

func (cache *RedisCache) ZUnion(ctx context.Context, key string, keys ...string) []string {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()

	fullKey := configmanager.GetInstance().DeploymentEnv + key
	args := make([]interface{}, 0, len(keys)+1)
	args = append(args, fullKey)
	args = append(args, len(keys))
	for _, key := range keys {
		args = append(args, configmanager.GetInstance().DeploymentEnv+key)
	}

	data, err := redis.Strings(c.Do("ZUNION", args...))
	if err != nil {
		metrics.GetInstance().RecordCacheMiss("redis_zunion", configmanager.GetInstance().ServiceLBName)
	}
	return data
}

func (cache *RedisCache) ZUnionWithExpire(ctx context.Context, key string, expireTime int, start int64, end int64, limit int, keys ...string) []string {
	if err := cache.semaphore.Acquire(ctx, 1); err != nil {
	}
	c := cache.pool.Get()
	defer cache.semaphore.Release(1)
	defer c.Close()

	fullKey := configmanager.GetInstance().DeploymentEnv + key
	args := make([]interface{}, 0, len(keys)+1)
	args = append(args, fullKey)
	args = append(args, len(keys))
	for _, key := range keys {
		args = append(args, configmanager.GetInstance().DeploymentEnv+key)
	}

	data, err := redis.Strings(c.Do("ZUNIONSTORE", args...))
	if expireTime > 0 {
		c.Do("EXPIRE", configmanager.GetInstance().DeploymentEnv+key, expireTime)
	}

	data = cache.ZRangeByScoreWithLimit(ctx, key, start, end, limit)

	if err != nil {
		metrics.GetInstance().RecordCacheMiss("redis_zunion", configmanager.GetInstance().ServiceLBName)
	}
	return data
}
