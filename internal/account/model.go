package account

import "time"

type Account struct {
	ID             int64
	UserID         int64
	Name           string
	Type           string
	Currency       string
	InitialBalance int64
	CreatedAt      time.Time
}
