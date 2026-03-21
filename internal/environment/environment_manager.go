// Copyright 2026-present raml-dev
// SPDX-License-Identifier: GPL-3.0-only

package environment

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	fs "solo/internal/tools"
	"strings"
)

type EnvironmentManager struct {
	path string
}

func NewEnvironmentManager() *EnvironmentManager {
	appConfigDir, err := fs.GetOrCreateConfigDir()
	if err != nil {
		return nil
	}

	return &EnvironmentManager{filepath.Join(appConfigDir, fs.CONFIG_ENV_DIR)}
}

func (em *EnvironmentManager) CreateEnvironment(name string) error {
	if name == "" {
		return errors.New("no environment name specified")
	}

	// check if a environment with name already exists
	exists, err := em.environmentExists(name)
	if err != nil {
		if _, ok := err.(*os.PathError); !ok {
			return err
		}
	}

	if exists {
		return fmt.Errorf("environment %s already exists", name)
	}

	environment := NewEnvironment(name)

	bytes, err := json.Marshal(environment)

	if err != nil {
		slog.Error("Failed to marshal environment", "name", name, "error", err)
		return err
	}

	if err := os.MkdirAll(em.path, 0700); err != nil {
		slog.Error("Failed to create environment directory", "path", em.path, "error", err)
		return err
	}

	fileName := name
	if !strings.HasSuffix(fileName, ".json") {
		fileName += ".json"
	}

	err = os.WriteFile(filepath.Join(em.path, fileName), bytes, 0600)

	if err != nil {
		slog.Error("Failed to create environment file", "name", name, "error", err)
		return err
	}

	slog.Info("Environment created", "name", name)
	return nil
}

func (em *EnvironmentManager) LoadEnvironments() (*[]string, error) {
	dirEntry, err := fs.ReadConfigDirectory(em.path)

	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(dirEntry))

	for _, e := range dirEntry {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		names = append(names, strings.TrimSuffix(e.Name(), ".json"))
	}

	return &names, nil
}

func (em *EnvironmentManager) LoadEnvironmentsContent() (*[]Environment, error) {
	dirEntry, err := fs.ReadConfigDirectory(em.path)
	if err != nil {
		return nil, err
	}

	environments := make([]Environment, 0, len(dirEntry))
	for _, e := range dirEntry {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		envName := strings.TrimSuffix(e.Name(), ".json")
		env, err := em.LoadEnvironment(envName)
		if err != nil {
			slog.Warn("Failed to load environment in batch", "name", envName, "error", err)
			continue
		}
		environments = append(environments, *env)
	}
	return &environments, nil
}

func (em *EnvironmentManager) LoadEnvironment(name string) (*Environment, error) {
	if name == "" {
		return nil, errors.New("no environment name specified")
	}

	slog.Debug("Loading environment", "name", name)

	fileName := name
	if !strings.HasSuffix(fileName, ".json") {
		fileName += ".json"
	}

	fileBytes, err := os.ReadFile(filepath.Join(em.path, fileName))

	if err != nil {
		slog.Debug("Failed to read environment file", "name", name, "error", err)
		return nil, err
	}
	var rC Environment

	err = json.Unmarshal(fileBytes, &rC)

	if err != nil {
		slog.Error("Failed to parse environment file", "name", name, "error", err)
		return nil, err
	}

	slog.Debug("Environment loaded", "name", name, "values_count", len(rC.Values))
	return &rC, nil
}

func (em *EnvironmentManager) UpdateEnvironment(updated *Environment) error {
	if updated.Name == "" {
		return errors.New("environment name is not specified")
	}

	bytes, err := json.MarshalIndent(updated, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(em.path, 0700); err != nil {
		slog.Error("Failed to create environment directory", "path", em.path, "error", err)
		return err
	}

	fileName := updated.Name
	if !strings.HasSuffix(fileName, ".json") {
		fileName += ".json"
	}

	fName := filepath.Join(em.path, fileName)
	return os.WriteFile(fName, bytes, 0666)
}

func (em *EnvironmentManager) DeleteEnvironment(name string) error {
	if name == "" {
		return errors.New("no environment name specified")
	}

	fileName := name
	if !strings.HasSuffix(fileName, ".json") {
		fileName += ".json"
	}

	err := os.Remove(filepath.Join(em.path, fileName))
	if err != nil {
		slog.Error("Failed to delete environment", "name", name, "error", err)
		return err
	}

	slog.Info("Environment deleted", "name", name)
	return nil
}

func (em *EnvironmentManager) GetValues(name string) (*map[string]ValueType, error) {
	if name == "" {
		return nil, errors.New("no environment name specified")
	}
	env, err := em.LoadEnvironment(name)
	if err != nil {
		return nil, err
	}
	return env.GetValues(), nil
}

func (em *EnvironmentManager) AddValue(environmentName, valueName string, value ValueType) error {
	if environmentName == "" {
		return errors.New("no environment name specified")
	}
	if valueName == "" {
		return errors.New("no value name specified")
	}
	env, err := em.LoadEnvironment(environmentName)
	if err != nil {
		return err
	}

	if err := env.AddValue(valueName, value); err != nil {
		return err
	}

	if err := em.UpdateEnvironment(env); err != nil {
		return err
	}

	slog.Debug("Value added", "environment", environmentName, "key", valueName)
	return nil
}

func (em *EnvironmentManager) RemoveValue(environmentName, valueName string) error {
	if environmentName == "" {
		return errors.New("no environment name specified")
	}

	env, err := em.LoadEnvironment(environmentName)
	if err != nil {
		return err
	}

	if err := env.RemoveValue(valueName); err != nil {
		return err
	}

	if err := em.UpdateEnvironment(env); err != nil {
		return err
	}

	slog.Debug("Value removed", "environment", environmentName, "key", valueName)
	return nil
}

func (em *EnvironmentManager) UpdateValue(environmentName, valueName string, updated ValueType) error {
	if environmentName == "" {
		return errors.New("no environment name specified")
	}
	if valueName == "" {
		return errors.New("no value name specified")
	}

	env, err := em.LoadEnvironment(environmentName)
	if err != nil {
		return err
	}

	if err := env.UpdateValue(valueName, updated); err != nil {
		return err
	}

	if err := em.UpdateEnvironment(env); err != nil {
		return err
	}

	slog.Debug("Value updated", "environment", environmentName, "key", valueName)
	return nil
}

func (em *EnvironmentManager) environmentExists(name string) (bool, error) {
	c, err := em.LoadEnvironment(name)
	if err != nil {
		return false, err
	}
	if c != nil {
		return true, nil
	}
	return false, nil
}

func (em *EnvironmentManager) buildEnvironmentFileName(name string) string {
	return filepath.Join(em.path, name+".json")
}
