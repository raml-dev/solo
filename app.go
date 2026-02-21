package main

import (
	"context"
	"fmt"
	"log/slog"
	"yapla/internal/collection"
	"yapla/internal/configuration"
	"yapla/internal/environment"
	"yapla/internal/requester"
	"yapla/internal/theme"
)

// App struct
type App struct {
	ctx                context.Context
	service            *requester.Service
	collectionManager  *collection.CollectionManager
	environmentManager *environment.EnvironmentManager
	configManager      *configuration.ConfigurationManager
}

type RequestOptions struct {
	Method   string                                 `json:"method"`
	URL      string                                 `json:"url"`
	Headers  map[string]any                         `json:"headers"`
	Body     string                                 `json:"body"`
	Settings *configuration.RequestSettingsOverride `json:"settings,omitempty"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	// Initialize Configuration Manager which now handles themes as well
	cm, err := configuration.NewConfigurationManager()
	if err != nil {
		slog.Warn("FATAL: Failed to initialize configuration manager", "error", err)
		// In the new design, config manager is critical. We could panic here.
		// For now, we'll let it proceed, but operations will fail.
	}

	return &App{
		service:            requester.NewService(cm),
		collectionManager:  collection.NewCollectionManager(),
		environmentManager: environment.NewEnvironmentManager(),
		configManager:      cm,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	slog.Info("Application starting")
	// Theme manager is no longer needed, ConfigurationManager handles it.
	slog.Info("Application started")
}

// Execute actually performs the action of calling the server
func (a *App) Execute(options RequestOptions) (*requester.ResponseData, error) {
	execOpts := requester.ExecutionOptions{
		Method:   options.Method,
		URL:      options.URL,
		Body:     options.Body,
		Headers:  options.Headers,
		Cookies:  nil, // TODO: Add cookies to RequestOptions from frontend if needed
		Settings: options.Settings,
	}
	return a.service.ExecuteRequest(execOpts)
}

// Theme Management Methods (now routed to ConfigManager)

// GetActiveTheme returns the currently active theme name
func (a *App) GetActiveTheme() string {
	if a.configManager == nil {
		return "default-light"
	}
	return a.configManager.GetActiveTheme()
}

// SetActiveTheme sets the active theme
func (a *App) SetActiveTheme(themeName string) error {
	if a.configManager == nil {
		return fmt.Errorf("configuration manager not initialized")
	}
	return a.configManager.SetActiveTheme(themeName)
}

// GetCustomThemes returns all user-created themes
func (a *App) GetCustomThemes() []theme.Theme {
	if a.configManager == nil {
		return []theme.Theme{}
	}
	return a.configManager.GetCustomThemes()
}

// GetAllThemes returns both predefined and custom themes
func (a *App) GetAllThemes() []theme.Theme {
	if a.configManager == nil {
		return []theme.Theme{}
	}
	return a.configManager.GetAllThemes()
}

// SaveCustomTheme saves or updates a custom theme
func (a *App) SaveCustomTheme(theme theme.Theme) error {
	if a.configManager == nil {
		return fmt.Errorf("configuration manager not initialized")
	}
	return a.configManager.SaveCustomTheme(theme)
}

// DeleteCustomTheme removes a custom theme
func (a *App) DeleteCustomTheme(themeName string) error {
	if a.configManager == nil {
		return fmt.Errorf("configuration manager not initialized")
	}
	return a.configManager.DeleteCustomTheme(themeName)
}

// GetThemeByName returns a specific theme by name
func (a *App) GetThemeByName(themeName string) (*theme.Theme, error) {
	if a.configManager == nil {
		return nil, fmt.Errorf("configuration manager not initialized")
	}
	return a.configManager.GetThemeByName(themeName)
}

// Collection Management Methods

func (a *App) CreateCollection(collectionName string) error {
	return a.collectionManager.CreateCollection(collectionName)
}

func (a *App) LoadCollections() (*[]string, error) {
	return a.collectionManager.LoadCollections()
}

func (a *App) LoadCollection(collectionName string) (*collection.Collection, error) {
	return a.collectionManager.LoadCollection(collectionName)
}

func (a *App) UpdateCollection(updated collection.Collection) error {
	return a.collectionManager.UpdateCollection(updated)
}

func (a *App) DeleteCollection(collectionName string) error {
	return a.collectionManager.DeleteCollection(collectionName)
}

// Configuration Management Methods

func (a *App) GetConfiguration() (configuration.Configuration, error) {
	if a.configManager == nil {
		return configuration.Configuration{}, fmt.Errorf("configuration manager not initialized")
	}
	return a.configManager.Get(), nil
}

func (a *App) UpdateConfiguration(cfg configuration.Configuration) error {
	if a.configManager == nil {
		return fmt.Errorf("configuration manager not initialized")
	}
	// Save also updates the internal in-memory config safely
	return a.configManager.Save(cfg)
}

func (a *App) GetDefaultConfiguration() (configuration.Configuration, error) {
	if a.configManager == nil {
		return configuration.Configuration{}, fmt.Errorf("configuration manager not initialized")
	}
	return a.configManager.GetDefaultConfiguration(), nil
}

// Request Management Methods

func (a *App) GetRequests(collectionName string) (*[]collection.Request, error) {
	return a.collectionManager.GetRequests(collectionName)
}

func (a *App) AddRequest(collectionName string, request collection.Request) error {
	return a.collectionManager.AddRequest(collectionName, request)
}

func (a *App) RemoveRequest(collectionName string, requestId string) error {
	return a.collectionManager.RemoveRequest(collectionName, requestId)
}

func (a *App) UpdateRequest(collectionName string, updated collection.Request) error {
	return a.collectionManager.UpdateRequest(collectionName, updated)
}

// Environment Management Methods

func (a *App) CreateEnvironment(name string) error {
	return a.environmentManager.CreateEnvironment(name)
}

func (a *App) LoadEnvironments() (*[]string, error) {
	return a.environmentManager.LoadEnvironments()
}

func (a *App) LoadEnvironment(name string) (*environment.Environment, error) {
	return a.environmentManager.LoadEnvironment(name)
}

func (a *App) UpdateEnvironment(updated environment.Environment) error {
	return a.environmentManager.UpdateEnvironment(&updated)
}

func (a *App) DeleteEnvironment(name string) error {
	return a.environmentManager.DeleteEnvironment(name)
}

// Environment Value Management Methods

func (a *App) GetValues(environmentName string) (*map[string]environment.ValueType, error) {
	return a.environmentManager.GetValues(environmentName)
}

func (a *App) AddValue(environmentName string, valueName string, value environment.ValueType) error {
	return a.environmentManager.AddValue(environmentName, valueName, value)
}

func (a *App) RemoveValue(environmentName string, valueName string) error {
	return a.environmentManager.RemoveValue(environmentName, valueName)
}

func (a *App) UpdateValue(environmentName string, valueName string, updated environment.ValueType) error {
	return a.environmentManager.UpdateValue(environmentName, valueName, updated)
}

func (a *App) ResolveRequestPlaceholders(reqId string, collName string, envId string) (*map[string]environment.ValueType, error) {
	req, err := a.collectionManager.GetRequest(collName, reqId)

	if err != nil {
		return nil, err
	}

	env, err := a.environmentManager.LoadEnvironment(envId)

	if err != nil {
		return nil, err
	}

	resolvedMap := env.GetSelectedValues(req.GetPlaceholders())

	return resolvedMap, nil
}

// Changes log level at runtime
func (a *App) SetDebugMode(enabled bool) {
	if enabled {

		programLevel.Set(slog.LevelDebug)
		slog.Debug("Enabling debug mode")

	} else {

		slog.Debug("Disabling debug mode")
		programLevel.Set(slog.LevelInfo)
	}
}
