package router

import (
	"go-api-boilerplate/config"
	"go-api-boilerplate/internal/api/handlers"
	"go-api-boilerplate/internal/database"
	"go-api-boilerplate/internal/middleware"
	"go-api-boilerplate/internal/utility/injector"
	"go-api-boilerplate/internal/utility/swagger"
	"go-api-boilerplate/internal/worker"
	"strings"

	"github.com/boni-fm/go-libsd3/pkg/log"
	"github.com/gofiber/fiber/v3"
)

// SetupRoutes registers all application routes on the provided Fiber app.
// All handlers are constructed via NewHandlersRegistry which wires their
// dependencies (repository, service, swagger utilities, worker pool) automatically.
//
// Swagger UI routes are only registered when isDevelopment is true, keeping the
// API documentation endpoint disabled in production builds.
func SetupRoutes(log *log.Logger, app *fiber.App, pool *worker.Pool, dcAdapter *database.DcAdapter, cfg *config.Config, dbInject injector.DBInjector) {
	handlersRegistry := handlers.NewHandlersRegistry(log, pool, dcAdapter, cfg, dbInject)

	//---
	// setup routing disini
	//---

	// > health probe routes
	app.Get("/live", middleware.LivenessHandler())

	// > swagger routes — only available in development / staging environments
	if cfg.IsDevelopment {
		app.Use("/swagger*", swagger.ProxyPathMiddleware())
		app.Get("/", func(c fiber.Ctx) error {
			proxyPath := c.Get("X-Forwarded-Prefix", "")
			if proxyPath != "" {
				proxyPath = strings.TrimSuffix(proxyPath, "/")
				return c.Redirect().Status(fiber.StatusTemporaryRedirect).To(proxyPath + "/swagger/index.html")
			}
			return c.Redirect().Status(fiber.StatusTemporaryRedirect).To("/swagger/index.html")
		})

		app.Get("/swagger/doc.json", handlersRegistry.GetSwaggerDocumentation)

		// Fiber v3 does not match "/swagger" (no trailing slash) against the
		// "/swagger/" or "/swagger/*" routes. Add an explicit redirect so that
		// browsers reaching /swagger are sent to the Swagger UI index page.
		app.Get("/swagger", func(c fiber.Ctx) error {
			proxyPath := c.Get("X-Forwarded-Prefix", "")
			if proxyPath != "" {
				proxyPath = strings.TrimSuffix(proxyPath, "/")
				return c.Redirect().Status(fiber.StatusMovedPermanently).To(proxyPath + "/swagger/index.html")
			}
			return c.Redirect().Status(fiber.StatusMovedPermanently).To("/swagger/index.html")
		})
		app.Get("/swagger/", handlersRegistry.GetSwaggerUI)
		app.Get("/swagger/*", handlersRegistry.GetSwaggerUI)
	}

	app.Get("/ping", handlersRegistry.PingPongHandler)

	// > user routes (example)
	api := app.Group("/api", middleware.MultiDCMiddleware(log, cfg, dcAdapter))
	{
		api.Post("/users", handlersRegistry.CreateUser)
		api.Get("/users", handlersRegistry.GetUsers)
		api.Put("/users/:user_name/password", handlersRegistry.UpdateUserPassword)
		api.Delete("/users/:user_name", handlersRegistry.DeleteUser)
	}
}
