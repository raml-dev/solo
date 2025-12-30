package tools

import "testing"

var tests = []struct {
	name     string
	input    string
	expected []string
}{
	{
		name:     "Host placeholder retrieved",
		input:    "https://{{host}}/api/path",
		expected: []string{"host"},
	},
	{
		name:     "Host placeholder retrieved, avoiding duplicates",
		input:    "https://{{host}}/api/{{host}}",
		expected: []string{"host"},
	},
	{
		name:     "No placeholder retrieved",
		input:    "https://host/api",
		expected: []string{},
	},
	{
		name:     "No placeholder retrieved",
		input:    "https://{{host]/api",
		expected: []string{},
	},
}

func TestFlagParser(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractPlaceholders(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("test %s fails: got %v, expected %v", tt.name, result, tt.expected)
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("test %s fails: got %v, expected %v", tt.name, result, tt.expected)
					return
				}
			}

		})
	}
}
