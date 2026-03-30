package services_test

import (
	"context"
	"errors"
	"testing"

	"go-api-boilerplate/internal/api/models"
	"go-api-boilerplate/internal/api/repository"
	"go-api-boilerplate/internal/api/services"

	"github.com/boni-fm/go-libsd3/pkg/log"
)

// ----- mock repository -----

// mockProfileRepo is an in-memory ProfileRepository used to test
// ProfileService without any database dependency.
type mockProfileRepo struct {
	// canned return for GetProfile
	profile *models.ProfileResponse
	getErr  error

	// injected error for UpsertProfile
	upsertErr error

	// last values captured by UpsertProfile
	lastUserName    string
	lastDisplayName string
	lastEmail       string
}

func (m *mockProfileRepo) GetProfile(_ context.Context, _ string) (*models.ProfileResponse, error) {
	return m.profile, m.getErr
}

func (m *mockProfileRepo) UpsertProfile(_ context.Context, userName, displayName, email string) error {
	m.lastUserName = userName
	m.lastDisplayName = displayName
	m.lastEmail = email
	return m.upsertErr
}

// compile-time check that mockProfileRepo satisfies the interface.
var _ repository.ProfileRepository = (*mockProfileRepo)(nil)

// ----- helpers -----

func newProfileService(t *testing.T, repo repository.ProfileRepository) *services.ProfileService {
	t.Helper()
	l := log.NewLoggerWithFilename("test")
	return services.NewProfileService(l, repo)
}

// ----- GetProfile tests -----

func TestProfileService_GetProfile_ReturnsProfile(t *testing.T) {
	want := &models.ProfileResponse{
		UserName:    "alice",
		DisplayName: "Alice Smith",
		Email:       "alice@example.com",
	}
	repo := &mockProfileRepo{profile: want}
	svc := newProfileService(t, repo)

	got, err := svc.GetProfile(context.Background(), "alice")
	if err != nil {
		t.Fatalf("GetProfile returned unexpected error: %v", err)
	}
	if got.UserName != want.UserName {
		t.Errorf("UserName: got %q, want %q", got.UserName, want.UserName)
	}
	if got.DisplayName != want.DisplayName {
		t.Errorf("DisplayName: got %q, want %q", got.DisplayName, want.DisplayName)
	}
	if got.Email != want.Email {
		t.Errorf("Email: got %q, want %q", got.Email, want.Email)
	}
}

func TestProfileService_GetProfile_RepoError(t *testing.T) {
	repo := &mockProfileRepo{getErr: errors.New("db error")}
	svc := newProfileService(t, repo)

	_, err := svc.GetProfile(context.Background(), "alice")
	if err == nil {
		t.Error("expected error from repo, got nil")
	}
}

// ----- UpsertProfile tests -----

func TestProfileService_UpsertProfile_PassesFieldsToRepo(t *testing.T) {
	repo := &mockProfileRepo{}
	svc := newProfileService(t, repo)

	if err := svc.UpsertProfile(context.Background(), "alice", "Alice Smith", "alice@example.com"); err != nil {
		t.Fatalf("UpsertProfile returned unexpected error: %v", err)
	}

	if repo.lastUserName != "alice" {
		t.Errorf("userName: got %q, want %q", repo.lastUserName, "alice")
	}
	if repo.lastDisplayName != "Alice Smith" {
		t.Errorf("displayName: got %q, want %q", repo.lastDisplayName, "Alice Smith")
	}
	if repo.lastEmail != "alice@example.com" {
		t.Errorf("email: got %q, want %q", repo.lastEmail, "alice@example.com")
	}
}

func TestProfileService_UpsertProfile_EmptyUserName(t *testing.T) {
	repo := &mockProfileRepo{}
	svc := newProfileService(t, repo)

	err := svc.UpsertProfile(context.Background(), "", "Alice Smith", "alice@example.com")
	if err == nil {
		t.Error("expected error for empty userName, got nil")
	}
}

func TestProfileService_UpsertProfile_RepoError(t *testing.T) {
	repo := &mockProfileRepo{upsertErr: errors.New("db error")}
	svc := newProfileService(t, repo)

	err := svc.UpsertProfile(context.Background(), "alice", "Alice Smith", "alice@example.com")
	if err == nil {
		t.Error("expected error from repo, got nil")
	}
}
