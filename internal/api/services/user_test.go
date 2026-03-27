package services_test

import (
	"context"
	"errors"
	"testing"

	"go-api-boilerplate/internal/api/models"
	"go-api-boilerplate/internal/api/repository"
	"go-api-boilerplate/internal/api/services"

	"github.com/boni-fm/go-libsd3/pkg/log"
	"golang.org/x/crypto/bcrypt"
)

// ----- mock repository -----

// mockUserRepo is an in-memory repository implementation used to test
// UserService without any database dependency.
type mockUserRepo struct {
	// last values captured by each call
	addedUserName   string
	addedHash       string
	updatedUserName string
	updatedHash     string
	deletedUserName string

	// injected errors
	addErr    error
	getErr    error
	updateErr error
	deleteErr error

	// canned return for GetAllUsers
	users []models.UserResponse
}

func (m *mockUserRepo) AddUser(_ context.Context, userName, passwordHash string) error {
	m.addedUserName = userName
	m.addedHash = passwordHash
	return m.addErr
}

func (m *mockUserRepo) GetAllUsers(_ context.Context) ([]models.UserResponse, error) {
	return m.users, m.getErr
}

func (m *mockUserRepo) UpdateUserPassword(_ context.Context, userName, passwordHash string) error {
	m.updatedUserName = userName
	m.updatedHash = passwordHash
	return m.updateErr
}

func (m *mockUserRepo) DeleteUser(_ context.Context, userName string) error {
	m.deletedUserName = userName
	return m.deleteErr
}

// compile-time check that mockUserRepo satisfies the interface.
var _ repository.UserRepository = (*mockUserRepo)(nil)

// ----- helpers -----

func newService(t *testing.T, repo repository.UserRepository) *services.UserService {
	t.Helper()
	l := log.NewLoggerWithFilename("test")
	return services.NewUserService(l, repo)
}

// ----- CreateUser tests -----

func TestUserService_CreateUser_HashesPassword(t *testing.T) {
	repo := &mockUserRepo{}
	svc := newService(t, repo)

	const plaintext = "super-secret"
	if err := svc.CreateUser(context.Background(), "alice", plaintext); err != nil {
		t.Fatalf("CreateUser returned unexpected error: %v", err)
	}

	// The stored value must NOT be the plaintext password.
	if repo.addedHash == plaintext {
		t.Error("password was stored as plaintext — bcrypt hashing not applied")
	}

	// The stored value must be a valid bcrypt hash of the plaintext.
	if err := bcrypt.CompareHashAndPassword([]byte(repo.addedHash), []byte(plaintext)); err != nil {
		t.Errorf("stored hash does not match original password: %v", err)
	}

	if repo.addedUserName != "alice" {
		t.Errorf("userName: got %q, want %q", repo.addedUserName, "alice")
	}
}

func TestUserService_CreateUser_RepoError(t *testing.T) {
	repo := &mockUserRepo{addErr: errors.New("db error")}
	svc := newService(t, repo)

	err := svc.CreateUser(context.Background(), "bob", "password")
	if err == nil {
		t.Error("expected error from repo, got nil")
	}
}

func TestUserService_CreateUser_DifferentHashEachCall(t *testing.T) {
	// bcrypt salts must differ between calls — ensure the service does not
	// cache or reuse hashes.
	repo1 := &mockUserRepo{}
	repo2 := &mockUserRepo{}
	svc1 := newService(t, repo1)
	svc2 := newService(t, repo2)

	_ = svc1.CreateUser(context.Background(), "u", "same-password")
	_ = svc2.CreateUser(context.Background(), "u", "same-password")

	if repo1.addedHash == repo2.addedHash {
		t.Error("bcrypt produced the same hash twice — salt is not being used correctly")
	}
}

// ----- GetUsers tests -----

func TestUserService_GetUsers_ReturnsUsers(t *testing.T) {
	want := []models.UserResponse{{UserName: "alice"}, {UserName: "bob"}}
	repo := &mockUserRepo{users: want}
	svc := newService(t, repo)

	got, err := svc.GetUsers(context.Background())
	if err != nil {
		t.Fatalf("GetUsers returned unexpected error: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("len(users): got %d, want %d", len(got), len(want))
	}
	for i, u := range got {
		if u.UserName != want[i].UserName {
			t.Errorf("users[%d].UserName: got %q, want %q", i, u.UserName, want[i].UserName)
		}
	}
}

func TestUserService_GetUsers_RepoError(t *testing.T) {
	repo := &mockUserRepo{getErr: errors.New("db error")}
	svc := newService(t, repo)

	_, err := svc.GetUsers(context.Background())
	if err == nil {
		t.Error("expected error from repo, got nil")
	}
}

// ----- UpdateUserPassword tests -----

func TestUserService_UpdateUserPassword_HashesPassword(t *testing.T) {
	repo := &mockUserRepo{}
	svc := newService(t, repo)

	const newPass = "new-secret"
	if err := svc.UpdateUserPassword(context.Background(), "alice", newPass); err != nil {
		t.Fatalf("UpdateUserPassword returned unexpected error: %v", err)
	}

	if repo.updatedHash == newPass {
		t.Error("new password was stored as plaintext — bcrypt hashing not applied")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(repo.updatedHash), []byte(newPass)); err != nil {
		t.Errorf("stored hash does not match new password: %v", err)
	}
	if repo.updatedUserName != "alice" {
		t.Errorf("userName: got %q, want %q", repo.updatedUserName, "alice")
	}
}

func TestUserService_UpdateUserPassword_RepoError(t *testing.T) {
	repo := &mockUserRepo{updateErr: errors.New("db error")}
	svc := newService(t, repo)

	err := svc.UpdateUserPassword(context.Background(), "alice", "newpass")
	if err == nil {
		t.Error("expected error from repo, got nil")
	}
}

// ----- DeleteUser tests -----

func TestUserService_DeleteUser_CallsRepo(t *testing.T) {
	repo := &mockUserRepo{}
	svc := newService(t, repo)

	if err := svc.DeleteUser(context.Background(), "alice"); err != nil {
		t.Fatalf("DeleteUser returned unexpected error: %v", err)
	}
	if repo.deletedUserName != "alice" {
		t.Errorf("deletedUserName: got %q, want %q", repo.deletedUserName, "alice")
	}
}

func TestUserService_DeleteUser_RepoError(t *testing.T) {
	repo := &mockUserRepo{deleteErr: errors.New("db error")}
	svc := newService(t, repo)

	err := svc.DeleteUser(context.Background(), "alice")
	if err == nil {
		t.Error("expected error from repo, got nil")
	}
}
