package rediscache

import (
	"context"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock/v3"
	"golang.org/x/sync/semaphore"
)

// Mock Redis connection for testing
func setupMockRedis() *redigomock.Conn {
	return redigomock.NewConn()
}

func TestRedisCache_Set(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		value          []byte
		setupMock      func(*redigomock.Conn)
		expectError    bool
		contextTimeout time.Duration
	}{
		{
			name:  "Successful SET operation",
			key:   "test-key",
			value: []byte("test-value"),
			setupMock: func(conn *redigomock.Conn) {
				conn.Command("SET", "DEVtest-key", []byte("test-value")).Expect("OK")
			},
			expectError:    false,
			contextTimeout: 5 * time.Second,
		},
		{
			name:  "SET operation with Redis error",
			key:   "test-key",
			value: []byte("test-value"),
			setupMock: func(conn *redigomock.Conn) {
				conn.Command("SET", "DEVtest-key", []byte("test-value")).ExpectError(redis.Error("connection failed"))
			},
			expectError:    true,
			contextTimeout: 5 * time.Second,
		},
		{
			name:  "Context timeout",
			key:   "test-key",
			value: []byte("test-value"),
			setupMock: func(conn *redigomock.Conn) {
				// Don't set up any expectations - this will cause the operation to hang
			},
			expectError:    true,
			contextTimeout: 1 * time.Millisecond, // Very short timeout
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := setupMockRedis()
			if tt.setupMock != nil {
				tt.setupMock(mockConn)
			}

			// Create a test cache with mock connection
			cache := &RedisCache{
				pool: &redis.Pool{
					Dial: func() (redis.Conn, error) {
						return mockConn, nil
					},
				},
				semaphore: testSemaphore(1),
			}

			ctx, cancel := context.WithTimeout(context.Background(), tt.contextTimeout)
			defer cancel()

			err := cache.Set(ctx, tt.key, tt.value)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}

			// Verify all expected commands were called
			if err := mockConn.ExpectationsWereMet(); err != nil && !tt.expectError {
				t.Errorf("Mock expectations not met: %v", err)
			}
		})
	}
}

func TestRedisCache_Get(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		setupMock      func(*redigomock.Conn)
		expectedValue  []byte
		expectError    bool
		contextTimeout time.Duration
	}{
		{
			name: "Successful GET operation",
			key:  "test-key",
			setupMock: func(conn *redigomock.Conn) {
				conn.Command("GET", "DEVtest-key").Expect([]byte("test-value"))
			},
			expectedValue:  []byte("test-value"),
			expectError:    false,
			contextTimeout: 5 * time.Second,
		},
		{
			name: "GET operation - key not found",
			key:  "missing-key",
			setupMock: func(conn *redigomock.Conn) {
				conn.Command("GET", "DEVmissing-key").Expect(nil)
			},
			expectedValue:  nil,
			expectError:    false,
			contextTimeout: 5 * time.Second,
		},
		{
			name: "GET operation with Redis error",
			key:  "test-key",
			setupMock: func(conn *redigomock.Conn) {
				conn.Command("GET", "DEVtest-key").ExpectError(redis.Error("connection failed"))
			},
			expectedValue:  []byte{},
			expectError:    true,
			contextTimeout: 5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := setupMockRedis()
			if tt.setupMock != nil {
				tt.setupMock(mockConn)
			}

			cache := &RedisCache{
				pool: &redis.Pool{
					Dial: func() (redis.Conn, error) {
						return mockConn, nil
					},
				},
				semaphore: testSemaphore(1),
			}

			ctx, cancel := context.WithTimeout(context.Background(), tt.contextTimeout)
			defer cancel()

			result, err := cache.Get(ctx, tt.key)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}

				if string(result) != string(tt.expectedValue) {
					t.Errorf("Expected value: %s, got: %s", string(tt.expectedValue), string(result))
				}
			}

			if err := mockConn.ExpectationsWereMet(); err != nil && !tt.expectError {
				t.Errorf("Mock expectations not met: %v", err)
			}
		})
	}
}

func TestRedisCache_AcquireCacheLock(t *testing.T) {
	tests := []struct {
		name           string
		lockKey        string
		expirationMs   int64
		setupMock      func(*redigomock.Conn)
		expectedResult bool
		expectError    bool
	}{
		{
			name:         "Successful lock acquisition",
			lockKey:      "test-lock",
			expirationMs: 5000,
			setupMock: func(conn *redigomock.Conn) {
				conn.Command("SET", "DEVtest-lock", "1", "NX", "PX", 5000).Expect("OK")
			},
			expectedResult: true,
			expectError:    false,
		},
		{
			name:         "Lock already exists",
			lockKey:      "existing-lock",
			expirationMs: 5000,
			setupMock: func(conn *redigomock.Conn) {
				conn.Command("SET", "DEVexisting-lock", "1", "NX", "PX", 5000).Expect(nil)
			},
			expectedResult: false,
			expectError:    false,
		},
		{
			name:         "Redis error during lock acquisition",
			lockKey:      "error-lock",
			expirationMs: 5000,
			setupMock: func(conn *redigomock.Conn) {
				conn.Command("SET", "DEVerror-lock", "1", "NX", "PX", 5000).ExpectError(redis.Error("connection failed"))
			},
			expectedResult: false,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := setupMockRedis()
			if tt.setupMock != nil {
				tt.setupMock(mockConn)
			}

			cache := &RedisCache{
				pool: &redis.Pool{
					Dial: func() (redis.Conn, error) {
						return mockConn, nil
					},
				},
				semaphore: testSemaphore(1),
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			result, err := cache.AcquireCacheLock(ctx, tt.lockKey, tt.expirationMs)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}

				if result != tt.expectedResult {
					t.Errorf("Expected result: %v, got: %v", tt.expectedResult, result)
				}
			}

			if err := mockConn.ExpectationsWereMet(); err != nil && !tt.expectError {
				t.Errorf("Mock expectations not met: %v", err)
			}
		})
	}
}

func TestRedisCache_SAdd(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		values      []interface{}
		setupMock   func(*redigomock.Conn)
		expectError bool
	}{
		{
			name:   "Successful SADD operation",
			key:    "test-set",
			values: []interface{}{"value1", "value2", "value3"},
			setupMock: func(conn *redigomock.Conn) {
				conn.Command("SADD", "DEVtest-set", "value1", "value2", "value3").Expect(int64(3))
			},
			expectError: false,
		},
		{
			name:        "Empty values array",
			key:         "test-set",
			values:      []interface{}{},
			setupMock:   nil, // No mock setup needed
			expectError: true,
		},
		{
			name:   "Redis error during SADD",
			key:    "error-set",
			values: []interface{}{"value1"},
			setupMock: func(conn *redigomock.Conn) {
				conn.Command("SADD", "DEVerror-set", "value1").ExpectError(redis.Error("connection failed"))
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := setupMockRedis()
			if tt.setupMock != nil {
				tt.setupMock(mockConn)
			}

			cache := &RedisCache{
				pool: &redis.Pool{
					Dial: func() (redis.Conn, error) {
						return mockConn, nil
					},
				},
				semaphore: testSemaphore(1),
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := cache.SAdd(ctx, tt.key, tt.values...)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}

			if tt.setupMock != nil {
				if err := mockConn.ExpectationsWereMet(); err != nil && !tt.expectError {
					t.Errorf("Mock expectations not met: %v", err)
				}
			}
		})
	}
}

func TestRedisCache_SCard(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		setupMock     func(*redigomock.Conn)
		expectedCount int64
		expectError   bool
	}{
		{
			name: "Successful SCARD operation",
			key:  "test-set",
			setupMock: func(conn *redigomock.Conn) {
				conn.Command("SCARD", "DEVtest-set").Expect(int64(5))
			},
			expectedCount: 5,
			expectError:   false,
		},
		{
			name:          "Empty key",
			key:           "",
			setupMock:     nil, // No mock setup needed
			expectedCount: 0,
			expectError:   true,
		},
		{
			name: "Redis error during SCARD",
			key:  "error-set",
			setupMock: func(conn *redigomock.Conn) {
				conn.Command("SCARD", "DEVerror-set").ExpectError(redis.Error("connection failed"))
			},
			expectedCount: 0,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := setupMockRedis()
			if tt.setupMock != nil {
				tt.setupMock(mockConn)
			}

			cache := &RedisCache{
				pool: &redis.Pool{
					Dial: func() (redis.Conn, error) {
						return mockConn, nil
					},
				},
				semaphore: testSemaphore(1),
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			count, err := cache.SCard(ctx, tt.key)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}

				if count != tt.expectedCount {
					t.Errorf("Expected count: %d, got: %d", tt.expectedCount, count)
				}
			}

			if tt.setupMock != nil {
				if err := mockConn.ExpectationsWereMet(); err != nil && !tt.expectError {
					t.Errorf("Mock expectations not met: %v", err)
				}
			}
		})
	}
}

// Integration test for cache operations
func TestRedisCache_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	mockConn := setupMockRedis()

	// Setup mock expectations for a sequence of operations
	mockConn.Command("SET", "DEVintegration-key", []byte("test-value")).Expect("OK")
	mockConn.Command("GET", "DEVintegration-key").Expect([]byte("test-value"))
	mockConn.Command("DEL", "DEVintegration-key").Expect(int64(1))

	cache := &RedisCache{
		pool: &redis.Pool{
			Dial: func() (redis.Conn, error) {
				return mockConn, nil
			},
		},
		semaphore: testSemaphore(1),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test SET operation
	err := cache.Set(ctx, "integration-key", []byte("test-value"))
	if err != nil {
		t.Fatalf("SET operation failed: %v", err)
	}

	// Test GET operation
	value, err := cache.Get(ctx, "integration-key")
	if err != nil {
		t.Fatalf("GET operation failed: %v", err)
	}
	if string(value) != "test-value" {
		t.Errorf("Expected value: test-value, got: %s", string(value))
	}

	// Test DEL operation
	err = cache.Del(ctx, "integration-key")
	if err != nil {
		t.Fatalf("DEL operation failed: %v", err)
	}

	// Verify all expectations were met
	if err := mockConn.ExpectationsWereMet(); err != nil {
		t.Errorf("Mock expectations not met: %v", err)
	}
}

// Benchmark tests
func BenchmarkRedisCache_Set(b *testing.B) {
	mockConn := setupMockRedis()
	mockConn.GenericCommand("SET").Expect("OK")

	cache := &RedisCache{
		pool: &redis.Pool{
			Dial: func() (redis.Conn, error) {
				return mockConn, nil
			},
		},
		semaphore: testSemaphore(100), // Higher limit for benchmarks
	}

	ctx := context.Background()
	value := []byte("benchmark-value")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = cache.Set(ctx, "benchmark-key", value)
		}
	})
}

func BenchmarkRedisCache_Get(b *testing.B) {
	mockConn := setupMockRedis()
	mockConn.GenericCommand("GET").Expect([]byte("benchmark-value"))

	cache := &RedisCache{
		pool: &redis.Pool{
			Dial: func() (redis.Conn, error) {
				return mockConn, nil
			},
		},
		semaphore: testSemaphore(100),
	}

	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = cache.Get(ctx, "benchmark-key")
		}
	})
}

// Helper function to create test semaphore
func testSemaphore(limit int64) *semaphore.Weighted {
	return semaphore.NewWeighted(limit)
}

// Test for semaphore behavior under high concurrency
func TestRedisCache_ConcurrentAccess(t *testing.T) {
	mockConn := setupMockRedis()

	// Allow multiple SET operations
	for i := 0; i < 10; i++ {
		mockConn.GenericCommand("SET").Expect("OK")
	}

	cache := &RedisCache{
		pool: &redis.Pool{
			Dial: func() (redis.Conn, error) {
				return mockConn, nil
			},
		},
		semaphore: testSemaphore(3), // Limited semaphore for testing
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start multiple goroutines
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			err := cache.Set(ctx, "concurrent-key", []byte("value"))
			if err != nil {
				t.Errorf("Goroutine %d failed: %v", id, err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		select {
		case <-done:
			// Success
		case <-time.After(10 * time.Second):
			t.Fatal("Timeout waiting for concurrent operations")
		}
	}
}
