package postgres

import (
	"context"
	"database/sql"

	"github.com/example/banking-service/internal/domain"
	"github.com/example/banking-service/internal/infrastructure/db"
)

type CreditRepository struct {
	tx *db.TxManager
}

func NewCreditRepository(tx *db.TxManager) *CreditRepository {
	return &CreditRepository{tx: tx}
}

func (r *CreditRepository) Create(ctx context.Context, credit *domain.Credit) error {
	query := `
		INSERT INTO credits (
			user_id,
			disbursement_account_id,
			repayment_account_id,
			principal_amount,
			outstanding_principal,
			annual_interest_rate,
			cbr_key_rate,
			bank_margin,
			term_months,
			annuity_payment,
			penalty_rate,
			status,
			issued_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at
	`

	err := r.tx.Executor(ctx).QueryRowContext(
		ctx,
		query,
		credit.UserID.String(),
		credit.DisbursementAccountID.String(),
		credit.RepaymentAccountID.String(),
		credit.PrincipalAmount.String(),
		credit.OutstandingPrincipal.String(),
		credit.AnnualInterestRate,
		credit.CBRKeyRate,
		credit.BankMargin,
		credit.TermMonths,
		credit.AnnuityPayment.String(),
		credit.PenaltyRate,
		string(credit.Status),
		credit.IssuedAt,
	).Scan(&credit.ID, &credit.CreatedAt, &credit.UpdatedAt)

	return mapDBError(err)
}

func (r *CreditRepository) FindByID(ctx context.Context, id domain.CreditID) (*domain.Credit, error) {
	query := baseCreditSelect() + " WHERE id = $1"

	credit, err := scanCredit(r.tx.Executor(ctx).QueryRowContext(ctx, query, id.String()))
	if err != nil {
		return nil, mapDBError(err)
	}

	return credit, nil
}

func (r *CreditRepository) ListByUserID(ctx context.Context, userID domain.UserID) ([]*domain.Credit, error) {
	query := baseCreditSelect() + " WHERE user_id = $1 ORDER BY issued_at DESC"

	rows, err := r.tx.Executor(ctx).QueryContext(ctx, query, userID.String())
	if err != nil {
		return nil, mapDBError(err)
	}
	defer rows.Close()

	var credits []*domain.Credit

	for rows.Next() {
		credit, err := scanCredit(rows)
		if err != nil {
			return nil, err
		}

		credits = append(credits, credit)
	}

	return credits, mapDBError(rows.Err())
}

func (r *CreditRepository) UpdateStatus(ctx context.Context, id domain.CreditID, status domain.CreditStatus) error {
	query := `
		UPDATE credits
		SET status = $2
		WHERE id = $1
	`

	result, err := r.tx.Executor(ctx).ExecContext(ctx, query, id.String(), string(status))
	if err != nil {
		return mapDBError(err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return domain.NewDomainError(domain.ErrorCodeNotFound, "credit not found", domain.ErrNotFound)
	}

	return nil
}

func (r *CreditRepository) UpdateOutstandingPrincipal(ctx context.Context, id domain.CreditID, credit *domain.Credit) error {
	query := `
		UPDATE credits
		SET outstanding_principal = $2
		WHERE id = $1
		RETURNING updated_at
	`

	err := r.tx.Executor(ctx).QueryRowContext(
		ctx,
		query,
		id.String(),
		credit.OutstandingPrincipal.String(),
	).Scan(&credit.UpdatedAt)

	return mapDBError(err)
}

func baseCreditSelect() string {
	return `
		SELECT
			id,
			user_id,
			disbursement_account_id,
			repayment_account_id,
			principal_amount::text,
			outstanding_principal::text,
			annual_interest_rate::text,
			cbr_key_rate::text,
			bank_margin::text,
			term_months,
			annuity_payment::text,
			penalty_rate::text,
			status,
			issued_at,
			closed_at,
			created_at,
			updated_at
		FROM credits
	`
}

type creditScanner interface {
	Scan(dest ...any) error
}

func scanCredit(scanner creditScanner) (*domain.Credit, error) {
	var credit domain.Credit
	var principalRaw string
	var outstandingRaw string
	var annuityRaw string
	var closedAt sql.NullTime

	err := scanner.Scan(
		&credit.ID,
		&credit.UserID,
		&credit.DisbursementAccountID,
		&credit.RepaymentAccountID,
		&principalRaw,
		&outstandingRaw,
		&credit.AnnualInterestRate,
		&credit.CBRKeyRate,
		&credit.BankMargin,
		&credit.TermMonths,
		&annuityRaw,
		&credit.PenaltyRate,
		&credit.Status,
		&credit.IssuedAt,
		&closedAt,
		&credit.CreatedAt,
		&credit.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	principal, err := scanMoney(principalRaw)
	if err != nil {
		return nil, err
	}

	outstanding, err := scanMoney(outstandingRaw)
	if err != nil {
		return nil, err
	}

	annuity, err := scanMoney(annuityRaw)
	if err != nil {
		return nil, err
	}

	credit.PrincipalAmount = principal
	credit.OutstandingPrincipal = outstanding
	credit.AnnuityPayment = annuity
	credit.ClosedAt = timePtr(closedAt)

	return &credit, nil
}
