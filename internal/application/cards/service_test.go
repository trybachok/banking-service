package cards

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/banking-service/internal/domain"
	"github.com/example/banking-service/pkg/money"
)

func TestServiceIssueCardAndGetDetails(t *testing.T) {
	accountRepository := newFakeAccountRepository()
	cardRepository := newFakeCardRepository()
	transactionRepository := newFakeTransactionRepository()

	account, err := domain.NewAccount(domain.UserID("user-1"), "40817810000000000001", domain.AccountTypeCurrent)
	if err != nil {
		t.Fatalf("unexpected account error: %v", err)
	}

	account.ID = domain.AccountID("account-1")
	accountRepository.accounts[account.ID] = account

	service := NewService(fakeTxManager{}, accountRepository, cardRepository, transactionRepository, fakeCardProtector{})

	issued, err := service.IssueCard(context.Background(), domain.UserID("user-1"), IssueCardRequest{
		AccountID: "account-1",
		Alias:     "Main card",
	})
	if err != nil {
		t.Fatalf("unexpected issue error: %v", err)
	}

	if issued.PAN == "" {
		t.Fatal("expected PAN")
	}

	if issued.CVV == "" {
		t.Fatal("expected CVV")
	}

	details, err := service.GetCardDetails(context.Background(), domain.UserID("user-1"), domain.CardID(issued.Card.ID))
	if err != nil {
		t.Fatalf("unexpected details error: %v", err)
	}

	if details.PAN == "" {
		t.Fatal("expected decrypted PAN")
	}
}

func TestServicePay(t *testing.T) {
	accountRepository := newFakeAccountRepository()
	cardRepository := newFakeCardRepository()
	transactionRepository := newFakeTransactionRepository()

	initialBalance, err := money.FromRubString("100.00")
	if err != nil {
		t.Fatalf("unexpected money error: %v", err)
	}

	account, err := domain.NewAccount(domain.UserID("user-1"), "40817810000000000001", domain.AccountTypeCurrent)
	if err != nil {
		t.Fatalf("unexpected account error: %v", err)
	}

	account.ID = domain.AccountID("account-1")
	account.Balance = initialBalance
	accountRepository.accounts[account.ID] = account

	service := NewService(fakeTxManager{}, accountRepository, cardRepository, transactionRepository, fakeCardProtector{})

	issued, err := service.IssueCard(context.Background(), domain.UserID("user-1"), IssueCardRequest{
		AccountID: "account-1",
	})
	if err != nil {
		t.Fatalf("unexpected issue error: %v", err)
	}

	result, err := service.Pay(context.Background(), domain.UserID("user-1"), domain.CardID(issued.Card.ID), CardPaymentRequest{
		Amount: "30.00",
		CVV:    issued.CVV,
	})
	if err != nil {
		t.Fatalf("unexpected pay error: %v", err)
	}

	if result.Account.Balance != "70.00" {
		t.Fatalf("expected balance 70.00, got %s", result.Account.Balance)
	}
}

func TestServicePayRejectsWrongCVV(t *testing.T) {
	accountRepository := newFakeAccountRepository()
	cardRepository := newFakeCardRepository()
	transactionRepository := newFakeTransactionRepository()

	account, err := domain.NewAccount(domain.UserID("user-1"), "40817810000000000001", domain.AccountTypeCurrent)
	if err != nil {
		t.Fatalf("unexpected account error: %v", err)
	}

	account.ID = domain.AccountID("account-1")
	accountRepository.accounts[account.ID] = account

	service := NewService(fakeTxManager{}, accountRepository, cardRepository, transactionRepository, fakeCardProtector{})

	issued, err := service.IssueCard(context.Background(), domain.UserID("user-1"), IssueCardRequest{
		AccountID: "account-1",
	})
	if err != nil {
		t.Fatalf("unexpected issue error: %v", err)
	}

	_, err = service.Pay(context.Background(), domain.UserID("user-1"), domain.CardID(issued.Card.ID), CardPaymentRequest{
		Amount: "10.00",
		CVV:    "999",
	})

	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected unauthorized error, got %v", err)
	}
}

type fakeCardProtector struct{}

func (p fakeCardProtector) Protect(_ context.Context, pan string, expiry string, cvv string) (*ProtectedCardData, error) {
	return &ProtectedCardData{
		PANEncrypted:    []byte(pan),
		ExpiryEncrypted: []byte(expiry),
		PANHMAC:         "hmac:" + pan,
		IntegrityHMAC:   "integrity:" + pan + ":" + expiry,
		CVVHash:         "cvv:" + cvv,
	}, nil
}

func (p fakeCardProtector) Reveal(_ context.Context, panEncrypted []byte, expiryEncrypted []byte) (*CardPlainData, error) {
	return &CardPlainData{
		PAN:    string(panEncrypted),
		Expiry: string(expiryEncrypted),
	}, nil
}

func (p fakeCardProtector) VerifyIntegrity(pan string, expiry string, integrityHMAC string) bool {
	return integrityHMAC == "integrity:"+pan+":"+expiry
}

func (p fakeCardProtector) VerifyCVV(cvv string, cvvHash string) error {
	if cvvHash != "cvv:"+cvv {
		return errors.New("invalid cvv")
	}

	return nil
}

func (p fakeCardProtector) PANHMAC(pan string) string {
	return "hmac:" + pan
}

type fakeTxManager struct{}

func (m fakeTxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type fakeAccountRepository struct {
	accounts map[domain.AccountID]*domain.Account
}

func newFakeAccountRepository() *fakeAccountRepository {
	return &fakeAccountRepository{
		accounts: map[domain.AccountID]*domain.Account{},
	}
}

func (r *fakeAccountRepository) Create(_ context.Context, account *domain.Account) error {
	r.accounts[account.ID] = account
	return nil
}

func (r *fakeAccountRepository) FindByID(_ context.Context, id domain.AccountID) (*domain.Account, error) {
	account, ok := r.accounts[id]
	if !ok {
		return nil, domain.NewDomainError(domain.ErrorCodeNotFound, "account not found", domain.ErrNotFound)
	}

	return account, nil
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
	account, ok := r.accounts[id]
	if !ok {
		return domain.NewDomainError(domain.ErrorCodeNotFound, "account not found", domain.ErrNotFound)
	}

	account.Status = status
	account.BlockedReason = reason

	return nil
}

type fakeCardRepository struct {
	cards map[domain.CardID]*domain.Card
	next  int
}

func newFakeCardRepository() *fakeCardRepository {
	return &fakeCardRepository{
		cards: map[domain.CardID]*domain.Card{},
		next:  1,
	}
}

func (r *fakeCardRepository) Create(_ context.Context, card *domain.Card) error {
	card.ID = domain.CardID("card-" + string(rune('0'+r.next)))
	r.next++
	r.cards[card.ID] = card
	return nil
}

func (r *fakeCardRepository) FindByID(_ context.Context, id domain.CardID) (*domain.Card, error) {
	card, ok := r.cards[id]
	if !ok {
		return nil, domain.NewDomainError(domain.ErrorCodeNotFound, "card not found", domain.ErrNotFound)
	}

	return card, nil
}

func (r *fakeCardRepository) ListByUserID(_ context.Context, userID domain.UserID) ([]*domain.Card, error) {
	result := []*domain.Card{}

	for _, card := range r.cards {
		if card.UserID == userID {
			result = append(result, card)
		}
	}

	return result, nil
}

func (r *fakeCardRepository) FindByPANHMAC(_ context.Context, panHMAC string) (*domain.Card, error) {
	for _, card := range r.cards {
		if card.PANHMAC == panHMAC {
			return card, nil
		}
	}

	return nil, domain.NewDomainError(domain.ErrorCodeNotFound, "card not found", domain.ErrNotFound)
}

func (r *fakeCardRepository) UpdateStatus(_ context.Context, id domain.CardID, status domain.CardStatus) error {
	card, ok := r.cards[id]
	if !ok {
		return domain.NewDomainError(domain.ErrorCodeNotFound, "card not found", domain.ErrNotFound)
	}

	card.Status = status
	return nil
}

type fakeTransactionRepository struct {
	transactions map[domain.TransactionID]*domain.Transaction
	next         int
}

func newFakeTransactionRepository() *fakeTransactionRepository {
	return &fakeTransactionRepository{
		transactions: map[domain.TransactionID]*domain.Transaction{},
		next:         1,
	}
}

func (r *fakeTransactionRepository) Create(_ context.Context, transaction *domain.Transaction) error {
	transaction.ID = domain.TransactionID("transaction-" + string(rune('0'+r.next)))
	r.next++
	r.transactions[transaction.ID] = transaction
	return nil
}

func (r *fakeTransactionRepository) FindByID(_ context.Context, id domain.TransactionID) (*domain.Transaction, error) {
	transaction, ok := r.transactions[id]
	if !ok {
		return nil, domain.NewDomainError(domain.ErrorCodeNotFound, "transaction not found", domain.ErrNotFound)
	}

	return transaction, nil
}

func (r *fakeTransactionRepository) FindByIdempotencyKey(_ context.Context, key string) (*domain.Transaction, error) {
	for _, transaction := range r.transactions {
		if transaction.IdempotencyKey == key {
			return transaction, nil
		}
	}

	return nil, domain.NewDomainError(domain.ErrorCodeNotFound, "transaction not found", domain.ErrNotFound)
}

func (r *fakeTransactionRepository) ListByUserID(_ context.Context, userID domain.UserID, _ int, _ int) ([]*domain.Transaction, error) {
	result := []*domain.Transaction{}

	for _, transaction := range r.transactions {
		if transaction.UserID == userID {
			result = append(result, transaction)
		}
	}

	return result, nil
}

func (r *fakeTransactionRepository) ListByAccountID(_ context.Context, accountID domain.AccountID, _ time.Time, _ time.Time) ([]*domain.Transaction, error) {
	result := []*domain.Transaction{}

	for _, transaction := range r.transactions {
		if transaction.SourceAccountID != nil && *transaction.SourceAccountID == accountID {
			result = append(result, transaction)
			continue
		}

		if transaction.DestinationAccountID != nil && *transaction.DestinationAccountID == accountID {
			result = append(result, transaction)
		}
	}

	return result, nil
}
