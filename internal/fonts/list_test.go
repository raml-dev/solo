// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package fonts

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
)

func resetCache() {
	cacheMu.Lock()
	cachedFamilies = nil
	cacheMu.Unlock()
}

func fixturePath(parts ...string) string {
	segments := append([]string{"..", "..", "test", "fonts"}, parts...)
	return filepath.Join(segments...)
}

func copyFile(t *testing.T, srcPath, dstPath string) {
	t.Helper()

	src, err := os.Open(srcPath)
	if err != nil {
		t.Fatalf("Open(%q) failed: %v", srcPath, err)
	}
	defer src.Close()

	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	dst, err := os.Create(dstPath)
	if err != nil {
		t.Fatalf("Create(%q) failed: %v", dstPath, err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		t.Fatalf("Copy failed: %v", err)
	}
}

func TestParseFontFile(t *testing.T) {
	t.Run("ttf sans fixture", func(t *testing.T) {
		path := fixturePath("inter", "Inter-Regular.ttf")

		fonts, err := parseFontFile(path)
		if err != nil {
			t.Fatalf("parseFontFile(%q) failed: %v", path, err)
		}

		want := []SystemFont{{Family: "Inter", IsMonospace: false}}
		if !reflect.DeepEqual(fonts, want) {
			t.Fatalf("parseFontFile(%q) = %#v, want %#v", path, fonts, want)
		}
	})

	t.Run("otf sans fixture", func(t *testing.T) {
		path := fixturePath("inter", "Inter-Regular.otf")

		fonts, err := parseFontFile(path)
		if err != nil {
			t.Fatalf("parseFontFile(%q) failed: %v", path, err)
		}

		want := []SystemFont{{Family: "Inter", IsMonospace: false}}
		if !reflect.DeepEqual(fonts, want) {
			t.Fatalf("parseFontFile(%q) = %#v, want %#v", path, fonts, want)
		}
	})

	t.Run("ttf monospace fixture", func(t *testing.T) {
		path := fixturePath("jetbrainsmono", "JetBrainsMono-Regular.ttf")

		fonts, err := parseFontFile(path)
		if err != nil {
			t.Fatalf("parseFontFile(%q) failed: %v", path, err)
		}

		want := []SystemFont{{Family: "JetBrains Mono", IsMonospace: true}}
		if !reflect.DeepEqual(fonts, want) {
			t.Fatalf("parseFontFile(%q) = %#v, want %#v", path, fonts, want)
		}
	})

	t.Run("ttc fixture", func(t *testing.T) {
		path := fixturePath("inter", "Inter.ttc")

		fonts, err := parseFontFile(path)
		if err != nil {
			t.Fatalf("parseFontFile(%q) failed: %v", path, err)
		}
		if len(fonts) != 36 {
			t.Fatalf("parseFontFile(%q) returned %d fonts, want 36", path, len(fonts))
		}

		families := make([]string, 0, len(fonts))
		for _, font := range fonts {
			families = append(families, font.Family)
			if font.IsMonospace {
				t.Fatalf("parseFontFile(%q) returned monospace TTC face: %#v", path, font)
			}
		}

		if !slices.Contains(families, "Inter") {
			t.Fatalf("parseFontFile(%q) families = %#v, want family %q", path, families, "Inter")
		}
		if !slices.Contains(families, "Inter Display") {
			t.Fatalf("parseFontFile(%q) families = %#v, want family %q", path, families, "Inter Display")
		}
	})

	t.Run("invalid font fixture", func(t *testing.T) {
		path := fixturePath("InvalidFont.otf")

		_, err := parseFontFile(path)
		if err == nil {
			t.Fatalf("parseFontFile(%q) succeeded, want error", path)
		}
	})
}

func TestListFamiliesSortedAndDeduplicated(t *testing.T) {
	resetCache()
	t.Cleanup(resetCache)

	tempDir := t.TempDir()

	copyFile(t, fixturePath("inter", "Inter-Regular.ttf"), filepath.Join(tempDir, "A", "Inter-Regular.ttf"))
	copyFile(t, fixturePath("inter", "Inter-Bold.ttf"), filepath.Join(tempDir, "B", "Inter-Bold.ttf"))
	copyFile(t, fixturePath("inter", "Inter-Regular.otf"), filepath.Join(tempDir, "C", "Inter-Regular.otf"))
	copyFile(t, fixturePath("inter", "Inter-Bold.otf"), filepath.Join(tempDir, "D", "Inter-Bold.otf"))
	copyFile(t, fixturePath("jetbrainsmono", "JetBrainsMono-Regular.ttf"), filepath.Join(tempDir, "E", "JetBrainsMono-Regular.ttf"))
	copyFile(t, fixturePath("InvalidFont.otf"), filepath.Join(tempDir, "broken", "InvalidFont.otf"))
	copyFile(t, fixturePath("NotAFont.txt"), filepath.Join(tempDir, "ignored", "NotAFont.txt"))

	originalGetFontDirs := getFontDirs
	getFontDirs = func() []string { return []string{tempDir} }
	t.Cleanup(func() { getFontDirs = originalGetFontDirs })

	got, err := ListFamilies(true)
	if err != nil {
		t.Fatalf("ListFamilies(true) failed: %v", err)
	}

	want := []SystemFont{
		{Family: "Inter", IsMonospace: false},
		{Family: "JetBrains Mono", IsMonospace: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ListFamilies(true) = %#v, want %#v", got, want)
	}
}

func TestListFamiliesUsesCacheUntilRefresh(t *testing.T) {
	resetCache()
	t.Cleanup(resetCache)

	tempDir := t.TempDir()
	copyFile(t, fixturePath("inter", "Inter-Regular.ttf"), filepath.Join(tempDir, "Inter-Regular.ttf"))

	originalGetFontDirs := getFontDirs
	getFontDirs = func() []string { return []string{tempDir} }
	t.Cleanup(func() { getFontDirs = originalGetFontDirs })

	got, err := ListFamilies(false)
	if err != nil {
		t.Fatalf("ListFamilies(false) failed: %v", err)
	}

	wantInitial := []SystemFont{{Family: "Inter", IsMonospace: false}}
	if !reflect.DeepEqual(got, wantInitial) {
		t.Fatalf("initial ListFamilies(false) = %#v, want %#v", got, wantInitial)
	}

	copyFile(t, fixturePath("jetbrainsmono", "JetBrainsMono-Regular.ttf"), filepath.Join(tempDir, "nested", "JetBrainsMono-Regular.ttf"))

	cached, err := ListFamilies(false)
	if err != nil {
		t.Fatalf("cached ListFamilies(false) failed: %v", err)
	}
	if !reflect.DeepEqual(cached, got) {
		t.Fatalf("cached ListFamilies(false) = %#v, want %#v", cached, got)
	}

	fresh, err := ListFamilies(true)
	if err != nil {
		t.Fatalf("ListFamilies(true) failed: %v", err)
	}

	wantFresh := []SystemFont{
		{Family: "Inter", IsMonospace: false},
		{Family: "JetBrains Mono", IsMonospace: true},
	}
	if !reflect.DeepEqual(fresh, wantFresh) {
		t.Fatalf("ListFamilies(true) = %#v, want %#v", fresh, wantFresh)
	}
}

func TestListFamiliesSkipsUnusableDirectories(t *testing.T) {
	resetCache()
	t.Cleanup(resetCache)

	tempDir := t.TempDir()
	missingDir := filepath.Join(tempDir, "missing")
	validDir := filepath.Join(tempDir, "valid")

	copyFile(t, fixturePath("inter", "Inter-Regular.ttf"), filepath.Join(validDir, "Inter-Regular.ttf"))

	originalGetFontDirs := getFontDirs
	getFontDirs = func() []string { return []string{missingDir, validDir} }
	t.Cleanup(func() { getFontDirs = originalGetFontDirs })

	got, err := ListFamilies(true)
	if err != nil {
		t.Fatalf("ListFamilies(true) failed: %v", err)
	}

	want := []SystemFont{{Family: "Inter", IsMonospace: false}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ListFamilies(true) = %#v, want %#v", got, want)
	}
}

func TestParseFontFileMissingFile(t *testing.T) {
	_, err := parseFontFile(filepath.Join(t.TempDir(), "missing.ttf"))
	if err == nil {
		t.Fatal("parseFontFile(missing) succeeded, want error")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("parseFontFile(missing) error = %v, want os.ErrNotExist", err)
	}
}
