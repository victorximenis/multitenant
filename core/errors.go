package core

import "fmt"

// ErrorCode represents specific error types
type ErrorCode string

const (
	// Tenant related errors
	ErrCodeTenantNotFound ErrorCode = "TENANT_NOT_FOUND"
	ErrCodeTenantInactive ErrorCode = "TENANT_INACTIVE"
	ErrCodeTenantExists   ErrorCode = "TENANT_EXISTS"
	ErrCodeTenantInvalid  ErrorCode = "TENANT_INVALID"

	// Database related errors
	ErrCodeDatabaseConnection ErrorCode = "DATABASE_CONNECTION"
	ErrCodeDatabaseQuery      ErrorCode = "DATABASE_QUERY"
	ErrCodeDatabaseTimeout    ErrorCode = "DATABASE_TIMEOUT"

	// Cache related errors
	ErrCodeCacheConnection ErrorCode = "CACHE_CONNECTION"
	ErrCodeCacheTimeout    ErrorCode = "CACHE_TIMEOUT"

	// Configuration errors
	ErrCodeConfigInvalid ErrorCode = "CONFIG_INVALID"
	ErrCodeConfigMissing ErrorCode = "CONFIG_MISSING"

	// Validation errors
	ErrCodeValidationFailed ErrorCode = "VALIDATION_FAILED"

	// Internal errors
	ErrCodeInternal ErrorCode = "INTERNAL_ERROR"
)

// MultitenantError represents a structured error with code and context
type MultitenantError struct {
	Code    ErrorCode              `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
	Cause   error                  `json:"-"`
}

// Error implements the error interface
func (e *MultitenantError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause
func (e *MultitenantError) Unwrap() error {
	return e.Cause
}

// WithDetail adds a detail to the error
func (e *MultitenantError) WithDetail(key string, value interface{}) *MultitenantError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// WithCause sets the underlying cause
func (e *MultitenantError) WithCause(cause error) *MultitenantError {
	e.Cause = cause
	return e
}

// NewError creates a new MultitenantError
func NewError(code ErrorCode, message string) *MultitenantError {
	return &MultitenantError{
		Code:    code,
		Message: message,
		Details: make(map[string]interface{}),
	}
}

// TenantNotFoundError represents an error when a tenant is not found
type TenantNotFoundError struct {
	Name string
}

// Error implements the error interface for TenantNotFoundError
func (e TenantNotFoundError) Error() string {
	return fmt.Sprintf("tenant not found: %s", e.Name)
}

// TenantInactiveError represents an error when a tenant is inactive
type TenantInactiveError struct {
	Name string
}

// Error implements the error interface for TenantInactiveError
func (e TenantInactiveError) Error() string {
	return fmt.Sprintf("tenant is inactive: %s", e.Name)
}

// Helper functions for common errors

// ErrTenantNotFound creates a tenant not found error
func ErrTenantNotFound(name string) *MultitenantError {
	return NewError(ErrCodeTenantNotFound, fmt.Sprintf("tenant not found: %s", name)).
		WithDetail("tenant_name", name)
}

// ErrTenantInactive creates a tenant inactive error
func ErrTenantInactive(name string) *MultitenantError {
	return NewError(ErrCodeTenantInactive, fmt.Sprintf("tenant is inactive: %s", name)).
		WithDetail("tenant_name", name)
}

// ErrTenantExists creates a tenant already exists error
func ErrTenantExists(name string) *MultitenantError {
	return NewError(ErrCodeTenantExists, fmt.Sprintf("tenant already exists: %s", name)).
		WithDetail("tenant_name", name)
}

// ErrTenantInvalid creates a tenant validation error
func ErrTenantInvalid(name string, reason string) *MultitenantError {
	return NewError(ErrCodeTenantInvalid, fmt.Sprintf("tenant invalid: %s", reason)).
		WithDetail("tenant_name", name).
		WithDetail("reason", reason)
}

// ErrDatabaseConnection creates a database connection error
func ErrDatabaseConnection(dsn string, cause error) *MultitenantError {
	return NewError(ErrCodeDatabaseConnection, "failed to connect to database").
		WithDetail("dsn", dsn).
		WithCause(cause)
}

// ErrDatabaseQuery creates a database query error
func ErrDatabaseQuery(query string, cause error) *MultitenantError {
	return NewError(ErrCodeDatabaseQuery, "database query failed").
		WithDetail("query", query).
		WithCause(cause)
}

// ErrCacheConnection creates a cache connection error
func ErrCacheConnection(url string, cause error) *MultitenantError {
	return NewError(ErrCodeCacheConnection, "failed to connect to cache").
		WithDetail("url", url).
		WithCause(cause)
}

// ErrConfigInvalid creates a configuration validation error
func ErrConfigInvalid(field string, reason string) *MultitenantError {
	return NewError(ErrCodeConfigInvalid, fmt.Sprintf("invalid configuration: %s", reason)).
		WithDetail("field", field).
		WithDetail("reason", reason)
}

// ErrValidationFailed creates a validation error
func ErrValidationFailed(entity string, reason string) *MultitenantError {
	return NewError(ErrCodeValidationFailed, fmt.Sprintf("validation failed for %s: %s", entity, reason)).
		WithDetail("entity", entity).
		WithDetail("reason", reason)
}

// IsErrorCode checks if an error has a specific error code
func IsErrorCode(err error, code ErrorCode) bool {
	if mtErr, ok := err.(*MultitenantError); ok {
		return mtErr.Code == code
	}
	return false
}

// GetErrorCode extracts the error code from an error
func GetErrorCode(err error) ErrorCode {
	if mtErr, ok := err.(*MultitenantError); ok {
		return mtErr.Code
	}
	return ErrCodeInternal
}
