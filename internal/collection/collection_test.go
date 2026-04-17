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

func initCollection() Collection {
	collection := NewCollection("collection_")

	for j := range REQUEST_NUM {
		collection.AddRequest(Request{Name: "request_" + strconv.Itoa(j)})
	}
	return collection
}
