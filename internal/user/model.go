package user

import "time"

type User struct {
	ID           int64
	Email        string
	PasswordHash string `json:"-"`
	CreatedAt    time.Time
}
