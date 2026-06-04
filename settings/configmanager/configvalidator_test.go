package configmanager

import (
	"strings"
	"testing"
)

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *config
		expectError bool
		errorField  string
		errorMsg    string
	}{
		{
			name: "Valid configuration",
			config: &config{
				ClassName:       "SSOSERVICE",
				Address:         "localhost:8080",
				DeploymentEnv:   "DEV",
				TokenExpiry:     3600,
				PageSize:        50,
				MaxPageSize:     1000,
				SessionKey:      "test-session-key-16chars",
				SessionSecret:   "test-session-secret-16chars",
				OtpExpirySeconds: 300,
				OtpResendSeconds: 60,
				RPCRequestExpirySeconds: 30,
				CacheContextTimeout: 5,
				Redis: struct {
					RedisMaxConnections     int    `json:"redisMaxConnections"`
					RedisMaxIdleConnections int    `json:"redisMaxIdleConnections"`
					RedisCon                string `json:"redisCon"`
					RedisAddress            string `json:"redisAddress"`
					PrintRedis              bool   `json:"printRedis"`
				}{
					RedisMaxConnections:     100,
					RedisMaxIdleConnections: 10,
				},
				PublisherBatchSize: 100,
				DBBatchSize:        50,
			},
			expectError: false,
		},
		{
			name: "Invalid service name",
			config: &config{
				ClassName:     "INVALIDSERVICE",
				Address:       "localhost:8080",
				DeploymentEnv: "DEV",
			},
			expectError: true,
			errorField:  "ClassName",
			errorMsg:    "must be one of",
		},
		{
			name: "Empty required fields",
			config: &config{
				ClassName:     "",
				Address:       "",
				DeploymentEnv: "",
			},
			expectError: true,
			errorField:  "ClassName",
			errorMsg:    "service name is required",
		},
		{
			name: "Invalid address format",
			config: &config{
				ClassName:     "SSOSERVICE",
				Address:       "invalid-address",
				DeploymentEnv: "DEV",
			},
			expectError: true,
			errorField:  "Address",
			errorMsg:    "must be in format",
		},
		{
			name: "Invalid deployment environment",
			config: &config{
				ClassName:     "SSOSERVICE",
				Address:       "localhost:8080",
				DeploymentEnv: "INVALID",
			},
			expectError: true,
			errorField:  "DeploymentEnv",
			errorMsg:    "must be one of",
		},
		{
			name: "Token expiry too short",
			config: &config{
				ClassName:     "SSOSERVICE",
				Address:       "localhost:8080",
				DeploymentEnv: "DEV",
				TokenExpiry:   100, // Less than 300 seconds
			},
			expectError: true,
			errorField:  "TokenExpiry",
			errorMsg:    "must be between 5 minutes",
		},
		{
			name: "Token expiry too long",
			config: &config{
				ClassName:     "SSOSERVICE",
				Address:       "localhost:8080",
				DeploymentEnv: "DEV",
				TokenExpiry:   100000, // More than 86400 seconds
			},
			expectError: true,
			errorField:  "TokenExpiry",
			errorMsg:    "must be between",
		},
		{
			name: "Session key too short",
			config: &config{
				ClassName:     "SSOSERVICE",
				Address:       "localhost:8080",
				DeploymentEnv: "DEV",
				TokenExpiry:   3600,
				SessionKey:    "short", // Less than 16 characters
			},
			expectError: true,
			errorField:  "SessionKey",
			errorMsg:    "must be at least 16 characters",
		},
		{
			name: "Invalid Redis configuration",
			config: &config{
				ClassName:     "SSOSERVICE",
				Address:       "localhost:8080",
				DeploymentEnv: "DEV",
				Redis: struct {
					RedisMaxConnections     int    `json:"redisMaxConnections"`
					RedisMaxIdleConnections int    `json:"redisMaxIdleConnections"`
					RedisCon                string `json:"redisCon"`
					RedisAddress            string `json:"redisAddress"`
					PrintRedis              bool   `json:"printRedis"`
				}{
					RedisMaxConnections:     -1, // Invalid
					RedisMaxIdleConnections: 10,
				},
			},
			expectError: true,
			errorField:  "Redis.RedisMaxConnections",
			errorMsg:    "must be positive",
		},
		{
			name: "Redis idle connections exceed max connections",
			config: &config{
				ClassName:     "SSOSERVICE",
				Address:       "localhost:8080",
				DeploymentEnv: "DEV",
				Redis: struct {
					RedisMaxConnections     int    `json:"redisMaxConnections"`
					RedisMaxIdleConnections int    `json:"redisMaxIdleConnections"`
					RedisCon                string `json:"redisCon"`
					RedisAddress            string `json:"redisAddress"`
					PrintRedis              bool   `json:"printRedis"`
				}{
					RedisMaxConnections:     10,
					RedisMaxIdleConnections: 20, // More than max
				},
			},
			expectError: true,
			errorField:  "Redis.RedisMaxIdleConnections",
			errorMsg:    "cannot exceed max connections",
		},
		{
			name: "Page size larger than max page size",
			config: &config{
				ClassName:     "SSOSERVICE",
				Address:       "localhost:8080",  
				DeploymentEnv: "DEV",
				PageSize:      1000,
				MaxPageSize:   500, // Less than page size
			},
			expectError: true,
			errorField:  "PageSize",
			errorMsg:    "cannot be larger than max page size",
		},
		{
			name: "OTP expiry too short",
			config: &config{
				ClassName:        "SSOSERVICE",
				Address:          "localhost:8080",
				DeploymentEnv:    "DEV",
				OtpExpirySeconds: 10, // Less than 30 seconds
			},
			expectError: true,
			errorField:  "OtpExpirySeconds",
			errorMsg:    "must be between 30 seconds",
		},
		{
			name: "OTP resend time invalid",
			config: &config{
				ClassName:        "SSOSERVICE",
				Address:          "localhost:8080",
				DeploymentEnv:    "DEV",
				OtpExpirySeconds: 300,
				OtpResendSeconds: 400, // More than expiry
			},
			expectError: true,
			errorField:  "OtpResendSeconds",
			errorMsg:    "must be positive and less than OTP expiry time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validateConfig()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				// Check if error contains expected field and message
				errStr := err.Error()
				if tt.errorField != "" && !strings.Contains(errStr, tt.errorField) {
					t.Errorf("Expected error to contain field '%s', got: %s", tt.errorField, errStr)
				}
				if tt.errorMsg != "" && !strings.Contains(errStr, tt.errorMsg) {
					t.Errorf("Expected error to contain message '%s', got: %s", tt.errorMsg, errStr)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestValidateHostnamePort(t *testing.T) {
	tests := []struct {
		name        string
		address     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid localhost",
			address:     "localhost:8080",
			expectError: false,
		},
		{
			name:        "Valid IP address",
			address:     "192.168.1.1:3000",
			expectError: false,
		},
		{
			name:        "Valid domain",
			address:     "api.example.com:443",
			expectError: false,
		},
		{
			name:        "Missing port",
			address:     "localhost",
			expectError: true,
			errorMsg:    "must be in format 'hostname:port'",
		},
		{
			name:        "Empty hostname",
			address:     ":8080",
			expectError: true,
			errorMsg:    "hostname cannot be empty",
		},
		{
			name:        "Invalid port - non-numeric",
			address:     "localhost:abc",
			expectError: true,
			errorMsg:    "port must be a valid integer",
		},
		{
			name:        "Invalid port - too low",
			address:     "localhost:0",
			expectError: true,
			errorMsg:    "port must be between 1 and 65535",
		},
		{
			name:        "Invalid port - too high",
			address:     "localhost:99999",
			expectError: true,
			errorMsg:    "port must be between 1 and 65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHostnamePort(tt.address)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got: %s", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestConfigValidationErrors_Error(t *testing.T) {
	tests := []struct {
		name     string
		errors   ConfigValidationErrors
		expected string
	}{
		{
			name:     "No errors",
			errors:   ConfigValidationErrors{},
			expected: "no configuration validation errors",
		},
		{
			name: "Single error",
			errors: ConfigValidationErrors{
				{Field: "TokenExpiry", Value: 100, Message: "too short"},
			},
			expected: "config validation failed for 'TokenExpiry': too short (value: 100)",
		},
		{
			name: "Multiple errors",
			errors: ConfigValidationErrors{
				{Field: "TokenExpiry", Value: 100, Message: "too short"},
				{Field: "PageSize", Value: -1, Message: "must be positive"},
			},
			expected: "config validation failed for 'TokenExpiry': too short (value: 100); config validation failed for 'PageSize': must be positive (value: -1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.errors.Error()
			if result != tt.expected {
				t.Errorf("Expected: %s, got: %s", tt.expected, result)
			}
		})
	}
}

func TestGetCacheContextTimeout(t *testing.T) {
	tests := []struct {
		name     string
		timeout  int64
		expected string // Duration as string for easy comparison
	}{
		{
			name:     "Valid timeout",
			timeout:  10,
			expected: "10s",
		},
		{
			name:     "Zero timeout - should return default",
			timeout:  0,
			expected: "5s",
		},
		{
			name:     "Negative timeout - should return default",
			timeout:  -5,
			expected: "5s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &config{CacheContextTimeout: tt.timeout}
			result := config.GetCacheContextTimeout()
			
			if result.String() != tt.expected {
				t.Errorf("Expected timeout: %s, got: %s", tt.expected, result.String())
			}
		})
	}
}

// Benchmark tests for validation performance
func BenchmarkConfigValidation(b *testing.B) {
	config := &config{
		ClassName:       "SSOSERVICE",
		Address:         "localhost:8080",
		DeploymentEnv:   "DEV",
		TokenExpiry:     3600,
		PageSize:        50,
		MaxPageSize:     1000,
		SessionKey:      "test-session-key-16chars",
		SessionSecret:   "test-session-secret-16chars",
		OtpExpirySeconds: 300,
		OtpResendSeconds: 60,
		RPCRequestExpirySeconds: 30,
		CacheContextTimeout: 5,
		Redis: struct {
			RedisMaxConnections     int    `json:"redisMaxConnections"`
			RedisMaxIdleConnections int    `json:"redisMaxIdleConnections"`
			RedisCon                string `json:"redisCon"`
			RedisAddress            string `json:"redisAddress"`
			PrintRedis              bool   `json:"printRedis"`
		}{
			RedisMaxConnections:     100,
			RedisMaxIdleConnections: 10,
		},
		PublisherBatchSize: 100,
		DBBatchSize:        50,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.validateConfig()
	}
}

func BenchmarkHostnamePortValidation(b *testing.B) {
	address := "localhost:8080"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validateHostnamePort(address)
	}
}