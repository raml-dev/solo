// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package auth

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestAuthManager_BearerTokenLifecycle(t *testing.T) {
	configDir := t.TempDir()
	manager := NewAuthManager(configDir)
	const token = "sensitive-bearer-token"

	tokenID, err := manager.StoreBearerToken("", token)
	if err != nil {
		t.Fatalf("StoreBearerToken() error = %v", err)
	}
	if tokenID == "" {
		t.Fatal("StoreBearerToken() returned an empty identifier")
	}

	stored, err := manager.GetBearerToken(tokenID)
	if err != nil {
		t.Fatalf("GetBearerToken() error = %v", err)
	}
	if stored != token {
		t.Fatalf("GetBearerToken() = %q, want %q", stored, token)
	}

	rawStore, err := os.ReadFile(filepath.Join(configDir, "auth_store.json"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if bytes.Contains(rawStore, []byte(token)) {
		t.Fatal("encrypted auth store contains the plaintext bearer token")
	}

	if err := manager.DeleteBearerToken(tokenID); err != nil {
		t.Fatalf("DeleteBearerToken() error = %v", err)
	}
	if _, err := manager.GetBearerToken(tokenID); err == nil {
		t.Fatal("GetBearerToken() succeeded after token deletion")
	}
}
