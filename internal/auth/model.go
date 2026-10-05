package auth

import (
	"time"

	"github.com/xixona38/finance-tracker/internal/user"
)

type LoginResult struct {
	User      user.User
	Token     string
	ExpiresAt time.Time
}
