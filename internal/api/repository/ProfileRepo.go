package repository

import (
	"context"
	"fmt"

	"go-api-boilerplate/internal/api/models"
	"go-api-boilerplate/internal/database"
)

// ProfileRepository defines the persistence contract for user profiles.
type ProfileRepository interface {
	// GetProfile returns the profile for the given user.
	GetProfile(ctx context.Context, userName string) (*models.ProfileResponse, error)

	// UpsertProfile creates or updates the profile for the given user.
	UpsertProfile(ctx context.Context, userName, displayName, email string) error
}

// PostgresProfileRepository is the production ProfileRepository backed by PostgreSQL.
// Create instances with NewPostgresProfileRepository.
type PostgresProfileRepository struct{}

// NewPostgresProfileRepository returns a new PostgresProfileRepository that satisfies
// the ProfileRepository interface.
func NewPostgresProfileRepository() ProfileRepository {
	return &PostgresProfileRepository{}
}

// GetProfile retrieves the profile row for the given user.
func (r *PostgresProfileRepository) GetProfile(
	ctx context.Context, userName string,
) (*models.ProfileResponse, error) {
	var p models.ProfileResponse
	q := `SELECT user_name, display_name, email
	      FROM dc_user_profile_t WHERE user_name = $1`
	if err := database.Db.SelectOne(ctx, &p, q, userName); err != nil {
		return nil, fmt.Errorf("GetProfile %q: %w", userName, err)
	}
	return &p, nil
}

// UpsertProfile inserts a new profile row or updates an existing one for the given user.
func (r *PostgresProfileRepository) UpsertProfile(
	ctx context.Context, userName, displayName, email string,
) error {
	q := `INSERT INTO dc_user_profile_t (user_name, display_name, email)
	      VALUES ($1, $2, $3)
	      ON CONFLICT (user_name) DO UPDATE
	      SET display_name = $2, email = $3`
	if _, err := database.Db.Exec(ctx, q, userName, displayName, email); err != nil {
		return fmt.Errorf("UpsertProfile %q: %w", userName, err)
	}
	return nil
}
