package importer

import (
	"path/filepath"
	"strings"
	"testing"
	"yapla/internal/collection"
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

	// 2. Verify number of requests found
	if len(coll.Requests) != 2 {
		t.Fatalf("Expected 2 requests, got %d", len(coll.Requests))
	}

	// 3. Find and test Get All Users request
	var req1 *collection.Request
	for i := range coll.Requests {
		if strings.Contains(coll.Requests[i].Name, "Get All Users") {
			req1 = &coll.Requests[i]
			break
		}
	}
	if req1 == nil {
		t.Fatalf("Request 'Get All Users' not found")
	}
	if req1.Name != "users / Get All Users" {
		t.Errorf("Expected request name 'users / Get All Users', got '%s'", req1.Name)
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
	var req2 *collection.Request
	for i := range coll.Requests {
		if strings.Contains(coll.Requests[i].Name, "Create User") {
			req2 = &coll.Requests[i]
			break
		}
	}
	if req2 == nil {
		t.Fatalf("Request 'Create User' not found")
	}
	if req2.Name != "users / Create User" {
		t.Errorf("Expected request name 'users / Create User', got '%s'", req2.Name)
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
