package router

import (
	"go-api-boilerplate/internal/api/handlers"
	"go-api-boilerplate/internal/utility/swagger"
	"go-api-boilerplate/internal/worker"

	"github.com/boni-fm/go-libsd3/pkg/log"
	"github.com/gofiber/fiber/v2"
)

// SetupRoutes registers all application routes on the provided Fiber app.
// All handlers are constructed via NewHandlersRegistry which wires their
// dependencies (repository, service, swagger utilities, worker pool) automatically.
func SetupRoutes(log *log.Logger, app *fiber.App, pool *worker.Pool) {
	handlers := handlers.NewHandlersRegistry(log, pool)

	//---
	// setup routing disini
	//---

	// > base routes
	app.Use("/swagger*", swagger.ProxyPathMiddleware())
	app.Get("/", func(c *fiber.Ctx) error {
		proxyPath := c.Get("X-Forwarded-Prefix", "")
		if proxyPath != "" {
			return c.Redirect(proxyPath+"/swagger", fiber.StatusTemporaryRedirect)
		}
		return c.Redirect("/swagger", fiber.StatusTemporaryRedirect)
	})

	app.Get("/ping", handlers.PingPongHandler)
	app.Get("/swagger/doc.json", handlers.GetSwaggerDocumentation)
	app.Get("/swagger/", handlers.GetSwaggerUI)
	app.Get("/swagger/*", handlers.GetSwaggerUI)

	// > user routes (example)
	app.Post("/api/users", handlers.CreateUser)
	app.Get("/api/users", handlers.GetUsers)
	app.Put("/api/users/:user_name/password", handlers.UpdateUserPassword)
	app.Delete("/api/users/:user_name", handlers.DeleteUser)
}
