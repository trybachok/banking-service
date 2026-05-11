package domain

import (
	"context"
	"time"
)

type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id UserID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	ExistsByEmailOrUsername(ctx context.Context, email string, username string) (bool, error)
	UpdateLastLoginAt(ctx context.Context, id UserID, at time.Time) error
	List(ctx context.Context, limit int, offset int) ([]*User, error)
}

type AccountRepository interface {
	Create(ctx context.Context, account *Account) error
	FindByID(ctx context.Context, id AccountID) (*Account, error)
	FindByIDForUpdate(ctx context.Context, id AccountID) (*Account, error)
	ListByUserID(ctx context.Context, userID UserID) ([]*Account, error)
	UpdateBalance(ctx context.Context, id AccountID, account *Account) error
	UpdateStatus(ctx context.Context, id AccountID, status AccountStatus, reason string) error
}

type CardRepository interface {
	Create(ctx context.Context, card *Card) error
	FindByID(ctx context.Context, id CardID) (*Card, error)
	ListByUserID(ctx context.Context, userID UserID) ([]*Card, error)
	FindByPANHMAC(ctx context.Context, panHMAC string) (*Card, error)
	UpdateStatus(ctx context.Context, id CardID, status CardStatus) error
}

type TransactionRepository interface {
	Create(ctx context.Context, transaction *Transaction) error
	FindByID(ctx context.Context, id TransactionID) (*Transaction, error)
	FindByIdempotencyKey(ctx context.Context, key string) (*Transaction, error)
	ListByUserID(ctx context.Context, userID UserID, limit int, offset int) ([]*Transaction, error)
	ListByAccountID(ctx context.Context, accountID AccountID, from time.Time, to time.Time) ([]*Transaction, error)
}

type CreditRepository interface {
	Create(ctx context.Context, credit *Credit) error
	FindByID(ctx context.Context, id CreditID) (*Credit, error)
	ListByUserID(ctx context.Context, userID UserID) ([]*Credit, error)
	UpdateStatus(ctx context.Context, id CreditID, status CreditStatus) error
	UpdateOutstandingPrincipal(ctx context.Context, id CreditID, credit *Credit) error
}

type PaymentScheduleRepository interface {
	CreateBatch(ctx context.Context, schedules []*PaymentSchedule) error
	ListByCreditID(ctx context.Context, creditID CreditID) ([]*PaymentSchedule, error)
	FindDuePayments(ctx context.Context, now time.Time, limit int) ([]*PaymentSchedule, error)
	FindByIDForUpdate(ctx context.Context, id PaymentScheduleID) (*PaymentSchedule, error)
	Update(ctx context.Context, schedule *PaymentSchedule) error
}

type EmailOutboxRepository interface {
	Create(ctx context.Context, userID UserID, recipientEmail string, subject string, templateCode string, payload map[string]any) error
	FindPending(ctx context.Context, now time.Time, limit int) ([]EmailOutboxMessage, error)
	MarkSent(ctx context.Context, id string, sentAt time.Time) error
	MarkFailed(ctx context.Context, id string, nextAttemptAt time.Time, lastError string) error
}

type EmailOutboxMessage struct {
	ID             string
	UserID         *UserID
	RecipientEmail string
	Subject        string
	TemplateCode   string
	Payload        map[string]any
	Attempts       int
	CreatedAt      time.Time
}

type MFARepository interface {
	Create(ctx context.Context, challenge *MFAChallenge) error
	FindByID(ctx context.Context, id MFAChallengeID) (*MFAChallenge, error)
	Update(ctx context.Context, challenge *MFAChallenge) error
}
