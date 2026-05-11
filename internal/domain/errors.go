package domain

import "errors"

var (
	ErrNotFound          = errors.New("resource not found")
	ErrConflict          = errors.New("resource conflict")
	ErrForbidden         = errors.New("forbidden")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrValidation        = errors.New("validation error")
	ErrInvalidState      = errors.New("invalid state")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrInvalidReference  = errors.New("invalid reference")
	ErrAlreadyExists     = errors.New("resource already exists")
)

type ErrorCode string

const (
	ErrorCodeNotFound          ErrorCode = "not_found"
	ErrorCodeConflict          ErrorCode = "conflict"
	ErrorCodeForbidden         ErrorCode = "forbidden"
	ErrorCodeUnauthorized      ErrorCode = "unauthorized"
	ErrorCodeValidation        ErrorCode = "validation_error"
	ErrorCodeInvalidState      ErrorCode = "invalid_state"
	ErrorCodeInsufficientFunds ErrorCode = "insufficient_funds"
	ErrorCodeInvalidReference  ErrorCode = "invalid_reference"
	ErrorCodeInternal          ErrorCode = "internal_error"
)

type DomainError struct {
	Code    ErrorCode
	Message string
	Cause   error
}

func (e *DomainError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	return string(e.Code)
}

func (e *DomainError) Unwrap() error {
	return e.Cause
}

func NewDomainError(code ErrorCode, message string, cause error) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}
