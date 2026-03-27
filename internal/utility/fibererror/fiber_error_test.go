package fibererror_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-api-boilerplate/internal/utility/fibererror"

	"github.com/gofiber/fiber/v2"
)

// newTestApp returns a minimal Fiber app that uses GlobalErrorHandler.
func newTestApp() *fiber.App {
	return fiber.New(fiber.Config{
		ErrorHandler: fibererror.GlobalErrorHandler,
	})
}

// decodeBody unmarshals the response body into a map for assertion.
func decodeBody(t *testing.T, body io.Reader) map[string]interface{} {
	t.Helper()
	var m map[string]interface{}
	if err := json.NewDecoder(body).Decode(&m); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	return m
}

func TestGlobalErrorHandler_FiberBadRequest(t *testing.T) {
	app := newTestApp()
	app.Get("/bad", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusBadRequest, "bad input")
	})

	req := httptest.NewRequest(http.MethodGet, "/bad", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
	body := decodeBody(t, resp.Body)
	if body["error"] != "Bad Request" {
		t.Errorf("error field: got %v, want \"Bad Request\"", body["error"])
	}
}

func TestGlobalErrorHandler_FiberNotFound(t *testing.T) {
	app := newTestApp()
	app.Get("/nf", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusNotFound, "not found")
	})

	req := httptest.NewRequest(http.MethodGet, "/nf", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	// NotFoundError serves a static file which may not exist in test env,
	// so we only verify the handler is called (no 200 returned).
	if resp.StatusCode == http.StatusOK {
		t.Error("expected non-200 status for NotFound handler")
	}
}

func TestGlobalErrorHandler_FiberGatewayTimeout(t *testing.T) {
	app := newTestApp()
	app.Get("/timeout", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusGatewayTimeout, "upstream timeout")
	})

	req := httptest.NewRequest(http.MethodGet, "/timeout", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusGatewayTimeout {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusGatewayTimeout)
	}
	body := decodeBody(t, resp.Body)
	if body["error"] == nil {
		t.Error("expected error field in response body")
	}
}

func TestGlobalErrorHandler_GenericError(t *testing.T) {
	app := newTestApp()
	app.Get("/panic", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusInternalServerError, "something broke")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}
	body := decodeBody(t, resp.Body)
	if body["code"] != float64(500) {
		t.Errorf("code field: got %v, want 500", body["code"])
	}
}

func TestResponseError_JSONShape(t *testing.T) {
	re := fibererror.ResponseError{
		Code:    400,
		Error:   "Bad Request",
		Message: "missing field",
	}
	b, err := json.Marshal(re)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	checks := map[string]interface{}{
		"code":    float64(400),
		"error":   "Bad Request",
		"message": "missing field",
	}
	for k, want := range checks {
		if m[k] != want {
			t.Errorf("key %q: got %v, want %v", k, m[k], want)
		}
	}
}
