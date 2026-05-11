package response

import (
	"errors"
	"net/http"

	"github.com/example/banking-service/internal/domain"
)

type ErrorResponse struct {
	Error ErrorPayload `json:"error"`
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func WriteError(w http.ResponseWriter, err error) {
	status := StatusFromError(err)
	code := CodeFromError(err)

	message := "internal server error"
	if status < http.StatusInternalServerError {
		message = err.Error()
	}

	WriteJSON(w, status, ErrorResponse{
		Error: ErrorPayload{
			Code:    code,
			Message: message,
		},
	})
}

func WriteErrorMessage(w http.ResponseWriter, status int, code string, message string) {
	WriteJSON(w, status, ErrorResponse{
		Error: ErrorPayload{
			Code:    code,
			Message: message,
		},
	})
}

func StatusFromError(err error) int {
	switch {
	case errors.Is(err, domain.ErrValidation):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict
	case errors.Is(err, domain.ErrInsufficientFunds):
		return http.StatusConflict
	case errors.Is(err, domain.ErrInvalidState):
		return http.StatusConflict
	case errors.Is(err, domain.ErrInvalidReference):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func CodeFromError(err error) string {
	var domainErr *domain.DomainError
	if errors.As(err, &domainErr) {
		return string(domainErr.Code)
	}

	switch {
	case errors.Is(err, domain.ErrValidation):
		return string(domain.ErrorCodeValidation)
	case errors.Is(err, domain.ErrUnauthorized):
		return string(domain.ErrorCodeUnauthorized)
	case errors.Is(err, domain.ErrForbidden):
		return string(domain.ErrorCodeForbidden)
	case errors.Is(err, domain.ErrNotFound):
		return string(domain.ErrorCodeNotFound)
	case errors.Is(err, domain.ErrConflict):
		return string(domain.ErrorCodeConflict)
	case errors.Is(err, domain.ErrInsufficientFunds):
		return string(domain.ErrorCodeInsufficientFunds)
	case errors.Is(err, domain.ErrInvalidReference):
		return string(domain.ErrorCodeInvalidReference)
	default:
		return string(domain.ErrorCodeInternal)
	}
}
