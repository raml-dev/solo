// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package fonts

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"

	"golang.org/x/image/font/sfnt"
)

type SystemFont struct {
	Family      string `json:"family"`
	IsMonospace bool   `json:"isMonospace"`
}

var errNoUsableFontMetadata = errors.New("fonts: no usable font metadata")

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

func parseFontFile(path string) ([]SystemFont, error) {
     file, err := os.Open(path)
     if err != nil {
         return nil, err
     }
     defer file.Close()
     ext := strings.ToLower(filepath.Ext(path))
     if ext == ".ttc" {
         collection, err := sfnt.ParseCollectionReaderAt(file)
         if err != nil {
             return nil, err
         }
         fonts := make([]SystemFont, 0, collection.NumFonts())
         for i := 0; i < collection.NumFonts(); i++ {
             font, err := collection.Font(i)
             if err != nil {
                 continue
             }
             parsed, err := parseSFNTFont(font)
             if err != nil {
                 continue
             }
             fonts = append(fonts, parsed)
         }
         if len(fonts) == 0 {
             return nil, errNoUsableFontMetadata
         }
         return fonts, nil
     }
     font, err := sfnt.ParseReaderAt(file)
     if err != nil {
         return nil, err
     }
     parsed, err := parseSFNTFont(font)
     if err != nil {
         return nil, err
     }
     return []SystemFont{parsed}, nil
 }

func parseSFNTFont(font *sfnt.Font) (SystemFont, error) {
	family, err := fontFamilyName(font)
	if err != nil {
		return SystemFont{}, err
	}

	post := font.PostTable()
	return SystemFont{
		Family:      family,
		IsMonospace: post != nil && post.IsFixedPitch,
	}, nil
}

func fontFamilyName(font *sfnt.Font) (string, error) {
	for _, id := range []sfnt.NameID{sfnt.NameIDTypographicFamily, sfnt.NameIDFamily} {
		name, err := font.Name(nil, id)
		if err != nil {
			continue
		}
		name = strings.TrimSpace(name)
		if name != "" {
			return name, nil
		}
	}

	return "", errNoUsableFontMetadata
}

func isSupportedFontExtension(ext string) bool {
	switch ext {
	case ".ttf", ".otf", ".ttc":
		return true
	default:
		return false
	}
}
