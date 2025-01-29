package api

import (
	"fmt"
	"net/http"
	"time"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	StatusCode int       `json:"status_code"`
	Timestamp  time.Time `json:"timestamp"`
	RequestID  string    `json:"request_id"`
	Error      Error     `json:"error"`
}

// Error contains the error details
type Error struct {
	Code    string      `json:"code"`              // Application-specific error code
	Message string      `json:"message"`           // User-friendly error message
	Details interface{} `json:"details,omitempty"` // Additional error context
}

// Common error codes
const (
	ErrCodeValidation     = "VALIDATION_ERROR"
	ErrCodeNotFound       = "NOT_FOUND"
	ErrCodeDatabase       = "DATABASE_ERROR"
	ErrCodeInternal       = "INTERNAL_ERROR"
	ErrCodeUnauthorized   = "UNAUTHORIZED"
	ErrCodeBadRequest     = "BAD_REQUEST"
	ErrCodeRateLimit      = "RATE_LIMIT_EXCEEDED"
)

// AppError represents an application-specific error
type AppError struct {
	Err     error
	Code    string
	Message string
	Status  int
}

func (e *AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewAppError creates a new application error
func NewAppError(err error, code string, message string, status int) *AppError {
	return &AppError{
		Err:     err,
		Code:    code,
		Message: message,
		Status:  status,
	}
}

// Common error creators
func NewValidationError(message string) *AppError {
	return &AppError{
		Code:    ErrCodeValidation,
		Message: message,
		Status:  http.StatusBadRequest,
	}
}

func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:    ErrCodeNotFound,
		Message: fmt.Sprintf("%s not found", resource),
		Status:  http.StatusNotFound,
	}
}

func NewDatabaseError(err error) *AppError {
	return &AppError{
		Err:     err,
		Code:    ErrCodeDatabase,
		Message: "Database operation failed",
		Status:  http.StatusInternalServerError,
	}
}

func NewInternalError(err error) *AppError {
	return &AppError{
		Err:     err,
		Code:    ErrCodeInternal,
		Message: "Internal server error",
		Status:  http.StatusInternalServerError,
	}
}

func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Code:    ErrCodeUnauthorized,
		Message: message,
		Status:  http.StatusUnauthorized,
	}
}

func NewBadRequestError(message string) *AppError {
	return &AppError{
		Code:    ErrCodeBadRequest,
		Message: message,
		Status:  http.StatusBadRequest,
	}
}

func NewTooManyRequestsError(message string) *AppError {
	return &AppError{
		Code:    ErrCodeRateLimit,
		Message: message,
		Status:  http.StatusTooManyRequests,
	}
}