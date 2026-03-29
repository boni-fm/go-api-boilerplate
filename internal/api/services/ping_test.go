package services_test

import (
	"testing"

	"go-api-boilerplate/internal/api/services"
)

func TestPingService_GetPing(t *testing.T) {
	svc := services.NewPingService()
	got := svc.GetPing()
	const want = "Pong"
	if got != want {
		t.Errorf("GetPing(): got %q, want %q", got, want)
	}
}

func TestPingService_GetPing_NotEmpty(t *testing.T) {
	svc := services.NewPingService()
	if svc.GetPing() == "" {
		t.Error("GetPing() must not return an empty string")
	}
}
