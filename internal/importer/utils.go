// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package importer

import (
	"fmt"
	"sort"
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
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(m))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", k, m[k]))
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
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		val := m[k]
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
