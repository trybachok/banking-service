package admin

import (
	"context"
	"strings"

	"github.com/example/banking-service/internal/domain"
)

type Service struct {
	users    domain.UserRepository
	accounts domain.AccountRepository
}

func NewService(users domain.UserRepository, accounts domain.AccountRepository) *Service {
	return &Service{
		users:    users,
		accounts: accounts,
	}
}

func (s *Service) ListUsers(ctx context.Context, limit int, offset int) ([]UserResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	if offset < 0 {
		offset = 0
	}

	users, err := s.users.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	result := make([]UserResponse, 0, len(users))
	for _, user := range users {
		result = append(result, UserResponse{
			ID:          user.ID.String(),
			Email:       user.Email,
			Username:    user.Username,
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			Role:        string(user.Role),
			Status:      string(user.Status),
			LastLoginAt: user.LastLoginAt,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		})
	}

	return result, nil
}

func (s *Service) BlockAccount(ctx context.Context, accountID domain.AccountID, reason string) (*BlockAccountResponse, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "Blocked by administrator"
	}

	if err := s.accounts.UpdateStatus(ctx, accountID, domain.AccountStatusBlocked, reason); err != nil {
		return nil, err
	}

	return &BlockAccountResponse{
		AccountID: accountID.String(),
		Status:    string(domain.AccountStatusBlocked),
		Reason:    reason,
	}, nil
}
