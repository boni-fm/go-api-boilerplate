package router

import (
	"go-api-boilerplate/internal/api/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	// Setup Default Routes
	app.Get("/ping", handlers.PingPongHandler)
}
