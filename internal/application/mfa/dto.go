package mfa

import "time"

type CreateChallengeRequest struct {
	Purpose string         `json:"purpose"`
	Context map[string]any `json:"context,omitempty"`
}

type VerifyChallengeRequest struct {
	Code string `json:"code"`
}

type ChallengeResponse struct {
	ID              string         `json:"id"`
	UserID          string         `json:"userId"`
	Purpose         string         `json:"purpose"`
	DeliveryChannel string         `json:"deliveryChannel"`
	Destination     string         `json:"destination"`
	Context         map[string]any `json:"context,omitempty"`
	Status          string         `json:"status"`
	Attempts        int            `json:"attempts"`
	ExpiresAt       time.Time      `json:"expiresAt"`
	VerifiedAt      *time.Time     `json:"verifiedAt,omitempty"`
	CreatedAt       time.Time      `json:"createdAt"`
}
