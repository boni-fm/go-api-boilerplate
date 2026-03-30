package services

import (
	"context"
	"fmt"

	"go-api-boilerplate/internal/api/models"
	"go-api-boilerplate/internal/api/repository"

	"github.com/boni-fm/go-libsd3/pkg/log"
)

// ProfileService owns all business rules for user profile operations.
// It depends on a ProfileRepository interface so that tests can inject
// a mock without a real database.
type ProfileService struct {
	log_ *log.Logger
	repo repository.ProfileRepository
}

// NewProfileService constructs a ProfileService with the given logger and repository.
// Inject a mock repository in tests to avoid any database dependency.
func NewProfileService(log_ *log.Logger, repo repository.ProfileRepository) *ProfileService {
	return &ProfileService{log_: log_, repo: repo}
}

// GetProfile returns the profile for the given user.
func (s *ProfileService) GetProfile(
	ctx context.Context, userName string,
) (*models.ProfileResponse, error) {
	profile, err := s.repo.GetProfile(ctx, userName)
	if err != nil {
		return nil, fmt.Errorf("ProfileService.GetProfile: %w", err)
	}
	return profile, nil
}

// UpsertProfile creates or updates the profile for the given user.
// Returns an error if userName is empty or the repository operation fails.
func (s *ProfileService) UpsertProfile(
	ctx context.Context, userName, displayName, email string,
) error {
	if userName == "" {
		return fmt.Errorf("ProfileService.UpsertProfile: userName is required")
	}
	return s.repo.UpsertProfile(ctx, userName, displayName, email)
}
