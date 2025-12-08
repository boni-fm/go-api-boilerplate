package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/favicon"
)

func FaviconMiddleware() fiber.Handler {
	return favicon.New(
		favicon.Config{
			File: "./static/public/favicon.ico",
			URL:  "/domar.ico",
		},
	)
}
