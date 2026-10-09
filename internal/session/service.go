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
	ErrInvalidIdleTTL     = errors.New("invalid session's idle time to live")
	ErrInvalidAbsoluteTTL = errors.New("invalid session's absolute time to live")
	ErrSessionExpired     = errors.New("session expired")
)

type Service struct {
	repo        *SessionRepository
	idleTTL     time.Duration
	absoluteTTL time.Duration
}

// NewService sets how long a session may stay idle and how long it may exist in total.
// The idle timeout must be between five minutes and one hour.
// The total lifetime must be between five and 24 hours.
func NewService(repo *SessionRepository, idleTTL, absoluteTTL time.Duration) (*Service, error) {
	if idleTTL < time.Minute*5 || idleTTL > time.Hour {
		return nil, ErrInvalidIdleTTL
	}
	if absoluteTTL < time.Hour*5 || absoluteTTL > time.Hour*24 {
		return nil, ErrInvalidAbsoluteTTL
	}

	return &Service{
		repo:        repo,
		idleTTL:     idleTTL,
		absoluteTTL: absoluteTTL,
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
	expiresAt := createdAt.Add(s.idleTTL)

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

// Validate finds the session for a token and checks both its idle timeout and total lifetime.
// It returns an error if the session is missing, has expired, or cannot be read from the database.
func (s *Service) Validate(ctx context.Context, token string) (*Session, error) {
	tokenHashBytes := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(tokenHashBytes[:])

	session, err := s.repo.FindByHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	absExpiresAt := session.CreatedAt.Add(s.absoluteTTL)
	current := time.Now()
	if current.Before(session.ExpiresAt) && current.Before(absExpiresAt) {
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

// Renew checks the session and requests a later expiration time after user activity.
// It limits the requested expiration time to the session's maximum total lifetime and keeps the same token.
func (s *Service) Renew(ctx context.Context, token string) (*Session, error) {
	session, err := s.Validate(ctx, token)
	if err != nil {
		return nil, err
	}

	absoluteExpiresAt := session.CreatedAt.Add(s.absoluteTTL)
	wantedExpiresAt := time.Now().Add(s.idleTTL)

	if wantedExpiresAt.Before(absoluteExpiresAt) {
		return s.repo.Extend(ctx, session.TokenHash, wantedExpiresAt)
	}

	return s.repo.Extend(ctx, session.TokenHash, absoluteExpiresAt)
}
