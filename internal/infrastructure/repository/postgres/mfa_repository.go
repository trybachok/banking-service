package postgres

import (
	"context"
	"database/sql"

	"github.com/example/banking-service/internal/domain"
	"github.com/example/banking-service/internal/infrastructure/db"
)

type MFARepository struct {
	tx *db.TxManager
}

func NewMFARepository(tx *db.TxManager) *MFARepository {
	return &MFARepository{tx: tx}
}

func (r *MFARepository) Create(ctx context.Context, challenge *domain.MFAChallenge) error {
	contextRaw, err := marshalJSONMap(challenge.Context)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO mfa_challenges (
			user_id,
			purpose,
			delivery_channel,
			destination,
			code_hash,
			context,
			status,
			attempts,
			expires_at
		)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9)
		RETURNING id, created_at
	`

	err = r.tx.Executor(ctx).QueryRowContext(
		ctx,
		query,
		challenge.UserID.String(),
		string(challenge.Purpose),
		string(challenge.DeliveryChannel),
		challenge.Destination,
		challenge.CodeHash,
		string(contextRaw),
		string(challenge.Status),
		challenge.Attempts,
		challenge.ExpiresAt,
	).Scan(&challenge.ID, &challenge.CreatedAt)

	return mapDBError(err)
}

func (r *MFARepository) FindByID(ctx context.Context, id domain.MFAChallengeID) (*domain.MFAChallenge, error) {
	query := baseMFASelect() + " WHERE id = $1"

	challenge, err := scanMFAChallenge(r.tx.Executor(ctx).QueryRowContext(ctx, query, id.String()))
	if err != nil {
		return nil, mapDBError(err)
	}

	return challenge, nil
}

func (r *MFARepository) Update(ctx context.Context, challenge *domain.MFAChallenge) error {
	contextRaw, err := marshalJSONMap(challenge.Context)
	if err != nil {
		return err
	}

	query := `
		UPDATE mfa_challenges
		SET context = $2::jsonb,
		    status = $3,
		    attempts = $4,
		    verified_at = $5
		WHERE id = $1
	`

	result, err := r.tx.Executor(ctx).ExecContext(
		ctx,
		query,
		challenge.ID.String(),
		string(contextRaw),
		string(challenge.Status),
		challenge.Attempts,
		challenge.VerifiedAt,
	)
	if err != nil {
		return mapDBError(err)
	}

	return ensureRowsAffected(result, "mfa challenge not found")
}

func baseMFASelect() string {
	return `
		SELECT
			id,
			user_id,
			purpose,
			delivery_channel,
			destination,
			code_hash,
			context,
			status,
			attempts,
			expires_at,
			verified_at,
			created_at
		FROM mfa_challenges
	`
}

func scanMFAChallenge(scanner interface{ Scan(dest ...any) error }) (*domain.MFAChallenge, error) {
	var challenge domain.MFAChallenge
	var contextRaw []byte
	var verifiedAt sql.NullTime

	err := scanner.Scan(
		&challenge.ID,
		&challenge.UserID,
		&challenge.Purpose,
		&challenge.DeliveryChannel,
		&challenge.Destination,
		&challenge.CodeHash,
		&contextRaw,
		&challenge.Status,
		&challenge.Attempts,
		&challenge.ExpiresAt,
		&verifiedAt,
		&challenge.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	contextMap, err := unmarshalJSONMap(contextRaw)
	if err != nil {
		return nil, err
	}

	challenge.Context = contextMap
	challenge.VerifiedAt = timePtr(verifiedAt)

	return &challenge, nil
}
