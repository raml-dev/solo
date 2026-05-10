// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package configuration

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"solo/internal/theme"
	"solo/internal/tools"
	"sync"
)

func clampBaseFontSizePx(v int) int {
	if v < tools.MIN_BASE_FONT_SIZE_PX || v > tools.MAX_BASE_FONT_SIZE_PX {
		return tools.DEFAULT_BASE_FONT_SIZE_PX
	}
	return v
}

type ConfigurationManager struct {
	mu        sync.RWMutex
	configDir string // Base directory where config.json resides
	config    *Configuration
}

func NewConfigurationManager() (*ConfigurationManager, error) {
	// Setup paths
	baseDir, err := tools.GetOrCreateConfigDir()
	if err != nil {
		slog.Error("Failed to get/create config directory", "error", err)
		return nil, err
	}

	cm := &ConfigurationManager{
		configDir: baseDir,
	}

	// Try to load existing config
	if err := cm.load(); err != nil {
		if os.IsNotExist(err) {
			// Create default if not exists
			slog.Info("Configuration file not found, creating default")
			defaultConfig := cm.createDefault()
			if err := cm.Save(defaultConfig); err != nil {
				slog.Error("Failed to save default configuration", "error", err)
				return nil, err
			}
		} else {
			slog.Error("Failed to load configuration", "error", err)
			return nil, err
		}
	}

	return cm, nil
}

func (cm *ConfigurationManager) createDefault() Configuration {
	return Configuration{
		General: GeneralSettings{
			ActiveTheme:              tools.DEFAULT_THEME,
			ThemeMode:                "system",
			DayTheme:                 tools.DEFAULT_THEME_LIGHT,
			NightTheme:               tools.DEFAULT_THEME_DARK,
			CheckForUpdates:          tools.DEFAULT_CHECK_UPDATES,
			IncludePrereleaseUpdates: tools.DEFAULT_INCLUDE_PRERELEASE_UPDATES,
			DebugMode:                false,
			BaseFontSizePx:           tools.DEFAULT_BASE_FONT_SIZE_PX,
			DefaultFontFamily:        "",
			MonoFontFamily:           "",
		},
		Request: RequestSettings{
			TimeoutSeconds:   tools.DEFAULT_TIMEOUT_SECONDS,
			FollowRedirects:  true,
			MaxRedirects:     tools.DEFAULT_MAX_REDIRECTS,
			ValidateSSL:      tools.DEFAULT_VALIDATE_SSL,
			DefaultUserAgent: tools.DEFAULT_USER_AGENT,
		},
		CustomThemes: []theme.Theme{},
	}
}

func (cm *ConfigurationManager) load() error {
	// Use ReadConfigFile from tools/fs.go which expects (path, filename)
	data, err := tools.ReadConfigFile(cm.configDir, tools.CONFIG_JSON_FILENAME)
	if err != nil {
		return err
	}

	var cfg Configuration
	if err := json.Unmarshal(data, &cfg); err != nil {
		slog.Error("Failed to parse configuration file", "error", err)
		return err
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.config = &cfg

	slog.Info("Configuration loaded")
	slog.Debug("Configuration details",
		"theme", cfg.General.ActiveTheme,
		"timeout", cfg.Request.TimeoutSeconds,
		"follow_redirects", cfg.Request.FollowRedirects,
		"validate_ssl", cfg.Request.ValidateSSL)

	return nil
}

func (cm *ConfigurationManager) Get() Configuration {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	if cm.config == nil {
		return cm.createDefault()
	}
	return *cm.config
}

func (cm *ConfigurationManager) Save(cfg Configuration) error {
	cfg.General.BaseFontSizePx = clampBaseFontSizePx(cfg.General.BaseFontSizePx)

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		slog.Error("Failed to marshal configuration", "error", err)
		return err
	}

	if err := tools.UpdateConfigFile(cm.configDir, tools.CONFIG_JSON_FILENAME, data); err != nil {
		slog.Error("Failed to write configuration file", "error", err)
		return err
	}

	cm.mu.Lock()
	cm.config = &cfg
	cm.mu.Unlock()

	slog.Info("Configuration saved")
	return nil
}

func (cm *ConfigurationManager) GetDefaultConfiguration() Configuration {
	return cm.createDefault()
}

// Theme management methods now part of ConfigurationManager

func (cm *ConfigurationManager) GetActiveTheme() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config.General.ActiveTheme
}

func (cm *ConfigurationManager) SetActiveTheme(themeID string) error {
	cm.mu.Lock()
	cm.config.General.ActiveTheme = themeID
	configToSave := *cm.config
	cm.mu.Unlock()
	return cm.Save(configToSave) // Save immediately to persist
}

func (cm *ConfigurationManager) GetSelectedEnvironment() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config.General.SelectedEnvironment
}

func (cm *ConfigurationManager) SetSelectedEnvironment(name string) error {
	cm.mu.Lock()
	cm.config.General.SelectedEnvironment = name
	configToSave := *cm.config
	cm.mu.Unlock()
	return cm.Save(configToSave)
}

func (cm *ConfigurationManager) GetAllThemes() []theme.Theme {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	predefined := theme.GetPredefinedThemes()
	return append(predefined, cm.config.CustomThemes...)
}

func (cm *ConfigurationManager) GetThemeByName(name string) (*theme.Theme, error) {
	allThemes := cm.GetAllThemes()
	for _, t := range allThemes {
		if t.ID == name {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("theme not found: %s", name)
}

func (cm *ConfigurationManager) GetCustomThemes() []theme.Theme {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.config.CustomThemes
}

func (cm *ConfigurationManager) SaveCustomTheme(th theme.Theme) error {
	cm.mu.Lock()
	found := false
	for i, t := range cm.config.CustomThemes {
		if t.ID == th.ID {
			cm.config.CustomThemes[i] = th
			found = true
			break
		}
	}

	if !found {
		cm.config.CustomThemes = append(cm.config.CustomThemes, th)
	}
	configToSave := *cm.config
	cm.mu.Unlock()

	return cm.Save(configToSave)
}

func (cm *ConfigurationManager) DeleteCustomTheme(themeName string) error {
	cm.mu.Lock()
	newThemes := []theme.Theme{}
	for _, t := range cm.config.CustomThemes {
		if t.ID != themeName {
			newThemes = append(newThemes, t)
		}
	}
	cm.config.CustomThemes = newThemes

	if cm.config.General.ActiveTheme == themeName {
		cm.config.General.ActiveTheme = tools.DEFAULT_THEME
	}
	configToSave := *cm.config
	cm.mu.Unlock()

	return cm.Save(configToSave)
}
