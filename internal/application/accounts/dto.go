package accounts

import (
	"time"

	"github.com/example/banking-service/internal/domain"
)

type CreateAccountRequest struct {
	AccountType string `json:"accountType,omitempty"`
}

type MoneyOperationRequest struct {
	AccountID      string `json:"accountId,omitempty"`
	Amount         string `json:"amount"`
	Description    string `json:"description,omitempty"`
	IdempotencyKey string `json:"idempotencyKey,omitempty"`
}

type AccountResponse struct {
	ID            string    `json:"id"`
	UserID        string    `json:"userId"`
	AccountNo     string    `json:"accountNo"`
	AccountType   string    `json:"accountType"`
	Currency      string    `json:"currency"`
	Balance       string    `json:"balance"`
	Status        string    `json:"status"`
	BlockedReason string    `json:"blockedReason,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type TransactionResponse struct {
	ID                   string     `json:"id"`
	UserID               string     `json:"userId"`
	SourceAccountID      *string    `json:"sourceAccountId,omitempty"`
	DestinationAccountID *string    `json:"destinationAccountId,omitempty"`
	CardID               *string    `json:"cardId,omitempty"`
	CreditID             *string    `json:"creditId,omitempty"`
	OperationType        string     `json:"operationType"`
	Status               string     `json:"status"`
	Amount               string     `json:"amount"`
	Currency             string     `json:"currency"`
	Description          string     `json:"description,omitempty"`
	IdempotencyKey       string     `json:"idempotencyKey,omitempty"`
	ExternalReference    string     `json:"externalReference,omitempty"`
	CreatedAt            time.Time  `json:"createdAt"`
	CompletedAt          *time.Time `json:"completedAt,omitempty"`
}

type AccountOperationResponse struct {
	Account     AccountResponse     `json:"account"`
	Transaction TransactionResponse `json:"transaction"`
}

func ToAccountResponse(account *domain.Account) AccountResponse {
	return AccountResponse{
		ID:            account.ID.String(),
		UserID:        account.UserID.String(),
		AccountNo:     account.AccountNo,
		AccountType:   string(account.AccountType),
		Currency:      account.Currency,
		Balance:       account.Balance.String(),
		Status:        string(account.Status),
		BlockedReason: account.BlockedReason,
		CreatedAt:     account.CreatedAt,
		UpdatedAt:     account.UpdatedAt,
	}
}

func ToTransactionResponse(transaction *domain.Transaction) TransactionResponse {
	return TransactionResponse{
		ID:                   transaction.ID.String(),
		UserID:               transaction.UserID.String(),
		SourceAccountID:      accountIDToStringPtr(transaction.SourceAccountID),
		DestinationAccountID: accountIDToStringPtr(transaction.DestinationAccountID),
		CardID:               cardIDToStringPtr(transaction.CardID),
		CreditID:             creditIDToStringPtr(transaction.CreditID),
		OperationType:        string(transaction.OperationType),
		Status:               string(transaction.Status),
		Amount:               transaction.Amount.String(),
		Currency:             transaction.Currency,
		Description:          transaction.Description,
		IdempotencyKey:       transaction.IdempotencyKey,
		ExternalReference:    transaction.ExternalReference,
		CreatedAt:            transaction.CreatedAt,
		CompletedAt:          transaction.CompletedAt,
	}
}

func accountIDToStringPtr(value *domain.AccountID) *string {
	if value == nil {
		return nil
	}

	result := value.String()
	return &result
}

func cardIDToStringPtr(value *domain.CardID) *string {
	if value == nil {
		return nil
	}

	result := value.String()
	return &result
}

func creditIDToStringPtr(value *domain.CreditID) *string {
	if value == nil {
		return nil
	}

	result := value.String()
	return &result
}
