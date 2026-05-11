package domain

import (
	"strings"
	"time"
)

type CardStatus string

const (
	CardStatusActive  CardStatus = "active"
	CardStatusBlocked CardStatus = "blocked"
	CardStatusExpired CardStatus = "expired"
	CardStatusClosed  CardStatus = "closed"
)

type Card struct {
	ID              CardID
	UserID          UserID
	AccountID       AccountID
	CardAlias       string
	PANMasked       string
	PANLast4        string
	PANEncrypted    []byte
	ExpiryEncrypted []byte
	PANHMAC         string
	IntegrityHMAC   string
	CVVHash         string
	Status          CardStatus
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type CreateCardData struct {
	UserID          UserID
	AccountID       AccountID
	CardAlias       string
	PANMasked       string
	PANLast4        string
	PANEncrypted    []byte
	ExpiryEncrypted []byte
	PANHMAC         string
	IntegrityHMAC   string
	CVVHash         string
}

type CardPlainDetails struct {
	PAN    string
	Expiry string
	CVV    string
}

func NewCard(data CreateCardData) (*Card, error) {
	if strings.TrimSpace(data.UserID.String()) == "" {
		return nil, NewDomainError(ErrorCodeValidation, "user id is required", ErrValidation)
	}

	if strings.TrimSpace(data.AccountID.String()) == "" {
		return nil, NewDomainError(ErrorCodeValidation, "account id is required", ErrValidation)
	}

	if strings.TrimSpace(data.PANMasked) == "" {
		return nil, NewDomainError(ErrorCodeValidation, "masked PAN is required", ErrValidation)
	}

	if len(data.PANLast4) != 4 {
		return nil, NewDomainError(ErrorCodeValidation, "PAN last4 must contain 4 digits", ErrValidation)
	}

	if len(data.PANEncrypted) == 0 {
		return nil, NewDomainError(ErrorCodeValidation, "encrypted PAN is required", ErrValidation)
	}

	if len(data.ExpiryEncrypted) == 0 {
		return nil, NewDomainError(ErrorCodeValidation, "encrypted expiry is required", ErrValidation)
	}

	if strings.TrimSpace(data.PANHMAC) == "" {
		return nil, NewDomainError(ErrorCodeValidation, "PAN HMAC is required", ErrValidation)
	}

	if strings.TrimSpace(data.IntegrityHMAC) == "" {
		return nil, NewDomainError(ErrorCodeValidation, "integrity HMAC is required", ErrValidation)
	}

	if strings.TrimSpace(data.CVVHash) == "" {
		return nil, NewDomainError(ErrorCodeValidation, "CVV hash is required", ErrValidation)
	}

	now := time.Now().UTC()

	return &Card{
		UserID:          data.UserID,
		AccountID:       data.AccountID,
		CardAlias:       strings.TrimSpace(data.CardAlias),
		PANMasked:       data.PANMasked,
		PANLast4:        data.PANLast4,
		PANEncrypted:    data.PANEncrypted,
		ExpiryEncrypted: data.ExpiryEncrypted,
		PANHMAC:         data.PANHMAC,
		IntegrityHMAC:   data.IntegrityHMAC,
		CVVHash:         data.CVVHash,
		Status:          CardStatusActive,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

func (c *Card) IsActive() bool {
	return c.Status == CardStatusActive
}
