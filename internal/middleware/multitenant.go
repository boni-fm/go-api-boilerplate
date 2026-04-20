// Package middleware contains all Fiber middleware registrations.
package middleware

import (
	"strings"

	"go-api-boilerplate/internal/database"

	"github.com/gofiber/fiber/v3"
)

// MultiTenantMiddleware resolves the database connection for each request
// and stores it in the request context via database.WithDB. Downstream
// repository calls retrieve the connection with database.DBFromContext.
//
// Tenant key resolution priority (highest → lowest):
//
//  1. Query parameter  ?kunci=<key>       — explicit per-request override
//  2. Request header   X-Kunci: <key>     — service-to-service / SDK usage
//  3. Nginx header     X-Forwarded-Prefix — first path segment used as key
//     (e.g. "X-Forwarded-Prefix: /g009sim" → key "g009sim")
//  4. Default connection (first key in appsettings.ini)
//
// If the registry is nil (e.g. in tests that bypass server setup), the
// middleware is a no-op and passes the request through unchanged.
func MultiTenantMiddleware(registry *database.Registry) fiber.Handler {
	return func(c fiber.Ctx) error {
		if registry == nil {
			return c.Next()
		}

		db := registry.Default()

		if kunci := resolveKunci(c); kunci != "" {
			if d, ok := registry.Get(kunci); ok {
				db = d
			}
		}

		c.SetContext(database.WithDB(c.Context(), db))
		return c.Next()
	}
}

// resolveKunci determines the tenant key from the request using the
// documented priority order.  Returns an empty string when no key is found.
func resolveKunci(c fiber.Ctx) string {
	// 1. Query parameter: ?kunci=<key>
	if k := c.Query("kunci"); k != "" {
		return k
	}

	// 2. Request header: X-Kunci: <key>
	if k := c.Get("X-Kunci"); k != "" {
		return k
	}

	// 3. Nginx reverse-proxy header: X-Forwarded-Prefix: /<key>[/optional/path]
	//    Extract the first non-empty path segment (strip leading "/").
	if prefix := c.Get("X-Forwarded-Prefix"); prefix != "" {
		// Trim surrounding slashes then split; take the first segment only.
		segment := strings.SplitN(strings.Trim(prefix, "/"), "/", 2)[0]
		if segment != "" {
			return segment
		}
	}

	return ""
}
