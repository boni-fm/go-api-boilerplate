package middleware

import (
	"time"

	"go-api-boilerplate/internal/utility/fibererror"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/sirupsen/logrus"
)

// RateLimiter returns a Fiber middleware that limits each IP to 100 requests
// per minute. When the limit is exceeded it responds with HTTP 429 using the
// standard ResponseError envelope so clients receive a consistent error shape.
func RateLimiter(log *logrus.Logger) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fibererror.ResponseError{
				Code:    fiber.StatusTooManyRequests,
				Error:   "Too Many Requests",
				Message: "Rate limit exceeded. Please try again later.",
			})
		},
	})
}
