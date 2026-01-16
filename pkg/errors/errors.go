// Package errors provides custom error types and utilities for the Platform Orchestrator.
package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Error codes for the platform
const (
	CodeInternal         = "INTERNAL_ERROR"
	CodeNotFound         = "NOT_FOUND"
	CodeBadRequest       = "BAD_REQUEST"
	CodeUnauthorized     = "UNAUTHORIZED"
	CodeForbidden        = "FORBIDDEN"
	CodeConflict         = "CONFLICT"
	CodeValidation       = "VALIDATION_ERROR"
	CodeTimeout          = "TIMEOUT"
	CodeUnavailable      = "SERVICE_UNAVAILABLE"
	CodeRateLimited      = "RATE_LIMITED"
	CodeDependencyFailed = "DEPENDENCY_FAILED"
)

// PlatformError represents a platform-specific error
type PlatformError struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	Details    string                 `json:"details,omitempty"`
	HTTPStatus int                    `json:"-"`
	Cause      error                  `json:"-"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// Error implements the error interface
func (e *PlatformError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause
func (e *PlatformError) Unwrap() error {
	return e.Cause
}

// WithCause adds a cause to the error
func (e *PlatformError) WithCause(err error) *PlatformError {
	e.Cause = err
	return e
}

// WithDetails adds details to the error
func (e *PlatformError) WithDetails(details string) *PlatformError {
	e.Details = details
	return e
}

// WithMetadata adds metadata to the error
func (e *PlatformError) WithMetadata(key string, value interface{}) *PlatformError {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
	return e
}

// New creates a new PlatformError
func New(code string, message string, httpStatus int) *PlatformError {
	return &PlatformError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// Internal creates an internal server error
func Internal(message string) *PlatformError {
	return New(CodeInternal, message, http.StatusInternalServerError)
}

// NotFound creates a not found error
func NotFound(resource string, identifier string) *PlatformError {
	return New(
		CodeNotFound,
		fmt.Sprintf("%s not found: %s", resource, identifier),
		http.StatusNotFound,
	)
}

// BadRequest creates a bad request error
func BadRequest(message string) *PlatformError {
	return New(CodeBadRequest, message, http.StatusBadRequest)
}

// Unauthorized creates an unauthorized error
func Unauthorized(message string) *PlatformError {
	if message == "" {
		message = "authentication required"
	}
	return New(CodeUnauthorized, message, http.StatusUnauthorized)
}

// Forbidden creates a forbidden error
func Forbidden(message string) *PlatformError {
	if message == "" {
		message = "access denied"
	}
	return New(CodeForbidden, message, http.StatusForbidden)
}

// Conflict creates a conflict error
func Conflict(resource string, reason string) *PlatformError {
	return New(
		CodeConflict,
		fmt.Sprintf("%s conflict: %s", resource, reason),
		http.StatusConflict,
	)
}

// Validation creates a validation error
func Validation(field string, message string) *PlatformError {
	return New(
		CodeValidation,
		fmt.Sprintf("validation failed for %s: %s", field, message),
		http.StatusBadRequest,
	)
}

// Timeout creates a timeout error
func Timeout(operation string) *PlatformError {
	return New(
		CodeTimeout,
		fmt.Sprintf("operation timed out: %s", operation),
		http.StatusRequestTimeout,
	)
}

// Unavailable creates a service unavailable error
func Unavailable(service string) *PlatformError {
	return New(
		CodeUnavailable,
		fmt.Sprintf("service unavailable: %s", service),
		http.StatusServiceUnavailable,
	)
}

// RateLimited creates a rate limited error
func RateLimited() *PlatformError {
	return New(
		CodeRateLimited,
		"rate limit exceeded",
		http.StatusTooManyRequests,
	)
}

// DependencyFailed creates a dependency failed error
func DependencyFailed(dependency string, err error) *PlatformError {
	return New(
		CodeDependencyFailed,
		fmt.Sprintf("dependency failed: %s", dependency),
		http.StatusServiceUnavailable,
	).WithCause(err)
}

// IsPlatformError checks if an error is a PlatformError
func IsPlatformError(err error) bool {
	var pe *PlatformError
	return errors.As(err, &pe)
}

// GetPlatformError extracts a PlatformError from an error
func GetPlatformError(err error) *PlatformError {
	var pe *PlatformError
	if errors.As(err, &pe) {
		return pe
	}
	return Internal(err.Error()).WithCause(err)
}

// IsNotFound checks if an error is a not found error
func IsNotFound(err error) bool {
	var pe *PlatformError
	if errors.As(err, &pe) {
		return pe.Code == CodeNotFound
	}
	return false
}

// IsConflict checks if an error is a conflict error
func IsConflict(err error) bool {
	var pe *PlatformError
	if errors.As(err, &pe) {
		return pe.Code == CodeConflict
	}
	return false
}

// IsUnauthorized checks if an error is an unauthorized error
func IsUnauthorized(err error) bool {
	var pe *PlatformError
	if errors.As(err, &pe) {
		return pe.Code == CodeUnauthorized
	}
	return false
}

// IsForbidden checks if an error is a forbidden error
func IsForbidden(err error) bool {
	var pe *PlatformError
	if errors.As(err, &pe) {
		return pe.Code == CodeForbidden
	}
	return false
}

// Wrap wraps an error with additional context
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// Wrapf wraps an error with a formatted message
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}
