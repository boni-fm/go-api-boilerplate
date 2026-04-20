package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-api-boilerplate/internal/database"
	"go-api-boilerplate/internal/middleware"

	"github.com/gofiber/fiber/v3"
)

// newTestApp wires MultiTenantMiddleware on a minimal Fiber app.
func newTestApp(registry *database.Registry) *fiber.App {
	app := fiber.New()
	app.Use(middleware.MultiTenantMiddleware(registry))
	app.Get("/test", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	return app
}

// TestMultiTenantMiddleware_NilRegistry ensures the middleware is a no-op
// when the registry is nil (as in tests that bypass full server setup).
func TestMultiTenantMiddleware_NilRegistry(t *testing.T) {
	app := newTestApp(nil)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

// TestMultiTenantMiddleware_EmptyRegistry exercises the default-fallback path.
func TestMultiTenantMiddleware_EmptyRegistry(t *testing.T) {
	r := database.NewRegistry()
	app := newTestApp(r)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

// TestResolveKunci_Priority verifies the query-param → X-Kunci → X-Forwarded-Prefix
// resolution order by reading the resolved value back via middleware.ResolvedKunci.
//
// ARC-005: this test now exercises the ACTUAL MultiTenantMiddleware and its
// internal resolveKunci function rather than duplicating the logic inline.
func TestResolveKunci_Priority(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		xKunci     string
		xFwdPrefix string
		wantKunci  string // "" means "no key sent — captured empty"
	}{
		{
			name:       "query param wins over all",
			url:        "/test?kunci=qparam",
			xKunci:     "header-key",
			xFwdPrefix: "/prefix-key",
			wantKunci:  "qparam",
		},
		{
			name:       "X-Kunci wins over prefix",
			url:        "/test",
			xKunci:     "header-key",
			xFwdPrefix: "/prefix-key",
			wantKunci:  "header-key",
		},
		{
			name:       "X-Forwarded-Prefix used when no higher source",
			url:        "/test",
			xFwdPrefix: "/g009sim",
			wantKunci:  "g009sim",
		},
		{
			name:       "X-Forwarded-Prefix with extra path segments",
			url:        "/test",
			xFwdPrefix: "/g009sim/extra/path",
			wantKunci:  "g009sim",
		},
		{
			name:       "X-Forwarded-Prefix without leading slash",
			url:        "/test",
			xFwdPrefix: "g009sim",
			wantKunci:  "g009sim",
		},
		{
			name:       "X-Forwarded-Prefix with only slash",
			url:        "/test",
			xFwdPrefix: "/",
			wantKunci:  "",
		},
		{
			name:      "no sources — empty kunci",
			url:       "/test",
			wantKunci: "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var captured string

			// Use a real registry so the middleware runs its full code path.
			registry := database.NewRegistry()
			app := fiber.New()
			app.Use(middleware.MultiTenantMiddleware(registry))
			app.Get("/test", func(c fiber.Ctx) error {
				captured = middleware.ResolvedKunci(c)
				return c.SendStatus(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			if tc.xKunci != "" {
				req.Header.Set("X-Kunci", tc.xKunci)
			}
			if tc.xFwdPrefix != "" {
				req.Header.Set("X-Forwarded-Prefix", tc.xFwdPrefix)
			}

			resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
			}
			if captured != tc.wantKunci {
				t.Errorf("kunci: got %q, want %q", captured, tc.wantKunci)
			}
		})
	}
}
