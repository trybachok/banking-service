package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/banking-service/internal/domain"
)

func TestServiceRegister(t *testing.T) {
	repository := newFakeUserRepository()
	service := NewService(repository, fakePasswordHasher{}, fakeTokenIssuer{})

	result, err := service.Register(context.Background(), RegisterRequest{
		Email:     "user@example.com",
		Username:  "user_name",
		Password:  "strong-password",
		FirstName: "Test",
		LastName:  "User",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.AccessToken == "" {
		t.Fatal("expected access token")
	}

	if result.User.ID == "" {
		t.Fatal("expected user id")
	}

	if result.User.Email != "user@example.com" {
		t.Fatalf("unexpected email: %s", result.User.Email)
	}
}

func TestServiceRegisterRejectsDuplicateUser(t *testing.T) {
	repository := newFakeUserRepository()
	service := NewService(repository, fakePasswordHasher{}, fakeTokenIssuer{})

	req := RegisterRequest{
		Email:    "user@example.com",
		Username: "user_name",
		Password: "strong-password",
	}

	if _, err := service.Register(context.Background(), req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := service.Register(context.Background(), req); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected conflict error, got: %v", err)
	}
}

func TestServiceLogin(t *testing.T) {
	repository := newFakeUserRepository()
	service := NewService(repository, fakePasswordHasher{}, fakeTokenIssuer{})

	_, err := service.Register(context.Background(), RegisterRequest{
		Email:    "user@example.com",
		Username: "user_name",
		Password: "strong-password",
	})
	if err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	result, err := service.Login(context.Background(), LoginRequest{
		Email:    "user@example.com",
		Password: "strong-password",
	})
	if err != nil {
		t.Fatalf("unexpected login error: %v", err)
	}

	if result.AccessToken == "" {
		t.Fatal("expected access token")
	}
}

func TestServiceLoginRejectsWrongPassword(t *testing.T) {
	repository := newFakeUserRepository()
	service := NewService(repository, fakePasswordHasher{}, fakeTokenIssuer{})

	_, err := service.Register(context.Background(), RegisterRequest{
		Email:    "user@example.com",
		Username: "user_name",
		Password: "strong-password",
	})
	if err != nil {
		t.Fatalf("unexpected register error: %v", err)
	}

	_, err = service.Login(context.Background(), LoginRequest{
		Email:    "user@example.com",
		Password: "wrong-password",
	})

	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected unauthorized error, got: %v", err)
	}
}

type fakePasswordHasher struct{}

func (h fakePasswordHasher) Hash(password string) (string, error) {
	return "hash:" + password, nil
}

func (h fakePasswordHasher) Compare(password string, passwordHash string) error {
	if passwordHash != "hash:"+password {
		return errors.New("password mismatch")
	}

	return nil
}

type fakeTokenIssuer struct{}

func (t fakeTokenIssuer) Issue(userID string, role string) (string, time.Time, error) {
	return "token:" + userID + ":" + role, time.Now().UTC().Add(24 * time.Hour), nil
}

type fakeUserRepository struct {
	usersByID       map[domain.UserID]*domain.User
	usersByEmail    map[string]*domain.User
	usersByUsername map[string]*domain.User
	nextID          int
}

func newFakeUserRepository() *fakeUserRepository {
	return &fakeUserRepository{
		usersByID:       map[domain.UserID]*domain.User{},
		usersByEmail:    map[string]*domain.User{},
		usersByUsername: map[string]*domain.User{},
		nextID:          1,
	}
}

func (r *fakeUserRepository) Create(_ context.Context, user *domain.User) error {
	user.ID = domain.UserID("user-1")
	r.usersByID[user.ID] = user
	r.usersByEmail[user.Email] = user
	r.usersByUsername[user.Username] = user
	r.nextID++

	return nil
}

func (r *fakeUserRepository) FindByID(_ context.Context, id domain.UserID) (*domain.User, error) {
	user, ok := r.usersByID[id]
	if !ok {
		return nil, domain.NewDomainError(domain.ErrorCodeNotFound, "user not found", domain.ErrNotFound)
	}

	return user, nil
}

func (r *fakeUserRepository) FindByEmail(_ context.Context, email string) (*domain.User, error) {
	user, ok := r.usersByEmail[email]
	if !ok {
		return nil, domain.NewDomainError(domain.ErrorCodeNotFound, "user not found", domain.ErrNotFound)
	}

	return user, nil
}

func (r *fakeUserRepository) FindByUsername(_ context.Context, username string) (*domain.User, error) {
	user, ok := r.usersByUsername[username]
	if !ok {
		return nil, domain.NewDomainError(domain.ErrorCodeNotFound, "user not found", domain.ErrNotFound)
	}

	return user, nil
}

func (r *fakeUserRepository) ExistsByEmailOrUsername(_ context.Context, email string, username string) (bool, error) {
	if _, ok := r.usersByEmail[email]; ok {
		return true, nil
	}

	if _, ok := r.usersByUsername[username]; ok {
		return true, nil
	}

	return false, nil
}

func (r *fakeUserRepository) UpdateLastLoginAt(_ context.Context, id domain.UserID, at time.Time) error {
	user, ok := r.usersByID[id]
	if !ok {
		return domain.NewDomainError(domain.ErrorCodeNotFound, "user not found", domain.ErrNotFound)
	}

	user.LastLoginAt = &at

	return nil
}

func (r *fakeUserRepository) List(_ context.Context, _ int, _ int) ([]*domain.User, error) {
	users := make([]*domain.User, 0, len(r.usersByID))

	for _, user := range r.usersByID {
		users = append(users, user)
	}

	return users, nil
}
