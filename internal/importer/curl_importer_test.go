// Copyright 2026-present raml-dev
// SPDX-License-Identifier: GPL-3.0-only

package importer

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestCurlImporter_HappyPath(t *testing.T) {
	imp := NewCurlImporter()

	tests := []struct {
		name        string
		input       string
		wantVerb    string
		wantURL     string
		wantBody    string
		wantBT      string // bodyType
		wantName    string
		wantHeaders map[string]string
		wantCookies map[string]string
	}{
		{
			name:     "simple_get",
			input:    "curl https://api.example.com/users",
			wantVerb: "GET",
			wantURL:  "https://api.example.com/users",
			wantName: "GET /users",
		},
		{
			name:     "explicit_delete",
			input:    "curl -X DELETE https://api.example.com/users/1",
			wantVerb: "DELETE",
			wantURL:  "https://api.example.com/users/1",
			wantName: "DELETE /users/1",
		},
		{
			name:        "post_with_json_body",
			input:       `curl -X POST -H 'Content-Type: application/json' -d '{"name":"Alice"}' https://api.example.com/users`,
			wantVerb:    "POST",
			wantURL:     "https://api.example.com/users",
			wantBody:    `{"name":"Alice"}`,
			wantBT:      "json",
			wantName:    "POST /users",
			wantHeaders: map[string]string{"Content-Type": "application/json"},
		},
		{
			name:     "post_inferred_from_data",
			input:    "curl -d 'x=1' https://api.example.com/submit",
			wantVerb: "POST",
			wantURL:  "https://api.example.com/submit",
			wantBody: "x=1",
			wantBT:   "text",
		},
		{
			name:     "text_body",
			input:    "curl -X POST -d 'plain text' https://api.example.com/log",
			wantVerb: "POST",
			wantBody: "plain text",
			wantBT:   "text",
		},
		{
			name:     "json_array_body",
			input:    `curl -X POST -d '[1,2,3]' https://api.example.com/items`,
			wantVerb: "POST",
			wantBody: "[1,2,3]",
			wantBT:   "json",
		},
		{
			name:  "basic_auth",
			input: "curl -u alice:secret https://api.example.com/me",
			wantHeaders: map[string]string{
				"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte("alice:secret")),
			},
		},
		{
			name:  "basic_auth_overridden_by_header",
			input: `curl -u alice:secret -H 'Authorization: Bearer tok' https://api.example.com/me`,
			wantHeaders: map[string]string{
				"Authorization": "Bearer tok",
			},
		},
		{
			name:        "cookies",
			input:       "curl -b 'session=abc; token=xyz' https://api.example.com/dashboard",
			wantCookies: map[string]string{"session": "abc", "token": "xyz"},
		},
		{
			name:    "url_flag",
			input:   "curl --url https://api.example.com/items",
			wantURL: "https://api.example.com/items",
		},
		{
			name:     "data_raw",
			input:    `curl --data-raw '{"a":1}' https://api.example.com/x`,
			wantBody: `{"a":1}`,
			wantBT:   "json",
		},
		{
			name:     "data_binary",
			input:    "curl --data-binary 'not json' https://api.example.com/x",
			wantBody: "not json",
			wantBT:   "text",
		},
		{
			name:     "multiline",
			input:    "curl \\\n  -X POST \\\n  https://api.example.com/users",
			wantVerb: "POST",
			wantURL:  "https://api.example.com/users",
		},
		{
			name:     "request_name_from_path",
			input:    "curl https://api.example.com/users/42",
			wantName: "GET /users/42",
		},
		{
			name:  "quoted_header_double",
			input: `curl -H "Authorization: Bearer tok" https://api.example.com/x`,
			wantHeaders: map[string]string{
				"Authorization": "Bearer tok",
			},
		},
		{
			name:     "ignored_flags",
			input:    "curl -s -v -L --compressed https://api.example.com/ping",
			wantVerb: "GET",
			wantURL:  "https://api.example.com/ping",
		},
		{
			name:        "long_flags",
			input:       `curl --request POST --header 'X-Foo: bar' --data '{}' --url https://api.example.com/x`,
			wantVerb:    "POST",
			wantURL:     "https://api.example.com/x",
			wantBody:    "{}",
			wantBT:      "json",
			wantHeaders: map[string]string{"X-Foo": "bar"},
		},
		{
			name:     "curl_uppercase",
			input:    "CURL https://api.example.com/ping",
			wantVerb: "GET",
			wantURL:  "https://api.example.com/ping",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := imp.ParseRequest(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantVerb != "" && req.Verb != tt.wantVerb {
				t.Errorf("verb: got %q, want %q", req.Verb, tt.wantVerb)
			}
			if tt.wantURL != "" && req.Url != tt.wantURL {
				t.Errorf("url: got %q, want %q", req.Url, tt.wantURL)
			}
			if tt.wantBody != "" && req.Body != tt.wantBody {
				t.Errorf("body: got %q, want %q", req.Body, tt.wantBody)
			}
			if tt.wantBT != "" && req.BodyType != tt.wantBT {
				t.Errorf("bodyType: got %q, want %q", req.BodyType, tt.wantBT)
			}
			if tt.wantName != "" && req.Name != tt.wantName {
				t.Errorf("name: got %q, want %q", req.Name, tt.wantName)
			}
			for k, v := range tt.wantHeaders {
				if got, ok := req.Headers[k]; !ok {
					t.Errorf("header %q: missing", k)
				} else if got != v {
					t.Errorf("header %q: got %q, want %q", k, got, v)
				}
			}
			for k, v := range tt.wantCookies {
				if got, ok := req.Cookies[k]; !ok {
					t.Errorf("cookie %q: missing", k)
				} else if got != v {
					t.Errorf("cookie %q: got %q, want %q", k, got, v)
				}
			}
		})
	}
}

func TestCurlImporter_Metadata(t *testing.T) {
	imp := NewCurlImporter()
	req, err := imp.ParseRequest("curl https://api.example.com/ping")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Id == "" {
		t.Error("request ID should not be empty")
	}
	if req.CreationTimestamp.IsZero() {
		t.Error("CreationTimestamp should not be zero")
	}
	if req.LastUpdateTimestamp.IsZero() {
		t.Error("LastUpdateTimestamp should not be zero")
	}
	if req.Headers == nil {
		t.Error("Headers map should not be nil")
	}
	if req.Cookies == nil {
		t.Error("Cookies map should not be nil")
	}
}

func TestCurlImporter_ErrorCases(t *testing.T) {
	imp := NewCurlImporter()

	tests := []struct {
		name        string
		input       string
		wantErrFrag string
	}{
		{
			name:        "not_curl",
			input:       "wget https://api.example.com/users",
			wantErrFrag: "not a curl command",
		},
		{
			name:        "empty_string",
			input:       "",
			wantErrFrag: "not a curl command",
		},
		{
			name:        "no_url",
			input:       "curl -X GET",
			wantErrFrag: "no URL",
		},
		{
			name:        "unterminated_single_quote",
			input:       "curl 'https://api.example.com/users",
			wantErrFrag: "unterminated single quote",
		},
		{
			name:        "unterminated_double_quote",
			input:       `curl "https://api.example.com/users`,
			wantErrFrag: "unterminated double quote",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := imp.ParseRequest(tt.input)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.wantErrFrag)) {
				t.Errorf("error %q should contain %q", err.Error(), tt.wantErrFrag)
			}
		})
	}
}

// ── tokenizer unit tests ───────────────────────────────────────────────────────

func TestTokenizeCurl(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "simple",
			input: "curl -X GET https://example.com",
			want:  []string{"curl", "-X", "GET", "https://example.com"},
		},
		{
			name:  "single_quotes",
			input: "curl -d 'hello world' https://x.com",
			want:  []string{"curl", "-d", "hello world", "https://x.com"},
		},
		{
			name:  "double_quotes",
			input: `curl -H "Content-Type: application/json" https://x.com`,
			want:  []string{"curl", "-H", "Content-Type: application/json", "https://x.com"},
		},
		{
			name:  "double_quote_escape",
			input: `curl -d "{\"key\":\"val\"}" https://x.com`,
			want:  []string{"curl", "-d", `{"key":"val"}`, "https://x.com"},
		},
		{
			name:  "line_continuation",
			input: "curl \\\n  -X POST \\\n  https://x.com",
			want:  []string{"curl", "-X", "POST", "https://x.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tokenizeCurl(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("token count: got %d %v, want %d %v", len(got), got, len(tt.want), tt.want)
			}
			for i, w := range tt.want {
				if got[i] != w {
					t.Errorf("token[%d]: got %q, want %q", i, got[i], w)
				}
			}
		})
	}
}
