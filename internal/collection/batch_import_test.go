// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package collection

import (
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
)

func TestImportBatchGlobalErrors(t *testing.T) {
	importOne := func(path string) (*Collection, []string, error) {
		coll := NewCollection(path)
		return &coll, nil, nil
	}

	tests := []struct {
		name      string
		manager   *CollectionManager
		paths     []string
		importOne func(string) (*Collection, []string, error)
		wantError string
	}{
		{
			name:      "nil manager",
			manager:   nil,
			paths:     []string{"valid"},
			importOne: importOne,
			wantError: "collection manager not initialized",
		},
		{
			name:      "nil import function",
			manager:   setupTestManager(t),
			paths:     []string{"valid"},
			importOne: nil,
			wantError: "collection import function not provided",
		},
		{
			name:      "empty path list",
			manager:   setupTestManager(t),
			paths:     []string{},
			importOne: importOne,
			wantError: "no collection import paths provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.manager != nil {
				defer cleanupTestDir(tt.manager.config)
			}

			result, err := tt.manager.ImportBatch(tt.paths, tt.importOne)
			if err == nil {
				t.Fatalf("expected global error containing %q, got nil", tt.wantError)
			}
			if !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("expected global error containing %q, got %q", tt.wantError, err.Error())
			}
			if len(result.Results) != 0 {
				t.Fatalf("expected no per-item results for global error, got %+v", result.Results)
			}
		})
	}
}

func TestImportBatchPerItemResults(t *testing.T) {
	cm := setupTestManager(t)
	defer cleanupTestDir(cm.config)

	var importCalls atomic.Int32
	importOne := func(path string) (*Collection, []string, error) {
		importCalls.Add(1)

		switch path {
		case "first":
			coll := NewCollection("imported-first")
			coll.Requests = []Request{{Name: "first-request"}}
			return &coll, []string{"first warning"}, nil
		case "bad":
			return nil, nil, fmt.Errorf("bad import")
		case "nil-collection":
			return nil, nil, nil
		case "save-error":
			return &Collection{Name: ""}, nil, nil
		case "last":
			coll := NewCollection("imported-last")
			coll.Requests = []Request{{Name: "last-request"}}
			return &coll, nil, nil
		default:
			return nil, nil, fmt.Errorf("unexpected path %s", path)
		}
	}

	paths := []string{"first", "", "bad", "nil-collection", "save-error", "last"}
	result, err := cm.ImportBatch(paths, importOne)
	if err != nil {
		t.Fatalf("expected no global error, got %v", err)
	}

	if got := int(importCalls.Load()); got != len(paths)-1 {
		t.Fatalf("expected empty path to skip import function; got %d calls", got)
	}

	if len(result.Results) != len(paths) {
		t.Fatalf("expected %d results, got %d", len(paths), len(result.Results))
	}

	for index, path := range paths {
		if result.Results[index].Path != path {
			t.Fatalf("result %d path = %q, want %q", index, result.Results[index].Path, path)
		}
	}

	expected := []struct {
		path        string
		name        string
		success     bool
		errorSubstr string
		warnings    []string
	}{
		{path: "first", name: "imported-first", success: true, warnings: []string{"first warning"}},
		{path: "", errorSubstr: "path is empty"},
		{path: "bad", errorSubstr: "bad import"},
		{path: "nil-collection", errorSubstr: "import returned no collection"},
		{path: "save-error", errorSubstr: "collection name is not specified"},
		{path: "last", name: "imported-last", success: true},
	}

	for index, want := range expected {
		got := result.Results[index]
		if got.Path != want.path {
			t.Fatalf("result %d path = %q, want %q", index, got.Path, want.path)
		}
		if got.Name != want.name {
			t.Fatalf("result %d name = %q, want %q", index, got.Name, want.name)
		}
		if got.Success != want.success {
			t.Fatalf("result %d success = %v, want %v", index, got.Success, want.success)
		}
		if want.errorSubstr == "" && got.Error != "" {
			t.Fatalf("result %d error = %q, want empty", index, got.Error)
		}
		if want.errorSubstr != "" && !strings.Contains(got.Error, want.errorSubstr) {
			t.Fatalf("result %d error = %q, want substring %q", index, got.Error, want.errorSubstr)
		}
		if len(got.Warnings) != len(want.warnings) {
			t.Fatalf("result %d warnings = %+v, want %+v", index, got.Warnings, want.warnings)
		}
		for warningIndex, warning := range want.warnings {
			if got.Warnings[warningIndex] != warning {
				t.Fatalf("result %d warning %d = %q, want %q", index, warningIndex, got.Warnings[warningIndex], warning)
			}
		}
	}

	first, err := cm.LoadCollection("imported-first")
	if err != nil {
		t.Fatalf("expected first successful import to be saved: %v", err)
	}
	if len(first.Requests) != 1 || first.Requests[0].Name != "first-request" {
		t.Fatalf("unexpected first saved collection requests: %+v", first.Requests)
	}

	last, err := cm.LoadCollection("imported-last")
	if err != nil {
		t.Fatalf("expected later successful import to be saved despite earlier failures: %v", err)
	}
	if len(last.Requests) != 1 || last.Requests[0].Name != "last-request" {
		t.Fatalf("unexpected last saved collection requests: %+v", last.Requests)
	}

	for _, name := range []string{"bad", "nil-collection", "save-error"} {
		if _, err := cm.LoadCollection(name); err == nil {
			t.Fatalf("expected failed import %q not to be saved", name)
		}
	}
}
