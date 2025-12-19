package router

import (
	"context"
	"go-api-boilerplate/internal/api/handlers"

	"github.com/boni-fm/go-libsd3/pkg/log"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(log *log.Logger, app *fiber.App) {
	// buat instance handlers registry
	ctx := context.Background()
	handlers := handlers.NewHandlersRegistry(log, ctx)

	//---
	// setup routing disini
	//---

	// > base routes
	app.Get("/ping", handlers.PingPongHandler)
	app.Get("/swagger/doc.json", handlers.GetSwaggerDocumentation)
	app.Get("/swagger", handlers.GetSwaggerUI)
	app.Get("/swagger/*", handlers.GetSwaggerUI)
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/swagger", fiber.StatusTemporaryRedirect)
	})

	// > user routes (example)
	// User routes
	app.Post("/api/users", handlers.CreateUser)
	app.Get("/api/users", handlers.GetUsers)
	app.Put("/api/users/:user_name/password", handlers.UpdateUserPassword)
	app.Delete("/api/users/:user_name", handlers.DeleteUser)
}
