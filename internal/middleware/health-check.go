package middleware

import (
	"context"
	"time"

	"go-api-boilerplate/internal/database"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
)

// readinessTimeout is the maximum time allowed for the DB ping during a
// readiness check. If the database doesn't respond within this window the
// pod is marked unready so the load balancer stops sending it traffic.
const readinessTimeout = 2 * time.Second

// LivenessHandler returns a Fiber handler for the liveness probe.
// Register it on the desired path (conventionally /live or /livez):
//
//	app.Get("/live", middleware.LivenessHandler())
//
// In Fiber v3, healthcheck endpoints are registered as individual routes
// rather than as a single unified middleware — this gives each probe its own
// path without requiring the caller to pre-configure endpoint names.
func LivenessHandler() fiber.Handler {
	return healthcheck.New(healthcheck.Config{
		Probe: func(_ fiber.Ctx) bool {
			return true
		},
	})
}

// ReadinessHandler returns a Fiber handler for the readiness probe.
// Register it on the desired path (conventionally /ready or /readyz):
//
//	app.Get("/ready", middleware.ReadinessHandler(registry))
//
// The probe pings the default PostgreSQL connection with a [readinessTimeout]
// deadline. A 503 is returned when the database is unreachable.
func ReadinessHandler(registry *database.Registry) fiber.Handler {
	return healthcheck.New(healthcheck.Config{
		Probe: func(_ fiber.Ctx) bool {
			if registry == nil {
				return false
			}
			db := registry.Default()
			if db == nil {
				return false
			}
			ctx, cancel := context.WithTimeout(context.Background(), readinessTimeout)
			defer cancel()
			return db.Ping(ctx) == nil
		},
	})
}
