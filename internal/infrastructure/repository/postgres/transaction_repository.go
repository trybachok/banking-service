package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/example/banking-service/internal/domain"
	"github.com/example/banking-service/internal/infrastructure/db"
)

type TransactionRepository struct {
	tx *db.TxManager
}

func NewTransactionRepository(tx *db.TxManager) *TransactionRepository {
	return &TransactionRepository{tx: tx}
}

func (r *TransactionRepository) Create(ctx context.Context, transaction *domain.Transaction) error {
	query := `
		INSERT INTO transactions (
			user_id,
			source_account_id,
			destination_account_id,
			card_id,
			credit_id,
			operation_type,
			status,
			amount,
			currency,
			description,
			idempotency_key,
			external_reference,
			completed_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'RUB', $9, $10, $11, $12)
		RETURNING id, created_at
	`

	err := r.tx.Executor(ctx).QueryRowContext(
		ctx,
		query,
		transaction.UserID.String(),
		accountIDPtrValue(transaction.SourceAccountID),
		accountIDPtrValue(transaction.DestinationAccountID),
		cardIDPtrValue(transaction.CardID),
		creditIDPtrValue(transaction.CreditID),
		string(transaction.OperationType),
		string(transaction.Status),
		transaction.Amount.String(),
		transaction.Description,
		transaction.IdempotencyKey,
		transaction.ExternalReference,
		transaction.CompletedAt,
	).Scan(&transaction.ID, &transaction.CreatedAt)

	return mapDBError(err)
}

func (r *TransactionRepository) FindByID(ctx context.Context, id domain.TransactionID) (*domain.Transaction, error) {
	query := baseTransactionSelect() + " WHERE id = $1"

	transaction, err := scanTransaction(r.tx.Executor(ctx).QueryRowContext(ctx, query, id.String()))
	if err != nil {
		return nil, mapDBError(err)
	}

	return transaction, nil
}

func (r *TransactionRepository) FindByIdempotencyKey(ctx context.Context, key string) (*domain.Transaction, error) {
	query := baseTransactionSelect() + " WHERE idempotency_key = $1"

	transaction, err := scanTransaction(r.tx.Executor(ctx).QueryRowContext(ctx, query, key))
	if err != nil {
		return nil, mapDBError(err)
	}

	return transaction, nil
}

func (r *TransactionRepository) ListByUserID(ctx context.Context, userID domain.UserID, limit int, offset int) ([]*domain.Transaction, error) {
	query := baseTransactionSelect() + " WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3"

	rows, err := r.tx.Executor(ctx).QueryContext(ctx, query, userID.String(), limit, offset)
	if err != nil {
		return nil, mapDBError(err)
	}
	defer rows.Close()

	return scanTransactions(rows)
}

func (r *TransactionRepository) ListByAccountID(ctx context.Context, accountID domain.AccountID, from time.Time, to time.Time) ([]*domain.Transaction, error) {
	query := baseTransactionSelect() + `
		WHERE (source_account_id = $1 OR destination_account_id = $1)
		  AND created_at >= $2
		  AND created_at <= $3
		ORDER BY created_at DESC
	`

	rows, err := r.tx.Executor(ctx).QueryContext(ctx, query, accountID.String(), from, to)
	if err != nil {
		return nil, mapDBError(err)
	}
	defer rows.Close()

	return scanTransactions(rows)
}

func baseTransactionSelect() string {
	return `
		SELECT
			id,
			user_id,
			source_account_id,
			destination_account_id,
			card_id,
			credit_id,
			operation_type,
			status,
			amount::text,
			currency,
			COALESCE(description, ''),
			COALESCE(idempotency_key, ''),
			COALESCE(external_reference, ''),
			created_at,
			completed_at
		FROM transactions
	`
}

type transactionScanner interface {
	Scan(dest ...any) error
}

func scanTransaction(scanner transactionScanner) (*domain.Transaction, error) {
	var transaction domain.Transaction
	var sourceAccountID sql.NullString
	var destinationAccountID sql.NullString
	var cardID sql.NullString
	var creditID sql.NullString
	var amountRaw string
	var completedAt sql.NullTime

	err := scanner.Scan(
		&transaction.ID,
		&transaction.UserID,
		&sourceAccountID,
		&destinationAccountID,
		&cardID,
		&creditID,
		&transaction.OperationType,
		&transaction.Status,
		&amountRaw,
		&transaction.Currency,
		&transaction.Description,
		&transaction.IdempotencyKey,
		&transaction.ExternalReference,
		&transaction.CreatedAt,
		&completedAt,
	)
	if err != nil {
		return nil, err
	}

	amount, err := scanMoney(amountRaw)
	if err != nil {
		return nil, err
	}

	transaction.SourceAccountID = accountIDPtr(sourceAccountID)
	transaction.DestinationAccountID = accountIDPtr(destinationAccountID)
	transaction.CardID = cardIDPtr(cardID)
	transaction.CreditID = creditIDPtr(creditID)
	transaction.Amount = amount
	transaction.CompletedAt = timePtr(completedAt)

	return &transaction, nil
}

func scanTransactions(rows *sql.Rows) ([]*domain.Transaction, error) {
	var transactions []*domain.Transaction

	for rows.Next() {
		transaction, err := scanTransaction(rows)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, transaction)
	}

	return transactions, mapDBError(rows.Err())
}
