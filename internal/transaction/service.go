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

func NewService(repo *Repository, repoAcc *account.Repository) *Service {
	return &Service{
		repo:    repo,
		repoAcc: repoAcc,
	}
}

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
