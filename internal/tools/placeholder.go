// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package tools

import (
	"maps"
	"regexp"
	"slices"
	"strings"
)

var placeholderRegex = regexp.MustCompile(PLACEHOLDER_REGEXP)

// Given a string, extract all the name of placeholders it contains, avoiding duplicates
func ExtractPlaceholders(text string) []string {
	found := make(map[string]struct{}, 0)
	matches := placeholderRegex.FindAllStringSubmatch(text, -1)
	for _, match := range matches {
		if len(match) > 1 {
			key := strings.TrimSpace(match[1])
			found[key] = struct{}{}
		}
	}

	return slices.Collect(maps.Keys(found))
}
