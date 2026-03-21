// Copyright 2026-present raml-dev
// SPDX-License-Identifier: GPL-3.0-only

package importer

import (
	"encoding/json"
	"path/filepath"
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

	// Verify we got 3 requests
	if len(coll.Requests) != 3 {
		t.Fatalf("Expected 3 requests, got %d", len(coll.Requests))
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

	// 2. Nested Request 1 (String URL, JSON Body)
	req2 := coll.Requests[1]
	if req2.Name != "Folder A / Nested Request 1" {
		t.Errorf("Expected second request name 'Folder A / Nested Request 1', got '%s'", req2.Name)
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

	// 3. Deeply Nested Request (Object URL, XML fallback)
	req3 := coll.Requests[2]
	if req3.Name != "Folder A / SubFolder B / Deeply Nested Request" {
		t.Errorf("Expected third request name 'Folder A / SubFolder B / Deeply Nested Request', got '%s'", req3.Name)
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

	// General checks for UUIDs and Timestamps
	for i, req := range coll.Requests {
		if req.Id == "" {
			t.Errorf("Request %d has empty ID", i)
		}
		if req.CreationTimestamp.IsZero() {
			t.Errorf("Request %d has zero creation timestamp", i)
		}
	}
}

func TestPostmanImporter_Import_FileNotFound(t *testing.T) {
	importer := NewPostmanImporter()
	_, err := importer.Import("does_not_exist.json")
	if err == nil {
		t.Fatal("Expected error for non-existent file, got nil")
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
