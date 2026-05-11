package domain

import (
	"time"

	"github.com/example/banking-service/pkg/money"
)

type CreditStatus string

const (
	CreditStatusActive    CreditStatus = "active"
	CreditStatusClosed    CreditStatus = "closed"
	CreditStatusOverdue   CreditStatus = "overdue"
	CreditStatusDefaulted CreditStatus = "defaulted"
)

type Credit struct {
	ID                    CreditID
	UserID                UserID
	DisbursementAccountID AccountID
	RepaymentAccountID    AccountID
	PrincipalAmount       money.Money
	OutstandingPrincipal  money.Money
	AnnualInterestRate    string
	CBRKeyRate            string
	BankMargin            string
	TermMonths            int
	AnnuityPayment        money.Money
	PenaltyRate           string
	Status                CreditStatus
	IssuedAt              time.Time
	ClosedAt              *time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type CreateCreditData struct {
	UserID                UserID
	DisbursementAccountID AccountID
	RepaymentAccountID    AccountID
	PrincipalAmount       money.Money
	AnnualInterestRate    string
	CBRKeyRate            string
	BankMargin            string
	TermMonths            int
	AnnuityPayment        money.Money
}

func NewCredit(data CreateCreditData) (*Credit, error) {
	if data.UserID == "" {
		return nil, NewDomainError(ErrorCodeValidation, "user id is required", ErrValidation)
	}

	if data.DisbursementAccountID == "" {
		return nil, NewDomainError(ErrorCodeValidation, "disbursement account id is required", ErrValidation)
	}

	if data.RepaymentAccountID == "" {
		return nil, NewDomainError(ErrorCodeValidation, "repayment account id is required", ErrValidation)
	}

	if !data.PrincipalAmount.IsPositive() {
		return nil, NewDomainError(ErrorCodeValidation, "principal amount must be positive", ErrValidation)
	}

	if data.TermMonths <= 0 {
		return nil, NewDomainError(ErrorCodeValidation, "term months must be greater than zero", ErrValidation)
	}

	if !data.AnnuityPayment.IsPositive() {
		return nil, NewDomainError(ErrorCodeValidation, "annuity payment must be positive", ErrValidation)
	}

	now := time.Now().UTC()

	return &Credit{
		UserID:                data.UserID,
		DisbursementAccountID: data.DisbursementAccountID,
		RepaymentAccountID:    data.RepaymentAccountID,
		PrincipalAmount:       data.PrincipalAmount,
		OutstandingPrincipal:  data.PrincipalAmount,
		AnnualInterestRate:    data.AnnualInterestRate,
		CBRKeyRate:            data.CBRKeyRate,
		BankMargin:            data.BankMargin,
		TermMonths:            data.TermMonths,
		AnnuityPayment:        data.AnnuityPayment,
		PenaltyRate:           "0.1000",
		Status:                CreditStatusActive,
		IssuedAt:              now,
		CreatedAt:             now,
		UpdatedAt:             now,
	}, nil
}
