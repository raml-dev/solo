// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package importer

import (
	"fmt"
	"strings"
)

func objectToQueryString(v interface{}) string {
	if v == nil {
		return ""
	}
	m, ok := v.(map[string]interface{})
	if !ok {
		return fmt.Sprintf("%v", v)
	}
	parts := make([]string, 0, len(m))
	for k, val := range m {
		parts = append(parts, fmt.Sprintf("%s=%v", k, val))
	}
	return strings.Join(parts, "&")
}

func objectToMultipartString(v interface{}, boundary string) string {
	if v == nil {
		return ""
	}
	m, ok := v.(map[string]interface{})
	if !ok {
		return ""
	}
	var b strings.Builder
	for k, val := range m {
		b.WriteString("--" + boundary + "\r\n")
		b.WriteString(fmt.Sprintf("Content-Disposition: form-data; name=\"%s\"\r\n\r\n", k))
		if val == nil {
			b.WriteString("\r\n")
		} else {
			b.WriteString(fmt.Sprintf("%v\r\n", val))
		}
	}
	b.WriteString("--" + boundary + "--\r\n")
	return b.String()
}
