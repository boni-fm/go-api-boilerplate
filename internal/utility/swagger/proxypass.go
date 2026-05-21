package swagger

import "github.com/gofiber/fiber/v3"

const (
	HeaderForwardedPrefix = "X-Forwarded-Prefix"
	LocalsProxyPath       = "proxyPath"
)

// ProxyPathMiddleware ambil proxy path dari header terus simpen ke locals
func ProxyPathMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		proxyPath := c.Get(HeaderForwardedPrefix, "")
		if proxyPath != "" {
			c.Locals(LocalsProxyPath, proxyPath)
		}
		return c.Next()
	}
}

// GetProxyPath ambil proxy path dari locals
func GetProxyPath(c fiber.Ctx) string {
	if proxyPath := c.Locals(LocalsProxyPath); proxyPath != nil {
		if path, ok := proxyPath.(string); ok {
			return path
		}
	}
	return ""
}
