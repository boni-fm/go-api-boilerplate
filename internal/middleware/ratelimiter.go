package middleware

import (
	"time"

	"go-api-boilerplate/internal/utility/fibererror"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/sirupsen/logrus"
)

// RateLimiter batasi tiap IP maksimal 100 request per menit.
// Kalau limit kelewat, balik HTTP 429 pakai ResponseError biar formatnya konsisten.
// Path /live di-skip biar probe Kubernetes gak kena throttle.
func RateLimiter(log *logrus.Logger) fiber.Handler {
	return limiter.New(limiter.Config{
		// skip /live biar health probe gak kena rate limit
		Next: func(c fiber.Ctx) bool {
			p := c.Path()
			return p == "/live"
		},
		Max:        100,
		Expiration: 1 * time.Minute,
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fibererror.ResponseError{
				Code:    fiber.StatusTooManyRequests,
				Error:   "Too Many Requests",
				Message: "Rate limit exceeded. Please try again later.",
			})
		},
	})
}
