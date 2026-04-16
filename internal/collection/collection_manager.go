// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package collection

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"solo/internal/tools"
	fs "solo/internal/tools"
	"strings"
)

type CollectionManager struct {
	config string
}

func NewCollectionManager() *CollectionManager {

	config, err := fs.GetMainConfig(fs.CONFIG_COLLECTION_DIR)
	if err != nil {
		return nil
	}

	return &CollectionManager{config}
}

func (cm *CollectionManager) GetConfigPath() (string, error) {
	if cm.config == "" {
		return "", errors.New("configuration path not set")
	}
	return cm.config, nil
}

func (cm *CollectionManager) CreateCollection(collectionName string) error {
	if collectionName == "" {
		return errors.New("no collection name specified")
	}

	// check if a collection with name already exists
	exists, err := cm.collectionExists(collectionName)
	if err != nil {
		if _, ok := err.(*os.PathError); !ok {
			return err
		}
	}

	if exists {
		return fmt.Errorf("collection %s already exists", collectionName)
	}

	collection := NewCollection(collectionName)

	bytes, err := json.Marshal(collection)

	if err != nil {
		slog.Error("Failed to marshal collection", "name", collectionName, "error", err)
		return err
	}

	fileName := collectionName
	if !strings.HasSuffix(fileName, ".json") {
		fileName += ".json"
	}

	if err := fs.CreateConfigFile(cm.config, fileName, bytes); err != nil {
		slog.Error("Failed to create collection file", "name", collectionName, "error", err)
		return err
	}

	slog.Info("Collection created", "name", collectionName)
	return nil
}

func (cm *CollectionManager) LoadCollections() (*[]string, error) {
	dirEntry, err := fs.ReadConfigDirectory(cm.config)

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

func (cm *CollectionManager) LoadCollectionsContent() (*[]Collection, error) {

	dirEntry, err := fs.ReadConfigDirectory(cm.config)

	if err != nil {
		return nil, err
	}

	collections := make([]Collection, 0, len(dirEntry))

	for _, e := range dirEntry {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		collectionName := strings.TrimSuffix(e.Name(), ".json")
		coll, err := cm.LoadCollection(collectionName)

		if err != nil {
			slog.Warn("Failed to load collection in batch", "name", collectionName, "error", err)
			continue
		}

		collections = append(collections, *coll)
	}

	return &collections, nil
}

func (cm *CollectionManager) LoadCollection(collectionName string) (*Collection, error) {
	if collectionName == "" {
		return nil, errors.New("no collection name specified")
	}

	slog.Debug("Loading collection", "name", collectionName)

	fileName := collectionName
	if !strings.HasSuffix(fileName, ".json") {
		fileName += ".json"
	}

	fileBytes, err := fs.ReadConfigFile(cm.config, fileName)

	if err != nil {
		slog.Debug("Failed to read collection file", "name", collectionName, "error", err)
		return nil, err
	}
	var rC Collection

	err = json.Unmarshal(fileBytes, &rC)

	if err != nil {
		slog.Error("Failed to parse collection file", "name", collectionName, "error", err)
		return nil, err
	}

	slog.Debug("Collection loaded", "name", collectionName, "requests_count", len(rC.Requests))
	return &rC, nil
}

func (cm *CollectionManager) UpdateCollection(updated Collection) error {
	if updated.Name == "" {
		return errors.New("collection name is not specified")
	}

	bytes, err := json.MarshalIndent(updated, "", "  ")
	if err != nil {
		return err
	}

	fileName := updated.Name
	if !strings.HasSuffix(fileName, ".json") {
		fileName += ".json"
	}

	return fs.UpdateConfigFile(cm.config, fileName, bytes)
}

func (cm *CollectionManager) DeleteCollection(collectionName string) error {
	if collectionName == "" {
		return errors.New("no collection name specified")
	}

	fileName := collectionName
	if !strings.HasSuffix(fileName, ".json") {
		fileName += ".json"
	}

	if err := tools.RemoveConfigFile(cm.config, fileName); err != nil {
		slog.Error("Failed to delete collection", "name", collectionName, "error", err)
		return err
	}

	slog.Info("Collection deleted", "name", collectionName)
	return nil
}

// requests

func (cm *CollectionManager) GetRequest(collectionName, requestId string) (*Request, error) {
	if collectionName == "" {
		return nil, errors.New("no collection name specified")
	}
	coll, err := cm.LoadCollection(collectionName)
	if err != nil {
		return nil, err
	}
	return coll.GetRequestById(requestId)
}

func (cm *CollectionManager) GetRequests(collectionName string) (*[]Request, error) {
	if collectionName == "" {
		return nil, errors.New("no collection name specified")
	}
	coll, err := cm.LoadCollection(collectionName)
	if err != nil {
		return nil, err
	}
	return coll.GetRequests(), nil
}

func (cm *CollectionManager) AddRequest(collectionName string, request Request) (*Request, error) {
	if collectionName == "" {
		return nil, errors.New("no collection name specified")
	}
	coll, err := cm.LoadCollection(collectionName)
	if err != nil {
		return nil, err
	}

	newRequest, err := coll.AddRequest(request)
	if err != nil {
		return nil, err
	}

	if err := cm.UpdateCollection(*coll); err != nil {
		return nil, err
	}

	slog.Debug("Request added", "collection", collectionName, "request_id", newRequest.Id, "request_name", newRequest.Name)
	return newRequest, nil
}

func (cm *CollectionManager) AddRequestToFolder(collectionName, folderId string, request Request) (*Request, error) {
	if collectionName == "" {
		return nil, errors.New("no collection name specified")
	}
	if folderId == "" {
		return nil, errors.New("no folder id specified")
	}
	coll, err := cm.LoadCollection(collectionName)
	if err != nil {
		return nil, err
	}

	newRequest, err := coll.AddRequestToFolder(folderId, request)
	if err != nil {
		return nil, err
	}

	if err := cm.UpdateCollection(*coll); err != nil {
		return nil, err
	}

	slog.Debug("Request added to folder", "collection", collectionName, "folder_id", folderId, "request_id", newRequest.Id)
	return newRequest, nil
}

func (cm *CollectionManager) RemoveRequest(collectionName string, requestId string) error {
	if collectionName == "" {
		return errors.New("no collection name specified")
	}
	if requestId == "" {
		return errors.New("no request id specified")
	}
	coll, err := cm.LoadCollection(collectionName)
	if err != nil {
		return err
	}

	if err := coll.RemoveRequest(requestId); err != nil {
		return err
	}

	if err := cm.UpdateCollection(*coll); err != nil {
		return err
	}

	slog.Debug("Request removed", "collection", collectionName, "request_id", requestId)
	return nil
}

func (cm *CollectionManager) UpdateRequest(collectionName string, updated Request) error {
	if collectionName == "" {
		return errors.New("no collection name specified")
	}
	coll, err := cm.LoadCollection(collectionName)
	if err != nil {
		return err
	}

	if err := coll.UpdateRequest(updated); err != nil {
		return err
	}

	if err := cm.UpdateCollection(*coll); err != nil {
		return err
	}

	slog.Debug("Request updated", "collection", collectionName, "request_id", updated.Id)
	return nil
}

// folders

func (cm *CollectionManager) GetFolders(collectionName string) (*[]Folder, error) {
	if collectionName == "" {
		return nil, errors.New("no collection name specified")
	}
	coll, err := cm.LoadCollection(collectionName)
	if err != nil {
		return nil, err
	}
	return coll.GetFolders(), nil
}

func (cm *CollectionManager) GetFolder(collectionName, folderId string) (*Folder, error) {
	if collectionName == "" {
		return nil, errors.New("no collection name specified")
	}
	if folderId == "" {
		return nil, errors.New("no folder id specified")
	}
	coll, err := cm.LoadCollection(collectionName)
	if err != nil {
		return nil, err
	}
	return coll.GetFolderById(folderId)
}

func (cm *CollectionManager) AddFolder(collectionName string, folder Folder) (*Folder, error) {
	if collectionName == "" {
		return nil, errors.New("no collection name specified")
	}
	coll, err := cm.LoadCollection(collectionName)
	if err != nil {
		return nil, err
	}

	newFolder, err := coll.AddFolder(folder)
	if err != nil {
		return nil, err
	}

	if err := cm.UpdateCollection(*coll); err != nil {
		return nil, err
	}

	slog.Debug("Folder added", "collection", collectionName, "folder_id", newFolder.Id, "folder_name", newFolder.Name)
	return newFolder, nil
}

func (cm *CollectionManager) AddSubFolder(collectionName, parentFolderId string, folder Folder) (*Folder, error) {
	if collectionName == "" {
		return nil, errors.New("no collection name specified")
	}
	if parentFolderId == "" {
		return nil, errors.New("no parent folder id specified")
	}
	coll, err := cm.LoadCollection(collectionName)
	if err != nil {
		return nil, err
	}

	newFolder, err := coll.AddSubFolder(parentFolderId, folder)
	if err != nil {
		return nil, err
	}

	if err := cm.UpdateCollection(*coll); err != nil {
		return nil, err
	}

	slog.Debug("Subfolder added", "collection", collectionName, "parent_folder_id", parentFolderId, "folder_id", newFolder.Id)
	return newFolder, nil
}

func (cm *CollectionManager) RemoveFolder(collectionName, folderId string) error {
	if collectionName == "" {
		return errors.New("no collection name specified")
	}
	if folderId == "" {
		return errors.New("no folder id specified")
	}
	coll, err := cm.LoadCollection(collectionName)
	if err != nil {
		return err
	}

	if err := coll.RemoveFolder(folderId); err != nil {
		return err
	}

	if err := cm.UpdateCollection(*coll); err != nil {
		return err
	}

	slog.Debug("Folder removed", "collection", collectionName, "folder_id", folderId)
	return nil
}

func (cm *CollectionManager) UpdateFolder(collectionName string, updated Folder) error {
	if collectionName == "" {
		return errors.New("no collection name specified")
	}
	coll, err := cm.LoadCollection(collectionName)
	if err != nil {
		return err
	}

	if err := coll.UpdateFolder(updated); err != nil {
		return err
	}

	if err := cm.UpdateCollection(*coll); err != nil {
		return err
	}

	slog.Debug("Folder updated", "collection", collectionName, "folder_id", updated.Id)
	return nil
}
