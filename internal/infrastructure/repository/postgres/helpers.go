package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/example/banking-service/internal/domain"
	"github.com/example/banking-service/pkg/money"
)

func scanMoney(value string) (money.Money, error) {
	amount, err := money.FromRubString(value)
	if err != nil {
		return money.Money{}, fmt.Errorf("scan money %q: %w", value, err)
	}

	return amount, nil
}

func accountIDPtrValue(value *domain.AccountID) any {
	if value == nil {
		return nil
	}

	return value.String()
}

func cardIDPtrValue(value *domain.CardID) any {
	if value == nil {
		return nil
	}

	return value.String()
}

func creditIDPtrValue(value *domain.CreditID) any {
	if value == nil {
		return nil
	}

	return value.String()
}

func transactionIDPtrValue(value *domain.TransactionID) any {
	if value == nil {
		return nil
	}

	return value.String()
}

func userIDPtrValue(value *domain.UserID) any {
	if value == nil {
		return nil
	}

	return value.String()
}

func accountIDPtr(value sql.NullString) *domain.AccountID {
	if !value.Valid {
		return nil
	}

	id := domain.AccountID(value.String)
	return &id
}

func cardIDPtr(value sql.NullString) *domain.CardID {
	if !value.Valid {
		return nil
	}

	id := domain.CardID(value.String)
	return &id
}

func creditIDPtr(value sql.NullString) *domain.CreditID {
	if !value.Valid {
		return nil
	}

	id := domain.CreditID(value.String)
	return &id
}

func transactionIDPtr(value sql.NullString) *domain.TransactionID {
	if !value.Valid {
		return nil
	}

	id := domain.TransactionID(value.String)
	return &id
}

func userIDPtr(value sql.NullString) *domain.UserID {
	if !value.Valid {
		return nil
	}

	id := domain.UserID(value.String)
	return &id
}

func timePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}

	return &value.Time
}

func marshalJSONMap(value map[string]any) ([]byte, error) {
	if value == nil {
		value = map[string]any{}
	}

	return json.Marshal(value)
}

func unmarshalJSONMap(raw []byte) (map[string]any, error) {
	if len(raw) == 0 {
		return map[string]any{}, nil
	}

	var value map[string]any
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}

	return value, nil
}
