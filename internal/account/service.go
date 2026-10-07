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

// NewService sets up the account service using the given repository.
func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// Create removes spaces around the name and checks the name length, account type, and currency.
// It saves the account if the details are valid; otherwise, it returns an ErrValidation error.
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

// List returns all accounts belonging to the given user.
// A user with no accounts gets an empty list, not an error.
func (s *Service) List(ctx context.Context, userID int64) ([]Account, error) {
	accounts, err := s.repo.List(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get list of accounts: %w", err)
	}

	return accounts, nil
}
