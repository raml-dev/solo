<script lang="ts">
  import { onMount } from "svelte";
  import { derived } from "svelte/store";
  import { configurationStore } from "../stores/configurationStore";
  import type { theme } from "../../../wailsjs/go/models";
  import { configuration } from "../../../wailsjs/go/models";
  import { GetDefaultConfiguration } from "../../../wailsjs/go/main/App";
  import Button from "./base/Button.svelte";
  import Modal from "./base/Modal.svelte";

  // --- Exported for parent binding ---
  export let isDirty = false;
  export let isLoading = false;
  export let save: () => Promise<void>;
  export let revert: () => void;

  // --- Component State ---
  let error: string | null = null;
  let successMessage: string | null = null;
  let showThemeCustomizer = false;
  let showThemes = false;

  // --- Theme Preview State ---
  let originalThemeName: string = "";
  let originalThemeColors: Record<string, string> = {};
  let previewThemeName: string = "";

  // --- Configuration Objects ---
  function createEmptyConfig() {
    const cfg = new configuration.Configuration();
    cfg.general = new configuration.GeneralSettings();
    cfg.request = new configuration.RequestSettings();
    cfg.customThemes = [] as theme.Theme[];
    return cfg;
  }

  // Default config from backend, used for placeholders
  let defaultConfig: configuration.Configuration = createEmptyConfig();
  // Pristine copy of the config, to detect changes
  let pristineConfig: configuration.Configuration = createEmptyConfig();
  // The config object bound to the form fields
  let editableConfig: configuration.Configuration = createEmptyConfig();

  // --- Dirty Check ---
  // Reactive statement to check if the form is dirty.
  // Compares editable config with pristine state AND theme preview with original.
  $: {
    if (pristineConfig.request && editableConfig.request) {
      const themeChanged = previewThemeName !== "" && previewThemeName !== originalThemeName;
      const stringPristine = JSON.stringify({
        general: { checkForUpdates: pristineConfig.general.checkForUpdates },
        request: pristineConfig.request
      });
      const stringEditable = JSON.stringify({
        general: { checkForUpdates: editableConfig.general.checkForUpdates },
        request: editableConfig.request
      });
      isDirty = stringPristine !== stringEditable || themeChanged;
    } else {
      isDirty = false;
    }
  }

  // --- Store Subscriptions ---
  const { config, allThemes } = configurationStore;
  const customThemes = derived([config], ([$config]) => $config?.customThemes || []);
  const predefinedThemes = derived([allThemes, customThemes], ([$allThemes, $customThemes]) => {
    if (!$allThemes) return [];
    const customThemeNames = new Set($customThemes.map((t) => t.name));
    return $allThemes.filter((t) => !customThemeNames.has(t.name));
  });

  // --- Lifecycle ---
  onMount(async () => {
    try {
      const loadedDefaults = await GetDefaultConfiguration();
      defaultConfig = new configuration.Configuration(loadedDefaults);
      if (!defaultConfig.general) defaultConfig.general = new configuration.GeneralSettings();
      if (!defaultConfig.request) defaultConfig.request = new configuration.RequestSettings();
    } catch (err) {
      console.error("Failed to load default configuration:", err);
    }

    const unsubscribe = config.subscribe((value) => {
      if (value?.request) {
        const deepCopiedValue = new configuration.Configuration(JSON.parse(JSON.stringify(value)));
        if (!deepCopiedValue.general) deepCopiedValue.general = new configuration.GeneralSettings();
        if (!deepCopiedValue.request) deepCopiedValue.request = new configuration.RequestSettings();
        editableConfig = deepCopiedValue;
        pristineConfig = new configuration.Configuration(
          JSON.parse(JSON.stringify(deepCopiedValue))
        );

        // Capture original theme state for preview/revert functionality
        if (!originalThemeName) {
          originalThemeName = value.general?.activeTheme || "default-light";
          previewThemeName = originalThemeName;

          // Store original theme colors for revert
          const themes = $allThemes;
          const originalTheme = themes.find((t) => t.name === originalThemeName);
          if (originalTheme) {
            originalThemeColors = { ...originalTheme.colors };
          }
        }
      }
    });
    unsubscribe();
  });

  // --- Theme Preview Helpers ---
  function applyThemeToDom(colors: Record<string, string>) {
    const root = document.documentElement;
    Object.entries(colors).forEach(([key, value]) => {
      root.style.setProperty(`--${key}`, value);
    });
  }

  function handleThemeChange(themeName: string) {
    // Find the theme and apply it as a preview (no persistence)
    const themes = $allThemes;
    const selectedTheme = themes.find((t) => t.name === themeName);
    if (selectedTheme) {
      previewThemeName = themeName;
      applyThemeToDom(selectedTheme.colors);
    }
  }

  function handleRevert() {
    // Restore original theme if preview differs
    if (previewThemeName !== originalThemeName && Object.keys(originalThemeColors).length > 0) {
      applyThemeToDom(originalThemeColors);
      previewThemeName = originalThemeName;
    }
  }

  // --- Event Handlers ---
  async function handleSaveSettings() {
    try {
      isLoading = true;
      error = null;
      successMessage = null;

      // Ensure numbers are valid integers
      if (editableConfig.request) {
        editableConfig.request.timeoutSeconds =
          parseInt(String(editableConfig.request.timeoutSeconds), 10) || 0;
        editableConfig.request.maxRedirects =
          parseInt(String(editableConfig.request.maxRedirects), 10) || 0;
      }

      const currentConfig = $config;
      currentConfig.general.checkForUpdates = editableConfig.general.checkForUpdates;
      currentConfig.request = editableConfig.request;

      await configurationStore.save(currentConfig);

      // Persist theme if changed
      if (previewThemeName !== originalThemeName) {
        await configurationStore.changeTheme(previewThemeName);
        // Update original state after successful save
        originalThemeName = previewThemeName;
        const themes = $allThemes;
        const savedTheme = themes.find((t) => t.name === previewThemeName);
        if (savedTheme) {
          originalThemeColors = { ...savedTheme.colors };
        }
      }

      // Update pristine state to reset isDirty flag
      pristineConfig = new configuration.Configuration(JSON.parse(JSON.stringify(editableConfig)));

      successMessage = "Settings saved successfully";
      setTimeout(() => {
        successMessage = null;
      }, 3000);
    } catch (err) {
      console.error("Error saving settings:", err);
      error = String(err);
    } finally {
      isLoading = false;
    }
  }

  // Assign exports after functions are defined
  save = handleSaveSettings;
  revert = handleRevert;

  function getThemePreview(theme: theme.Theme): string {
    return theme.colors["bg-primary"] || "#ffffff";
  }
</script>

<div class="configuration-container">
  <!-- General Settings Section -->
  <div class="config-section">
    <h4>General</h4>
    <div class="form-group">
      <label class="checkbox-label">
        <input type="checkbox" bind:checked={editableConfig.general.checkForUpdates} />
        Check for updates on startup
      </label>
    </div>
  </div>

  <!-- Request Defaults Section -->
  <div class="config-section">
    <h4>Request Defaults</h4>
    <div class="form-group">
      <label for="timeout">Timeout (seconds)</label>
      <input
        id="timeout"
        type="number"
        bind:value={editableConfig.request.timeoutSeconds}
        min="0"
        step="1"
        placeholder={`Default: ${defaultConfig.request.timeoutSeconds}`}
      />
    </div>
    <div class="form-group">
      <label for="user-agent">Default User Agent</label>
      <input
        id="user-agent"
        type="text"
        bind:value={editableConfig.request.defaultUserAgent}
        placeholder={`Default: ${defaultConfig.request.defaultUserAgent}`}
      />
    </div>
    <div class="form-group">
      <label class="checkbox-label">
        <input type="checkbox" bind:checked={editableConfig.request.followRedirects} />
        Follow Redirects
      </label>
    </div>
    <div class="form-group">
      <label for="max-redirects">Max Redirects</label>
      <input
        id="max-redirects"
        type="number"
        bind:value={editableConfig.request.maxRedirects}
        min="0"
        step="1"
        placeholder={`Default: ${defaultConfig.request.maxRedirects}`}
        disabled={!editableConfig.request.followRedirects}
      />
    </div>
    <div class="form-group">
      <label class="checkbox-label">
        <input type="checkbox" bind:checked={editableConfig.request.validateSSL} />
        Validate SSL Certificates
      </label>
    </div>
    <div class="form-group">
      <label for="proxy">Proxy URL</label>
      <input
        id="proxy"
        type="text"
        bind:value={editableConfig.request.proxyUrl}
        placeholder="http://user:pass@host:port (optional)"
      />
    </div>
  </div>

  <!-- Theme Selector Section -->
  <div class="config-section">
    <div
      class="section-header accordion-header"
      on:click={() => (showThemes = !showThemes)}
      on:keypress={() => (showThemes = !showThemes)}
      role="button"
      tabindex="0"
    >
      <h4>Theme</h4>
      <span class="accordion-icon" class:expanded={showThemes}>›</span>
    </div>

    {#if showThemes}
      <div class="theme-section-content">
        <div class="theme-section">
          <h5 class="text-sm font-medium text-muted">Predefined Themes</h5>
          <div class="theme-grid">
            {#each $predefinedThemes as theme (theme.name)}
              <button
                class="theme-card"
                class:active={previewThemeName === theme.name}
                on:click={() => handleThemeChange(theme.name)}
                title={`Activate ${theme.name} theme`}
              >
                <div class="theme-preview" style="background: {getThemePreview(theme)}" />
                <div class="theme-colors">
                  <span class="color-dot" style="background: {theme.colors.primary}" />
                  <span class="color-dot" style="background: {theme.colors.success}" />
                  <span class="color-dot" style="background: {theme.colors.warning}" />
                  <span class="color-dot" style="background: {theme.colors.danger}" />
                </div>
                <span class="theme-name">{theme.name}</span>
              </button>
            {/each}
          </div>
        </div>
        {#if $customThemes.length > 0}
          <div class="theme-section">
            <h5 class="text-sm font-medium text-muted">Custom Themes</h5>
            <div class="theme-grid">
              {#each $customThemes as theme (theme.name)}
                <button
                  class="theme-card"
                  class:active={previewThemeName === theme.name}
                  on:click={() => handleThemeChange(theme.name)}
                  title={`Activate ${theme.name} theme`}
                >
                  <div class="theme-preview" style="background: {getThemePreview(theme)}" />
                  <div class="theme-colors">
                    <span class="color-dot" style="background: {theme.colors.primary}" />
                    <span class="color-dot" style="background: {theme.colors.success}" />
                    <span class="color-dot" style="background: {theme.colors.warning}" />
                    <span class="color-dot" style="background: {theme.colors.danger}" />
                  </div>
                  <span class="theme-name">{theme.name}</span>
                </button>
              {/each}
            </div>
          </div>
        {/if}
        <Button variant="secondary" click={() => (showThemeCustomizer = true)}
          >Create Custom Theme</Button
        >
        {#if showThemeCustomizer}
          <Modal title="Create Custom Theme" toggleFn={() => (showThemeCustomizer = false)}>
            <p class="text-muted">Theme customizer UI will be implemented here.</p>
          </Modal>
        {/if}
      </div>
    {/if}
  </div>
  {#if error}
    <div class="error">{error}</div>
  {/if}
  {#if successMessage}
    <div class="success">{successMessage}</div>
  {/if}
</div>

<style>
  .config-section {
    padding: var(--space-md) var(--space-lg);
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
  }

  h4 {
    margin: 0 0 var(--space-md) 0;
    font-size: var(--font-size-md);
    font-weight: var(--font-weight-medium);
    color: var(--primary);
    border-bottom: 1px solid var(--border);
    padding-bottom: var(--space-xs);
  }

  h5 {
    margin-bottom: var(--space-sm);
  }

  .section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-sm);
  }

  .accordion-header {
    cursor: pointer;
    padding: var(--space-xs);
    border-radius: var(--radius-md);
    transition: background-color var(--transition-fast);
  }

  .accordion-header:hover {
    background-color: var(--bg-tertiary);
  }

  .accordion-icon {
    font-size: 1.5rem;
    font-weight: bold;
    color: var(--text-muted);
    transition: transform var(--transition-base);
  }

  .accordion-icon.expanded {
    transform: rotate(90deg);
  }

  .theme-section-content {
    padding-top: var(--space-md);
    border-top: 1px solid var(--border-dark);
    margin-top: var(--space-sm);
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: var(--space-xs);
  }

  label:not(.checkbox-label) {
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-medium);
    color: var(--text-muted);
  }

  .checkbox-label {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
    cursor: pointer;
    font-size: var(--font-size-sm);
    user-select: none;
  }

  input[type="text"],
  input[type="number"] {
    padding: var(--space-sm);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    background: var(--bg-secondary);
    color: var(--text);
    font-size: var(--font-size-sm);
  }

  input[type="text"]:focus,
  input[type="number"]:focus {
    outline: none;
    border-color: var(--primary);
  }

  input:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .error {
    margin: 0 var(--space-lg) var(--space-md);
    color: var(--danger);
    font-size: var(--font-size-sm);
    padding: var(--space-sm);
    background: var(--status-danger-bg);
    border-radius: var(--radius-sm);
  }

  .success {
    margin: 0 var(--space-lg) var(--space-md);
    color: var(--success);
    font-size: var(--font-size-sm);
    padding: var(--space-sm);
    background: var(--bg-tertiary);
    border: 1px solid var(--success);
    border-radius: var(--radius-sm);
  }

  /* Styles from ThemeSelector */
  .theme-section {
    margin-top: 0;
  }

  .theme-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
    gap: var(--space-md);
    margin-bottom: var(--space-lg);
  }

  .theme-card {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-sm);
    padding: var(--space-md);
    border: 2px solid var(--border);
    border-radius: var(--radius-lg);
    background: var(--bg-secondary);
    cursor: pointer;
    transition: all var(--transition-base);
    font-family: var(--font-sans);
    text-align: center;
  }

  .theme-card:hover {
    border-color: var(--border-dark);
    transform: translateY(-2px);
  }

  .theme-card.active {
    border-color: var(--primary);
    background: var(--bg-tertiary);
  }

  .theme-preview {
    width: 100%;
    height: 60px;
    border-radius: var(--radius-md);
    border: 1px solid var(--border);
  }

  .theme-colors {
    display: flex;
    gap: var(--space-xs);
  }

  .color-dot {
    width: 12px;
    height: 12px;
    border-radius: 50%;
    border: 1px solid rgba(0, 0, 0, 0.1);
  }

  .theme-name {
    font-size: var(--font-size-xs);
    font-weight: var(--font-weight-medium);
    text-transform: capitalize;
  }
</style>
