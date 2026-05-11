package domain

import "time"

type MFAPurpose string

const (
	MFAPurposeTransfer     MFAPurpose = "transfer"
	MFAPurposeCardPayment  MFAPurpose = "card_payment"
	MFAPurposeCreditIssue  MFAPurpose = "credit_issue"
	MFAPurposeAccountBlock MFAPurpose = "account_block"
)

type MFAStatus string

const (
	MFAStatusPending  MFAStatus = "pending"
	MFAStatusVerified MFAStatus = "verified"
	MFAStatusExpired  MFAStatus = "expired"
	MFAStatusFailed   MFAStatus = "failed"
)

type MFADeliveryChannel string

const (
	MFADeliveryChannelEmail MFADeliveryChannel = "email"
)

type MFAChallenge struct {
	ID              MFAChallengeID
	UserID          UserID
	Purpose         MFAPurpose
	DeliveryChannel MFADeliveryChannel
	Destination     string
	CodeHash        string
	Context         map[string]any
	Status          MFAStatus
	Attempts        int
	ExpiresAt       time.Time
	VerifiedAt      *time.Time
	CreatedAt       time.Time
}

func (m *MFAChallenge) IsVerified() bool {
	return m.Status == MFAStatusVerified && m.VerifiedAt != nil
}

func (m *MFAChallenge) IsExpired(now time.Time) bool {
	return now.After(m.ExpiresAt)
}
