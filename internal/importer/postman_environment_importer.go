package importer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"yapla/internal/environment"
)

type PostmanEnvironmentImporter struct{}

func NewPostmanEnvironmentImporter() *PostmanEnvironmentImporter {
	return &PostmanEnvironmentImporter{}
}

type postmanEnvironment struct {
	Name   string                `json:"name"`
	Values []postmanEnvironmentValue `json:"values"`
}

type postmanEnvironmentValue struct {
	Key     string      `json:"key"`
	Value   interface{} `json:"value"`
	Enabled bool        `json:"enabled"`
}

func (p *PostmanEnvironmentImporter) Import(path string) (*environment.Environment, error) {
	fileData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read file: %w", err)
	}

	var payload postmanEnvironment
	if err := json.Unmarshal(fileData, &payload); err != nil {
		return nil, fmt.Errorf("error parsing Postman environment: %w", err)
	}

	if payload.Name == "" {
		base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		if base != "" {
			payload.Name = base
		} else {
			payload.Name = "Imported Environment"
		}
	}

	now := time.Now()
	env := &environment.Environment{
		Id:                  generateUUID(),
		Name:                payload.Name,
		Values:              map[string]environment.ValueType{},
		CreationTimestamp:   now,
		LastUpdateTimestamp: now,
	}

	for _, entry := range payload.Values {
		value := ""
		if entry.Value != nil {
			value = fmt.Sprintf("%v", entry.Value)
		}
		env.Values[entry.Key] = environment.ValueType{
			Value: value,
			Type:  "string",
		}
	}

	return env, nil
}
