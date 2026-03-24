// Copyright 2026-present raml-dev
// SPDX-License-Identifier: GPL-3.0-only

package tools

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetOrCreateConfigDir(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", tmpHome)
	// On Windows, os.UserConfigDir uses AppData
	t.Setenv("AppData", tmpHome)

	path, err := GetOrCreateConfigDir()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Expected directory %s to be created", path)
	}

	expectedBase := MAIN_DIR
	if filepath.Base(path) != expectedBase {
		t.Errorf("Expected base directory %s, got %s", expectedBase, filepath.Base(path))
	}
}

func TestGetMainConfig(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", tmpHome)
	t.Setenv("AppData", tmpHome)

	configName := "test-config"
	path, err := GetMainConfig(configName)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Expected directory %s to be created", path)
	}

	if filepath.Base(path) != configName {
		t.Errorf("Expected base directory %s, got %s", configName, filepath.Base(path))
	}

	// Verify it's a subdirectory of the app config dir
	parent := filepath.Dir(path)
	if filepath.Base(parent) != MAIN_DIR {
		t.Errorf("Expected parent directory %s, got %s", MAIN_DIR, filepath.Base(parent))
	}
}

func TestGetMainConfig_Error(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", tmpHome)
	t.Setenv("AppData", tmpHome)

	// Create a file with the same name as the directory we want to create
	mainConfigDir, _ := GetOrCreateConfigDir()
	configName := "blocked-by-file"
	blockedPath := filepath.Join(mainConfigDir, configName)
	if err := os.WriteFile(blockedPath, []byte("I am a file"), 0644); err != nil {
		t.Fatalf("Failed to create blocking file: %v", err)
	}

	_, err := GetMainConfig(configName)
	if err == nil {
		t.Errorf("Expected error when directory creation is blocked by a file, but got none")
	}
}

func TestReadConfigDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Test non-existent directory
	entries, err := ReadConfigDirectory(filepath.Join(tmpDir, "non-existent"))
	if err != nil {
		t.Errorf("ReadConfigDirectory on non-existent path should not return error, got %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries for non-existent path, got %d", len(entries))
	}

	// Test existing directory with files
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	entries, err = ReadConfigDirectory(tmpDir)
	if err != nil {
		t.Errorf("Unexpected error reading existing directory: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
}

func TestConfigFileOperations(t *testing.T) {
	tmpDir := t.TempDir()
	fileName := "config.test"
	content := []byte("hello world")

	// Test Create
	err := CreateConfigFile(tmpDir, fileName, content)
	if err != nil {
		t.Fatalf("CreateConfigFile failed: %v", err)
	}

	// Verify file exists and content matches
	readContent, err := ReadConfigFile(tmpDir, fileName)
	if err != nil {
		t.Fatalf("ReadConfigFile failed: %v", err)
	}
	if string(readContent) != string(content) {
		t.Errorf("Expected content %s, got %s", string(content), string(readContent))
	}

	// Test Update
	newContent := []byte("updated content")
	err = UpdateConfigFile(tmpDir, fileName, newContent)
	if err != nil {
		t.Fatalf("UpdateConfigFile failed: %v", err)
	}

	readContent, err = ReadConfigFile(tmpDir, fileName)
	if err != nil {
		t.Fatalf("ReadConfigFile failed after update: %v", err)
	}
	if string(readContent) != string(newContent) {
		t.Errorf("Expected content %s, got %s", string(newContent), string(readContent))
	}

	// Test Remove
	err = RemoveConfigFile(tmpDir, fileName)
	if err != nil {
		t.Fatalf("RemoveConfigFile failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpDir, fileName)); !os.IsNotExist(err) {
		t.Errorf("File should be removed but still exists")
	}
}
