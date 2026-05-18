// Package repository provides the concrete PostgreSQL implementation of the
// UserRepository interface. All functions accept pre-hashed password values;
// hashing is the responsibility of the service (business) layer.
//
// The database connection is retrieved from the request context via
// database.DBFromContext — it must be set by MultiTenantMiddleware before any
// repository method is called.
package repository

import (
	"context"

	"go-api-boilerplate/internal/api/models"
	"go-api-boilerplate/internal/database"

	"github.com/boni-fm/go-libsd3/pkg/db/postgres"
)

// UserRepository is the production UserRepository backed by PostgreSQL.
// Create instances with NewPostgresUserRepository; never embed directly in tests.
type UserRepository struct {
	db *postgres.Database
}

// NewPostgresUserRepository returns a new PostgresUserRepository that satisfies
// the UserRepository interface.
func NewUserRepository(db *postgres.Database) *UserRepository {
	return &UserRepository{db: db}
}

// AddUser inserts a new user row. passwordHash must already be a bcrypt hash
// produced by the service layer — this function stores it verbatim.
func (r *UserRepository) AddUser(ctx context.Context, userName, passwordHash string) error {
	if r.db == nil {
		return database.ErrNoDB
	}
	query := `INSERT INTO dc_user_t (user_name, user_password, user_app_modul) VALUES ($1, $2, 'GOLANG')`
	_, err := r.db.Exec(ctx, query, userName, passwordHash)
	return err
}

// GetAllUsers retrieves every user record, returning only the username.
// Passwords (even hashed) are never returned to callers.
func (r *UserRepository) GetAllUsers(ctx context.Context) ([]models.UserResponse, error) {
	if r.db == nil {
		return nil, database.ErrNoDB
	}
	var users []models.UserResponse
	query := `SELECT user_name FROM dc_user_t`
	if err := r.db.SelectAll(ctx, &users, query); err != nil {
		return nil, err
	}
	return users, nil
}

// UpdateUserPassword replaces the stored password hash for the given user.
// passwordHash must already be a bcrypt hash produced by the service layer.
func (r *UserRepository) UpdateUserPassword(ctx context.Context, userName, passwordHash string) error {
	if r.db == nil {
		return database.ErrNoDB
	}
	query := `UPDATE dc_user_t SET user_password = $1 WHERE user_name = $2`
	_, err := r.db.Exec(ctx, query, passwordHash, userName)
	return err
}

// DeleteUser permanently removes the user row identified by userName.
func (r *UserRepository) DeleteUser(ctx context.Context, userName string) error {
	if r.db == nil {
		return database.ErrNoDB
	}
	query := `DELETE FROM dc_user_t WHERE user_name = $1`
	_, err := r.db.Exec(ctx, query, userName)
	return err
}
