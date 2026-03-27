// Package handlers defines the HTTP handler layer for the application.
// This file declares the interfaces that handlers depend on, following
// Go's "accept interfaces, return structs" idiom. Declaring them here
// (at the point of use) avoids circular imports and makes the dependency
// contract explicit for every engineer reading the handler code.
package handlers

import (
	"context"

	"go-api-boilerplate/internal/api/models"
)

// UserServiceIface defines the user-domain operations required by the
// handler layer. The concrete *services.UserService satisfies this
// interface; a mock implementation can be injected in unit tests.
type UserServiceIface interface {
	// CreateUser hashes the password and persists a new user record.
	CreateUser(ctx context.Context, userName, password string) error

	// GetUsers returns all users without exposing password hashes.
	GetUsers(ctx context.Context) ([]models.UserResponse, error)

	// UpdateUserPassword hashes the new password and stores it.
	UpdateUserPassword(ctx context.Context, userName, newPassword string) error

	// DeleteUser permanently removes the user by username.
	DeleteUser(ctx context.Context, userName string) error
}
