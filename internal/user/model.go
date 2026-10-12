package user

import (
	"errors"
	"time"
)

type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

var (
	ErrInvalidData       = errors.New("invalid data entry")
	ErrMailAlreadyExists = errors.New("email already exists")
	ErrUserNotFound      = errors.New("user not found")
)
