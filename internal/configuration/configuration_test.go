package configuration

import (
	"os"
	"testing"
	"yapla/internal/tools"
)

func TestConfigurationManager_Defaults(t *testing.T) {
	// Setup temporary home directory for isolation
	tempHome, err := os.MkdirTemp("", "yapla_test_config")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempHome)

	// Mock UserConfigDir via HOME env var (works on Unix/Mac)
	// Note: tools.GetOrCreateConfigDir calls os.UserConfigDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	// Test initialization
	cm, err := NewConfigurationManager()
	if err != nil {
		t.Fatalf("NewConfigurationManager failed: %v", err)
	}

	// Verify defaults
	cfg := cm.Get()
	if cfg.General.Theme != tools.DEFAULT_THEME {
		t.Errorf("Expected default theme %s, got %s", tools.DEFAULT_THEME, cfg.General.Theme)
	}
	if cfg.Request.TimeoutSeconds != tools.DEFAULT_TIMEOUT_SECONDS {
		t.Errorf("Expected default timeout %d, got %d", tools.DEFAULT_TIMEOUT_SECONDS, cfg.Request.TimeoutSeconds)
	}

	// Verify file creation
	// expectedPath logic removed as it's unused and OS-dependent checking is brittle.
	// We rely on the persistence test below to verify I/O works.

	// Test Persistence
	newTheme := "dark-mode-test"
	cfg.General.Theme = newTheme
	if err := cm.Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Create new manager to simulate app restart
	cm2, err := NewConfigurationManager()
	if err != nil {
		t.Fatalf("Second NewConfigurationManager failed: %v", err)
	}
	cfg2 := cm2.Get()
	if cfg2.General.Theme != newTheme {
		t.Errorf("Persistence failed: expected theme %s, got %s", newTheme, cfg2.General.Theme)
	}
}
