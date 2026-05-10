// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package configuration

import (
	"os"
	"path/filepath"
	"solo/internal/testutil"
	"solo/internal/theme"
	"solo/internal/tools"
	"testing"
)

func TestConfigurationManager_Defaults(t *testing.T) {
	testutil.IsolateUserConfigDir(t)

	cm, err := NewConfigurationManager()
	if err != nil {
		t.Fatalf("NewConfigurationManager failed: %v", err)
	}

	cfg := cm.Get()
	if cfg.General.ActiveTheme != tools.DEFAULT_THEME {
		t.Errorf("Expected default theme %s, got %s", tools.DEFAULT_THEME, cfg.General.ActiveTheme)
	}
	if cfg.Request.TimeoutSeconds != tools.DEFAULT_TIMEOUT_SECONDS {
		t.Errorf("Expected default timeout %d, got %d", tools.DEFAULT_TIMEOUT_SECONDS, cfg.Request.TimeoutSeconds)
	}
	if cfg.General.IncludePrereleaseUpdates != tools.DEFAULT_INCLUDE_PRERELEASE_UPDATES {
		t.Errorf(
			"Expected prerelease toggle default %t, got %t",
			tools.DEFAULT_INCLUDE_PRERELEASE_UPDATES,
			cfg.General.IncludePrereleaseUpdates,
		)
	}
	if cfg.General.BaseFontSizePx != tools.DEFAULT_BASE_FONT_SIZE_PX {
		t.Errorf("Expected default base font size %d, got %d", tools.DEFAULT_BASE_FONT_SIZE_PX, cfg.General.BaseFontSizePx)
	}
	if cfg.General.DefaultFontFamily != "" {
		t.Errorf("Expected default sans font family to be empty, got %q", cfg.General.DefaultFontFamily)
	}
	if cfg.General.MonoFontFamily != "" {
		t.Errorf("Expected default mono font family to be empty, got %q", cfg.General.MonoFontFamily)
	}

	newTheme := "nord"
	cfg.General.ActiveTheme = newTheme
	cfg.General.IncludePrereleaseUpdates = true
	if err := cm.Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	cm2, err := NewConfigurationManager()
	if err != nil {
		t.Fatalf("Second NewConfigurationManager failed: %v", err)
	}
	cfg2 := cm2.Get()
	if cfg2.General.ActiveTheme != newTheme {
		t.Errorf("Persistence failed: expected theme %s, got %s", newTheme, cfg2.General.ActiveTheme)
	}
	if !cfg2.General.IncludePrereleaseUpdates {
		t.Error("Expected prerelease toggle to persist")
	}
}

func TestConfigurationManager_LegacyTypographyFieldsRemainUnsetOnLoad(t *testing.T) {
	testutil.IsolateUserConfigDir(t)

	configDir, err := tools.GetOrCreateConfigDir()
	if err != nil {
		t.Fatalf("GetOrCreateConfigDir failed: %v", err)
	}

	legacyConfig := []byte(`{
  "general": {
    "activeTheme": "ocean",
    "themeMode": "system",
    "dayTheme": "ocean",
    "nightTheme": "nord",
    "checkForUpdates": true,
    "includePrereleaseUpdates": false,
    "debugMode": false
  },
  "request": {
    "timeoutSeconds": 30,
    "followRedirects": true,
    "maxRedirects": 10,
    "validateSSL": true,
    "defaultUserAgent": "Solo/1.0"
  },
  "customThemes": []
}`)
	if err := os.WriteFile(filepath.Join(configDir, tools.CONFIG_JSON_FILENAME), legacyConfig, 0o600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	cm, err := NewConfigurationManager()
	if err != nil {
		t.Fatalf("NewConfigurationManager failed: %v", err)
	}

	cfg := cm.Get()
	if cfg.General.BaseFontSizePx != 0 {
		t.Errorf("Expected legacy base font size to remain unset (0), got %d", cfg.General.BaseFontSizePx)
	}
	if cfg.General.DefaultFontFamily != "" {
		t.Errorf("Expected legacy sans font family to remain empty, got %q", cfg.General.DefaultFontFamily)
	}
	if cfg.General.MonoFontFamily != "" {
		t.Errorf("Expected legacy mono font family to remain empty, got %q", cfg.General.MonoFontFamily)
	}
}

func TestConfigurationManager_ClampsInvalidBaseFontSize(t *testing.T) {
	testutil.IsolateUserConfigDir(t)

	cm, err := NewConfigurationManager()
	if err != nil {
		t.Fatalf("NewConfigurationManager failed: %v", err)
	}

	cfg := cm.Get()
	cfg.General.BaseFontSizePx = 99
	if err := cm.Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	cm2, err := NewConfigurationManager()
	if err != nil {
		t.Fatalf("Second NewConfigurationManager failed: %v", err)
	}

	cfg2 := cm2.Get()
	if cfg2.General.BaseFontSizePx != tools.DEFAULT_BASE_FONT_SIZE_PX {
		t.Errorf("Expected clamped base font size %d, got %d", tools.DEFAULT_BASE_FONT_SIZE_PX, cfg2.General.BaseFontSizePx)
	}
}

func TestConfigurationManager_ThemeManagement(t *testing.T) {
	testutil.IsolateUserConfigDir(t)

	cm, err := NewConfigurationManager()
	if err != nil {
		t.Fatalf("NewConfigurationManager failed: %v", err)
	}

	if cm.GetActiveTheme() != tools.DEFAULT_THEME {
		t.Errorf("Expected default theme %s, got %s", tools.DEFAULT_THEME, cm.GetActiveTheme())
	}
	allThemes := cm.GetAllThemes()
	predefinedThemes := theme.GetPredefinedThemes()
	if len(allThemes) != len(predefinedThemes) {
		t.Errorf("Expected %d themes initially, got %d", len(predefinedThemes), len(allThemes))
	}

	customTheme := theme.Theme{
		ID:    "custom-test",
		Label: "Custom Test",
		Config: theme.ThemeConfig{
			Type: "custom",
			Mode: theme.ThemeModeSystem,
			Seeds: theme.ThemeSeeds{
				Primary: "#123456",
				Success: "#10b981",
				Warning: "#f59e0b",
				Danger:  "#ef4444",
				Neutral: "#52525b",
			},
		},
	}
	if err := cm.SaveCustomTheme(customTheme); err != nil {
		t.Fatalf("SaveCustomTheme failed: %v", err)
	}

	cm2, err := NewConfigurationManager()
	if err != nil {
		t.Fatalf("Second NewConfigurationManager failed: %v", err)
	}
	customThemes := cm2.GetCustomThemes()
	if len(customThemes) != 1 {
		t.Fatalf("Expected 1 custom theme after saving, got %d", len(customThemes))
	}
	if customThemes[0].ID != "custom-test" {
		t.Errorf("Expected custom theme id 'custom-test', got '%s'", customThemes[0].ID)
	}

	themeByName, err := cm2.GetThemeByName("custom-test")
	if err != nil {
		t.Fatalf("GetThemeByName should find the new custom theme, but failed: %v", err)
	}
	if themeByName.Config.Seeds.Primary != "#123456" {
		t.Errorf("Expected primary color '#123456', got '%s'", themeByName.Config.Seeds.Primary)
	}

	if err := cm2.SetActiveTheme("custom-test"); err != nil {
		t.Fatalf("SetActiveTheme failed: %v", err)
	}
	if cm2.GetActiveTheme() != "custom-test" {
		t.Errorf("Expected active theme 'custom-test', got '%s'", cm2.GetActiveTheme())
	}

	cm3, err := NewConfigurationManager()
	if err != nil {
		t.Fatalf("Third NewConfigurationManager failed: %v", err)
	}
	if cm3.GetActiveTheme() != "custom-test" {
		t.Errorf("Active theme should persist across instances. Expected 'custom-test', got '%s'", cm3.GetActiveTheme())
	}

	updatedTheme := customTheme
	updatedTheme.Config.Seeds.Primary = "#abcdef"
	if err := cm3.SaveCustomTheme(updatedTheme); err != nil {
		t.Fatalf("Updating custom theme failed: %v", err)
	}

	themeByName, err = cm3.GetThemeByName("custom-test")
	if err != nil {
		t.Fatalf("GetThemeByName should find the updated custom theme, but failed: %v", err)
	}
	if themeByName.Config.Seeds.Primary != "#abcdef" {
		t.Errorf("Custom theme color was not updated. Expected '#abcdef', got '%s'", themeByName.Config.Seeds.Primary)
	}

	if err := cm3.DeleteCustomTheme("custom-test"); err != nil {
		t.Fatalf("DeleteCustomTheme failed: %v", err)
	}

	cm4, err := NewConfigurationManager()
	if err != nil {
		t.Fatalf("Fourth NewConfigurationManager failed: %v", err)
	}
	if len(cm4.GetCustomThemes()) != 0 {
		t.Errorf("Custom themes should be empty after deletion, got %d", len(cm4.GetCustomThemes()))
	}

	_, err = cm4.GetThemeByName("custom-test")
	if err == nil {
		t.Errorf("GetThemeByName should fail for a deleted theme, but it succeeded")
	}

	if cm4.GetActiveTheme() != tools.DEFAULT_THEME {
		t.Errorf("Deleting active theme should reset active theme to default. Expected '%s', got '%s'", tools.DEFAULT_THEME, cm4.GetActiveTheme())
	}
}
