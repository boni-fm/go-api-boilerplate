package services

import (
	"context"
	"go-api-boilerplate/internal/api/models"
	user_repo "go-api-boilerplate/internal/api/repository"

	"github.com/boni-fm/go-libsd3/pkg/log"
)

type UserService struct {
	log_ *log.Logger
	ctx  context.Context
}

func NewUserService(log_ *log.Logger, ctx context.Context) *UserService {
	return &UserService{
		log_: log_,
		ctx:  ctx,
	}
}

// crud
func (us *UserService) CreateUser(user_name, user_password string) error {
	return user_repo.AddUser(us.ctx, user_name, user_password)
}

func (us *UserService) GetUsers() ([]models.User, error) {
	return user_repo.GetAllUsers(us.ctx)
}

func (us *UserService) UpdateUserPassword(user_name, new_password string) error {
	return user_repo.UpdateUserPassword(us.ctx, user_name, new_password)
}

func (us *UserService) DeleteUser(user_name string) error {
	return user_repo.DeleteUser(us.ctx, user_name)
}
