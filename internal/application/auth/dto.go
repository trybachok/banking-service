package auth

import "time"

type RegisterRequest struct {
	Email     string `json:"email"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken      string       `json:"accessToken"`
	TokenType        string       `json:"tokenType"`
	ExpiresAt        time.Time    `json:"expiresAt"`
	ExpiresInSeconds int64        `json:"expiresInSeconds"`
	User             UserResponse `json:"user"`
}

type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Role      string `json:"role"`
	Status    string `json:"status"`
}
