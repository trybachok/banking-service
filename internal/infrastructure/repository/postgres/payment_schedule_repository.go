package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/example/banking-service/internal/domain"
	"github.com/example/banking-service/internal/infrastructure/db"
	"github.com/example/banking-service/pkg/money"
)

type PaymentScheduleRepository struct {
	tx *db.TxManager
}

func NewPaymentScheduleRepository(tx *db.TxManager) *PaymentScheduleRepository {
	return &PaymentScheduleRepository{tx: tx}
}

func (r *PaymentScheduleRepository) CreateBatch(ctx context.Context, schedules []*domain.PaymentSchedule) error {
	query := `
		INSERT INTO payment_schedules (
			credit_id,
			installment_no,
			due_date,
			principal_amount,
			interest_amount,
			penalty_amount,
			total_amount,
			paid_amount,
			status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	for _, schedule := range schedules {
		err := r.tx.Executor(ctx).QueryRowContext(
			ctx,
			query,
			schedule.CreditID.String(),
			schedule.InstallmentNo,
			schedule.DueDate,
			schedule.PrincipalAmount.String(),
			schedule.InterestAmount.String(),
			schedule.PenaltyAmount.String(),
			schedule.TotalAmount.String(),
			schedule.PaidAmount.String(),
			string(schedule.Status),
		).Scan(&schedule.ID, &schedule.CreatedAt, &schedule.UpdatedAt)

		if err != nil {
			return mapDBError(err)
		}
	}

	return nil
}

func (r *PaymentScheduleRepository) ListByCreditID(ctx context.Context, creditID domain.CreditID) ([]*domain.PaymentSchedule, error) {
	query := basePaymentScheduleSelect() + " WHERE credit_id = $1 ORDER BY installment_no ASC"

	rows, err := r.tx.Executor(ctx).QueryContext(ctx, query, creditID.String())
	if err != nil {
		return nil, mapDBError(err)
	}
	defer rows.Close()

	return scanPaymentSchedules(rows)
}

func (r *PaymentScheduleRepository) FindDuePayments(ctx context.Context, now time.Time, limit int) ([]*domain.PaymentSchedule, error) {
	query := basePaymentScheduleSelect() + `
		WHERE status IN ('pending', 'overdue')
		  AND due_date <= $1
		ORDER BY due_date ASC
		LIMIT $2
	`

	rows, err := r.tx.Executor(ctx).QueryContext(ctx, query, now, limit)
	if err != nil {
		return nil, mapDBError(err)
	}
	defer rows.Close()

	return scanPaymentSchedules(rows)
}

func (r *PaymentScheduleRepository) FindByIDForUpdate(ctx context.Context, id domain.PaymentScheduleID) (*domain.PaymentSchedule, error) {
	query := basePaymentScheduleSelect() + " WHERE id = $1 FOR UPDATE"

	schedule, err := scanPaymentSchedule(r.tx.Executor(ctx).QueryRowContext(ctx, query, id.String()))
	if err != nil {
		return nil, mapDBError(err)
	}

	return schedule, nil
}

func (r *PaymentScheduleRepository) Update(ctx context.Context, schedule *domain.PaymentSchedule) error {
	query := `
		UPDATE payment_schedules
		SET penalty_amount = $2,
		    total_amount = $3,
		    paid_amount = $4,
		    status = $5,
		    last_penalty_applied_at = $6,
		    paid_transaction_id = $7
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.tx.Executor(ctx).QueryRowContext(
		ctx,
		query,
		schedule.ID.String(),
		schedule.PenaltyAmount.String(),
		schedule.TotalAmount.String(),
		schedule.PaidAmount.String(),
		string(schedule.Status),
		schedule.LastPenaltyAppliedAt,
		transactionIDPtrValue(schedule.PaidTransactionID),
	).Scan(&schedule.UpdatedAt)

	return mapDBError(err)
}

func basePaymentScheduleSelect() string {
	return `
		SELECT
			id,
			credit_id,
			installment_no,
			due_date,
			principal_amount::text,
			interest_amount::text,
			penalty_amount::text,
			total_amount::text,
			paid_amount::text,
			status,
			last_penalty_applied_at,
			paid_transaction_id,
			created_at,
			updated_at
		FROM payment_schedules
	`
}

type paymentScheduleScanner interface {
	Scan(dest ...any) error
}

func scanPaymentSchedule(scanner paymentScheduleScanner) (*domain.PaymentSchedule, error) {
	var schedule domain.PaymentSchedule
	var principalRaw string
	var interestRaw string
	var penaltyRaw string
	var totalRaw string
	var paidRaw string
	var lastPenaltyAt sql.NullTime
	var paidTransactionID sql.NullString

	err := scanner.Scan(
		&schedule.ID,
		&schedule.CreditID,
		&schedule.InstallmentNo,
		&schedule.DueDate,
		&principalRaw,
		&interestRaw,
		&penaltyRaw,
		&totalRaw,
		&paidRaw,
		&schedule.Status,
		&lastPenaltyAt,
		&paidTransactionID,
		&schedule.CreatedAt,
		&schedule.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	amounts, err := scanScheduleAmounts(principalRaw, interestRaw, penaltyRaw, totalRaw, paidRaw)
	if err != nil {
		return nil, err
	}

	schedule.PrincipalAmount = amounts[0]
	schedule.InterestAmount = amounts[1]
	schedule.PenaltyAmount = amounts[2]
	schedule.TotalAmount = amounts[3]
	schedule.PaidAmount = amounts[4]
	schedule.LastPenaltyAppliedAt = timePtr(lastPenaltyAt)
	schedule.PaidTransactionID = transactionIDPtr(paidTransactionID)

	return &schedule, nil
}

func scanScheduleAmounts(values ...string) ([]money.Money, error) {
	result := make([]money.Money, 0, len(values))

	for _, value := range values {
		amount, err := scanMoney(value)
		if err != nil {
			return nil, err
		}

		result = append(result, amount)
	}

	return result, nil
}

func scanPaymentSchedules(rows *sql.Rows) ([]*domain.PaymentSchedule, error) {
	var schedules []*domain.PaymentSchedule

	for rows.Next() {
		schedule, err := scanPaymentSchedule(rows)
		if err != nil {
			return nil, err
		}

		schedules = append(schedules, schedule)
	}

	return schedules, mapDBError(rows.Err())
}
