package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-api-boilerplate/internal/api/handlers"
	"go-api-boilerplate/internal/api/models"

	"github.com/boni-fm/go-libsd3/pkg/log"
	"github.com/gofiber/fiber/v2"
)

// ----- mock service -----

type mockProfileSvc struct {
	profile   *models.ProfileResponse
	getErr    error
	upsertErr error
}

func (m *mockProfileSvc) GetProfile(_ context.Context, _ string) (*models.ProfileResponse, error) {
	return m.profile, m.getErr
}

func (m *mockProfileSvc) UpsertProfile(_ context.Context, _, _, _ string) error {
	return m.upsertErr
}

// compile-time check
var _ handlers.ProfileServiceIface = (*mockProfileSvc)(nil)

// ----- helpers -----

func newProfileRouter(t *testing.T, svc handlers.ProfileServiceIface) *fiber.App {
	t.Helper()
	l := log.NewLoggerWithFilename("test")
	hr := handlers.NewHandlersRegistryForTest(l, nil)
	hr.ProfileService = svc
	app := fiber.New()
	app.Get("/api/users/:user_name/profile", hr.GetProfile)
	app.Put("/api/users/:user_name/profile", hr.UpsertProfile)
	return app
}

// ----- GetProfile -----

func TestGetProfile_Success(t *testing.T) {
	svc := &mockProfileSvc{
		profile: &models.ProfileResponse{
			UserName:    "alice",
			DisplayName: "Alice Smith",
			Email:       "alice@example.com",
		},
	}
	app := newProfileRouter(t, svc)

	req := httptest.NewRequest(http.MethodGet, "/api/users/alice/profile", nil)
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if result["success"] != true {
		t.Errorf("success: got %v, want true", result["success"])
	}
	if result["data"] == nil {
		t.Error("expected non-nil data in response")
	}
}

func TestGetProfile_ServiceError(t *testing.T) {
	svc := &mockProfileSvc{getErr: errors.New("db error")}
	app := newProfileRouter(t, svc)

	req := httptest.NewRequest(http.MethodGet, "/api/users/alice/profile", nil)
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestGetProfile_ErrorResponseShape(t *testing.T) {
	svc := &mockProfileSvc{getErr: errors.New("db error")}
	app := newProfileRouter(t, svc)

	req := httptest.NewRequest(http.MethodGet, "/api/users/alice/profile", nil)
	resp, _ := app.Test(req, 5000)
	rawBody, _ := io.ReadAll(resp.Body)

	var m map[string]interface{}
	if err := json.Unmarshal(rawBody, &m); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	for _, key := range []string{"code", "error", "message"} {
		if m[key] == nil {
			t.Errorf("expected key %q in error response, got none; body: %s", key, rawBody)
		}
	}
}

// ----- UpsertProfile -----

func TestUpsertProfile_Success(t *testing.T) {
	svc := &mockProfileSvc{}
	app := newProfileRouter(t, svc)

	body := bytes.NewBufferString(`{"display_name":"Alice Smith","email":"alice@example.com"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/users/alice/profile", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestUpsertProfile_MissingFields(t *testing.T) {
	svc := &mockProfileSvc{}
	app := newProfileRouter(t, svc)

	// Both display_name and email are empty — should be rejected.
	body := bytes.NewBufferString(`{}`)
	req := httptest.NewRequest(http.MethodPut, "/api/users/alice/profile", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestUpsertProfile_InvalidJSON(t *testing.T) {
	svc := &mockProfileSvc{}
	app := newProfileRouter(t, svc)

	req := httptest.NewRequest(http.MethodPut, "/api/users/alice/profile", bytes.NewBufferString("{not-json"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestUpsertProfile_ServiceError(t *testing.T) {
	svc := &mockProfileSvc{upsertErr: errors.New("db error")}
	app := newProfileRouter(t, svc)

	body := bytes.NewBufferString(`{"display_name":"Alice"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/users/alice/profile", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}
