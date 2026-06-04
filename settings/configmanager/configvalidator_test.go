package configmanager

import (
	"strings"
	"testing"

	configmodels "github.com/glodb/keel/settings/configmanager/configmodels"
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
				Address:                 "localhost:8080",
				DeploymentEnv:           "DEV",
				RPCRequestExpirySeconds: 30,
				CacheContextTimeout:     5,
				PageSize:                50,
				MaxPageSize:             1000,
				PublisherBatchSize:      100,
			},
			expectError: false,
		},
		{
			name: "Empty address",
			config: &config{
				Address:       "",
				DeploymentEnv: "DEV",
			},
			expectError: true,
			errorField:  "Address",
			errorMsg:    "server address is required",
		},
		{
			name: "Invalid address format",
			config: &config{
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
				Address:       "localhost:8080",
				DeploymentEnv: "INVALID",
			},
			expectError: true,
			errorField:  "DeploymentEnv",
			errorMsg:    "must be one of",
		},
		{
			name: "Redis idle connections exceed max connections",
			config: &config{
				Address:       "localhost:8080",
				DeploymentEnv: "DEV",
				Redis: configmodels.RedisConnection{
					RedisMaxConnections:     10,
					RedisMaxIdleConnections: 20,
				},
			},
			expectError: true,
			errorField:  "Redis.RedisMaxIdleConnections",
			errorMsg:    "cannot exceed max connections",
		},
		{
			name: "Page size larger than max page size",
			config: &config{
				Address:       "localhost:8080",
				DeploymentEnv: "DEV",
				PageSize:      1000,
				MaxPageSize:   500,
			},
			expectError: true,
			errorField:  "PageSize",
			errorMsg:    "cannot be larger than max page size",
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
		{name: "Valid localhost", address: "localhost:8080"},
		{name: "Valid IP address", address: "192.168.1.1:3000"},
		{name: "Valid domain", address: "api.example.com:443"},
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
		errs     ConfigValidationErrors
		expected string
	}{
		{
			name:     "No errors",
			errs:     ConfigValidationErrors{},
			expected: "no configuration validation errors",
		},
		{
			name: "Single error",
			errs: ConfigValidationErrors{
				{Field: "Address", Value: "", Message: "server address is required"},
			},
			expected: "config validation failed for 'Address': server address is required (value: )",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.errs.Error()
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
		expected string
	}{
		{name: "Valid timeout", timeout: 10, expected: "10s"},
		{name: "Zero timeout - should return default", timeout: 0, expected: "5s"},
		{name: "Negative timeout - should return default", timeout: -5, expected: "5s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config{CacheContextTimeout: tt.timeout}
			result := cfg.GetCacheContextTimeout()
			if result.String() != tt.expected {
				t.Errorf("Expected timeout: %s, got: %s", tt.expected, result.String())
			}
		})
	}
}

func BenchmarkConfigValidation(b *testing.B) {
	cfg := &config{
		Address:                 "localhost:8080",
		DeploymentEnv:           "DEV",
		RPCRequestExpirySeconds: 30,
		CacheContextTimeout:     5,
		PageSize:                50,
		MaxPageSize:             1000,
		PublisherBatchSize:      100,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cfg.validateConfig()
	}
}

func BenchmarkHostnamePortValidation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validateHostnamePort("localhost:8080")
	}
}
