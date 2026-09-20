package account

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

func (r *Repository) Create(ctx context.Context, account Account) (*Account, error) {
	query := `
		INSERT INTO accounts (name, type, currency, initial_balance)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at;
	`

	err := r.pool.QueryRow(ctx, query, account.Name, account.Type, account.Currency, account.InitialBalance).Scan(&account.ID, &account.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to execute sql query: %w", err)
	}

	return &account, nil
}
