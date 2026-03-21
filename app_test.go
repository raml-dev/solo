package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"solo/internal/configuration"
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
	// Mock startup context if needed, but here we just test logic
	app.startup(context.Background())

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
