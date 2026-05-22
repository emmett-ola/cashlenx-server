package errors

import "fmt"

// ErrorCode represents standardized error codes
type ErrorCode string

const (
	ErrNotFound         ErrorCode = "NOT_FOUND"
	ErrInvalidInput     ErrorCode = "INVALID_INPUT"
	ErrDatabase         ErrorCode = "DATABASE_ERROR"
	ErrUnauthorized     ErrorCode = "UNAUTHORIZED"
	ErrAlreadyExists    ErrorCode = "ALREADY_EXISTS"
	ErrConnectionFailed ErrorCode = "CONNECTION_FAILED"
	ErrForbidden        ErrorCode = "FORBIDDEN"
	ErrInternal         ErrorCode = "INTERNAL_ERROR"
	ErrValidation       ErrorCode = "VALIDATION_ERROR"
	ErrRateLimited      ErrorCode = "RATE_LIMITED"
)

// AppError represents an application error with a code, message, and optional cause
type AppError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Field   string    `json:"field,omitempty"`
	Cause   error     `json:"-"`
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// GetCode returns the error code as string
func (e *AppError) GetCode() string {
	return string(e.Code)
}

// GetMessage returns the error message
func (e *AppError) GetMessage() string {
	return e.Message
}

// Unwrap returns the underlying cause error
func (e *AppError) Unwrap() error {
	return e.Cause
}

// NewAppError creates a new application error
func NewAppError(code ErrorCode, message string, cause error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// NewNotFoundError creates a NOT_FOUND error
func NewNotFoundError(message string) *AppError {
	return &AppError{
		Code:    ErrNotFound,
		Message: message,
	}
}

// NewInvalidInputError creates an INVALID_INPUT error
func NewInvalidInputError(message string) *AppError {
	return &AppError{
		Code:    ErrInvalidInput,
		Message: message,
	}
}

// NewBadRequestError creates an INVALID_INPUT error (alias for NewInvalidInputError)
func NewBadRequestError(message string) *AppError {
	return NewInvalidInputError(message)
}

// NewDatabaseError creates a DATABASE_ERROR
func NewDatabaseError(message string, cause error) *AppError {
	return &AppError{
		Code:    ErrDatabase,
		Message: message,
		Cause:   cause,
	}
}

// NewValidationError creates a VALIDATION_ERROR
func NewValidationError(message string) *AppError {
	return &AppError{
		Code:    ErrValidation,
		Message: message,
	}
}

// NewFieldValidationError creates a VALIDATION_ERROR with field information
func NewFieldValidationError(field, message string) *AppError {
	return &AppError{
		Code:    ErrValidation,
		Message: message,
		Field:   field,
	}
}

// NewAlreadyExistsError creates an ALREADY_EXISTS error
func NewAlreadyExistsError(message string) *AppError {
	return &AppError{
		Code:    ErrAlreadyExists,
		Message: message,
	}
}

func NewFieldAlreadyExistsError(field, message string) *AppError {
	return &AppError{
		Code:    ErrAlreadyExists,
		Message: message,
		Field:   field,
	}
}

// NewUnauthorizedError creates an UNAUTHORIZED error
func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Code:    ErrUnauthorized,
		Message: message,
	}
}

// NewForbiddenError creates a FORBIDDEN error
func NewForbiddenError(message string) *AppError {
	return &AppError{
		Code:    ErrForbidden,
		Message: message,
	}
}

// NewInternalError creates an INTERNAL_ERROR
func NewInternalError(message string, cause error) *AppError {
	return &AppError{
		Code:    ErrInternal,
		Message: message,
		Cause:   cause,
	}
}

// NewRateLimitedError creates a RATE_LIMITED error.
func NewRateLimitedError(message string) *AppError {
	return &AppError{
		Code:    ErrRateLimited,
		Message: message,
	}
}

// NewConnectionFailedError creates a CONNECTION_FAILED error
func NewConnectionFailedError(message string, cause error) *AppError {
	return &AppError{
		Code:    ErrConnectionFailed,
		Message: message,
		Cause:   cause,
	}
}

// IsNotFound checks if error is a NOT_FOUND error
func IsNotFound(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == ErrNotFound
	}
	return false
}

// IsValidationError checks if error is a VALIDATION_ERROR
func IsValidationError(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == ErrValidation
	}
	return false
}

// IsDatabaseError checks if error is a DATABASE_ERROR
func IsDatabaseError(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == ErrDatabase
	}
	return false
}

// IsUnauthorizedError checks if error is an UNAUTHORIZED error
func IsUnauthorizedError(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == ErrUnauthorized
	}
	return false
}

// IsForbiddenError checks if error is a FORBIDDEN error
func IsForbiddenError(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == ErrForbidden
	}
	return false
}

// IsAlreadyExistsError checks if error is an ALREADY_EXISTS error
func IsAlreadyExistsError(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == ErrAlreadyExists
	}
	return false
}

// IsRateLimitedError checks if error is a RATE_LIMITED error.
func IsRateLimitedError(err error) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == ErrRateLimited
	}
	return false
}
