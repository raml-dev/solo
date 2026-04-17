// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package theme

// ThemeMode controls how a theme should be interpreted by the UI.
type ThemeMode string

const (
	ThemeModeLight  ThemeMode = "light"
	ThemeModeDark   ThemeMode = "dark"
	ThemeModeSystem ThemeMode = "system"
)

// ThemeSeeds are the semantic seed colors from which full ramps are derived.
type ThemeSeeds struct {
	Primary string `json:"primary"`
	Success string `json:"success"`
	Warning string `json:"warning"`
	Danger  string `json:"danger"`
	Neutral string `json:"neutral"`
	Surface string `json:"surface,omitempty"`
}

// ThemeConfig is the persisted config for either a preset or a custom theme.
type ThemeConfig struct {
	Type     string     `json:"type"` // "preset" | "custom"
	PresetID string     `json:"presetId,omitempty"`
	Mode     ThemeMode  `json:"mode"`
	Seeds    ThemeSeeds `json:"seeds,omitempty"`
}

// Theme is the new canonical theme model.
type Theme struct {
	ID     string      `json:"id"`
	Label  string      `json:"label"`
	Config ThemeConfig `json:"config"`
}

func preset(id, label string, seeds ThemeSeeds) Theme {
	return Theme{
		ID:    id,
		Label: label,
		Config: ThemeConfig{
			Type:     "preset",
			PresetID: id,
			Mode:     ThemeModeSystem,
			Seeds:    seeds,
		},
	}
}

// GetPredefinedThemes returns all built-in preset themes.
func GetPredefinedThemes() []Theme {
	return []Theme{
		preset("ocean", "Ocean", ThemeSeeds{
			Primary: "#0ea5e9",
			Success: "#10b981",
			Warning: "#f59e0b",
			Danger:  "#ef4444",
			Neutral: "#52525b",
		}),
		preset("ember", "Ember", ThemeSeeds{
			Primary: "#f97316",
			Success: "#22c55e",
			Warning: "#f59e0b",
			Danger:  "#ef4444",
			Neutral: "#52525b",
		}),
		preset("forest", "Forest", ThemeSeeds{
			Primary: "#22c55e",
			Success: "#16a34a",
			Warning: "#eab308",
			Danger:  "#dc2626",
			Neutral: "#52525b",
		}),
		preset("violet", "Violet", ThemeSeeds{
			Primary: "#8b5cf6",
			Success: "#10b981",
			Warning: "#f59e0b",
			Danger:  "#ef4444",
			Neutral: "#52525b",
		}),
		preset("nord", "Nord", ThemeSeeds{
			Primary: "#5e81ac",
			Success: "#a3be8c",
			Warning: "#ebcb8b",
			Danger:  "#bf616a",
			Neutral: "#4c566a",
			Surface: "#2e3440",
		}),
		preset("pastel", "Pastel", ThemeSeeds{
			Primary: "#7dd3fc",
			Success: "#86efac",
			Warning: "#fde68a",
			Danger:  "#fda4af",
			Neutral: "#94a3b8",
			Surface: "#e2e8f0",
		}),
	}
}
