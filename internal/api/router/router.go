package router

import (
	"go-api-boilerplate/internal/api/handlers"
	"go-api-boilerplate/internal/middleware"
	"go-api-boilerplate/internal/utility/swagger"
	"go-api-boilerplate/internal/worker"
	"strings"

	"github.com/boni-fm/go-libsd3/pkg/log"
	"github.com/gofiber/fiber/v3"
)

// SetupRoutes registers all application routes on the provided Fiber app.
// All handlers are constructed via NewHandlersRegistry which wires their
// dependencies (repository, service, swagger utilities, worker pool) automatically.
func SetupRoutes(log *log.Logger, app *fiber.App, pool *worker.Pool) {
	handlers := handlers.NewHandlersRegistry(log, pool)

	//---
	// setup routing disini
	//---

	// > health probe routes
	// In Fiber v3, healthcheck probes are registered as explicit routes instead
	// of a single unified middleware. This allows each probe to have its own path
	// and configuration. The RateLimiter middleware skips /live and /ready via its
	// Next function so probes are never inadvertently throttled.
	app.Get("/live", middleware.LivenessHandler())
	app.Get("/ready", middleware.ReadinessHandler())

	// > base routes
	app.Use("/swagger*", swagger.ProxyPathMiddleware())
	app.Get("/", func(c fiber.Ctx) error {
		proxyPath := c.Get("X-Forwarded-Prefix", "")
		if proxyPath != "" {
			proxyPath = strings.TrimSuffix(proxyPath, "/")
			return c.Redirect().Status(fiber.StatusTemporaryRedirect).To(proxyPath + "/swagger/index.html")
		}
		return c.Redirect().Status(fiber.StatusTemporaryRedirect).To("/swagger/index.html")
	})

	app.Get("/ping", handlers.PingPongHandler)
	app.Get("/swagger/doc.json", handlers.GetSwaggerDocumentation)
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
	app.Get("/swagger/", handlers.GetSwaggerUI)
	app.Get("/swagger/*", handlers.GetSwaggerUI)

	// > user routes (example)
	app.Post("/api/users", handlers.CreateUser)
	app.Get("/api/users", handlers.GetUsers)
	app.Put("/api/users/:user_name/password", handlers.UpdateUserPassword)
	app.Delete("/api/users/:user_name", handlers.DeleteUser)
}
