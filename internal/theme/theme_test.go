package theme

import (
	"testing"
)

func TestGetPredefinedThemes(t *testing.T) {
	predefinedThemes := GetPredefinedThemes()

	if len(predefinedThemes) == 0 {
		t.Fatal("GetPredefinedThemes returned an empty slice, expected at least one theme.")
	}

	foundZincLight := false
	foundZincDark := false

	for _, theme := range predefinedThemes {
		if theme.Name == "" {
			t.Error("Found a theme with an empty name.")
		}
		if len(theme.Colors) == 0 {
			t.Errorf("Theme '%s' has no colors defined.", theme.Name)
		}
		if theme.Name == "zinc-light" {
			foundZincLight = true
		}
		if theme.Name == "zinc-dark" {
			foundZincDark = true
		}
	}

	if !foundZincLight {
		t.Error("Predefined theme 'zinc-light' was not found.")
	}
	if !foundZincDark {
		t.Error("Predefined theme 'zinc-dark' was not found.")
	}

	// Ensure all themes have required color keys
	requiredKeys := []string{"primary", "bg-primary", "bg-secondary", "bg-tertiary", "border", "text", "text-muted"}
	for _, theme := range predefinedThemes {
		for _, key := range requiredKeys {
			if _, ok := theme.Colors[key]; !ok {
				t.Errorf("Theme '%s' is missing required color key '%s'.", theme.Name, key)
			}
		}
	}
}
