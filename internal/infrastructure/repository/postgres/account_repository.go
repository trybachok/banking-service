package postgres

import (
	"context"

	"github.com/example/banking-service/internal/domain"
	"github.com/example/banking-service/internal/infrastructure/db"
)

type AccountRepository struct {
	tx *db.TxManager
}

func NewAccountRepository(tx *db.TxManager) *AccountRepository {
	return &AccountRepository{tx: tx}
}

func (r *AccountRepository) Create(ctx context.Context, account *domain.Account) error {
	query := `
		INSERT INTO accounts (
			user_id,
			account_no,
			account_type,
			currency,
			balance,
			status,
			blocked_reason
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	err := r.tx.Executor(ctx).QueryRowContext(
		ctx,
		query,
		account.UserID.String(),
		account.AccountNo,
		string(account.AccountType),
		account.Currency,
		account.Balance.String(),
		string(account.Status),
		account.BlockedReason,
	).Scan(&account.ID, &account.CreatedAt, &account.UpdatedAt)

	return mapDBError(err)
}

func (r *AccountRepository) FindByID(ctx context.Context, id domain.AccountID) (*domain.Account, error) {
	return r.findByID(ctx, id, false)
}

func (r *AccountRepository) FindByIDForUpdate(ctx context.Context, id domain.AccountID) (*domain.Account, error) {
	return r.findByID(ctx, id, true)
}

func (r *AccountRepository) ListByUserID(ctx context.Context, userID domain.UserID) ([]*domain.Account, error) {
	query := `
		SELECT
			id,
			user_id,
			account_no,
			account_type,
			currency,
			balance::text,
			status,
			COALESCE(blocked_reason, ''),
			created_at,
			updated_at
		FROM accounts
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.tx.Executor(ctx).QueryContext(ctx, query, userID.String())
	if err != nil {
		return nil, mapDBError(err)
	}
	defer rows.Close()

	var accounts []*domain.Account

	for rows.Next() {
		account, err := scanAccount(rows)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}

	return accounts, mapDBError(rows.Err())
}

func (r *AccountRepository) UpdateBalance(ctx context.Context, id domain.AccountID, account *domain.Account) error {
	query := `
		UPDATE accounts
		SET balance = $2
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.tx.Executor(ctx).QueryRowContext(
		ctx,
		query,
		id.String(),
		account.Balance.String(),
	).Scan(&account.UpdatedAt)

	return mapDBError(err)
}

func (r *AccountRepository) UpdateStatus(ctx context.Context, id domain.AccountID, status domain.AccountStatus, reason string) error {
	query := `
		UPDATE accounts
		SET status = $2,
		    blocked_reason = $3
		WHERE id = $1
	`

	result, err := r.tx.Executor(ctx).ExecContext(ctx, query, id.String(), string(status), reason)
	if err != nil {
		return mapDBError(err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return domain.NewDomainError(domain.ErrorCodeNotFound, "account not found", domain.ErrNotFound)
	}

	return nil
}

func (r *AccountRepository) findByID(ctx context.Context, id domain.AccountID, forUpdate bool) (*domain.Account, error) {
	query := `
		SELECT
			id,
			user_id,
			account_no,
			account_type,
			currency,
			balance::text,
			status,
			COALESCE(blocked_reason, ''),
			created_at,
			updated_at
		FROM accounts
		WHERE id = $1
	`

	if forUpdate {
		query += " FOR UPDATE"
	}

	account, err := scanAccount(r.tx.Executor(ctx).QueryRowContext(ctx, query, id.String()))
	if err != nil {
		return nil, mapDBError(err)
	}

	return account, nil
}

type accountScanner interface {
	Scan(dest ...any) error
}

func scanAccount(scanner accountScanner) (*domain.Account, error) {
	var account domain.Account
	var balanceRaw string

	err := scanner.Scan(
		&account.ID,
		&account.UserID,
		&account.AccountNo,
		&account.AccountType,
		&account.Currency,
		&balanceRaw,
		&account.Status,
		&account.BlockedReason,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	balance, err := scanMoney(balanceRaw)
	if err != nil {
		return nil, err
	}

	account.Balance = balance

	return &account, nil
}
