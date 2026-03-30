package middleware

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
)

// defaultRequestTimeout is the maximum time a request handler is allowed to
// run before the context is cancelled. Downstream code (database queries,
// HTTP calls to other services) should honour ctx.Done() and return early
// when the deadline is exceeded.
//
// Individual routes that need a longer deadline (e.g. report generation)
// can override this per-handler by creating their own derived context.
const defaultRequestTimeout = 30 * time.Second

// TimeoutMiddleware wraps every request's context with a timeout-bounded
// child context. This ensures that database queries and other I/O operations
// that accept a context.Context will be cancelled automatically if the handler
// runs for too long.
//
// In Fiber v3, c.Context() returns the standard context.Context set via
// c.SetContext(). Downstream code (services, repositories) should accept
// and respect this context so deadlines propagate correctly.
func TimeoutMiddleware(timeout time.Duration) fiber.Handler {
	if timeout <= 0 {
		timeout = defaultRequestTimeout
	}
	return func(c fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), timeout)
		defer cancel()

		c.SetContext(ctx)
		return c.Next()
	}
}
