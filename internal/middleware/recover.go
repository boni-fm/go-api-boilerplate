// Package middleware contains all Fiber middleware registrations.
package middleware

import (
	"runtime/debug"

	"github.com/boni-fm/go-libsd3/pkg/log"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

// RecoverMiddleware returns Fiber's battle-tested built-in recovery middleware.
// It catches any panic raised inside a downstream handler or middleware,
// logs the stack trace via the application logger, and then delegates to the
// application's GlobalErrorHandler so the response uses the standard
// ResponseError envelope instead of a custom ad-hoc shape.
//
// Why replace the custom defer-based implementation?
//   - The old custom recover wrote the error response inside a defer and silently
//     discarded write errors ("_ = c.Status(…).JSON(…)").
//   - It bypassed GlobalErrorHandler, producing a fiber.Map instead of ResponseError.
//   - It had no stack-trace capture, making panic root-cause analysis impossible.
func RecoverMiddleware(log *log.Logger) fiber.Handler {
	return recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c fiber.Ctx, e interface{}) {
			log.Errorf("Recovered from panic: %v\n%s", e, debug.Stack())
		},
	})
}
