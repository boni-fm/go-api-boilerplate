// Package services contains the business-logic layer for the application.
// Each service owns its domain's rules; persistence is delegated to a
// repository interface so that the service can be tested in isolation.
package services

import (
	"context"
	"fmt"

	"go-api-boilerplate/internal/api/models"
	"go-api-boilerplate/internal/api/repository"

	"github.com/boni-fm/go-libsd3/pkg/log"
	"golang.org/x/crypto/bcrypt"
)

// UserService handles all user-related business operations.
// It depends on a UserRepository interface, allowing the concrete
// database implementation to be swapped out during testing.
type UserService struct {
	log_ *log.Logger
	repo repository.UserRepository
}

// NewUserService constructs a UserService with the given logger and repository.
// Inject a mock repository in tests to avoid any database dependency.
func NewUserService(log_ *log.Logger, repo repository.UserRepository) *UserService {
	return &UserService{
		log_: log_,
		repo: repo,
	}
}

// CreateUser hashes the plain-text password with bcrypt and persists the new user.
// Returns an error if hashing fails or the repository operation fails.
func (us *UserService) CreateUser(ctx context.Context, userName, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	return us.repo.AddUser(ctx, userName, string(hash))
}

// GetUsers returns all registered users without exposing password hashes.
func (us *UserService) GetUsers(ctx context.Context) ([]models.UserResponse, error) {
	return us.repo.GetAllUsers(ctx)
}

// UpdateUserPassword hashes the new plain-text password with bcrypt and
// updates the stored hash for the given user.
func (us *UserService) UpdateUserPassword(ctx context.Context, userName, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	return us.repo.UpdateUserPassword(ctx, userName, string(hash))
}

// DeleteUser permanently removes the user identified by userName.
func (us *UserService) DeleteUser(ctx context.Context, userName string) error {
	return us.repo.DeleteUser(ctx, userName)
}
