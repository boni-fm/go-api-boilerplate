// Package repository defines the data-access layer for the application.
// It exposes interfaces that higher-level packages (services) depend on,
// decoupling business logic from the concrete database implementation and
// enabling straightforward unit testing via mock implementations.
package repository

import (
	"context"

	"go-api-boilerplate/internal/api/models"
)

// UserRepository defines the contract for all user persistence operations.
// Consumers (e.g. UserService) depend on this interface rather than the
// concrete PostgreSQL implementation, which enables clean unit testing.
type UserRepository interface {
	// AddUser inserts a new user record. passwordHash must be a valid bcrypt hash.
	AddUser(ctx context.Context, userName, passwordHash string) error

	// GetAllUsers returns all users without exposing their password hashes.
	GetAllUsers(ctx context.Context) ([]models.UserResponse, error)

	// UpdateUserPassword replaces the stored password hash for a given user.
	UpdateUserPassword(ctx context.Context, userName, passwordHash string) error

	// DeleteUser permanently removes a user record by username.
	DeleteUser(ctx context.Context, userName string) error
}
