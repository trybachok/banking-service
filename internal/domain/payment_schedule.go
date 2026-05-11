package domain

import (
	"time"

	"github.com/example/banking-service/pkg/money"
)

type PaymentScheduleStatus string

const (
	PaymentScheduleStatusPending       PaymentScheduleStatus = "pending"
	PaymentScheduleStatusPaid          PaymentScheduleStatus = "paid"
	PaymentScheduleStatusOverdue       PaymentScheduleStatus = "overdue"
	PaymentScheduleStatusPartiallyPaid PaymentScheduleStatus = "partially_paid"
	PaymentScheduleStatusCancelled     PaymentScheduleStatus = "cancelled"
)

type PaymentSchedule struct {
	ID                   PaymentScheduleID
	CreditID             CreditID
	InstallmentNo        int
	DueDate              time.Time
	PrincipalAmount      money.Money
	InterestAmount       money.Money
	PenaltyAmount        money.Money
	TotalAmount          money.Money
	PaidAmount           money.Money
	Status               PaymentScheduleStatus
	LastPenaltyAppliedAt *time.Time
	PaidTransactionID    *TransactionID
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func (p *PaymentSchedule) IsDue(now time.Time) bool {
	return p.Status == PaymentScheduleStatusPending && !p.DueDate.After(now)
}

func (p *PaymentSchedule) IsPaid() bool {
	return p.Status == PaymentScheduleStatusPaid
}
