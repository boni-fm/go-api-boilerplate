package middleware

import (
	"github.com/boni-fm/go-libsd3/pkg/log"
	"github.com/gofiber/fiber/v3"
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

// InitAllMiddleware registers all global middleware in the correct order.
//
// Order matters:
//  1. RequestID  — must be first so every subsequent log line, error response
//     and downstream service call can include the correlation ID.
//  2. Logger     — logs the request after RequestID is set.
//  3. Recover    — catches panics and returns a standard error response.
//  4. Timeout    — wraps each request's context with a deadline so DB queries
//     and outbound I/O do not block indefinitely.
//  5. Favicon    — serves favicon without hitting the router.
//  6. RateLimiter — protects downstream handlers from excessive traffic.
//
// Note: In Fiber v3, healthcheck endpoints are registered as individual routes
// (GET /live and GET /ready) in the router rather than as a middleware. The
// RateLimiter skips these paths via its Next function to prevent probes from
// being throttled.
func (md *MiddlewareDependencies) InitAllMiddleware() {
	md.App.Use(
		RequestIDMiddleware(),
		LoggerMiddleware(md.Log.Logger),
		RecoverMiddleware(md.Log),
		TimeoutMiddleware(defaultRequestTimeout),
		FaviconMiddleware(),
		RateLimiter(md.Log.Logger),
	)
}
