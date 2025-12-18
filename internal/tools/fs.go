package fs

import (
	"os"
	"path/filepath"
)

const (
	MAIN_DIR              = "yapla"
	CONFIG_JSON_FILENAME  = "config.json"
	CONFIG_COLLECTION_DIR = "collections"
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
