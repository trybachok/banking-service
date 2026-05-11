package postgres

import (
	"context"

	"github.com/example/banking-service/internal/domain"
	"github.com/example/banking-service/internal/infrastructure/db"
)

type CardRepository struct {
	tx *db.TxManager
}

func NewCardRepository(tx *db.TxManager) *CardRepository {
	return &CardRepository{tx: tx}
}

func (r *CardRepository) Create(ctx context.Context, card *domain.Card) error {
	query := `
		INSERT INTO cards (
			user_id,
			account_id,
			card_alias,
			pan_masked,
			pan_last4,
			pan_encrypted,
			expiry_encrypted,
			pan_hmac,
			integrity_hmac,
			cvv_hash,
			status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`

	err := r.tx.Executor(ctx).QueryRowContext(
		ctx,
		query,
		card.UserID.String(),
		card.AccountID.String(),
		card.CardAlias,
		card.PANMasked,
		card.PANLast4,
		card.PANEncrypted,
		card.ExpiryEncrypted,
		card.PANHMAC,
		card.IntegrityHMAC,
		card.CVVHash,
		string(card.Status),
	).Scan(&card.ID, &card.CreatedAt, &card.UpdatedAt)

	return mapDBError(err)
}

func (r *CardRepository) FindByID(ctx context.Context, id domain.CardID) (*domain.Card, error) {
	query := baseCardSelect() + " WHERE id = $1"

	card, err := scanCard(r.tx.Executor(ctx).QueryRowContext(ctx, query, id.String()))
	if err != nil {
		return nil, mapDBError(err)
	}

	return card, nil
}

func (r *CardRepository) ListByUserID(ctx context.Context, userID domain.UserID) ([]*domain.Card, error) {
	query := baseCardSelect() + " WHERE user_id = $1 ORDER BY created_at DESC"

	rows, err := r.tx.Executor(ctx).QueryContext(ctx, query, userID.String())
	if err != nil {
		return nil, mapDBError(err)
	}
	defer rows.Close()

	var cards []*domain.Card

	for rows.Next() {
		card, err := scanCard(rows)
		if err != nil {
			return nil, err
		}

		cards = append(cards, card)
	}

	return cards, mapDBError(rows.Err())
}

func (r *CardRepository) FindByPANHMAC(ctx context.Context, panHMAC string) (*domain.Card, error) {
	query := baseCardSelect() + " WHERE pan_hmac = $1"

	card, err := scanCard(r.tx.Executor(ctx).QueryRowContext(ctx, query, panHMAC))
	if err != nil {
		return nil, mapDBError(err)
	}

	return card, nil
}

func (r *CardRepository) UpdateStatus(ctx context.Context, id domain.CardID, status domain.CardStatus) error {
	query := `
		UPDATE cards
		SET status = $2
		WHERE id = $1
	`

	result, err := r.tx.Executor(ctx).ExecContext(ctx, query, id.String(), string(status))
	if err != nil {
		return mapDBError(err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return domain.NewDomainError(domain.ErrorCodeNotFound, "card not found", domain.ErrNotFound)
	}

	return nil
}

func baseCardSelect() string {
	return `
		SELECT
			id,
			user_id,
			account_id,
			COALESCE(card_alias, ''),
			pan_masked,
			pan_last4,
			pan_encrypted,
			expiry_encrypted,
			pan_hmac,
			integrity_hmac,
			cvv_hash,
			status,
			created_at,
			updated_at
		FROM cards
	`
}

type cardScanner interface {
	Scan(dest ...any) error
}

func scanCard(scanner cardScanner) (*domain.Card, error) {
	var card domain.Card

	err := scanner.Scan(
		&card.ID,
		&card.UserID,
		&card.AccountID,
		&card.CardAlias,
		&card.PANMasked,
		&card.PANLast4,
		&card.PANEncrypted,
		&card.ExpiryEncrypted,
		&card.PANHMAC,
		&card.IntegrityHMAC,
		&card.CVVHash,
		&card.Status,
		&card.CreatedAt,
		&card.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &card, nil
}
