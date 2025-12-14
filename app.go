package main

import (
	"context"
	"fmt"
	"io"
	"yapla/internal/requester"
	"yapla/internal/theme"
)

// App struct
type App struct {
	ctx     context.Context
	service *requester.Service
	themeManager *theme.ThemeManager
}

type RequestOptions struct {
	Method  string         `json:"method"`
	URL     string         `json:"url"`
	Headers map[string]any `json:"headers"`
	Body    string         `json:"body"`
}

type ResponseData struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Duration   int64             `json:"duration"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{service: requester.NewService(nil)}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize theme manager
	tm, err := theme.NewThemeManager()
	if err != nil {
		fmt.Printf("Error initializing theme manager: %v\n", err)
		// Continue anyway with nil theme manager
	}
	a.themeManager = tm
}

// Greet returns a greeting for the given name
func (a *App) Execute(options RequestOptions) (ResponseData, error) {
	result := ResponseData{}

	resp, err := a.service.Execute(options.Method, options.URL, options.Body, options.Headers, nil)

	if err != nil {
		return ResponseData{}, err
	}

	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode

	// 6. Lettura della Risposta
	bodyBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return result, err
	}

	result.Body = string(bodyBytes)

	respHeaders := make(map[string]string, len(resp.Header))

	for k, v := range resp.Header {
		respHeaders[k] = v[0]
	}

	result.Headers = respHeaders

	return result, nil
}

// Theme Management Methods

// GetActiveTheme returns the currently active theme name
func (a *App) GetActiveTheme() string {
	if a.themeManager == nil {
		return "default-light"
	}
	return a.themeManager.GetActiveTheme()
}

// SetActiveTheme sets the active theme
func (a *App) SetActiveTheme(themeName string) error {
	if a.themeManager == nil {
		return fmt.Errorf("theme manager not initialized")
	}
	return a.themeManager.SetActiveTheme(themeName)
}

// GetPredefinedThemes returns all built-in themes
func (a *App) GetPredefinedThemes() []theme.Theme {
	if a.themeManager == nil {
		return []theme.Theme{}
	}
	return a.themeManager.GetPredefinedThemes()
}

// GetCustomThemes returns all user-created themes
func (a *App) GetCustomThemes() []theme.Theme {
	if a.themeManager == nil {
		return []theme.Theme{}
	}
	return a.themeManager.GetCustomThemes()
}

// GetAllThemes returns both predefined and custom themes
func (a *App) GetAllThemes() []theme.Theme {
	if a.themeManager == nil {
		return []theme.Theme{}
	}
	return a.themeManager.GetAllThemes()
}

// SaveCustomTheme saves or updates a custom theme
func (a *App) SaveCustomTheme(theme theme.Theme) error {
	if a.themeManager == nil {
		return fmt.Errorf("theme manager not initialized")
	}
	return a.themeManager.SaveCustomTheme(theme)
}

// DeleteCustomTheme removes a custom theme
func (a *App) DeleteCustomTheme(themeName string) error {
	if a.themeManager == nil {
		return fmt.Errorf("theme manager not initialized")
	}
	return a.themeManager.DeleteCustomTheme(themeName)
}

// GetThemeByName returns a specific theme by name
func (a *App) GetThemeByName(themeName string) (*theme.Theme, error) {
	if a.themeManager == nil {
		return nil, fmt.Errorf("theme manager not initialized")
	}
	
	allThemes := a.themeManager.GetAllThemes()
	for _, theme := range allThemes {
		if theme.Name == themeName {
			return &theme, nil
		}
	}
	
	return nil, fmt.Errorf("theme not found: %s", themeName)
}
