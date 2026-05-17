// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package fonts

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
)

var getFontDirs = fontDirs

var (
	cacheMu        sync.RWMutex
	cachedFamilies []SystemFont
)

func ListFamilies(refresh bool) ([]SystemFont, error) {
	if !refresh {
		cacheMu.RLock()
		if cachedFamilies != nil {
			cached := slices.Clone(cachedFamilies)
			cacheMu.RUnlock()
			return cached, nil
		}
		cacheMu.RUnlock()
	}

	families, err := scanFamilies()
	if err != nil {
		return nil, err
	}

	cacheMu.Lock()
	cachedFamilies = slices.Clone(families)
	cacheMu.Unlock()

	return families, nil
}

func scanFamilies() ([]SystemFont, error) {
	seen := map[string]SystemFont{}

	for _, dir := range getFontDirs() {
		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			continue
		}

		_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if !isSupportedFontExtension(strings.ToLower(filepath.Ext(d.Name()))) {
				return nil
			}

			fonts, err := parseFontFile(path)
			if err != nil {
				return nil
			}

			for _, font := range fonts {
				if font.Family == "" {
					continue
				}
				if _, exists := seen[font.Family]; exists {
					continue
				}
				seen[font.Family] = font
			}
			return nil
		})
	}

	families := make([]SystemFont, 0, len(seen))
	for _, font := range seen {
		families = append(families, font)
	}

	sort.Slice(families, func(i, j int) bool {
		return families[i].Family < families[j].Family
	})
	return families, nil
}

func isSupportedFontExtension(ext string) bool {
	switch ext {
	case ".ttf", ".otf", ".ttc":
		return true
	default:
		return false
	}
}
