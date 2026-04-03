// Copyright 2026-present raml-dev
// SPDX-License-Identifier: GPL-3.0-only

package requester

import (
	"solo/internal/auth"
	"solo/internal/host"
	"solo/internal/script"
	"testing"
)

func TestService_Execute_ContentTypeInjection(goTest *testing.T) {
	// Setup dependencies (minimal for this test)
	hm := host.NewHostManager()
	am := auth.NewAuthManager(goTest.TempDir())
	sm := script.NewScriptManager(nil)
	
	s := NewService(nil, sm, hm, am)

	tests := []struct {
		name           string
		opts           ExecutionOptions
		wantHeader     string
	}{
		{
			name: "inject text/plain when body is present and header is missing",
			opts: ExecutionOptions{
				Method: "POST",
				URL:    "http://example.com",
				Body:   `{"foo":"bar"}`,
				Headers: map[string]any{},
			},
			wantHeader:   "text/plain",
		},
		{
			name: "do not override existing Content-Type",
			opts: ExecutionOptions{
				Method: "POST",
				URL:    "http://example.com",
				Body:   `{"foo":"bar"}`,
				Headers: map[string]any{
					"Content-Type": "application/json",
				},
			},
			wantHeader:   "application/json",
		},
		{
			name: "do not inject when body is empty",
			opts: ExecutionOptions{
				Method: "POST",
				URL:    "http://example.com",
				Body:   "",
				Headers: map[string]any{},
			},
			wantHeader:   "",
		},
		{
			name: "allow pre-request script to override injected Content-Type",
			opts: ExecutionOptions{
				Method: "POST",
				URL:    "http://example.com",
				Body:   `{"foo":"bar"}`,
				Headers: map[string]any{},
				// Lua script to change content type
				PreRequestScript: `request.headers["Content-Type"] = "application/lua"`,
			},
			wantHeader: "application/lua",
		},
	}

	for _, tt := range tests {
		goTest.Run(tt.name, func(t *testing.T) {
			_, req, _ := s.Execute(tt.opts)
			
			if req == nil {
				t.Fatalf("Expected request object to be returned")
			}

			gotHeader := req.Header.Get("Content-Type")
			if gotHeader != tt.wantHeader {
				t.Errorf("Content-Type = %q, want %q", gotHeader, tt.wantHeader)
			}
		})
	}
}
