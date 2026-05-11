package admin

import "time"

type UserResponse struct {
	ID          string     `json:"id"`
	Email       string     `json:"email"`
	Username    string     `json:"username"`
	FirstName   string     `json:"firstName,omitempty"`
	LastName    string     `json:"lastName,omitempty"`
	Role        string     `json:"role"`
	Status      string     `json:"status"`
	LastLoginAt *time.Time `json:"lastLoginAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type BlockAccountRequest struct {
	Reason string `json:"reason"`
}

type BlockAccountResponse struct {
	AccountID string `json:"accountId"`
	Status    string `json:"status"`
	Reason    string `json:"reason"`
}
