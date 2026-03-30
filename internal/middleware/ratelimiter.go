package middleware

import (
	"time"

	"go-api-boilerplate/internal/utility/fibererror"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/sirupsen/logrus"
)

// RateLimiter returns a Fiber middleware that limits each IP to 100 requests
// per minute. When the limit is exceeded it responds with HTTP 429 using the
// standard ResponseError envelope so clients receive a consistent error shape.
//
// Health-check paths (/live, /ready) are excluded so that Kubernetes probes
// are never inadvertently rate-limited, regardless of probe frequency.
func RateLimiter(log *logrus.Logger) fiber.Handler {
	return limiter.New(limiter.Config{
		// Skip rate limiting for health probes so Kubernetes probes are never
		// inadvertently throttled. In Fiber v3, health checks are registered
		// as routes rather than middleware, so they pass through the full
		// middleware stack — this guard ensures they remain unaffected.
		Next: func(c fiber.Ctx) bool {
			p := c.Path()
			return p == "/live" || p == "/ready"
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
