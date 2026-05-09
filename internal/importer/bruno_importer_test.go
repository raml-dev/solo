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

func TestBrunoImporter_Import(t *testing.T) {
	importer := NewBrunoImporter()

	// Use the testdata directory for Bruno
	testDir := filepath.Join("..", "..", "test", "testdata", "bruno_collection")

	coll, err := importer.Import(testDir)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if coll == nil {
		t.Fatalf("Expected collection to not be nil")
	}

	// 1. Verify Collection Name from bruno.json
	if coll.Name != "Bruno Test Collection" {
		t.Errorf("Expected collection name 'Bruno Test Collection', got '%s'", coll.Name)
	}

	// We have 2 tagged requests (in users/) and 3 root requests (form, multipart, query)
	totalReqs := countRequestsInCollection(coll)
	if totalReqs != 5 {
		t.Errorf("total request count: got %d, want 5", totalReqs)
	}

	// Helper to find request by name
	findReq := func(reqs []collection.Request, name string) *collection.Request {
		for i := range reqs {
			if reqs[i].Name == name {
				return &reqs[i]
			}
		}
		return nil
	}

	// Verify root requests
	// Query Test
	r := findReq(coll.Requests, "Query Test")
	if r == nil {
		t.Fatal("Query Test request not found")
	}
	if !strings.Contains(r.Url, "?limit=10&offset=0&filter=active") {
		t.Errorf("Query Test URL: got %q", r.Url)
	}

	// Form Test
	r = findReq(coll.Requests, "Form Test")
	if r == nil {
		t.Fatal("Form Test request not found")
	}
	if r.Headers["Content-Type"] != "application/x-www-form-urlencoded" {
		t.Errorf("Form Test Content-Type: got %q", r.Headers["Content-Type"])
	}
	if !strings.Contains(r.Body, "username=admin") || !strings.Contains(r.Body, "password=password123") {
		t.Errorf("Form Test body: got %q", r.Body)
	}

	// Multipart Test
	r = findReq(coll.Requests, "Multipart Test")
	if r == nil {
		t.Fatal("Multipart Test request not found")
	}
	if !strings.Contains(r.Headers["Content-Type"], "multipart/form-data") {
		t.Errorf("Multipart Test Content-Type: got %q", r.Headers["Content-Type"])
	}
	if !strings.Contains(r.Body, "name=\"title\"") || !strings.Contains(r.Body, "Project Logo") {
		t.Errorf("Multipart Test body: got %q", r.Body)
	}

	// Verify tagged requests in users folder
	var usersFolder *collection.Folder
	for i := range coll.Folders {
		if coll.Folders[i].Name == "users" {
			usersFolder = &coll.Folders[i]
			break
		}
	}
	if usersFolder == nil {
		t.Fatal("Expected folder name 'users' not found")
	}

	req1 := findReq(usersFolder.Requests, "Get All Users")
	if req1 == nil {
		t.Fatal("Request 'Get All Users' not found")
	}
	if req1.Verb != "GET" {
		t.Errorf("Expected verb 'GET', got '%s'", req1.Verb)
	}
	if req1.Url != "{{host}}/api/v1/users" {
		t.Errorf("Expected URL '{{host}}/api/v1/users', got '%s'", req1.Url)
	}
	if req1.Headers["X-Custom-Header"] != "TestValue" {
		t.Errorf("Expected 'X-Custom-Header' to be 'TestValue', got '%s'", req1.Headers["X-Custom-Header"])
	}

	req2 := findReq(usersFolder.Requests, "Create User")
	if req2 == nil {
		t.Fatal("Request 'Create User' not found")
	}
	if req2.Verb != "POST" {
		t.Errorf("Expected verb 'POST', got '%s'", req2.Verb)
	}
	if req2.BodyType != "json" {
		t.Errorf("Expected body type 'json', got '%s'", req2.BodyType)
	}
	if !strings.Contains(req2.Body, `"name": "Bruno"`) {
		t.Errorf("Expected body to contain '\"name\": \"Bruno\"', got '%s'", req2.Body)
	}
}

func TestBrunoImporter_Import_DirNotFound(t *testing.T) {
	importer := NewBrunoImporter()
	_, err := importer.Import("non_existent_bruno_dir")
	if err == nil {
		t.Fatal("Expected error for non-existent directory, got nil")
	}
}

func TestBrunoImporter_Import_IgnoresEnvironmentBruFiles(t *testing.T) {
	importer := NewBrunoImporter()
	testDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(testDir, "bruno.json"), []byte(`{"name":"GitHub API","type":"collection","version":"1"}`), 0644); err != nil {
		t.Fatalf("Failed to write bruno.json: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(testDir, "Repository"), 0755); err != nil {
		t.Fatalf("Failed to create request directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testDir, "Repository", "Repository Info.bru"), []byte(`meta {
  name: Repository Info
  type: http
}

get {
  url: {{baseUrl}}/repos/usebruno/bruno-website
  body: none
}
`), 0644); err != nil {
		t.Fatalf("Failed to write request .bru: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(testDir, "environments"), 0755); err != nil {
		t.Fatalf("Failed to create environments directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testDir, "environments", "Github.bru"), []byte(`vars {
  baseUrl: https://api.github.com
}
`), 0644); err != nil {
		t.Fatalf("Failed to write environment .bru: %v", err)
	}

	coll, err := importer.Import(testDir)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(coll.Requests) != 0 {
		t.Fatalf("Expected 0 root requests, got %d", len(coll.Requests))
	}
	if len(coll.Folders) != 1 {
		t.Fatalf("Expected 1 folder, got %d", len(coll.Folders))
	}
	if coll.Folders[0].Name != "Repository" {
		t.Fatalf("Expected folder 'Repository', got '%s'", coll.Folders[0].Name)
	}
	if len(coll.Folders[0].Requests) != 1 {
		t.Fatalf("Expected 1 request in Repository folder, got %d", len(coll.Folders[0].Requests))
	}
	if coll.Folders[0].Requests[0].Name != "Repository Info" {
		t.Fatalf("Expected request 'Repository Info', got '%s'", coll.Folders[0].Requests[0].Name)
	}
}
