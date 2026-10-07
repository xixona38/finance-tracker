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

// NewService sets up session management with the given session lifetime.
// The lifetime must be at least five minutes and no more than one hour.
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

// Create starts a session for the user and generates a random token for the client.
// Only the token's hash is saved in the database; the original token is returned with the session.
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

// Validate finds the session for a token and checks that it has not expired.
// It returns an error if the session is missing, has expired, or cannot be read from the database.
func (s *Service) Validate(ctx context.Context, token string) (*Session, error) {
	tokenHashBytes := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(tokenHashBytes[:])

	session, err := s.repo.FindByHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	if time.Now().Before(session.ExpiresAt) {
		return session, nil
	}

	return nil, ErrSessionExpired
}

// Revoke deletes the session for a token so the token can no longer be used.
// It also succeeds if the session has already been deleted.
func (s *Service) Revoke(ctx context.Context, token string) error {
	tokenHashBytes := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(tokenHashBytes[:])

	if err := s.repo.Delete(ctx, tokenHash); err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}
	return nil
}
