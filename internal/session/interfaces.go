package session

import (
	"context"
	"time"
)

type Repository interface {
	Create(ctx context.Context, session *Session) error
	FindByHash(ctx context.Context, hash string) (*Session, error)
	Delete(ctx context.Context, hash string) error
	Extend(ctx context.Context, hash string, expiresAt time.Time) (*Session, error)
}
