// theme.go
package theme

// Theme represents a color theme configuration
type Theme struct {
	Name   string            `json:"name"`
	Colors map[string]string `json:"colors"`
}

// sharedAccents returns the semantic accent colors shared across all themes.
// Each theme overrides bg/text/border; accents stay consistent.
func sharedLight() map[string]string {
	return map[string]string{
		"success":            "#16a34a",
		"warning":            "#d97706",
		"danger":             "#dc2626",
		"info":               "#0891b2",
		"status-success-bg":  "#dcfce7",
		"status-success-text": "#14532d",
		"status-warning-bg":  "#fef9c3",
		"status-warning-text": "#713f12",
		"status-danger-bg":   "#fee2e2",
		"status-danger-text": "#7f1d1d",
		"status-info-bg":     "#e0f2fe",
		"status-info-text":   "#0c4a6e",
		"method-get-bg":      "#dbeafe",
		"method-get-text":    "#1e40af",
		"method-post-bg":     "#dcfce7",
		"method-post-text":   "#14532d",
		"method-put-bg":      "#fef9c3",
		"method-put-text":    "#713f12",
		"method-delete-bg":   "#fee2e2",
		"method-delete-text": "#7f1d1d",
		"method-patch-bg":    "#ede9fe",
		"method-patch-text":  "#3b0764",
	}
}

func sharedDark() map[string]string {
	return map[string]string{
		"success":            "#22c55e",
		"warning":            "#f59e0b",
		"danger":             "#ef4444",
		"info":               "#22d3ee",
		"status-success-bg":  "#14532d",
		"status-success-text": "#bbf7d0",
		"status-warning-bg":  "#713f12",
		"status-warning-text": "#fef08a",
		"status-danger-bg":   "#7f1d1d",
		"status-danger-text": "#fecaca",
		"status-info-bg":     "#0c4a6e",
		"status-info-text":   "#bae6fd",
		"method-get-bg":      "#1e3a5f",
		"method-get-text":    "#93c5fd",
		"method-post-bg":     "#14532d",
		"method-post-text":   "#86efac",
		"method-put-bg":      "#713f12",
		"method-put-text":    "#fde68a",
		"method-delete-bg":   "#7f1d1d",
		"method-delete-text": "#fca5a5",
		"method-patch-bg":    "#2e1065",
		"method-patch-text":  "#c4b5fd",
	}
}

func merge(base, override map[string]string) map[string]string {
	result := make(map[string]string, len(base)+len(override))
	for k, v := range base {
		result[k] = v
	}
	for k, v := range override {
		result[k] = v
	}
	return result
}

// GetPredefinedThemes returns all built-in themes (shadcn-svelte palettes, light + dark)
func GetPredefinedThemes() []Theme {
	return []Theme{
		// ── Zinc ────────────────────────────────────────────────────────────
		{
			Name: "zinc-light",
			Colors: merge(sharedLight(), map[string]string{
				"primary":      "#18181b",
				"primary-dark": "#09090b",
				"bg-primary":   "#ffffff",
				"bg-secondary": "#fafafa",
				"bg-tertiary":  "#f4f4f5",
				"border":       "#e4e4e7",
				"border-dark":  "#d4d4d8",
				"text":         "#09090b",
				"text-muted":   "#71717a",
				"text-light":   "#a1a1aa",
			}),
		},
		{
			Name: "zinc-dark",
			Colors: merge(sharedDark(), map[string]string{
				"primary":      "#fafafa",
				"primary-dark": "#e4e4e7",
				"bg-primary":   "#09090b",
				"bg-secondary": "#18181b",
				"bg-tertiary":  "#27272a",
				"border":       "#3f3f46",
				"border-dark":  "#52525b",
				"text":         "#fafafa",
				"text-muted":   "#a1a1aa",
				"text-light":   "#71717a",
			}),
		},

		// ── Slate ────────────────────────────────────────────────────────────
		{
			Name: "slate-light",
			Colors: merge(sharedLight(), map[string]string{
				"primary":      "#0f172a",
				"primary-dark": "#020617",
				"bg-primary":   "#ffffff",
				"bg-secondary": "#f8fafc",
				"bg-tertiary":  "#f1f5f9",
				"border":       "#e2e8f0",
				"border-dark":  "#cbd5e1",
				"text":         "#0f172a",
				"text-muted":   "#64748b",
				"text-light":   "#94a3b8",
			}),
		},
		{
			Name: "slate-dark",
			Colors: merge(sharedDark(), map[string]string{
				"primary":      "#f8fafc",
				"primary-dark": "#e2e8f0",
				"bg-primary":   "#020617",
				"bg-secondary": "#0f172a",
				"bg-tertiary":  "#1e293b",
				"border":       "#1e293b",
				"border-dark":  "#334155",
				"text":         "#f8fafc",
				"text-muted":   "#94a3b8",
				"text-light":   "#64748b",
			}),
		},

		// ── Stone ────────────────────────────────────────────────────────────
		{
			Name: "stone-light",
			Colors: merge(sharedLight(), map[string]string{
				"primary":      "#1c1917",
				"primary-dark": "#0c0a09",
				"bg-primary":   "#ffffff",
				"bg-secondary": "#fafaf9",
				"bg-tertiary":  "#f5f5f4",
				"border":       "#e7e5e4",
				"border-dark":  "#d6d3d1",
				"text":         "#1c1917",
				"text-muted":   "#78716c",
				"text-light":   "#a8a29e",
			}),
		},
		{
			Name: "stone-dark",
			Colors: merge(sharedDark(), map[string]string{
				"primary":      "#fafaf9",
				"primary-dark": "#e7e5e4",
				"bg-primary":   "#0c0a09",
				"bg-secondary": "#1c1917",
				"bg-tertiary":  "#292524",
				"border":       "#44403c",
				"border-dark":  "#57534e",
				"text":         "#fafaf9",
				"text-muted":   "#a8a29e",
				"text-light":   "#78716c",
			}),
		},

		// ── Gray ─────────────────────────────────────────────────────────────
		{
			Name: "gray-light",
			Colors: merge(sharedLight(), map[string]string{
				"primary":      "#111827",
				"primary-dark": "#030712",
				"bg-primary":   "#ffffff",
				"bg-secondary": "#f9fafb",
				"bg-tertiary":  "#f3f4f6",
				"border":       "#e5e7eb",
				"border-dark":  "#d1d5db",
				"text":         "#111827",
				"text-muted":   "#6b7280",
				"text-light":   "#9ca3af",
			}),
		},
		{
			Name: "gray-dark",
			Colors: merge(sharedDark(), map[string]string{
				"primary":      "#f9fafb",
				"primary-dark": "#e5e7eb",
				"bg-primary":   "#030712",
				"bg-secondary": "#111827",
				"bg-tertiary":  "#1f2937",
				"border":       "#1f2937",
				"border-dark":  "#374151",
				"text":         "#f9fafb",
				"text-muted":   "#9ca3af",
				"text-light":   "#6b7280",
			}),
		},

		// ── Neutral ───────────────────────────────────────────────────────────
		{
			Name: "neutral-light",
			Colors: merge(sharedLight(), map[string]string{
				"primary":      "#171717",
				"primary-dark": "#0a0a0a",
				"bg-primary":   "#ffffff",
				"bg-secondary": "#fafafa",
				"bg-tertiary":  "#f5f5f5",
				"border":       "#e5e5e5",
				"border-dark":  "#d4d4d4",
				"text":         "#171717",
				"text-muted":   "#737373",
				"text-light":   "#a3a3a3",
			}),
		},
		{
			Name: "neutral-dark",
			Colors: merge(sharedDark(), map[string]string{
				"primary":      "#fafafa",
				"primary-dark": "#e5e5e5",
				"bg-primary":   "#0a0a0a",
				"bg-secondary": "#171717",
				"bg-tertiary":  "#262626",
				"border":       "#404040",
				"border-dark":  "#525252",
				"text":         "#fafafa",
				"text-muted":   "#a3a3a3",
				"text-light":   "#737373",
			}),
		},

		// ── Rose ──────────────────────────────────────────────────────────────
		{
			Name: "rose-light",
			Colors: merge(sharedLight(), map[string]string{
				"primary":      "#e11d48",
				"primary-dark": "#be123c",
				"bg-primary":   "#ffffff",
				"bg-secondary": "#fff1f2",
				"bg-tertiary":  "#ffe4e6",
				"border":       "#fecdd3",
				"border-dark":  "#fda4af",
				"text":         "#0f172a",
				"text-muted":   "#64748b",
				"text-light":   "#94a3b8",
			}),
		},
		{
			Name: "rose-dark",
			Colors: merge(sharedDark(), map[string]string{
				"primary":      "#fb7185",
				"primary-dark": "#f43f5e",
				"bg-primary":   "#0f0a0b",
				"bg-secondary": "#1a0f11",
				"bg-tertiary":  "#2d1418",
				"border":       "#4c1d25",
				"border-dark":  "#6b2737",
				"text":         "#fff1f2",
				"text-muted":   "#fda4af",
				"text-light":   "#fb7185",
			}),
		},

		// ── Orange ────────────────────────────────────────────────────────────
		{
			Name: "orange-light",
			Colors: merge(sharedLight(), map[string]string{
				"primary":      "#ea580c",
				"primary-dark": "#c2410c",
				"bg-primary":   "#ffffff",
				"bg-secondary": "#fff7ed",
				"bg-tertiary":  "#ffedd5",
				"border":       "#fed7aa",
				"border-dark":  "#fdba74",
				"text":         "#0f172a",
				"text-muted":   "#64748b",
				"text-light":   "#94a3b8",
			}),
		},
		{
			Name: "orange-dark",
			Colors: merge(sharedDark(), map[string]string{
				"primary":      "#fb923c",
				"primary-dark": "#f97316",
				"bg-primary":   "#0d0900",
				"bg-secondary": "#1a1000",
				"bg-tertiary":  "#2d1f00",
				"border":       "#431f00",
				"border-dark":  "#7c2d12",
				"text":         "#fff7ed",
				"text-muted":   "#fdba74",
				"text-light":   "#fb923c",
			}),
		},

		// ── Green ─────────────────────────────────────────────────────────────
		{
			Name: "green-light",
			Colors: merge(sharedLight(), map[string]string{
				"primary":      "#16a34a",
				"primary-dark": "#15803d",
				"bg-primary":   "#ffffff",
				"bg-secondary": "#f0fdf4",
				"bg-tertiary":  "#dcfce7",
				"border":       "#bbf7d0",
				"border-dark":  "#86efac",
				"text":         "#0f172a",
				"text-muted":   "#64748b",
				"text-light":   "#94a3b8",
			}),
		},
		{
			Name: "green-dark",
			Colors: merge(sharedDark(), map[string]string{
				"primary":      "#4ade80",
				"primary-dark": "#22c55e",
				"bg-primary":   "#030a05",
				"bg-secondary": "#071210",
				"bg-tertiary":  "#0d2218",
				"border":       "#14532d",
				"border-dark":  "#166534",
				"text":         "#f0fdf4",
				"text-muted":   "#86efac",
				"text-light":   "#4ade80",
			}),
		},

		// ── Blue ──────────────────────────────────────────────────────────────
		{
			Name: "blue-light",
			Colors: merge(sharedLight(), map[string]string{
				"primary":      "#2563eb",
				"primary-dark": "#1d4ed8",
				"bg-primary":   "#ffffff",
				"bg-secondary": "#eff6ff",
				"bg-tertiary":  "#dbeafe",
				"border":       "#bfdbfe",
				"border-dark":  "#93c5fd",
				"text":         "#0f172a",
				"text-muted":   "#64748b",
				"text-light":   "#94a3b8",
			}),
		},
		{
			Name: "blue-dark",
			Colors: merge(sharedDark(), map[string]string{
				"primary":      "#60a5fa",
				"primary-dark": "#3b82f6",
				"bg-primary":   "#030711",
				"bg-secondary": "#0c1325",
				"bg-tertiary":  "#172044",
				"border":       "#1e3a8a",
				"border-dark":  "#1d4ed8",
				"text":         "#eff6ff",
				"text-muted":   "#93c5fd",
				"text-light":   "#60a5fa",
			}),
		},

		// ── Yellow ────────────────────────────────────────────────────────────
		{
			Name: "yellow-light",
			Colors: merge(sharedLight(), map[string]string{
				"primary":      "#ca8a04",
				"primary-dark": "#a16207",
				"bg-primary":   "#ffffff",
				"bg-secondary": "#fefce8",
				"bg-tertiary":  "#fef9c3",
				"border":       "#fef08a",
				"border-dark":  "#fde047",
				"text":         "#0f172a",
				"text-muted":   "#64748b",
				"text-light":   "#94a3b8",
			}),
		},
		{
			Name: "yellow-dark",
			Colors: merge(sharedDark(), map[string]string{
				"primary":      "#facc15",
				"primary-dark": "#eab308",
				"bg-primary":   "#0d0b00",
				"bg-secondary": "#1a1600",
				"bg-tertiary":  "#2d2700",
				"border":       "#422006",
				"border-dark":  "#713f12",
				"text":         "#fefce8",
				"text-muted":   "#fde047",
				"text-light":   "#facc15",
			}),
		},

		// ── Violet ────────────────────────────────────────────────────────────
		{
			Name: "violet-light",
			Colors: merge(sharedLight(), map[string]string{
				"primary":      "#7c3aed",
				"primary-dark": "#6d28d9",
				"bg-primary":   "#ffffff",
				"bg-secondary": "#f5f3ff",
				"bg-tertiary":  "#ede9fe",
				"border":       "#ddd6fe",
				"border-dark":  "#c4b5fd",
				"text":         "#0f172a",
				"text-muted":   "#64748b",
				"text-light":   "#94a3b8",
			}),
		},
		{
			Name: "violet-dark",
			Colors: merge(sharedDark(), map[string]string{
				"primary":      "#a78bfa",
				"primary-dark": "#7c3aed",
				"bg-primary":   "#060311",
				"bg-secondary": "#0e0820",
				"bg-tertiary":  "#1c1240",
				"border":       "#2e1065",
				"border-dark":  "#4c1d95",
				"text":         "#f5f3ff",
				"text-muted":   "#c4b5fd",
				"text-light":   "#a78bfa",
			}),
		},
	}
}
