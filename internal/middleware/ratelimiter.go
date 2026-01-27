package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/sirupsen/logrus"
)

func RateLimiter(log *logrus.Logger) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{
				"code":    429,
				"error":   "Too Many Request",
				"message": "Oops ~ Request mencapai limit...",
			})
		},
	},
	)
}
