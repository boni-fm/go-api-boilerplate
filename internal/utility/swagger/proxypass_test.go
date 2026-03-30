package swagger_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go-api-boilerplate/internal/utility/swagger"

	"github.com/gofiber/fiber/v3"
)

func TestProxyPathMiddleware_WithHeader(t *testing.T) {
	app := fiber.New()
	app.Use(swagger.ProxyPathMiddleware())
	app.Get("/test", func(c fiber.Ctx) error {
		path := swagger.GetProxyPath(c)
		return c.SendString(path)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Forwarded-Prefix", "/my-service")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	buf := make([]byte, 64)
	n, _ := resp.Body.Read(buf)
	if string(buf[:n]) != "/my-service" {
		t.Errorf("proxy path: got %q, want %q", string(buf[:n]), "/my-service")
	}
}

func TestProxyPathMiddleware_WithoutHeader(t *testing.T) {
	app := fiber.New()
	app.Use(swagger.ProxyPathMiddleware())
	app.Get("/test", func(c fiber.Ctx) error {
		path := swagger.GetProxyPath(c)
		return c.SendString(path)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	buf := make([]byte, 64)
	n, _ := resp.Body.Read(buf)
	if string(buf[:n]) != "" {
		t.Errorf("proxy path: got %q, want empty string", string(buf[:n]))
	}
}

func TestGetProxyPath_NilLocals(t *testing.T) {
	app := fiber.New()
	// No middleware — Locals will be nil/missing.
	app.Get("/test", func(c fiber.Ctx) error {
		path := swagger.GetProxyPath(c)
		return c.SendString(path)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	buf := make([]byte, 64)
	n, _ := resp.Body.Read(buf)
	if string(buf[:n]) != "" {
		t.Errorf("proxy path: got %q, want empty string", string(buf[:n]))
	}
}
