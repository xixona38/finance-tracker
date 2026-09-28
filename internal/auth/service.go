package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/alexedwards/argon2id"
	"github.com/xixona38/finance-tracker/internal/user"
)

var (
	ErrValidation         = errors.New("validation failed")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type Service struct {
	repo *user.Repository
}

func NewService(repo *user.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Register(ctx context.Context, email, password string) (*user.User, error) {
	addr, err := mail.ParseAddress(strings.TrimSpace(email))
	if err != nil {
		return nil, fmt.Errorf("%w: failed to parse email", ErrValidation)
	}

	if addr.Address != strings.TrimSpace(email) {
		return nil, fmt.Errorf("%w: invalid email", ErrValidation)
	}

	if utf8.RuneCountInString(password) < 15 || utf8.RuneCountInString(password) > 128 {
		return nil, fmt.Errorf("%w: password must not exceed 128 characters", ErrValidation)
	}

	passHash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return nil, fmt.Errorf("failed to hash the password: %w", err)
	}

	createdUser, err := s.repo.Create(ctx, user.User{
		Email:        addr.Address,
		PasswordHash: passHash,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return createdUser, nil
}

func (s *Service) Authenticate(ctx context.Context, email, password string) (*user.User, error) {
	addr, err := mail.ParseAddress(strings.TrimSpace(email))
	if err != nil {
		return nil, fmt.Errorf("%w: invalid email address", ErrValidation)
	}

	if addr.Address != strings.TrimSpace(email) {
		return nil, fmt.Errorf("%w: invalid email", ErrValidation)
	}

	if utf8.RuneCountInString(password) > 128 {
		return nil, fmt.Errorf("%w: password must be between 128 characters", ErrValidation)
	}

	foundUser, err := s.repo.GetByEmail(ctx, addr.Address)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	same, err := argon2id.ComparePasswordAndHash(password, foundUser.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to verify password: %w", err)
	}

	if same {
		return foundUser, nil
	}

	return nil, ErrInvalidCredentials
}
