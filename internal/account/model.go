package account

import "time"

type Account struct {
	ID             int64     `json:"id"`
	UserID         int64     `json:"user_id"`
	Name           string    `json:"name"`
	Type           string    `json:"type"`
	Currency       string    `json:"currency"`
	InitialBalance int64     `json:"initial_balance"`
	CreatedAt      time.Time `json:"created_at"`
}
