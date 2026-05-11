package mfa

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/example/banking-service/internal/domain"
)

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(password string, passwordHash string) error
}

type Service struct {
	users       domain.UserRepository
	challenges  domain.MFARepository
	emailOutbox domain.EmailOutboxRepository
	hasher      PasswordHasher
	ttl         time.Duration
}

func NewService(
	users domain.UserRepository,
	challenges domain.MFARepository,
	emailOutbox domain.EmailOutboxRepository,
	hasher PasswordHasher,
) *Service {
	return &Service{
		users:       users,
		challenges:  challenges,
		emailOutbox: emailOutbox,
		hasher:      hasher,
		ttl:         10 * time.Minute,
	}
}

func (s *Service) CreateChallenge(ctx context.Context, userID domain.UserID, req CreateChallengeRequest) (*ChallengeResponse, error) {
	purpose, err := parsePurpose(req.Purpose)
	if err != nil {
		return nil, err
	}

	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if !user.IsActive() {
		return nil, domain.NewDomainError(domain.ErrorCodeForbidden, "user is not active", domain.ErrForbidden)
	}

	code, err := generateMFACode()
	if err != nil {
		return nil, err
	}

	codeHash, err := s.hasher.Hash(code)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	challenge := &domain.MFAChallenge{
		UserID:          user.ID,
		Purpose:         purpose,
		DeliveryChannel: domain.MFADeliveryChannelEmail,
		Destination:     user.Email,
		CodeHash:        codeHash,
		Context:         req.Context,
		Status:          domain.MFAStatusPending,
		Attempts:        0,
		ExpiresAt:       now.Add(s.ttl),
		CreatedAt:       now,
	}

	if challenge.Context == nil {
		challenge.Context = map[string]any{}
	}

	if err := s.challenges.Create(ctx, challenge); err != nil {
		return nil, err
	}

	if err := s.emailOutbox.Create(
		ctx,
		user.ID,
		user.Email,
		"Banking operation confirmation code",
		"mfa_code",
		map[string]any{
			"code":      code,
			"purpose":   string(purpose),
			"expiresAt": challenge.ExpiresAt,
		},
	); err != nil {
		return nil, err
	}

	return toChallengeResponse(challenge), nil
}

func (s *Service) VerifyChallenge(ctx context.Context, userID domain.UserID, challengeID domain.MFAChallengeID, req VerifyChallengeRequest) (*ChallengeResponse, error) {
	challenge, err := s.challenges.FindByID(ctx, challengeID)
	if err != nil {
		return nil, err
	}

	if challenge.UserID != userID {
		return nil, domain.NewDomainError(domain.ErrorCodeForbidden, "mfa challenge does not belong to current user", domain.ErrForbidden)
	}

	now := time.Now().UTC()

	if challenge.Status != domain.MFAStatusPending {
		return nil, domain.NewDomainError(domain.ErrorCodeInvalidState, "mfa challenge is not pending", domain.ErrInvalidState)
	}

	if challenge.IsExpired(now) {
		challenge.Status = domain.MFAStatusExpired
		_ = s.challenges.Update(ctx, challenge)
		return nil, domain.NewDomainError(domain.ErrorCodeInvalidState, "mfa challenge expired", domain.ErrInvalidState)
	}

	if err := s.hasher.Compare(req.Code, challenge.CodeHash); err != nil {
		challenge.Attempts++

		if challenge.Attempts >= 3 {
			challenge.Status = domain.MFAStatusFailed
		}

		_ = s.challenges.Update(ctx, challenge)

		return nil, domain.NewDomainError(domain.ErrorCodeUnauthorized, "invalid mfa code", domain.ErrUnauthorized)
	}

	challenge.Status = domain.MFAStatusVerified
	challenge.VerifiedAt = &now

	if err := s.challenges.Update(ctx, challenge); err != nil {
		return nil, err
	}

	return toChallengeResponse(challenge), nil
}

func (s *Service) EnsureVerifiedChallenge(ctx context.Context, userID domain.UserID, purpose domain.MFAPurpose, challengeID domain.MFAChallengeID) error {
	challenge, err := s.challenges.FindByID(ctx, challengeID)
	if err != nil {
		return err
	}

	if challenge.UserID != userID {
		return domain.NewDomainError(domain.ErrorCodeForbidden, "mfa challenge does not belong to current user", domain.ErrForbidden)
	}

	if challenge.Purpose != purpose {
		return domain.NewDomainError(domain.ErrorCodeForbidden, "mfa challenge purpose mismatch", domain.ErrForbidden)
	}

	if !challenge.IsVerified() {
		return domain.NewDomainError(domain.ErrorCodeUnauthorized, "verified mfa challenge is required", domain.ErrUnauthorized)
	}

	if challenge.IsExpired(time.Now().UTC()) {
		return domain.NewDomainError(domain.ErrorCodeInvalidState, "mfa challenge expired", domain.ErrInvalidState)
	}

	return nil
}

func parsePurpose(value string) (domain.MFAPurpose, error) {
	switch domain.MFAPurpose(value) {
	case domain.MFAPurposeTransfer,
		domain.MFAPurposeCardPayment,
		domain.MFAPurposeCreditIssue,
		domain.MFAPurposeAccountBlock:
		return domain.MFAPurpose(value), nil
	default:
		return "", domain.NewDomainError(domain.ErrorCodeValidation, "unsupported mfa purpose", domain.ErrValidation)
	}
}

func generateMFACode() (string, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", fmt.Errorf("generate mfa code: %w", err)
	}

	return fmt.Sprintf("%06d", value.Int64()), nil
}

func toChallengeResponse(challenge *domain.MFAChallenge) *ChallengeResponse {
	return &ChallengeResponse{
		ID:              challenge.ID.String(),
		UserID:          challenge.UserID.String(),
		Purpose:         string(challenge.Purpose),
		DeliveryChannel: string(challenge.DeliveryChannel),
		Destination:     challenge.Destination,
		Context:         challenge.Context,
		Status:          string(challenge.Status),
		Attempts:        challenge.Attempts,
		ExpiresAt:       challenge.ExpiresAt,
		VerifiedAt:      challenge.VerifiedAt,
		CreatedAt:       challenge.CreatedAt,
	}
}

func IsMFAError(err error) bool {
	return errors.Is(err, domain.ErrUnauthorized) ||
		errors.Is(err, domain.ErrForbidden) ||
		errors.Is(err, domain.ErrInvalidState)
}
