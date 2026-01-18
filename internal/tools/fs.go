package tools

import (
	"fmt"
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

func GetMainConfig(name string) (string, error) {
	configPath, err := GetOrCreateConfigDir()

	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", configPath, name), nil
}

func ReadConfigDirectory(configPath string) ([]os.DirEntry, error) {
	return os.ReadDir(configPath)
}

func CreateConfigFile(configPath, fileName string, content []byte) error {
	if err := os.MkdirAll(configPath, 0700); err != nil {
		return err
	}
	fName := fmt.Sprintf("%s/%s", configPath, fileName)
	return os.WriteFile(fName, content, 0600)
}

func UpdateConfigFile(configPath, fileName string, content []byte) error {
	fName := fmt.Sprintf("%s/%s", configPath, fileName)
	return os.WriteFile(fName, content, 0666)
}

func RemoveConfigFile(configPath, fileName string) error {
	fName := fmt.Sprintf("%s/%s", configPath, fileName)
	return os.Remove(fName)
}

func ReadConfigFile(configPath, fileName string) ([]byte, error) {
	fName := fmt.Sprintf("%s/%s", configPath, fileName)
	return os.ReadFile(fName)

}
