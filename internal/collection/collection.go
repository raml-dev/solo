package collection

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"time"

	"github.com/google/uuid"
)

// Collection represents a group of HTTP requests.
// NOTE: This type is not safe for concurrent use.
// If concurrent access is needed in the future, add sync.RWMutex.
type Collection struct {
	CreationTimestamp   time.Time `json:"creationTimestamp"`
	LastUpdateTimestamp time.Time `json:"lastUpdateTimestamp"`
	Requests            []Request `json:"requests"`
	Name                string    `json:"name"`
	Id                  string    `json:"id"`
}

func NewCollection(name string) Collection {
	tsp := time.Now()
	return Collection{
		Id:                  uuid.NewString(),
		CreationTimestamp:   tsp,
		LastUpdateTimestamp: tsp,
		Name:                name,
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

	request.CreationTimestamp = now
	request.LastUpdateTimestamp = now

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
	c.LastUpdateTimestamp = time.Now()

	return nil
}

func (c Collection) UpdateRequest(updated Request) error {
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
	r.LastUpdateTimestamp = now
	c.LastUpdateTimestamp = now

	c.Requests[idx] = *r

	return nil
}
