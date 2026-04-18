// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package importer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"solo/internal/collection"
	"strings"
	"time"
)

// BrunoImporter implements the Importer interface for Bruno collections.
type BrunoImporter struct{}

func NewBrunoImporter() *BrunoImporter {
	return &BrunoImporter{}
}

// Import processes a directory containing a Bruno collection.
func (b *BrunoImporter) Import(dirPath string) (*collection.Collection, error) {
	// 1. Read collection name from bruno.json
	brunoJSONPath := filepath.Join(dirPath, "bruno.json")
	fileData, err := os.ReadFile(brunoJSONPath)
	if err != nil {
		return nil, fmt.Errorf("bruno.json not found in the directory: %w", err)
	}
	slog.Info("Bruno Importer: Found and read bruno.json")

	var brunoConfig struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(fileData, &brunoConfig); err != nil {
		return nil, fmt.Errorf("error parsing bruno.json: %w", err)
	}
	slog.Info("Bruno Importer: Parsed collection name", "name", brunoConfig.Name)

	now := time.Now()
	coll := &collection.Collection{
		Id:                  generateUUID(),
		Name:                brunoConfig.Name,
		Requests:            []collection.Request{},
		Folders:             []collection.Folder{},
		CreationTimestamp:   now,
		LastUpdateTimestamp: now,
	}

	slog.Info("Bruno Importer: Starting directory walk", "path", dirPath)
	var filesFound int
	// 2. Walk the directory to find .bru files
	err = filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			slog.Error("Bruno Importer: Error accessing path", "path", path, "error", err)
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".bru") {
			slog.Info("Bruno Importer: Found .bru file", "path", path)
			filesFound++
			request, folderNames, err := parseBruFile(path, dirPath)
			if err != nil {
				// Log the error but continue with other files
				slog.Warn("Bruno Importer: Skipping file due to parse error", "path", path, "error", err)
				return nil
			}

			if request != nil {
				slog.Info("Bruno Importer: Successfully parsed request", "name", request.Name, "url", request.Url)
				addRequestToCollectionTree(coll, folderNames, *request)
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %w", err)
	}

	slog.Info("Bruno Importer: Directory walk completed", "bru_files_found", filesFound, "requests_parsed", countRequestsInCollection(coll))
	return coll, nil
}

// parseBruFile reads a single .bru file and converts it into a models.Request.
func parseBruFile(filePath string, basePath string) (*collection.Request, []string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	req := collection.Request{
		Id:                  generateUUID(),
		Headers:             make(map[string]string),
		Cookies:             make(map[string]string),
		CreationTimestamp:   time.Now(),
		LastUpdateTimestamp: time.Now(),
	}

	// Determine request name from path
	relPath, _ := filepath.Rel(basePath, filePath)
	nameParts := strings.Split(filepath.ToSlash(relPath), "/")
	folderNames := append([]string(nil), nameParts[:len(nameParts)-1]...)
	// Remove .bru extension from the last part
	lastPart := nameParts[len(nameParts)-1]
	req.Name = strings.TrimSuffix(lastPart, ".bru")

	scanner := bufio.NewScanner(file)
	var currentSection string
	var isBodyBlock bool
	isHTTPRequest := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "meta {") {
			currentSection = "meta"
			isBodyBlock = false
			continue
		} else if strings.HasPrefix(line, "get {") || strings.HasPrefix(line, "post {") || strings.HasPrefix(line, "put {") || strings.HasPrefix(line, "patch {") || strings.HasPrefix(line, "delete {") {
			currentSection = "http"
			isBodyBlock = false
			isHTTPRequest = true
			req.Verb = strings.ToUpper(strings.TrimSuffix(line, " {"))
			continue
		} else if strings.HasPrefix(line, "headers {") {
			currentSection = "headers"
			isBodyBlock = false
			continue
		} else if strings.HasPrefix(line, "body:json {") {
			currentSection = "body"
			req.BodyType = "json"
			isBodyBlock = true
			continue
		} else if strings.HasPrefix(line, "body {") {
			currentSection = "body"
			req.BodyType = "text" // Assume text for simple body
			isBodyBlock = true
			continue
		} else if line == "}" {
			if isBodyBlock {
				isBodyBlock = false // End of a multi-line body block
			} else {
				currentSection = "" // End of a section
			}
			continue
		}

		if isBodyBlock && currentSection == "body" {
			// This is a line inside a multi-line body
			req.Body += scanner.Text() + "\n" // Use scanner.Text() to preserve indentation
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue // Not a key-value line
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch currentSection {
		case "meta":
			if key == "name" {
				req.Name = value
			}
		case "http":
			if key == "url" {
				req.Url = value
			}
			if key == "body" && value != "none" {
				req.BodyType = value
			}
		case "headers":
			req.Headers[key] = value
		}
	}

	// Clean up body (remove trailing newline)
	req.Body = strings.TrimRight(req.Body, "\n")

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	if !isHTTPRequest || req.Verb == "" || req.Url == "" {
		return nil, nil, nil
	}

	return &req, folderNames, nil
}

func addRequestToCollectionTree(coll *collection.Collection, folderNames []string, req collection.Request) {
	if len(folderNames) == 0 {
		coll.Requests = append(coll.Requests, req)
		return
	}

	currentFolders := &coll.Folders
	for _, folderName := range folderNames {
		folderIdx := indexFolderByName(*currentFolders, folderName)
		if folderIdx == -1 {
			*currentFolders = append(*currentFolders, collection.NewFolder(folderName))
			folderIdx = len(*currentFolders) - 1
		}
		currentFolders = &(*currentFolders)[folderIdx].Folders
	}

	parent := findFolderByPath(&coll.Folders, folderNames)
	if parent == nil {
		coll.Requests = append(coll.Requests, req)
		return
	}

	parent.Requests = append(parent.Requests, req)
}

func indexFolderByName(folders []collection.Folder, name string) int {
	for i := range folders {
		if folders[i].Name == name {
			return i
		}
	}

	return -1
}

func findFolderByPath(folders *[]collection.Folder, folderNames []string) *collection.Folder {
	currentFolders := folders
	var current *collection.Folder

	for _, folderName := range folderNames {
		idx := indexFolderByName(*currentFolders, folderName)
		if idx == -1 {
			return nil
		}
		current = &(*currentFolders)[idx]
		currentFolders = &current.Folders
	}

	return current
}

func countRequestsInCollection(coll *collection.Collection) int {
	total := len(coll.Requests)
	for i := range coll.Folders {
		total += countRequestsInFolder(&coll.Folders[i])
	}

	return total
}

func countRequestsInFolder(folder *collection.Folder) int {
	total := len(folder.Requests)
	for i := range folder.Folders {
		total += countRequestsInFolder(&folder.Folders[i])
	}

	return total
}
