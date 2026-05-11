package analytics

import (
	"context"
	"time"

	"github.com/example/banking-service/internal/domain"
	"github.com/example/banking-service/pkg/money"
)

type Service struct {
	accounts     domain.AccountRepository
	transactions domain.TransactionRepository
	credits      domain.CreditRepository
	schedules    domain.PaymentScheduleRepository
}

func NewService(
	accounts domain.AccountRepository,
	transactions domain.TransactionRepository,
	credits domain.CreditRepository,
	schedules domain.PaymentScheduleRepository,
) *Service {
	return &Service{
		accounts:     accounts,
		transactions: transactions,
		credits:      credits,
		schedules:    schedules,
	}
}

func (s *Service) GetAnalytics(ctx context.Context, userID domain.UserID, now time.Time) (*AnalyticsResponse, error) {
	monthly, err := s.GetMonthlyAnalytics(ctx, userID, now)
	if err != nil {
		return nil, err
	}

	creditLoad, err := s.GetCreditLoad(ctx, userID, now)
	if err != nil {
		return nil, err
	}

	return &AnalyticsResponse{
		Monthly:    *monthly,
		CreditLoad: *creditLoad,
	}, nil
}

func (s *Service) GetMonthlyAnalytics(ctx context.Context, userID domain.UserID, now time.Time) (*MonthlyAnalyticsResponse, error) {
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Nanosecond)

	accounts, err := s.accounts.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	userAccountIDs := map[domain.AccountID]bool{}
	for _, account := range accounts {
		userAccountIDs[account.ID] = true
	}

	seenTransactions := map[domain.TransactionID]bool{}

	income := money.Zero()
	expenses := money.Zero()
	internalTransfers := money.Zero()

	for _, account := range accounts {
		items, err := s.transactions.ListByAccountID(ctx, account.ID, periodStart, periodEnd)
		if err != nil {
			return nil, err
		}

		for _, transaction := range items {
			if seenTransactions[transaction.ID] {
				continue
			}
			seenTransactions[transaction.ID] = true

			sourceBelongsToUser := transaction.SourceAccountID != nil && userAccountIDs[*transaction.SourceAccountID]
			destinationBelongsToUser := transaction.DestinationAccountID != nil && userAccountIDs[*transaction.DestinationAccountID]

			if sourceBelongsToUser && destinationBelongsToUser {
				internalTransfers = internalTransfers.Add(transaction.Amount)
				continue
			}

			switch transaction.OperationType {
			case domain.TransactionTypeDeposit, domain.TransactionTypeCreditDisbursement, domain.TransactionTypeRefund:
				if destinationBelongsToUser {
					income = income.Add(transaction.Amount)
				}
			case domain.TransactionTypeWithdrawal, domain.TransactionTypeCardPayment, domain.TransactionTypeCreditPayment, domain.TransactionTypePenalty:
				if sourceBelongsToUser {
					expenses = expenses.Add(transaction.Amount)
				}
			case domain.TransactionTypeTransfer:
				if destinationBelongsToUser {
					income = income.Add(transaction.Amount)
				}
				if sourceBelongsToUser {
					expenses = expenses.Add(transaction.Amount)
				}
			}
		}
	}

	netCashFlow, err := subtractMoneyOrZero(income, expenses)
	if err != nil {
		netCashFlow = money.Zero()
	}

	return &MonthlyAnalyticsResponse{
		PeriodStart:       periodStart,
		PeriodEnd:         periodEnd,
		Income:            income.String(),
		Expenses:          expenses.String(),
		InternalTransfers: internalTransfers.String(),
		NetCashFlow:       netCashFlow.String(),
	}, nil
}

func (s *Service) GetCreditLoad(ctx context.Context, userID domain.UserID, now time.Time) (*CreditLoadResponse, error) {
	credits, err := s.credits.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	activeCredits := 0
	outstanding := money.Zero()
	monthlyPayment := money.Zero()
	overdueAmount := money.Zero()
	overduePayments := 0

	for _, credit := range credits {
		if credit.Status == domain.CreditStatusActive || credit.Status == domain.CreditStatusOverdue {
			activeCredits++
			outstanding = outstanding.Add(credit.OutstandingPrincipal)
			monthlyPayment = monthlyPayment.Add(credit.AnnuityPayment)
		}

		schedules, err := s.schedules.ListByCreditID(ctx, credit.ID)
		if err != nil {
			return nil, err
		}

		for _, schedule := range schedules {
			if schedule.Status == domain.PaymentScheduleStatusOverdue ||
				(schedule.Status == domain.PaymentScheduleStatusPending && !schedule.DueDate.After(now)) {
				overduePayments++
				overdueAmount = overdueAmount.Add(schedule.TotalAmount)
			}
		}
	}

	return &CreditLoadResponse{
		ActiveCredits:         activeCredits,
		OutstandingPrincipal:  outstanding.String(),
		MonthlyPayment:        monthlyPayment.String(),
		OverduePayments:       overduePayments,
		OverdueAmount:         overdueAmount.String(),
		CreditLoadDescription: describeCreditLoad(activeCredits, overduePayments),
	}, nil
}

func (s *Service) PredictBalance(ctx context.Context, userID domain.UserID, accountID domain.AccountID, days int, now time.Time) (*BalancePredictionResponse, error) {
	if days <= 0 || days > 365 {
		return nil, domain.NewDomainError(domain.ErrorCodeValidation, "prediction period must be between 1 and 365 days", domain.ErrValidation)
	}

	account, err := s.accounts.FindByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	if account.UserID != userID {
		return nil, domain.NewDomainError(domain.ErrorCodeForbidden, "account does not belong to current user", domain.ErrForbidden)
	}

	targetDate := now.AddDate(0, 0, days)
	currentBalance := account.Balance
	predictedBalance := account.Balance
	plannedPayments := money.Zero()
	items := []BalancePredictionItem{}

	credits, err := s.credits.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	for _, credit := range credits {
		if credit.RepaymentAccountID != accountID {
			continue
		}

		schedules, err := s.schedules.ListByCreditID(ctx, credit.ID)
		if err != nil {
			return nil, err
		}

		for _, schedule := range schedules {
			if schedule.Status == domain.PaymentScheduleStatusPaid ||
				schedule.Status == domain.PaymentScheduleStatusCancelled {
				continue
			}

			if schedule.DueDate.Before(now) || schedule.DueDate.After(targetDate) {
				continue
			}

			plannedPayments = plannedPayments.Add(schedule.TotalAmount)

			nextBalance, err := predictedBalance.Sub(schedule.TotalAmount)
			if err != nil {
				nextBalance = money.Zero()
			}

			predictedBalance = nextBalance

			items = append(items, BalancePredictionItem{
				Date:         schedule.DueDate,
				Description:  "Credit payment #" + intToString(schedule.InstallmentNo),
				Amount:       schedule.TotalAmount.String(),
				BalanceAfter: predictedBalance.String(),
			})
		}
	}

	return &BalancePredictionResponse{
		AccountID:        account.ID.String(),
		Days:             days,
		CurrentBalance:   currentBalance.String(),
		PredictedBalance: predictedBalance.String(),
		PlannedPayments:  plannedPayments.String(),
		Items:            items,
	}, nil
}

func subtractMoneyOrZero(left money.Money, right money.Money) (money.Money, error) {
	if left.MinorUnits() < right.MinorUnits() {
		return money.Zero(), nil
	}

	return left.Sub(right)
}

func describeCreditLoad(activeCredits int, overduePayments int) string {
	if activeCredits == 0 {
		return "no active credits"
	}

	if overduePayments > 0 {
		return "has overdue payments"
	}

	return "normal"
}

func intToString(value int) string {
	result := ""
	if value == 0 {
		return "0"
	}

	for value > 0 {
		digit := value % 10
		result = string(rune('0'+digit)) + result
		value = value / 10
	}

	return result
}
