package swagger

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// DocumentModifier handles swagger document modifications
type DocumentModifier struct {
	docPath string
}

// NewDocumentModifier creates a new DocumentModifier that reads from the
// default docs/swagger.json location relative to the current working directory.
func NewDocumentModifier() *DocumentModifier {
	return &DocumentModifier{
		docPath: GetDocPath(),
	}
}

// NewDocumentModifierWithPath creates a DocumentModifier that reads from the
// specified path. Use this in tests to avoid a dependency on the generated
// swagger.json file in the project's docs/ directory.
func NewDocumentModifierWithPath(docPath string) *DocumentModifier {
	return &DocumentModifier{docPath: docPath}
}

// GetModifiedDocument reads and modifies swagger document based on proxy configuration
func (dm *DocumentModifier) GetModifiedDocument(proxyPath string) (map[string]interface{}, error) {
	// Read swagger file
	doc, err := dm.readSwaggerDoc()
	if err != nil {
		return nil, fmt.Errorf("failed to read swagger doc: %w", err)
	}

	// Apply server modifications
	servers := dm.buildServers(proxyPath)
	doc["servers"] = servers

	// For Swagger 2.0 compatibility - modify basePath
	if proxyPath != "" {
		normalizedPath := normalizeProxyPath(proxyPath)
		doc["basePath"] = normalizedPath
		// Remove host when using proxy
		delete(doc, "host")
	}

	return doc, nil
}

// readSwaggerDoc reads and parses the swagger.json file
func (dm *DocumentModifier) readSwaggerDoc() (map[string]interface{}, error) {
	fileContent, err := os.ReadFile(dm.docPath)
	if err != nil {
		return nil, err
	}

	var doc map[string]interface{}
	if err := json.Unmarshal(fileContent, &doc); err != nil {
		return nil, err
	}

	return doc, nil
}

// buildServers constructs the servers array based on proxy path
func (dm *DocumentModifier) buildServers(proxyPath string) []map[string]interface{} {
	servers := []map[string]interface{}{}

	if proxyPath != "" {
		// Normalize proxy path
		normalizedPath := normalizeProxyPath(proxyPath)

		// Add proxy server first (this will be the default)
		servers = append(servers, map[string]interface{}{
			"url":         normalizedPath,
			"description": "Reverse Proxy Path",
		})

		// Add root server as alternative
		servers = append(servers, map[string]interface{}{
			"url":         "/",
			"description": "Root Url Server",
		})
	} else {
		// No proxy, just root server
		servers = append(servers, map[string]interface{}{
			"url":         "/",
			"description": "Root Url Server",
		})
	}

	return servers
}

// normalizeProxyPath ensures the proxy path starts with / or http
func normalizeProxyPath(path string) string {
	path = strings.TrimSpace(path)

	if path == "" {
		return "/"
	}

	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}

	if !strings.HasPrefix(path, "/") {
		return "/" + path
	}

	// Remove trailing slash
	path = strings.TrimSuffix(path, "/")

	return path
}
