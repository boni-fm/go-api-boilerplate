package user_repo

import (
	"context"
	"go-api-boilerplate/internal/api/models"
	"go-api-boilerplate/internal/database"
)

// CRUD
// CREATE
func AddUser(ctx context.Context, user_name, user_password string) error {
	query := `INSERT INTO dc_user_t (user_name, user_password) VALUES ($1, $2)`
	err := database.Db.Query(ctx, query, user_name, user_password)
	if err != nil {
		return err.Scan()
	}
	return nil
}

// READ
func GetAllUsers(ctx context.Context) ([]models.User, error) {
	var users []models.User
	query := `SELECT user_name, user_password FROM dc_user_t`
	err := database.Db.SelectAll(ctx, &users, query)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// UPDATE
func UpdateUserPassword(ctx context.Context, user_name, new_password string) error {
	query := `UPDATE dc_user_t SET user_password = $1 WHERE user_name = $2`
	err := database.Db.Query(ctx, query, new_password, user_name)
	if err != nil {
		return err.Scan()
	}
	return nil
}

// DELETE
func DeleteUser(ctx context.Context, user_name string) error {
	query := `DELETE FROM dc_user_t WHERE user_name = $1`
	err := database.Db.Query(ctx, query, user_name)
	if err != nil {
		return err.Scan()
	}
	return nil
}
