// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package fonts

import "solo/internal/tools"

func IsValidFontFamily(family string) bool {
	families, err := ListFamilies(false)
	if err != nil {
		return false
	}
	for _, f := range families {
		if f.Family == family {
			return true
		}
	}
	return false
}

func IsValidMonoFontFamily(family string) bool {
	families, err := ListFamilies(false)
	if err != nil {
		return false
	}
	for _, f := range families {
		if f.Family == family {
      // found the requested family, return true only if it is monospace
			return f.IsMonospace
		}
	}
	return false
}

func ClampBaseFontSizePx(v int) int {
	if v < tools.MIN_BASE_FONT_SIZE_PX || v > tools.MAX_BASE_FONT_SIZE_PX {
		return tools.DEFAULT_BASE_FONT_SIZE_PX
	}
	return v
}
