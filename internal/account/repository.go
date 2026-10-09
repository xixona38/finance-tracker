package account

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository sets up account storage using the given database pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

// Create saves a new account and its owner in the database.
// It returns the account with the ID and creation time assigned by the database.
func (r *Repository) Create(ctx context.Context, account Account) (*Account, error) {
	query := `
		INSERT INTO accounts (name, user_id, type, currency, initial_balance)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at;
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		account.Name,
		account.UserID,
		account.Type,
		account.Currency,
		account.InitialBalance,
	).Scan(&account.ID, &account.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to execute sql query: %w", err)
	}

	return &account, nil
}

// List returns one user's accounts, ordered from lowest to highest ID.
// It returns an empty slice if the user has no accounts.
func (r *Repository) List(ctx context.Context, userID int64) ([]Account, error) {
	query := `
		SELECT id, user_id, name, type, currency, initial_balance, created_at FROM accounts
		WHERE user_id=$1
		ORDER BY id
	`

	accounts := make([]Account, 0, 4)

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to execute sql query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var res Account
		err := rows.Scan(
			&res.ID,
			&res.UserID,
			&res.Name,
			&res.Type,
			&res.Currency,
			&res.InitialBalance,
			&res.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to read a row: %w", err)
		}
		accounts = append(accounts, res)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("an error occurred while reading rows: %w", err)
	}

	return accounts, nil
}

// GetAnAccount finds an account by its ID and owner.
// It returns ErrAccountNotFound if the account is missing or belongs to another user.
func (r *Repository) GetAnAccount(ctx context.Context, userID, accID int64) (*Account, error) {
	query := `
		SELECT id, user_id, name, type, currency, initial_balance, created_at
		FROM accounts
		WHERE user_id=$1 AND id=$2;
	`

	var acc Account
	err := r.pool.QueryRow(ctx, query, userID, accID).Scan(
		&acc.ID,
		&acc.UserID,
		&acc.Name,
		&acc.Type,
		&acc.Currency,
		&acc.InitialBalance,
		&acc.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to execute sql query: %w", err)
	}

	return &acc, nil
}
