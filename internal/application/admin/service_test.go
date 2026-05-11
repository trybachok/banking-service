package admin

import (
	"context"
	"testing"
	"time"

	"github.com/example/banking-service/internal/domain"
)

func TestServiceListUsers(t *testing.T) {
	users := newFakeUserRepository()
	accounts := newFakeAccountRepository()

	users.users[domain.UserID("user-1")] = &domain.User{
		ID:        domain.UserID("user-1"),
		Email:     "user@example.com",
		Username:  "user",
		Role:      domain.UserRoleCustomer,
		Status:    domain.UserStatusActive,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	service := NewService(users, accounts)

	result, err := service.ListUsers(context.Background(), 50, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 user, got %d", len(result))
	}
}

func TestServiceBlockAccount(t *testing.T) {
	users := newFakeUserRepository()
	accounts := newFakeAccountRepository()

	account := &domain.Account{
		ID:     domain.AccountID("account-1"),
		UserID: domain.UserID("user-1"),
		Status: domain.AccountStatusActive,
	}

	accounts.accounts[account.ID] = account

	service := NewService(users, accounts)

	result, err := service.BlockAccount(context.Background(), account.ID, "Fraud check")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != string(domain.AccountStatusBlocked) {
		t.Fatalf("expected blocked status, got %s", result.Status)
	}

	if accounts.accounts[account.ID].Status != domain.AccountStatusBlocked {
		t.Fatal("expected account to be blocked")
	}
}

type fakeUserRepository struct {
	users map[domain.UserID]*domain.User
}

func newFakeUserRepository() *fakeUserRepository {
	return &fakeUserRepository{users: map[domain.UserID]*domain.User{}}
}

func (r *fakeUserRepository) Create(_ context.Context, user *domain.User) error {
	r.users[user.ID] = user
	return nil
}

func (r *fakeUserRepository) FindByID(_ context.Context, id domain.UserID) (*domain.User, error) {
	return r.users[id], nil
}

func (r *fakeUserRepository) FindByEmail(_ context.Context, email string) (*domain.User, error) {
	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, domain.NewDomainError(domain.ErrorCodeNotFound, "user not found", domain.ErrNotFound)
}

func (r *fakeUserRepository) FindByUsername(_ context.Context, username string) (*domain.User, error) {
	for _, user := range r.users {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, domain.NewDomainError(domain.ErrorCodeNotFound, "user not found", domain.ErrNotFound)
}

func (r *fakeUserRepository) ExistsByEmailOrUsername(_ context.Context, email string, username string) (bool, error) {
	for _, user := range r.users {
		if user.Email == email || user.Username == username {
			return true, nil
		}
	}
	return false, nil
}

func (r *fakeUserRepository) UpdateLastLoginAt(_ context.Context, id domain.UserID, at time.Time) error {
	user := r.users[id]
	user.LastLoginAt = &at
	return nil
}

func (r *fakeUserRepository) List(_ context.Context, _ int, _ int) ([]*domain.User, error) {
	result := []*domain.User{}
	for _, user := range r.users {
		result = append(result, user)
	}
	return result, nil
}

type fakeAccountRepository struct {
	accounts map[domain.AccountID]*domain.Account
}

func newFakeAccountRepository() *fakeAccountRepository {
	return &fakeAccountRepository{accounts: map[domain.AccountID]*domain.Account{}}
}

func (r *fakeAccountRepository) Create(_ context.Context, account *domain.Account) error {
	r.accounts[account.ID] = account
	return nil
}

func (r *fakeAccountRepository) FindByID(_ context.Context, id domain.AccountID) (*domain.Account, error) {
	return r.accounts[id], nil
}

func (r *fakeAccountRepository) FindByIDForUpdate(ctx context.Context, id domain.AccountID) (*domain.Account, error) {
	return r.FindByID(ctx, id)
}

func (r *fakeAccountRepository) ListByUserID(_ context.Context, userID domain.UserID) ([]*domain.Account, error) {
	result := []*domain.Account{}
	for _, account := range r.accounts {
		if account.UserID == userID {
			result = append(result, account)
		}
	}
	return result, nil
}

func (r *fakeAccountRepository) UpdateBalance(_ context.Context, id domain.AccountID, account *domain.Account) error {
	r.accounts[id] = account
	return nil
}

func (r *fakeAccountRepository) UpdateStatus(_ context.Context, id domain.AccountID, status domain.AccountStatus, reason string) error {
	account := r.accounts[id]
	account.Status = status
	account.BlockedReason = reason
	return nil
}
