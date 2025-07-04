package apperrors

import (
	"errors"
	"net/http"
)

// Error represents a custom error with a code and message.
type Error struct {
	Code    int
	Message string
	Err     error
}

// Error returns the error message.
func (e *Error) Error() string {
	return e.Message
}

// Unwrap provides compatibility for errors.Is and errors.As.
func (e *Error) Unwrap() error {
	return e.Err
}

// New creates a new custom error.
func New(code int, message string, err error) *Error {
	return &Error{Code: code, Message: message, Err: err}
}

// Define specific error types for the application.
// This allows us to handle them consistently in our middleware.
var (
	ErrNotFound         = errors.New("not found")
	ErrValidation       = errors.New("validation failed")
	ErrConflict         = errors.New("resource conflict or already exists")
	ErrPermissionDenied = errors.New("permission denied")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrInternalServer   = errors.New("internal server error")
)

// MapDomainErrors maps our custom domain errors to HTTP status codes.
func MapDomainErrors(err error) *Error {
	switch {
	case errors.Is(err, ErrNotFound):
		return New(http.StatusNotFound, err.Error(), err)
	case errors.Is(err, ErrValidation):
		return New(http.StatusBadRequest, err.Error(), err)
	case errors.Is(err, ErrConflict):
		return New(http.StatusConflict, err.Error(), err)
	case errors.Is(err, ErrPermissionDenied):
		return New(http.StatusForbidden, err.Error(), err)
	case errors.Is(err, ErrUnauthorized):
		return New(http.StatusUnauthorized, err.Error(), err)
	default:
		return New(http.StatusInternalServerError, "An unexpected error occurred", err)
	}
}
