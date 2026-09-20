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

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

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
