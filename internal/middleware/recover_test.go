package middleware_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-api-boilerplate/internal/middleware"
	"go-api-boilerplate/internal/utility/fibererror"

	"github.com/boni-fm/go-libsd3/pkg/log"
	"github.com/gofiber/fiber/v3"
)

// newRecoverApp builds a minimal Fiber app with the RecoverMiddleware and
// GlobalErrorHandler registered, mirroring the production configuration.
func newRecoverApp(t *testing.T) *fiber.App {
	t.Helper()
	l := log.NewLoggerWithFilename("test")
	app := fiber.New(fiber.Config{
		ErrorHandler: fibererror.GlobalErrorHandler,
	})
	app.Use(middleware.RecoverMiddleware(l))
	return app
}

// TestRecoverMiddleware_PanicReturns500 verifies that a panic inside a handler
// does not crash the server but instead returns HTTP 500.
func TestRecoverMiddleware_PanicReturns500(t *testing.T) {
	app := newRecoverApp(t)
	app.Get("/panic", func(c fiber.Ctx) error {
		panic("something went terribly wrong")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

// TestRecoverMiddleware_PanicResponseUsesStandardShape verifies that the
// response body produced after a panic uses the standard ResponseError JSON
// envelope (code, error, message) rather than a custom ad-hoc shape.
func TestRecoverMiddleware_PanicResponseUsesStandardShape(t *testing.T) {
	app := newRecoverApp(t)
	app.Get("/panic", func(c fiber.Ctx) error {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}

	rawBody, _ := io.ReadAll(resp.Body)
	var body map[string]interface{}
	if err := json.Unmarshal(rawBody, &body); err != nil {
		t.Fatalf("response is not valid JSON: %v — body: %s", err, rawBody)
	}

	for _, key := range []string{"code", "error", "message"} {
		if body[key] == nil {
			t.Errorf("expected key %q in panic response body; got: %s", key, rawBody)
		}
	}
}

// TestRecoverMiddleware_NoPanicContinuesNormally verifies that the middleware
// does not interfere with ordinary (non-panicking) handler execution.
func TestRecoverMiddleware_NoPanicContinuesNormally(t *testing.T) {
	app := newRecoverApp(t)
	app.Get("/ok", func(c fiber.Ctx) error {
		return c.Status(http.StatusOK).SendString("all good")
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("app.Test returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
}
