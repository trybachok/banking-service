package auth

import (
	"context"
	"errors"
	"time"

	"github.com/example/banking-service/internal/domain"
)

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(password string, passwordHash string) error
}

type TokenIssuer interface {
	Issue(userID string, role string) (string, time.Time, error)
}

type Service struct {
	users  domain.UserRepository
	hashes PasswordHasher
	tokens TokenIssuer
}

func NewService(
	users domain.UserRepository,
	hashes PasswordHasher,
	tokens TokenIssuer,
) *Service {
	return &Service{
		users:  users,
		hashes: hashes,
		tokens: tokens,
	}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	exists, err := s.users.ExistsByEmailOrUsername(ctx, req.Email, req.Username)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, domain.NewDomainError(domain.ErrorCodeConflict, "email or username already exists", domain.ErrConflict)
	}

	passwordHash, err := s.hashes.Hash(req.Password)
	if err != nil {
		return nil, err
	}

	user, err := domain.NewUser(domain.RegisterUserData{
		Email:        req.Email,
		Username:     req.Username,
		Password:     req.Password,
		PasswordHash: passwordHash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
	})
	if err != nil {
		return nil, err
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	return s.issueAuthResponse(user)
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	user, err := s.users.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.NewDomainError(domain.ErrorCodeUnauthorized, "invalid email or password", domain.ErrUnauthorized)
		}

		return nil, err
	}

	if !user.IsActive() {
		return nil, domain.NewDomainError(domain.ErrorCodeForbidden, "user is not active", domain.ErrForbidden)
	}

	if err := s.hashes.Compare(req.Password, user.PasswordHash); err != nil {
		return nil, domain.NewDomainError(domain.ErrorCodeUnauthorized, "invalid email or password", domain.ErrUnauthorized)
	}

	now := time.Now().UTC()
	if err := s.users.UpdateLastLoginAt(ctx, user.ID, now); err != nil {
		return nil, err
	}

	user.LastLoginAt = &now

	return s.issueAuthResponse(user)
}

func (s *Service) issueAuthResponse(user *domain.User) (*AuthResponse, error) {
	token, expiresAt, err := s.tokens.Issue(user.ID.String(), string(user.Role))
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		AccessToken:      token,
		TokenType:        "Bearer",
		ExpiresAt:        expiresAt,
		ExpiresInSeconds: int64(time.Until(expiresAt).Seconds()),
		User: UserResponse{
			ID:        user.ID.String(),
			Email:     user.Email,
			Username:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      string(user.Role),
			Status:    string(user.Status),
		},
	}, nil
}
