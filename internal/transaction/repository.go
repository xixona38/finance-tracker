package transaction

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository sets up transaction storage using the given database pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

// Create saves an expense, income, or transfer in the database.
// It fills in the transaction's ID and creation time and returns the same transaction.
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

// List returns a page of the user's transactions, ordered by transaction date from newest to oldest.
// Transactions with the same date are ordered by ID, from highest to lowest.
// If accID is set, it includes only transactions where that account sends or receives money.
// It returns an empty slice when no transactions match.
func (r *Repository) List(
	ctx context.Context,
	userID int64,
	accID *int64,
	limit,
	offset int,
) ([]*Transaction, error) {
	args := make([]any, 0, 4)
	args = append(args, userID, limit, offset)

	query := `
		SELECT
			id,
			user_id,
			from_acc_id,
			to_acc_id,
			type,
			amount,
			description,
			occurred_at,
			created_at
		FROM transactions
		WHERE user_id=$1
		ORDER BY occurred_at DESC, id DESC
		LIMIT $2 OFFSET $3;
	`
	if accID != nil {
		query = `
		SELECT
			id,
			user_id,
			from_acc_id,
			to_acc_id,
			type,
			amount,
			description,
			occurred_at,
			created_at
		FROM transactions
		WHERE user_id=$1 AND (from_acc_id=$4 OR to_acc_id=$4)
		ORDER BY occurred_at DESC, id DESC
		LIMIT $2 OFFSET $3;
	`
		args = append(args, *accID)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute sql query: %w", err)
	}
	defer rows.Close()

	transactions := make([]*Transaction, 0, limit)

	for rows.Next() {
		var tr Transaction
		if err := rows.Scan(
			&tr.ID,
			&tr.UserID,
			&tr.FromAccID,
			&tr.ToAccID,
			&tr.Type,
			&tr.Amount,
			&tr.Description,
			&tr.OccurredAt,
			&tr.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to read a row: %w", err)
		}
		transactions = append(transactions, &tr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("an error occurred while reading rows: %w", err)
	}

	return transactions, nil
}
