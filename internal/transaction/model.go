package transaction

import (
	"errors"
	"time"
)

type Transaction struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	FromAccID   *int64    `json:"from_acc_id"`
	ToAccID     *int64    `json:"to_acc_id"`
	Type        string    `json:"type"`
	Amount      int64     `json:"amount"`
	Description string    `json:"description"`
	OccurredAt  time.Time `json:"occurred_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type ListFilter struct {
	AccountID *int64
	Type      string
	Limit     int
	Offset    int
	From      *time.Time
	To        *time.Time
}

var (
	ErrTransactionNotFound = errors.New("transaction  not found")
	ErrInvalidData         = errors.New("invalid data entry")
)
