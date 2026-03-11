package collection

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"yapla/internal/tools"
	fs "yapla/internal/tools"
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

	if err := fs.CreateConfigFile(cm.config, collectionName, bytes); err != nil {
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
		names = append(names, e.Name())
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
		collectionName := e.Name()
		if filepath.Ext(collectionName) == ".json" {
			collectionName = collectionName[:len(collectionName)-5]
		}
		coll, err := cm.LoadCollection(collectionName)

		if err != nil {
			slog.Warn("Failed to load collection in batch", "name", collectionName, "error", err)
			continue
		}

		collections = append(collections, *coll)
	}

	if len(collections) == 0 {
		return nil, fmt.Errorf("no collection found in %s", cm.config)
	}

	return &collections, nil
}

func (cm *CollectionManager) LoadCollection(collectionName string) (*Collection, error) {
	if collectionName == "" {
		return nil, errors.New("no collection name specified")
	}

	slog.Debug("Loading collection", "name", collectionName)

	fileBytes, err := fs.ReadConfigFile(cm.config, collectionName)

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

	return fs.UpdateConfigFile(cm.config, updated.Name, bytes)
}

func (cm *CollectionManager) DeleteCollection(collectionName string) error {
	if collectionName == "" {
		return errors.New("no collection name specified")
	}

	if err := tools.RemoveConfigFile(cm.config, collectionName); err != nil {
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
