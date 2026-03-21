// Copyright 2026-present raml-dev
// SPDX-License-Identifier: GPL-3.0-only

package collection

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
