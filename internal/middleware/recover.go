package middleware

import (
	"runtime/debug"

	"github.com/boni-fm/go-libsd3/pkg/log"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

// RecoverMiddleware nangkep panic dari handler atau middleware di bawahnya,
// log stack trace-nya, terus delegasiin ke GlobalErrorHandler
// biar response-nya tetap pakai format ResponseError yang konsisten.
func RecoverMiddleware(log *log.Logger) fiber.Handler {
	return recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c fiber.Ctx, e interface{}) {
			log.Errorf("Recovered from panic: %v\n%s", e, debug.Stack())
		},
	})
}
