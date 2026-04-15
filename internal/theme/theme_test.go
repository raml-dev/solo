// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package theme

import "testing"

func TestGetPredefinedThemes(t *testing.T) {
	predefinedThemes := GetPredefinedThemes()

	if len(predefinedThemes) == 0 {
		t.Fatal("GetPredefinedThemes returned an empty slice, expected at least one theme.")
	}

	foundOcean := false
	foundNord := false

	for _, th := range predefinedThemes {
		if th.ID == "" {
			t.Error("Found a theme with an empty id.")
		}
		if th.Label == "" {
			t.Errorf("Theme '%s' has empty label.", th.ID)
		}
		if th.Config.Type == "" {
			t.Errorf("Theme '%s' has empty config type.", th.ID)
		}
		if th.Config.Seeds.Primary == "" || th.Config.Seeds.Neutral == "" {
			t.Errorf("Theme '%s' is missing required seed colors.", th.ID)
		}
		if th.ID == "ocean" {
			foundOcean = true
		}
		if th.ID == "nord" {
			foundNord = true
		}
	}

	if !foundOcean {
		t.Error("Predefined theme 'ocean' was not found.")
	}
	if !foundNord {
		t.Error("Predefined theme 'nord' was not found.")
	}
}
