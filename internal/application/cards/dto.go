package cards

import (
	"time"

	accountapp "github.com/example/banking-service/internal/application/accounts"
	"github.com/example/banking-service/internal/domain"
)

type IssueCardRequest struct {
	AccountID string `json:"accountId"`
	Alias     string `json:"alias,omitempty"`
}

type CardPaymentRequest struct {
	Amount         string `json:"amount"`
	CVV            string `json:"cvv"`
	Description    string `json:"description,omitempty"`
	IdempotencyKey string `json:"idempotencyKey,omitempty"`
}

type CardResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	AccountID string    `json:"accountId"`
	Alias     string    `json:"alias,omitempty"`
	PANMasked string    `json:"panMasked"`
	PANLast4  string    `json:"panLast4"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CardDetailsResponse struct {
	Card   CardResponse `json:"card"`
	PAN    string       `json:"pan,omitempty"`
	Expiry string       `json:"expiry,omitempty"`
	CVV    string       `json:"cvv,omitempty"`
	Notice string       `json:"notice,omitempty"`
}

type CardPaymentResponse struct {
	Card        CardResponse                   `json:"card"`
	Account     accountapp.AccountResponse     `json:"account"`
	Transaction accountapp.TransactionResponse `json:"transaction"`
}

type ProtectedCardData struct {
	PANEncrypted    []byte
	ExpiryEncrypted []byte
	PANHMAC         string
	IntegrityHMAC   string
	CVVHash         string
}

type CardPlainData struct {
	PAN    string
	Expiry string
}

func ToCardResponse(card *domain.Card) CardResponse {
	return CardResponse{
		ID:        card.ID.String(),
		UserID:    card.UserID.String(),
		AccountID: card.AccountID.String(),
		Alias:     card.CardAlias,
		PANMasked: card.PANMasked,
		PANLast4:  card.PANLast4,
		Status:    string(card.Status),
		CreatedAt: card.CreatedAt,
		UpdatedAt: card.UpdatedAt,
	}
}
