// Copyright 2026-present raml-dev
// SPDX-License-Identifier: GPL-3.0-only

package importer

import (
	"path/filepath"
	"testing"
)

func TestBrunoEnvironmentImporter(t *testing.T) {
	fixture := filepath.Join("..", "..", "test", "testdata", "bruno_environment.bru")
	imp := NewBrunoEnvironmentImporter()
	env, err := imp.Import(fixture)
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if env.Name != "bruno_environment" {
		t.Fatalf("expected env name 'bruno_environment', got %q", env.Name)
	}

	if env.Values["baseUrl"].Value != "https://api.example.com" {
		t.Fatalf("expected baseUrl value, got %q", env.Values["baseUrl"].Value)
	}

	if env.Values["apiKey"].Value != "secret123" {
		t.Fatalf("expected apiKey value, got %q", env.Values["apiKey"].Value)
	}
}
