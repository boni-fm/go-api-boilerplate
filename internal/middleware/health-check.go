package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
)

func HealthCheckMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		healthcheck.New(healthcheck.Config{
			LivenessProbe: func(c *fiber.Ctx) bool {
				return true
			},
			LivenessEndpoint: "/live",
		},
		)
		return c.Next()
	}
}
