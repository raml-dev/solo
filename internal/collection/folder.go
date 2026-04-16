package collection

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"time"

	"github.com/google/uuid"
)

type Folder struct {
	CreationTimestamp   time.Time `json:"creationTimestamp"`
	LastUpdateTimestamp time.Time `json:"lastUpdateTimestamp"`
	Requests            []Request `json:"requests"`
	Name                string    `json:"name"`
	Id                  string    `json:"id"`
	SubFolders          []Folder  `json:"subFolders"`
}

func NewFolder(name string) Folder {

	tsp := time.Now()
	return Folder{
		Id:                  uuid.NewString(),
		CreationTimestamp:   tsp,
		LastUpdateTimestamp: tsp,
		Name:                name,
		Requests:            make([]Request, 0),
		SubFolders:          make([]Folder, 0),
	}
}

func (f *Folder) GetRequests() *[]Request {
	return &f.Requests
}

func (f *Folder) GetSubFolders() *[]Folder {
	return &f.SubFolders
}

// this function search a request by its id also in all subfolders
func (f *Folder) GetRequestById(id string) (*Request, error) {
	// search in request at first level of the struct
	for i := range f.Requests {
		if f.Requests[i].Id == id {
			return &f.Requests[i], nil
		}
	}

	// recursive check
	for i := range f.SubFolders {
		req, err := f.SubFolders[i].GetRequestById(id)
		if err == nil {
			return req, nil
		}
	}

	return nil, fmt.Errorf("request with id %s does not exist", id)
}

func (f *Folder) AddRequestToFolder(folderId string, request Request) error {
	if request.Id == "" {
		request.Id = uuid.NewString()
	}

	now := time.Now()

	request.CreationTimestamp = now
	request.LastUpdateTimestamp = now

	if f.Id == folderId {
		if f.exists(request.Id) {
			return fmt.Errorf("request with id %s already exists in folder %s", request.Id, folderId)
		}

		f.Requests = append(f.Requests, request)
		f.LastUpdateTimestamp = now
		return nil
	}

	// search recursively for folderId in subfolders
	for i := range f.SubFolders {
		err := f.SubFolders[i].AddRequestToFolder(folderId, request)
		if err == nil {
			f.LastUpdateTimestamp = now
			return nil
		}
	}

	return fmt.Errorf("cannot add request with id %s to the folder %s", request.Id, folderId)
}

func (f *Folder) RemoveRequestFromFolder(folderId string, requestId string) error {
	now := time.Now()

	if f.Id == folderId {
		updatedRequests := slices.DeleteFunc(f.Requests, func(r Request) bool { return r.Id == requestId })
		if len(updatedRequests) != len(f.Requests) {
			f.Requests = updatedRequests
			f.LastUpdateTimestamp = now
			return nil
		}

		return fmt.Errorf("request with id %s does not exist in folder %s", requestId, folderId)
	}

	for i := range f.SubFolders {
		err := f.SubFolders[i].RemoveRequestFromFolder(folderId, requestId)
		if err == nil {
			f.LastUpdateTimestamp = now
			return nil
		}
	}

	return fmt.Errorf("cannot remove request with id %s from folder %s", requestId, folderId)
}

func (f *Folder) UpdateRequestInFolder(folderId string, updated Request) error {
	if updated.Id == "" {
		return errors.New("missing identifier for request")
	}

	now := time.Now()

	if f.Id == folderId {
		for i := range f.Requests {
			if f.Requests[i].Id != updated.Id {
				continue
			}

			vUpdated := reflect.ValueOf(updated)
			vExisting := reflect.ValueOf(&f.Requests[i]).Elem()

			for fieldIdx := 0; fieldIdx < vUpdated.NumField(); fieldIdx++ {
				fieldName := vUpdated.Type().Field(fieldIdx).Name

				if fieldName == "Id" ||
					fieldName == "CreationTimestamp" ||
					fieldName == "LastUpdateTimestamp" {
					continue
				}

				if !vExisting.Field(fieldIdx).CanSet() {
					continue
				}

				updatedField := vUpdated.Field(fieldIdx)
				existingField := vExisting.Field(fieldIdx)

				if fieldName == "Auth" {
					existingField.Set(updatedField)
					continue
				}

				if !reflect.DeepEqual(updatedField.Interface(), existingField.Interface()) {
					existingField.Set(updatedField)
				}
			}

			f.Requests[i].LastUpdateTimestamp = now
			f.LastUpdateTimestamp = now

			return nil
		}

		return fmt.Errorf("request with id %s does not exist in folder %s", updated.Id, folderId)
	}

	for i := range f.SubFolders {
		err := f.SubFolders[i].UpdateRequestInFolder(folderId, updated)
		if err == nil {
			f.LastUpdateTimestamp = now
			return nil
		}
	}

	return fmt.Errorf("cannot update request with id %s in folder %s", updated.Id, folderId)
}

func (f *Folder) exists(requestId string) bool {
	for _, r := range f.Requests {
		if r.Id == requestId {
			return true
		}
	}

	return false
}
