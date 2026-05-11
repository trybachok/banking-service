package transfers

import (
	"context"

	accountapp "github.com/example/banking-service/internal/application/accounts"
	"github.com/example/banking-service/internal/domain"
)

type Service struct {
	tx           domain.TxManager
	accounts     domain.AccountRepository
	transactions domain.TransactionRepository
}

func NewService(
	tx domain.TxManager,
	accounts domain.AccountRepository,
	transactions domain.TransactionRepository,
) *Service {
	return &Service{
		tx:           tx,
		accounts:     accounts,
		transactions: transactions,
	}
}

func (s *Service) Transfer(ctx context.Context, userID domain.UserID, req TransferRequest) (*TransferResponse, error) {
	amount, err := accountapp.ParseAmountForTransfer(req.Amount)
	if err != nil {
		return nil, err
	}

	sourceAccountID := domain.AccountID(req.SourceAccountID)
	destinationAccountID := domain.AccountID(req.DestinationAccountID)

	if sourceAccountID == destinationAccountID {
		return nil, domain.NewDomainError(domain.ErrorCodeValidation, "source and destination accounts must be different", domain.ErrValidation)
	}

	var response *TransferResponse

	err = s.tx.WithinTx(ctx, func(txCtx context.Context) error {
		sourceAccount, err := s.accounts.FindByIDForUpdate(txCtx, sourceAccountID)
		if err != nil {
			return err
		}

		if sourceAccount.UserID != userID {
			return domain.NewDomainError(domain.ErrorCodeForbidden, "source account does not belong to current user", domain.ErrForbidden)
		}

		destinationAccount, err := s.accounts.FindByIDForUpdate(txCtx, destinationAccountID)
		if err != nil {
			return err
		}

		if err := sourceAccount.Debit(amount); err != nil {
			return err
		}

		if err := destinationAccount.Credit(amount); err != nil {
			return err
		}

		if err := s.accounts.UpdateBalance(txCtx, sourceAccount.ID, sourceAccount); err != nil {
			return err
		}

		if err := s.accounts.UpdateBalance(txCtx, destinationAccount.ID, destinationAccount); err != nil {
			return err
		}

		transaction, err := domain.NewCompletedTransaction(domain.CreateTransactionData{
			UserID:               userID,
			SourceAccountID:      &sourceAccountID,
			DestinationAccountID: &destinationAccountID,
			OperationType:        domain.TransactionTypeTransfer,
			Amount:               amount,
			Description:          req.Description,
			IdempotencyKey:       req.IdempotencyKey,
		})
		if err != nil {
			return err
		}

		if err := s.transactions.Create(txCtx, transaction); err != nil {
			return err
		}

		response = &TransferResponse{
			SourceAccount:      accountapp.ToAccountResponse(sourceAccount),
			DestinationAccount: accountapp.ToAccountResponse(destinationAccount),
			Transaction:        accountapp.ToTransactionResponse(transaction),
		}

		return nil
	})

	return response, err
}
