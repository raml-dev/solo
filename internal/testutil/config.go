// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package testutil

import "testing"

// IsolateUserConfigDir redirects user-config resolution to a per-test temp dir.
func IsolateUserConfigDir(t *testing.T) string {
	t.Helper()

	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("LOCALAPPDATA", tmp)
	t.Setenv("AppData", tmp)

	return tmp
}
