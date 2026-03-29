package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-api-boilerplate/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

// ── RequestIDMiddleware tests ─────────────────────────────────────────────────

func TestRequestIDMiddleware_GeneratesUUID(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.RequestIDMiddleware())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}

	rid := resp.Header.Get("X-Request-ID")
	if rid == "" {
		t.Error("X-Request-ID response header is empty; middleware should generate one")
	}
	// UUIDv4 format: 8-4-4-4-12 hex digits = 36 chars
	if len(rid) != 36 {
		t.Errorf("X-Request-ID length: got %d, want 36 (UUID format)", len(rid))
	}
}

func TestRequestIDMiddleware_PropagatesIncomingHeader(t *testing.T) {
	const incoming = "ext-trace-abc-123"

	app := fiber.New()
	app.Use(middleware.RequestIDMiddleware())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", incoming)
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}

	rid := resp.Header.Get("X-Request-ID")
	if rid != incoming {
		t.Errorf("X-Request-ID: got %q, want %q (should propagate incoming value)", rid, incoming)
	}
}

func TestRequestIDMiddleware_SetsLocals(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.RequestIDMiddleware())
	app.Get("/", func(c *fiber.Ctx) error {
		rid, ok := c.Locals(middleware.LocalsRequestID).(string)
		if !ok || rid == "" {
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		return c.SendString(rid)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	if len(body) == 0 {
		t.Error("body is empty; handler should have returned the request ID from locals")
	}
}

func TestRequestIDMiddleware_UniquePerRequest(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.RequestIDMiddleware())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	ids := make(map[string]bool)
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		resp, err := app.Test(req, 5000)
		if err != nil {
			t.Fatal(err)
		}
		rid := resp.Header.Get("X-Request-ID")
		if ids[rid] {
			t.Fatalf("duplicate X-Request-ID %q on request %d", rid, i)
		}
		ids[rid] = true
	}
}
