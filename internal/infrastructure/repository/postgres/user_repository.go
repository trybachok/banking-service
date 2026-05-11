package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/example/banking-service/internal/domain"
	"github.com/example/banking-service/internal/infrastructure/db"
)

type UserRepository struct {
	tx *db.TxManager
}

func NewUserRepository(tx *db.TxManager) *UserRepository {
	return &UserRepository{tx: tx}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (
			email,
			username,
			password_hash,
			first_name,
			last_name,
			role,
			status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	err := r.tx.Executor(ctx).QueryRowContext(
		ctx,
		query,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		string(user.Role),
		string(user.Status),
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	return mapDBError(err)
}

func (r *UserRepository) FindByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	return r.findOne(ctx, "WHERE id = $1", id.String())
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.findOne(ctx, "WHERE email_normalized = lower($1)", email)
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	return r.findOne(ctx, "WHERE username_normalized = lower($1)", username)
}

func (r *UserRepository) ExistsByEmailOrUsername(ctx context.Context, email string, username string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM users
			WHERE email_normalized = lower($1)
			   OR username_normalized = lower($2)
		)
	`

	var exists bool
	err := r.tx.Executor(ctx).QueryRowContext(ctx, query, email, username).Scan(&exists)

	return exists, mapDBError(err)
}

func (r *UserRepository) UpdateLastLoginAt(ctx context.Context, id domain.UserID, at time.Time) error {
	query := `
		UPDATE users
		SET last_login_at = $2
		WHERE id = $1
	`

	result, err := r.tx.Executor(ctx).ExecContext(ctx, query, id.String(), at)
	if err != nil {
		return mapDBError(err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return domain.NewDomainError(domain.ErrorCodeNotFound, "user not found", domain.ErrNotFound)
	}

	return nil
}

func (r *UserRepository) List(ctx context.Context, limit int, offset int) ([]*domain.User, error) {
	query := `
		SELECT
			id,
			email,
			username,
			password_hash,
			COALESCE(first_name, ''),
			COALESCE(last_name, ''),
			role,
			status,
			last_login_at,
			created_at,
			updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.tx.Executor(ctx).QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, mapDBError(err)
	}
	defer rows.Close()

	var users []*domain.User

	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, mapDBError(rows.Err())
}

func (r *UserRepository) findOne(ctx context.Context, where string, arg any) (*domain.User, error) {
	query := `
		SELECT
			id,
			email,
			username,
			password_hash,
			COALESCE(first_name, ''),
			COALESCE(last_name, ''),
			role,
			status,
			last_login_at,
			created_at,
			updated_at
		FROM users
	` + where

	user, err := scanUser(r.tx.Executor(ctx).QueryRowContext(ctx, query, arg))
	if err != nil {
		return nil, mapDBError(err)
	}

	return user, nil
}

type userScanner interface {
	Scan(dest ...any) error
}

func scanUser(scanner userScanner) (*domain.User, error) {
	var user domain.User
	var lastLoginAt sql.NullTime

	err := scanner.Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.Status,
		&lastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	user.LastLoginAt = timePtr(lastLoginAt)

	return &user, nil
}
