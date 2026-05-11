package credits

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/example/banking-service/internal/domain"
	"github.com/example/banking-service/pkg/money"
)

type KeyRateProvider interface {
	GetKeyRate(ctx context.Context) (string, error)
}

type Service struct {
	tx           domain.TxManager
	accounts     domain.AccountRepository
	credits      domain.CreditRepository
	schedules    domain.PaymentScheduleRepository
	transactions domain.TransactionRepository
	keyRates     KeyRateProvider
}

func NewService(
	tx domain.TxManager,
	accounts domain.AccountRepository,
	credits domain.CreditRepository,
	schedules domain.PaymentScheduleRepository,
	transactions domain.TransactionRepository,
	keyRates KeyRateProvider,
) *Service {
	return &Service{
		tx:           tx,
		accounts:     accounts,
		credits:      credits,
		schedules:    schedules,
		transactions: transactions,
		keyRates:     keyRates,
	}
}

func (s *Service) IssueCredit(ctx context.Context, userID domain.UserID, req IssueCreditRequest) (*IssueCreditResponse, error) {
	principal, err := parsePositiveMoney(req.PrincipalAmount, "principal amount")
	if err != nil {
		return nil, err
	}

	if req.TermMonths <= 0 {
		return nil, domain.NewDomainError(domain.ErrorCodeValidation, "term months must be greater than zero", domain.ErrValidation)
	}

	bankMargin := req.BankMargin
	if bankMargin == "" {
		bankMargin = "4.0000"
	}

	cbrKeyRate, err := s.keyRates.GetKeyRate(ctx)
	if err != nil {
		return nil, err
	}

	annualRate, err := sumRates(cbrKeyRate, bankMargin)
	if err != nil {
		return nil, err
	}

	annuityPayment, err := CalculateAnnuityPayment(principal, annualRate, req.TermMonths)
	if err != nil {
		return nil, err
	}

	disbursementAccountID := domain.AccountID(req.DisbursementAccountID)
	repaymentAccountID := domain.AccountID(req.RepaymentAccountID)

	var response *IssueCreditResponse

	err = s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		disbursementAccount, err := s.accounts.FindByIDForUpdate(txCtx, disbursementAccountID)
		if err != nil {
			return err
		}

		if disbursementAccount.UserID != userID {
			return domain.NewDomainError(domain.ErrorCodeForbidden, "disbursement account does not belong to current user", domain.ErrForbidden)
		}

		repaymentAccount, err := s.accounts.FindByID(txCtx, repaymentAccountID)
		if err != nil {
			return err
		}

		if repaymentAccount.UserID != userID {
			return domain.NewDomainError(domain.ErrorCodeForbidden, "repayment account does not belong to current user", domain.ErrForbidden)
		}

		credit, err := domain.NewCredit(domain.CreateCreditData{
			UserID:                userID,
			DisbursementAccountID: disbursementAccountID,
			RepaymentAccountID:    repaymentAccountID,
			PrincipalAmount:       principal,
			AnnualInterestRate:    annualRate,
			CBRKeyRate:            cbrKeyRate,
			BankMargin:            bankMargin,
			TermMonths:            req.TermMonths,
			AnnuityPayment:        annuityPayment,
		})
		if err != nil {
			return err
		}

		if err := s.credits.Create(txCtx, credit); err != nil {
			return err
		}

		paymentSchedules, err := generatePaymentSchedule(credit)
		if err != nil {
			return err
		}

		if err := s.schedules.CreateBatch(txCtx, paymentSchedules); err != nil {
			return err
		}

		if err := disbursementAccount.Credit(principal); err != nil {
			return err
		}

		if err := s.accounts.UpdateBalance(txCtx, disbursementAccount.ID, disbursementAccount); err != nil {
			return err
		}

		transaction, err := domain.NewCompletedTransaction(domain.CreateTransactionData{
			UserID:               userID,
			DestinationAccountID: &disbursementAccountID,
			CreditID:             &credit.ID,
			OperationType:        domain.TransactionTypeCreditDisbursement,
			Amount:               principal,
			Description:          "Credit disbursement",
			IdempotencyKey:       "credit-disbursement-" + credit.ID.String(),
		})
		if err != nil {
			return err
		}

		if err := s.transactions.Create(txCtx, transaction); err != nil {
			return err
		}

		response = &IssueCreditResponse{
			Credit:   ToCreditResponse(credit),
			Schedule: ToPaymentScheduleResponses(paymentSchedules),
		}

		return nil
	})

	return response, err
}

func (s *Service) ListCredits(ctx context.Context, userID domain.UserID) ([]CreditResponse, error) {
	credits, err := s.credits.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]CreditResponse, 0, len(credits))
	for _, credit := range credits {
		result = append(result, ToCreditResponse(credit))
	}

	return result, nil
}

func (s *Service) GetSchedule(ctx context.Context, userID domain.UserID, creditID domain.CreditID) ([]PaymentScheduleResponse, error) {
	credit, err := s.credits.FindByID(ctx, creditID)
	if err != nil {
		return nil, err
	}

	if credit.UserID != userID {
		return nil, domain.NewDomainError(domain.ErrorCodeForbidden, "credit does not belong to current user", domain.ErrForbidden)
	}

	schedules, err := s.schedules.ListByCreditID(ctx, creditID)
	if err != nil {
		return nil, err
	}

	return ToPaymentScheduleResponses(schedules), nil
}

func (s *Service) ProcessDuePayments(ctx context.Context, now time.Time, limit int) (*ProcessPaymentsResult, error) {
	if limit <= 0 {
		limit = 100
	}

	dueSchedules, err := s.schedules.FindDuePayments(ctx, now, limit)
	if err != nil {
		return nil, err
	}

	result := &ProcessPaymentsResult{}

	for _, due := range dueSchedules {
		result.Processed++

		err := s.processOnePayment(ctx, due.ID, now)
		if err == nil {
			result.Paid++
			continue
		}

		if isInsufficientFunds(err) {
			result.Penalized++
			continue
		}

		result.Failed++
	}

	return result, nil
}

func (s *Service) processOnePayment(ctx context.Context, scheduleID domain.PaymentScheduleID, now time.Time) error {
	return s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		schedule, err := s.schedules.FindByIDForUpdate(txCtx, scheduleID)
		if err != nil {
			return err
		}

		if schedule.IsPaid() {
			return nil
		}

		credit, err := s.credits.FindByID(txCtx, schedule.CreditID)
		if err != nil {
			return err
		}

		account, err := s.accounts.FindByIDForUpdate(txCtx, credit.RepaymentAccountID)
		if err != nil {
			return err
		}

		if err := account.Debit(schedule.TotalAmount); err != nil {
			if isInsufficientFunds(err) {
				return s.applyPenalty(txCtx, credit, schedule, now)
			}

			return err
		}

		if err := s.accounts.UpdateBalance(txCtx, account.ID, account); err != nil {
			return err
		}

		accountID := account.ID
		creditID := credit.ID

		transaction, err := domain.NewCompletedTransaction(domain.CreateTransactionData{
			UserID:          credit.UserID,
			SourceAccountID: &accountID,
			CreditID:        &creditID,
			OperationType:   domain.TransactionTypeCreditPayment,
			Amount:          schedule.TotalAmount,
			Description:     fmt.Sprintf("Credit payment #%d", schedule.InstallmentNo),
			IdempotencyKey:  "credit-payment-" + schedule.ID.String(),
		})
		if err != nil {
			return err
		}

		if err := s.transactions.Create(txCtx, transaction); err != nil {
			return err
		}

		paidAmount := schedule.TotalAmount
		schedule.PaidAmount = paidAmount
		schedule.Status = domain.PaymentScheduleStatusPaid
		schedule.PaidTransactionID = &transaction.ID

		if err := s.schedules.Update(txCtx, schedule); err != nil {
			return err
		}

		newOutstanding, err := credit.OutstandingPrincipal.Sub(schedule.PrincipalAmount)
		if err != nil {
			newOutstanding = money.Zero()
		}

		credit.OutstandingPrincipal = newOutstanding

		if err := s.credits.UpdateOutstandingPrincipal(txCtx, credit.ID, credit); err != nil {
			return err
		}

		if credit.OutstandingPrincipal.IsZero() {
			if err := s.credits.UpdateStatus(txCtx, credit.ID, domain.CreditStatusClosed); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *Service) applyPenalty(ctx context.Context, credit *domain.Credit, schedule *domain.PaymentSchedule, now time.Time) error {
	penalty, err := calculatePenalty(schedule.TotalAmount)
	if err != nil {
		return err
	}

	schedule.PenaltyAmount = schedule.PenaltyAmount.Add(penalty)
	schedule.TotalAmount = schedule.TotalAmount.Add(penalty)
	schedule.Status = domain.PaymentScheduleStatusOverdue
	schedule.LastPenaltyAppliedAt = &now

	if err := s.schedules.Update(ctx, schedule); err != nil {
		return err
	}

	if credit.Status != domain.CreditStatusOverdue {
		if err := s.credits.UpdateStatus(ctx, credit.ID, domain.CreditStatusOverdue); err != nil {
			return err
		}
	}

	return domain.NewDomainError(domain.ErrorCodeInsufficientFunds, "insufficient funds for credit payment, penalty applied", domain.ErrInsufficientFunds)
}

func generatePaymentSchedule(credit *domain.Credit) ([]*domain.PaymentSchedule, error) {
	schedules := make([]*domain.PaymentSchedule, 0, credit.TermMonths)
	now := credit.IssuedAt

	principalPartMinor := credit.PrincipalAmount.MinorUnits() / int64(credit.TermMonths)
	remainder := credit.PrincipalAmount.MinorUnits() % int64(credit.TermMonths)

	for i := 1; i <= credit.TermMonths; i++ {
		principalMinor := principalPartMinor
		if i == credit.TermMonths {
			principalMinor += remainder
		}

		principalPart, err := money.FromMinorUnits(principalMinor)
		if err != nil {
			return nil, err
		}

		interestPart, err := credit.AnnuityPayment.Sub(principalPart)
		if err != nil {
			interestPart = money.Zero()
		}

		schedules = append(schedules, &domain.PaymentSchedule{
			CreditID:        credit.ID,
			InstallmentNo:   i,
			DueDate:         now.AddDate(0, i, 0),
			PrincipalAmount: principalPart,
			InterestAmount:  interestPart,
			PenaltyAmount:   money.Zero(),
			TotalAmount:     credit.AnnuityPayment,
			PaidAmount:      money.Zero(),
			Status:          domain.PaymentScheduleStatusPending,
			CreatedAt:       now,
			UpdatedAt:       now,
		})
	}

	return schedules, nil
}

func parsePositiveMoney(value string, fieldName string) (money.Money, error) {
	amount, err := money.FromRubString(value)
	if err != nil {
		return money.Money{}, domain.NewDomainError(domain.ErrorCodeValidation, fieldName+" is invalid", err)
	}

	if !amount.IsPositive() {
		return money.Money{}, domain.NewDomainError(domain.ErrorCodeValidation, fieldName+" must be positive", domain.ErrValidation)
	}

	return amount, nil
}

func sumRates(left string, right string) (string, error) {
	leftValue, err := strconv.ParseFloat(left, 64)
	if err != nil {
		return "", domain.NewDomainError(domain.ErrorCodeValidation, "cbr key rate is invalid", err)
	}

	rightValue, err := strconv.ParseFloat(right, 64)
	if err != nil {
		return "", domain.NewDomainError(domain.ErrorCodeValidation, "bank margin is invalid", err)
	}

	return fmt.Sprintf("%.4f", leftValue+rightValue), nil
}

func calculatePenalty(amount money.Money) (money.Money, error) {
	penaltyMinor := amount.MinorUnits() / 10
	if penaltyMinor <= 0 {
		penaltyMinor = 1
	}

	return money.FromMinorUnits(penaltyMinor)
}

func isInsufficientFunds(err error) bool {
	return errors.Is(err, domain.ErrInsufficientFunds)
}
