package domain

import (
	"strings"
	"time"

	"github.com/example/banking-service/pkg/money"
)

const CurrencyRUB = "RUB"

type AccountType string

const (
	AccountTypeCurrent         AccountType = "current"
	AccountTypeCreditRepayment AccountType = "credit_repayment"
)

type AccountStatus string

const (
	AccountStatusActive  AccountStatus = "active"
	AccountStatusBlocked AccountStatus = "blocked"
	AccountStatusClosed  AccountStatus = "closed"
)

type Account struct {
	ID            AccountID
	UserID        UserID
	AccountNo     string
	AccountType   AccountType
	Currency      string
	Balance       money.Money
	Status        AccountStatus
	BlockedReason string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewAccount(userID UserID, accountNo string, accountType AccountType) (*Account, error) {
	if strings.TrimSpace(userID.String()) == "" {
		return nil, NewDomainError(ErrorCodeValidation, "user id is required", ErrValidation)
	}

	if strings.TrimSpace(accountNo) == "" {
		return nil, NewDomainError(ErrorCodeValidation, "account number is required", ErrValidation)
	}

	if accountType == "" {
		accountType = AccountTypeCurrent
	}

	now := time.Now().UTC()

	return &Account{
		UserID:      userID,
		AccountNo:   strings.TrimSpace(accountNo),
		AccountType: accountType,
		Currency:    CurrencyRUB,
		Balance:     money.Zero(),
		Status:      AccountStatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (a *Account) CanDebit(amount money.Money) bool {
	return a.Status == AccountStatusActive && a.Balance.MinorUnits() >= amount.MinorUnits()
}

func (a *Account) Credit(amount money.Money) error {
	if !amount.IsPositive() {
		return NewDomainError(ErrorCodeValidation, "amount must be positive", ErrValidation)
	}

	if a.Status != AccountStatusActive {
		return NewDomainError(ErrorCodeInvalidState, "account is not active", ErrInvalidState)
	}

	a.Balance = a.Balance.Add(amount)
	a.UpdatedAt = time.Now().UTC()

	return nil
}

func (a *Account) Debit(amount money.Money) error {
	if !amount.IsPositive() {
		return NewDomainError(ErrorCodeValidation, "amount must be positive", ErrValidation)
	}

	if a.Status != AccountStatusActive {
		return NewDomainError(ErrorCodeInvalidState, "account is not active", ErrInvalidState)
	}

	if !a.CanDebit(amount) {
		return NewDomainError(ErrorCodeInsufficientFunds, "insufficient funds", ErrInsufficientFunds)
	}

	result, err := a.Balance.Sub(amount)
	if err != nil {
		return NewDomainError(ErrorCodeInsufficientFunds, "insufficient funds", err)
	}

	a.Balance = result
	a.UpdatedAt = time.Now().UTC()

	return nil
}
