// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package collection

import (
	"errors"
	"fmt"
	"os"
	"sync"
)

const maxConcurrentCollectionImports = 4

type BatchImportResult struct {
	Results []BatchImportItemResult `json:"results"`
}

type BatchImportItemResult struct {
	Path     string   `json:"path"`
	Name     string   `json:"name,omitempty"`
	Success  bool     `json:"success"`
	Conflict bool     `json:"conflict,omitempty"`
	Error    string   `json:"error,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

type batchImportParsedResult struct {
	collection *Collection
	warnings   []string
	err        error
}

func (cm *CollectionManager) ImportBatch(
	paths []string,
	overwriteExisting bool,
	importOne func(path string) (*Collection, []string, error),
) (BatchImportResult, error) {
	if cm == nil {
		return BatchImportResult{}, fmt.Errorf("collection manager not initialized")
	}
	if _, err := cm.GetConfigPath(); err != nil {
		return BatchImportResult{}, fmt.Errorf("collection manager configuration unusable: %w", err)
	}
	if importOne == nil {
		return BatchImportResult{}, fmt.Errorf("collection import function not provided")
	}
	if len(paths) == 0 {
		return BatchImportResult{}, fmt.Errorf("no collection import paths provided")
	}

	parsed := parseBatchCollectionImports(paths, importOne)
	result := BatchImportResult{
		Results: make([]BatchImportItemResult, len(paths)),
	}

	for index, item := range parsed {
		result.Results[index].Path = paths[index]
		if item.err != nil {
			result.Results[index].Error = item.err.Error()
			continue
		}
		if item.collection == nil {
			result.Results[index].Error = "import returned no collection"
			continue
		}

		result.Results[index].Name = item.collection.Name
		result.Results[index].Warnings = item.warnings

		if item.collection.Name != "" {
			exists, err := cm.collectionExists(item.collection.Name)
			if err != nil {
				var pathErr *os.PathError
				if errors.As(err, &pathErr) {
					err = nil
				}
			}
			if err != nil {
				result.Results[index].Error = err.Error()
				continue
			}
			if exists && !overwriteExisting {
				result.Results[index].Conflict = true
				result.Results[index].Error = fmt.Sprintf("collection %s already exists", item.collection.Name)
				continue
			}
		}

		if err := cm.UpdateCollection(*item.collection); err != nil {
			result.Results[index].Error = err.Error()
			continue
		}

		result.Results[index].Success = true
	}

	return result, nil
}

func parseBatchCollectionImports(
	paths []string,
	importOne func(path string) (*Collection, []string, error),
) []batchImportParsedResult {
	parsed := make([]batchImportParsedResult, len(paths))

	limit := maxConcurrentCollectionImports
	if len(paths) < limit {
		limit = len(paths)
	}

	jobs := make(chan int)
	var wg sync.WaitGroup
	wg.Add(limit)
	for range limit {
		go func() {
			defer wg.Done()
			for index := range jobs {
				path := paths[index]
				if path == "" {
					parsed[index].err = fmt.Errorf("path is empty")
					continue
				}

				coll, warnings, err := importOne(path)
				parsed[index] = batchImportParsedResult{
					collection: coll,
					warnings:   warnings,
					err:        err,
				}
			}
		}()
	}

	for index := range paths {
		jobs <- index
	}
	close(jobs)
	wg.Wait()

	return parsed
}
