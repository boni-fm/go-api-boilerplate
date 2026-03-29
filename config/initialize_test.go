package config_test

import (
	"os"
	"testing"

	"go-api-boilerplate/config"
)

// writeTempINI creates a temporary INI file with the given content and returns
// its path. The caller is responsible for removing the file after use.
func writeTempINI(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "appsettings-*.ini")
	if err != nil {
		t.Fatalf("failed to create temp ini file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("failed to write temp ini file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoadConfigIniFromPath_AllFieldsPresent(t *testing.T) {
	ini := `
[CONFIG]
AppName       = MyService
IsDevelopment = true
Port          = 9090
Kunci         = secret123
`
	path := writeTempINI(t, ini)
	cfg := config.LoadConfigIniFromPath(path)

	if cfg.AppName != "MyService" {
		t.Errorf("AppName: got %q, want %q", cfg.AppName, "MyService")
	}
	if !cfg.IsDevelopment {
		t.Error("IsDevelopment: got false, want true")
	}
	if cfg.Port != "9090" {
		t.Errorf("Port: got %q, want %q", cfg.Port, "9090")
	}
	if cfg.Kunci != "secret123" {
		t.Errorf("Kunci: got %q, want %q", cfg.Kunci, "secret123")
	}
}

func TestLoadConfigIniFromPath_Defaults(t *testing.T) {
	// Minimal INI — only the section header, no keys.
	path := writeTempINI(t, "[CONFIG]\n")
	cfg := config.LoadConfigIniFromPath(path)

	if cfg.AppName != "Go API Boilerplate" {
		t.Errorf("AppName default: got %q, want %q", cfg.AppName, "Go API Boilerplate")
	}
	if cfg.IsDevelopment {
		t.Error("IsDevelopment default: got true, want false")
	}
	if cfg.Port != "8080" {
		t.Errorf("Port default: got %q, want %q", cfg.Port, "8080")
	}
	if cfg.Kunci != "" {
		t.Errorf("Kunci default: got %q, want empty string", cfg.Kunci)
	}
}

func TestLoadConfigIniFromPath_IsDevelopmentFalse(t *testing.T) {
	ini := "[CONFIG]\nIsDevelopment = false\n"
	path := writeTempINI(t, ini)
	cfg := config.LoadConfigIniFromPath(path)
	if cfg.IsDevelopment {
		t.Error("IsDevelopment: got true, want false")
	}
}
