package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorCode represents application error codes
type ErrorCode string

const (
	// General errors
	ErrCodeInternal         ErrorCode = "INTERNAL_ERROR"
	ErrCodeBadRequest       ErrorCode = "BAD_REQUEST"
	ErrCodeNotFound         ErrorCode = "NOT_FOUND"
	ErrCodeUnauthorized     ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden        ErrorCode = "FORBIDDEN"
	ErrCodeConflict         ErrorCode = "CONFLICT"
	ErrCodeValidation       ErrorCode = "VALIDATION_ERROR"
	
	// Auth errors
	ErrCodeInvalidCredentials ErrorCode = "INVALID_CREDENTIALS"
	ErrCodeInvalidToken       ErrorCode = "INVALID_TOKEN"
	ErrCodeTokenExpired       ErrorCode = "TOKEN_EXPIRED"
	ErrCodeInvalidAPIKey      ErrorCode = "INVALID_API_KEY"
	
	// User errors
	ErrCodeUserNotFound     ErrorCode = "USER_NOT_FOUND"
	ErrCodeUserExists       ErrorCode = "USER_ALREADY_EXISTS"
	ErrCodeInvalidEmail     ErrorCode = "INVALID_EMAIL"
	
	// Project errors
	ErrCodeProjectNotFound  ErrorCode = "PROJECT_NOT_FOUND"
	ErrCodeProjectExists    ErrorCode = "PROJECT_ALREADY_EXISTS"
	
	// Config errors
	ErrCodeConfigNotFound       ErrorCode = "CONFIG_NOT_FOUND"
	ErrCodeConfigExists         ErrorCode = "CONFIG_ALREADY_EXISTS"
	ErrCodeConfigVersionMismatch ErrorCode = "CONFIG_VERSION_MISMATCH"
	ErrCodeConfigValidation     ErrorCode = "CONFIG_VALIDATION_ERROR"
	
	// Schema errors
	ErrCodeSchemaNotFound   ErrorCode = "SCHEMA_NOT_FOUND"
	ErrCodeSchemaExists     ErrorCode = "SCHEMA_ALREADY_EXISTS"
	ErrCodeSchemaInvalid    ErrorCode = "SCHEMA_INVALID"
	
	// Role errors
	ErrCodeRoleNotFound     ErrorCode = "ROLE_NOT_FOUND"
	ErrCodeInsufficientRole ErrorCode = "INSUFFICIENT_ROLE"
	
	// Rate limiting
	ErrCodeRateLimitExceeded ErrorCode = "RATE_LIMIT_EXCEEDED"
)

// AppError represents an application error
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	Err        error     `json:"-"`
	StatusCode int       `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap implements errors.Unwrap interface
func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: getHTTPStatusCode(code),
	}
}

// Wrap wraps an error with additional context
func Wrap(err error, code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Err:        err,
		StatusCode: getHTTPStatusCode(code),
	}
}

// WithDetails adds details to an error
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// IsAppError checks if error is AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// GetAppError extracts AppError from error
func GetAppError(err error) (*AppError, bool) {
	var appErr *AppError
	ok := errors.As(err, &appErr)
	return appErr, ok
}

// getHTTPStatusCode maps error codes to HTTP status codes
func getHTTPStatusCode(code ErrorCode) int {
	switch code {
	case ErrCodeBadRequest, ErrCodeValidation, ErrCodeInvalidEmail:
		return http.StatusBadRequest
	case ErrCodeUnauthorized, ErrCodeInvalidCredentials, ErrCodeInvalidToken, 
	     ErrCodeTokenExpired, ErrCodeInvalidAPIKey:
		return http.StatusUnauthorized
	case ErrCodeForbidden, ErrCodeInsufficientRole:
		return http.StatusForbidden
	case ErrCodeNotFound, ErrCodeUserNotFound, ErrCodeProjectNotFound, 
	     ErrCodeConfigNotFound, ErrCodeSchemaNotFound, ErrCodeRoleNotFound:
		return http.StatusNotFound
	case ErrCodeConflict, ErrCodeUserExists, ErrCodeProjectExists, 
	     ErrCodeConfigExists, ErrCodeSchemaExists, ErrCodeConfigVersionMismatch:
		return http.StatusConflict
	case ErrCodeRateLimitExceeded:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

// Common error constructors for convenience

func BadRequest(message string) *AppError {
	return New(ErrCodeBadRequest, message)
}

func NotFound(resource, id string) *AppError {
	return New(ErrCodeNotFound, fmt.Sprintf("%s with id '%s' not found", resource, id))
}

func Unauthorized(message string) *AppError {
	return New(ErrCodeUnauthorized, message)
}

func Forbidden(message string) *AppError {
	return New(ErrCodeForbidden, message)
}

func Conflict(message string) *AppError {
	return New(ErrCodeConflict, message)
}

func Internal(err error, message string) *AppError {
	return Wrap(err, ErrCodeInternal, message)
}

func ValidationError(details string) *AppError {
	return New(ErrCodeValidation, "Validation failed").WithDetails(details)
}

