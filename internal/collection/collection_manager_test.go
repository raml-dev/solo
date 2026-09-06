// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package collection

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateCollection(t *testing.T) {
	tests := []struct {
		name           string
		collectionName string
		setupExisting  bool
		expectError    bool
		errorMsg       string
	}{
		{
			name:           "Create new collection",
			collectionName: "test-collection",
			setupExisting:  false,
			expectError:    false,
		},
		{
			name:           "Create collection with empty name",
			collectionName: "",
			setupExisting:  false,
			expectError:    true,
			errorMsg:       "no collection name specified",
		},
		{
			name:           "Create duplicate collection",
			collectionName: "duplicate-collection",
			setupExisting:  true,
			expectError:    true,
			errorMsg:       "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := setupTestManager(t)
			defer cleanupTestDir(cm.config)

			if tt.setupExisting {
				if err := cm.CreateCollection(tt.collectionName); err != nil {
					t.Fatalf("Failed to setup existing collection: %v", err)
				}
			}

			err := cm.CreateCollection(tt.collectionName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Verify file exists and is readable
				data, readErr := os.ReadFile(buildCollectionFileName(cm.config, tt.collectionName))
				if readErr != nil {
					t.Errorf("Failed to read created collection: %v", readErr)
				}

				var coll Collection
				if unmarshalErr := json.Unmarshal(data, &coll); unmarshalErr != nil {
					t.Errorf("Failed to unmarshal collection: %v", unmarshalErr)
				}

				if coll.Name != tt.collectionName {
					t.Errorf("Expected collection name %s, got %s", tt.collectionName, coll.Name)
				}
			}
		})
	}
}

func TestLoadCollection(t *testing.T) {
	tests := []struct {
		name           string
		collectionName string
		setupFunc      func(*CollectionManager, string) error
		expectError    bool
	}{
		{
			name:           "Load existing collection",
			collectionName: "existing-collection",
			setupFunc: func(cm *CollectionManager, name string) error {
				return cm.CreateCollection(name)
			},
			expectError: false,
		},
		{
			name:           "Load non-existent collection",
			collectionName: "non-existent",
			setupFunc:      nil,
			expectError:    true,
		},
		{
			name:           "Load with empty name",
			collectionName: "",
			setupFunc:      nil,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := setupTestManager(t)
			defer cleanupTestDir(cm.config)

			if tt.setupFunc != nil {
				if err := tt.setupFunc(cm, tt.collectionName); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			coll, err := cm.LoadCollection(tt.collectionName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if coll.Name != tt.collectionName {
					t.Errorf("Expected collection name %s, got %s", tt.collectionName, coll.Name)
				}
			}
		})
	}
}

func TestUpdateCollection(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*CollectionManager) Collection
		expectError bool
	}{
		{
			name: "Update existing collection",
			setup: func(cm *CollectionManager) Collection {
				coll := NewCollection("update-test")
				cm.CreateCollection(coll.Name)
				coll.AddRequest(Request{Name: "test-request"})
				return coll
			},
			expectError: false,
		},
		{
			name: "Update with empty name",
			setup: func(cm *CollectionManager) Collection {
				return Collection{Name: ""}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := setupTestManager(t)
			defer cleanupTestDir(cm.config)

			coll := tt.setup(cm)
			err := cm.UpdateCollection(coll)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Verify update was persisted
				loaded, loadErr := cm.LoadCollection(coll.Name)
				if loadErr != nil {
					t.Errorf("Failed to load updated collection: %v", loadErr)
				}
				if len(loaded.Requests) != len(coll.Requests) {
					t.Errorf("Expected %d requests, got %d", len(coll.Requests), len(loaded.Requests))
				}
			}
		})
	}
}

func TestDeleteCollection(t *testing.T) {
	tests := []struct {
		name           string
		collectionName string
		setupExisting  bool
		expectError    bool
	}{
		{
			name:           "Delete existing collection",
			collectionName: "to-delete",
			setupExisting:  true,
			expectError:    false,
		},
		{
			name:           "Delete non-existent collection",
			collectionName: "non-existent",
			setupExisting:  false,
			expectError:    true,
		},
		{
			name:           "Delete with empty name",
			collectionName: "",
			setupExisting:  false,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := setupTestManager(t)
			defer cleanupTestDir(cm.config)

			if tt.setupExisting {
				if err := cm.CreateCollection(tt.collectionName); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			err := cm.DeleteCollection(tt.collectionName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Verify file is deleted
				if _, statErr := os.Stat(buildCollectionFileName(cm.config, tt.collectionName)); !os.IsNotExist(statErr) {
					t.Errorf("Collection file should not exist after deletion")
				}
			}
		})
	}
}

func TestCollectionManagerAddRequest(t *testing.T) {
	tests := []struct {
		name           string
		collectionName string
		request        Request
		expectError    bool
	}{
		{
			name:           "Add request to existing collection",
			collectionName: "test-collection",
			request:        Request{Name: "new-request"},
			expectError:    false,
		},
		{
			name:           "Add request with empty collection name",
			collectionName: "",
			request:        Request{Name: "test"},
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := setupTestManager(t)
			defer cleanupTestDir(cm.config)

			if tt.collectionName != "" {
				cm.CreateCollection(tt.collectionName)
			}

			_, err := cm.AddRequest(tt.collectionName, tt.request)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Verify request was added
				coll, _ := cm.LoadCollection(tt.collectionName)
				if len(coll.Requests) != 1 {
					t.Errorf("Expected 1 request, got %d", len(coll.Requests))
				}
			}
		})
	}
}

func TestCollectionManagerFolderCRUD(t *testing.T) {
	cm := setupTestManager(t)
	defer cleanupTestDir(cm.config)

	const collectionName = "folders-test"
	if err := cm.CreateCollection(collectionName); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	root, err := cm.AddFolder(collectionName, "", Folder{Id: "f-root", Name: "root"})
	if err != nil {
		t.Fatalf("add folder failed: %v", err)
	}
	if root == nil || root.Id != "f-root" {
		t.Fatalf("expected folder f-root, got %+v", root)
	}

	sub, err := cm.AddFolder(collectionName, "f-root", Folder{Id: "f-sub", Name: "sub"})
	if err != nil {
		t.Fatalf("add subfolder failed: %v", err)
	}
	if sub == nil || sub.Id != "f-sub" {
		t.Fatalf("expected folder f-sub, got %+v", sub)
	}

	found, err := cm.GetFolder(collectionName, "f-sub")
	if err != nil {
		t.Fatalf("get folder failed: %v", err)
	}
	if found.Name != "sub" {
		t.Fatalf("expected name sub, got %s", found.Name)
	}

	if err := cm.UpdateFolder(collectionName, Folder{Id: "f-sub", Name: "sub-updated"}); err != nil {
		t.Fatalf("update folder failed: %v", err)
	}

	updated, err := cm.GetFolder(collectionName, "f-sub")
	if err != nil {
		t.Fatalf("get updated folder failed: %v", err)
	}
	if updated.Name != "sub-updated" {
		t.Fatalf("expected name sub-updated, got %s", updated.Name)
	}

	if err := cm.RemoveFolder(collectionName, "f-sub"); err != nil {
		t.Fatalf("remove folder failed: %v", err)
	}

	if _, err := cm.GetFolder(collectionName, "f-sub"); err == nil {
		t.Fatal("expected error for removed folder, got nil")
	}
}

func TestCollectionManagerAddRequestToFolder(t *testing.T) {
	cm := setupTestManager(t)
	defer cleanupTestDir(cm.config)

	const collectionName = "folder-request-test"
	if err := cm.CreateCollection(collectionName); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	if _, err := cm.AddFolder(collectionName, "", Folder{Id: "f-root", Name: "root"}); err != nil {
		t.Fatalf("add folder failed: %v", err)
	}

	added, err := cm.AddRequestToFolder(collectionName, "f-root", Request{Name: "inside-folder"})
	if err != nil {
		t.Fatalf("add request to folder failed: %v", err)
	}

	if added == nil {
		t.Fatal("expected request, got nil")
	}

	coll, err := cm.LoadCollection(collectionName)
	if err != nil {
		t.Fatalf("load collection failed: %v", err)
	}

	if len(coll.Folders) != 1 || len(coll.Folders[0].Requests) != 1 {
		t.Fatalf("expected 1 folder request, got %+v", coll.Folders)
	}
}

type fakeAuthSecretStore struct {
	tokens  map[string]string
	deleted map[string]bool
	nextID  int
}

func newFakeAuthSecretStore() *fakeAuthSecretStore {
	return &fakeAuthSecretStore{tokens: map[string]string{}, deleted: map[string]bool{}}
}

func (s *fakeAuthSecretStore) StoreBearerToken(tokenID, token string) (string, error) {
	if tokenID == "" {
		s.nextID++
		tokenID = fmt.Sprintf("token-%d", s.nextID)
	}
	s.tokens[tokenID] = token
	return tokenID, nil
}

func (s *fakeAuthSecretStore) GetBearerToken(tokenID string) (string, error) {
	token, ok := s.tokens[tokenID]
	if !ok {
		return "", fmt.Errorf("token %q not found", tokenID)
	}
	return token, nil
}

func (s *fakeAuthSecretStore) DeleteBearerToken(tokenID string) error {
	delete(s.tokens, tokenID)
	s.deleted[tokenID] = true
	return nil
}

func TestCollectionManager_BearerTokenPersistenceAndDuplication(t *testing.T) {
	cm := setupTestManager(t)
	defer cleanupTestDir(cm.config)
	secretStore := newFakeAuthSecretStore()
	cm.secretStore = secretStore

	const collectionName = "bearer-auth-test"
	if err := cm.CreateCollection(collectionName); err != nil {
		t.Fatalf("CreateCollection() error = %v", err)
	}

	added, err := cm.AddRequest(collectionName, Request{
		Name: "secured",
		Auth: &AuthConfiguration{Mode: AuthModeBearer, BearerToken: "plaintext-secret"},
	})
	if err != nil {
		t.Fatalf("AddRequest() error = %v", err)
	}
	if added.Auth == nil || added.Auth.BearerTokenID == "" || added.Auth.BearerToken != "" {
		t.Fatalf("persisted auth = %+v, want token identifier without plaintext", added.Auth)
	}
	originalTokenID := added.Auth.BearerTokenID

	rawCollection, err := os.ReadFile(buildCollectionFileName(cm.config, collectionName))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if strings.Contains(string(rawCollection), "plaintext-secret") {
		t.Fatal("collection JSON contains the plaintext bearer token")
	}

	placeholderRequest, err := cm.AddRequest(collectionName, Request{
		Name: "environment secured",
		Auth: &AuthConfiguration{Mode: AuthModeBearer, BearerToken: "{{apiToken}}"},
	})
	if err != nil {
		t.Fatalf("AddRequest() with placeholder error = %v", err)
	}
	if placeholderRequest.Auth == nil || placeholderRequest.Auth.BearerToken != "{{apiToken}}" {
		t.Fatalf("placeholder auth = %+v, want readable placeholder", placeholderRequest.Auth)
	}
	if placeholderRequest.Auth.BearerTokenID != "" {
		t.Fatalf("placeholder token identifier = %q, want empty", placeholderRequest.Auth.BearerTokenID)
	}

	duplicate := *added
	duplicate.Id = ""
	duplicate.Name = "secured copy"
	duplicated, err := cm.AddRequest(collectionName, duplicate)
	if err != nil {
		t.Fatalf("duplicating bearer request: %v", err)
	}
	if duplicated.Auth.BearerTokenID == originalTokenID {
		t.Fatal("duplicated request shares the original bearer token identifier")
	}
	if got := secretStore.tokens[duplicated.Auth.BearerTokenID]; got != "plaintext-secret" {
		t.Fatalf("duplicated secret = %q, want plaintext-secret", got)
	}

	if err := cm.RemoveRequest(collectionName, added.Id); err != nil {
		t.Fatalf("RemoveRequest() error = %v", err)
	}
	if !secretStore.deleted[originalTokenID] {
		t.Fatal("removing a request did not remove its bearer credential")
	}
}

// Helper functions
func setupTestManager(t *testing.T) *CollectionManager {
	tmpDir := filepath.Join(os.TempDir(), "solo-test-"+t.Name())
	if err := os.MkdirAll(tmpDir, 0700); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	return &CollectionManager{config: tmpDir}
}

func cleanupTestDir(path string) {
	os.RemoveAll(path)
}

func buildCollectionFileName(configPath, collectionName string) string {
	return fmt.Sprintf("%s/%s.json", configPath, collectionName)
}
