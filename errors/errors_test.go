package errors

import (
	"errors"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		appError *AppError
		want     string
	}{
		{
			name: "Error without cause",
			appError: &AppError{
				Code:    ErrNotFound,
				Message: "resource not found",
			},
			want: "[NOT_FOUND] resource not found",
		},
		{
			name: "Error with cause",
			appError: &AppError{
				Code:    ErrDatabase,
				Message: "query failed",
				Cause:   errors.New("connection timeout"),
			},
			want: "[DATABASE_ERROR] query failed: connection timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.appError.Error(); got != tt.want {
				t.Errorf("AppError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError("user not found")
	if err.Code != ErrNotFound {
		t.Errorf("Expected code %v, got %v", ErrNotFound, err.Code)
	}
	if err.Message != "user not found" {
		t.Errorf("Expected message 'user not found', got %v", err.Message)
	}
}

func TestNewDatabaseError(t *testing.T) {
	cause := errors.New("connection failed")
	err := NewDatabaseError("database operation failed", cause)
	if err.Code != ErrDatabase {
		t.Errorf("Expected code %v, got %v", ErrDatabase, err.Code)
	}
	if err.Cause != cause {
		t.Errorf("Expected cause to be set")
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "AppError with NOT_FOUND code",
			err:  NewNotFoundError("not found"),
			want: true,
		},
		{
			name: "AppError with different code",
			err:  NewInvalidInputError("invalid"),
			want: false,
		},
		{
			name: "Standard error",
			err:  errors.New("standard error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFound(tt.err); got != tt.want {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidationError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "AppError with VALIDATION_ERROR code",
			err:  NewValidationError("validation failed"),
			want: true,
		},
		{
			name: "AppError with different code",
			err:  NewNotFoundError("not found"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidationError(tt.err); got != tt.want {
				t.Errorf("IsValidationError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := NewDatabaseError("database error", cause)

	unwrapped := errors.Unwrap(err)
	if unwrapped != cause {
		t.Errorf("Expected unwrapped error to be %v, got %v", cause, unwrapped)
	}
}
