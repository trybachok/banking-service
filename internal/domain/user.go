package domain

import (
	"regexp"
	"strings"
	"time"
)

type UserRole string

const (
	UserRoleCustomer UserRole = "customer"
	UserRoleAdmin    UserRole = "admin"
)

type UserStatus string

const (
	UserStatusActive  UserStatus = "active"
	UserStatusBlocked UserStatus = "blocked"
	UserStatusDeleted UserStatus = "deleted"
)

type User struct {
	ID           UserID
	Email        string
	Username     string
	PasswordHash string
	FirstName    string
	LastName     string
	Role         UserRole
	Status       UserStatus
	LastLoginAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type RegisterUserData struct {
	Email        string
	Username     string
	Password     string
	PasswordHash string
	FirstName    string
	LastName     string
}

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_.-]{3,64}$`)

func ValidateEmail(email string) error {
	normalized := strings.TrimSpace(email)

	if normalized == "" {
		return NewDomainError(ErrorCodeValidation, "email is required", ErrValidation)
	}

	if !strings.Contains(normalized, "@") {
		return NewDomainError(ErrorCodeValidation, "email must contain @", ErrValidation)
	}

	return nil
}

func ValidateUsername(username string) error {
	normalized := strings.TrimSpace(username)

	if !usernamePattern.MatchString(normalized) {
		return NewDomainError(ErrorCodeValidation, "username must match ^[A-Za-z0-9_.-]{3,64}$", ErrValidation)
	}

	return nil
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return NewDomainError(ErrorCodeValidation, "password must contain at least 8 characters", ErrValidation)
	}

	if len([]byte(password)) > 72 {
		return NewDomainError(ErrorCodeValidation, "password must not exceed 72 bytes for bcrypt", ErrValidation)
	}

	return nil
}

func NewUser(data RegisterUserData) (*User, error) {
	if err := ValidateEmail(data.Email); err != nil {
		return nil, err
	}

	if err := ValidateUsername(data.Username); err != nil {
		return nil, err
	}

	if err := ValidatePassword(data.Password); err != nil {
		return nil, err
	}

	if strings.TrimSpace(data.PasswordHash) == "" {
		return nil, NewDomainError(ErrorCodeValidation, "password hash is required", ErrValidation)
	}

	now := time.Now().UTC()

	return &User{
		Email:        strings.TrimSpace(data.Email),
		Username:     strings.TrimSpace(data.Username),
		PasswordHash: data.PasswordHash,
		FirstName:    strings.TrimSpace(data.FirstName),
		LastName:     strings.TrimSpace(data.LastName),
		Role:         UserRoleCustomer,
		Status:       UserStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin
}
