// internal/api/handlers/ProfileHandler.go
package handlers

import (
	"context"

	"go-api-boilerplate/internal/api/models"
	"go-api-boilerplate/internal/utility/fibererror"

	"github.com/gofiber/fiber/v2"
)

// GetProfile godoc
// @Summary      Get user profile
// @Description  Returns extended profile information for a user
// @Tags         profile
// @Produce      json
// @Param        user_name  path      string  true  "Username"
// @Success      200        {object}  models.ProfileResponse
// @Failure      404        {object}  fibererror.ResponseError
// @Failure      500        {object}  fibererror.ResponseError
// @Router       /api/users/{user_name}/profile [get]
func (hr *HandlersRegistry) GetProfile(c *fiber.Ctx) error {
	userName := c.Params("user_name")

	profile, err := hr.ProfileService.GetProfile(c.Context(), userName)
	if err != nil {
		hr.log_.Errorf("GetProfile error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fibererror.ResponseError{
			Code:    fiber.StatusInternalServerError,
			Error:   "Internal Server Error",
			Message: "Failed to fetch profile",
		})
	}

	// Dispatch a non-critical "profile viewed" audit event to the worker pool
	// so it is recorded asynchronously without blocking the response.
	if hr.Pool != nil {
		user := userName
		if ok := hr.Pool.Submit(func(_ context.Context) {
			hr.log_.Infof("[audit] profile viewed: %s", user)
		}); !ok {
			hr.log_.Warn("worker pool saturated — profile-view audit event dropped")
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    profile,
	})
}

// UpsertProfile godoc
// @Summary      Create or update user profile
// @Description  Creates a new profile or updates the display name and email for an existing user
// @Tags         profile
// @Accept       json
// @Produce      json
// @Param        user_name  path      string                       true  "Username"
// @Param        request    body      models.UpdateProfileRequest  true  "Profile data"
// @Success      200        {object}  map[string]interface{}
// @Failure      400        {object}  fibererror.ResponseError
// @Failure      500        {object}  fibererror.ResponseError
// @Router       /api/users/{user_name}/profile [put]
func (hr *HandlersRegistry) UpsertProfile(c *fiber.Ctx) error {
	userName := c.Params("user_name")

	var req models.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fibererror.ResponseError{
			Code:    fiber.StatusBadRequest,
			Error:   "Bad Request",
			Message: "Invalid request body",
		})
	}

	if req.DisplayName == "" && req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fibererror.ResponseError{
			Code:    fiber.StatusBadRequest,
			Error:   "Bad Request",
			Message: "display_name or email is required",
		})
	}

	if err := hr.ProfileService.UpsertProfile(c.Context(), userName, req.DisplayName, req.Email); err != nil {
		hr.log_.Errorf("UpsertProfile error: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fibererror.ResponseError{
			Code:    fiber.StatusInternalServerError,
			Error:   "Internal Server Error",
			Message: "Failed to upsert profile",
		})
	}

	hr.log_.Infof("Profile upserted for: %s", userName)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Profile updated",
	})
}
