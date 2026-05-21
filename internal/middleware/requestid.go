package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/gofiber/utils/v2"
)

// LocalsRequestID disimpen buat kompatibilitas. Di Fiber v3, request ID diambil
// pakai requestid.FromContext(c). Kalau butuh di Locals (misal buat format logger),
// copy manual aja: c.Locals(LocalsRequestID, requestid.FromContext(c))
const LocalsRequestID = "requestid"

// RequestIDMiddleware generate UUIDv4 sebagai request ID dan taruh di header X-Request-ID.
// Kalau request masuk udah bawa header X-Request-ID (misal dari API gateway),
// value yang ada dipakai ulang — berguna buat distributed tracing.
func RequestIDMiddleware() fiber.Handler {
	return requestid.New(requestid.Config{
		// UUIDv4 lebih gampang dibaca dan di-grep di log dibanding SecureToken (default v3)
		Generator: utils.UUIDv4,
		Header:    fiber.HeaderXRequestID,
	})
}
