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
	// BadRequestError must use ResponseError — verify all three fields are present.
	for _, key := range []string{"code", "error", "message"} {
		if body[key] == nil {
			t.Errorf("expected key %q in bad-request response body", key)
		}
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
	// NotFoundError calls c.Status(404).SendFile("./static/public/404.html").
	// When the static file does not exist (as in the test environment), fasthttp
	// sends a bare HTTP 404 response, so we assert the status code directly.
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusNotFound)
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
	// GatewayTimeoutError must use ResponseError — verify all three fields are present.
	for _, key := range []string{"code", "error", "message"} {
		if body[key] == nil {
			t.Errorf("expected key %q in gateway-timeout response body", key)
		}
	}
	if body["error"] != "Gateway Timeout" {
		t.Errorf("error field: got %v, want \"Gateway Timeout\"", body["error"])
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

// TestGlobalErrorHandler_UnhandledFiberError_PreservesStatusCode is the regression
// test for the critical bug where any Fiber error code not explicitly listed in the
// switch (e.g. 401, 403, 405, 408, 413, 429) was previously misrouted to
// InternalServerError, returning HTTP 500 instead of the correct status.
func TestGlobalErrorHandler_UnhandledFiberError_PreservesStatusCode(t *testing.T) {
	tests := []struct {
		name       string
		code       int
		wantError  string
		wantStatus int
	}{
		{"unauthorized", fiber.StatusUnauthorized, "Unauthorized", http.StatusUnauthorized},
		{"forbidden", fiber.StatusForbidden, "Forbidden", http.StatusForbidden},
		{"method not allowed", fiber.StatusMethodNotAllowed, "Method Not Allowed", http.StatusMethodNotAllowed},
		{"conflict", fiber.StatusConflict, "Conflict", http.StatusConflict},
		{"too many requests", fiber.StatusTooManyRequests, "Too Many Requests", http.StatusTooManyRequests},
		{"unprocessable entity", fiber.StatusUnprocessableEntity, "Unprocessable Entity", http.StatusUnprocessableEntity},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := newTestApp()
			code := tc.code // capture for closure
			app.Get("/test", func(c *fiber.Ctx) error {
				return fiber.NewError(code, "detail")
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}

			if resp.StatusCode != tc.wantStatus {
				t.Errorf("HTTP status: got %d, want %d", resp.StatusCode, tc.wantStatus)
			}

			body := decodeBody(t, resp.Body)
			if body["code"] != float64(tc.wantStatus) {
				t.Errorf("code field: got %v, want %d", body["code"], tc.wantStatus)
			}
			if body["error"] != tc.wantError {
				t.Errorf("error field: got %v, want %q", body["error"], tc.wantError)
			}
		})
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
