package middleware

import (
	"go-api-boilerplate/internal/utility/fibererror"

	"github.com/boni-fm/go-libsd3/pkg/log"
	"github.com/gofiber/fiber/v2"
)

type MiddlewareDepedencies struct {
	Log           *log.Logger
	App           *fiber.App
	IsDevelopment bool
}

func NewMiddlewareDepedencies(log *log.Logger, app *fiber.App, isDevelopment bool) *MiddlewareDepedencies {
	return &MiddlewareDepedencies{
		Log:           log,
		App:           app,
		IsDevelopment: isDevelopment,
	}
}

func (md *MiddlewareDepedencies) InitAllMiddleware() {
	md.App.Use(
		LoggerMiddleware(md.Log.Logger),
		RecoverMiddleware(),
		HealthCheckMiddleware(),
	)
}

func (md *MiddlewareDepedencies) InitErrorMiddleware() {
	md.App.Use(
		fibererror.NotFoundError,
		fibererror.GateawayTimeoutError,
		fibererror.BadRequestError,
		fibererror.InternalServerError,
	)
}
