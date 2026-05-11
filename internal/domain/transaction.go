package domain

import (
	"time"

	"github.com/example/banking-service/pkg/money"
)

type TransactionType string

const (
	TransactionTypeDeposit            TransactionType = "deposit"
	TransactionTypeWithdrawal         TransactionType = "withdrawal"
	TransactionTypeTransfer           TransactionType = "transfer"
	TransactionTypeCardPayment        TransactionType = "card_payment"
	TransactionTypeCreditDisbursement TransactionType = "credit_disbursement"
	TransactionTypeCreditPayment      TransactionType = "credit_payment"
	TransactionTypePenalty            TransactionType = "penalty"
	TransactionTypeRefund             TransactionType = "refund"
)

type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusReversed  TransactionStatus = "reversed"
)

type Transaction struct {
	ID                   TransactionID
	UserID               UserID
	SourceAccountID      *AccountID
	DestinationAccountID *AccountID
	CardID               *CardID
	CreditID             *CreditID
	OperationType        TransactionType
	Status               TransactionStatus
	Amount               money.Money
	Currency             string
	Description          string
	IdempotencyKey       string
	ExternalReference    string
	CreatedAt            time.Time
	CompletedAt          *time.Time
}

type CreateTransactionData struct {
	UserID               UserID
	SourceAccountID      *AccountID
	DestinationAccountID *AccountID
	CardID               *CardID
	CreditID             *CreditID
	OperationType        TransactionType
	Amount               money.Money
	Description          string
	IdempotencyKey       string
	ExternalReference    string
}

func NewCompletedTransaction(data CreateTransactionData) (*Transaction, error) {
	if data.UserID == "" {
		return nil, NewDomainError(ErrorCodeValidation, "user id is required", ErrValidation)
	}

	if data.OperationType == "" {
		return nil, NewDomainError(ErrorCodeValidation, "operation type is required", ErrValidation)
	}

	if !data.Amount.IsPositive() {
		return nil, NewDomainError(ErrorCodeValidation, "amount must be positive", ErrValidation)
	}

	now := time.Now().UTC()

	return &Transaction{
		UserID:               data.UserID,
		SourceAccountID:      data.SourceAccountID,
		DestinationAccountID: data.DestinationAccountID,
		CardID:               data.CardID,
		CreditID:             data.CreditID,
		OperationType:        data.OperationType,
		Status:               TransactionStatusCompleted,
		Amount:               data.Amount,
		Currency:             CurrencyRUB,
		Description:          data.Description,
		IdempotencyKey:       data.IdempotencyKey,
		ExternalReference:    data.ExternalReference,
		CreatedAt:            now,
		CompletedAt:          &now,
	}, nil
}
