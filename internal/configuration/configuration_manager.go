package configuration

import (
	"encoding/json"
	"log/slog"
	"os"
	"sync"
	"yapla/internal/tools"
)

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
			Theme:           tools.DEFAULT_THEME,
			CheckForUpdates: tools.DEFAULT_CHECK_UPDATES,
		},
		Request: RequestSettings{
			TimeoutSeconds:   tools.DEFAULT_TIMEOUT_SECONDS,
			FollowRedirects:  true,
			MaxRedirects:     tools.DEFAULT_MAX_REDIRECTS,
			ValidateSSL:      tools.DEFAULT_VALIDATE_SSL,
			DefaultUserAgent: tools.DEFAULT_USER_AGENT,
		},
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
		"theme", cfg.General.Theme,
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
