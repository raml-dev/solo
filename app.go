package main

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"yapla/internal/collection"
	"yapla/internal/configuration"
	"yapla/internal/environment"
	"yapla/internal/git"
	"yapla/internal/host"
	"yapla/internal/importer"
	"yapla/internal/requester"
	"yapla/internal/runner"
	"yapla/internal/script"
	"yapla/internal/theme"
	"yapla/internal/tools"

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
	scriptManager      *script.ScriptManager
	runner             *runner.Runner
	gitManager         *git.Manager
  closingMu           sync.Mutex
  isClosing           bool
}

type RequestOptions struct {
	Method              string                                 `json:"method"`
	URL                 string                                 `json:"url"`
	Headers             map[string]any                         `json:"headers"`
	Body                string                                 `json:"body"`
	Settings            *configuration.RequestSettingsOverride `json:"settings,omitempty"`
	PreRequestScript    string                                 `json:"preRequestScript,omitempty"`
	PostResponseScript  string                                 `json:"postResponseScript,omitempty"`
}

// RunParallel performs parallel HTTP requests for load testing.
func (a *App) RunParallel(options RequestOptions, concurrency, iterations int, stopOnError bool) (runner.RunnerStats, error) {
	if a.runner == nil {
		return runner.RunnerStats{}, fmt.Errorf("runner not initialized")
	}

	execOpts := requester.ExecutionOptions{
		Method:             options.Method,
		URL:                options.URL,
		Body:               options.Body,
		Headers:            options.Headers,
		Cookies:            nil,
		Settings:           options.Settings,
		PreRequestScript:   options.PreRequestScript,
		PostResponseScript: options.PostResponseScript,
	}

	opts := runner.RunnerOptions{
		Concurrency: concurrency,
		Iterations:  iterations,
		StopOnError: stopOnError,
		Request:     execOpts,
	}

	onResult := func(res runner.RunnerResult) {
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "runner:result", res)
		}
	}

	return a.runner.Run(a.ctx, opts, onResult), nil
}

// dummy function to emit RunnerResult
func (a *App) GetRunnerResult() runner.RunnerResult {
    return runner.RunnerResult{}
}

// NewApp creates a new App application struct
func NewApp() *App {
	cm, err := configuration.NewConfigurationManager()
	if err != nil {
		slog.Warn("FATAL: Failed to initialize configuration manager", "error", err)
	}

	hm := host.NewHostManager()

	// ScriptManager is created without context here; context is set in startup()
	sm := script.NewScriptManager(context.TODO())
	service := requester.NewService(cm, sm, hm)

	return &App{
		service:            service,
		collectionManager:  collection.NewCollectionManager(),
		environmentManager: environment.NewEnvironmentManager(),
		configManager:      cm,
		hostManager:        hm,
		scriptManager:      sm,
		runner:             runner.NewRunner(service),
		gitManager:         git.NewManager(),
    closingMu:          sync.Mutex{},
    isClosing:          false,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// Inject the real Wails context into ScriptManager so env.set() can emit events
	if a.scriptManager != nil {
		a.scriptManager.SetContext(ctx)
	}
	slog.Info("Application started")
}

// beforeClose is called when the user tries to close the application.
// We emit an event to the frontend to check for unsaved changes and veto the close.
func (a *App) beforeClose(ctx context.Context) bool {
    a.closingMu.Lock()
    defer a.closingMu.Unlock()

    if a.isClosing {
        return false // Already confirmed — let it close
    }

    runtime.EventsEmit(ctx, "app:request-close")
    return true // Veto the close
}

func (a *App) ForceQuit() {
    a.closingMu.Lock()
    a.isClosing = true
    a.closingMu.Unlock()

    if a.ctx != nil {
        runtime.Quit(a.ctx)
    }
}

// Execute performs the HTTP request with the given options.
func (a *App) Execute(options RequestOptions) (*requester.ResponseData, error) {
	execOpts := requester.ExecutionOptions{
		Method:             options.Method,
		URL:                options.URL,
		Body:               options.Body,
		Headers:            options.Headers,
		Cookies:            nil,
		Settings:           options.Settings,
		PreRequestScript:   options.PreRequestScript,
		PostResponseScript: options.PostResponseScript,
	}
	return a.service.ExecuteRequest(execOpts)
}

// GetSessionVars returns the current in-memory session variables set by scripts.
func (a *App) GetSessionVars() map[string]string {
	if a.scriptManager == nil {
		return map[string]string{}
	}
	return a.scriptManager.GetSessionVars()
}

// RemoveSessionVar removes a single session variable set by scripts.
func (a *App) RemoveSessionVar(key string) {
	if a.scriptManager != nil {
		a.scriptManager.RemoveSessionVar(key)
	}
}

// ClearSessionVars clears all session variables set by scripts.
func (a *App) ClearSessionVars() {
	if a.scriptManager != nil {
		a.scriptManager.ClearSessionVars()
	}
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

// ImportPostmanCollection imports a Postman v2.1 collection file into Yapla.
func (a *App) ImportPostmanCollection(path string) error {
	imp := importer.NewPostmanImporter()

	coll, err := imp.Import(path)
	if err != nil {
		return fmt.Errorf("failed to import collection: %w", err)
	}

	// Salviamo la collection importata usando il manager esistente.
	// UpdateCollection scriverà l'intero oggetto JSON su disco usando il suo Name.
	if err := a.collectionManager.UpdateCollection(*coll); err != nil {
		return fmt.Errorf("failed to save imported collection: %w", err)
	}

	return nil
}

// ImportBrunoCollection imports a Bruno collection from a directory.
func (a *App) ImportBrunoCollection(path string) error {
	imp := importer.NewBrunoImporter()

	coll, err := imp.Import(path)
	if err != nil {
		return fmt.Errorf("failed to import Bruno collection: %w", err)
	}

	if err := a.collectionManager.UpdateCollection(*coll); err != nil {
		return fmt.Errorf("failed to save imported Bruno collection: %w", err)
	}

	return nil
}

// ImportPostmanEnvironment imports a Postman environment JSON file.
func (a *App) ImportPostmanEnvironment(path string, overwrite bool) error {
	imp := importer.NewPostmanEnvironmentImporter()

	env, err := imp.Import(path)
	if err != nil {
		return fmt.Errorf("failed to import Postman environment: %w", err)
	}

	if a.environmentManager != nil {
		existing, loadErr := a.environmentManager.LoadEnvironment(env.Name)
		if loadErr == nil && existing != nil && !overwrite {
			return fmt.Errorf("environment %s already exists", env.Name)
		}
	}

	if err := a.environmentManager.UpdateEnvironment(env); err != nil {
		return fmt.Errorf("failed to save imported environment: %w", err)
	}

	return nil
}

// ImportBrunoEnvironment imports a Bruno environment .bru file.
func (a *App) ImportBrunoEnvironment(path string, overwrite bool) error {
	imp := importer.NewBrunoEnvironmentImporter()

	env, err := imp.Import(path)
	if err != nil {
		return fmt.Errorf("failed to import Bruno environment: %w", err)
	}

	if a.environmentManager != nil {
		existing, loadErr := a.environmentManager.LoadEnvironment(env.Name)
		if loadErr == nil && existing != nil && !overwrite {
			return fmt.Errorf("environment %s already exists", env.Name)
		}
	}

	if err := a.environmentManager.UpdateEnvironment(env); err != nil {
		return fmt.Errorf("failed to save imported environment: %w", err)
	}

	return nil
}

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
func (a *App) AddRequest(collectionName string, request collection.Request) (*collection.Request, error) {
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

// Git Management Methods

// IdentifyGitProvider returns the provider name for the given Git URL.
func (a *App) IdentifyGitProvider(url string) (string, error) {
	if a.gitManager == nil {
		return "", fmt.Errorf("git manager not initialized")
	}
	return a.gitManager.IdentifyProvider(url), nil
}

// GetGitRemoteBranches returns a list of branches for a given Git URL.
func (a *App) GetGitRemoteBranches(url string) ([]string, error) {
	if a.gitManager == nil {
		return nil, fmt.Errorf("git manager not initialized")
	}
	return a.gitManager.GetRemoteBranches(url)
}

// BrowseGitRemote returns a list of files and folders in a remote Git repository.
func (a *App) BrowseGitRemote(url string) ([]git.GitResource, error) {
	if a.gitManager == nil {
		return nil, fmt.Errorf("git manager not initialized")
	}
	return a.gitManager.BrowseRemote(url)
}

// SetupGitCollection clones a repository and sets up sparse-checkout for a specific collection path.
func (a *App) SetupGitCollection(url, remotePath, localName, providerType string) error {
	if a.gitManager == nil {
		return fmt.Errorf("git manager not initialized")
	}

	// Generate a unique storage directory name based on URL hash to avoid conflicts
	hash := sha1.Sum([]byte(url))
	storageDirName := fmt.Sprintf("%x", hash[:8])

	// Determine local target directory: ~/.yapla/git_storage/<hash>
	configRoot, err := tools.GetOrCreateConfigDir()
	if err != nil {
		return err
	}
	targetDir := filepath.Join(configRoot, tools.GIT_STORAGE_DIR, storageDirName)

	if err := a.gitManager.SetupGitCollection(url, remotePath, targetDir); err != nil {
		return err
	}

	// Detect format and create/update the local Yapla collection metadata
	var coll collection.Collection

	// Load the file content to create the metadata
	fullPath := filepath.Join(targetDir, remotePath)
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		return fmt.Errorf("file not found after git setup: %w", err)
	}

	if fileInfo.IsDir() {
		// Bruno folder
		imp := importer.NewBrunoImporter()
		c, err := imp.Import(fullPath)
		if err != nil {
			return err
		}
		coll = *c
	} else {
		// File (Yapla or Postman)
		data, err := os.ReadFile(fullPath)
		if err != nil {
			return err
		}

		slog.Debug("File read from Git repo", "path", fullPath, "size", len(data))

		// Try to detect format by attempting a Yapla-native unmarshal first.
		// Yapla files contain "creationTimestamp" and "requests" fields;
		// Postman files contain "info" and "item" fields instead.
		// NOTE: "yapla_version" does NOT exist in the Collection struct,
		// so it must never be used as a detection key.
		var tryYapla collection.Collection
		if err := json.Unmarshal(data, &tryYapla); err == nil && tryYapla.Name != "" && tryYapla.Id != "" {
			slog.Debug("Detected Yapla native format", "name", tryYapla.Name, "requests", len(tryYapla.Requests))
			coll = tryYapla
		} else {
			slog.Debug("Detected Postman format, calling importer")
			// Postman
			imp := importer.NewPostmanImporter()
			c, err := imp.Import(fullPath)
			if err != nil {
				return err
			}
			coll = *c
		}
	}

	// Update metadata
	// If localName is empty, we use the name found in the file, otherwise we keep localName
	if localName == "" {
		if coll.Name == "" {
			// Fallback to filename if still empty
			coll.Name = filepath.Base(remotePath)
		}
	} else {
		coll.Name = localName
	}

	coll.GitRemote = url
	coll.GitPath = remotePath
	coll.GitProvider = providerType

	return a.collectionManager.UpdateCollection(coll)
}

// SyncGitCollection performs pull, commit and push for a Git-backed collection.
func (a *App) SyncGitCollection(collectionId string) error {
	gitRepoDir, targetColl, err := a.resolveGitCollectionDir(collectionId)
	if err != nil {
		return err
	}
	gitFilePath := filepath.Join(gitRepoDir, targetColl.GitPath)

	// 3. Save current Yapla state to the file in the Git repo before syncing
	// This ensures we push our latest local changes
	jsonData, err := json.MarshalIndent(targetColl, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(gitFilePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to update file in git repo: %w", err)
	}

	// 4. Perform Sync (Pull, Commit, Push)
	if err := a.gitManager.SyncGitCollection(gitRepoDir, targetColl.Name, ""); err != nil {
		return err
	}

	// 5. Reload the collection from the synced git file to update local Yapla state
	updatedData, err := os.ReadFile(gitFilePath)
	if err == nil {
		var updatedColl collection.Collection
		if err := json.Unmarshal(updatedData, &updatedColl); err == nil {
			// Keep our local metadata
			updatedColl.Id = targetColl.Id
			updatedColl.Name = targetColl.Name
			updatedColl.GitRemote = targetColl.GitRemote
			updatedColl.GitPath = targetColl.GitPath
			updatedColl.GitProvider = targetColl.GitProvider
			return a.collectionManager.UpdateCollection(updatedColl)
		}
	}

	return nil
}

// SetupGitEnvironment clones a repository and sets up sparse-checkout for a specific environment path.
func (a *App) SetupGitEnvironment(url, remotePath, localName, providerType string) error {
	if a.gitManager == nil {
		return fmt.Errorf("git manager not initialized")
	}

	// Generate a unique storage directory name based on URL hash
	hash := sha1.Sum([]byte(url))
	storageDirName := fmt.Sprintf("env_%x", hash[:8])

	// Determine local target directory: ~/.yapla/git_storage/<hash>
	configRoot, err := tools.GetOrCreateConfigDir()
	if err != nil {
		return err
	}
	targetDir := filepath.Join(configRoot, tools.GIT_STORAGE_DIR, storageDirName)

	if err := a.gitManager.SetupGitCollection(url, remotePath, targetDir); err != nil {
		return err
	}

	// Detect format and create/update the local Yapla environment metadata
	var env environment.Environment

	fullPath := filepath.Join(targetDir, remotePath)
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		return fmt.Errorf("environment file not found after git setup: %w", err)
	}

	if fileInfo.IsDir() {
		// Bruno environment folder (actually Bruno usually has .bru files)
		imp := importer.NewBrunoEnvironmentImporter()
		e, err := imp.Import(fullPath)
		if err != nil {
			return err
		}
		env = *e
	} else {
		data, err := os.ReadFile(fullPath)
		if err != nil {
			return err
		}

		slog.Debug("Environment file read from Git repo", "path", fullPath, "size", len(data))

		// Try to detect format
		content := string(data)
		if strings.Contains(content, "\"creation_timestamp\"") && strings.Contains(content, "\"values\"") {
			// Yapla Native
			if err := json.Unmarshal(data, &env); err != nil {
				return err
			}
		} else if strings.Contains(content, "\"values\"") && strings.Contains(content, "\"_postman_variable_scope\"") {
			// Postman
			imp := importer.NewPostmanEnvironmentImporter()
			e, err := imp.Import(fullPath)
			if err != nil {
				return err
			}
			env = *e
		} else if strings.HasSuffix(fullPath, ".bru") {
			// Bruno single file
			imp := importer.NewBrunoEnvironmentImporter()
			e, err := imp.Import(fullPath)
			if err != nil {
				return err
			}
			env = *e
		} else {
			return fmt.Errorf("unsupported environment format")
		}
	}

	// Update metadata
	if localName == "" {
		if env.Name == "" {
			env.Name = filepath.Base(remotePath)
		}
	} else {
		env.Name = localName
	}

	env.GitRemote = url
	env.GitPath = remotePath
	env.GitProvider = providerType

	return a.environmentManager.UpdateEnvironment(&env)
}

// SyncGitEnvironment performs pull, commit and push for a Git-backed environment.
func (a *App) SyncGitEnvironment(environmentId string) error {
	gitRepoDir, targetEnv, err := a.resolveGitEnvDir(environmentId)
	if err != nil {
		return err
	}
	gitFilePath := filepath.Join(gitRepoDir, targetEnv.GitPath)

	// Save current state to git repo
	jsonData, err := json.MarshalIndent(targetEnv, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(gitFilePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to update environment file in git repo: %w", err)
	}

	// Perform Sync
	if err := a.gitManager.SyncGitCollection(gitRepoDir, targetEnv.Name, ""); err != nil {
		return err
	}

	// Reload and update local state
	updatedData, err := os.ReadFile(gitFilePath)
	if err == nil {
		var updatedEnv environment.Environment
		if err := json.Unmarshal(updatedData, &updatedEnv); err == nil {
			updatedEnv.Id = targetEnv.Id
			updatedEnv.Name = targetEnv.Name
			updatedEnv.GitRemote = targetEnv.GitRemote
			updatedEnv.GitPath = targetEnv.GitPath
			updatedEnv.GitProvider = targetEnv.GitProvider
			return a.environmentManager.UpdateEnvironment(&updatedEnv)
		}
	}

	return nil
}

// SelectDirectory opens a native file dialog to select a directory.
func (a *App) SelectDirectory(title string) (string, error) {
	slog.Debug("Opening directory dialog", "title", title)
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: title,
	})
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

// ── Git helpers ──────────────────────────────────────────────────────────────

// resolveGitCollectionDir resolves a collectionId to its local git repo dir
// and returns the collection metadata. Avoids duplicating the hash logic everywhere.
func (a *App) resolveGitCollectionDir(collectionId string) (gitRepoDir string, coll *collection.Collection, err error) {
	collsContent, err := a.collectionManager.LoadCollectionsContent()
	if err != nil {
		return "", nil, err
	}

	var target *collection.Collection
	for i, c := range *collsContent {
		if c.Id == collectionId {
			target = &(*collsContent)[i]
			break
		}
	}

	if target == nil || target.GitRemote == "" {
		return "", nil, fmt.Errorf("collection not found or not Git-backed")
	}

	configRoot, err := tools.GetOrCreateConfigDir()
	if err != nil {
		return "", nil, err
	}

	hash := sha1.Sum([]byte(target.GitRemote))
	storageDirName := fmt.Sprintf("%x", hash[:8])
	dir := filepath.Join(configRoot, tools.GIT_STORAGE_DIR, storageDirName)

	return dir, target, nil
}

// ── Git Status Panel ─────────────────────────────────────────────────────────

// GetGitCollectionStatus returns the current git status for a Git-backed collection.
func (a *App) GetGitCollectionStatus(collectionId string) (git.CollectionStatus, error) {
	if a.gitManager == nil {
		return git.CollectionStatus{}, fmt.Errorf("git manager not initialized")
	}
	dir, _, err := a.resolveGitCollectionDir(collectionId)
	if err != nil {
		return git.CollectionStatus{}, err
	}
	return a.gitManager.GetCollectionStatus(dir)
}

// GitKeepOurs resolves all conflicts in a Git-backed collection by keeping our version.
func (a *App) GitKeepOurs(collectionId string) error {
	if a.gitManager == nil {
		return fmt.Errorf("git manager not initialized")
	}
	dir, _, err := a.resolveGitCollectionDir(collectionId)
	if err != nil {
		return err
	}
	return a.gitManager.KeepOurs(dir)
}

// GitKeepTheirs resolves all conflicts in a Git-backed collection by keeping the remote version.
func (a *App) GitKeepTheirs(collectionId string) error {
	if a.gitManager == nil {
		return fmt.Errorf("git manager not initialized")
	}
	dir, coll, err := a.resolveGitCollectionDir(collectionId)
	if err != nil {
		return err
	}
	if err := a.gitManager.KeepTheirs(dir); err != nil {
		return err
	}
	// After accepting remote changes, reload the collection from the git file
	// so the local Yapla state reflects the remote version.
	gitFilePath := filepath.Join(dir, coll.GitPath)
	data, err := os.ReadFile(gitFilePath)
	if err != nil {
		return nil // best-effort, don't fail the whole operation
	}
	var updatedColl collection.Collection
	if err := json.Unmarshal(data, &updatedColl); err != nil {
		return nil
	}
	updatedColl.Id = coll.Id
	updatedColl.Name = coll.Name
	updatedColl.GitRemote = coll.GitRemote
	updatedColl.GitPath = coll.GitPath
	updatedColl.GitProvider = coll.GitProvider
	return a.collectionManager.UpdateCollection(updatedColl)
}

// GitAbortRebase aborts an in-progress rebase for a Git-backed collection.
func (a *App) GitAbortRebase(collectionId string) error {
	if a.gitManager == nil {
		return fmt.Errorf("git manager not initialized")
	}
	dir, _, err := a.resolveGitCollectionDir(collectionId)
	if err != nil {
		return err
	}
	return a.gitManager.AbortRebase(dir)
}

// GitDiscardChanges discards all local uncommitted changes for a Git-backed collection.
func (a *App) GitDiscardChanges(collectionId string) error {
	if a.gitManager == nil {
		return fmt.Errorf("git manager not initialized")
	}
	dir, _, err := a.resolveGitCollectionDir(collectionId)
	if err != nil {
		return err
	}
	return a.gitManager.DiscardAllChanges(dir)
}

// ── Open in Terminal ─────────────────────────────────────────────────────────

// OpenCollectionInTerminal opens the system terminal at the git storage directory
// for the given Git-backed collection.
func (a *App) OpenCollectionInTerminal(collectionId string) error {
	if a.gitManager == nil {
		return fmt.Errorf("git manager not initialized")
	}
	dir, _, err := a.resolveGitCollectionDir(collectionId)
	if err != nil {
		return err
	}
	return a.gitManager.OpenInTerminal(dir)
}

// ── Git Environment helpers ───────────────────────────────────────────────────

// resolveGitEnvDir resolves an environmentId to its local git repo dir
// and returns the environment metadata.
func (a *App) resolveGitEnvDir(environmentId string) (gitRepoDir string, env *environment.Environment, err error) {
	envsContent, err := a.environmentManager.LoadEnvironmentsContent()
	if err != nil {
		return "", nil, err
	}

	var target *environment.Environment
	for i, e := range *envsContent {
		if e.Id == environmentId {
			target = &(*envsContent)[i]
			break
		}
	}

	if target == nil || target.GitRemote == "" {
		return "", nil, fmt.Errorf("environment not found or not Git-backed")
	}

	configRoot, err := tools.GetOrCreateConfigDir()
	if err != nil {
		return "", nil, err
	}

	hash := sha1.Sum([]byte(target.GitRemote))
	storageDirName := fmt.Sprintf("env_%x", hash[:8])
	dir := filepath.Join(configRoot, tools.GIT_STORAGE_DIR, storageDirName)

	return dir, target, nil
}

// GetGitEnvironmentStatus returns the current git status for a Git-backed environment.
func (a *App) GetGitEnvironmentStatus(environmentId string) (git.CollectionStatus, error) {
	if a.gitManager == nil {
		return git.CollectionStatus{}, fmt.Errorf("git manager not initialized")
	}
	dir, _, err := a.resolveGitEnvDir(environmentId)
	if err != nil {
		return git.CollectionStatus{}, err
	}
	return a.gitManager.GetCollectionStatus(dir)
}

// GitEnvKeepOurs resolves all conflicts in a Git-backed environment by keeping our version.
func (a *App) GitEnvKeepOurs(environmentId string) error {
	if a.gitManager == nil {
		return fmt.Errorf("git manager not initialized")
	}
	dir, _, err := a.resolveGitEnvDir(environmentId)
	if err != nil {
		return err
	}
	return a.gitManager.KeepOurs(dir)
}

// GitEnvKeepTheirs resolves all conflicts in a Git-backed environment by keeping the remote version.
func (a *App) GitEnvKeepTheirs(environmentId string) error {
	if a.gitManager == nil {
		return fmt.Errorf("git manager not initialized")
	}
	dir, env, err := a.resolveGitEnvDir(environmentId)
	if err != nil {
		return err
	}
	if err := a.gitManager.KeepTheirs(dir); err != nil {
		return err
	}
	gitFilePath := filepath.Join(dir, env.GitPath)
	data, err := os.ReadFile(gitFilePath)
	if err != nil {
		return nil
	}
	var updatedEnv environment.Environment
	if err := json.Unmarshal(data, &updatedEnv); err != nil {
		return nil
	}
	updatedEnv.Id = env.Id
	updatedEnv.Name = env.Name
	updatedEnv.GitRemote = env.GitRemote
	updatedEnv.GitPath = env.GitPath
	updatedEnv.GitProvider = env.GitProvider
	return a.environmentManager.UpdateEnvironment(&updatedEnv)
}

// GitEnvAbortRebase aborts an in-progress rebase for a Git-backed environment.
func (a *App) GitEnvAbortRebase(environmentId string) error {
	if a.gitManager == nil {
		return fmt.Errorf("git manager not initialized")
	}
	dir, _, err := a.resolveGitEnvDir(environmentId)
	if err != nil {
		return err
	}
	return a.gitManager.AbortRebase(dir)
}

// GitEnvDiscardChanges discards all local uncommitted changes for a Git-backed environment.
func (a *App) GitEnvDiscardChanges(environmentId string) error {
	if a.gitManager == nil {
		return fmt.Errorf("git manager not initialized")
	}
	dir, _, err := a.resolveGitEnvDir(environmentId)
	if err != nil {
		return err
	}
	return a.gitManager.DiscardAllChanges(dir)
}

// OpenEnvironmentInTerminal opens the system terminal at the git storage directory
// for the given Git-backed environment.
func (a *App) OpenEnvironmentInTerminal(environmentId string) error {
	if a.gitManager == nil {
		return fmt.Errorf("git manager not initialized")
	}
	dir, _, err := a.resolveGitEnvDir(environmentId)
	if err != nil {
		return err
	}
	return a.gitManager.OpenInTerminal(dir)
}
