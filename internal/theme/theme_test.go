package theme

import (
	"testing"
)

func TestGetPredefinedThemes(t *testing.T) {
	predefinedThemes := GetPredefinedThemes()

	if len(predefinedThemes) == 0 {
		t.Fatal("GetPredefinedThemes returned an empty slice, expected at least one theme.")
	}

	foundDefaultLight := false
	foundDefaultDark := false

	for _, theme := range predefinedThemes {
		if theme.Name == "" {
			t.Error("Found a theme with an empty name.")
		}
		if len(theme.Colors) == 0 {
			t.Errorf("Theme '%s' has no colors defined.", theme.Name)
		}

		if theme.Name == "default-light" {
			foundDefaultLight = true
		}
		if theme.Name == "default-dark" {
			foundDefaultDark = true
		}
	}

	if !foundDefaultLight {
		t.Error("Predefined theme 'default-light' was not found.")
	}
	if !foundDefaultDark {
		t.Error("Predefined theme 'default-dark' was not found.")
	}
}
