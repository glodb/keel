package configmanager

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

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
	var errs ConfigValidationErrors

	errs = append(errs, c.validateBasicFields()...)
	errs = append(errs, c.validateDatabaseConfig()...)
	errs = append(errs, c.validateTimeoutConfig()...)
	errs = append(errs, c.validateBusinessLogic()...)

	if len(errs) > 0 {
		for _, err := range errs {
			logger.Log().Error("Configuration validation error",
				logger.StringField("field", err.Field),
				logger.StringField("message", err.Message),
				logger.AnyField("value", err.Value))
		}
		return errs
	}

	return nil
}

// validateBasicFields validates basic required fields
func (c *config) validateBasicFields() ConfigValidationErrors {
	var errs ConfigValidationErrors

	if strings.TrimSpace(c.Address) == "" {
		errs = append(errs, ConfigValidationError{
			Field:   "Address",
			Value:   c.Address,
			Message: "server address is required",
		})
	} else {
		if err := validateHostnamePort(c.Address); err != nil {
			errs = append(errs, ConfigValidationError{
				Field:   "Address",
				Value:   c.Address,
				Message: err.Error(),
			})
		}
	}

	return errs
}

// validateDatabaseConfig validates database-related configuration
func (c *config) validateDatabaseConfig() ConfigValidationErrors {
	var errs ConfigValidationErrors

	if c.Mongo.Host != "" {
		if c.Mongo.Port != "" {
			if port, err := strconv.Atoi(c.Mongo.Port); err != nil || port < 1 || port > 65535 {
				errs = append(errs, ConfigValidationError{
					Field:   "Mongo.Port",
					Value:   c.Mongo.Port,
					Message: "MongoDB port must be a valid port number (1-65535)",
				})
			}
		}
		if c.Mongo.MongoMaxConnections <= 0 {
			errs = append(errs, ConfigValidationError{
				Field:   "Mongo.MongoMaxConnections",
				Value:   c.Mongo.MongoMaxConnections,
				Message: "MongoDB max connections must be positive",
			})
		} else if c.Mongo.MongoMaxConnections > 1000 {
			errs = append(errs, ConfigValidationError{
				Field:   "Mongo.MongoMaxConnections",
				Value:   c.Mongo.MongoMaxConnections,
				Message: "MongoDB max connections should not exceed 1000 for performance reasons",
			})
		}
	}

	if c.PSql.Host != "" {
		if c.PSql.Port != "" {
			if port, err := strconv.Atoi(c.PSql.Port); err != nil || port < 1 || port > 65535 {
				errs = append(errs, ConfigValidationError{
					Field:   "PSql.Port",
					Value:   c.PSql.Port,
					Message: "PostgreSQL port must be a valid port number (1-65535)",
				})
			}
		}
	}

	if c.MySql.Host != "" {
		if c.MySql.Port != "" {
			if port, err := strconv.Atoi(c.MySql.Port); err != nil || port < 1 || port > 65535 {
				errs = append(errs, ConfigValidationError{
					Field:   "MySql.Port",
					Value:   c.MySql.Port,
					Message: "MySQL port must be a valid port number (1-65535)",
				})
			}
		}
	}

	if c.Redis.RedisMaxConnections > 0 {
		if c.Redis.RedisMaxConnections > 10000 {
			errs = append(errs, ConfigValidationError{
				Field:   "Redis.RedisMaxConnections",
				Value:   c.Redis.RedisMaxConnections,
				Message: "Redis max connections should not exceed 10000",
			})
		}
		if c.Redis.RedisMaxIdleConnections > 0 && c.Redis.RedisMaxIdleConnections > c.Redis.RedisMaxConnections {
			errs = append(errs, ConfigValidationError{
				Field:   "Redis.RedisMaxIdleConnections",
				Value:   c.Redis.RedisMaxIdleConnections,
				Message: "Redis max idle connections cannot exceed max connections",
			})
		}
	}

	return errs
}

// validateTimeoutConfig validates timeout and duration settings
func (c *config) validateTimeoutConfig() ConfigValidationErrors {
	var errs ConfigValidationErrors

	if c.RPCRequestExpirySeconds > 0 && c.RPCRequestExpirySeconds > 300 {
		errs = append(errs, ConfigValidationError{
			Field:   "RPCRequestExpirySeconds",
			Value:   c.RPCRequestExpirySeconds,
			Message: "RPC request expiry should not exceed 300 seconds",
		})
	}

	if c.CacheContextTimeout > 60 {
		errs = append(errs, ConfigValidationError{
			Field:   "CacheContextTimeout",
			Value:   c.CacheContextTimeout,
			Message: "cache context timeout should not exceed 60 seconds",
		})
	}

	return errs
}

// validateBusinessLogic validates business logic constraints
func (c *config) validateBusinessLogic() ConfigValidationErrors {
	var errs ConfigValidationErrors

	if c.PageSize > 0 && c.MaxPageSize > 0 && c.PageSize > c.MaxPageSize {
		errs = append(errs, ConfigValidationError{
			Field:   "PageSize",
			Value:   c.PageSize,
			Message: "page size cannot be larger than max page size",
		})
	}

	if c.MaxPageSize > 10000 {
		errs = append(errs, ConfigValidationError{
			Field:   "MaxPageSize",
			Value:   c.MaxPageSize,
			Message: "max page size should not exceed 10000 for performance reasons",
		})
	}

	if c.PublisherBatchSize > 10000 {
		errs = append(errs, ConfigValidationError{
			Field:   "PublisherBatchSize",
			Value:   c.PublisherBatchSize,
			Message: "publisher batch size should not exceed 10000 for performance reasons",
		})
	}

	if c.RedisRetries > 10 {
		errs = append(errs, ConfigValidationError{
			Field:   "RedisRetries",
			Value:   c.RedisRetries,
			Message: "Redis retries should not exceed 10",
		})
	}

	if c.MessageSendingMilliSeconds < 0 {
		errs = append(errs, ConfigValidationError{
			Field:   "MessageSendingMilliSeconds",
			Value:   c.MessageSendingMilliSeconds,
			Message: "message sending interval must not be negative",
		})
	}

	return errs
}

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
