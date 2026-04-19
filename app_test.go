// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"solo/internal/collection"
	"solo/internal/configuration"
	"solo/internal/environment"
	"solo/internal/tools"
	"testing"
	"time"
)

func TestApp_ConfigurationIntegration(t *testing.T) {
	// 1. Setup Env Isolation
	tempHome, err := os.MkdirTemp("", "solo_app_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempHome)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	// 2. Initialize App
	app := NewApp()

	// Disable update checks in this integration test to avoid background
	// network/event side-effects unrelated to request timeout behavior.
	initialCfg, err := app.GetConfiguration()
	if err != nil {
		t.Fatalf("Failed to get initial config: %v", err)
	}
	initialCfg.General.CheckForUpdates = false
	if err := app.UpdateConfiguration(initialCfg); err != nil {
		t.Fatalf("Failed to disable update checks for test: %v", err)
	}

	// Mock startup context if needed, but here we just test logic.
	app.startup(context.TODO())

	// 3. Create a slow mock server
	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer slowServer.Close()

	// 4. Test with Default Timeout (should pass, default is usually 30s)
	// First ensure default is loaded
	cfg, err := app.GetConfiguration()
	if err != nil {
		t.Fatalf("Failed to get config: %v", err)
	}
	if cfg.Request.TimeoutSeconds != tools.DEFAULT_TIMEOUT_SECONDS {
		t.Errorf("Expected default timeout %d, got %d", tools.DEFAULT_TIMEOUT_SECONDS, cfg.Request.TimeoutSeconds)
	}

	opts := RequestOptions{
		Method: "GET",
		URL:    slowServer.URL,
	}

	resp, err := app.Execute(opts)
	if err != nil {
		t.Fatalf("Request with default timeout failed unexpected: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	// 5. Update Configuration to very short timeout (e.g., 0s which might mean no timeout, so use 1ms via custom logic?
	// Wait, our logic is: client.Timeout = time.Duration(cfg.Request.TimeoutSeconds) * time.Second.
	// Since cfg uses int seconds, minimum non-zero is 1 second.
	// 1 second is longer than 100ms sleep. So we need the server to sleep longer than 1s.

	// Let's create a slower handler for the timeout test
	verySlowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer verySlowServer.Close()

	// Set timeout to 1 second
	cfg.Request.TimeoutSeconds = 1
	err = app.UpdateConfiguration(cfg)
	if err != nil {
		t.Fatalf("Failed to update config: %v", err)
	}

	// 6. Test with Short Timeout (should fail)
	opts.URL = verySlowServer.URL
	_, err = app.Execute(opts)

	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}

	// Verify it's actually a timeout error (Client.Timeout returns "Client.Timeout exceeded...")
	// Specific error checking might be brittle, but ensuring an error occurred is good enough for now.
	t.Logf("Got expected error: %v", err)

	// 7. Test Override Logic
	// Reset global config to safe defaults (30s timeout)
	cfg.Request.TimeoutSeconds = 30
	if err := app.UpdateConfiguration(cfg); err != nil {
		t.Fatalf("Failed to reset config: %v", err)
	}

	// Create request with 1s override
	overrideTimeout := 1
	optsOverride := RequestOptions{
		Method: "GET",
		URL:    verySlowServer.URL,
		Settings: &configuration.RequestSettingsOverride{
			TimeoutSeconds: &overrideTimeout,
		},
	}

	// Should fail due to 1s local timeout < 2s server delay
	_, err = app.Execute(optsOverride)
	if err == nil {
		t.Fatal("Expected override timeout error, got nil")
	}
	t.Logf("Got expected override error: %v", err)
}

func TestApp_Execute_CollectionVariablesFallbackAndEnvPrecedence(t *testing.T) {
	tempHome, err := os.MkdirTemp("", "solo_app_collection_vars")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempHome)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	app := NewApp()
	app.startup(context.TODO())

	if err := app.CreateCollection("Orders"); err != nil {
		t.Fatalf("failed to create collection: %v", err)
	}

	coll, err := app.LoadCollection("Orders")
	if err != nil {
		t.Fatalf("failed to load collection: %v", err)
	}

	collectionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/orders" {
			t.Errorf("collection fallback path = %q, want %q", got, "/orders")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer collectionServer.Close()

	envServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/orders" {
			t.Errorf("env precedence path = %q, want %q", got, "/orders")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer envServer.Close()

	coll.Variables = map[string]collection.ValueType{
		"baseUrl": {Value: collectionServer.URL, Type: "text"},
	}
	if err := app.UpdateCollection(*coll); err != nil {
		t.Fatalf("failed to update collection variables: %v", err)
	}

	env := environment.NewEnvironment("dev")
	env.Values["baseUrl"] = environment.ValueType{Value: "", Type: "text"}
	if err := app.UpdateEnvironment(env); err != nil {
		t.Fatalf("failed to create environment: %v", err)
	}
	if err := app.SetSelectedEnvironment("dev"); err != nil {
		t.Fatalf("failed to select environment: %v", err)
	}

	resp, err := app.Execute(RequestOptions{
		Method:         "GET",
		URL:            "{{baseUrl}}/orders",
		CollectionName: "Orders",
	})
	if err != nil {
		t.Fatalf("collection fallback execute failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("collection fallback status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	env.Values["baseUrl"] = environment.ValueType{Value: envServer.URL, Type: "text"}
	if err := app.UpdateEnvironment(env); err != nil {
		t.Fatalf("failed to update environment override: %v", err)
	}

	resp, err = app.Execute(RequestOptions{
		Method:         "GET",
		URL:            "{{baseUrl}}/orders",
		CollectionName: "Orders",
	})
	if err != nil {
		t.Fatalf("env precedence execute failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("env precedence status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestApp_LoadGitBackedCollection_PostmanWithFolders(t *testing.T) {
	app := NewApp()

	coll, err := app.loadGitBackedCollection(".", filepath.Join("test", "testdata", "postman_v2_1.json"))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if coll.Name != "Test Collection" {
		t.Fatalf("Expected collection name 'Test Collection', got '%s'", coll.Name)
	}
	if len(coll.Requests) != 1 {
		t.Fatalf("Expected 1 root request, got %d", len(coll.Requests))
	}
	if len(coll.Folders) != 1 {
		t.Fatalf("Expected 1 root folder, got %d", len(coll.Folders))
	}
	if coll.Folders[0].Name != "Folder A" {
		t.Fatalf("Expected folder 'Folder A', got '%s'", coll.Folders[0].Name)
	}
}

func TestApp_LoadGitBackedCollection_OpenAPIJSON(t *testing.T) {
	app := NewApp()

	coll, err := app.loadGitBackedCollection(".", filepath.Join("test", "testdata", "openapi_3_0.json"))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if coll.Name != "OpenAPI Test Collection" {
		t.Fatalf("Expected collection name 'OpenAPI Test Collection', got '%s'", coll.Name)
	}
	if len(coll.Folders) != 1 || coll.Folders[0].Name != "users" {
		t.Fatalf("Expected OpenAPI tag folder 'users', got %+v", coll.Folders)
	}
}

func TestApp_LoadGitBackedCollection_SwaggerYAML(t *testing.T) {
	app := NewApp()

	coll, err := app.loadGitBackedCollection(".", filepath.Join("test", "testdata", "swagger_2_0.yaml"))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if coll.Name != "Swagger Test Collection" {
		t.Fatalf("Expected collection name 'Swagger Test Collection', got '%s'", coll.Name)
	}
	if len(coll.Folders) != 1 || coll.Folders[0].Name != "users" {
		t.Fatalf("Expected Swagger tag folder 'users', got %+v", coll.Folders)
	}
}

func TestApp_LoadGitBackedCollection_BrunoDirectoryWithFolders(t *testing.T) {
	app := NewApp()

	coll, err := app.loadGitBackedCollection(".", filepath.Join("test", "testdata", "bruno_collection"))
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if coll.Name != "Bruno Test Collection" {
		t.Fatalf("Expected collection name 'Bruno Test Collection', got '%s'", coll.Name)
	}
	if len(coll.Requests) != 0 {
		t.Fatalf("Expected 0 root requests, got %d", len(coll.Requests))
	}
	if len(coll.Folders) != 1 {
		t.Fatalf("Expected 1 root folder, got %d", len(coll.Folders))
	}
	if coll.Folders[0].Name != "users" {
		t.Fatalf("Expected folder 'users', got '%s'", coll.Folders[0].Name)
	}
}

func TestApp_LoadGitBackedCollection_BrunoRootDotPath(t *testing.T) {
	app := NewApp()

	coll, err := app.loadGitBackedCollection(filepath.Join("test", "testdata", "bruno_collection"), ".")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if coll.Name != "Bruno Test Collection" {
		t.Fatalf("Expected collection name 'Bruno Test Collection', got '%s'", coll.Name)
	}
	if len(coll.Folders) != 1 || coll.Folders[0].Name != "users" {
		t.Fatalf("Expected Bruno root import to resolve folder 'users', got %+v", coll.Folders)
	}
}

func TestApp_LoadGitBackedCollection_SoloNativePreservesFolders(t *testing.T) {
	tempDir := t.TempDir()
	app := NewApp()

	nativeCollection := collection.NewCollection("Native Git Collection")
	folder := collection.NewFolder("apis")
	req := collection.Request{
		Id:                  "req-1",
		Name:                "Get APIs",
		Url:                 "https://example.com/apis",
		Verb:                "GET",
		Headers:             map[string]string{},
		Cookies:             map[string]string{},
		CreationTimestamp:   time.Now(),
		LastUpdateTimestamp: time.Now(),
	}
	folder.Requests = append(folder.Requests, req)
	nativeCollection.Folders = append(nativeCollection.Folders, folder)

	data, err := json.Marshal(nativeCollection)
	if err != nil {
		t.Fatalf("Failed to marshal native collection: %v", err)
	}

	filePath := filepath.Join(tempDir, "collection.json")
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		t.Fatalf("Failed to write native collection file: %v", err)
	}

	coll, err := app.loadGitBackedCollection(tempDir, "collection.json")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(coll.Folders) != 1 {
		t.Fatalf("Expected 1 folder, got %d", len(coll.Folders))
	}
	if len(coll.Folders[0].Requests) != 1 {
		t.Fatalf("Expected 1 request in folder, got %d", len(coll.Folders[0].Requests))
	}
	if coll.Folders[0].Requests[0].Name != "Get APIs" {
		t.Fatalf("Expected request 'Get APIs', got '%s'", coll.Folders[0].Requests[0].Name)
	}
}

func TestApp_LoadGitBackedCollection_SoloNativeWithoutID(t *testing.T) {
	tempDir := t.TempDir()
	app := NewApp()

	payload := map[string]any{
		"name":                "Solo Without ID",
		"requests":            []any{},
		"folders":             []any{},
		"creationTimestamp":   time.Now(),
		"lastUpdateTimestamp": time.Now(),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal native collection payload: %v", err)
	}

	filePath := filepath.Join(tempDir, "collection.json")
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		t.Fatalf("Failed to write native collection file: %v", err)
	}

	coll, err := app.loadGitBackedCollection(tempDir, "collection.json")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if coll.Name != "Solo Without ID" {
		t.Fatalf("Expected collection name 'Solo Without ID', got '%s'", coll.Name)
	}
	if len(coll.Requests) != 0 {
		t.Fatalf("Expected 0 root requests, got %d", len(coll.Requests))
	}
	if len(coll.Folders) != 0 {
		t.Fatalf("Expected 0 folders, got %d", len(coll.Folders))
	}
}
