package main

import (
	"context"
	"fmt"
	"log/slog"
	"yapla/internal/collection"
	"yapla/internal/configuration"
	"yapla/internal/environment"
	"yapla/internal/host"
	"yapla/internal/requester"
	"yapla/internal/theme"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx                context.Context
	service            *requester.Service
	collectionManager  *collection.CollectionManager
	environmentManager *environment.EnvironmentManager
	configManager      *configuration.ConfigurationManager
	hostManager        *host.HostManager
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
		hostManager:        host.NewHostManager(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	slog.Info("Application starting")
	// Theme manager is no longer needed, ConfigurationManager handles it.
	slog.Info("Application started")
}

// Execute performs the HTTP request with the given options.
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

// GetActiveTheme returns the name of the currently active theme.
func (a *App) GetActiveTheme() string {
	if a.configManager == nil {
		return "default-light"
	}
	return a.configManager.GetActiveTheme()
}

// SetActiveTheme sets the active theme by name and persists the change.
func (a *App) SetActiveTheme(themeName string) error {
	if a.configManager == nil {
		return fmt.Errorf("configuration manager not initialized")
	}
	return a.configManager.SetActiveTheme(themeName)
}

// GetCustomThemes returns all user-created themes.
func (a *App) GetCustomThemes() []theme.Theme {
	if a.configManager == nil {
		return []theme.Theme{}
	}
	return a.configManager.GetCustomThemes()
}

// GetAllThemes returns all available themes, both predefined and custom.
func (a *App) GetAllThemes() []theme.Theme {
	if a.configManager == nil {
		return []theme.Theme{}
	}
	return a.configManager.GetAllThemes()
}

// SaveCustomTheme saves a new theme or updates an existing one.
func (a *App) SaveCustomTheme(theme theme.Theme) error {
	if a.configManager == nil {
		return fmt.Errorf("configuration manager not initialized")
	}
	return a.configManager.SaveCustomTheme(theme)
}

// DeleteCustomTheme removes a custom theme by name.
func (a *App) DeleteCustomTheme(themeName string) error {
	if a.configManager == nil {
		return fmt.Errorf("configuration manager not initialized")
	}
	return a.configManager.DeleteCustomTheme(themeName)
}

// GetThemeByName finds and returns a single theme (predefined or custom) by its name.
func (a *App) GetThemeByName(themeName string) (*theme.Theme, error) {
	if a.configManager == nil {
		return nil, fmt.Errorf("configuration manager not initialized")
	}
	return a.configManager.GetThemeByName(themeName)
}

// Host Management Methods

// GetAllHosts returns a list of all configured hosts.
func (a *App) GetAllHosts() ([]host.Host, error) {
	if a.hostManager == nil {
		return nil, fmt.Errorf("host manager not initialized")
	}
	return a.hostManager.GetAllHosts(), nil
}

// UpsertHost creates a new host configuration or updates an existing one.
func (a *App) UpsertHost(h host.Host) error {
	if a.hostManager == nil {
		return fmt.Errorf("host manager not initialized")
	}
	return a.hostManager.UpsertHost(h)
}

// DeleteHost removes a host configuration by its name (hostname).
func (a *App) DeleteHost(hostname string) error {
	if a.hostManager == nil {
		return fmt.Errorf("host manager not initialized")
	}
	return a.hostManager.DeleteHost(hostname)
}

// GetSelectedEnvironment returns the persisted selected environment name
func (a *App) GetSelectedEnvironment() string {
	if a.configManager == nil {
		return ""
	}
	return a.configManager.GetSelectedEnvironment()
}

// SetSelectedEnvironment persists the selected environment name
func (a *App) SetSelectedEnvironment(name string) error {
	if a.configManager == nil {
		return fmt.Errorf("configuration manager not initialized")
	}
	return a.configManager.SetSelectedEnvironment(name)
}

// Collection Management Methods

// CreateCollection creates a new, empty collection.
func (a *App) CreateCollection(collectionName string) error {
	return a.collectionManager.CreateCollection(collectionName)
}

// LoadCollections returns a list of all available collection names.
func (a *App) LoadCollections() (*[]string, error) {
	return a.collectionManager.LoadCollections()
}

// LoadCollection returns the full content of a single collection by name.
func (a *App) LoadCollection(collectionName string) (*collection.Collection, error) {
	return a.collectionManager.LoadCollection(collectionName)
}

// UpdateCollection saves the changes to an entire collection.
func (a *App) UpdateCollection(updated collection.Collection) error {
	return a.collectionManager.UpdateCollection(updated)
}

// DeleteCollection removes a collection by name.
func (a *App) DeleteCollection(collectionName string) error {
	return a.collectionManager.DeleteCollection(collectionName)
}

// Configuration Management Methods

// GetConfiguration returns the entire application configuration object.
func (a *App) GetConfiguration() (configuration.Configuration, error) {
	if a.configManager == nil {
		return configuration.Configuration{}, fmt.Errorf("configuration manager not initialized")
	}
	return a.configManager.Get(), nil
}

// UpdateConfiguration saves the entire application configuration object.
func (a *App) UpdateConfiguration(cfg configuration.Configuration) error {
	if a.configManager == nil {
		return fmt.Errorf("configuration manager not initialized")
	}
	// Save also updates the internal in-memory config safely
	return a.configManager.Save(cfg)
}

// GetDefaultConfiguration returns the default application configuration.
func (a *App) GetDefaultConfiguration() (configuration.Configuration, error) {
	if a.configManager == nil {
		return configuration.Configuration{}, fmt.Errorf("configuration manager not initialized")
	}
	return a.configManager.GetDefaultConfiguration(), nil
}

// Request Management Methods

// GetRequests returns all requests within a specific collection.
func (a *App) GetRequests(collectionName string) (*[]collection.Request, error) {
	return a.collectionManager.GetRequests(collectionName)
}

// AddRequest adds a new request to a specific collection.
func (a *App) AddRequest(collectionName string, request collection.Request) error {
	return a.collectionManager.AddRequest(collectionName, request)
}

// RemoveRequest removes a request from a collection using its ID.
func (a *App) RemoveRequest(collectionName string, requestId string) error {
	return a.collectionManager.RemoveRequest(collectionName, requestId)
}

// UpdateRequest updates an existing request within a collection.
func (a *App) UpdateRequest(collectionName string, updated collection.Request) error {
	return a.collectionManager.UpdateRequest(collectionName, updated)
}

// Environment Management Methods

// CreateEnvironment creates a new, empty environment.
func (a *App) CreateEnvironment(name string) error {
	return a.environmentManager.CreateEnvironment(name)
}

// LoadEnvironments returns a list of all available environment names.
func (a *App) LoadEnvironments() (*[]string, error) {
	return a.environmentManager.LoadEnvironments()
}

// LoadEnvironment returns the full content of a single environment by name.
func (a *App) LoadEnvironment(name string) (*environment.Environment, error) {
	return a.environmentManager.LoadEnvironment(name)
}

// UpdateEnvironment saves the changes to an entire environment.
func (a *App) UpdateEnvironment(updated environment.Environment) error {
	return a.environmentManager.UpdateEnvironment(&updated)
}

// DeleteEnvironment removes an environment by name.
func (a *App) DeleteEnvironment(name string) error {
	return a.environmentManager.DeleteEnvironment(name)
}

// Environment Value Management Methods

// GetValues returns all key-value pairs for a specific environment.
func (a *App) GetValues(environmentName string) (*map[string]environment.ValueType, error) {
	return a.environmentManager.GetValues(environmentName)
}

// AddValue adds a new variable to a specific environment.
func (a *App) AddValue(environmentName string, valueName string, value environment.ValueType) error {
	return a.environmentManager.AddValue(environmentName, valueName, value)
}

// RemoveValue removes a variable from a specific environment.
func (a *App) RemoveValue(environmentName string, valueName string) error {
	return a.environmentManager.RemoveValue(environmentName, valueName)
}

// UpdateValue updates an existing variable in a specific environment.
func (a *App) UpdateValue(environmentName string, valueName string, updated environment.ValueType) error {
	return a.environmentManager.UpdateValue(environmentName, valueName, updated)
}

// ResolveRequestPlaceholders identifies placeholders in a request and resolves them
// using values from the specified environment.
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

// SetDebugMode changes the application's log level at runtime.
func (a *App) SetDebugMode(enabled bool) {
	if enabled {

		programLevel.Set(slog.LevelDebug)
		slog.Debug("Enabling debug mode")

	} else {

		slog.Debug("Disabling debug mode")
		programLevel.Set(slog.LevelInfo)
	}
}

// SelectFile opens a native file dialog to select a file.
// It takes a title for the dialog and a pattern for file filtering (e.g., "*.pem;*.crt").
func (a *App) SelectFile(title, patterns, displayName string) (string, error) {
	slog.Debug("Opening file dialog", "title", title, "patterns", patterns)
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: title,
		Filters: []runtime.FileFilter{
			{
				DisplayName: displayName,
				Pattern:     patterns, // e.g., "*.pem;*.crt;*.key"
			},
		},
	})
}
