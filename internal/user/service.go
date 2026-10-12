package user

import (
	"context"
)

type Service struct {
	repo *Repository
}

// NewService sets up user lookups using the given repository.
func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// GetByID checks that the ID is positive, then loads the user's public details.
// It passes back ErrUserNotFound if the user does not exist.
func (s *Service) GetByID(ctx context.Context, userID int64) (*User, error) {
	if userID <= 0 {
		return nil, ErrInvalidData
	}
	return s.repo.GetByID(ctx, userID)
}
