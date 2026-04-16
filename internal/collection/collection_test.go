// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package collection

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/google/uuid"
)

const (
	REQUEST_NUM = 2
)

func TestGetRequests(t *testing.T) {

	tests := []struct {
		name       string
		collection Collection
		expected   int
	}{
		{
			name:       "Get some requests",
			collection: initCollection(),
			expected:   REQUEST_NUM,
		},
		{
			name:       "Get no requests",
			collection: NewCollection("collection"),
			expected:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := tt.collection.GetRequests()

			if len(*res) != tt.expected {
				t.Errorf("\"%s\" should return %d instead of %d", tt.name, tt.expected, len(*res))
			}
		})
	}

}

func TestGetRequestById(t *testing.T) {

	tests := []struct {
		name     string
		idIndex  int
		expected map[string]string
		error    error
	}{
		{
			name:     "Get request by id",
			idIndex:  1,
			expected: map[string]string{"Name": "request_1"},
			error:    nil,
		},
		{
			name:     "Specified request does not exists",
			idIndex:  -1,
			expected: map[string]string{},
			error:    fmt.Errorf(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collection := initCollection()

			var testId string
			if tt.idIndex >= 0 {
				testId = (*collection.GetRequests())[tt.idIndex].Id
				tt.expected["Id"] = testId
			} else {
				testId = uuid.NewString()
			}

			req, err := collection.GetRequestById(testId)

			if tt.error != nil && err == nil {
				t.Errorf("\"%s\" expect error %s but not error has been raised", tt.name, tt.error.Error())
			} else if tt.error == nil && err != nil {
				t.Errorf("\"%s\" raised error %s but no error was expected", tt.name, err.Error())
			}

			r := reflect.ValueOf(req)

			for k, v := range tt.expected {
				structField := r.Elem().FieldByName(k)

				if !structField.IsValid() ||
					structField.Kind() != reflect.String ||
					structField.String() != v {
					t.Errorf("\"%s\" expect %s:%s but found value %v", tt.name, k, v, structField.Interface())
				}
			}

		})
	}

}

func TestAddRequest(t *testing.T) {

	tests := []struct {
		name         string
		requestIndex int
		expected     int
		error        error
	}{
		{
			name:         "Add new request",
			requestIndex: -1,
			expected:     3,
			error:        nil,
		},
		{
			name:         "Add existent request",
			requestIndex: 1,
			expected:     2,
			error:        fmt.Errorf(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collection := initCollection()

			var testRequest Request
			if tt.requestIndex >= 0 {
				existingRequest := (*collection.GetRequests())[tt.requestIndex]
				testRequest = Request{Id: existingRequest.Id}
			} else {
				testRequest = Request{Name: uuid.NewString()}
			}

			_, err := collection.AddRequest(testRequest)
			resLen := len(*collection.GetRequests())

			if tt.error != nil && err == nil {
				t.Errorf("\"%s\" expect error %s but not error has been raised", tt.name, tt.error.Error())
			} else if tt.error == nil && err != nil {
				t.Errorf("\"%s\" raised error %s but no error was expected", tt.name, err.Error())
			}
			if resLen != tt.expected {
				t.Errorf("\"%s\" expect to have %d requests in collection but found %d", tt.name, tt.expected, resLen)
			}
		})
	}
}

func TestRemoveRequest(t *testing.T) {

	tests := []struct {
		name     string
		idIndex  int
		expected int
		error    error
	}{
		{
			name:     "Remove a request",
			idIndex:  1,
			expected: 1,
			error:    nil,
		},
		{
			name:     "Remove non-existent request",
			idIndex:  -1,
			expected: 2,
			error:    fmt.Errorf(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collection := initCollection()

			var testId string
			if tt.idIndex >= 0 {
				testId = (*collection.GetRequests())[tt.idIndex].Id
			} else {
				testId = uuid.NewString()
			}

			err := collection.RemoveRequest(testId)
			resLen := len(*collection.GetRequests())

			if tt.error != nil && err == nil {
				t.Errorf("\"%s\" expect error %s but not error has been raised", tt.name, tt.error.Error())
			} else if tt.error == nil && err != nil {
				t.Errorf("\"%s\" raised error %s but no error was expected", tt.name, err.Error())
			}

			if resLen != tt.expected {
				t.Errorf("\"%s\" expect to have %d requests in collection but found %d", tt.name, tt.expected, resLen)
			}
		})
	}
}

func TestUpdateRequest(t *testing.T) {

	tests := []struct {
		name           string
		request        Request
		expectedName   string
		expectedLength int
		error          error
	}{
		{
			name:           "Update existent request",
			request:        Request{Name: "new-name"},
			expectedName:   "new-name",
			expectedLength: 2,
			error:          nil,
		},
		{
			name:           "Update request without id",
			request:        Request{Name: "new-name"},
			expectedName:   "request_0",
			expectedLength: 2,
			error:          errors.New("missing identifier for request"),
		},
		{
			name:           "Update non-existent request",
			request:        Request{Id: "test", Name: "new-name"},
			expectedName:   "request_0",
			expectedLength: 2,
			error:          errors.New("Request with id test does not exists"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collection := initCollection()
			originalRequest := (*collection.GetRequests())[0]

			if tt.name == "Update existent request" {
				tt.request.Id = originalRequest.Id
			}

			err := collection.UpdateRequest(tt.request)

			resLen := len(*collection.GetRequests())

			if tt.error != nil && err == nil {
				t.Errorf("\"%s\" expect error %s but not error has been raised", tt.name, tt.error.Error())
			} else if tt.error == nil && err != nil {
				t.Errorf("\"%s\" raised error %s but no error was expected", tt.name, err.Error())
			}
			if resLen != tt.expectedLength {
				t.Errorf("\"%s\" expect to have %d requests in collection but found %d", tt.name, tt.expectedLength, resLen)
			}

			updatedRequest := (*collection.GetRequests())[0]

			if updatedRequest.Name != tt.expectedName {
				t.Errorf("expected %s name, found %s", tt.expectedName, updatedRequest.Name)
			}

			if updatedRequest.Id != originalRequest.Id {
				t.Errorf("expected Id %s to remain unchanged, but found %s", originalRequest.Id, updatedRequest.Id)
			}

			if (*collection.GetRequests())[1].Name != "request_1" {
				t.Error("Update operation should not affect other requests")
			}
		})
	}
}

func TestAddFolder(t *testing.T) {
	collection := NewCollection("collection")

	added, err := collection.AddFolder(Folder{Id: "f-1", Name: "folder-1"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if added == nil {
		t.Fatal("expected folder, got nil")
	}

	if len(*collection.GetFolders()) != 1 {
		t.Fatalf("expected 1 folder, got %d", len(*collection.GetFolders()))
	}
}

func TestGetFolderById(t *testing.T) {
	collection := NewCollection("collection")
	collection.Folders = []Folder{
		{
			Id: "f-1",
			SubFolders: []Folder{
				{Id: "f-2"},
			},
		},
	}

	folder, err := collection.GetFolderById("f-2")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if folder == nil || folder.Id != "f-2" {
		t.Fatalf("expected folder f-2, got %+v", folder)
	}
}

func TestAddSubFolder(t *testing.T) {
	collection := NewCollection("collection")
	collection.Folders = []Folder{
		{Id: "f-1"},
	}

	added, err := collection.AddSubFolder("f-1", Folder{Id: "f-2", Name: "sub"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if added == nil || added.Id != "f-2" {
		t.Fatalf("expected subfolder f-2, got %+v", added)
	}

	if len(collection.Folders[0].SubFolders) != 1 {
		t.Fatalf("expected 1 subfolder, got %d", len(collection.Folders[0].SubFolders))
	}
}

func TestRemoveFolder(t *testing.T) {
	collection := NewCollection("collection")
	collection.Folders = []Folder{
		{
			Id: "f-1",
			SubFolders: []Folder{
				{Id: "f-2"},
			},
		},
	}

	err := collection.RemoveFolder("f-2")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(collection.Folders[0].SubFolders) != 0 {
		t.Fatalf("expected 0 subfolders, got %d", len(collection.Folders[0].SubFolders))
	}
}

func TestUpdateFolder(t *testing.T) {
	collection := NewCollection("collection")
	collection.Folders = []Folder{
		{
			Id:   "f-1",
			Name: "old-name",
			SubFolders: []Folder{
				{
					Id:   "f-2",
					Name: "old-sub",
				},
			},
		},
	}

	err := collection.UpdateFolder(Folder{Id: "f-2", Name: "new-sub"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	updated, err := collection.GetFolderById("f-2")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if updated.Name != "new-sub" {
		t.Fatalf("expected name new-sub, got %s", updated.Name)
	}
}

func initCollection() Collection {
	collection := NewCollection("collection_")

	for j := range REQUEST_NUM {
		collection.AddRequest(Request{Name: "request_" + strconv.Itoa(j)})
	}
	return collection
}
