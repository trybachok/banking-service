package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/example/banking-service/internal/domain"
	"github.com/example/banking-service/internal/infrastructure/db"
)

type EmailOutboxRepository struct {
	tx *db.TxManager
}

func NewEmailOutboxRepository(tx *db.TxManager) *EmailOutboxRepository {
	return &EmailOutboxRepository{tx: tx}
}

func (r *EmailOutboxRepository) Create(
	ctx context.Context,
	userID domain.UserID,
	recipientEmail string,
	subject string,
	templateCode string,
	payload map[string]any,
) error {
	payloadRaw, err := marshalJSONMap(payload)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO email_outbox (
			user_id,
			recipient_email,
			subject,
			template_code,
			payload,
			status
		)
		VALUES ($1, $2, $3, $4, $5::jsonb, 'pending')
	`

	_, err = r.tx.Executor(ctx).ExecContext(
		ctx,
		query,
		userID.String(),
		recipientEmail,
		subject,
		templateCode,
		string(payloadRaw),
	)

	return mapDBError(err)
}

func (r *EmailOutboxRepository) FindPending(ctx context.Context, now time.Time, limit int) ([]domain.EmailOutboxMessage, error) {
	query := `
		SELECT
			id,
			user_id,
			recipient_email,
			subject,
			template_code,
			payload,
			attempts,
			created_at
		FROM email_outbox
		WHERE status = 'pending'
		  AND next_attempt_at <= $1
		ORDER BY created_at ASC
		LIMIT $2
	`

	rows, err := r.tx.Executor(ctx).QueryContext(ctx, query, now, limit)
	if err != nil {
		return nil, mapDBError(err)
	}
	defer rows.Close()

	var messages []domain.EmailOutboxMessage

	for rows.Next() {
		message, err := scanEmailOutboxMessage(rows)
		if err != nil {
			return nil, err
		}

		messages = append(messages, message)
	}

	return messages, mapDBError(rows.Err())
}

func (r *EmailOutboxRepository) MarkSent(ctx context.Context, id string, sentAt time.Time) error {
	query := `
		UPDATE email_outbox
		SET status = 'sent',
		    sent_at = $2
		WHERE id = $1
	`

	result, err := r.tx.Executor(ctx).ExecContext(ctx, query, id, sentAt)
	if err != nil {
		return mapDBError(err)
	}

	return ensureRowsAffected(result, "email outbox message not found")
}

func (r *EmailOutboxRepository) MarkFailed(ctx context.Context, id string, nextAttemptAt time.Time, lastError string) error {
	query := `
		UPDATE email_outbox
		SET status = CASE
		        WHEN attempts + 1 >= 3 THEN 'failed'
		        ELSE 'pending'
		    END,
		    attempts = attempts + 1,
		    next_attempt_at = $2,
		    last_error = $3
		WHERE id = $1
	`

	result, err := r.tx.Executor(ctx).ExecContext(ctx, query, id, nextAttemptAt, lastError)
	if err != nil {
		return mapDBError(err)
	}

	return ensureRowsAffected(result, "email outbox message not found")
}

func scanEmailOutboxMessage(scanner interface{ Scan(dest ...any) error }) (domain.EmailOutboxMessage, error) {
	var message domain.EmailOutboxMessage
	var userID sql.NullString
	var payloadRaw []byte

	err := scanner.Scan(
		&message.ID,
		&userID,
		&message.RecipientEmail,
		&message.Subject,
		&message.TemplateCode,
		&payloadRaw,
		&message.Attempts,
		&message.CreatedAt,
	)
	if err != nil {
		return message, err
	}

	payload, err := unmarshalJSONMap(payloadRaw)
	if err != nil {
		return message, err
	}

	message.UserID = userIDPtr(userID)
	message.Payload = payload

	return message, nil
}

func ensureRowsAffected(result sql.Result, message string) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return domain.NewDomainError(domain.ErrorCodeNotFound, message, domain.ErrNotFound)
	}

	return nil
}
