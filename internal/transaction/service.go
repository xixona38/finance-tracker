package transaction

import (
	"context"
	"errors"
	"fmt"

	"github.com/xixona38/finance-tracker/internal/account"
)

type Service struct {
	repo    *Repository
	repoAcc *account.Repository
}

var (
	ErrInvalidData = errors.New("invalid data entry")
)

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

// List checks the page size, offset, and optional account ID, then loads the user's transactions.
// A nil account ID leaves the results unfiltered by account.
func (s *Service) List(ctx context.Context, userID int64, accID *int64, limit, offset int) ([]*Transaction, error) {
	if limit < 1 || limit > 100 {
		return nil, fmt.Errorf("limit must be between 1 and 100: %w", ErrInvalidData)
	}
	if offset < 0 {
		return nil, fmt.Errorf("offset must be zero or greater: %w", ErrInvalidData)
	}
	if accID != nil {
		if *accID < 1 {
			return nil, fmt.Errorf("account id must be greater than zero: %w", ErrInvalidData)
		}
	}

	return s.repo.List(ctx, userID, accID, limit, offset)
}
