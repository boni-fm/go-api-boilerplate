package handlers

import (
	"go-api-boilerplate/internal/api/models"
	"go-api-boilerplate/internal/api/services"
	"time"

	"github.com/gofiber/fiber/v3"
)

// Ping godoc
// @Summary Ping Check
// @Description Ping Pong endpoint to check service availability
// @Tags _
// @Produce json
// @Success 200 {object} models.PingPongResponse "Pong response"
// @Router /ping [get]
func (hr *HandlersRegistry) PingPongHandler(c fiber.Ctx) error {
	service := services.NewPingService()
	message := service.GetPing()

	response := models.PingPongResponse{
		IsSuccess: true,
		Message:   message,
		Timestamp: time.Now(),
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
