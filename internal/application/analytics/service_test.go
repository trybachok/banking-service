package analytics

import (
	"context"
	"testing"
	"time"

	"github.com/example/banking-service/internal/domain"
	"github.com/example/banking-service/pkg/money"
)

func TestServiceGetMonthlyAnalytics(t *testing.T) {
	accountRepository := newFakeAccountRepository()
	transactionRepository := newFakeTransactionRepository()
	creditRepository := newFakeCreditRepository()
	scheduleRepository := newFakeScheduleRepository()

	account := newTestAccount(t, "account-1", "user-1", "1000.00")
	accountRepository.accounts[account.ID] = account

	depositAmount, _ := money.FromRubString("500.00")
	withdrawAmount, _ := money.FromRubString("100.00")

	destinationAccountID := account.ID
	sourceAccountID := account.ID

	transactionRepository.transactions = append(transactionRepository.transactions,
		&domain.Transaction{
			ID:                   domain.TransactionID("tx-1"),
			UserID:               domain.UserID("user-1"),
			DestinationAccountID: &destinationAccountID,
			OperationType:        domain.TransactionTypeDeposit,
			Status:               domain.TransactionStatusCompleted,
			Amount:               depositAmount,
			Currency:             domain.CurrencyRUB,
			CreatedAt:            time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC),
		},
		&domain.Transaction{
			ID:              domain.TransactionID("tx-2"),
			UserID:          domain.UserID("user-1"),
			SourceAccountID: &sourceAccountID,
			OperationType:   domain.TransactionTypeWithdrawal,
			Status:          domain.TransactionStatusCompleted,
			Amount:          withdrawAmount,
			Currency:        domain.CurrencyRUB,
			CreatedAt:       time.Date(2026, 5, 11, 0, 0, 0, 0, time.UTC),
		},
	)

	service := NewService(accountRepository, transactionRepository, creditRepository, scheduleRepository)

	result, err := service.GetMonthlyAnalytics(context.Background(), domain.UserID("user-1"), time.Date(2026, 5, 11, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Income != "500.00" {
		t.Fatalf("expected income 500.00, got %s", result.Income)
	}

	if result.Expenses != "100.00" {
		t.Fatalf("expected expenses 100.00, got %s", result.Expenses)
	}

	if result.NetCashFlow != "400.00" {
		t.Fatalf("expected net cash flow 400.00, got %s", result.NetCashFlow)
	}
}

func TestServicePredictBalance(t *testing.T) {
	accountRepository := newFakeAccountRepository()
	transactionRepository := newFakeTransactionRepository()
	creditRepository := newFakeCreditRepository()
	scheduleRepository := newFakeScheduleRepository()

	account := newTestAccount(t, "account-1", "user-1", "1000.00")
	accountRepository.accounts[account.ID] = account

	principal, _ := money.FromRubString("10000.00")
	payment, _ := money.FromRubString("250.00")

	credit := &domain.Credit{
		ID:                   domain.CreditID("credit-1"),
		UserID:               domain.UserID("user-1"),
		RepaymentAccountID:   account.ID,
		OutstandingPrincipal: principal,
		AnnuityPayment:       payment,
		Status:               domain.CreditStatusActive,
	}

	creditRepository.credits[credit.ID] = credit

	scheduleRepository.schedules = append(scheduleRepository.schedules, &domain.PaymentSchedule{
		ID:            domain.PaymentScheduleID("schedule-1"),
		CreditID:      credit.ID,
		InstallmentNo: 1,
		DueDate:       time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC),
		TotalAmount:   payment,
		Status:        domain.PaymentScheduleStatusPending,
	})

	service := NewService(accountRepository, transactionRepository, creditRepository, scheduleRepository)

	result, err := service.PredictBalance(
		context.Background(),
		domain.UserID("user-1"),
		account.ID,
		30,
		time.Date(2026, 5, 11, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.CurrentBalance != "1000.00" {
		t.Fatalf("expected current balance 1000.00, got %s", result.CurrentBalance)
	}

	if result.PredictedBalance != "750.00" {
		t.Fatalf("expected predicted balance 750.00, got %s", result.PredictedBalance)
	}

	if result.PlannedPayments != "250.00" {
		t.Fatalf("expected planned payments 250.00, got %s", result.PlannedPayments)
	}
}

func newTestAccount(t *testing.T, id string, userID string, balance string) *domain.Account {
	t.Helper()

	amount, err := money.FromRubString(balance)
	if err != nil {
		t.Fatalf("unexpected money error: %v", err)
	}

	account, err := domain.NewAccount(domain.UserID(userID), "40817810000000000001", domain.AccountTypeCurrent)
	if err != nil {
		t.Fatalf("unexpected account error: %v", err)
	}

	account.ID = domain.AccountID(id)
	account.Balance = amount

	return account
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

type fakeTransactionRepository struct {
	transactions []*domain.Transaction
}

func newFakeTransactionRepository() *fakeTransactionRepository {
	return &fakeTransactionRepository{
		transactions: []*domain.Transaction{},
	}
}

func (r *fakeTransactionRepository) Create(_ context.Context, transaction *domain.Transaction) error {
	r.transactions = append(r.transactions, transaction)
	return nil
}

func (r *fakeTransactionRepository) FindByID(_ context.Context, id domain.TransactionID) (*domain.Transaction, error) {
	for _, transaction := range r.transactions {
		if transaction.ID == id {
			return transaction, nil
		}
	}

	return nil, domain.NewDomainError(domain.ErrorCodeNotFound, "transaction not found", domain.ErrNotFound)
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

func (r *fakeTransactionRepository) ListByAccountID(_ context.Context, accountID domain.AccountID, from time.Time, to time.Time) ([]*domain.Transaction, error) {
	result := []*domain.Transaction{}

	for _, transaction := range r.transactions {
		if transaction.CreatedAt.Before(from) || transaction.CreatedAt.After(to) {
			continue
		}

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

type fakeCreditRepository struct {
	credits map[domain.CreditID]*domain.Credit
}

func newFakeCreditRepository() *fakeCreditRepository {
	return &fakeCreditRepository{
		credits: map[domain.CreditID]*domain.Credit{},
	}
}

func (r *fakeCreditRepository) Create(_ context.Context, credit *domain.Credit) error {
	r.credits[credit.ID] = credit
	return nil
}

func (r *fakeCreditRepository) FindByID(_ context.Context, id domain.CreditID) (*domain.Credit, error) {
	credit, ok := r.credits[id]
	if !ok {
		return nil, domain.NewDomainError(domain.ErrorCodeNotFound, "credit not found", domain.ErrNotFound)
	}

	return credit, nil
}

func (r *fakeCreditRepository) ListByUserID(_ context.Context, userID domain.UserID) ([]*domain.Credit, error) {
	result := []*domain.Credit{}

	for _, credit := range r.credits {
		if credit.UserID == userID {
			result = append(result, credit)
		}
	}

	return result, nil
}

func (r *fakeCreditRepository) UpdateStatus(_ context.Context, id domain.CreditID, status domain.CreditStatus) error {
	credit, ok := r.credits[id]
	if !ok {
		return domain.NewDomainError(domain.ErrorCodeNotFound, "credit not found", domain.ErrNotFound)
	}

	credit.Status = status
	return nil
}

func (r *fakeCreditRepository) UpdateOutstandingPrincipal(_ context.Context, id domain.CreditID, credit *domain.Credit) error {
	r.credits[id] = credit
	return nil
}

type fakeScheduleRepository struct {
	schedules []*domain.PaymentSchedule
}

func newFakeScheduleRepository() *fakeScheduleRepository {
	return &fakeScheduleRepository{
		schedules: []*domain.PaymentSchedule{},
	}
}

func (r *fakeScheduleRepository) CreateBatch(_ context.Context, schedules []*domain.PaymentSchedule) error {
	r.schedules = append(r.schedules, schedules...)
	return nil
}

func (r *fakeScheduleRepository) ListByCreditID(_ context.Context, creditID domain.CreditID) ([]*domain.PaymentSchedule, error) {
	result := []*domain.PaymentSchedule{}

	for _, schedule := range r.schedules {
		if schedule.CreditID == creditID {
			result = append(result, schedule)
		}
	}

	return result, nil
}

func (r *fakeScheduleRepository) FindDuePayments(_ context.Context, now time.Time, _ int) ([]*domain.PaymentSchedule, error) {
	result := []*domain.PaymentSchedule{}

	for _, schedule := range r.schedules {
		if schedule.IsDue(now) {
			result = append(result, schedule)
		}
	}

	return result, nil
}

func (r *fakeScheduleRepository) FindByIDForUpdate(_ context.Context, id domain.PaymentScheduleID) (*domain.PaymentSchedule, error) {
	for _, schedule := range r.schedules {
		if schedule.ID == id {
			return schedule, nil
		}
	}

	return nil, domain.NewDomainError(domain.ErrorCodeNotFound, "schedule not found", domain.ErrNotFound)
}

func (r *fakeScheduleRepository) Update(_ context.Context, schedule *domain.PaymentSchedule) error {
	for index, item := range r.schedules {
		if item.ID == schedule.ID {
			r.schedules[index] = schedule
			return nil
		}
	}

	r.schedules = append(r.schedules, schedule)
	return nil
}
