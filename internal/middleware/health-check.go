package middleware

import (
	"context"
	"time"

	"go-api-boilerplate/internal/database"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
)

// readinessTimeout is the maximum time allowed for the DB ping during a
// readiness check. If the database doesn't respond within this window the
// pod is marked unready so the load balancer stops sending it traffic.
const readinessTimeout = 2 * time.Second

// HealthCheckMiddleware returns Fiber's healthcheck middleware configured with
// both a liveness and a readiness probe:
//
//   - GET /live   → 200 OK when the process is running (liveness).
//   - GET /ready  → 200 OK only when PostgreSQL responds to a Ping within
//     [readinessTimeout]; 503 Service Unavailable otherwise.
//
// In Kubernetes, wire these as:
//
//	livenessProbe:  httpGet: { path: /live,  port: 8080 }
//	readinessProbe: httpGet: { path: /ready, port: 8080 }
func HealthCheckMiddleware() fiber.Handler {
	return healthcheck.New(healthcheck.Config{
		LivenessProbe: func(_ *fiber.Ctx) bool {
			return true
		},
		LivenessEndpoint: "/live",

		ReadinessProbe:    readinessProbe,
		ReadinessEndpoint: "/ready",
	})
}

// readinessProbe checks that the database is reachable. It creates a
// short-lived context with a timeout so that a slow or partitioned DB
// does not block the probe indefinitely.
func readinessProbe(_ *fiber.Ctx) bool {
	db := database.GetDatabase()
	if db == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), readinessTimeout)
	defer cancel()

	return db.Ping(ctx) == nil
}
