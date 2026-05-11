package credits

import (
	"time"

	"github.com/example/banking-service/internal/domain"
)

type IssueCreditRequest struct {
	DisbursementAccountID string `json:"disbursementAccountId"`
	RepaymentAccountID    string `json:"repaymentAccountId"`
	PrincipalAmount       string `json:"principalAmount"`
	TermMonths            int    `json:"termMonths"`
	BankMargin            string `json:"bankMargin,omitempty"`
}

type CreditResponse struct {
	ID                    string     `json:"id"`
	UserID                string     `json:"userId"`
	DisbursementAccountID string     `json:"disbursementAccountId"`
	RepaymentAccountID    string     `json:"repaymentAccountId"`
	PrincipalAmount       string     `json:"principalAmount"`
	OutstandingPrincipal  string     `json:"outstandingPrincipal"`
	AnnualInterestRate    string     `json:"annualInterestRate"`
	CBRKeyRate            string     `json:"cbrKeyRate"`
	BankMargin            string     `json:"bankMargin"`
	TermMonths            int        `json:"termMonths"`
	AnnuityPayment        string     `json:"annuityPayment"`
	PenaltyRate           string     `json:"penaltyRate"`
	Status                string     `json:"status"`
	IssuedAt              time.Time  `json:"issuedAt"`
	ClosedAt              *time.Time `json:"closedAt,omitempty"`
	CreatedAt             time.Time  `json:"createdAt"`
	UpdatedAt             time.Time  `json:"updatedAt"`
}

type PaymentScheduleResponse struct {
	ID                   string     `json:"id"`
	CreditID             string     `json:"creditId"`
	InstallmentNo        int        `json:"installmentNo"`
	DueDate              time.Time  `json:"dueDate"`
	PrincipalAmount      string     `json:"principalAmount"`
	InterestAmount       string     `json:"interestAmount"`
	PenaltyAmount        string     `json:"penaltyAmount"`
	TotalAmount          string     `json:"totalAmount"`
	PaidAmount           string     `json:"paidAmount"`
	Status               string     `json:"status"`
	LastPenaltyAppliedAt *time.Time `json:"lastPenaltyAppliedAt,omitempty"`
	PaidTransactionID    *string    `json:"paidTransactionId,omitempty"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
}

type IssueCreditResponse struct {
	Credit   CreditResponse            `json:"credit"`
	Schedule []PaymentScheduleResponse `json:"schedule"`
}

type ProcessPaymentsResult struct {
	Processed int `json:"processed"`
	Paid      int `json:"paid"`
	Penalized int `json:"penalized"`
	Failed    int `json:"failed"`
}

func ToCreditResponse(credit *domain.Credit) CreditResponse {
	return CreditResponse{
		ID:                    credit.ID.String(),
		UserID:                credit.UserID.String(),
		DisbursementAccountID: credit.DisbursementAccountID.String(),
		RepaymentAccountID:    credit.RepaymentAccountID.String(),
		PrincipalAmount:       credit.PrincipalAmount.String(),
		OutstandingPrincipal:  credit.OutstandingPrincipal.String(),
		AnnualInterestRate:    credit.AnnualInterestRate,
		CBRKeyRate:            credit.CBRKeyRate,
		BankMargin:            credit.BankMargin,
		TermMonths:            credit.TermMonths,
		AnnuityPayment:        credit.AnnuityPayment.String(),
		PenaltyRate:           credit.PenaltyRate,
		Status:                string(credit.Status),
		IssuedAt:              credit.IssuedAt,
		ClosedAt:              credit.ClosedAt,
		CreatedAt:             credit.CreatedAt,
		UpdatedAt:             credit.UpdatedAt,
	}
}

func ToPaymentScheduleResponse(schedule *domain.PaymentSchedule) PaymentScheduleResponse {
	return PaymentScheduleResponse{
		ID:                   schedule.ID.String(),
		CreditID:             schedule.CreditID.String(),
		InstallmentNo:        schedule.InstallmentNo,
		DueDate:              schedule.DueDate,
		PrincipalAmount:      schedule.PrincipalAmount.String(),
		InterestAmount:       schedule.InterestAmount.String(),
		PenaltyAmount:        schedule.PenaltyAmount.String(),
		TotalAmount:          schedule.TotalAmount.String(),
		PaidAmount:           schedule.PaidAmount.String(),
		Status:               string(schedule.Status),
		LastPenaltyAppliedAt: schedule.LastPenaltyAppliedAt,
		PaidTransactionID:    transactionIDToStringPtr(schedule.PaidTransactionID),
		CreatedAt:            schedule.CreatedAt,
		UpdatedAt:            schedule.UpdatedAt,
	}
}

func ToPaymentScheduleResponses(schedules []*domain.PaymentSchedule) []PaymentScheduleResponse {
	result := make([]PaymentScheduleResponse, 0, len(schedules))

	for _, schedule := range schedules {
		result = append(result, ToPaymentScheduleResponse(schedule))
	}

	return result
}

func transactionIDToStringPtr(value *domain.TransactionID) *string {
	if value == nil {
		return nil
	}

	result := value.String()
	return &result
}
