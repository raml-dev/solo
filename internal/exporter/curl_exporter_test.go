// Copyright 2026-present raml-dev
// SPDX-License-Identifier: GPL-3.0-only

package exporter

import (
	"strings"
	"testing"
)

func TestGenerateCurl(t *testing.T) {
	tests := []struct {
		name          string
		input         CurlExportRequest
		wantFragments []string // all must be present in output
		wantAbsent    []string // none must be present in output
	}{
		{
			name: "simple_get",
			input: CurlExportRequest{
				Method: "GET",
				URL:    "https://api.example.com/users",
			},
			wantFragments: []string{
				"curl \\",
				"-X GET",
				"'https://api.example.com/users'",
			},
			wantAbsent: []string{"-d"},
		},
		{
			name: "post_with_json_body",
			input: CurlExportRequest{
				Method:  "POST",
				URL:     "https://api.example.com/users",
				Headers: map[string]string{"Content-Type": "application/json"},
				Body:    `{"name":"Alice"}`,
			},
			wantFragments: []string{
				"-X POST",
				"-H 'Content-Type: application/json'",
				`-d '{"name":"Alice"}'`,
			},
		},
		{
			name: "multiple_headers_sorted",
			input: CurlExportRequest{
				Method: "GET",
				URL:    "https://api.example.com/x",
				Headers: map[string]string{
					"Z-Last":        "z",
					"A-First":       "a",
					"M-Middle":      "m",
				},
			},
			wantFragments: []string{
				"-H 'A-First: a'",
				"-H 'M-Middle: m'",
				"-H 'Z-Last: z'",
			},
		},
		{
			name: "cookie_header",
			input: CurlExportRequest{
				Method:  "GET",
				URL:     "https://api.example.com/dash",
				Headers: map[string]string{"Cookie": "session=abc; token=xyz"},
			},
			wantFragments: []string{
				"-H 'Cookie: session=abc; token=xyz'",
			},
		},
		{
			name: "body_with_single_quote",
			input: CurlExportRequest{
				Method: "POST",
				URL:    "https://api.example.com/x",
				Body:   `it's here`,
			},
			wantFragments: []string{
				`-d 'it'\''s here'`,
			},
		},
		{
			name: "url_with_single_quote",
			input: CurlExportRequest{
				Method: "GET",
				URL:    "https://api.example.com/path?q=it's",
			},
			wantFragments: []string{
				`'https://api.example.com/path?q=it'\''s'`,
			},
		},
		{
			name: "no_body",
			input: CurlExportRequest{
				Method: "DELETE",
				URL:    "https://api.example.com/users/1",
			},
			wantAbsent: []string{"-d"},
		},
		{
			name: "no_headers",
			input: CurlExportRequest{
				Method: "GET",
				URL:    "https://api.example.com/ping",
			},
			wantAbsent: []string{"-H"},
		},
		{
			name: "header_value_with_single_quote",
			input: CurlExportRequest{
				Method:  "GET",
				URL:     "https://api.example.com/x",
				Headers: map[string]string{"X-Note": "it's fine"},
			},
			wantFragments: []string{
				`-H 'X-Note: it'\''s fine'`,
			},
		},
		{
			name: "empty_body_not_included",
			input: CurlExportRequest{
				Method: "POST",
				URL:    "https://api.example.com/x",
				Body:   "",
			},
			wantAbsent: []string{"-d"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateCurl(tt.input)

			for _, frag := range tt.wantFragments {
				if !strings.Contains(got, frag) {
					t.Errorf("expected output to contain %q\nfull output:\n%s", frag, got)
				}
			}
			for _, frag := range tt.wantAbsent {
				if strings.Contains(got, frag) {
					t.Errorf("expected output NOT to contain %q\nfull output:\n%s", frag, got)
				}
			}
		})
	}
}

func TestGenerateCurl_Deterministic(t *testing.T) {
	req := CurlExportRequest{
		Method: "POST",
		URL:    "https://api.example.com/users",
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer tok",
			"X-Request-Id":  "123",
		},
		Body: `{"name":"Alice"}`,
	}

	first := GenerateCurl(req)
	for i := 0; i < 20; i++ {
		if got := GenerateCurl(req); got != first {
			t.Fatalf("non-deterministic output on iteration %d\nfirst:\n%s\ngot:\n%s", i, first, got)
		}
	}
}

func TestGenerateCurl_MultilineFormat(t *testing.T) {
	req := CurlExportRequest{
		Method:  "POST",
		URL:     "https://api.example.com/users",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    `{"name":"Alice"}`,
	}

	got := GenerateCurl(req)
	lines := strings.Split(got, "\n")

	// All lines except the last must end with " \"
	for i, line := range lines[:len(lines)-1] {
		if !strings.HasSuffix(line, "\\") {
			t.Errorf("line %d should end with '\\', got: %q", i, line)
		}
	}

	// Last line must NOT end with " \"
	last := lines[len(lines)-1]
	if strings.HasSuffix(last, "\\") {
		t.Errorf("last line should not end with '\\', got: %q", last)
	}
}

func TestGenerateCurl_HeaderOrder(t *testing.T) {
	req := CurlExportRequest{
		Method: "GET",
		URL:    "https://api.example.com/x",
		Headers: map[string]string{
			"Z-Header": "z",
			"A-Header": "a",
			"M-Header": "m",
		},
	}

	got := GenerateCurl(req)
	posA := strings.Index(got, "A-Header")
	posM := strings.Index(got, "M-Header")
	posZ := strings.Index(got, "Z-Header")

	if posA < 0 || posM < 0 || posZ < 0 {
		t.Fatal("one or more headers not found in output")
	}
	if !(posA < posM && posM < posZ) {
		t.Errorf("headers not in alphabetical order: A=%d M=%d Z=%d\noutput:\n%s", posA, posM, posZ, got)
	}
}

func TestEscapeSingleQuote(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"no quotes", "no quotes"},
		{"it's", `it'\''s`},
		{"a'b'c", `a'\''b'\''c`},
		{"'''", `'\'''\'''\''`},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := escapeSingleQuote(tt.input)
			if got != tt.want {
				t.Errorf("escapeSingleQuote(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
