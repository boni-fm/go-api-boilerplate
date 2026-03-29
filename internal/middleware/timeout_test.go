package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"go-api-boilerplate/internal/middleware"

	"github.com/gofiber/fiber/v2"
)

// ── TimeoutMiddleware tests ───────────────────────────────────────────────────

func TestTimeoutMiddleware_SetsDeadlineOnContext(t *testing.T) {
	app := fiber.New()
	app.Use(middleware.TimeoutMiddleware(5 * time.Second))

	var hasDeadline atomic.Bool
	app.Get("/", func(c *fiber.Ctx) error {
		_, ok := c.UserContext().Deadline()
		hasDeadline.Store(ok)
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if !hasDeadline.Load() {
		t.Error("UserContext should have a deadline after TimeoutMiddleware")
	}
}

func TestTimeoutMiddleware_ContextCancelledAfterTimeout(t *testing.T) {
	const timeout = 50 * time.Millisecond

	app := fiber.New()
	app.Use(middleware.TimeoutMiddleware(timeout))

	var ctxErr atomic.Value
	app.Get("/", func(c *fiber.Ctx) error {
		ctx := c.UserContext()
		// Wait for the deadline to pass.
		select {
		case <-ctx.Done():
			ctxErr.Store(ctx.Err())
		case <-time.After(2 * time.Second):
			ctxErr.Store(context.DeadlineExceeded) // fallback — should not reach
		}
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	_, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}

	stored := ctxErr.Load()
	if stored == nil {
		t.Fatal("context error was never stored; timeout did not fire")
	}
	if stored != context.DeadlineExceeded {
		t.Errorf("context error: got %v, want %v", stored, context.DeadlineExceeded)
	}
}

func TestTimeoutMiddleware_NonPositiveDefaultsToSafe(t *testing.T) {
	// Passing 0 should not panic; the middleware should use a sane default.
	app := fiber.New()
	app.Use(middleware.TimeoutMiddleware(0))

	var hasDeadline atomic.Bool
	app.Get("/", func(c *fiber.Ctx) error {
		_, ok := c.UserContext().Deadline()
		hasDeadline.Store(ok)
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if !hasDeadline.Load() {
		t.Error("UserContext should have a deadline even with timeout=0 (should use default)")
	}
}
