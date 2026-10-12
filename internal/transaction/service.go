package transaction

import (
	"context"
	"fmt"

	"github.com/xixona38/finance-tracker/internal/account"
)

type Service struct {
	repo    *Repository
	repoAcc *account.Repository
}

// NewService sets up transaction management using transaction storage and account lookups.
func NewService(repo *Repository, repoAcc *account.Repository) *Service {
	return &Service{
		repo:    repo,
		repoAcc: repoAcc,
	}
}

// Create checks the amount, date, and accounts for an expense, income, or transfer.
// It saves the transaction only if the details are valid and all referenced accounts belong to the user.
func (s *Service) Create(ctx context.Context, tr *Transaction) (*Transaction, error) {
	if tr.Amount < 1 {
		return nil, ErrInvalidData
	}

	if tr.OccurredAt.IsZero() {
		return nil, fmt.Errorf("transaction date required: %w", ErrInvalidData)
	}

	switch tr.Type {
	case "expense":
		if tr.FromAccID == nil || tr.ToAccID != nil {
			return nil, ErrInvalidData
		}
		_, err := s.repoAcc.GetAnAccount(ctx, tr.UserID, *tr.FromAccID)
		if err != nil {
			return nil, err
		}
	case "income":
		if tr.FromAccID != nil || tr.ToAccID == nil {
			return nil, ErrInvalidData
		}
		_, err := s.repoAcc.GetAnAccount(ctx, tr.UserID, *tr.ToAccID)
		if err != nil {
			return nil, err
		}
	case "transfer":
		if tr.FromAccID == nil || tr.ToAccID == nil || *tr.FromAccID == *tr.ToAccID {
			return nil, ErrInvalidData
		}
		_, err := s.repoAcc.GetAnAccount(ctx, tr.UserID, *tr.FromAccID)
		if err != nil {
			return nil, err
		}
		_, err = s.repoAcc.GetAnAccount(ctx, tr.UserID, *tr.ToAccID)
		if err != nil {
			return nil, err
		}
	default:
		return nil, ErrInvalidData
	}

	return s.repo.Create(ctx, tr)
}

// List checks the page size, offset, and optional account and type filters, then loads the user's transactions.
// A nil account ID or an empty type leaves that filter unused.
func (s *Service) List(ctx context.Context, userID int64, filters ListFilter) ([]*Transaction, error) {
	if filters.Limit < 1 || filters.Limit > 100 {
		return nil, fmt.Errorf("limit must be between 1 and 100: %w", ErrInvalidData)
	}
	if filters.Offset < 0 {
		return nil, fmt.Errorf("offset must be zero or greater: %w", ErrInvalidData)
	}
	if filters.AccountID != nil {
		if *filters.AccountID < 1 {
			return nil, fmt.Errorf("account id must be greater than zero: %w", ErrInvalidData)
		}
	}
	if filters.Type != "expense" && filters.Type != "income" && filters.Type != "transfer" && filters.Type != "" {
		return nil, fmt.Errorf("type must be 'income' or 'expense' or 'transfer' or no type at all: %w", ErrInvalidData)
	}

	return s.repo.List(ctx, userID, filters)
}

// Delete checks that both IDs are positive, then asks the repository to delete the user's transaction.
// It passes back ErrTransactionNotFound if no matching transaction belongs to the user.
func (s *Service) Delete(ctx context.Context, trID, userID int64) error {
	if trID <= 0 || userID <= 0 {
		return ErrInvalidData
	}
	return s.repo.Delete(ctx, trID, userID)
}
