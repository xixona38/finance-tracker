package session

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrSessionNotFound = errors.New("session not found")
)

type SessionRepository struct {
	pool *pgxpool.Pool
}

// NewSessionRepo creates a session repository using an existing PostgreSQL connection pool.
func NewSessionRepo(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{
		pool: pool,
	}
}

// Create stores the supplied session's token hash, user ID, and timestamps.
func (s *SessionRepository) Create(ctx context.Context, session *Session) error {
	query := `
		INSERT INTO sessions (token_hash, user_id, created_at, expires_at)
		VALUES ($1, $2, $3, $4)
	`

	if _, err := s.pool.Exec(ctx, query,
		session.TokenHash,
		session.UserID,
		session.CreatedAt,
		session.ExpiresAt,
	); err != nil {
		return fmt.Errorf("failed to execute sql query: %w", err)
	}

	return nil
}

// FindByHash retrieves a session by its stored token hash.
// It returns ErrSessionNotFound when no matching record exists.
func (s *SessionRepository) FindByHash(ctx context.Context, hash string) (*Session, error) {
	query := `
		SELECT token_hash, user_id, created_at, expires_at FROM sessions
		WHERE token_hash=$1
	`

	var session Session

	err := s.pool.QueryRow(ctx, query, hash).Scan(&session.TokenHash, &session.UserID, &session.CreatedAt, &session.ExpiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to execute sql query: %w", err)
	}

	return &session, nil

}

// Delete removes a session by its token hash and wraps SQL execution errors.
// It returns nil if the record is already absent.
func (s *SessionRepository) Delete(ctx context.Context, hash string) error {
	query := `
		DELETE FROM sessions
		WHERE token_hash=$1
	`

	if _, err := s.pool.Exec(ctx, query, hash); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}
