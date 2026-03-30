package handlers_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-api-boilerplate/internal/api/handlers"
	"go-api-boilerplate/internal/api/models"

	"github.com/boni-fm/go-libsd3/pkg/log"
	"github.com/gofiber/fiber/v3"
)

func TestPingPongHandler_Returns200(t *testing.T) {
	app := fiber.New()
	l := log.NewLoggerWithFilename("test")
	hr := handlers.NewHandlersRegistryForTest(l, nil)
	app.Get("/ping", hr.PingPongHandler)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestPingPongHandler_ResponseShape(t *testing.T) {
	app := fiber.New()
	l := log.NewLoggerWithFilename("test")
	hr := handlers.NewHandlersRegistryForTest(l, nil)
	app.Get("/ping", hr.PingPongHandler)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}

	body, _ := io.ReadAll(resp.Body)
	var r models.PingPongResponse
	if err := json.Unmarshal(body, &r); err != nil {
		t.Fatalf("failed to unmarshal ping response: %v", err)
	}
	if !r.IsSuccess {
		t.Error("is_success: got false, want true")
	}
	if r.Message != "Pong" {
		t.Errorf("message: got %q, want %q", r.Message, "Pong")
	}
	if r.Timestamp.IsZero() {
		t.Error("timestamp must not be zero")
	}
	// Timestamp should be recent (within the last minute).
	if time.Since(r.Timestamp) > time.Minute {
		t.Errorf("timestamp %v is too old", r.Timestamp)
	}
}
