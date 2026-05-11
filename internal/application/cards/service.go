package cards

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	accountapp "github.com/example/banking-service/internal/application/accounts"
	"github.com/example/banking-service/internal/domain"
)

type CardProtector interface {
	Protect(ctx context.Context, pan string, expiry string, cvv string) (*ProtectedCardData, error)
	Reveal(ctx context.Context, panEncrypted []byte, expiryEncrypted []byte) (*CardPlainData, error)
	VerifyIntegrity(pan string, expiry string, integrityHMAC string) bool
	VerifyCVV(cvv string, cvvHash string) error
	PANHMAC(pan string) string
}

type Service struct {
	tx           domain.TxManager
	accounts     domain.AccountRepository
	cards        domain.CardRepository
	transactions domain.TransactionRepository
	protector    CardProtector
}

func NewService(
	tx domain.TxManager,
	accounts domain.AccountRepository,
	cards domain.CardRepository,
	transactions domain.TransactionRepository,
	protector CardProtector,
) *Service {
	return &Service{
		tx:           tx,
		accounts:     accounts,
		cards:        cards,
		transactions: transactions,
		protector:    protector,
	}
}

func (s *Service) IssueCard(ctx context.Context, userID domain.UserID, req IssueCardRequest) (*CardDetailsResponse, error) {
	accountID := domain.AccountID(req.AccountID)

	account, err := s.accounts.FindByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	if account.UserID != userID {
		return nil, domain.NewDomainError(domain.ErrorCodeForbidden, "account does not belong to current user", domain.ErrForbidden)
	}

	pan, err := s.generateUniquePAN(ctx)
	if err != nil {
		return nil, err
	}

	expiry := generateExpiry()
	cvv, err := generateCVV()
	if err != nil {
		return nil, err
	}

	protected, err := s.protector.Protect(ctx, pan, expiry, cvv)
	if err != nil {
		return nil, err
	}

	card, err := domain.NewCard(domain.CreateCardData{
		UserID:          userID,
		AccountID:       accountID,
		CardAlias:       req.Alias,
		PANMasked:       MaskPAN(pan),
		PANLast4:        Last4(pan),
		PANEncrypted:    protected.PANEncrypted,
		ExpiryEncrypted: protected.ExpiryEncrypted,
		PANHMAC:         protected.PANHMAC,
		IntegrityHMAC:   protected.IntegrityHMAC,
		CVVHash:         protected.CVVHash,
	})
	if err != nil {
		return nil, err
	}

	if err := s.cards.Create(ctx, card); err != nil {
		return nil, err
	}

	return &CardDetailsResponse{
		Card:   ToCardResponse(card),
		PAN:    pan,
		Expiry: expiry,
		CVV:    cvv,
		Notice: "CVV is shown only once and is stored only as bcrypt hash",
	}, nil
}

func (s *Service) ListCards(ctx context.Context, userID domain.UserID) ([]CardResponse, error) {
	cardsList, err := s.cards.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]CardResponse, 0, len(cardsList))
	for _, card := range cardsList {
		result = append(result, ToCardResponse(card))
	}

	return result, nil
}

func (s *Service) GetCardDetails(ctx context.Context, userID domain.UserID, cardID domain.CardID) (*CardDetailsResponse, error) {
	card, err := s.cards.FindByID(ctx, cardID)
	if err != nil {
		return nil, err
	}

	if card.UserID != userID {
		return nil, domain.NewDomainError(domain.ErrorCodeForbidden, "card does not belong to current user", domain.ErrForbidden)
	}

	plain, err := s.protector.Reveal(ctx, card.PANEncrypted, card.ExpiryEncrypted)
	if err != nil {
		return nil, err
	}

	if !s.protector.VerifyIntegrity(plain.PAN, plain.Expiry, card.IntegrityHMAC) {
		return nil, domain.NewDomainError(domain.ErrorCodeInvalidState, "card integrity check failed", domain.ErrInvalidState)
	}

	return &CardDetailsResponse{
		Card:   ToCardResponse(card),
		PAN:    plain.PAN,
		Expiry: plain.Expiry,
		Notice: "CVV cannot be shown because it is stored only as bcrypt hash",
	}, nil
}

func (s *Service) Pay(ctx context.Context, userID domain.UserID, cardID domain.CardID, req CardPaymentRequest) (*CardPaymentResponse, error) {
	amount, err := accountapp.ParseAmountForTransfer(req.Amount)
	if err != nil {
		return nil, err
	}

	var result *CardPaymentResponse

	err = s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		card, err := s.cards.FindByID(txCtx, cardID)
		if err != nil {
			return err
		}

		if card.UserID != userID {
			return domain.NewDomainError(domain.ErrorCodeForbidden, "card does not belong to current user", domain.ErrForbidden)
		}

		if !card.IsActive() {
			return domain.NewDomainError(domain.ErrorCodeInvalidState, "card is not active", domain.ErrInvalidState)
		}

		if err := s.protector.VerifyCVV(req.CVV, card.CVVHash); err != nil {
			return domain.NewDomainError(domain.ErrorCodeUnauthorized, "invalid cvv", domain.ErrUnauthorized)
		}

		account, err := s.accounts.FindByIDForUpdate(txCtx, card.AccountID)
		if err != nil {
			return err
		}

		if account.UserID != userID {
			return domain.NewDomainError(domain.ErrorCodeForbidden, "account does not belong to current user", domain.ErrForbidden)
		}

		if err := account.Debit(amount); err != nil {
			return err
		}

		if err := s.accounts.UpdateBalance(txCtx, account.ID, account); err != nil {
			return err
		}

		sourceAccountID := account.ID
		usedCardID := card.ID

		transaction, err := domain.NewCompletedTransaction(domain.CreateTransactionData{
			UserID:          userID,
			SourceAccountID: &sourceAccountID,
			CardID:          &usedCardID,
			OperationType:   domain.TransactionTypeCardPayment,
			Amount:          amount,
			Description:     req.Description,
			IdempotencyKey:  req.IdempotencyKey,
		})
		if err != nil {
			return err
		}

		if err := s.transactions.Create(txCtx, transaction); err != nil {
			return err
		}

		result = &CardPaymentResponse{
			Card:        ToCardResponse(card),
			Account:     accountapp.ToAccountResponse(account),
			Transaction: accountapp.ToTransactionResponse(transaction),
		}

		return nil
	})

	return result, err
}

func (s *Service) generateUniquePAN(ctx context.Context) (string, error) {
	const maxAttempts = 10

	for attempt := 0; attempt < maxAttempts; attempt++ {
		pan, err := GeneratePAN()
		if err != nil {
			return "", err
		}

		hmacValue := s.protector.PANHMAC(pan)

		_, err = s.cards.FindByPANHMAC(ctx, hmacValue)
		if errors.Is(err, domain.ErrNotFound) {
			return pan, nil
		}

		if err != nil {
			return "", err
		}
	}

	return "", domain.NewDomainError(domain.ErrorCodeConflict, "failed to generate unique card number", domain.ErrConflict)
}

func generateExpiry() string {
	now := time.Now().UTC().AddDate(3, 0, 0)

	return fmt.Sprintf("%02d/%02d", int(now.Month()), now.Year()%100)
}

func generateCVV() (string, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(1000))
	if err != nil {
		return "", fmt.Errorf("generate cvv: %w", err)
	}

	return fmt.Sprintf("%03d", value.Int64()), nil
}
