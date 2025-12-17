package collections

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"time"
	"yapla/internal/configuration"

	"github.com/google/uuid"
)

type Collection struct {
	creationTimestamp   time.Time
	lastUpdateTimestamp time.Time
	Requests            []Request
	Name                string
	Id                  string
	fsPath              string
}

type Request struct {
	Name                string
	Url                 string
	Verb                string
	Body                string
	Id                  string
	Headers             map[string]any
	Cookies             map[string]any
	creationTimestamp   time.Time
	lastUpdateTimestamp time.Time
}

func NewCollection(config *configuration.Configuration, name string) *Collection {
	tsp := time.Now()
	return &Collection{
		Id:                  uuid.NewString(),
		creationTimestamp:   tsp,
		lastUpdateTimestamp: tsp,
		Name:                name,
		fsPath:              config.Collection.Path,
	}
}

func (c *Collection) GetRequests() *[]Request {
	return &c.Requests
}

func (c *Collection) GetRequestById(id string) (*Request, error) {
	_, r := c.get(id)
	if r == nil {
		return nil, fmt.Errorf("request with id %s does not exist", id)
	}
	return r, nil

}

func (c *Collection) AddRequest(request Request) error {
	if request.Id == "" {
		request.Id = uuid.NewString()
	}

	now := time.Now()

	request.creationTimestamp = now
	request.lastUpdateTimestamp = now

	exists := c.exists(request.Id)

	if exists {
		return fmt.Errorf("Request %s with id %s already exists", request.Name, request.Id)
	}

	c.Requests = append(c.Requests, request)

	return nil
}

func (c *Collection) RemoveRequest(id string) error {
	exists := c.exists(id)

	if !exists {
		return fmt.Errorf("Request with id %s does not exists", id)
	}

	// remove request from c.Requests
	requests := slices.DeleteFunc(c.Requests,
		func(r Request) bool { return r.Id == id })

	if len(requests) != len(c.Requests)-1 {
		return fmt.Errorf("error removing request %s", id)
	}

	c.Requests = requests
	c.lastUpdateTimestamp = time.Now()

	return nil
}

func (c *Collection) UpdateRequest(updated Request) error {
	if updated.Id == "" {
		return errors.New("missing identifier for request")
	}

	idx, r := c.get(updated.Id)

	if idx == -1 || r == nil {
		return fmt.Errorf("Request with id %s does not exists", updated.Id)
	}

	vUpdated := reflect.ValueOf(updated)
	vExisting := reflect.ValueOf(r).Elem()

	for i := 0; i < vUpdated.NumField(); i++ {
		fieldName := vUpdated.Type().Field(i).Name

		if fieldName == "Id" ||
			fieldName == "creationTimestamp" ||
			fieldName == "lastUpdateTimestamp" {
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
	r.lastUpdateTimestamp = now
	c.lastUpdateTimestamp = now

	c.Requests[idx] = *r

	return nil
}

// Utilities
func (c *Collection) exists(id string) bool {
	return slices.ContainsFunc(c.Requests,
		func(r Request) bool {
			return r.Id == id
		})
}

func (c *Collection) get(id string) (int, *Request) {
	for i, r := range c.Requests {
		if r.Id == id {
			return i, &c.Requests[i]
		}
	}
	return -1, nil
}
