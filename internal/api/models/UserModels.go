package models

type User struct {
	UserName string `db:"user_name"`
	Password string `db:"user_password"`
}

type CreateUserRequest struct {
	UserName string `json:"user_name" example:"john_doe" description:"Username for the new user"`
	Password string `json:"password" example:"password123" description:"Password for the new user"`
}

type UpdateUserPasswordRequest struct {
	NewPassword string `json:"new_password" example:"newpassword123" description:"New password for the user"`
}

type UserResponse struct {
	UserName string `json:"user_name" db:"user_name" example:"john_doe"`
}

type UsersListResponse struct {
	Success bool           `json:"success" example:"true"`
	Data    []UserResponse `json:"data"`
	Message string         `json:"message" example:"Users retrieved successfully"`
}
