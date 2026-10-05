package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrMailAlreadyExists = errors.New("email already exists")
	ErrUserNotFound      = errors.New("user not found")
)

type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a user repository using an existing PostgreSQL connection pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

// Create stores the email and password hash and returns the user with a database-generated
// ID and creation timestamp. It returns ErrMailAlreadyExists for a duplicate email.
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

// GetByEmail retrieves a user, including the password hash, by exact email match.
// It returns ErrUserNotFound when no matching user exists.
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
