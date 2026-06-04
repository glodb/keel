package configmanager

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/glodb/keel/settings/logger"
)

// ConfigValidationError represents a configuration validation error
type ConfigValidationError struct {
	Field   string      `json:"field"`
	Value   interface{} `json:"value"`
	Message string      `json:"message"`
}

func (e ConfigValidationError) Error() string {
	return fmt.Sprintf("config validation failed for '%s': %s (value: %v)", e.Field, e.Message, e.Value)
}

// ConfigValidationErrors represents multiple configuration validation errors
type ConfigValidationErrors []ConfigValidationError

func (e ConfigValidationErrors) Error() string {
	if len(e) == 0 {
		return "no configuration validation errors"
	}

	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// validateConfig validates the entire configuration struct
func (c *config) validateConfig() error {
	var errors ConfigValidationErrors

	// Basic field validation
	errors = append(errors, c.validateBasicFields()...)
	errors = append(errors, c.validateDatabaseConfig()...)
	errors = append(errors, c.validateTimeoutConfig()...)
	errors = append(errors, c.validateSecurityConfig()...)
	errors = append(errors, c.validateBusinessLogic()...)

	if len(errors) > 0 {
		// Log all errors for debugging
		for _, err := range errors {
			logger.Log().Error("Configuration validation error",
				logger.StringField("field", err.Field),
				logger.StringField("message", err.Message),
				logger.AnyField("value", err.Value))
		}
		return errors
	}

	return nil
}

// validateBasicFields validates basic required fields
func (c *config) validateBasicFields() ConfigValidationErrors {
	var errors ConfigValidationErrors

	// Validate ClassName (service name)
	if strings.TrimSpace(c.ClassName) == "" {
		errors = append(errors, ConfigValidationError{
			Field:   "ClassName",
			Value:   c.ClassName,
			Message: "service name is required",
		})
	} else {
		validServices := []string{"SSOSERVICE", "OTPSERVICE",
			"NOTIFICATIONSERVICE", "NOTIFICATIONSENDERSERVICE", "SUPPORTSERVICE",
			"SETTINGSSERVICE", "ADMINSERVICE", "EVENTSSERVICE", "PALSSERVICE", "CHATSERVICE",
			"PROCESSORSERVICE", "EVENTCREATORCRON", "EVENTPROCESSORCRON", "VENUESERVICE",
			"FACILITYSERVICE", "FACILITYBOOKINGSERVICE", "CHATPROCESSORSERVICE"}
		normalizedClassName := strings.ToUpper(strings.TrimSpace(c.ClassName))
		isValid := false
		for _, valid := range validServices {
			if normalizedClassName == valid {
				isValid = true
				break
			}
		}
		if !isValid {
			errors = append(errors, ConfigValidationError{
				Field:   "ClassName",
				Value:   c.ClassName,
				Message: fmt.Sprintf("must be one of: %s", strings.Join(validServices, ", ")),
			})
		}
	}

	// Validate Address (hostname:port)
	if strings.TrimSpace(c.Address) == "" {
		errors = append(errors, ConfigValidationError{
			Field:   "Address",
			Value:   c.Address,
			Message: "server address is required",
		})
	} else {
		if err := validateHostnamePort(c.Address); err != nil {
			errors = append(errors, ConfigValidationError{
				Field:   "Address",
				Value:   c.Address,
				Message: err.Error(),
			})
		}
	}

	// Validate DeploymentEnv
	if strings.TrimSpace(c.DeploymentEnv) == "" {
		errors = append(errors, ConfigValidationError{
			Field:   "DeploymentEnv",
			Value:   c.DeploymentEnv,
			Message: "deployment environment is required",
		})
	} else {
		validEnvs := []string{"DEV", "UAT", "PROD", "TEST"}
		normalizedEnv := strings.ToUpper(strings.TrimSpace(c.DeploymentEnv))
		isValid := false
		for _, valid := range validEnvs {
			if normalizedEnv == valid {
				isValid = true
				break
			}
		}
		if !isValid {
			errors = append(errors, ConfigValidationError{
				Field:   "DeploymentEnv",
				Value:   c.DeploymentEnv,
				Message: fmt.Sprintf("must be one of: %s", strings.Join(validEnvs, ", ")),
			})
		}
	}

	return errors
}

// validateDatabaseConfig validates database-related configuration
func (c *config) validateDatabaseConfig() ConfigValidationErrors {
	var errors ConfigValidationErrors

	// Validate MongoDB configuration
	if c.Mongo.Host != "" {
		if c.Mongo.Port != "" {
			if port, err := strconv.Atoi(c.Mongo.Port); err != nil || port < 1 || port > 65535 {
				errors = append(errors, ConfigValidationError{
					Field:   "Mongo.Port",
					Value:   c.Mongo.Port,
					Message: "MongoDB port must be a valid port number (1-65535)",
				})
			}
		}

		if c.Mongo.MongoMaxConnections <= 0 {
			errors = append(errors, ConfigValidationError{
				Field:   "Mongo.MongoMaxConnections",
				Value:   c.Mongo.MongoMaxConnections,
				Message: "MongoDB max connections must be positive",
			})
		} else if c.Mongo.MongoMaxConnections > 1000 {
			errors = append(errors, ConfigValidationError{
				Field:   "Mongo.MongoMaxConnections",
				Value:   c.Mongo.MongoMaxConnections,
				Message: "MongoDB max connections should not exceed 1000 for performance reasons",
			})
		}
	}

	// Validate PostgreSQL configuration
	if c.PSql.Host != "" {
		if c.PSql.Port != "" {
			if port, err := strconv.Atoi(c.PSql.Port); err != nil || port < 1 || port > 65535 {
				errors = append(errors, ConfigValidationError{
					Field:   "PSql.Port",
					Value:   c.PSql.Port,
					Message: "PostgreSQL port must be a valid port number (1-65535)",
				})
			}
		}
	}

	// Validate MySQL configuration
	if c.MySql.Host != "" {
		if c.MySql.Port != "" {
			if port, err := strconv.Atoi(c.MySql.Port); err != nil || port < 1 || port > 65535 {
				errors = append(errors, ConfigValidationError{
					Field:   "MySql.Port",
					Value:   c.MySql.Port,
					Message: "MySQL port must be a valid port number (1-65535)",
				})
			}
		}
	}

	// Validate Redis configuration
	if c.Redis.RedisMaxConnections <= 0 {
		errors = append(errors, ConfigValidationError{
			Field:   "Redis.RedisMaxConnections",
			Value:   c.Redis.RedisMaxConnections,
			Message: "Redis max connections must be positive",
		})
	} else if c.Redis.RedisMaxConnections > 10000 {
		errors = append(errors, ConfigValidationError{
			Field:   "Redis.RedisMaxConnections",
			Value:   c.Redis.RedisMaxConnections,
			Message: "Redis max connections should not exceed 10000",
		})
	}

	if c.Redis.RedisMaxIdleConnections <= 0 {
		errors = append(errors, ConfigValidationError{
			Field:   "Redis.RedisMaxIdleConnections",
			Value:   c.Redis.RedisMaxIdleConnections,
			Message: "Redis max idle connections must be positive",
		})
	}

	if c.Redis.RedisMaxIdleConnections > c.Redis.RedisMaxConnections {
		errors = append(errors, ConfigValidationError{
			Field:   "Redis.RedisMaxIdleConnections",
			Value:   c.Redis.RedisMaxIdleConnections,
			Message: "Redis max idle connections cannot exceed max connections",
		})
	}

	return errors
}

// validateTimeoutConfig validates timeout and duration settings
func (c *config) validateTimeoutConfig() ConfigValidationErrors {
	var errors ConfigValidationErrors

	// Validate TokenExpiry (5 minutes to 24 hours)
	if c.TokenExpiry < 300 || c.TokenExpiry > 86400 {
		errors = append(errors, ConfigValidationError{
			Field:   "TokenExpiry",
			Value:   c.TokenExpiry,
			Message: "token expiry must be between 5 minutes (300s) and 24 hours (86400s)",
		})
	}

	// Validate OtpExpirySeconds (30 seconds to 30 minutes) only when otp is enabled
	if c.OtpExpirySeconds > 0 {
		if c.OtpExpirySeconds < 30 || c.OtpExpirySeconds > 1800 {
			errors = append(errors, ConfigValidationError{
				Field:   "OtpExpirySeconds",
				Value:   c.OtpExpirySeconds,
				Message: "OTP expiry must be between 30 seconds and 30 minutes (1800s)",
			})
		}
	}

	// Validate RPCRequestExpirySeconds
	if c.RPCRequestExpirySeconds <= 0 || c.RPCRequestExpirySeconds > 300 {
		errors = append(errors, ConfigValidationError{
			Field:   "RPCRequestExpirySeconds",
			Value:   c.RPCRequestExpirySeconds,
			Message: "RPC request expiry must be between 1 and 300 seconds",
		})
	}

	// Validate CacheContextTimeout
	if c.CacheContextTimeout <= 0 || c.CacheContextTimeout > 60 {
		errors = append(errors, ConfigValidationError{
			Field:   "CacheContextTimeout",
			Value:   c.CacheContextTimeout,
			Message: "cache context timeout must be between 1 and 60 seconds",
		})
	}

	return errors
}

// validateSecurityConfig validates security-related settings
func (c *config) validateSecurityConfig() ConfigValidationErrors {
	var errors ConfigValidationErrors

	// Validate SessionKey
	if strings.TrimSpace(c.SessionKey) == "" {
		errors = append(errors, ConfigValidationError{
			Field:   "SessionKey",
			Value:   "[REDACTED]",
			Message: "session key is required for security",
		})
	} else if len(c.SessionKey) < 16 {
		errors = append(errors, ConfigValidationError{
			Field:   "SessionKey",
			Value:   "[REDACTED]",
			Message: "session key must be at least 16 characters for security",
		})
	}

	// Validate secure cookie settings if enabled
	if c.UseSecureCookie {
		if strings.TrimSpace(c.SecureCookieHash) == "" {
			errors = append(errors, ConfigValidationError{
				Field:   "SecureCookieHash",
				Value:   "[REDACTED]",
				Message: "secure cookie hash is required when secure cookies are enabled",
			})
		}

		if strings.TrimSpace(c.SecureCookieBlock) == "" {
			errors = append(errors, ConfigValidationError{
				Field:   "SecureCookieBlock",
				Value:   "[REDACTED]",
				Message: "secure cookie block is required when secure cookies are enabled",
			})
		}

		if c.SecureSessionExpirySeconds <= 0 {
			errors = append(errors, ConfigValidationError{
				Field:   "SecureSessionExpirySeconds",
				Value:   c.SecureSessionExpirySeconds,
				Message: "secure session expiry must be positive when secure cookies are enabled",
			})
		}
	}

	return errors
}

// validateBusinessLogic validates business logic constraints
func (c *config) validateBusinessLogic() ConfigValidationErrors {
	var errors ConfigValidationErrors

	// Validate pagination settings
	if c.PageSize <= 0 {
		errors = append(errors, ConfigValidationError{
			Field:   "PageSize",
			Value:   c.PageSize,
			Message: "page size must be positive",
		})
	}

	if c.MaxPageSize <= 0 {
		errors = append(errors, ConfigValidationError{
			Field:   "MaxPageSize",
			Value:   c.MaxPageSize,
			Message: "max page size must be positive",
		})
	}

	if c.PageSize > c.MaxPageSize {
		errors = append(errors, ConfigValidationError{
			Field:   "PageSize",
			Value:   c.PageSize,
			Message: "page size cannot be larger than max page size",
		})
	}

	// Reasonable limits for pagination
	if c.MaxPageSize > 10000 {
		errors = append(errors, ConfigValidationError{
			Field:   "MaxPageSize",
			Value:   c.MaxPageSize,
			Message: "max page size should not exceed 10000 for performance reasons",
		})
	}

	// Validate batch processing settings
	if c.PublisherBatchSize <= 0 {
		errors = append(errors, ConfigValidationError{
			Field:   "PublisherBatchSize",
			Value:   c.PublisherBatchSize,
			Message: "publisher batch size must be positive",
		})
	} else if c.PublisherBatchSize > 10000 {
		errors = append(errors, ConfigValidationError{
			Field:   "PublisherBatchSize",
			Value:   c.PublisherBatchSize,
			Message: "publisher batch size should not exceed 10000 for performance reasons",
		})
	}

	if c.DBBatchSize <= 0 {
		errors = append(errors, ConfigValidationError{
			Field:   "DBBatchSize",
			Value:   c.DBBatchSize,
			Message: "database batch size must be positive",
		})
	} else if c.DBBatchSize > 1000 {
		errors = append(errors, ConfigValidationError{
			Field:   "DBBatchSize",
			Value:   c.DBBatchSize,
			Message: "database batch size should not exceed 1000 for performance reasons",
		})
	}

	// Validate retry settings
	if c.RedisRetries < 0 {
		errors = append(errors, ConfigValidationError{
			Field:   "RedisRetries",
			Value:   c.RedisRetries,
			Message: "Redis retries cannot be negative",
		})
	} else if c.RedisRetries > 10 {
		errors = append(errors, ConfigValidationError{
			Field:   "RedisRetries",
			Value:   c.RedisRetries,
			Message: "Redis retries should not exceed 10",
		})
	}

	// Validate timing settings
	if c.MessageSendingMilliSeconds <= 0 {
		errors = append(errors, ConfigValidationError{
			Field:   "MessageSendingMilliSeconds",
			Value:   c.MessageSendingMilliSeconds,
			Message: "message sending interval must be positive",
		})
	}

	return errors
}

// Helper functions

// validateHostnamePort validates hostname:port format
func validateHostnamePort(address string) error {
	host, portStr, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("must be in format 'hostname:port': %w", err)
	}

	if host == "" {
		return errors.New("hostname cannot be empty")
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("port must be a valid integer: %w", err)
	}

	if port < 1 || port > 65535 {
		return errors.New("port must be between 1 and 65535")
	}

	return nil
}

// validateTimeoutSeconds validates timeout values in seconds
func validateTimeoutSeconds(value int64, min, max int64) error {
	if value < min || value > max {
		return fmt.Errorf("timeout must be between %d and %d seconds", min, max)
	}
	return nil
}

// GetCacheContextTimeout returns cache context timeout as time.Duration
func (c *config) GetCacheContextTimeout() time.Duration {
	if c.CacheContextTimeout <= 0 {
		return 5 * time.Second // Default 5 seconds
	}
	return time.Duration(c.CacheContextTimeout) * time.Second
}
