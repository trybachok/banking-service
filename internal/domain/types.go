package domain

type UserID string
type AccountID string
type CardID string
type CreditID string
type TransactionID string
type PaymentScheduleID string
type MFAChallengeID string

func (id UserID) String() string {
	return string(id)
}

func (id AccountID) String() string {
	return string(id)
}

func (id CardID) String() string {
	return string(id)
}

func (id CreditID) String() string {
	return string(id)
}

func (id TransactionID) String() string {
	return string(id)
}

func (id PaymentScheduleID) String() string {
	return string(id)
}

func (id MFAChallengeID) String() string {
	return string(id)
}
