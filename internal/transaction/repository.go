package transaction

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
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
// An account filter keeps transactions where that account sends or receives money.
// A type filter keeps only expenses, income, or transfers, as requested.
// It returns an empty slice when no transactions match.
func (r *Repository) List(
	ctx context.Context,
	userID int64,
	filters ListFilter,
) ([]*Transaction, error) {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	qb := psql.Select(
		"id",
		"user_id",
		"from_acc_id",
		"to_acc_id",
		"type",
		"amount",
		"description",
		"occurred_at",
		"created_at",
	).From("transactions").Where(squirrel.Eq{"user_id": userID})

	if filters.AccountID != nil {
		qb = qb.Where(squirrel.Or{
			squirrel.Eq{"from_acc_id": *filters.AccountID},
			squirrel.Eq{"to_acc_id": *filters.AccountID},
		})
	}

	if filters.Type != "" {
		qb = qb.Where(squirrel.Eq{"type": filters.Type})
	}

	qb = qb.OrderBy("occurred_at DESC", "id DESC").Limit(uint64(filters.Limit)).Offset(uint64(filters.Offset))

	query, args, err := qb.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build sql query: %w", err)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute sql query: %w", err)
	}
	defer rows.Close()

	transactions := make([]*Transaction, 0, filters.Limit)

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

// Delete removes a transaction only if it belongs to the given user.
// It returns ErrTransactionNotFound if the transaction is missing or belongs to someone else.
func (r *Repository) Delete(ctx context.Context, trID, userID int64) error {
	query := `
		DELETE FROM transactions
		WHERE id = $1 AND user_id = $2;
	`
	pg, err := r.pool.Exec(ctx, query, trID, userID)
	if err != nil {
		return fmt.Errorf("failed to execute sql query: %w", err)
	}
	if pg.RowsAffected() == 0 {
		return ErrTransactionNotFound
	}
	return nil
}
