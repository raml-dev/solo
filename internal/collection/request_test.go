// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package collection

import (
	"encoding/json"
	"slices"
	"sort"
	"testing"
)

func TestRequest_GetPlaceholders_IncludesParams(t *testing.T) {
	tests := []struct {
		name string
		req  Request
		want []string
	}{
		{
			name: "params values are scanned for placeholders",
			req: Request{
				Url:    "http://example.com",
				Params: map[string]string{"key": "{{param_val}}"},
			},
			want: []string{"param_val"},
		},
		{
			name: "placeholders deduplicated across url, headers, cookies, params",
			req: Request{
				Url:     "http://example.com/{{shared}}",
				Headers: map[string]string{"X-Val": "{{shared}}"},
				Params:  map[string]string{"p": "{{shared}}"},
			},
			want: []string{"shared"},
		},
		{
			name: "nil params does not panic",
			req: Request{
				Url:    "http://example.com/{{only_url}}",
				Params: nil,
			},
			want: []string{"only_url"},
		},
		{
			name: "empty params does not panic",
			req: Request{
				Url:    "http://example.com",
				Params: map[string]string{},
			},
			want: []string{},
		},
		{
			name: "multiple distinct placeholders in params",
			req: Request{
				Params: map[string]string{
					"a": "{{alpha}}",
					"b": "{{beta}}",
				},
			},
			want: []string{"alpha", "beta"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.req.GetPlaceholders()
			sort.Strings(got)
			sort.Strings(tt.want)

			if len(got) != len(tt.want) {
				t.Fatalf("GetPlaceholders() = %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("GetPlaceholders()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestRequest_JSON_ParamsOmitEmpty(t *testing.T) {
	// A request without params must not emit the "params" key in JSON.
	req := Request{
		Id:   "r1",
		Name: "no-params",
		Url:  "http://example.com",
		Verb: "GET",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if _, found := raw["params"]; found {
		t.Error("expected 'params' key to be absent when Params is nil, but it was present")
	}
}

func TestRequest_JSON_ParamsSerialisedWhenPresent(t *testing.T) {
	req := Request{
		Id:     "r2",
		Name:   "with-params",
		Url:    "http://example.com",
		Verb:   "GET",
		Params: map[string]string{"foo": "bar"},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded Request
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if v, ok := decoded.Params["foo"]; !ok || v != "bar" {
		t.Errorf("Params[foo] = %q, want %q", v, "bar")
	}
}

func TestRequest_JSON_BackwardCompatibility(t *testing.T) {
	// Old JSON without the "params" field must deserialise without errors
	// and yield a nil Params map.
	oldJSON := `{"id":"r3","name":"old","url":"http://example.com","verb":"GET","headers":{},"cookies":{}}`

	var req Request
	if err := json.Unmarshal([]byte(oldJSON), &req); err != nil {
		t.Fatalf("unmarshal old JSON error: %v", err)
	}

	if req.Params != nil {
		t.Errorf("expected nil Params for old JSON without params field, got %v", req.Params)
	}

	// GetPlaceholders must still work on the old-format request.
	placeholders := req.GetPlaceholders()
	if !slices.Equal(placeholders, []string{}) && len(placeholders) != 0 {
		t.Errorf("expected no placeholders, got %v", placeholders)
	}
}
