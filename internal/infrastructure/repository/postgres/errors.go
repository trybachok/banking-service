package postgres

import (
	"database/sql"
	"errors"

	"github.com/example/banking-service/internal/domain"

	"github.com/lib/pq"
)

func mapDBError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return domain.NewDomainError(domain.ErrorCodeNotFound, "resource not found", domain.ErrNotFound)
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch string(pqErr.Code) {
		case "23505":
			return domain.NewDomainError(domain.ErrorCodeConflict, "resource already exists", domain.ErrConflict)
		case "23503":
			return domain.NewDomainError(domain.ErrorCodeInvalidReference, "invalid related entity reference", domain.ErrInvalidReference)
		case "23514":
			return domain.NewDomainError(domain.ErrorCodeValidation, "database constraint violation", domain.ErrValidation)
		case "40001", "40P01":
			return domain.NewDomainError(domain.ErrorCodeConflict, "transaction conflict, retry required", err)
		default:
			return err
		}
	}

	return err
}
