package credits

import (
	"context"
	"testing"
	"time"

	"github.com/example/banking-service/internal/domain"
	"github.com/example/banking-service/pkg/money"
)

func TestServiceIssueCredit(t *testing.T) {
	accountRepository := newFakeAccountRepository()
	creditRepository := newFakeCreditRepository()
	scheduleRepository := newFakePaymentScheduleRepository()
	transactionRepository := newFakeTransactionRepository()

	account, err := domain.NewAccount(domain.UserID("user-1"), "40817810000000000001", domain.AccountTypeCurrent)
	if err != nil {
		t.Fatalf("unexpected account error: %v", err)
	}

	account.ID = domain.AccountID("account-1")
	accountRepository.accounts[account.ID] = account

	service := NewService(
		fakeTxManager{},
		accountRepository,
		creditRepository,
		scheduleRepository,
		transactionRepository,
		fakeKeyRateProvider{},
	)

	result, err := service.IssueCredit(context.Background(), domain.UserID("user-1"), IssueCreditRequest{
		DisbursementAccountID: "account-1",
		RepaymentAccountID:    "account-1",
		PrincipalAmount:       "100000.00",
		TermMonths:            12,
		BankMargin:            "4.0000",
	})
	if err != nil {
		t.Fatalf("unexpected issue credit error: %v", err)
	}

	if result.Credit.ID == "" {
		t.Fatal("expected credit id")
	}

	if len(result.Schedule) != 12 {
		t.Fatalf("expected 12 schedule rows, got %d", len(result.Schedule))
	}

	if result.Credit.CBRKeyRate != "16.0000" {
		t.Fatalf("expected cbr key rate 16.0000, got %s", result.Credit.CBRKeyRate)
	}
}

type fakeKeyRateProvider struct{}

func (p fakeKeyRateProvider) GetKeyRate(_ context.Context) (string, error) {
	return "16.0000", nil
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

type fakeCreditRepository struct {
	credits map[domain.CreditID]*domain.Credit
	next    int
}

func newFakeCreditRepository() *fakeCreditRepository {
	return &fakeCreditRepository{
		credits: map[domain.CreditID]*domain.Credit{},
		next:    1,
	}
}

func (r *fakeCreditRepository) Create(_ context.Context, credit *domain.Credit) error {
	credit.ID = domain.CreditID("credit-" + string(rune('0'+r.next)))
	r.next++
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

type fakePaymentScheduleRepository struct {
	schedules map[domain.PaymentScheduleID]*domain.PaymentSchedule
	next      int
}

func newFakePaymentScheduleRepository() *fakePaymentScheduleRepository {
	return &fakePaymentScheduleRepository{
		schedules: map[domain.PaymentScheduleID]*domain.PaymentSchedule{},
		next:      1,
	}
}

func (r *fakePaymentScheduleRepository) CreateBatch(_ context.Context, schedules []*domain.PaymentSchedule) error {
	for _, schedule := range schedules {
		schedule.ID = domain.PaymentScheduleID("schedule-" + string(rune('0'+r.next)))
		r.next++
		r.schedules[schedule.ID] = schedule
	}

	return nil
}

func (r *fakePaymentScheduleRepository) ListByCreditID(_ context.Context, creditID domain.CreditID) ([]*domain.PaymentSchedule, error) {
	result := []*domain.PaymentSchedule{}

	for _, schedule := range r.schedules {
		if schedule.CreditID == creditID {
			result = append(result, schedule)
		}
	}

	return result, nil
}

func (r *fakePaymentScheduleRepository) FindDuePayments(_ context.Context, now time.Time, _ int) ([]*domain.PaymentSchedule, error) {
	result := []*domain.PaymentSchedule{}

	for _, schedule := range r.schedules {
		if schedule.IsDue(now) {
			result = append(result, schedule)
		}
	}

	return result, nil
}

func (r *fakePaymentScheduleRepository) FindByIDForUpdate(_ context.Context, id domain.PaymentScheduleID) (*domain.PaymentSchedule, error) {
	schedule, ok := r.schedules[id]
	if !ok {
		return nil, domain.NewDomainError(domain.ErrorCodeNotFound, "schedule not found", domain.ErrNotFound)
	}

	return schedule, nil
}

func (r *fakePaymentScheduleRepository) Update(_ context.Context, schedule *domain.PaymentSchedule) error {
	r.schedules[schedule.ID] = schedule
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

func TestCalculatePenalty(t *testing.T) {
	amount, err := money.FromRubString("100.00")
	if err != nil {
		t.Fatalf("unexpected money error: %v", err)
	}

	penalty, err := calculatePenalty(amount)
	if err != nil {
		t.Fatalf("unexpected penalty error: %v", err)
	}

	if penalty.String() != "10.00" {
		t.Fatalf("expected 10.00, got %s", penalty.String())
	}
}
