package handlers

import (
	"go-api-boilerplate/internal/api/models"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Ping godoc
// @Summary Ping Check
// @Description Ping Pong endpoint to check service availability
// @Tags _
// @Produce json
// @Success 200 {object} models.PingPongResponse "Pong response"
// @Router /ping [get]
func PingPongHandler(c *fiber.Ctx) error {
	response := models.PingPongResponse{
		Message:   "Pong",
		Timestamp: time.Now(),
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
