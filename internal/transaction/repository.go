package transaction

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

func (r *Repository) Create(ctx context.Context, tr *Transaction) (*Transaction, error) {
	query := `
		INSERT INTO transactions (
			user_id,
			from_acc_id, 
			to_acc_id,
			type,
			amount,
			description,
			occurred_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at;
	`
	err := r.pool.QueryRow(
		ctx,
		query,
		tr.UserID,
		tr.FromAccID,
		tr.ToAccID,
		tr.Type,
		tr.Amount,
		tr.Description,
		tr.OccurredAt,
	).Scan(&tr.ID, &tr.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to execute sql query: %w", err)
	}

	return tr, nil
}
