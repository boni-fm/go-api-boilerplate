package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
)

func LivenessHandler() fiber.Handler {
	return healthcheck.New(healthcheck.Config{
		Probe: func(_ fiber.Ctx) bool {
			return true
		},
	})
}
