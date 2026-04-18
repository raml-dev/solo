// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package importer

import (
	"os"
	"path/filepath"
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

	if len(coll.Requests) != 0 {
		t.Fatalf("Expected 0 root requests, got %d", len(coll.Requests))
	}

	if len(coll.Folders) != 1 {
		t.Fatalf("Expected 1 root folder, got %d", len(coll.Folders))
	}

	usersFolder := coll.Folders[0]
	if usersFolder.Name != "users" {
		t.Fatalf("Expected folder name 'users', got '%s'", usersFolder.Name)
	}
	if len(usersFolder.Requests) != 2 {
		t.Fatalf("Expected folder 'users' to contain 2 requests, got %d", len(usersFolder.Requests))
	}
	if len(usersFolder.Folders) != 0 {
		t.Fatalf("Expected folder 'users' to contain no subfolders, got %d", len(usersFolder.Folders))
	}

	req1 := usersFolder.Requests[0]
	req2 := usersFolder.Requests[1]

	if req1.Name == "Create User" && req2.Name == "Get All Users" {
		req1, req2 = req2, req1
	}

	// 3. Find and test Get All Users request
	if req1.Name != "Get All Users" {
		t.Errorf("Expected request name 'Get All Users', got '%s'", req1.Name)
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
	if req1.Body != "" {
		t.Errorf("Expected empty body, got '%s'", req1.Body)
	}

	// 4. Find and test Create User request
	if req2.Name != "Create User" {
		t.Errorf("Expected request name 'Create User', got '%s'", req2.Name)
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
	if usersFolder.Id == "" || usersFolder.CreationTimestamp.IsZero() {
		t.Errorf("Expected folder 'users' to have ID and creation timestamp")
	}
	if req1.Id == "" || req1.CreationTimestamp.IsZero() {
		t.Errorf("Expected request 'Get All Users' to have ID and creation timestamp")
	}
	if req2.Id == "" || req2.CreationTimestamp.IsZero() {
		t.Errorf("Expected request 'Create User' to have ID and creation timestamp")
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
