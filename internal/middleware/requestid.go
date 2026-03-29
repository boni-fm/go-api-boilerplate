package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/fiber/v2/utils"
)

// LocalsRequestID is the key under which the request identifier is stored in
// Fiber's c.Locals. Handlers and downstream middleware can retrieve it with:
//
//	rid, _ := c.Locals(LocalsRequestID).(string)
const LocalsRequestID = "requestid"

// RequestIDMiddleware returns Fiber's built-in request-id middleware configured
// with UUIDv4 generation (privacy-safe — does not leak request count) and the
// standard X-Request-ID header.
//
// If the incoming request already carries an X-Request-ID header (e.g. from an
// API gateway or load balancer), the existing value is reused, enabling
// end-to-end distributed tracing without additional infrastructure.
func RequestIDMiddleware() fiber.Handler {
	return requestid.New(requestid.Config{
		// UUIDv4 is cryptographically random — unlike the default utils.UUID
		// which is sequential and leaks total request count.
		Generator:  utils.UUIDv4,
		Header:     fiber.HeaderXRequestID,
		ContextKey: LocalsRequestID,
	})
}
