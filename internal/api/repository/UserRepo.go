package repository

import (
	"context"
	"go-api-boilerplate/internal/api/models"
	"go-api-boilerplate/internal/database"
)

// CRUD
// CREATE
func AddUser(ctx context.Context, user_name, user_password string) error {
	query := `INSERT INTO dc_user_t (user_name, user_password, user_app_modul) VALUES ($1, $2, 'GOLANG')`
	_, err := database.Db.Exec(ctx, query, user_name, user_password)
	if err != nil {
		return err
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
	_, err := database.Db.Exec(ctx, query, new_password, user_name)
	if err != nil {
		return err
	}
	return nil
}

// DELETE
func DeleteUser(ctx context.Context, user_name string) error {
	query := `DELETE FROM dc_user_t WHERE user_name = $1`
	_, err := database.Db.Exec(ctx, query, user_name)
	if err != nil {
		return err
	}
	return nil
}
