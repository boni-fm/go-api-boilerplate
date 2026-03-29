package middleware

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
)

// defaultRequestTimeout is the maximum time a request handler is allowed to
// run before the context is cancelled. Downstream code (database queries,
// HTTP calls to other services) should honour ctx.Done() and return early
// when the deadline is exceeded.
//
// Individual routes that need a longer deadline (e.g. report generation)
// can override this per-handler by creating their own derived context.
const defaultRequestTimeout = 30 * time.Second

// TimeoutMiddleware wraps every request's UserContext with a timeout-bounded
// child context. This ensures that database queries and other I/O operations
// that accept a context.Context will be cancelled automatically if the handler
// runs for too long.
//
// Fiber reuses contexts across requests (fasthttp pool), so we must not store
// the derived context anywhere — we only set it for the current request cycle
// via c.SetUserContext.
func TimeoutMiddleware(timeout time.Duration) fiber.Handler {
	if timeout <= 0 {
		timeout = defaultRequestTimeout
	}
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), timeout)
		defer cancel()

		c.SetUserContext(ctx)
		return c.Next()
	}
}
