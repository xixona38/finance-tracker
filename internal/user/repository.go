package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository sets up user storage using the given database pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

// Create saves the user's email and password hash and returns their assigned ID and creation time.
// It returns ErrMailAlreadyExists if another user already has that email.
func (r *Repository) Create(ctx context.Context, user User) (*User, error) {
	query := `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, created_at;
	`

	err := r.pool.QueryRow(ctx, query, user.Email, user.PasswordHash).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "users_email_key" {
				return nil, ErrMailAlreadyExists
			}
		}
		return nil, fmt.Errorf("failed to execute sql query: %w", err)

	}

	return &user, nil
}

// GetByEmail finds a user by their exact email address and includes their saved password hash.
// It returns ErrUserNotFound if no user has that email.
func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var foundUser User

	query := `
		SELECT id, email, password_hash, created_at FROM users
		WHERE email=$1;
	`

	if err := r.pool.QueryRow(ctx, query, email).Scan(
		&foundUser.ID,
		&foundUser.Email,
		&foundUser.PasswordHash,
		&foundUser.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to execute sql query: %w", err)
	}

	return &foundUser, nil
}

// GetByID loads the user's ID, email, and creation time without reading the password hash.
// It returns ErrUserNotFound if no user has that ID.
func (r *Repository) GetByID(ctx context.Context, userID int64) (*User, error) {
	query := `
		SELECT id, email, created_at FROM users
		WHERE id = $1;
	`
	var user User
	if err := r.pool.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to execute sql query: %w", err)
	}
	return &user, nil
}
