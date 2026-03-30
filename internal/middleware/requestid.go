package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/gofiber/utils/v2"
)

// LocalsRequestID is retained for backward compatibility. In Fiber v3,
// the request ID is stored via the typed context key inside the requestid
// middleware and is retrieved with requestid.FromContext(c). If you need it
// in a string-keyed Locals slot (e.g. for the logger format), copy it there
// in a thin wrapper middleware:
//
//	c.Locals(LocalsRequestID, requestid.FromContext(c))
const LocalsRequestID = "requestid"

// RequestIDMiddleware returns Fiber's built-in request-id middleware configured
// with UUIDv4 generation and the standard X-Request-ID header.
//
// UUIDv4 uses a cryptographically secure random number generator internally
// (via crypto/rand) so the IDs are unguessable and do not leak request count.
//
// If the incoming request already carries an X-Request-ID header (e.g. from an
// API gateway or load balancer), the existing value is reused, enabling
// end-to-end distributed tracing without additional infrastructure.
//
// In Fiber v3, the request ID is stored using an unexported typed key. Retrieve
// it in handlers and downstream middleware with requestid.FromContext(c).
func RequestIDMiddleware() fiber.Handler {
	return requestid.New(requestid.Config{
		// UUIDv4 is cryptographically random — unlike utils.SecureToken (the
		// v3 default) it produces a human-readable UUID format that is easier
		// to grep in log files.
		Generator: utils.UUIDv4,
		Header:    fiber.HeaderXRequestID,
	})
}
