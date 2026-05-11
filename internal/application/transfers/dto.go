package transfers

import "github.com/example/banking-service/internal/application/accounts"

type TransferRequest struct {
	SourceAccountID      string `json:"sourceAccountId"`
	DestinationAccountID string `json:"destinationAccountId"`
	Amount               string `json:"amount"`
	Description          string `json:"description,omitempty"`
	IdempotencyKey       string `json:"idempotencyKey,omitempty"`
}

type TransferResponse struct {
	SourceAccount      accounts.AccountResponse     `json:"sourceAccount"`
	DestinationAccount accounts.AccountResponse     `json:"destinationAccount"`
	Transaction        accounts.TransactionResponse `json:"transaction"`
}
