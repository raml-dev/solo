// theme.go
// Place in project root
package theme

// Theme represents a color theme configuration
type Theme struct {
	Name   string            `json:"name"`
	Colors map[string]string `json:"colors"`
}

// GetPredefinedThemes returns the built-in themes
func GetPredefinedThemes() []Theme {
	return []Theme{
		{
			Name: "default-light",
			Colors: map[string]string{
				"primary":      "#2563eb",
				"primary-dark": "#1e40af",
				"success":      "#10b981",
				"warning":      "#f59e0b",
				"danger":       "#ef4444",
				"info":         "#06b6d4",
				"bg-primary":   "#ffffff",
				"bg-secondary": "#f9fafb",
				"bg-tertiary":  "#f3f4f6",
				"border":       "#e5e7eb",
				"border-dark":  "#d1d5db",
				"text":         "#111827",
				"text-muted":   "#6b7280",
				"text-light":   "#9ca3af",
			},
		},
		{
			Name: "default-dark",
			Colors: map[string]string{
				"primary":      "#3b82f6",
				"primary-dark": "#2563eb",
				"success":      "#10b981",
				"warning":      "#f59e0b",
				"danger":       "#ef4444",
				"info":         "#06b6d4",
				"bg-primary":   "#111827",
				"bg-secondary": "#1f2937",
				"bg-tertiary":  "#374151",
				"border":       "#374151",
				"border-dark":  "#4b5563",
				"text":         "#f9fafb",
				"text-muted":   "#d1d5db",
				"text-light":   "#9ca3af",
			},
		},
		{
			Name: "dracula",
			Colors: map[string]string{
				"primary":      "#bd93f9",
				"primary-dark": "#9580d6",
				"success":      "#50fa7b",
				"warning":      "#f1fa8c",
				"danger":       "#ff5555",
				"info":         "#8be9fd",
				"bg-primary":   "#282a36",
				"bg-secondary": "#1e1f29",
				"bg-tertiary":  "#383a59",
				"border":       "#44475a",
				"border-dark":  "#6272a4",
				"text":         "#f8f8f2",
				"text-muted":   "#e6e6e6",
				"text-light":   "#bfbfbf",
			},
		},
		{
			Name: "nord",
			Colors: map[string]string{
				"primary":      "#88c0d0",
				"primary-dark": "#5e81ac",
				"success":      "#a3be8c",
				"warning":      "#ebcb8b",
				"danger":       "#bf616a",
				"info":         "#81a1c1",
				"bg-primary":   "#2e3440",
				"bg-secondary": "#3b4252",
				"bg-tertiary":  "#434c5e",
				"border":       "#4c566a",
				"border-dark":  "#5e6b82",
				"text":         "#eceff4",
				"text-muted":   "#d8dee9",
				"text-light":   "#9099ab",
			},
		},
		{
			Name: "monokai",
			Colors: map[string]string{
				"primary":      "#66d9ef",
				"primary-dark": "#4db8d8",
				"success":      "#a6e22e",
				"warning":      "#e6db74",
				"danger":       "#f92672",
				"info":         "#ae81ff",
				"bg-primary":   "#272822",
				"bg-secondary": "#1e1f1b",
				"bg-tertiary":  "#3e3d32",
				"border":       "#49483e",
				"border-dark":  "#75715e",
				"text":         "#f8f8f2",
				"text-muted":   "#e6e6e6",
				"text-light":   "#75715e",
			},
		},
		{
			Name: "solarized-light",
			Colors: map[string]string{
				"primary":      "#268bd2",
				"primary-dark": "#1c6aa3",
				"success":      "#859900",
				"warning":      "#b58900",
				"danger":       "#dc322f",
				"info":         "#2aa198",
				"bg-primary":   "#fdf6e3",
				"bg-secondary": "#eee8d5",
				"bg-tertiary":  "#e4ddc8",
				"border":       "#d3cbb7",
				"border-dark":  "#c9c0a6",
				"text":         "#073642",
				"text-muted":   "#586e75",
				"text-light":   "#839496",
			},
		},
	}
}
