package swagger

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// DocumentModifier ngurusin modifikasi dokumen swagger
type DocumentModifier struct {
	docPath string
}

// NewDocumentModifier bikin DocumentModifier baru, baca dari docs/swagger.json
func NewDocumentModifier() *DocumentModifier {
	return &DocumentModifier{
		docPath: GetDocPath(),
	}
}

// NewDocumentModifierWithPath bikin DocumentModifier dengan path custom —
// berguna buat testing biar gak bergantung sama file swagger.json asli
func NewDocumentModifierWithPath(docPath string) *DocumentModifier {
	return &DocumentModifier{docPath: docPath}
}

// GetModifiedDocument baca swagger doc terus modif sesuai proxy config
func (dm *DocumentModifier) GetModifiedDocument(proxyPath string) (map[string]interface{}, error) {
	doc, err := dm.readSwaggerDoc()
	if err != nil {
		return nil, fmt.Errorf("gagal baca swagger doc: %w", err)
	}

	doc["servers"] = dm.buildServers(proxyPath)

	// Swagger 2.0 compat — set basePath kalau ada proxy
	if proxyPath != "" {
		normalizedPath := normalizeProxyPath(proxyPath)
		doc["basePath"] = normalizedPath
		delete(doc, "host") // hapus host kalau pakai proxy
	}

	return doc, nil
}

// readSwaggerDoc baca dan parse file swagger.json
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

// buildServers bangun array servers berdasarkan proxy path
func (dm *DocumentModifier) buildServers(proxyPath string) []map[string]interface{} {
	servers := []map[string]interface{}{}

	if proxyPath != "" {
		normalizedPath := normalizeProxyPath(proxyPath)
		// proxy path duluan biar jadi default
		servers = append(servers, map[string]interface{}{
			"url":         normalizedPath,
			"description": "Reverse Proxy Path",
		})
		servers = append(servers, map[string]interface{}{
			"url":         "/",
			"description": "Root Url Server",
		})
	} else {
		servers = append(servers, map[string]interface{}{
			"url":         "/",
			"description": "Root Url Server",
		})
	}

	return servers
}

// normalizeProxyPath mastiin proxy path diawali / atau http
func normalizeProxyPath(path string) string {
	path = strings.TrimSpace(path)

	if path == "" {
		return "/"
	}

	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return strings.TrimSuffix(path, "/")
}
