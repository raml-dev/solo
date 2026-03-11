package importer

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"time"
	"yapla/internal/environment"
)

type BrunoEnvironmentImporter struct{}

func NewBrunoEnvironmentImporter() *BrunoEnvironmentImporter {
	return &BrunoEnvironmentImporter{}
}

func (b *BrunoEnvironmentImporter) Import(path string) (*environment.Environment, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inVars := false
	values := map[string]environment.ValueType{}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "vars {") {
			inVars = true
			continue
		}

		if inVars && line == "}" {
			inVars = false
			continue
		}

		if !inVars {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			continue
		}

		values[key] = environment.ValueType{
			Value: value,
			Type:  "string",
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	name := strings.TrimSuffix(filepath.Base(path), ".bru")
	if name == "" {
		name = "Imported Environment"
	}

	now := time.Now()
	return &environment.Environment{
		Id:                  generateUUID(),
		Name:                name,
		Values:              values,
		CreationTimestamp:   now,
		LastUpdateTimestamp: now,
	}, nil
}

