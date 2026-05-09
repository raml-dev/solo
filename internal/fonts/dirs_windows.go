// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

//go:build windows

package fonts

import (
	"os"
	"path/filepath"
)

func fontDirs() []string {
	windir := os.Getenv("WINDIR")
	home, _ := os.UserHomeDir()
	return []string{
		filepath.Join(windir, "Fonts"),
		filepath.Join(home, "AppData", "Local", "Microsoft", "Windows", "Fonts"),
	}
}
