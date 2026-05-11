package analytics

import "time"

type MonthlyAnalyticsResponse struct {
	PeriodStart       time.Time `json:"periodStart"`
	PeriodEnd         time.Time `json:"periodEnd"`
	Income            string    `json:"income"`
	Expenses          string    `json:"expenses"`
	InternalTransfers string    `json:"internalTransfers"`
	NetCashFlow       string    `json:"netCashFlow"`
}

type CreditLoadResponse struct {
	ActiveCredits         int    `json:"activeCredits"`
	OutstandingPrincipal  string `json:"outstandingPrincipal"`
	MonthlyPayment        string `json:"monthlyPayment"`
	OverduePayments       int    `json:"overduePayments"`
	OverdueAmount         string `json:"overdueAmount"`
	CreditLoadDescription string `json:"creditLoadDescription"`
}

type AnalyticsResponse struct {
	Monthly    MonthlyAnalyticsResponse `json:"monthly"`
	CreditLoad CreditLoadResponse       `json:"creditLoad"`
}

type BalancePredictionResponse struct {
	AccountID        string                  `json:"accountId"`
	Days             int                     `json:"days"`
	CurrentBalance   string                  `json:"currentBalance"`
	PredictedBalance string                  `json:"predictedBalance"`
	PlannedPayments  string                  `json:"plannedPayments"`
	Items            []BalancePredictionItem `json:"items"`
}

type BalancePredictionItem struct {
	Date         time.Time `json:"date"`
	Description  string    `json:"description"`
	Amount       string    `json:"amount"`
	BalanceAfter string    `json:"balanceAfter"`
}
