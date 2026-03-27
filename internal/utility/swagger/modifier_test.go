package swagger_test

import (
	"encoding/json"
	"os"
	"testing"

	"go-api-boilerplate/internal/utility/swagger"
)

// writeSwaggerJSON writes a minimal swagger.json to a temp file and returns the path.
func writeSwaggerJSON(t *testing.T, extra map[string]interface{}) string {
	t.Helper()
	doc := map[string]interface{}{
		"swagger": "2.0",
		"info":    map[string]interface{}{"title": "Test API", "version": "1.0"},
		"paths":   map[string]interface{}{},
	}
	for k, v := range extra {
		doc[k] = v
	}
	b, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("failed to marshal swagger doc: %v", err)
	}
	f, err := os.CreateTemp(t.TempDir(), "swagger-*.json")
	if err != nil {
		t.Fatalf("failed to create temp swagger file: %v", err)
	}
	if _, err := f.Write(b); err != nil {
		t.Fatalf("failed to write temp swagger file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestGetModifiedDocument_NoProxyPath(t *testing.T) {
	path := writeSwaggerJSON(t, nil)
	dm := swagger.NewDocumentModifierWithPath(path)

	doc, err := dm.GetModifiedDocument("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	servers, ok := doc["servers"].([]map[string]interface{})
	if !ok {
		t.Fatal("servers field missing or wrong type")
	}
	if len(servers) != 1 {
		t.Fatalf("servers count: got %d, want 1", len(servers))
	}
	if servers[0]["url"] != "/" {
		t.Errorf("server url: got %v, want \"/\"", servers[0]["url"])
	}
}

func TestGetModifiedDocument_WithProxyPath(t *testing.T) {
	path := writeSwaggerJSON(t, nil)
	dm := swagger.NewDocumentModifierWithPath(path)

	doc, err := dm.GetModifiedDocument("/my-prefix")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	servers, ok := doc["servers"].([]map[string]interface{})
	if !ok {
		t.Fatal("servers field missing or wrong type")
	}
	if len(servers) != 2 {
		t.Fatalf("servers count: got %d, want 2", len(servers))
	}
	if servers[0]["url"] != "/my-prefix" {
		t.Errorf("first server url: got %v, want \"/my-prefix\"", servers[0]["url"])
	}
	// basePath should be set when proxy path is present.
	if doc["basePath"] != "/my-prefix" {
		t.Errorf("basePath: got %v, want \"/my-prefix\"", doc["basePath"])
	}
	// host key should be removed when a proxy path is set.
	if _, hasHost := doc["host"]; hasHost {
		t.Error("host key should be removed when proxy path is set")
	}
}

func TestGetModifiedDocument_MissingFile(t *testing.T) {
	dm := swagger.NewDocumentModifierWithPath("/nonexistent/swagger.json")
	_, err := dm.GetModifiedDocument("")
	if err == nil {
		t.Error("expected an error for missing file, got nil")
	}
}

func TestGetModifiedDocument_InvalidJSON(t *testing.T) {
	f, _ := os.CreateTemp(t.TempDir(), "swagger-*.json")
	f.WriteString("not-valid-json{{{")
	f.Close()

	dm := swagger.NewDocumentModifierWithPath(f.Name())
	_, err := dm.GetModifiedDocument("")
	if err == nil {
		t.Error("expected an error for invalid JSON, got nil")
	}
}
