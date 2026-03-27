// Package repository provides the concrete PostgreSQL implementation of the
// UserRepository interface. All functions accept pre-hashed password values;
// hashing is the responsibility of the service (business) layer.
package repository

import (
	"context"

	"go-api-boilerplate/internal/api/models"
	"go-api-boilerplate/internal/database"
)

// PostgresUserRepository is the production UserRepository backed by PostgreSQL.
// Create instances with NewPostgresUserRepository; never embed directly in tests.
type PostgresUserRepository struct{}

// NewPostgresUserRepository returns a new PostgresUserRepository that satisfies
// the UserRepository interface.
func NewPostgresUserRepository() UserRepository {
	return &PostgresUserRepository{}
}

// AddUser inserts a new user row. passwordHash must already be a bcrypt hash
// produced by the service layer — this function stores it verbatim.
func (r *PostgresUserRepository) AddUser(ctx context.Context, userName, passwordHash string) error {
	query := `INSERT INTO dc_user_t (user_name, user_password, user_app_modul) VALUES ($1, $2, 'GOLANG')`
	_, err := database.Db.Exec(ctx, query, userName, passwordHash)
	return err
}

// GetAllUsers retrieves every user record, returning only the username.
// Passwords (even hashed) are never returned to callers.
func (r *PostgresUserRepository) GetAllUsers(ctx context.Context) ([]models.UserResponse, error) {
	var users []models.UserResponse
	query := `SELECT user_name FROM dc_user_t`
	if err := database.Db.SelectAll(ctx, &users, query); err != nil {
		return nil, err
	}
	return users, nil
}

// UpdateUserPassword replaces the stored password hash for the given user.
// passwordHash must already be a bcrypt hash produced by the service layer.
func (r *PostgresUserRepository) UpdateUserPassword(ctx context.Context, userName, passwordHash string) error {
	query := `UPDATE dc_user_t SET user_password = $1 WHERE user_name = $2`
	_, err := database.Db.Exec(ctx, query, passwordHash, userName)
	return err
}

// DeleteUser permanently removes the user row identified by userName.
func (r *PostgresUserRepository) DeleteUser(ctx context.Context, userName string) error {
	query := `DELETE FROM dc_user_t WHERE user_name = $1`
	_, err := database.Db.Exec(ctx, query, userName)
	return err
}
