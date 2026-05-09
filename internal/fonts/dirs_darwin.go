// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

//go:build darwin

package fonts

import (
	"os"
	"path/filepath"
)

func fontDirs() []string {
	home, _ := os.UserHomeDir()
	return []string{
		filepath.Join(home, "Library", "Fonts"),
		"/Library/Fonts",
		"/System/Library/Fonts",
	}
}
