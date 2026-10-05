package session

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidTTL     = errors.New("invalid session time to live")
	ErrSessionExpired = errors.New("session expired")
)

type Service struct {
	repo *SessionRepository
	ttl  time.Duration
}

// NewService creates a session service with the supplied lifetime.
// It accepts lifetimes between five minutes and one hour, inclusive.
func NewService(repo *SessionRepository, ttl time.Duration) (*Service, error) {
	if ttl < time.Minute*5 {
		return nil, ErrInvalidTTL
	}

	if ttl > time.Hour {
		return nil, ErrInvalidTTL
	}
	return &Service{
		repo: repo,
		ttl:  ttl,
	}, nil
}

// Create generates a random token and stores a session for the specified user.
// It stores the token's SHA-256 hash and returns the original token and saved session.
func (s *Service) Create(ctx context.Context, userID int64) (string, *Session, error) {
	tokenInBytes := make([]byte, 32)
	rand.Read(tokenInBytes)
	token := hex.EncodeToString(tokenInBytes)
	tokenHash := sha256.Sum256([]byte(token))

	createdAt := time.Now()
	expiresAt := createdAt.Add(s.ttl)

	session := Session{
		UserID:    userID,
		TokenHash: hex.EncodeToString(tokenHash[:]),
		CreatedAt: createdAt,
		ExpiresAt: expiresAt,
	}

	if err := s.repo.Create(ctx, &session); err != nil {
		return "", nil, fmt.Errorf("failed to create a new session: %w", err)
	}

	return token, &session, nil
}

// Validate hashes the token, retrieves its session, and checks the expiration time.
// It returns the session or a missing session, expiration, or database error.
func (s *Service) Validate(ctx context.Context, token string) (*Session, error) {
	tokenHashBytes := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(tokenHashBytes[:])

	session, err := s.repo.FindByHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	if !session.ExpiresAt.Before(time.Now()) {
		return session, nil
	}

	return nil, ErrSessionExpired
}

// Revoke hashes the token and deletes the corresponding session.
// Revoking a session that is already absent succeeds.
func (s *Service) Revoke(ctx context.Context, token string) error {
	tokenHashBytes := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(tokenHashBytes[:])

	if err := s.repo.Delete(ctx, tokenHash); err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}
	return nil
}
