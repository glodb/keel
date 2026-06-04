package cache

import (
	"context"
	"time"

	"github.com/glodb/keel/models/cachemodels"

	"github.com/gomodule/redigo/redis"
)

// Cache interface defines all cache operations
type Cache interface {
	// Connection management
	GetConnection(ctx context.Context) redis.Conn
	ReleaseConnection(conn redis.Conn)

	GetContext() (context.Context, context.CancelFunc)
	GetCacheContext() context.Context
	GetCacheContextWithDeadline(deadline time.Time) (context.Context, context.CancelFunc)
	GetCacheContextWithValue(key, value interface{}) (context.Context, context.CancelFunc)
	GetContextRemainingTime(ctx context.Context) (time.Duration, bool)

	// Basic operations
	Set(ctx context.Context, key string, value []byte) error
	Get(ctx context.Context, key string) ([]byte, error)
	Del(ctx context.Context, key string) error
	DelMany(ctx context.Context, keys []string) error
	Exists(ctx context.Context, key string) (bool, error)

	// Set with expiration
	SetEx(ctx context.Context, key string, value []byte, expiryTime int) error

	// Integer operations
	SetInt(ctx context.Context, key string, value int) error
	GetInt(ctx context.Context, key string) (int64, error)
	Increment(ctx context.Context, key string) (int64, error)
	Decrement(ctx context.Context, key string) (int64, error)

	// String operations
	SetString(ctx context.Context, key string, value string) error
	GetString(ctx context.Context, key string) (string, error)
	Append(ctx context.Context, key string, value interface{}) error

	// Lock operations
	AcquireCacheLock(ctx context.Context, lockKey string, expirationMilli int64) (bool, error)
	AcquireCacheLock2(lockKey string, expirationMilli int64) (bool, error)
	ReleaseLock(ctx context.Context, lockKey string) error

	// Set operations
	SAdd(ctx context.Context, key string, values ...interface{}) error
	SPop(ctx context.Context, key string, count int) []string
	SPopTransaction(ctx context.Context, key string, count int, lockKey string) []string
	SAddTransaction(ctx context.Context, key string, values []interface{}, lockKey string) error
	SMembers(ctx context.Context, key string) []string
	SRem(ctx context.Context, key string, value ...interface{}) error
	SCard(ctx context.Context, key string) (int64, error)
	SetISMember(ctx context.Context, key string, member string) bool
	SetMISMember(ctx context.Context, key string, values ...interface{}) []int64
	SUnion(ctx context.Context, keys ...interface{}) []string

	// Expiration
	Expire(ctx context.Context, key string, expirationTime int) error
	TTL(ctx context.Context, key string) int64

	// ZSet operations
	ZAdd(ctx context.Context, key string, seq int64, value interface{}) error
	ZMultiAdd(ctx context.Context, key string, items []cachemodels.ZMultiAddItem) error
	ZRange(ctx context.Context, key string, start int, end int) []string
	ZRangeWithScore(ctx context.Context, key string, start int, end int) []string
	ZRevRange(ctx context.Context, key string, start int, end int) []string
	ZRevRangeWithScore(ctx context.Context, key string, start int, end int) []string
	ZRem(ctx context.Context, key string, value ...interface{}) error
	ZRemRangeByScore(ctx context.Context, key string, start int, end int) error
	ZCount(ctx context.Context, key string, start int, end int) (int64, error)
	ZCard(ctx context.Context, key string) (int64, error)
	ZScore(ctx context.Context, key string, member string) (float64, error)
	ZRangeByScoreWithLimit(ctx context.Context, key string, start int64, end int64, limit int) []string
	ZRevRangeByScoreWithLimit(ctx context.Context, key string, start int64, end int64, limit int) []string
	ZPaginate(ctx context.Context, key string, page int, limit int) ([]string, error)
	ZUnion(ctx context.Context, key string, keys ...string) []string
	ZUnionWithExpire(ctx context.Context, key string, expireTime int, start int64, end int64, limit int, keys ...string) []string

	// Hash operations
	HashMultiSet(ctx context.Context, key string, args map[string]interface{})
	HashMultiSetString(ctx context.Context, key string, args map[string]string)
	HashMultiSetInt(ctx context.Context, key string, args map[string]int)
	HashSet(ctx context.Context, key string, args []interface{})
	HashGet(ctx context.Context, key string, field string) string
	HashGetNoPrint(ctx context.Context, key string, field string) (string, error)
	HashGetBytes(ctx context.Context, key string, field string) []byte
	HashDel(ctx context.Context, key string, field string) error
	HashLen(ctx context.Context, key string) (int64, error)
	HashGetAll(ctx context.Context, key string) map[string]string
	HashGetInt(ctx context.Context, key string, field string) int
	HashGetBool(ctx context.Context, key string, field string) bool
	HashGetInt64(ctx context.Context, key string, field string) int64
	HashGetFloat64(ctx context.Context, key string, field string) float64
	HashGetInt64WithError(ctx context.Context, key string, field string) (int64, error)
	HashMGet(ctx context.Context, key string, fields []interface{}) []string
	HashMGetInts(ctx context.Context, key string, field []interface{}) []int
	HashIncrementBy(ctx context.Context, key string, field string, val int64) int64
	HashIncrementByFloat(ctx context.Context, key string, field string, val float64) int64

	// List operations
	LPush(ctx context.Context, key string, args []byte)
	LTrim(ctx context.Context, key string, start int, end int)
	LRange(ctx context.Context, key string, start int, end int) []string

	// Utility operations
	GetKeys(ctx context.Context, pattern string) []string
}
