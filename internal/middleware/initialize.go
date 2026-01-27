package middleware

import (
	"github.com/boni-fm/go-libsd3/pkg/log"
	"github.com/gofiber/fiber/v2"
)

type MiddlewareDependencies struct {
	Log           *log.Logger
	App           *fiber.App
	IsDevelopment bool
}

func NewMiddlewareDependencies(log *log.Logger, app *fiber.App, isDevelopment bool) *MiddlewareDependencies {
	return &MiddlewareDependencies{
		Log:           log,
		App:           app,
		IsDevelopment: isDevelopment,
	}
}

func (md *MiddlewareDependencies) InitAllMiddleware() {
	md.App.Use(
		LoggerMiddleware(md.Log.Logger),
		RecoverMiddleware(md.Log),
		HealthCheckMiddleware(),
		FaviconMiddleware(),
		RateLimiter(md.Log.Logger),
	)
}
