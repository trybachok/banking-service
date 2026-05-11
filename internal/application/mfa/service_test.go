package mfa

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/banking-service/internal/domain"
)

func TestServiceCreateAndVerifyChallenge(t *testing.T) {
	users := newFakeUserRepository()
	challenges := newFakeMFARepository()
	outbox := newFakeEmailOutboxRepository()
	hasher := newFakePasswordHasher()

	user := &domain.User{
		ID:       domain.UserID("user-1"),
		Email:    "user@example.com",
		Username: "user",
		Role:     domain.UserRoleCustomer,
		Status:   domain.UserStatusActive,
	}
	users.users[user.ID] = user

	service := NewService(users, challenges, outbox, hasher)

	created, err := service.CreateChallenge(context.Background(), user.ID, CreateChallengeRequest{
		Purpose: string(domain.MFAPurposeTransfer),
	})
	if err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}

	if created.ID == "" {
		t.Fatal("expected challenge id")
	}

	if len(outbox.messages) != 1 {
		t.Fatalf("expected 1 email outbox message, got %d", len(outbox.messages))
	}

	code, _ := outbox.messages[0].Payload["code"].(string)

	verified, err := service.VerifyChallenge(context.Background(), user.ID, domain.MFAChallengeID(created.ID), VerifyChallengeRequest{
		Code: code,
	})
	if err != nil {
		t.Fatalf("unexpected verify error: %v", err)
	}

	if verified.Status != string(domain.MFAStatusVerified) {
		t.Fatalf("expected verified status, got %s", verified.Status)
	}

	err = service.EnsureVerifiedChallenge(context.Background(), user.ID, domain.MFAPurposeTransfer, domain.MFAChallengeID(created.ID))
	if err != nil {
		t.Fatalf("expected verified challenge to pass: %v", err)
	}
}

func TestServiceVerifyRejectsWrongCode(t *testing.T) {
	users := newFakeUserRepository()
	challenges := newFakeMFARepository()
	outbox := newFakeEmailOutboxRepository()
	hasher := newFakePasswordHasher()

	user := &domain.User{
		ID:       domain.UserID("user-1"),
		Email:    "user@example.com",
		Username: "user",
		Role:     domain.UserRoleCustomer,
		Status:   domain.UserStatusActive,
	}
	users.users[user.ID] = user

	service := NewService(users, challenges, outbox, hasher)

	created, err := service.CreateChallenge(context.Background(), user.ID, CreateChallengeRequest{
		Purpose: string(domain.MFAPurposeTransfer),
	})
	if err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}

	_, err = service.VerifyChallenge(context.Background(), user.ID, domain.MFAChallengeID(created.ID), VerifyChallengeRequest{
		Code: "000000",
	})

	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected unauthorized error, got %v", err)
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

func newFakePasswordHasher() fakePasswordHasher {
	return fakePasswordHasher{}
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
	user, ok := r.users[id]
	if !ok {
		return nil, domain.NewDomainError(domain.ErrorCodeNotFound, "user not found", domain.ErrNotFound)
	}
	return user, nil
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
	user, ok := r.users[id]
	if !ok {
		return domain.NewDomainError(domain.ErrorCodeNotFound, "user not found", domain.ErrNotFound)
	}
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

type fakeMFARepository struct {
	challenges map[domain.MFAChallengeID]*domain.MFAChallenge
	next       int
}

func newFakeMFARepository() *fakeMFARepository {
	return &fakeMFARepository{
		challenges: map[domain.MFAChallengeID]*domain.MFAChallenge{},
		next:       1,
	}
}

func (r *fakeMFARepository) Create(_ context.Context, challenge *domain.MFAChallenge) error {
	challenge.ID = domain.MFAChallengeID("mfa-" + string(rune('0'+r.next)))
	r.next++
	r.challenges[challenge.ID] = challenge
	return nil
}

func (r *fakeMFARepository) FindByID(_ context.Context, id domain.MFAChallengeID) (*domain.MFAChallenge, error) {
	challenge, ok := r.challenges[id]
	if !ok {
		return nil, domain.NewDomainError(domain.ErrorCodeNotFound, "mfa challenge not found", domain.ErrNotFound)
	}
	return challenge, nil
}

func (r *fakeMFARepository) Update(_ context.Context, challenge *domain.MFAChallenge) error {
	r.challenges[challenge.ID] = challenge
	return nil
}

type fakeEmailOutboxRepository struct {
	messages []domain.EmailOutboxMessage
}

func newFakeEmailOutboxRepository() *fakeEmailOutboxRepository {
	return &fakeEmailOutboxRepository{messages: []domain.EmailOutboxMessage{}}
}

func (r *fakeEmailOutboxRepository) Create(_ context.Context, userID domain.UserID, recipientEmail string, subject string, templateCode string, payload map[string]any) error {
	r.messages = append(r.messages, domain.EmailOutboxMessage{
		ID:             "email-1",
		UserID:         &userID,
		RecipientEmail: recipientEmail,
		Subject:        subject,
		TemplateCode:   templateCode,
		Payload:        payload,
		CreatedAt:      time.Now().UTC(),
	})
	return nil
}

func (r *fakeEmailOutboxRepository) FindPending(_ context.Context, _ time.Time, _ int) ([]domain.EmailOutboxMessage, error) {
	return r.messages, nil
}

func (r *fakeEmailOutboxRepository) MarkSent(_ context.Context, _ string, _ time.Time) error {
	return nil
}

func (r *fakeEmailOutboxRepository) MarkFailed(_ context.Context, _ string, _ time.Time, _ string) error {
	return nil
}
