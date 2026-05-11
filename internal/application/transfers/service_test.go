package transfers

import (
	"context"
	"testing"
	"time"

	accountapp "github.com/example/banking-service/internal/application/accounts"
	"github.com/example/banking-service/internal/domain"
)

func TestServiceTransfer(t *testing.T) {
	accountRepository := newFakeAccountRepository()
	transactionRepository := newFakeTransactionRepository()

	accountService := accountapp.NewService(fakeTxManager{}, accountRepository, transactionRepository)
	transferService := NewService(fakeTxManager{}, accountRepository, transactionRepository)

	source, err := accountService.CreateAccount(context.Background(), domain.UserID("user-1"), accountapp.CreateAccountRequest{})
	if err != nil {
		t.Fatalf("unexpected source create error: %v", err)
	}

	destination, err := accountService.CreateAccount(context.Background(), domain.UserID("user-2"), accountapp.CreateAccountRequest{})
	if err != nil {
		t.Fatalf("unexpected destination create error: %v", err)
	}

	_, err = accountService.Deposit(context.Background(), domain.UserID("user-1"), accountapp.MoneyOperationRequest{
		AccountID: source.ID,
		Amount:    "100.00",
	})
	if err != nil {
		t.Fatalf("unexpected deposit error: %v", err)
	}

	result, err := transferService.Transfer(context.Background(), domain.UserID("user-1"), TransferRequest{
		SourceAccountID:      source.ID,
		DestinationAccountID: destination.ID,
		Amount:               "30.00",
	})
	if err != nil {
		t.Fatalf("unexpected transfer error: %v", err)
	}

	if result.SourceAccount.Balance != "70.00" {
		t.Fatalf("expected source balance 70.00, got %s", result.SourceAccount.Balance)
	}

	if result.DestinationAccount.Balance != "30.00" {
		t.Fatalf("expected destination balance 30.00, got %s", result.DestinationAccount.Balance)
	}
}

type fakeTxManager struct{}

func (m fakeTxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type fakeAccountRepository struct {
	accounts map[domain.AccountID]*domain.Account
	nextID   int
}

func newFakeAccountRepository() *fakeAccountRepository {
	return &fakeAccountRepository{
		accounts: map[domain.AccountID]*domain.Account{},
		nextID:   1,
	}
}

func (r *fakeAccountRepository) Create(_ context.Context, account *domain.Account) error {
	account.ID = domain.AccountID("account-" + string(rune('0'+r.nextID)))
	r.nextID++
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

type fakeTransactionRepository struct {
	transactions map[domain.TransactionID]*domain.Transaction
	nextID       int
}

func newFakeTransactionRepository() *fakeTransactionRepository {
	return &fakeTransactionRepository{
		transactions: map[domain.TransactionID]*domain.Transaction{},
		nextID:       1,
	}
}

func (r *fakeTransactionRepository) Create(_ context.Context, transaction *domain.Transaction) error {
	transaction.ID = domain.TransactionID("transaction-" + string(rune('0'+r.nextID)))
	r.nextID++
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
