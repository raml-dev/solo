package environment

import (
	"encoding/json"
	"os"
	"path/filepath"
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
			defer cleanupTestDir(em.path)

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
			defer cleanupTestDir(em.path)

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
			defer cleanupTestDir(em.path)

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
			defer cleanupTestDir(em.path)

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
			defer cleanupTestDir(em.path)

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
			defer cleanupTestDir(em.path)

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
			defer cleanupTestDir(em.path)

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
			defer cleanupTestDir(em.path)

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
			defer cleanupTestDir(em.path)

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
	tmpDir := filepath.Join(os.TempDir(), "yapla-test-"+t.Name())
	if err := os.MkdirAll(tmpDir, 0700); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	return &EnvironmentManager{path: tmpDir}
}

func cleanupTestDir(path string) {
	os.RemoveAll(path)
}
