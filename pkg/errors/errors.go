// Package errors provides centralized error handling for NorthStack
package errors

import (
	"fmt"
	"net/http"
)

// Code represents an error code
type Code string

// Error codes
const (
	CodeNotFound           Code = "NOT_FOUND"
	CodeInvalidInput       Code = "INVALID_INPUT"
	CodeUnauthorized       Code = "UNAUTHORIZED"
	CodeForbidden          Code = "FORBIDDEN"
	CodeConflict           Code = "CONFLICT"
	CodeInternalError      Code = "INTERNAL_ERROR"
	CodeRateLimitExceeded  Code = "RATE_LIMIT_EXCEEDED"
	CodeServiceUnavailable Code = "SERVICE_UNAVAILABLE"
	CodeValidationFailed   Code = "VALIDATION_FAILED"
	CodeBuildFailed        Code = "BUILD_FAILED"
	CodeDeploymentFailed   Code = "DEPLOYMENT_FAILED"
	CodeDatabaseError      Code = "DATABASE_ERROR"
)

// AppError represents an application error
type AppError struct {
	Code       Code        `json:"code"`
	Message    string      `json:"message"`
	Details    interface{} `json:"details,omitempty"`
	HTTPStatus int         `json:"-"`
	Err        error       `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(details interface{}) *AppError {
	e.Details = details
	return e
}

// WithError wraps an underlying error
func (e *AppError) WithError(err error) *AppError {
	e.Err = err
	return e
}

// NewError creates a new AppError
func NewError(code Code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// Predefined errors
var (
	ErrNotFound = NewError(
		CodeNotFound,
		"Resource not found",
		http.StatusNotFound,
	)
	ErrUnauthorized = NewError(
		CodeUnauthorized,
		"Authentication required",
		http.StatusUnauthorized,
	)
	ErrForbidden = NewError(
		CodeForbidden,
		"Access denied",
		http.StatusForbidden,
	)
	ErrInvalidInput = NewError(
		CodeInvalidInput,
		"Invalid input",
		http.StatusBadRequest,
	)
	ErrConflict = NewError(
		CodeConflict,
		"Resource conflict",
		http.StatusConflict,
	)
	ErrInternalError = NewError(
		CodeInternalError,
		"Internal server error",
		http.StatusInternalServerError,
	)
	ErrRateLimitExceeded = NewError(
		CodeRateLimitExceeded,
		"Rate limit exceeded",
		http.StatusTooManyRequests,
	)
	ErrServiceUnavailable = NewError(
		CodeServiceUnavailable,
		"Service temporarily unavailable",
		http.StatusServiceUnavailable,
	)
)

// NotFound creates a not found error
func NotFound(resourceAndID ...string) *AppError {
	message := "Resource not found"
	if len(resourceAndID) >= 1 {
		if len(resourceAndID) >= 2 {
			message = fmt.Sprintf("%s not found: %s", resourceAndID[0], resourceAndID[1])
		} else {
			message = fmt.Sprintf("%s not found", resourceAndID[0])
		}
	}
	return NewError(
		CodeNotFound,
		message,
		http.StatusNotFound,
	)
}

// Forbidden creates a forbidden error
func Forbidden(message string) *AppError {
	return NewError(
		CodeForbidden,
		message,
		http.StatusForbidden,
	)
}

// Internal creates an internal error
func Internal(message string) *AppError {
	return NewError(
		CodeInternalError,
		message,
		http.StatusInternalServerError,
	)
}

// InvalidInput creates an invalid input error
func InvalidInput(message string) *AppError {
	return NewError(
		CodeInvalidInput,
		message,
		http.StatusBadRequest,
	)
}

// ValidationFailed creates a validation error
func ValidationFailed(errors map[string]string) *AppError {
	return NewError(
		CodeValidationFailed,
		"Validation failed",
		http.StatusBadRequest,
	).WithDetails(errors)
}

// Conflict creates a conflict error
func Conflict(resource string) *AppError {
	return NewError(
		CodeConflict,
		fmt.Sprintf("%s already exists", resource),
		http.StatusConflict,
	)
}

// BuildFailed creates a build failure error
func BuildFailed(reason string) *AppError {
	return NewError(
		CodeBuildFailed,
		fmt.Sprintf("Build failed: %s", reason),
		http.StatusInternalServerError,
	)
}

// DeploymentFailed creates a deployment failure error
func DeploymentFailed(reason string) *AppError {
	return NewError(
		CodeDeploymentFailed,
		fmt.Sprintf("Deployment failed: %s", reason),
		http.StatusInternalServerError,
	)
}

// Unauthorized creates an unauthorized error
func Unauthorized(message string) *AppError {
	return NewError(
		CodeUnauthorized,
		message,
		http.StatusUnauthorized,
	)
}

// BadRequest creates a bad request error
func BadRequest(message string) *AppError {
	return NewError(
		CodeInvalidInput,
		message,
		http.StatusBadRequest,
	)
}

// IsNotFound checks if error is a not found error
func IsNotFound(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == CodeNotFound
	}
	return false
}

// IsUnauthorized checks if error is an unauthorized error
func IsUnauthorized(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == CodeUnauthorized
	}
	return false
}

// Wrap wraps an error with context
func Wrap(err error, message string) *AppError {
	return NewError(
		CodeInternalError,
		message,
		http.StatusInternalServerError,
	).WithError(err)
}

// DependencyFailed creates a dependency failure error
func DependencyFailed(service string, err error) *AppError {
	return NewError(
		CodeServiceUnavailable,
		fmt.Sprintf("Dependency %s failed", service),
		http.StatusServiceUnavailable,
	).WithError(err)
}

// CodeRateLimited is the rate limit error code
const CodeRateLimited Code = "RATE_LIMITED"

// PlatformError for legacy compatibility
type PlatformError struct {
	Code       string
	Message    string
	Details    string
	HTTPStatus int
	Metadata   map[string]interface{}
}

// GetPlatformError converts an error to PlatformError
func GetPlatformError(err error) *PlatformError {
	if appErr, ok := err.(*AppError); ok {
		return &PlatformError{
			Code:       string(appErr.Code),
			Message:    appErr.Message,
			HTTPStatus: appErr.HTTPStatus,
			Metadata:   map[string]interface{}{},
		}
	}
	return &PlatformError{
		Code:       "INTERNAL_ERROR",
		Message:    err.Error(),
		HTTPStatus: http.StatusInternalServerError,
		Metadata:   map[string]interface{}{},
	}
}
