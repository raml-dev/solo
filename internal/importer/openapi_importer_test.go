// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package importer

import (
	"os"
	"path/filepath"
	"solo/internal/collection"
	"strings"
	"testing"
)

const testdataDir = "../../test/testdata"

// findRequest finds a request by verb+url suffix, returns nil if not found.
func findRequest(t *testing.T, results []requestSummary, verb, urlSuffix string) *requestSummary {
	t.Helper()
	for i, r := range results {
		if r.verb == verb && strings.HasSuffix(r.url, urlSuffix) {
			return &results[i]
		}
	}
	return nil
}

type requestSummary struct {
	name     string
	verb     string
	url      string
	bodyType string
	body     string
	headers  map[string]string
}

func summarize(t *testing.T, path string) ([]requestSummary, []string) {
	t.Helper()
	imp := NewOpenAPIImporter()
	result, err := imp.Import(path)
	if err != nil {
		t.Fatalf("Import(%q) unexpected error: %v", path, err)
	}
	summaries := make([]requestSummary, 0, len(result.Collection.Requests))
	for _, req := range result.Collection.Requests {
		summaries = append(summaries, requestSummary{
			name:     req.Name,
			verb:     req.Verb,
			url:      req.Url,
			bodyType: req.BodyType,
			body:     req.Body,
			headers:  req.Headers,
		})
	}
	return summaries, result.Warnings
}

func summarizeRequests(reqs []collection.Request) []requestSummary {
	summaries := make([]requestSummary, 0, len(reqs))
	for _, req := range reqs {
		summaries = append(summaries, requestSummary{
			name:     req.Name,
			verb:     req.Verb,
			url:      req.Url,
			bodyType: req.BodyType,
			body:     req.Body,
			headers:  req.Headers,
		})
	}
	return summaries
}

// ── OpenAPI 3.x ──────────────────────────────────────────────────────────────

func testOpenAPI3(t *testing.T, path string) {
	t.Helper()
	imp := NewOpenAPIImporter()
	result, err := imp.Import(path)
	if err != nil {
		t.Fatalf("Import(%q) unexpected error: %v", path, err)
	}

	coll := result.Collection
	if coll.Name != "OpenAPI Test Collection" {
		t.Errorf("collection name: got %q, want %q", coll.Name, "OpenAPI Test Collection")
	}
	if got := coll.Variables["baseUrl"].Value; got != "https://api.example.com" {
		t.Errorf("collection baseUrl variable: got %q, want %q", got, "https://api.example.com")
	}
	if len(coll.Requests) != 1 {
		t.Fatalf("root request count: got %d, want 1", len(coll.Requests))
	}
	if len(coll.Folders) != 1 {
		t.Fatalf("folder count: got %d, want 1", len(coll.Folders))
	}
	if coll.Folders[0].Name != "users" {
		t.Fatalf("folder name: got %q, want %q", coll.Folders[0].Name, "users")
	}
	if len(coll.Folders[0].Requests) != 3 {
		t.Fatalf("tagged request count: got %d, want 3", len(coll.Folders[0].Requests))
	}

	reqs := summarizeRequests(coll.Folders[0].Requests)
	rootReqs := summarizeRequests(coll.Requests)
	warnings := result.Warnings

	// GET /users — operationId takes precedence as name
	r := findRequest(t, reqs, "GET", "/users")
	if r == nil {
		t.Fatal("GET /users not found")
	}
	if r.name != "listUsers" {
		t.Errorf("GET /users name: got %q, want %q", r.name, "listUsers")
	}
	if r.url != "{{baseUrl}}/users" {
		t.Errorf("GET /users url: got %q", r.url)
	}
	if _, ok := r.headers["X-Request-Id"]; !ok {
		t.Error("GET /users: expected header X-Request-Id")
	}
	if r.headers["X-Request-Id"] != "" {
		t.Errorf("GET /users: header X-Request-Id should be empty placeholder, got %q", r.headers["X-Request-Id"])
	}
	if r.bodyType != "" {
		t.Errorf("GET /users: expected no bodyType, got %q", r.bodyType)
	}

	// POST /users — summary used as name (no operationId)
	r = findRequest(t, reqs, "POST", "/users")
	if r == nil {
		t.Fatal("POST /users not found")
	}
	if r.name != "Create user" {
		t.Errorf("POST /users name: got %q, want %q", r.name, "Create user")
	}
	if r.bodyType != "json" {
		t.Errorf("POST /users bodyType: got %q, want %q", r.bodyType, "json")
	}
	if r.body != "{}" {
		t.Errorf("POST /users body: got %q, want %q", r.body, "{}")
	}

	// PUT /users/{id} — operationId
	r = findRequest(t, reqs, "PUT", "/users/{{id}}")
	if r == nil {
		t.Fatal("PUT /users/{id} not found")
	}
	if r.name != "updateUser" {
		t.Errorf("PUT /users/{id} name: got %q, want %q", r.name, "updateUser")
	}
	if r.url != "{{baseUrl}}/users/{{id}}" {
		t.Errorf("PUT /users/{id} url: got %q", r.url)
	}
	if r.bodyType != "json" {
		t.Errorf("PUT /users/{id} bodyType: got %q, want %q", r.bodyType, "json")
	}

	// DELETE /users/{id} — fallback name (no operationId, no summary) and no tags => root
	r = findRequest(t, rootReqs, "DELETE", "/users/{{id}}")
	if r == nil {
		t.Fatal("DELETE /users/{id} not found")
	}
	if r.name != "DELETE /users/{id}" {
		t.Errorf("DELETE /users/{id} name: got %q, want %q", r.name, "DELETE /users/{id}")
	}
	if r.url != "{{baseUrl}}/users/{{id}}" {
		t.Errorf("DELETE /users/{id} url: got %q, want %q", r.url, "{{baseUrl}}/users/{{id}}")
	}
	if r.bodyType != "" {
		t.Errorf("DELETE /users/{id}: expected no bodyType, got %q", r.bodyType)
	}
	if result.BasePath != "" {
		t.Errorf("openapi3 basePath: got %q, want empty", result.BasePath)
	}
	if len(result.Servers) != 1 || result.Servers[0] != "https://api.example.com" {
		t.Errorf("openapi3 servers: got %v, want %v", result.Servers, []string{"https://api.example.com"})
	}

	// Security warnings — fixture has 2 schemes: bearerAuth, apiKey
	if len(warnings) == 0 {
		t.Error("expected security warnings, got none")
	}
	if !strings.Contains(warnings[0], "bearerAuth") && !strings.Contains(warnings[0], "apiKey") {
		t.Errorf("warning should mention scheme names, got: %q", warnings[0])
	}
}

func TestOpenAPIImporter_OpenAPI3_JSON(t *testing.T) {
	testOpenAPI3(t, filepath.Join(testdataDir, "openapi_3_0.json"))
}

func TestOpenAPIImporter_OpenAPI3_YAML(t *testing.T) {
	testOpenAPI3(t, filepath.Join(testdataDir, "openapi_3_0.yaml"))
}

// ── Swagger 2.x ──────────────────────────────────────────────────────────────

func testSwagger2(t *testing.T, path string) {
	t.Helper()
	imp := NewOpenAPIImporter()
	result, err := imp.Import(path)
	if err != nil {
		t.Fatalf("Import(%q) unexpected error: %v", path, err)
	}

	coll := result.Collection
	if coll.Name != "Swagger Test Collection" {
		t.Errorf("collection name: got %q, want %q", coll.Name, "Swagger Test Collection")
	}
	if got := coll.Variables["baseUrl"].Value; got != "https://api.example.com/v1" {
		t.Errorf("collection baseUrl variable: got %q, want %q", got, "https://api.example.com/v1")
	}
	if len(coll.Requests) != 1 {
		t.Fatalf("root request count: got %d, want 1", len(coll.Requests))
	}
	if len(coll.Folders) != 1 {
		t.Fatalf("folder count: got %d, want 1", len(coll.Folders))
	}
	if coll.Folders[0].Name != "users" {
		t.Fatalf("folder name: got %q, want %q", coll.Folders[0].Name, "users")
	}
	if len(coll.Folders[0].Requests) != 3 {
		t.Fatalf("tagged request count: got %d, want 3", len(coll.Folders[0].Requests))
	}

	reqs := summarizeRequests(coll.Folders[0].Requests)
	rootReqs := summarizeRequests(coll.Requests)
	warnings := result.Warnings

	// GET /users
	r := findRequest(t, reqs, "GET", "/users")
	if r == nil {
		t.Fatal("GET /users not found")
	}
	if r.url != "{{baseUrl}}/users" {
		t.Errorf("GET /users url: got %q, want %q", r.url, "{{baseUrl}}/users")
	}
	if r.name != "listUsers" {
		t.Errorf("GET /users name: got %q, want %q", r.name, "listUsers")
	}
	if _, ok := r.headers["X-Request-Id"]; !ok {
		t.Error("GET /users: expected header X-Request-Id")
	}

	// POST /users — body param + root consumes → BodyType json
	r = findRequest(t, reqs, "POST", "/users")
	if r == nil {
		t.Fatal("POST /users not found")
	}
	if r.name != "createUser" {
		t.Errorf("POST /users name: got %q, want %q", r.name, "createUser")
	}
	if r.bodyType != "json" {
		t.Errorf("POST /users bodyType: got %q, want %q", r.bodyType, "json")
	}
	if r.body != "{}" {
		t.Errorf("POST /users body: got %q, want %q", r.body, "{}")
	}

	// PUT /users/{id} — operation-level consumes overrides root
	r = findRequest(t, reqs, "PUT", "/users/{{id}}")
	if r == nil {
		t.Fatal("PUT /users/{id} not found")
	}
	if r.bodyType != "json" {
		t.Errorf("PUT /users/{id} bodyType: got %q, want %q", r.bodyType, "json")
	}
	if r.url != "{{baseUrl}}/users/{{id}}" {
		t.Errorf("PUT /users/{id} url: got %q, want %q", r.url, "{{baseUrl}}/users/{{id}}")
	}

	// DELETE /users/{id} — fallback name and no tags => root
	r = findRequest(t, rootReqs, "DELETE", "/users/{{id}}")
	if r == nil {
		t.Fatal("DELETE /users/{id} not found")
	}
	if r.name != "DELETE /users/{id}" {
		t.Errorf("DELETE /users/{id} name: got %q, want %q", r.name, "DELETE /users/{id}")
	}
	if r.url != "{{baseUrl}}/users/{{id}}" {
		t.Errorf("DELETE /users/{id} url: got %q, want %q", r.url, "{{baseUrl}}/users/{{id}}")
	}
	if result.BasePath != "/v1" {
		t.Errorf("swagger2 basePath: got %q, want %q", result.BasePath, "/v1")
	}
	if len(result.Servers) != 0 {
		t.Errorf("swagger2 servers: got %v, want none", result.Servers)
	}

	// Security warnings — fixture has 2 definitions: basicAuth, apiKey
	if len(warnings) == 0 {
		t.Error("expected security warnings, got none")
	}
	if !strings.Contains(warnings[0], "basicAuth") && !strings.Contains(warnings[0], "apiKey") {
		t.Errorf("warning should mention scheme names, got: %q", warnings[0])
	}
}

func TestOpenAPIImporter_Swagger2_JSON(t *testing.T) {
	testSwagger2(t, filepath.Join(testdataDir, "swagger_2_0.json"))
}

func TestOpenAPIImporter_Swagger2_YAML(t *testing.T) {
	testSwagger2(t, filepath.Join(testdataDir, "swagger_2_0.yaml"))
}

// ── Swagger 2.x no consumes — defaults to JSON ────────────────────────────────

func TestOpenAPIImporter_Swagger2_NoConsumes(t *testing.T) {
	// Inline minimal document: body param, no consumes at root or operation level
	// We write it to a temp file to reuse the file-based Import() path.
	content := `{
		"swagger": "2.0",
		"info": { "title": "No Consumes", "version": "1.0" },
		"host": "api.example.com",
		"basePath": "/",
		"paths": {
			"/items": {
				"post": {
					"operationId": "createItem",
					"parameters": [
						{ "name": "body", "in": "body", "schema": { "type": "object" } }
					],
					"responses": { "201": { "description": "Created" } }
				}
			}
		}
	}`

	tmp := t.TempDir() + "/no_consumes.json"
	if err := writeTemp(tmp, content); err != nil {
		t.Fatal(err)
	}

	imp := NewOpenAPIImporter()
	result, err := imp.Import(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Collection.Requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(result.Collection.Requests))
	}
	req := result.Collection.Requests[0]
	if req.BodyType != "json" {
		t.Errorf("bodyType: got %q, want %q", req.BodyType, "json")
	}
}

// ── Error cases ───────────────────────────────────────────────────────────────

func TestOpenAPIImporter_ErrorCases(t *testing.T) {
	imp := NewOpenAPIImporter()

	t.Run("file_not_found", func(t *testing.T) {
		_, err := imp.Import("does_not_exist.json")
		if err == nil {
			t.Fatal("expected error for missing file, got nil")
		}
	})

	t.Run("unsupported_version_openapi_1", func(t *testing.T) {
		tmp := t.TempDir() + "/bad.json"
		_ = writeTemp(tmp, `{"openapi":"1.0","info":{"title":"X"},"paths":{}}`)
		_, err := imp.Import(tmp)
		if err == nil {
			t.Fatal("expected error for openapi 1.0, got nil")
		}
		if !strings.Contains(strings.ToLower(err.Error()), "unsupported") {
			t.Errorf("error should mention 'unsupported', got: %v", err)
		}
	})

	t.Run("unsupported_version_swagger_1", func(t *testing.T) {
		tmp := t.TempDir() + "/bad.json"
		_ = writeTemp(tmp, `{"swagger":"1.0","info":{"title":"X"},"paths":{}}`)
		_, err := imp.Import(tmp)
		if err == nil {
			t.Fatal("expected error for swagger 1.0, got nil")
		}
		if !strings.Contains(strings.ToLower(err.Error()), "unsupported") {
			t.Errorf("error should mention 'unsupported', got: %v", err)
		}
	})

	t.Run("empty_paths_openapi3", func(t *testing.T) {
		tmp := t.TempDir() + "/empty.json"
		_ = writeTemp(tmp, `{"openapi":"3.0.0","info":{"title":"Empty"},"paths":{}}`)
		result, err := imp.Import(tmp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Collection.Requests) != 0 {
			t.Errorf("expected 0 requests, got %d", len(result.Collection.Requests))
		}
		if len(result.Collection.Folders) != 0 {
			t.Errorf("expected 0 folders, got %d", len(result.Collection.Folders))
		}
	})

	t.Run("empty_paths_swagger2", func(t *testing.T) {
		tmp := t.TempDir() + "/empty.json"
		_ = writeTemp(tmp, `{"swagger":"2.0","info":{"title":"Empty"},"host":"x.com","paths":{}}`)
		result, err := imp.Import(tmp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Collection.Requests) != 0 {
			t.Errorf("expected 0 requests, got %d", len(result.Collection.Requests))
		}
		if len(result.Collection.Folders) != 0 {
			t.Errorf("expected 0 folders, got %d", len(result.Collection.Folders))
		}
	})

	t.Run("no_servers_openapi3_uses_baseurl_placeholder", func(t *testing.T) {
		tmp := t.TempDir() + "/no_servers.json"
		_ = writeTemp(tmp, `{"openapi":"3.0.0","info":{"title":"T"},"paths":{"/ping":{"get":{"operationId":"ping"}}}}`)
		result, err := imp.Import(tmp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Collection.Requests) != 1 {
			t.Fatalf("expected 1 request, got %d", len(result.Collection.Requests))
		}
		if result.Collection.Requests[0].Url != "{{baseUrl}}/ping" {
			t.Errorf("url: got %q, want %q", result.Collection.Requests[0].Url, "{{baseUrl}}/ping")
		}
		if len(result.Servers) != 0 {
			t.Errorf("servers: got %v, want none", result.Servers)
		}
		if _, ok := result.Collection.Variables["baseUrl"]; ok {
			t.Errorf("collection baseUrl variable should be absent when no servers are declared")
		}
	})

	t.Run("no_servers_openapi3_uses_normalized_basepath_as_baseurl", func(t *testing.T) {
		testCases := []struct {
			name         string
			basePath     string
			wantBaseURL  string
			wantVariable bool
		}{
			{
				name:         "normalizes_missing_leading_slash_and_trailing_slash",
				basePath:     "api/",
				wantBaseURL:  "/api",
				wantVariable: true,
			},
			{
				name:         "root_basepath_collapses_to_empty",
				basePath:     "/",
				wantBaseURL:  "",
				wantVariable: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				tmp := t.TempDir() + "/no_servers_basepath.json"
				_ = writeTemp(tmp, `{"openapi":"3.0.0","info":{"title":"T"},"servers":[],"basePath":"`+tc.basePath+`","paths":{"/ping":{"get":{"operationId":"ping"}}}}`)
				result, err := imp.Import(tmp)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				baseURL, ok := result.Collection.Variables["baseUrl"]
				if ok != tc.wantVariable {
					t.Fatalf("baseUrl variable presence = %v, want %v", ok, tc.wantVariable)
				}
				if tc.wantVariable && baseURL.Value != tc.wantBaseURL {
					t.Errorf("baseUrl value: got %q, want %q", baseURL.Value, tc.wantBaseURL)
				}
			})
		}
	})

	t.Run("no_host_swagger2_uses_baseurl_placeholder", func(t *testing.T) {
		tmp := t.TempDir() + "/no_host.json"
		_ = writeTemp(tmp, `{"swagger":"2.0","info":{"title":"T"},"basePath":"/api","paths":{"/ping":{"get":{"operationId":"ping"}}}}`)
		result, err := imp.Import(tmp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Collection.Requests) != 1 {
			t.Fatalf("expected 1 request, got %d", len(result.Collection.Requests))
		}
		if result.Collection.Requests[0].Url != "{{baseUrl}}/ping" {
			t.Errorf("url: got %q, want %q", result.Collection.Requests[0].Url, "{{baseUrl}}/ping")
		}
		if result.BasePath != "/api" {
			t.Errorf("basePath: got %q, want %q", result.BasePath, "/api")
		}
		baseURL, ok := result.Collection.Variables["baseUrl"]
		if !ok {
			t.Fatal("collection baseUrl variable should be present when swagger basePath is defined")
		}
		if baseURL.Value != "/api" {
			t.Errorf("baseUrl value: got %q, want %q", baseURL.Value, "/api")
		}
	})

	t.Run("no_host_swagger2_uses_normalized_basepath_as_baseurl", func(t *testing.T) {
		testCases := []struct {
			name         string
			basePath     string
			wantBaseURL  string
			wantVariable bool
		}{
			{
				name:         "normalizes_missing_leading_slash_and_trailing_slash",
				basePath:     "api/",
				wantBaseURL:  "/api",
				wantVariable: true,
			},
			{
				name:         "root_basepath_collapses_to_empty",
				basePath:     "/",
				wantBaseURL:  "",
				wantVariable: false,
			},
			{
				name:         "missing_basepath_keeps_baseurl_absent",
				basePath:     "",
				wantBaseURL:  "",
				wantVariable: false,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				tmp := t.TempDir() + "/no_host_normalized.json"
				doc := `{"swagger":"2.0","info":{"title":"T"},"paths":{"/ping":{"get":{"operationId":"ping"}}}}`
				if tc.basePath != "" {
					doc = `{"swagger":"2.0","info":{"title":"T"},"basePath":"` + tc.basePath + `","paths":{"/ping":{"get":{"operationId":"ping"}}}}`
				}
				_ = writeTemp(tmp, doc)
				result, err := imp.Import(tmp)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				baseURL, ok := result.Collection.Variables["baseUrl"]
				if ok != tc.wantVariable {
					t.Fatalf("baseUrl variable presence = %v, want %v", ok, tc.wantVariable)
				}
				if tc.wantVariable && baseURL.Value != tc.wantBaseURL {
					t.Errorf("baseUrl value: got %q, want %q", baseURL.Value, tc.wantBaseURL)
				}
			})
		}
	})
}

// ── Security warnings ─────────────────────────────────────────────────────────

func TestOpenAPIImporter_SecurityWarnings(t *testing.T) {
	imp := NewOpenAPIImporter()

	t.Run("openapi3_with_security_schemes", func(t *testing.T) {
		result, err := imp.Import(filepath.Join(testdataDir, "openapi_3_0.json"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Warnings) == 0 {
			t.Fatal("expected warnings for security schemes, got none")
		}
		w := result.Warnings[0]
		if !strings.Contains(w, "bearerAuth") && !strings.Contains(w, "apiKey") {
			t.Errorf("warning should mention scheme names, got: %q", w)
		}
		if !strings.Contains(strings.ToLower(w), "header") && !strings.Contains(strings.ToLower(w), "environment") {
			t.Errorf("warning should mention headers/environment, got: %q", w)
		}
	})

	t.Run("swagger2_with_security_definitions", func(t *testing.T) {
		result, err := imp.Import(filepath.Join(testdataDir, "swagger_2_0.json"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Warnings) == 0 {
			t.Fatal("expected warnings for security definitions, got none")
		}
		w := result.Warnings[0]
		if !strings.Contains(w, "basicAuth") && !strings.Contains(w, "apiKey") {
			t.Errorf("warning should mention scheme names, got: %q", w)
		}
	})

	t.Run("no_security_schemes_no_warnings", func(t *testing.T) {
		tmp := t.TempDir() + "/no_sec.json"
		_ = writeTemp(tmp, `{"openapi":"3.0.0","info":{"title":"T"},"paths":{}}`)
		result, err := imp.Import(tmp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Warnings) != 0 {
			t.Errorf("expected no warnings, got: %v", result.Warnings)
		}
	})
}

// ── IDs and timestamps ────────────────────────────────────────────────────────

func TestOpenAPIImporter_RequestMetadata(t *testing.T) {
	imp := NewOpenAPIImporter()
	result, err := imp.Import(filepath.Join(testdataDir, "openapi_3_0.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	allRequests := append([]collection.Request{}, result.Collection.Requests...)
	for _, folder := range result.Collection.Folders {
		allRequests = append(allRequests, folder.Requests...)
	}
	for i, req := range allRequests {
		if req.Id == "" {
			t.Errorf("request %d has empty ID", i)
		}
		if req.CreationTimestamp.IsZero() {
			t.Errorf("request %d has zero CreationTimestamp", i)
		}
		if req.LastUpdateTimestamp.IsZero() {
			t.Errorf("request %d has zero LastUpdateTimestamp", i)
		}
	}
	if result.Collection.Id == "" {
		t.Error("collection has empty ID")
	}

	// All request IDs must be unique
	seen := make(map[string]bool)
	for _, req := range allRequests {
		if seen[req.Id] {
			t.Errorf("duplicate request ID: %s", req.Id)
		}
		seen[req.Id] = true
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func writeTemp(path, content string) error {
	return os.WriteFile(path, []byte(content), 0600)
}
