package account

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

var ErrValidation = errors.New("validation failed")

type Service struct {
	repo *Repository
}

// NewService creates an account service with the supplied repository.
func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// Create trims the account name, validates its length, type, and RUB currency,
// and saves the account. Validation failures are wrapped with ErrValidation.
func (s *Service) Create(ctx context.Context, account Account) (*Account, error) {
	nameWithoutSpaces := strings.TrimSpace(account.Name)
	nameLen := utf8.RuneCountInString(nameWithoutSpaces)
	if nameLen > 100 || nameLen < 1 {
		return nil, fmt.Errorf("%w: name length must be between 1 and 100 symbols", ErrValidation)
	}

	account.Name = nameWithoutSpaces

	if account.Type != "card" && account.Type != "cash" && account.Type != "credit" && account.Type != "savings" {
		return nil, fmt.Errorf("%w: type field can be card/cash/credit/savings only", ErrValidation)
	}

	if account.Currency != "RUB" {
		return nil, fmt.Errorf("%w: available currency is RUB only", ErrValidation)
	}

	return s.repo.Create(ctx, account)
}

// List retrieves accounts for the specified user and wraps repository errors.
// An empty account list is a successful result.
func (s *Service) List(ctx context.Context, userID int64) ([]Account, error) {
	accounts, err := s.repo.List(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get list of accounts: %w", err)
	}

	return accounts, nil
}
