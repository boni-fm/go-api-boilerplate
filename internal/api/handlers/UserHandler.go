// internal/api/handlers/UserHandler.go
package handlers

import (
	"go-api-boilerplate/internal/api/models"
	"go-api-boilerplate/internal/utility/fibererror"

	"github.com/gofiber/fiber/v2"
)

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user account with username and password
// @Tags users
// @Accept json
// @Produce json
// @Param request body models.CreateUserRequest true "User creation request"
// @Success 201 {object} map[string]interface{} "User created successfully"
// @Failure 400 {object} fibererror.ResponseError "Invalid request format or missing fields"
// @Failure 500 {object} fibererror.ResponseError "Internal server error"
// @Router /api/users [post]
func (hr *HandlersRegistry) CreateUser(c *fiber.Ctx) error {
	var req models.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fibererror.ResponseError{
			Code:    fiber.StatusBadRequest,
			Error:   "Bad Request",
			Message: "Invalid request body",
		})
	}

	if req.UserName == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fibererror.ResponseError{
			Code:    fiber.StatusBadRequest,
			Error:   "Bad Request",
			Message: "user_name and password are required",
		})
	}

	if err := hr.UserService.CreateUser(c.Context(), req.UserName, req.Password); err != nil {
		hr.log_.Errorf("Create user error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fibererror.ResponseError{
			Code:    fiber.StatusInternalServerError,
			Error:   "Internal Server Error",
			Message: "Failed to create user",
		})
	}

	hr.log_.Infof("User created: %s", req.UserName)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "User created",
		"user":    req.UserName,
	})
}

// GetUsers godoc
// @Summary Get all users
// @Description Retrieve a list of all registered users
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} models.UsersListResponse "Users retrieved successfully"
// @Failure 500 {object} fibererror.ResponseError "Failed to fetch users"
// @Router /api/users [get]
func (hr *HandlersRegistry) GetUsers(c *fiber.Ctx) error {
	users, err := hr.UserService.GetUsers(c.Context())
	if err != nil {
		hr.log_.Errorf("Get users error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fibererror.ResponseError{
			Code:    fiber.StatusInternalServerError,
			Error:   "Internal Server Error",
			Message: "Failed to fetch users",
		})
	}

	hr.log_.Infof("Fetched %d users", len(users))
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    users,
	})
}

// UpdateUserPassword godoc
// @Summary Update user password
// @Description Update the password for a specific user by username
// @Tags users
// @Accept json
// @Produce json
// @Param user_name path string true "Username of the user to update"
// @Param request body models.UpdateUserPasswordRequest true "New password"
// @Success 200 {object} map[string]interface{} "Password updated successfully"
// @Failure 400 {object} fibererror.ResponseError "Invalid request format or missing fields"
// @Failure 500 {object} fibererror.ResponseError "Failed to update password"
// @Router /api/users/{user_name}/password [put]
func (hr *HandlersRegistry) UpdateUserPassword(c *fiber.Ctx) error {
	userName := c.Params("user_name")

	var req models.UpdateUserPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fibererror.ResponseError{
			Code:    fiber.StatusBadRequest,
			Error:   "Bad Request",
			Message: "Invalid request body",
		})
	}

	if req.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fibererror.ResponseError{
			Code:    fiber.StatusBadRequest,
			Error:   "Bad Request",
			Message: "new_password is required",
		})
	}

	if err := hr.UserService.UpdateUserPassword(c.Context(), userName, req.NewPassword); err != nil {
		hr.log_.Errorf("Update password error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fibererror.ResponseError{
			Code:    fiber.StatusInternalServerError,
			Error:   "Internal Server Error",
			Message: "Failed to update password",
		})
	}

	hr.log_.Infof("Password updated for: %s", userName)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Password updated",
	})
}

// DeleteUser godoc
// @Summary Delete a user
// @Description Delete a user account by username
// @Tags users
// @Accept json
// @Produce json
// @Param user_name path string true "Username of the user to delete"
// @Success 200 {object} map[string]interface{} "User deleted successfully"
// @Failure 500 {object} fibererror.ResponseError "Failed to delete user"
// @Router /api/users/{user_name} [delete]
func (hr *HandlersRegistry) DeleteUser(c *fiber.Ctx) error {
	userName := c.Params("user_name")

	if err := hr.UserService.DeleteUser(c.Context(), userName); err != nil {
		hr.log_.Errorf("Delete user error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fibererror.ResponseError{
			Code:    fiber.StatusInternalServerError,
			Error:   "Internal Server Error",
			Message: "Failed to delete user",
		})
	}

	hr.log_.Infof("User deleted: %s", userName)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "User deleted",
	})
}
