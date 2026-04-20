// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package collection

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Collection represents a group of HTTP requests.
// NOTE: This type is not safe for concurrent use.
// If concurrent access is needed in the future, add sync.RWMutex.
type ValueType struct {
	Value string `json:"value"`
	Type  string `json:"type"`
}

type Collection struct {
	CreationTimestamp   time.Time            `json:"creationTimestamp"`
	LastUpdateTimestamp time.Time            `json:"lastUpdateTimestamp"`
	Requests            []Request            `json:"requests"`
	Name                string               `json:"name"`
	Id                  string               `json:"id"`
	Auth                AuthConfiguration    `json:"auth,omitempty"`
	Variables           map[string]ValueType `json:"variables,omitempty"`
	GitRemote           string               `json:"gitRemote,omitempty"`
	GitPath             string               `json:"gitPath,omitempty"`
	GitProvider         string               `json:"gitProvider,omitempty"`
	Folders             []Folder             `json:"folders"`
}

func NewCollection(name string) Collection {
	tsp := time.Now()
	return Collection{
		Id:                  uuid.NewString(),
		CreationTimestamp:   tsp,
		LastUpdateTimestamp: tsp,
		Name:                name,
		Requests:            make([]Request, 0),
		Variables:           make(map[string]ValueType),
		Folders:             make([]Folder, 0),
	}
}

func (c *Collection) GetRequests() *[]Request {
	return &c.Requests
}

func (c *Collection) GetFolders() *[]Folder {
	return &c.Folders
}

func (c *Collection) GetSelectedVariables(keys []string) *map[string]ValueType {
	result := make(map[string]ValueType)
	for _, key := range keys {
		if value, ok := c.Variables[key]; ok {
			result[key] = value
		}
	}
	return &result
}

func (c *Collection) VariableStringValues() map[string]string {
	values := make(map[string]string, len(c.Variables))
	for key, value := range c.Variables {
		values[key] = value.Value
	}
	return values
}

func (c *Collection) ResolveVariableValue(key string, envValues map[string]string) (string, bool) {
	if value, ok := envValues[key]; ok {
		if strings.TrimSpace(value) != "" {
			return value, true
		}
		if collectionValue, ok := c.Variables[key]; ok {
			return collectionValue.Value, true
		}
		return value, true
	}

	if collectionValue, ok := c.Variables[key]; ok {
		return collectionValue.Value, true
	}

	return "", false
}

func (c *Collection) GetRequestById(id string) (*Request, error) {
	_, r := c.get(id)
	if r != nil {
		return r, nil
	}

	for i := range c.Folders {
		req, err := c.Folders[i].GetRequestById(id)
		if err == nil {
			return req, nil
		}
	}

	return nil, fmt.Errorf("request with id %s does not exist", id)

}

func (c *Collection) AddRequest(request Request) (*Request, error) {
	if request.Id == "" {
		request.Id = uuid.NewString()
	}

	now := time.Now()

	request.CreationTimestamp = now
	request.LastUpdateTimestamp = now

	exists := c.exists(request.Id)

	if exists {
		return nil, fmt.Errorf("Request %s with id %s already exists", request.Name, request.Id)
	}

	c.Requests = append(c.Requests, request)
	c.LastUpdateTimestamp = now

	return &request, nil
}

func (c *Collection) RemoveRequest(id string) error {
	exists := c.exists(id)

	if exists {
		// remove request from c.Requests
		requests := slices.DeleteFunc(c.Requests,
			func(r Request) bool { return r.Id == id })

		if len(requests) != len(c.Requests)-1 {
			return fmt.Errorf("error removing request %s", id)
		}

		c.Requests = requests
		c.LastUpdateTimestamp = time.Now()

		return nil
	}

	folderId, found := findFolderIDByRequestID(c.Folders, id)
	if !found {
		return fmt.Errorf("Request with id %s does not exists", id)
	}

	for i := range c.Folders {
		if err := c.Folders[i].RemoveRequestFromFolder(folderId, id); err == nil {
			c.LastUpdateTimestamp = time.Now()
			return nil
		}
	}

	return fmt.Errorf("error removing request %s", id)
}

func (c *Collection) AddRequestToFolder(folderId string, request Request) (*Request, error) {
	if folderId == "" {
		return nil, errors.New("missing folder identifier")
	}

	if request.Id == "" {
		request.Id = uuid.NewString()
	}

	for i := range c.Folders {
		if err := c.Folders[i].AddRequestToFolder(folderId, request); err == nil {
			c.LastUpdateTimestamp = time.Now()
			return c.GetRequestById(request.Id)
		}
	}

	return nil, fmt.Errorf("folder with id %s does not exist", folderId)
}

func (c *Collection) UpdateRequest(updated Request) error {
	if updated.Id == "" {
		return errors.New("missing identifier for request")
	}

	idx, r := c.get(updated.Id)

	if idx != -1 && r != nil {
		vUpdated := reflect.ValueOf(updated)
		vExisting := reflect.ValueOf(r).Elem()

		for i := 0; i < vUpdated.NumField(); i++ {
			fieldName := vUpdated.Type().Field(i).Name

			if fieldName == "Id" ||
				fieldName == "CreationTimestamp" ||
				fieldName == "LastUpdateTimestamp" {
				continue
			}

			if !vExisting.Field(i).CanSet() {
				continue
			}

			updatedField := vUpdated.Field(i)
			existingField := vExisting.Field(i)

			// Special handling for Auth pointer to ensure it's always copied if different
			if fieldName == "Auth" {
				existingField.Set(updatedField)
				continue
			}

			if !reflect.DeepEqual(updatedField.Interface(), existingField.Interface()) {
				existingField.Set(updatedField)
			}
		}

		now := time.Now()
		r.LastUpdateTimestamp = now
		c.LastUpdateTimestamp = now

		c.Requests[idx] = *r

		return nil
	}

	folderId, found := findFolderIDByRequestID(c.Folders, updated.Id)
	if !found {
		return fmt.Errorf("Request with id %s does not exists", updated.Id)
	}

	for i := range c.Folders {
		if err := c.Folders[i].UpdateRequestInFolder(folderId, updated); err == nil {
			c.LastUpdateTimestamp = time.Now()
			return nil
		}
	}

	return fmt.Errorf("Request with id %s does not exists", updated.Id)
}

func (c *Collection) GetFolderById(id string) (*Folder, error) {
	f := findFolderById(c.Folders, id)
	if f == nil {
		return nil, fmt.Errorf("folder with id %s does not exist", id)
	}

	return f, nil
}

func (c *Collection) AddFolder(parentFolderId string, folder Folder) (*Folder, error) {
	if folder.Id == "" {
		folder.Id = uuid.NewString()
	}

	if existing, _ := c.GetFolderById(folder.Id); existing != nil {
		return nil, fmt.Errorf("folder with id %s already exists", folder.Id)
	}

	now := time.Now()
	folder.CreationTimestamp = now
	folder.LastUpdateTimestamp = now

	if parentFolderId == "" {
		c.Folders = append(c.Folders, folder)
	} else {
		added := addSubFolderByParent(&c.Folders, parentFolderId, folder, now)
		if !added {
			return nil, fmt.Errorf("folder with id %s does not exist", parentFolderId)
		}
	}

	c.LastUpdateTimestamp = now

	subFolder, _ := c.GetFolderById(folder.Id)
	return subFolder, nil
}

func (c *Collection) RemoveFolder(folderId string) error {
	now := time.Now()
	removed := removeFolderById(&c.Folders, folderId, now)
	if !removed {
		return fmt.Errorf("folder with id %s does not exist", folderId)
	}

	c.LastUpdateTimestamp = now
	return nil
}

func (c *Collection) UpdateFolder(updated Folder) error {
	if updated.Id == "" {
		return errors.New("missing identifier for folder")
	}

	folder, err := c.GetFolderById(updated.Id)
	if err != nil {
		return err
	}

	vUpdated := reflect.ValueOf(updated)
	vExisting := reflect.ValueOf(folder).Elem()

	for i := 0; i < vUpdated.NumField(); i++ {
		fieldName := vUpdated.Type().Field(i).Name

		if fieldName == "Id" ||
			fieldName == "CreationTimestamp" ||
			fieldName == "LastUpdateTimestamp" {
			continue
		}

		if !vExisting.Field(i).CanSet() {
			continue
		}

		updatedField := vUpdated.Field(i)
		existingField := vExisting.Field(i)

		if !reflect.DeepEqual(updatedField.Interface(), existingField.Interface()) {
			existingField.Set(updatedField)
		}
	}

	now := time.Now()
	folder.LastUpdateTimestamp = now
	c.LastUpdateTimestamp = now

	return nil
}

func findFolderById(folders []Folder, id string) *Folder {
	for i := range folders {
		if folders[i].Id == id {
			return &folders[i]
		}

		f := findFolderById(folders[i].Folders, id)
		if f != nil {
			return f
		}
	}

	return nil
}

func addSubFolderByParent(folders *[]Folder, parentFolderId string, folder Folder, now time.Time) bool {
	for i := range *folders {
		if (*folders)[i].Id == parentFolderId {
			(*folders)[i].Folders = append((*folders)[i].Folders, folder)
			(*folders)[i].LastUpdateTimestamp = now
			return true
		}

		added := addSubFolderByParent(&(*folders)[i].Folders, parentFolderId, folder, now)
		if added {
			(*folders)[i].LastUpdateTimestamp = now
			return true
		}
	}

	return false
}

func removeFolderById(folders *[]Folder, folderId string, now time.Time) bool {
	updatedFolders := slices.DeleteFunc(*folders, func(f Folder) bool { return f.Id == folderId })
	if len(updatedFolders) != len(*folders) {
		*folders = updatedFolders
		return true
	}

	for i := range *folders {
		removed := removeFolderById(&(*folders)[i].Folders, folderId, now)
		if removed {
			(*folders)[i].LastUpdateTimestamp = now
			return true
		}
	}

	return false
}

func findFolderIDByRequestID(folders []Folder, requestId string) (string, bool) {
	for i := range folders {
		for j := range folders[i].Requests {
			if folders[i].Requests[j].Id == requestId {
				return folders[i].Id, true
			}
		}

		folderId, found := findFolderIDByRequestID(folders[i].Folders, requestId)
		if found {
			return folderId, true
		}
	}

	return "", false
}
