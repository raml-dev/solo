// Copyright 2026-present raml-dev
// SPDX-License-Identifier: GPL-3.0-only

package importer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPostmanEnvironmentImporter(t *testing.T) {
	fixture := filepath.Join("..", "..", "test", "testdata", "postman_environment.json")
	imp := NewPostmanEnvironmentImporter()
	env, err := imp.Import(fixture)
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if env.Name != "Postman Env" {
		t.Fatalf("expected env name 'Postman Env', got %q", env.Name)
	}

	if env.Values["baseUrl"].Value != "https://api.example.com" {
		t.Fatalf("expected baseUrl value, got %q", env.Values["baseUrl"].Value)
	}

	if env.Values["port"].Value != "8080" {
		t.Fatalf("expected port value '8080', got %q", env.Values["port"].Value)
	}

	if env.Values["enabledFlag"].Value != "true" {
		t.Fatalf("expected enabledFlag value 'true', got %q", env.Values["enabledFlag"].Value)
	}

	if env.Values["secret"].Value != "supersecret" {
		t.Fatalf("expected secret value, got %q", env.Values["secret"].Value)
	}

	if _, err := os.Stat(fixture); err != nil {
		t.Fatalf("fixture missing: %v", err)
	}
}
