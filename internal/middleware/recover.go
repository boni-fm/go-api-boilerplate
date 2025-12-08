package middleware

import (
	"github.com/gofiber/fiber/v2"
)

func RecoverMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			{
				if r := recover(); r != nil {
					c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"code":    fiber.StatusInternalServerError,
						"error":   "Internal Server Error",
						"message": "PANIC! Server Crash! - Recovered from panic...",
					})
				}
			}
		}()
		return c.Next()
	}
}
