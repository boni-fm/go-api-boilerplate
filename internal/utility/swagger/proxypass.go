package swagger

import "github.com/gofiber/fiber/v3"

const (
	HeaderForwardedPrefix = "X-Forwarded-Prefix"
	LocalsProxyPath       = "proxyPath"
)

// ProxyPathMiddleware extracts and stores the proxy path from headers
func ProxyPathMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		proxyPath := c.Get(HeaderForwardedPrefix, "")
		if proxyPath != "" {
			c.Locals(LocalsProxyPath, proxyPath)
		}
		return c.Next()
	}
}

// GetProxyPath retrieves the proxy path from context locals
func GetProxyPath(c fiber.Ctx) string {
	if proxyPath := c.Locals(LocalsProxyPath); proxyPath != nil {
		if path, ok := proxyPath.(string); ok {
			return path
		}
	}
	return ""
}
