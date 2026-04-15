// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package exporter

import (
	"sort"
	"strings"
)

// CurlExportRequest holds the already-resolved request data to be rendered as a cURL command.
type CurlExportRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

// GenerateCurl converts a resolved HTTP request into a multi-line cURL command string.
// Header keys are sorted alphabetically for deterministic output.
// Single quotes inside URL, header values, and body are escaped with the '\” technique.
func GenerateCurl(req CurlExportRequest) string {
	var lines []string

	lines = append(lines, "curl \\")
	lines = append(lines, "  -X "+req.Method+" \\")
	lines = append(lines, "  '"+escapeSingleQuote(req.URL)+"'")

	// Sort header keys for deterministic output
	keys := make([]string, 0, len(req.Headers))
	for k := range req.Headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := req.Headers[k]
		lines = append(lines, "  -H '"+escapeSingleQuote(k+": "+v)+"'")
	}

	if req.Body != "" {
		lines = append(lines, "  -d '"+escapeSingleQuote(req.Body)+"'")
	}

	// Append " \" to all lines except the last
	result := make([]string, len(lines))
	for i, line := range lines {
		if i < len(lines)-1 {
			// Lines that already end with " \" (method line) stay as-is;
			// other lines need the continuation appended.
			if !strings.HasSuffix(line, "\\") {
				result[i] = line + " \\"
			} else {
				result[i] = line
			}
		} else {
			// Last line: strip any trailing " \" that might have been added
			result[i] = strings.TrimSuffix(line, " \\")
		}
	}

	return strings.Join(result, "\n")
}

// escapeSingleQuote replaces ' with '\” for use inside single-quoted shell strings.
func escapeSingleQuote(s string) string {
	return strings.ReplaceAll(s, "'", `'\''`)
}
