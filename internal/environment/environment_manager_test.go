// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package environment

import (
	"encoding/json"
	"os"
	"path/filepath"
	"solo/internal/configuration"
	"solo/internal/testutil"
	"testing"
)

func TestCreateEnvironment(t *testing.T) {
	tests := []struct {
		name            string
		environmentName string
		setupExisting   bool
		expectError     bool
		errorMsg        string
	}{
		{
			name:            "Create new environment",
			environmentName: "test-environment",
			setupExisting:   false,
			expectError:     false,
		},
		{
			name:            "Create environment with empty name",
			environmentName: "",
			setupExisting:   false,
			expectError:     true,
			errorMsg:        "no environment name specified",
		},
		{
			name:            "Create duplicate environment",
			environmentName: "duplicate-environment",
			setupExisting:   true,
			expectError:     true,
			errorMsg:        "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			em := setupTestManager(t)

			if tt.setupExisting {
				if err := em.CreateEnvironment(tt.environmentName); err != nil {
					t.Fatalf("Failed to setup existing environment: %v", err)
				}
			}

			err := em.CreateEnvironment(tt.environmentName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Verify file exists and is readable
				data, readErr := os.ReadFile(em.buildEnvironmentFileName(tt.environmentName))
				if readErr != nil {
					t.Errorf("Failed to read created environment: %v", readErr)
				}

				var env Environment
				if unmarshalErr := json.Unmarshal(data, &env); unmarshalErr != nil {
					t.Errorf("Failed to unmarshal environment: %v", unmarshalErr)
				}

				if env.Name != tt.environmentName {
					t.Errorf("Expected environment name %s, got %s", tt.environmentName, env.Name)
				}
			}
		})
	}
}

func TestLoadEnvironment(t *testing.T) {
	tests := []struct {
		name            string
		environmentName string
		setupFunc       func(*EnvironmentManager, string) error
		expectError     bool
	}{
		{
			name:            "Load existing environment",
			environmentName: "existing-environment",
			setupFunc: func(em *EnvironmentManager, name string) error {
				return em.CreateEnvironment(name)
			},
			expectError: false,
		},
		{
			name:            "Load non-existent environment",
			environmentName: "non-existent",
			setupFunc:       nil,
			expectError:     true,
		},
		{
			name:            "Load with empty name",
			environmentName: "",
			setupFunc:       nil,
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			em := setupTestManager(t)

			if tt.setupFunc != nil {
				if err := tt.setupFunc(em, tt.environmentName); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			env, err := em.LoadEnvironment(tt.environmentName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if env.Name != tt.environmentName {
					t.Errorf("Expected environment name %s, got %s", tt.environmentName, env.Name)
				}
			}
		})
	}
}

func TestLoadEnvironments(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(*EnvironmentManager) error
		expectedNum int
		expectError bool
	}{
		{
			name: "Load multiple environments",
			setupFunc: func(em *EnvironmentManager) error {
				if err := em.CreateEnvironment("env1"); err != nil {
					return err
				}
				if err := em.CreateEnvironment("env2"); err != nil {
					return err
				}
				return em.CreateEnvironment("env3")
			},
			expectedNum: 3,
			expectError: false,
		},
		{
			name:        "Load from empty directory",
			setupFunc:   nil,
			expectedNum: 0,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			em := setupTestManager(t)

			if tt.setupFunc != nil {
				if err := tt.setupFunc(em); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			envs, err := em.LoadEnvironments()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(*envs) != tt.expectedNum {
					t.Errorf("Expected %d environments, got %d", tt.expectedNum, len(*envs))
				}
			}
		})
	}
}

func TestUpdateEnvironment(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*EnvironmentManager) *Environment
		expectError bool
	}{
		{
			name: "Update existing environment",
			setup: func(em *EnvironmentManager) *Environment {
				env := NewEnvironment("update-test")
				em.CreateEnvironment(env.Name)
				env.AddValue("key1", ValueType{Value: "value1", Type: "string"})
				return &env
			},
			expectError: false,
		},
		{
			name: "Update with empty name",
			setup: func(em *EnvironmentManager) *Environment {
				return &Environment{Name: ""}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			em := setupTestManager(t)

			env := tt.setup(em)
			err := em.UpdateEnvironment(env)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Verify update was persisted
				loaded, loadErr := em.LoadEnvironment(env.Name)
				if loadErr != nil {
					t.Errorf("Failed to load updated environment: %v", loadErr)
				}
				if len(loaded.Values) != len(env.Values) {
					t.Errorf("Expected %d values, got %d", len(env.Values), len(loaded.Values))
				}
			}
		})
	}
}

func TestDeleteEnvironment(t *testing.T) {
	tests := []struct {
		name            string
		environmentName string
		setupExisting   bool
		expectError     bool
	}{
		{
			name:            "Delete existing environment",
			environmentName: "to-delete",
			setupExisting:   true,
			expectError:     false,
		},
		{
			name:            "Delete non-existent environment",
			environmentName: "non-existent",
			setupExisting:   false,
			expectError:     true,
		},
		{
			name:            "Delete with empty name",
			environmentName: "",
			setupExisting:   false,
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			em := setupTestManager(t)

			if tt.setupExisting {
				if err := em.CreateEnvironment(tt.environmentName); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			err := em.DeleteEnvironment(tt.environmentName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Verify file is deleted
				if _, statErr := os.Stat(em.buildEnvironmentFileName(tt.environmentName)); !os.IsNotExist(statErr) {
					t.Errorf("Environment file should not exist after deletion")
				}
			}
		})
	}
}

func TestGetValuesFromManager(t *testing.T) {
	tests := []struct {
		name            string
		environmentName string
		setupFunc       func(*EnvironmentManager, string) error
		expectedNum     int
		expectError     bool
	}{
		{
			name:            "Get values from environment with values",
			environmentName: "test-env",
			setupFunc: func(em *EnvironmentManager, name string) error {
				if err := em.CreateEnvironment(name); err != nil {
					return err
				}
				em.AddValue(name, "key1", ValueType{Value: "value1", Type: "string"})
				return em.AddValue(name, "key2", ValueType{Value: "value2", Type: "string"})
			},
			expectedNum: 2,
			expectError: false,
		},
		{
			name:            "Get values from empty environment",
			environmentName: "empty-env",
			setupFunc: func(em *EnvironmentManager, name string) error {
				return em.CreateEnvironment(name)
			},
			expectedNum: 0,
			expectError: false,
		},
		{
			name:            "Get values with empty name",
			environmentName: "",
			setupFunc:       nil,
			expectedNum:     0,
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			em := setupTestManager(t)

			if tt.setupFunc != nil {
				if err := tt.setupFunc(em, tt.environmentName); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			values, err := em.GetValues(tt.environmentName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if len(*values) != tt.expectedNum {
					t.Errorf("Expected %d values, got %d", tt.expectedNum, len(*values))
				}
			}
		})
	}
}

func TestEnvironmentManagerAddValue(t *testing.T) {
	tests := []struct {
		name            string
		environmentName string
		valueName       string
		value           ValueType
		expectError     bool
	}{
		{
			name:            "Add value to existing environment",
			environmentName: "test-environment",
			valueName:       "key1",
			value:           ValueType{Value: "value1", Type: "string"},
			expectError:     false,
		},
		{
			name:            "Add value with empty environment name",
			environmentName: "",
			valueName:       "key1",
			value:           ValueType{Value: "value1", Type: "string"},
			expectError:     true,
		},
		{
			name:            "Add value with empty value name",
			environmentName: "test-environment",
			valueName:       "",
			value:           ValueType{Value: "value1", Type: "string"},
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			em := setupTestManager(t)

			if tt.environmentName != "" {
				em.CreateEnvironment(tt.environmentName)
			}

			err := em.AddValue(tt.environmentName, tt.valueName, tt.value)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Verify value was added
				env, _ := em.LoadEnvironment(tt.environmentName)
				if len(env.Values) != 1 {
					t.Errorf("Expected 1 value, got %d", len(env.Values))
				}
				if val, ok := env.Values[tt.valueName]; !ok {
					t.Errorf("Value %s not found in environment", tt.valueName)
				} else if val.Value != tt.value.Value {
					t.Errorf("Expected value %s, got %s", tt.value.Value, val.Value)
				}
			}
		})
	}
}

func TestEnvironmentManagerRemoveValue(t *testing.T) {
	tests := []struct {
		name            string
		environmentName string
		valueName       string
		setupFunc       func(*EnvironmentManager, string, string) error
		expectError     bool
	}{
		{
			name:            "Remove existing value",
			environmentName: "test-environment",
			valueName:       "key1",
			setupFunc: func(em *EnvironmentManager, envName, valName string) error {
				if err := em.CreateEnvironment(envName); err != nil {
					return err
				}
				return em.AddValue(envName, valName, ValueType{Value: "value1", Type: "string"})
			},
			expectError: false,
		},
		{
			name:            "Remove with empty environment name",
			environmentName: "",
			valueName:       "key1",
			setupFunc:       nil,
			expectError:     true,
		},
		{
			name:            "Remove non-existent value",
			environmentName: "test-environment",
			valueName:       "non-existent",
			setupFunc: func(em *EnvironmentManager, envName, valName string) error {
				return em.CreateEnvironment(envName)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			em := setupTestManager(t)

			if tt.setupFunc != nil {
				if err := tt.setupFunc(em, tt.environmentName, tt.valueName); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			err := em.RemoveValue(tt.environmentName, tt.valueName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Verify value was removed
				env, _ := em.LoadEnvironment(tt.environmentName)
				if _, ok := env.Values[tt.valueName]; ok {
					t.Errorf("Value %s should not exist after removal", tt.valueName)
				}
			}
		})
	}
}

func TestEnvironmentManagerUpdateValue(t *testing.T) {
	tests := []struct {
		name            string
		environmentName string
		valueName       string
		setupFunc       func(*EnvironmentManager, string, string) error
		updatedValue    ValueType
		expectError     bool
	}{
		{
			name:            "Update existing value",
			environmentName: "test-environment",
			valueName:       "key1",
			setupFunc: func(em *EnvironmentManager, envName, valName string) error {
				if err := em.CreateEnvironment(envName); err != nil {
					return err
				}
				return em.AddValue(envName, valName, ValueType{Value: "original", Type: "string"})
			},
			updatedValue: ValueType{Value: "updated", Type: "string"},
			expectError:  false,
		},
		{
			name:            "Update with empty environment name",
			environmentName: "",
			valueName:       "key1",
			setupFunc:       nil,
			updatedValue:    ValueType{Value: "updated", Type: "string"},
			expectError:     true,
		},
		{
			name:            "Update with empty value name",
			environmentName: "test-environment",
			valueName:       "",
			setupFunc: func(em *EnvironmentManager, envName, valName string) error {
				return em.CreateEnvironment(envName)
			},
			updatedValue: ValueType{Value: "updated", Type: "string"},
			expectError:  true,
		},
		{
			name:            "Update non-existent value",
			environmentName: "test-environment",
			valueName:       "non-existent",
			setupFunc: func(em *EnvironmentManager, envName, valName string) error {
				return em.CreateEnvironment(envName)
			},
			updatedValue: ValueType{Value: "updated", Type: "string"},
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			em := setupTestManager(t)

			if tt.setupFunc != nil {
				if err := tt.setupFunc(em, tt.environmentName, tt.valueName); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			err := em.UpdateValue(tt.environmentName, tt.valueName, tt.updatedValue)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Verify value was updated
				env, _ := em.LoadEnvironment(tt.environmentName)
				if val, ok := env.Values[tt.valueName]; !ok {
					t.Errorf("Value %s not found in environment", tt.valueName)
				} else if val.Value != tt.updatedValue.Value {
					t.Errorf("Expected value %s, got %s", tt.updatedValue.Value, val.Value)
				}
			}
		})
	}
}

// Helper functions
func setupTestManager(t *testing.T) *EnvironmentManager {
	tempHome := testutil.IsolateUserConfigDir(t)

	cm, err := configuration.NewConfigurationManager()
	if err != nil {
		t.Fatalf("Failed to create configuration manager: %v", err)
	}

	// EnvironmentManager needs a real directory in the temp home
	envDir := filepath.Join(tempHome, "environments")
	if err := os.MkdirAll(envDir, 0700); err != nil {
		t.Fatalf("Failed to create environments directory: %v", err)
	}

	em := &EnvironmentManager{path: envDir, cm: cm}

	return em
}

func setupTestEnvironmentConfig(t *testing.T) string {
	return testutil.IsolateUserConfigDir(t)
}

func TestNewEnvironmentManager(t *testing.T) {
	setupTestEnvironmentConfig(t)

	cm, err := configuration.NewConfigurationManager()
	if err != nil {
		t.Fatalf("Failed to create configuration manager: %v", err)
	}

	// Test case 1: "default" environment is created if it doesn't exist
	em := NewEnvironmentManager(cm)
	if em == nil {
		t.Fatal("Expected NewEnvironmentManager to return a manager, got nil")
	}

	// Verify "default" environment file exists
	defaultEnvPath := filepath.Join(em.path, DEFAULT_ENVIRONMENT_NAME+".json")
	if _, err := os.Stat(defaultEnvPath); os.IsNotExist(err) {
		t.Errorf("Expected default environment file to be created at %s", defaultEnvPath)
	}

	// Verify "default" environment is selected if none was selected
	if cm.GetSelectedEnvironment() != DEFAULT_ENVIRONMENT_NAME {
		t.Errorf("Expected selected environment to be %s, got %s", DEFAULT_ENVIRONMENT_NAME, cm.GetSelectedEnvironment())
	}

	// Test case 2: "default" environment is not overwritten if it already exists
	// First, add a value to the default environment
	err = em.AddValue(DEFAULT_ENVIRONMENT_NAME, "test-key", ValueType{Value: "test-value", Type: "string"})
	if err != nil {
		t.Fatalf("Failed to add value to default environment: %v", err)
	}

	// Re-initialize the manager
	em2 := NewEnvironmentManager(cm)

	// Verify the value still exists
	env, err := em2.LoadEnvironment(DEFAULT_ENVIRONMENT_NAME)
	if err != nil {
		t.Fatalf("Failed to load default environment: %v", err)
	}
	if _, ok := env.Values["test-key"]; !ok {
		t.Error("Expected test-key to persist in default environment after re-initialization")
	}

	// Test case 3: Selected environment is NOT changed if it's already set and exists
	otherEnv := "other-env"
	err = em2.CreateEnvironment(otherEnv)
	if err != nil {
		t.Fatalf("Failed to create other environment: %v", err)
	}
	cm.SetSelectedEnvironment(otherEnv)

	_ = NewEnvironmentManager(cm)
	if cm.GetSelectedEnvironment() != otherEnv {
		t.Errorf("Expected selected environment to remain %s, got %s", otherEnv, cm.GetSelectedEnvironment())
	}

	// Test case 4: Selected environment is reset to "default" if it's set but does not exist
	cm.SetSelectedEnvironment("non-existent-env")
	_ = NewEnvironmentManager(cm)
	if cm.GetSelectedEnvironment() != DEFAULT_ENVIRONMENT_NAME {
		t.Errorf("Expected selected environment to be reset to %s when missing, got %s", DEFAULT_ENVIRONMENT_NAME, cm.GetSelectedEnvironment())
	}
}

func TestDeleteEnvironment_DefaultProtectionAndFallback(t *testing.T) {
	setupTestEnvironmentConfig(t)

	cm, err := configuration.NewConfigurationManager()
	if err != nil {
		t.Fatalf("Failed to create configuration manager: %v", err)
	}

	em := NewEnvironmentManager(cm)

	// Test case 1: Cannot delete "default" environment
	err = em.DeleteEnvironment(DEFAULT_ENVIRONMENT_NAME)
	if err == nil {
		t.Error("Expected error when deleting default environment, got nil")
	} else if err.Error() != "default environment cannot be deleted" {
		t.Errorf("Expected error 'default environment cannot be deleted', got '%v'", err)
	}

	// Test case 2: Deleting selected environment falls back to "default"
	otherEnv := "to-delete"
	err = em.CreateEnvironment(otherEnv)
	if err != nil {
		t.Fatalf("Failed to create environment: %v", err)
	}

	cm.SetSelectedEnvironment(otherEnv)
	if cm.GetSelectedEnvironment() != otherEnv {
		t.Fatalf("Failed to set selected environment")
	}

	err = em.DeleteEnvironment(otherEnv)
	if err != nil {
		t.Fatalf("Failed to delete environment: %v", err)
	}

	if cm.GetSelectedEnvironment() != DEFAULT_ENVIRONMENT_NAME {
		t.Errorf("Expected selected environment to fall back to %s after deletion, got %s", DEFAULT_ENVIRONMENT_NAME, cm.GetSelectedEnvironment())
	}

	// Test case 3: Deleting non-selected environment does NOT change selection
	env1 := "env1"
	env2 := "env2"
	em.CreateEnvironment(env1)
	em.CreateEnvironment(env2)
	cm.SetSelectedEnvironment(env1)

	err = em.DeleteEnvironment(env2)
	if err != nil {
		t.Fatalf("Failed to delete environment: %v", err)
	}

	if cm.GetSelectedEnvironment() != env1 {
		t.Errorf("Selection changed unexpectedly. Expected %s, got %s", env1, cm.GetSelectedEnvironment())
	}
}

func TestRenameEnvironment(t *testing.T) {
	tests := []struct {
		name          string
		oldName       string
		newName       string
		setupExisting bool
		isSelected    bool
		expectError   bool
	}{
		{
			name:          "Rename existing environment",
			oldName:       "old-name",
			newName:       "new-name",
			setupExisting: true,
			expectError:   false,
		},
		{
			name:          "Rename non-existent environment",
			oldName:       "non-existent",
			newName:       "some-name",
			setupExisting: false,
			expectError:   true,
		},
		{
			name:          "Rename with empty new name",
			oldName:       "old-name",
			newName:       "",
			setupExisting: true,
			expectError:   true,
		},
		{
			name:          "Rename default environment (protected)",
			oldName:       DEFAULT_ENVIRONMENT_NAME,
			newName:       "new-default",
			setupExisting: true,
			expectError:   true,
		},
		{
			name:          "Rename to existing name (conflict)",
			oldName:       "env1",
			newName:       "env2",
			setupExisting: true,
			expectError:   true,
		},
		{
			name:          "Rename selected environment updates selection",
			oldName:       "selected-env",
			newName:       "renamed-env",
			setupExisting: true,
			isSelected:    true,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			em := setupTestManager(t)

			if tt.setupExisting {
				if tt.oldName != DEFAULT_ENVIRONMENT_NAME {
					if err := em.CreateEnvironment(tt.oldName); err != nil {
						t.Fatalf("Setup failed: %v", err)
					}
				}
				if tt.name == "Rename to existing name (conflict)" {
					if err := em.CreateEnvironment(tt.newName); err != nil {
						t.Fatalf("Setup failed: %v", err)
					}
				}
			}

			if tt.isSelected {
				em.cm.SetSelectedEnvironment(tt.oldName)
			}

			err := em.RenameEnvironment(tt.oldName, tt.newName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Verify new file exists
				if _, statErr := os.Stat(em.buildEnvironmentFileName(tt.newName)); os.IsNotExist(statErr) {
					t.Errorf("New environment file should exist")
				}

				// Verify old file is gone
				if _, statErr := os.Stat(em.buildEnvironmentFileName(tt.oldName)); !os.IsNotExist(statErr) {
					t.Errorf("Old environment file should not exist")
				}

				// Verify internal name was updated
				loaded, _ := em.LoadEnvironment(tt.newName)
				if loaded.Name != tt.newName {
					t.Errorf("Expected internal name %s, got %s", tt.newName, loaded.Name)
				}

				// Verify selection update
				if tt.isSelected && em.cm.GetSelectedEnvironment() != tt.newName {
					t.Errorf("Expected selected environment to be %s, got %s", tt.newName, em.cm.GetSelectedEnvironment())
				}
			}
		})
	}
}
