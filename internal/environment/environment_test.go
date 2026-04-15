// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package environment

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"testing"
)

const (
	VALUE_NUM = 2
)

func TestGetValues(t *testing.T) {

	tests := []struct {
		name        string
		environment Environment
		expected    int
	}{
		{
			name:        "Get some values",
			environment: initEnvironment(),
			expected:    VALUE_NUM,
		},
		{
			name:        "Get no values",
			environment: NewEnvironment("environment"),
			expected:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := tt.environment.GetValues()

			if len(*res) != tt.expected {
				t.Errorf("\"%s\" should return %d instead of %d", tt.name, tt.expected, len(*res))
			}
		})
	}

}

func TestGetValueByName(t *testing.T) {

	tests := []struct {
		name      string
		valueName string
		expected  map[string]string
		error     error
	}{
		{
			name:      "Get value by name",
			valueName: "value_1",
			expected:  map[string]string{"Value": "test_1", "Type": "string"},
			error:     nil,
		},
		{
			name:      "Get non-existent value",
			valueName: "non_existent",
			expected:  map[string]string{"Value": "", "Type": ""},
			error:     nil,
		},
		{
			name:      "Get value with empty name",
			valueName: "",
			expected:  map[string]string{},
			error:     errors.New("name value cannot be empty"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environment := initEnvironment()

			value, err := environment.GetValueByName(tt.valueName)

			if tt.error != nil && err == nil {
				t.Errorf("\"%s\" expect error %s but not error has been raised", tt.name, tt.error.Error())
			} else if tt.error == nil && err != nil {
				t.Errorf("\"%s\" raised error %s but no error was expected", tt.name, err.Error())
			}

			if err == nil {
				r := reflect.ValueOf(value)

				for k, v := range tt.expected {
					structField := r.Elem().FieldByName(k)

					if !structField.IsValid() ||
						structField.Kind() != reflect.String ||
						structField.String() != v {
						t.Errorf("\"%s\" expect %s:%s but found value %v", tt.name, k, v, structField.Interface())
					}
				}
			}

		})
	}

}

func TestAddValue(t *testing.T) {

	tests := []struct {
		name         string
		valueName    string
		valueContent ValueType
		expected     int
		error        error
	}{
		{
			name:         "Add new value",
			valueName:    "new_value",
			valueContent: ValueType{Value: "test", Type: "string"},
			expected:     3,
			error:        nil,
		},
		{
			name:         "Add existent value",
			valueName:    "value_1",
			valueContent: ValueType{Value: "test", Type: "string"},
			expected:     2,
			error:        fmt.Errorf("value with name value_1 already exists"),
		},
		{
			name:         "Add value with empty name",
			valueName:    "",
			valueContent: ValueType{Value: "test", Type: "string"},
			expected:     2,
			error:        errors.New("value name must be specified"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environment := initEnvironment()

			err := environment.AddValue(tt.valueName, tt.valueContent)
			resLen := len(*environment.GetValues())

			if tt.error != nil && err == nil {
				t.Errorf("\"%s\" expect error %s but not error has been raised", tt.name, tt.error.Error())
			} else if tt.error == nil && err != nil {
				t.Errorf("\"%s\" raised error %s but no error was expected", tt.name, err.Error())
			}
			if resLen != tt.expected {
				t.Errorf("\"%s\" expect to have %d values in environment but found %d", tt.name, tt.expected, resLen)
			}
		})
	}
}

func TestRemoveValue(t *testing.T) {

	tests := []struct {
		name      string
		valueName string
		expected  int
		error     error
	}{
		{
			name:      "Remove a value",
			valueName: "value_1",
			expected:  1,
			error:     nil,
		},
		{
			name:      "Remove non-existent value",
			valueName: "non_existent",
			expected:  2,
			error:     fmt.Errorf("value with name non_existent does not exists"),
		},
		{
			name:      "Remove value with empty name",
			valueName: "",
			expected:  2,
			error:     errors.New("value name must be specified"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environment := initEnvironment()

			err := environment.RemoveValue(tt.valueName)
			resLen := len(*environment.GetValues())

			if tt.error != nil && err == nil {
				t.Errorf("\"%s\" expect error %s but not error has been raised", tt.name, tt.error.Error())
			} else if tt.error == nil && err != nil {
				t.Errorf("\"%s\" raised error %s but no error was expected", tt.name, err.Error())
			}

			if resLen != tt.expected {
				t.Errorf("\"%s\" expect to have %d values in environment but found %d", tt.name, tt.expected, resLen)
			}
		})
	}
}

func TestUpdateValue(t *testing.T) {

	tests := []struct {
		name          string
		valueName     string
		valueContent  ValueType
		expectedValue string
		expectedType  string
		expectedLen   int
		error         error
	}{
		{
			name:          "Update existent value",
			valueName:     "value_0",
			valueContent:  ValueType{Value: "updated", Type: "string"},
			expectedValue: "updated",
			expectedType:  "string",
			expectedLen:   2,
			error:         nil,
		},
		{
			name:          "Update value with empty name",
			valueName:     "",
			valueContent:  ValueType{Value: "updated", Type: "string"},
			expectedValue: "test_0",
			expectedType:  "string",
			expectedLen:   2,
			error:         errors.New("value name must be specified"),
		},
		{
			name:          "Update non-existent value",
			valueName:     "non_existent",
			valueContent:  ValueType{Value: "updated", Type: "string"},
			expectedValue: "test_0",
			expectedType:  "string",
			expectedLen:   2,
			error:         fmt.Errorf("value with name non_existent does not exists"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environment := initEnvironment()

			err := environment.UpdateValue(tt.valueName, tt.valueContent)

			resLen := len(*environment.GetValues())

			if tt.error != nil && err == nil {
				t.Errorf("\"%s\" expect error %s but not error has been raised", tt.name, tt.error.Error())
			} else if tt.error == nil && err != nil {
				t.Errorf("\"%s\" raised error %s but no error was expected", tt.name, err.Error())
			}
			if resLen != tt.expectedLen {
				t.Errorf("\"%s\" expect to have %d values in environment but found %d", tt.name, tt.expectedLen, resLen)
			}

			updatedValue := environment.Values["value_0"]

			if updatedValue.Value != tt.expectedValue {
				t.Errorf("expected %s value, found %s", tt.expectedValue, updatedValue.Value)
			}

			if updatedValue.Type != tt.expectedType {
				t.Errorf("expected %s type, found %s", tt.expectedType, updatedValue.Type)
			}

			if environment.Values["value_1"].Value != "test_1" {
				t.Error("Update operation should not affect other values")
			}
		})
	}
}

func TestGetSelectedValues(t *testing.T) {

	tests := []struct {
		name     string
		keys     []string
		expected int
	}{
		{
			name:     "Get selected values",
			keys:     []string{"value_0", "value_1"},
			expected: 2,
		},
		{
			name:     "Get some selected values",
			keys:     []string{"value_0", "non_existent"},
			expected: 1,
		},
		{
			name:     "Get no selected values",
			keys:     []string{"non_existent"},
			expected: 0,
		},
		{
			name:     "Get selected values with empty keys",
			keys:     []string{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environment := initEnvironment()

			res := environment.GetSelectedValues(tt.keys)

			if len(*res) != tt.expected {
				t.Errorf("\"%s\" should return %d instead of %d", tt.name, tt.expected, len(*res))
			}

			for _, k := range tt.keys {
				if val, ok := environment.Values[k]; ok {
					if resVal, exists := (*res)[k]; !exists {
						t.Errorf("\"%s\" expected key %s to be in result", tt.name, k)
					} else if resVal != val {
						t.Errorf("\"%s\" expected value %v for key %s, found %v", tt.name, val, k, resVal)
					}
				}
			}
		})
	}
}

func initEnvironment() Environment {
	environment := NewEnvironment("environment_")

	for j := range VALUE_NUM {
		environment.AddValue("value_"+strconv.Itoa(j), ValueType{
			Value: "test_" + strconv.Itoa(j),
			Type:  "string",
		})
	}
	return environment
}
