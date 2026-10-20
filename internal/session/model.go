package session

import (
	"errors"
	"time"
)

type Session struct {
	TokenHash string
	UserID    int64
	CreatedAt time.Time
	ExpiresAt time.Time
}

var (
	ErrSessionNotFound    = errors.New("session not found")
	ErrInvalidIdleTTL     = errors.New("invalid session's idle time to live")
	ErrInvalidAbsoluteTTL = errors.New("invalid session's absolute time to live")
	ErrSessionExpired     = errors.New("session expired")
)
