package repository

import (
	"context"

	"go-api-boilerplate/internal/api/models"
	"go-api-boilerplate/internal/database"

	"github.com/boni-fm/go-libsd3/pkg/db/postgres"
)

type UserRepository struct {
	db *postgres.Database
}

func NewUserRepository(db *postgres.Database) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) AddUser(ctx context.Context, userName, passwordHash string) error {
	if r.db == nil {
		return database.ErrNoDB
	}
	query := `INSERT INTO dc_user_t (user_name, user_password, user_app_modul) VALUES ($1, $2, 'GOLANG')`
	_, err := r.db.Exec(ctx, query, userName, passwordHash)
	return err
}

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

func (r *UserRepository) UpdateUserPassword(ctx context.Context, userName, passwordHash string) error {
	if r.db == nil {
		return database.ErrNoDB
	}
	query := `UPDATE dc_user_t SET user_password = $1 WHERE user_name = $2`
	_, err := r.db.Exec(ctx, query, passwordHash, userName)
	return err
}

func (r *UserRepository) DeleteUser(ctx context.Context, userName string) error {
	if r.db == nil {
		return database.ErrNoDB
	}
	query := `DELETE FROM dc_user_t WHERE user_name = $1`
	_, err := r.db.Exec(ctx, query, userName)
	return err
}
