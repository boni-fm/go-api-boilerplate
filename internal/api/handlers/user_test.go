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
	"time"

	"go-api-boilerplate/internal/api/handlers"
	"go-api-boilerplate/internal/api/models"

	"github.com/boni-fm/go-libsd3/pkg/log"
	"github.com/gofiber/fiber/v3"
)

// ----- mock service -----

type mockUserSvc struct {
	createErr error
	getUsers  []models.UserResponse
	getErr    error
	updateErr error
	deleteErr error
}

func (m *mockUserSvc) CreateUser(_ context.Context, _, _ string) error {
	return m.createErr
}
func (m *mockUserSvc) GetUsers(_ context.Context) ([]models.UserResponse, error) {
	return m.getUsers, m.getErr
}
func (m *mockUserSvc) UpdateUserPassword(_ context.Context, _, _ string) error {
	return m.updateErr
}
func (m *mockUserSvc) DeleteUser(_ context.Context, _ string) error {
	return m.deleteErr
}

// compile-time check
var _ handlers.UserServiceIface = (*mockUserSvc)(nil)

// ----- helpers -----

func newTestRouter(t *testing.T, svc handlers.UserServiceIface) *fiber.App {
	t.Helper()
	l := log.NewLoggerWithFilename("test")
	hr := handlers.NewHandlersRegistryForTest(l, svc)
	app := fiber.New()
	app.Post("/api/users", hr.CreateUser)
	app.Get("/api/users", hr.GetUsers)
	app.Put("/api/users/:user_name/password", hr.UpdateUserPassword)
	app.Delete("/api/users/:user_name", hr.DeleteUser)
	return app
}

func jsonBody(t *testing.T, v interface{}) io.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}
	return bytes.NewReader(b)
}

// ----- CreateUser -----

func TestCreateUser_Success(t *testing.T) {
	svc := &mockUserSvc{}
	app := newTestRouter(t, svc)

	body := jsonBody(t, map[string]string{"user_name": "alice", "password": "secret"})
	req := httptest.NewRequest(http.MethodPost, "/api/users", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusCreated)
	}
}

func TestCreateUser_MissingFields(t *testing.T) {
	tests := []struct {
		name string
		body map[string]string
	}{
		{"missing user_name", map[string]string{"password": "secret"}},
		{"missing password", map[string]string{"user_name": "alice"}},
		{"both missing", map[string]string{}},
	}

	svc := &mockUserSvc{}
	app := newTestRouter(t, svc)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/users", jsonBody(t, tc.body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusBadRequest)
			}
		})
	}
}

func TestCreateUser_InvalidJSON(t *testing.T) {
	svc := &mockUserSvc{}
	app := newTestRouter(t, svc)

	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString("{not-json"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestCreateUser_ServiceError(t *testing.T) {
	svc := &mockUserSvc{createErr: errors.New("db error")}
	app := newTestRouter(t, svc)

	body := jsonBody(t, map[string]string{"user_name": "alice", "password": "secret"})
	req := httptest.NewRequest(http.MethodPost, "/api/users", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

// ----- GetUsers -----

func TestGetUsers_Success(t *testing.T) {
	users := []models.UserResponse{{UserName: "alice"}, {UserName: "bob"}}
	svc := &mockUserSvc{getUsers: users}
	app := newTestRouter(t, svc)

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if result["success"] != true {
		t.Errorf("success: got %v, want true", result["success"])
	}
	data, ok := result["data"].([]interface{})
	if !ok || len(data) != 2 {
		t.Errorf("data: expected 2 users, got %v", result["data"])
	}
}

func TestGetUsers_ServiceError(t *testing.T) {
	svc := &mockUserSvc{getErr: errors.New("db error")}
	app := newTestRouter(t, svc)

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

// ----- UpdateUserPassword -----

func TestUpdateUserPassword_Success(t *testing.T) {
	svc := &mockUserSvc{}
	app := newTestRouter(t, svc)

	body := jsonBody(t, map[string]string{"new_password": "newpass"})
	req := httptest.NewRequest(http.MethodPut, "/api/users/alice/password", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestUpdateUserPassword_MissingNewPassword(t *testing.T) {
	svc := &mockUserSvc{}
	app := newTestRouter(t, svc)

	body := jsonBody(t, map[string]string{})
	req := httptest.NewRequest(http.MethodPut, "/api/users/alice/password", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestUpdateUserPassword_ServiceError(t *testing.T) {
	svc := &mockUserSvc{updateErr: errors.New("db error")}
	app := newTestRouter(t, svc)

	body := jsonBody(t, map[string]string{"new_password": "newpass"})
	req := httptest.NewRequest(http.MethodPut, "/api/users/alice/password", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

// ----- DeleteUser -----

func TestDeleteUser_Success(t *testing.T) {
	svc := &mockUserSvc{}
	app := newTestRouter(t, svc)

	req := httptest.NewRequest(http.MethodDelete, "/api/users/alice", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestDeleteUser_ServiceError(t *testing.T) {
	svc := &mockUserSvc{deleteErr: errors.New("db error")}
	app := newTestRouter(t, svc)

	req := httptest.NewRequest(http.MethodDelete, "/api/users/alice", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

// ----- ErrorResponseShape -----

func TestErrorResponses_UseStandardShape(t *testing.T) {
	// Verify that all error responses include the standard {code, error, message} shape.
	svc := &mockUserSvc{createErr: errors.New("some error")}
	app := newTestRouter(t, svc)

	body := jsonBody(t, map[string]string{"user_name": "x", "password": "y"})
	req := httptest.NewRequest(http.MethodPost, "/api/users", body)
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
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
