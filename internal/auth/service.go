package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/alexedwards/argon2id"
	"github.com/xixona38/finance-tracker/internal/session"
	"github.com/xixona38/finance-tracker/internal/user"
)

var (
	ErrValidation         = errors.New("validation failed")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type Service struct {
	repo    *user.Repository
	session *session.Service
}

// NewService creates an authentication service with a user repository and session service.
func NewService(repo *user.Repository, session *session.Service) *Service {
	return &Service{
		repo:    repo,
		session: session,
	}
}

// Register validates the email and password, hashes the password, and stores a new user.
// It returns the created user or a validation, duplicate email, or internal error.
func (s *Service) Register(ctx context.Context, email, password string) (*user.User, error) {
	addr, err := mail.ParseAddress(strings.TrimSpace(email))
	if err != nil {
		return nil, fmt.Errorf("%w: failed to parse email", ErrValidation)
	}

	if addr.Address != strings.TrimSpace(email) {
		return nil, fmt.Errorf("%w: invalid email", ErrValidation)
	}

	if utf8.RuneCountInString(password) < 15 || utf8.RuneCountInString(password) > 128 {
		return nil, fmt.Errorf("%w: password must be between 15 and 128 characters", ErrValidation)
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

// Authenticate validates credentials and verifies the password against the stored hash.
// It returns the matching user or ErrInvalidCredentials for an unknown email or incorrect password.
func (s *Service) Authenticate(ctx context.Context, email, password string) (*user.User, error) {
	addr, err := mail.ParseAddress(strings.TrimSpace(email))
	if err != nil {
		return nil, fmt.Errorf("%w: invalid email address", ErrValidation)
	}

	if addr.Address != strings.TrimSpace(email) {
		return nil, fmt.Errorf("%w: invalid email", ErrValidation)
	}

	if utf8.RuneCountInString(password) > 128 || utf8.RuneCountInString(password) < 15 {
		return nil, fmt.Errorf("%w: password must be between 15 and 128 characters", ErrValidation)
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

// Login authenticates the user and creates a session.
// It returns the user, the session token, and the session expiration time.
func (s *Service) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	user, err := s.Authenticate(ctx, email, password)
	if err != nil {
		return nil, err
	}

	token, session, err := s.session.Create(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &LoginResult{
		User:      *user,
		Token:     token,
		ExpiresAt: session.ExpiresAt,
	}, nil
}
