package middleware

import (
	"github.com/boni-fm/go-libsd3/pkg/log"
	"github.com/gofiber/fiber/v2"
)

func RecoverMiddleware(log *log.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("Recovered from panic: %v", r)
				_ = c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code":    fiber.StatusInternalServerError,
					"error":   "Internal Server Error",
					"message": "PANIC! Server Crash! - Recovered from panic...",
				})
			}
		}()

		return c.Next()
	}
}
