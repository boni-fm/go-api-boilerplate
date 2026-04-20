// Package middleware contains all Fiber middleware registrations.
package middleware

import (
	"go-api-boilerplate/internal/database"

	"github.com/gofiber/fiber/v3"
)

// MultiTenantMiddleware resolves the database connection for each request
// based on the X-Kunci header and stores it in the request context via
// database.WithDB. Downstream repository calls retrieve the connection with
// database.DBFromContext.
//
// Routing logic:
//   - If the X-Kunci header is present and matches a registered key, that
//     connection is used.
//   - If the header is absent or the key is not registered, the default
//     connection (first key in appsettings.ini) is used as a fallback.
//
// If the registry is nil (e.g. in tests that bypass server setup), the
// middleware is a no-op and passes the request through unchanged.
func MultiTenantMiddleware(registry *database.Registry) fiber.Handler {
	return func(c fiber.Ctx) error {
		if registry == nil {
			return c.Next()
		}

		var db = registry.Default()

		if kunci := c.Get("X-Kunci"); kunci != "" {
			if d, ok := registry.Get(kunci); ok {
				db = d
			}
		}

		c.SetContext(database.WithDB(c.Context(), db))
		return c.Next()
	}
}
