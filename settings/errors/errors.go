package errors

import (
	"fmt"
	"net/http"
	"time"
)

// ErrorCode represents standardized error codes across the application
type ErrorCode string

const (
	// Database Errors
	ErrCodeDatabaseConnection  ErrorCode = "DB_CONNECTION_ERROR"
	ErrCodeDatabaseTimeout     ErrorCode = "DB_TIMEOUT_ERROR"
	ErrCodeDatabaseConstraint  ErrorCode = "DB_CONSTRAINT_ERROR"
	ErrCodeRecordNotFound      ErrorCode = "RECORD_NOT_FOUND"
	ErrCodeRecordAlreadyExists ErrorCode = "RECORD_ALREADY_EXISTS"
	ErrCodeDatabaseTransaction ErrorCode = "DB_TRANSACTION_ERROR"

	// Authentication & Authorization Errors
	ErrCodeUnauthorized       ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden          ErrorCode = "FORBIDDEN"
	ErrCodeInvalidToken       ErrorCode = "INVALID_TOKEN"
	ErrCodeTokenExpired       ErrorCode = "TOKEN_EXPIRED"
	ErrCodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"
	ErrCodeSessionNotFound    ErrorCode = "SESSION_NOT_FOUND"
	ErrCodeSessionExpired     ErrorCode = "SESSION_EXPIRED"

	// Validation Errors
	ErrCodeValidationFailed     ErrorCode = "VALIDATION_FAILED"
	ErrCodeInvalidInput         ErrorCode = "INVALID_INPUT"
	ErrCodeMissingRequiredField ErrorCode = "MISSING_REQUIRED_FIELD"
	ErrCodeInvalidFormat        ErrorCode = "INVALID_FORMAT"

	// Business Logic Errors
	ErrCodeBusinessRuleViolation ErrorCode = "BUSINESS_RULE_VIOLATION"
	ErrCodeInsufficientFunds     ErrorCode = "INSUFFICIENT_FUNDS"
	ErrCodeResourceLocked        ErrorCode = "RESOURCE_LOCKED"
	ErrCodeOperationNotAllowed   ErrorCode = "OPERATION_NOT_ALLOWED"

	// External Service Errors
	ErrCodeExternalServiceUnavailable ErrorCode = "EXTERNAL_SERVICE_UNAVAILABLE"
	ErrCodeExternalServiceTimeout     ErrorCode = "EXTERNAL_SERVICE_TIMEOUT"
	ErrCodeThirdPartyAPIError         ErrorCode = "THIRD_PARTY_API_ERROR"

	// Rate Limiting & Quota Errors
	ErrCodeRateLimitExceeded ErrorCode = "RATE_LIMIT_EXCEEDED"
	ErrCodeQuotaExceeded     ErrorCode = "QUOTA_EXCEEDED"
	ErrCodeTooManyRequests   ErrorCode = "TOO_MANY_REQUESTS"

	// Configuration & System Errors
	ErrCodeConfigurationError  ErrorCode = "CONFIGURATION_ERROR"
	ErrCodeInternalServerError ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrCodeServiceUnavailable  ErrorCode = "SERVICE_UNAVAILABLE"
	ErrCodeTimeout             ErrorCode = "TIMEOUT"
	ErrCodeCircuitBreakerOpen  ErrorCode = "CIRCUIT_BREAKER_OPEN"

	// Cache Errors
	ErrCodeCacheUnavailable     ErrorCode = "CACHE_UNAVAILABLE"
	ErrCodeCacheKeyNotFound     ErrorCode = "CACHE_KEY_NOT_FOUND"
	ErrCodeCacheConnectionError ErrorCode = "CACHE_CONNECTION_ERROR"
)

// Severity represents the severity level of an error
type Severity string

const (
	SeverityLow      Severity = "LOW"
	SeverityMedium   Severity = "MEDIUM"
	SeverityHigh     Severity = "HIGH"
	SeverityCritical Severity = "CRITICAL"
)

// AppError represents a structured application error with rich context
type AppError struct {
	Code        ErrorCode              `json:"code"`
	Message     string                 `json:"message"`
	Details     string                 `json:"details,omitempty"`
	HTTPStatus  int                    `json:"http_status"`
	Severity    Severity               `json:"severity"`
	Timestamp   time.Time              `json:"timestamp"`
	RequestID   string                 `json:"request_id,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	TraceID     string                 `json:"trace_id,omitempty"`
	SpanID      string                 `json:"span_id,omitempty"`
	Component   string                 `json:"component,omitempty"`
	Operation   string                 `json:"operation,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Cause       error                  `json:"-"` // Original error, not serialized
	Retryable   bool                   `json:"retryable"`
	UserMessage string                 `json:"user_message,omitempty"` // User-friendly message
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error for error wrapping
func (e *AppError) Unwrap() error {
	return e.Cause
}

// IsRetryable returns whether the error is retryable
func (e *AppError) IsRetryable() bool {
	return e.Retryable
}

// GetUserMessage returns a user-friendly error message
func (e *AppError) GetUserMessage() string {
	if e.UserMessage != "" {
		return e.UserMessage
	}
	// Return generic message based on severity
	switch e.Severity {
	case SeverityCritical, SeverityHigh:
		return "A system error occurred. Please try again later or contact support."
	case SeverityMedium:
		return "An error occurred while processing your request. Please try again."
	default:
		return e.Message
	}
}

// WithMetadata adds metadata to the error
func (e *AppError) WithMetadata(key string, value interface{}) *AppError {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
	return e
}

// WithContext adds contextual information to the error
func (e *AppError) WithContext(requestID, userID, traceID, spanID string) *AppError {
	e.RequestID = requestID
	e.UserID = userID
	e.TraceID = traceID
	e.SpanID = spanID
	return e
}

// WithComponent adds component and operation information
func (e *AppError) WithComponent(component, operation string) *AppError {
	e.Component = component
	e.Operation = operation
	return e
}

// NewAppError creates a new application error
func NewAppError(code ErrorCode, message string, cause error) *AppError {
	httpStatus := getHTTPStatusFromCode(code)
	severity := getSeverityFromCode(code)
	retryable := isRetryableError(code)

	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Severity:   severity,
		Timestamp:  time.Now(),
		Cause:      cause,
		Retryable:  retryable,
	}
}

// NewAppErrorWithDetails creates a new application error with details
func NewAppErrorWithDetails(code ErrorCode, message, details string, cause error) *AppError {
	err := NewAppError(code, message, cause)
	err.Details = details
	return err
}

// NewDatabaseError creates a database-specific error
func NewDatabaseError(operation string, cause error) *AppError {
	code := ErrCodeDatabaseConnection
	if cause != nil {
		// Analyze the underlying error to determine the specific code
		errStr := cause.Error()
		if contains(errStr, "timeout") {
			code = ErrCodeDatabaseTimeout
		} else if contains(errStr, "constraint") {
			code = ErrCodeDatabaseConstraint
		} else if contains(errStr, "not found") {
			code = ErrCodeRecordNotFound
		} else if contains(errStr, "duplicate", "exists") {
			code = ErrCodeRecordAlreadyExists
		}
	}

	return NewAppError(code, fmt.Sprintf("Database operation failed: %s", operation), cause).
		WithComponent("database", operation)
}

// NewAuthenticationError creates an authentication-specific error
func NewAuthenticationError(message string, cause error) *AppError {
	return NewAppError(ErrCodeUnauthorized, message, cause).
		WithComponent("authentication", "verify")
}

// NewValidationError creates a validation-specific error
func NewValidationError(field, message string, cause error) *AppError {
	return NewAppError(ErrCodeValidationFailed, fmt.Sprintf("Validation failed for field '%s': %s", field, message), cause).
		WithComponent("validation", "validate").
		WithMetadata("field", field)
}

// getHTTPStatusFromCode maps error codes to HTTP status codes
func getHTTPStatusFromCode(code ErrorCode) int {
	switch code {
	case ErrCodeUnauthorized, ErrCodeInvalidToken, ErrCodeTokenExpired, ErrCodeInvalidCredentials:
		return http.StatusUnauthorized
	case ErrCodeForbidden:
		return http.StatusForbidden
	case ErrCodeRecordNotFound, ErrCodeSessionNotFound:
		return http.StatusNotFound
	case ErrCodeValidationFailed, ErrCodeInvalidInput, ErrCodeMissingRequiredField, ErrCodeInvalidFormat:
		return http.StatusBadRequest
	case ErrCodeRecordAlreadyExists:
		return http.StatusConflict
	case ErrCodeRateLimitExceeded, ErrCodeQuotaExceeded, ErrCodeTooManyRequests:
		return http.StatusTooManyRequests
	case ErrCodeExternalServiceUnavailable, ErrCodeServiceUnavailable:
		return http.StatusServiceUnavailable
	case ErrCodeTimeout, ErrCodeExternalServiceTimeout:
		return http.StatusRequestTimeout
	case ErrCodeCircuitBreakerOpen:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

// getSeverityFromCode determines the severity based on error code
func getSeverityFromCode(code ErrorCode) Severity {
	switch code {
	case ErrCodeDatabaseConnection, ErrCodeInternalServerError, ErrCodeServiceUnavailable:
		return SeverityCritical
	case ErrCodeDatabaseTimeout, ErrCodeExternalServiceUnavailable, ErrCodeCircuitBreakerOpen:
		return SeverityHigh
	case ErrCodeUnauthorized, ErrCodeForbidden, ErrCodeBusinessRuleViolation:
		return SeverityMedium
	default:
		return SeverityLow
	}
}

// isRetryableError determines if an error is retryable
func isRetryableError(code ErrorCode) bool {
	switch code {
	case ErrCodeDatabaseTimeout, ErrCodeExternalServiceUnavailable, ErrCodeExternalServiceTimeout,
		ErrCodeTimeout, ErrCodeServiceUnavailable, ErrCodeCacheUnavailable:
		return true
	case ErrCodeCircuitBreakerOpen, ErrCodeRateLimitExceeded:
		return true // Can retry after some time
	default:
		return false
	}
}

// contains checks if a string contains any of the given substrings
func contains(s string, substrings ...string) bool {
	for _, substring := range substrings {
		if len(s) >= len(substring) {
			for i := 0; i <= len(s)-len(substring); i++ {
				if s[i:i+len(substring)] == substring {
					return true
				}
			}
		}
	}
	return false
}

// WrapError wraps an existing error with application error context
func WrapError(err error, code ErrorCode, message string) *AppError {
	if err == nil {
		return nil
	}

	// If it's already an AppError, enhance it
	if appErr, ok := err.(*AppError); ok {
		appErr.Message = message
		appErr.Code = code
		return appErr
	}

	return NewAppError(code, message, err)
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetErrorCode extracts the error code from an error
func GetErrorCode(err error) ErrorCode {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code
	}
	return ErrCodeInternalServerError
}

// IsRetryable checks if an error is retryable
func IsRetryable(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.IsRetryable()
	}
	return false
}
