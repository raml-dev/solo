package tools

import (
	"os"
	"path/filepath"
)

func GetOrCreateConfigDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	appConfigDir := filepath.Join(configDir, MAIN_DIR)
	if err := os.MkdirAll(appConfigDir, 0755); err != nil {
		return "", err
	}

	return appConfigDir, nil
}
