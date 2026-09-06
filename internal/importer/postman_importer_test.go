// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package importer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"solo/internal/collection"
	"strings"
	"testing"
)

func TestPostmanImporter_Import(t *testing.T) {
	importer := NewPostmanImporter()

	// Use testdata file
	path := filepath.Join("..", "..", "test", "testdata", "postman_v2_1.json")

	coll, err := importer.Import(path)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if coll == nil {
		t.Fatalf("Expected collection to not be nil")
	}

	// Verify Collection Name
	if coll.Name != "Test Collection" {
		t.Errorf("Expected collection name 'Test Collection', got '%s'", coll.Name)
	}

	if len(coll.Requests) != 1 {
		t.Fatalf("Expected 1 root request, got %d", len(coll.Requests))
	}

	if len(coll.Folders) != 1 {
		t.Fatalf("Expected 1 root folder, got %d", len(coll.Folders))
	}

	// 1. Root Request (Object URL)
	req1 := coll.Requests[0]
	if req1.Name != "Root Request" {
		t.Errorf("Expected first request name 'Root Request', got '%s'", req1.Name)
	}
	if req1.Url != "https://api.example.com/data" {
		t.Errorf("Expected first request URL 'https://api.example.com/data', got '%s'", req1.Url)
	}
	if req1.Verb != "GET" {
		t.Errorf("Expected first request verb 'GET', got '%s'", req1.Verb)
	}
	if len(req1.Headers) != 0 {
		t.Errorf("Expected 0 headers, got %d", len(req1.Headers))
	}
	// BodyType should be empty since no body/headers
	if req1.BodyType != "" {
		t.Errorf("Expected empty body type for first request, got '%s'", req1.BodyType)
	}

	folderA := coll.Folders[0]
	if folderA.Name != "Folder A" {
		t.Fatalf("Expected root folder name 'Folder A', got '%s'", folderA.Name)
	}
	if len(folderA.Requests) != 1 {
		t.Fatalf("Expected Folder A to contain 1 request, got %d", len(folderA.Requests))
	}
	if len(folderA.Folders) != 1 {
		t.Fatalf("Expected Folder A to contain 1 subfolder, got %d", len(folderA.Folders))
	}

	// 2. Nested Request 1 (String URL, JSON Body)
	req2 := folderA.Requests[0]
	if req2.Name != "Nested Request 1" {
		t.Errorf("Expected nested request name 'Nested Request 1', got '%s'", req2.Name)
	}
	if req2.Url != "https://api.example.com/users" {
		t.Errorf("Expected second request URL 'https://api.example.com/users', got '%s'", req2.Url)
	}
	if req2.Verb != "POST" {
		t.Errorf("Expected second request verb 'POST', got '%s'", req2.Verb)
	}
	if req2.Headers["Content-Type"] != "application/json" {
		t.Errorf("Expected Content-Type header 'application/json', got '%s'", req2.Headers["Content-Type"])
	}
	if req2.BodyType != "json" { // Based on the "options.raw.language"
		t.Errorf("Expected body type 'json', got '%s'", req2.BodyType)
	}
	if !strings.Contains(req2.Body, "value") {
		t.Errorf("Expected body to contain 'value', got '%s'", req2.Body)
	}

	subFolder := folderA.Folders[0]
	if subFolder.Name != "SubFolder B" {
		t.Fatalf("Expected subfolder name 'SubFolder B', got '%s'", subFolder.Name)
	}
	if len(subFolder.Requests) != 1 {
		t.Fatalf("Expected SubFolder B to contain 1 request, got %d", len(subFolder.Requests))
	}

	// 3. Deeply Nested Request (Object URL, XML fallback)
	req3 := subFolder.Requests[0]
	if req3.Name != "Deeply Nested Request" {
		t.Errorf("Expected third request name 'Deeply Nested Request', got '%s'", req3.Name)
	}
	if req3.Url != "https://api.example.com/update" {
		t.Errorf("Expected third request URL 'https://api.example.com/update', got '%s'", req3.Url)
	}
	if req3.Verb != "PUT" {
		t.Errorf("Expected third request verb 'PUT', got '%s'", req3.Verb)
	}
	if req3.Headers["Content-Type"] != "application/xml" {
		t.Errorf("Expected Content-Type header 'application/xml', got '%s'", req3.Headers["Content-Type"])
	}
	if req3.BodyType != "xml" { // Deduced from Content-Type
		t.Errorf("Expected body type 'xml', got '%s'", req3.BodyType)
	}
	if req3.Body != "<xml></xml>" {
		t.Errorf("Expected body '<xml></xml>', got '%s'", req3.Body)
	}

	if req1.Id == "" || req1.CreationTimestamp.IsZero() {
		t.Errorf("Expected root request to have ID and creation timestamp")
	}
	if req2.Id == "" || req2.CreationTimestamp.IsZero() {
		t.Errorf("Expected nested request to have ID and creation timestamp")
	}
	if req3.Id == "" || req3.CreationTimestamp.IsZero() {
		t.Errorf("Expected deeply nested request to have ID and creation timestamp")
	}
	if folderA.Id == "" || folderA.CreationTimestamp.IsZero() {
		t.Errorf("Expected Folder A to have ID and creation timestamp")
	}
	if subFolder.Id == "" || subFolder.CreationTimestamp.IsZero() {
		t.Errorf("Expected SubFolder B to have ID and creation timestamp")
	}
}

func TestPostmanImporter_Import_FileNotFound(t *testing.T) {
	importer := NewPostmanImporter()
	_, err := importer.Import("does_not_exist.json")
	if err == nil {
		t.Fatal("Expected error for non-existent file, got nil")
	}
}

func TestPostmanImporter_ImportsAndInheritsAuth(t *testing.T) {
	path := filepath.Join(t.TempDir(), "auth.postman_collection.json")
	content := `{
  "info": {"name": "Auth collection"},
  "auth": {"type": "bearer", "bearer": [{"key": "token", "value": "collection-token"}]},
  "item": [
    {"name": "Inherited", "request": {"method": "GET", "url": "https://example.com/inherited"}},
    {"name": "No auth", "request": {"method": "GET", "url": "https://example.com/none", "auth": {"type": "noauth"}}},
    {"name": "Folder", "auth": {"type": "bearer", "bearer": [{"key": "token", "value": "folder-token"}]}, "item": [
      {"name": "Folder inherited", "request": {"method": "GET", "url": "https://example.com/folder"}},
      {"name": "OAuth access token", "request": {"method": "GET", "url": "https://example.com/oauth", "auth": {"type": "oauth2", "oauth2": [{"key": "accessToken", "value": "oauth-token"}]}}}
    ]}
  ]
}`
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	coll, err := NewPostmanImporter().Import(path)
	if err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if got := coll.Requests[0].Auth.BearerToken; got != "collection-token" {
		t.Fatalf("collection inherited token = %q", got)
	}
	if coll.Requests[1].Auth != nil {
		t.Fatalf("noauth request auth = %+v, want nil", coll.Requests[1].Auth)
	}
	if got := coll.Folders[0].Requests[0].Auth.BearerToken; got != "folder-token" {
		t.Fatalf("folder inherited token = %q", got)
	}
	if got := coll.Folders[0].Requests[1].Auth.BearerToken; got != "oauth-token" {
		t.Fatalf("OAuth access token = %q", got)
	}
}

func TestPostmanURL_UnmarshalJSON(t *testing.T) {
	importer := NewPostmanImporter() // Used to force syntax check of postmanURL

	_ = importer

	// Test Unmarshal String
	jsonStr := `{"url": "http://example.com"}`
	var reqStr postmanRequest
	if err := json.Unmarshal([]byte(jsonStr), &reqStr); err != nil {
		t.Fatalf("Failed to unmarshal string URL: %v", err)
	}
	if reqStr.URL.Raw != "http://example.com" {
		t.Errorf("Expected URL 'http://example.com', got '%s'", reqStr.URL.Raw)
	}

	// Test Unmarshal Object
	jsonObj := `{"url": {"raw": "http://example.com/obj", "host": ["example", "com"]}}`
	var reqObj postmanRequest
	if err := json.Unmarshal([]byte(jsonObj), &reqObj); err != nil {
		t.Fatalf("Failed to unmarshal object URL: %v", err)
	}
	if reqObj.URL.Raw != "http://example.com/obj" {
		t.Errorf("Expected URL 'http://example.com/obj', got '%s'", reqObj.URL.Raw)
	}
}

func TestPostmanImporter_FormBodies(t *testing.T) {
	content := `{
		"info": { "name": "Form Test" },
		"item": [
			{
				"name": "Urlencoded",
				"request": {
					"method": "POST",
					"url": "http://example.com/form",
					"body": {
						"mode": "urlencoded",
						"urlencoded": [
							{ "key": "foo", "value": "bar" },
							{ "key": "baz", "value": "qux" }
						]
					}
				}
			},
			{
				"name": "Formdata",
				"request": {
					"method": "POST",
					"url": "http://example.com/multipart",
					"body": {
						"mode": "formdata",
						"formdata": [
							{ "key": "text", "value": "hello", "type": "text" },
							{ "key": "file", "type": "file" }
						]
					}
				}
			}
		]
	}`

	tmp := t.TempDir() + "/postman_forms.json"
	_ = os.WriteFile(tmp, []byte(content), 0600)

	imp := NewPostmanImporter()
	coll, err := imp.Import(tmp)
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if len(coll.Requests) != 2 {
		t.Fatalf("Expected 2 requests, got %d", len(coll.Requests))
	}

	// Verify Urlencoded
	var req *collection.Request
	for i := range coll.Requests {
		if coll.Requests[i].Name == "Urlencoded" {
			req = &coll.Requests[i]
		}
	}
	if req == nil {
		t.Fatal("Urlencoded request not found")
	}
	if req.Headers["Content-Type"] != "application/x-www-form-urlencoded" {
		t.Errorf("Urlencoded Content-Type: got %q", req.Headers["Content-Type"])
	}
	if !strings.Contains(req.Body, "foo=bar") || !strings.Contains(req.Body, "baz=qux") {
		t.Errorf("Urlencoded body: got %q", req.Body)
	}

	// Verify Formdata
	req = nil
	for i := range coll.Requests {
		if coll.Requests[i].Name == "Formdata" {
			req = &coll.Requests[i]
		}
	}
	if req == nil {
		t.Fatal("Formdata request not found")
	}
	if !strings.Contains(req.Headers["Content-Type"], "multipart/form-data") {
		t.Errorf("Formdata Content-Type: got %q", req.Headers["Content-Type"])
	}
	if !strings.Contains(req.Body, "name=\"text\"") || !strings.Contains(req.Body, "hello") {
		t.Errorf("Formdata body text: got %q", req.Body)
	}
	if !strings.Contains(req.Body, "name=\"file\"") || !strings.Contains(req.Body, "[BINARY_FILE_CONTENT]") {
		t.Errorf("Formdata body file: got %q", req.Body)
	}
}
