package collection

import (
	"encoding/json"
	"errors"
	"fmt"
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
		return err
	}

	return fs.CreateConfigFile(cm.config, collectionName, bytes)

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
			fmt.Printf("%s", err.Error())
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
	fileBytes, err := fs.ReadConfigFile(cm.config, collectionName)

	if err != nil {
		return nil, err
	}
	var rC Collection

	err = json.Unmarshal(fileBytes, &rC)

	if err != nil {
		return nil, err
	}

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

	return tools.RemoveConfigFile(cm.config, collectionName)

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

func (cm *CollectionManager) AddRequest(collectionName string, request Request) error {
	if collectionName == "" {
		return errors.New("no collection name specified")
	}
	coll, err := cm.LoadCollection(collectionName)
	if err != nil {
		return err
	}

	if err := coll.AddRequest(request); err != nil {
		return err
	}

	return cm.UpdateCollection(*coll)
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

	return cm.UpdateCollection(*coll)
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

	return cm.UpdateCollection(*coll)
}
