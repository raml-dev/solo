package environment

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Environment struct {
	Id                  string               `json:"id"`
	Name                string               `json:"name"`
	Values              map[string]ValueType `json:"values"`
	CreationTimestamp   time.Time            `json:"creation_timestamp"`
	LastUpdateTimestamp time.Time            `json:"last_update_timestamp"`
}

type ValueType struct {
	Value string `json:"value"`
	Type  string `json:"type"`
}

func NewEnvironment(name string) Environment {
	currentTimestamp := time.Now()
	return Environment{
		Id:                  uuid.NewString(),
		Name:                name,
		Values:              make(map[string]ValueType),
		CreationTimestamp:   currentTimestamp,
		LastUpdateTimestamp: currentTimestamp,
	}
}

func (e *Environment) GetValues() *map[string]ValueType {
	return &e.Values
}
func (e *Environment) GetValueByName(name string) (*ValueType, error) {

	if name == "" {
		return nil, errors.New("name value cannot be empty")
	}

	value := e.Values[name]
	return &value, nil
}
func (e *Environment) AddValue(name string, value ValueType) error {

	if name == "" {
		return errors.New("value name must be specified")
	}

	if _, ok := e.Values[name]; ok {
		return fmt.Errorf("value with name %s already exists", name)
	}

	e.Values[name] = value

	return nil
}
func (e *Environment) RemoveValue(name string) error {
	if name == "" {
		return errors.New("value name must be specified")
	}

	if _, ok := e.Values[name]; !ok {
		return fmt.Errorf("value with name %s does not exists", name)
	}

	delete(e.Values, name)

	return nil
}
func (e *Environment) UpdateValue(name string, value ValueType) error {
	if name == "" {
		return errors.New("value name must be specified")
	}

	if _, ok := e.Values[name]; !ok {
		return fmt.Errorf("value with name %s does not exists", name)
	}

	e.Values[name] = value

	return nil
}
