package accounts

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"github.com/example/banking-service/internal/domain"
	"github.com/example/banking-service/pkg/money"
)

type Service struct {
	tx           domain.TxManager
	accounts     domain.AccountRepository
	transactions domain.TransactionRepository
}

func NewService(
	tx domain.TxManager,
	accounts domain.AccountRepository,
	transactions domain.TransactionRepository,
) *Service {
	return &Service{
		tx:           tx,
		accounts:     accounts,
		transactions: transactions,
	}
}

func (s *Service) CreateAccount(ctx context.Context, userID domain.UserID, req CreateAccountRequest) (*AccountResponse, error) {
	accountType, err := parseAccountType(req.AccountType)
	if err != nil {
		return nil, err
	}

	accountNo, err := generateAccountNumber()
	if err != nil {
		return nil, err
	}

	account, err := domain.NewAccount(userID, accountNo, accountType)
	if err != nil {
		return nil, err
	}

	if err := s.accounts.Create(ctx, account); err != nil {
		return nil, err
	}

	response := ToAccountResponse(account)

	return &response, nil
}

func (s *Service) ListAccounts(ctx context.Context, userID domain.UserID) ([]AccountResponse, error) {
	accountsList, err := s.accounts.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	response := make([]AccountResponse, 0, len(accountsList))
	for _, account := range accountsList {
		response = append(response, ToAccountResponse(account))
	}

	return response, nil
}

func (s *Service) Deposit(ctx context.Context, userID domain.UserID, req MoneyOperationRequest) (*AccountOperationResponse, error) {
	amount, err := parsePositiveAmount(req.Amount)
	if err != nil {
		return nil, err
	}

	accountID := domain.AccountID(req.AccountID)

	var response *AccountOperationResponse

	err = s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		account, err := s.accounts.FindByIDForUpdate(txCtx, accountID)
		if err != nil {
			return err
		}

		if err := ensureAccountOwner(account, userID); err != nil {
			return err
		}

		if err := account.Credit(amount); err != nil {
			return err
		}

		if err := s.accounts.UpdateBalance(txCtx, account.ID, account); err != nil {
			return err
		}

		destinationAccountID := account.ID

		transaction, err := domain.NewCompletedTransaction(domain.CreateTransactionData{
			UserID:               userID,
			DestinationAccountID: &destinationAccountID,
			OperationType:        domain.TransactionTypeDeposit,
			Amount:               amount,
			Description:          req.Description,
			IdempotencyKey:       req.IdempotencyKey,
		})
		if err != nil {
			return err
		}

		if err := s.transactions.Create(txCtx, transaction); err != nil {
			return err
		}

		response = &AccountOperationResponse{
			Account:     ToAccountResponse(account),
			Transaction: ToTransactionResponse(transaction),
		}

		return nil
	})

	return response, err
}

func (s *Service) Withdraw(ctx context.Context, userID domain.UserID, req MoneyOperationRequest) (*AccountOperationResponse, error) {
	amount, err := parsePositiveAmount(req.Amount)
	if err != nil {
		return nil, err
	}

	accountID := domain.AccountID(req.AccountID)

	var response *AccountOperationResponse

	err = s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		account, err := s.accounts.FindByIDForUpdate(txCtx, accountID)
		if err != nil {
			return err
		}

		if err := ensureAccountOwner(account, userID); err != nil {
			return err
		}

		if err := account.Debit(amount); err != nil {
			return err
		}

		if err := s.accounts.UpdateBalance(txCtx, account.ID, account); err != nil {
			return err
		}

		sourceAccountID := account.ID

		transaction, err := domain.NewCompletedTransaction(domain.CreateTransactionData{
			UserID:          userID,
			SourceAccountID: &sourceAccountID,
			OperationType:   domain.TransactionTypeWithdrawal,
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

		response = &AccountOperationResponse{
			Account:     ToAccountResponse(account),
			Transaction: ToTransactionResponse(transaction),
		}

		return nil
	})

	return response, err
}

func parseAccountType(value string) (domain.AccountType, error) {
	normalized := strings.TrimSpace(value)

	if normalized == "" {
		return domain.AccountTypeCurrent, nil
	}

	switch domain.AccountType(normalized) {
	case domain.AccountTypeCurrent, domain.AccountTypeCreditRepayment:
		return domain.AccountType(normalized), nil
	default:
		return "", domain.NewDomainError(domain.ErrorCodeValidation, "unsupported account type", domain.ErrValidation)
	}
}

func parsePositiveAmount(value string) (money.Money, error) {
	amount, err := money.FromRubString(value)
	if err != nil {
		return money.Money{}, domain.NewDomainError(domain.ErrorCodeValidation, "invalid amount", err)
	}

	if !amount.IsPositive() {
		return money.Money{}, domain.NewDomainError(domain.ErrorCodeValidation, "amount must be positive", domain.ErrValidation)
	}

	return amount, nil
}

func ensureAccountOwner(account *domain.Account, userID domain.UserID) error {
	if account.UserID != userID {
		return domain.NewDomainError(domain.ErrorCodeForbidden, "account does not belong to current user", domain.ErrForbidden)
	}

	return nil
}

func generateAccountNumber() (string, error) {
	result := "40817810"

	for len(result) < 20 {
		digit, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", fmt.Errorf("generate account number: %w", err)
		}

		result += digit.String()
	}

	return result, nil
}

func ParseAmountForTransfer(value string) (money.Money, error) {
	return parsePositiveAmount(value)
}
