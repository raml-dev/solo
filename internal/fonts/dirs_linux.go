// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

//go:build linux

package fonts

import (
	"os"
	"path/filepath"
)

func fontDirs() []string {
	home, _ := os.UserHomeDir()
	return []string{
		filepath.Join(home, ".fonts"),
		filepath.Join(home, ".local", "share", "fonts"),
		"/usr/share/fonts",
		"/usr/local/share/fonts",
	}
}
