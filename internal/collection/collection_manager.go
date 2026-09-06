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
	"regexp"
	"solo/internal/tools"
	fs "solo/internal/tools"
	"strings"
)

type CollectionManager struct {
	config      string
	secretStore AuthSecretStore
}

var bearerTokenPlaceholderPattern = regexp.MustCompile(tools.PLACEHOLDER_REGEXP)

func isBearerTokenPlaceholder(value string) bool {
	trimmedValue := strings.TrimSpace(value)
	matches := bearerTokenPlaceholderPattern.FindAllString(trimmedValue, -1)
	return len(matches) == 1 && matches[0] == trimmedValue
}

func NewCollectionManager(secretStores ...AuthSecretStore) *CollectionManager {

	config, err := fs.GetMainConfig(fs.CONFIG_COLLECTION_DIR)
	if err != nil {
		return nil
	}

	var secretStore AuthSecretStore
	if len(secretStores) > 0 {
		secretStore = secretStores[0]
	}

	return &CollectionManager{config: config, secretStore: secretStore}
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
	previous, _ := cm.LoadCollection(updated.Name)
	if err := cm.prepareCollectionAuthForStorage(&updated); err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(updated, "", "  ")
	if err != nil {
		return err
	}

	fileName := updated.Name
	if !strings.HasSuffix(fileName, ".json") {
		fileName += ".json"
	}

	if err := fs.UpdateConfigFile(cm.config, fileName, bytes); err != nil {
		return err
	}
	cm.deleteUnreferencedBearerTokens(previous, &updated)
	return nil
}

func (cm *CollectionManager) DeleteCollection(collectionName string) error {
	if collectionName == "" {
		return errors.New("no collection name specified")
	}

	coll, loadErr := cm.LoadCollection(collectionName)

	fileName := collectionName
	if !strings.HasSuffix(fileName, ".json") {
		fileName += ".json"
	}

	if err := tools.RemoveConfigFile(cm.config, fileName); err != nil {
		slog.Error("Failed to delete collection", "name", collectionName, "error", err)
		return err
	}
	if loadErr == nil {
		cm.deleteCollectionBearerTokens(coll)
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

	if err := cm.cloneBearerTokenForNewRequest(&request); err != nil {
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

	if err := cm.cloneBearerTokenForNewRequest(&request); err != nil {
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

func (cm *CollectionManager) cloneBearerTokenForNewRequest(request *Request) error {
	if request == nil || request.Auth == nil || request.Auth.BearerTokenID == "" || request.Auth.BearerToken != "" {
		return nil
	}
	if cm.secretStore == nil {
		return errors.New("auth secret store not configured")
	}

	token, err := cm.secretStore.GetBearerToken(request.Auth.BearerTokenID)
	if err != nil {
		return err
	}
	request.Auth.BearerTokenID = ""
	request.Auth.BearerToken = token
	return nil
}

func (cm *CollectionManager) prepareCollectionAuthForStorage(coll *Collection) error {
	if coll == nil {
		return nil
	}
	if err := cm.prepareAuthForStorage(&coll.Auth); err != nil {
		return err
	}
	for i := range coll.Requests {
		if err := cm.prepareAuthForStorage(coll.Requests[i].Auth); err != nil {
			return fmt.Errorf("request %q: %w", coll.Requests[i].Name, err)
		}
	}
	for i := range coll.Folders {
		if err := cm.prepareFolderAuthForStorage(&coll.Folders[i]); err != nil {
			return err
		}
	}
	return nil
}

func (cm *CollectionManager) prepareFolderAuthForStorage(folder *Folder) error {
	for i := range folder.Requests {
		if err := cm.prepareAuthForStorage(folder.Requests[i].Auth); err != nil {
			return fmt.Errorf("request %q: %w", folder.Requests[i].Name, err)
		}
	}
	for i := range folder.Folders {
		if err := cm.prepareFolderAuthForStorage(&folder.Folders[i]); err != nil {
			return err
		}
	}
	return nil
}

func (cm *CollectionManager) prepareAuthForStorage(authConfig *AuthConfiguration) error {
	if authConfig == nil {
		return nil
	}
	authConfig.Normalize()
	if authConfig.BearerToken == "" {
		return nil
	}
	if isBearerTokenPlaceholder(authConfig.BearerToken) {
		authConfig.BearerTokenID = ""
		return nil
	}
	if cm.secretStore == nil {
		return errors.New("auth secret store not configured")
	}

	tokenID, err := cm.secretStore.StoreBearerToken(authConfig.BearerTokenID, authConfig.BearerToken)
	if err != nil {
		return err
	}
	authConfig.BearerTokenID = tokenID
	authConfig.BearerToken = ""
	return nil
}

func (cm *CollectionManager) deleteCollectionBearerTokens(coll *Collection) {
	if coll == nil || cm.secretStore == nil {
		return
	}
	seen := make(map[string]struct{})
	collectBearerTokenIDs(coll, seen)
	for tokenID := range seen {
		if err := cm.secretStore.DeleteBearerToken(tokenID); err != nil {
			slog.Warn("Failed to remove bearer token for deleted collection", "token_id", tokenID, "error", err)
		}
	}
}

func (cm *CollectionManager) deleteUnreferencedBearerTokens(previous, current *Collection) {
	if previous == nil || cm.secretStore == nil {
		return
	}
	previousIDs := make(map[string]struct{})
	currentIDs := make(map[string]struct{})
	collectBearerTokenIDs(previous, previousIDs)
	collectBearerTokenIDs(current, currentIDs)
	for tokenID := range previousIDs {
		if _, stillReferenced := currentIDs[tokenID]; stillReferenced {
			continue
		}
		if err := cm.secretStore.DeleteBearerToken(tokenID); err != nil {
			slog.Warn("Failed to remove unreferenced bearer token", "token_id", tokenID, "error", err)
		}
	}
}

func collectBearerTokenIDs(coll *Collection, tokenIDs map[string]struct{}) {
	if coll.Auth.BearerTokenID != "" {
		tokenIDs[coll.Auth.BearerTokenID] = struct{}{}
	}
	for i := range coll.Requests {
		if coll.Requests[i].Auth != nil && coll.Requests[i].Auth.BearerTokenID != "" {
			tokenIDs[coll.Requests[i].Auth.BearerTokenID] = struct{}{}
		}
	}
	for i := range coll.Folders {
		collectFolderBearerTokenIDs(&coll.Folders[i], tokenIDs)
	}
}

func collectFolderBearerTokenIDs(folder *Folder, tokenIDs map[string]struct{}) {
	for i := range folder.Requests {
		if folder.Requests[i].Auth != nil && folder.Requests[i].Auth.BearerTokenID != "" {
			tokenIDs[folder.Requests[i].Auth.BearerTokenID] = struct{}{}
		}
	}
	for i := range folder.Folders {
		collectFolderBearerTokenIDs(&folder.Folders[i], tokenIDs)
	}
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

func (cm *CollectionManager) AddFolder(collectionName string, parentFolderId string, folder Folder) (*Folder, error) {
	if collectionName == "" {
		return nil, errors.New("no collection name specified")
	}
	coll, err := cm.LoadCollection(collectionName)
	if err != nil {
		return nil, err
	}

	newFolder, err := coll.AddFolder(parentFolderId, folder)
	if err != nil {
		return nil, err
	}

	if err := cm.UpdateCollection(*coll); err != nil {
		return nil, err
	}

	slog.Debug("Folder added", "collection", collectionName, "parent_folder_id", parentFolderId, "folder_id", newFolder.Id, "folder_name", newFolder.Name)
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
