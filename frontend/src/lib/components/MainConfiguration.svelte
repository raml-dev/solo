<script lang="ts">
  import { onMount } from "svelte";
  import { configurationStore } from "../stores/configurationStore";
  import { notifications } from "../stores/notificationStore";
  import type { theme } from "../../../wailsjs/go/models";
  import { configuration } from "../../../wailsjs/go/models";
  import { GetDefaultConfiguration } from "../../../wailsjs/go/main/App";
  import { debounce } from "../utils/debounce";
  import Dropdown from "./base/Dropdown.svelte";
  import ThemePreview from "./Settings/ThemePreview.svelte";

  export let toggleFn: () => void;

  // --- Nav ---
  type SettingsSection = "general" | "themes";
  let activeSection: SettingsSection = "themes";

  const NAV_ITEMS: { id: SettingsSection; label: string }[] = [
    { id: "general", label: "General" },
    { id: "themes",  label: "Themes"  },
  ];

  // --- Store ---
  const { config, allThemes } = configurationStore;

  $: themeOptions = ($allThemes || []).map((t) => ({ value: t.name, label: formatThemeName(t.name) }));

  function findTheme(name: string) {
    return ($allThemes || []).find((t) => t.name === name) || null;
  }

  function formatThemeName(name: string): string {
    return name.replace(/-/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
  }

  // --- Theme UI state (letti dal config al mount) ---
  let themeMode: "sync" | "manual" = "manual";
  let activeThemeName = "";
  let dayTheme   = "zinc-light";
  let nightTheme = "zinc-dark";

  // Inizializza da config (una volta sola, appena config+temi sono pronti)
  let initialized = false;
  $: if (!initialized && $config?.general?.activeTheme && ($allThemes || []).length > 0) {
    activeThemeName = $config.general.activeTheme;
    themeMode       = ($config.general.themeMode as "sync" | "manual") || "manual";
    dayTheme        = $config.general.dayTheme   || "zinc-light";
    nightTheme      = $config.general.nightTheme || "zinc-dark";
    initialized = true;
  }

  // --- System preference ---
  const prefersDark = window.matchMedia("(prefers-color-scheme: dark)");
  function getSystemTheme(): "light" | "dark" {
    return prefersDark.matches ? "dark" : "light";
  }

  // --- Applica tema + persiste tutto nel config ---
  async function applyAndSaveTheme(themeName: string) {
    if (!findTheme(themeName)) return;
    activeThemeName = themeName;
    try {
      // Salva activeTheme, themeMode, dayTheme, nightTheme in un'unica write
      const current = $config;
      current.general.activeTheme = themeName;
      current.general.themeMode   = themeMode;
      current.general.dayTheme    = dayTheme;
      current.general.nightTheme  = nightTheme;
      await configurationStore.save(current);
      // Applica colori al DOM (changeTheme li applica già via derived store)
      await configurationStore.changeTheme(themeName);
    } catch { /* shown by store */ }
  }

  // In sync mode: applica il tema giusto per il sistema corrente
  async function applySync() {
    const target = getSystemTheme() === "dark" ? nightTheme : dayTheme;
    await applyAndSaveTheme(target);
  }

  async function handleThemeModeChange(mode: "sync" | "manual") {
    themeMode = mode;
    if (mode === "sync") {
      await applySync();
    } else {
      // Persiste solo il cambio di modalità, tema rimane invariato
      const current = $config;
      current.general.themeMode = "manual";
      await configurationStore.save(current);
    }
  }

  async function handleManualThemeSelect(name: string) {
    await applyAndSaveTheme(name);
  }

  // Listener per cambi OS in sync mode
  onMount(() => {
    const listener = () => { if (themeMode === "sync") applySync(); };
    prefersDark.addEventListener("change", listener);
    return () => prefersDark.removeEventListener("change", listener);
  });

  // --- General / Request settings ---
  function createEmptyConfig() {
    const cfg = new configuration.Configuration();
    cfg.general = new configuration.GeneralSettings();
    cfg.request = new configuration.RequestSettings();
    cfg.customThemes = [] as theme.Theme[];
    return cfg;
  }

  let defaultConfig = createEmptyConfig();
  let editableConfig = createEmptyConfig();
  let saveStatus: "idle" | "saving" | "saved" = "idle";
  let lastPersistedSignature: string | null = null;

  function toSignature(cfg: configuration.Configuration): string {
    return JSON.stringify({
      checkForUpdates: cfg.general?.checkForUpdates ?? false,
      request: {
        timeoutSeconds:   cfg.request?.timeoutSeconds   ?? 0,
        defaultUserAgent: cfg.request?.defaultUserAgent ?? "",
        followRedirects:  cfg.request?.followRedirects  ?? true,
        maxRedirects:     cfg.request?.maxRedirects      ?? 0,
        validateSSL:      cfg.request?.validateSSL       ?? true,
        proxyUrl:         cfg.request?.proxyUrl          ?? "",
      }
    });
  }

  onMount(async () => {
    try {
      const loaded = await GetDefaultConfiguration();
      defaultConfig = new configuration.Configuration(loaded);
      if (!defaultConfig.general) defaultConfig.general = new configuration.GeneralSettings();
      if (!defaultConfig.request) defaultConfig.request = new configuration.RequestSettings();
    } catch (err) {
      notifications.error("Failed to load default configuration", String(err));
    }

    const unsub = config.subscribe((value) => {
      if (value?.request) {
        const copy = new configuration.Configuration(JSON.parse(JSON.stringify(value)));
        if (!copy.general) copy.general = new configuration.GeneralSettings();
        if (!copy.request) copy.request = new configuration.RequestSettings();
        editableConfig = copy;
        lastPersistedSignature = toSignature(copy);
      }
    });
    return () => unsub();
  });

  async function persistRequestSettings() {
    try {
      saveStatus = "saving";
      editableConfig.request.timeoutSeconds = parseInt(String(editableConfig.request.timeoutSeconds), 10) || 0;
      editableConfig.request.maxRedirects   = parseInt(String(editableConfig.request.maxRedirects),   10) || 0;
      const current = $config;
      current.general.checkForUpdates = editableConfig.general.checkForUpdates;
      current.request = editableConfig.request;
      const sig = toSignature(current);
      if (sig === lastPersistedSignature) { saveStatus = "idle"; return; }
      await configurationStore.save(current);
      lastPersistedSignature = sig;
      saveStatus = "saved";
      setTimeout(() => { saveStatus = "idle"; }, 2000);
    } catch { saveStatus = "idle"; }
  }

  const debouncedSave = debounce(persistRequestSettings, 800);
  $: if (editableConfig.request || editableConfig.general) debouncedSave();
</script>

<div class="settings-modal">
  <!-- Sidebar -->
  <nav class="settings-nav">
    {#each NAV_ITEMS as item}
      <button
        class="nav-item"
        class:active={activeSection === item.id}
        on:click={() => (activeSection = item.id)}
      >
        {item.label}
      </button>
    {/each}
  </nav>

  <!-- Content -->
  <div class="settings-content">

    {#if activeSection === "themes"}
      <div class="section-body">
        <h2 class="section-title">Themes</h2>
        <p class="section-desc">Personalize your experience with themes that match your style. Manually select a theme or sync with system settings and let the machine set your day and night themes.</p>

        <!-- Mode selector -->
        <div class="theme-mode-row">
          <span class="theme-mode-label">Theme selection</span>
          <div class="radio-group">
            <label class="radio-label">
              <input type="radio" bind:group={themeMode} value="sync" on:change={() => handleThemeModeChange("sync")} />
              Sync with system
            </label>
            <label class="radio-label">
              <input type="radio" bind:group={themeMode} value="manual" on:change={() => handleThemeModeChange("manual")} />
              Manual
            </label>
          </div>
        </div>

        {#if themeMode === "manual"}
          <!-- Manual: griglia di tutti i temi, click = applica -->
          <div class="theme-grid-manual">
            {#each ($allThemes || []) as t (t.name)}
              <button
                class="theme-tile"
                class:active-tile={activeThemeName === t.name}
                on:click={() => handleManualThemeSelect(t.name)}
              >
                <div class="theme-tile-preview">
                  <ThemePreview themeColors={t.colors} />
                </div>
                <span class="theme-tile-name">{formatThemeName(t.name)}</span>
                {#if activeThemeName === t.name}
                  <span class="active-badge">ACTIVE</span>
                {/if}
              </button>
            {/each}
          </div>

        {:else}
          <!-- Sync: mostra solo il tema attivo corrente -->
          <div class="sync-active-theme">
            <div class="sync-active-preview">
              <ThemePreview themeColors={findTheme(activeThemeName)?.colors || {}} />
            </div>
            <div class="sync-active-info">
              <span class="sync-active-name">{formatThemeName(activeThemeName) || "—"}</span>
              <span class="sync-active-sub">Currently active — follows your system appearance</span>
            </div>
          </div>
        {/if}
      </div>

    {:else if activeSection === "general"}
      <div class="section-body">
        <h2 class="section-title">General</h2>
        <p class="section-desc">Configure general application behavior and request defaults.</p>

        <div class="form-group">
          <label class="checkbox-label">
            <input type="checkbox" bind:checked={editableConfig.general.checkForUpdates} />
            Check for updates on startup
          </label>
        </div>

        <h3 class="subsection-title">Request Defaults</h3>

        <div class="form-row">
          <div class="form-group">
            <label for="timeout">Timeout (seconds)</label>
            <input id="timeout" type="number" bind:value={editableConfig.request.timeoutSeconds}
              min="0" step="1" placeholder={`Default: ${defaultConfig.request.timeoutSeconds}`} />
          </div>
          <div class="form-group">
            <label for="max-redirects">Max Redirects</label>
            <input id="max-redirects" type="number" bind:value={editableConfig.request.maxRedirects}
              min="0" step="1" placeholder={`Default: ${defaultConfig.request.maxRedirects}`}
              disabled={!editableConfig.request.followRedirects} />
          </div>
        </div>

        <div class="form-group">
          <label for="user-agent">Default User Agent</label>
          <input id="user-agent" type="text" bind:value={editableConfig.request.defaultUserAgent}
            placeholder={`Default: ${defaultConfig.request.defaultUserAgent}`} />
        </div>

        <div class="form-group">
          <label for="proxy">Proxy URL</label>
          <input id="proxy" type="text" bind:value={editableConfig.request.proxyUrl}
            placeholder="http://user:pass@host:port (optional)" />
        </div>

        <div class="checkboxes">
          <label class="checkbox-label">
            <input type="checkbox" bind:checked={editableConfig.request.followRedirects} />
            Follow Redirects
          </label>
          <label class="checkbox-label">
            <input type="checkbox" bind:checked={editableConfig.request.validateSSL} />
            Validate SSL Certificates
          </label>
        </div>

        {#if saveStatus === "saving"}
          <p class="save-status">Saving…</p>
        {:else if saveStatus === "saved"}
          <p class="save-status saved">Saved ✓</p>
        {/if}
      </div>
    {/if}
  </div>
</div>

<style>
  .settings-modal {
    display: flex;
    height: 100%;
    overflow: hidden;
    border-radius: var(--radius-lg);
  }

  /* ---- Sidebar ---- */
  .settings-nav {
    width: 200px;
    flex-shrink: 0;
    background: var(--bg-secondary);
    border-right: 1px solid var(--border);
    padding: var(--space-lg) var(--space-sm);
    display: flex;
    flex-direction: column;
    gap: 2px;
    border-radius: var(--radius-lg) 0 0 var(--radius-lg);
  }

  .nav-item {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
    padding: var(--space-sm) var(--space-md);
    border-radius: var(--radius-md);
    background: none;
    border: none;
    cursor: pointer;
    color: var(--text-muted);
    font-size: var(--font-size-sm);
    text-align: left;
    transition: background 0.15s, color 0.15s;
  }
  .nav-item:hover { background: var(--bg-tertiary); color: var(--text); }
  .nav-item.active { background: var(--bg-tertiary); color: var(--text); font-weight: var(--font-weight-semibold); }

  /* ---- Content ---- */
  .settings-content {
    flex: 1;
    overflow-y: auto;
    position: relative;
    background: var(--bg-primary);
    border-radius: 0 var(--radius-lg) var(--radius-lg) 0;
  }

  .section-body {
    padding: var(--space-xl) var(--space-xl) var(--space-xl) var(--space-xl);
    display: flex;
    flex-direction: column;
    gap: var(--space-lg);
  }

  .section-title {
    margin: 0;
    font-size: 1.3rem;
    font-weight: var(--font-weight-semibold);
    color: var(--text);
  }

  .section-desc {
    margin: 0;
    font-size: var(--font-size-sm);
    color: var(--text-muted);
    line-height: 1.5;
  }

  /* ---- Theme mode row ---- */
  .theme-mode-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-bottom: var(--space-md);
    border-bottom: 1px solid var(--border);
  }

  .theme-mode-label {
    font-size: var(--font-size-sm);
    color: var(--text-muted);
  }

  .radio-group {
    display: flex;
    gap: var(--space-lg);
  }

  .radio-label {
    display: flex;
    align-items: center;
    gap: var(--space-xs);
    font-size: var(--font-size-sm);
    cursor: pointer;
    color: var(--text);
  }

  /* ---- Sync active theme display ---- */
  .sync-active-theme {
    display: flex;
    align-items: center;
    gap: var(--space-lg);
    padding: var(--space-md);
    border: 1px solid var(--primary);
    border-radius: var(--radius-lg);
    background: var(--bg-secondary);
  }

  .sync-active-preview {
    width: 180px;
    flex-shrink: 0;
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    overflow: hidden;
    aspect-ratio: 240 / 130;
  }

  .sync-active-info {
    display: flex;
    flex-direction: column;
    gap: var(--space-xs);
  }

  .sync-active-name {
    font-size: var(--font-size-md);
    font-weight: var(--font-weight-semibold);
    color: var(--text);
    text-transform: capitalize;
  }

  .sync-active-sub {
    font-size: var(--font-size-sm);
    color: var(--text-muted);
  }

  /* ---- Theme grid (manual mode) ---- */
  .theme-grid-manual {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
    gap: var(--space-md);
  }

  .theme-tile {
    display: flex;
    flex-direction: column;
    align-items: stretch;
    gap: var(--space-xs);
    padding: var(--space-sm);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    background: var(--bg-secondary);
    cursor: pointer;
    transition: border-color 0.15s, box-shadow 0.15s;
    text-align: left;
    font-family: inherit;
  }
  .theme-tile:hover { border-color: var(--border-dark); }
  .theme-tile.active-tile {
    border-color: var(--primary);
    box-shadow: 0 0 0 2px color-mix(in srgb, var(--primary) 20%, transparent);
  }

  .theme-tile-preview {
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    overflow: hidden;
    aspect-ratio: 240 / 130;
    background: var(--bg-tertiary);
  }

  .theme-tile-name {
    font-size: var(--font-size-xs, 0.72rem);
    color: var(--text);
    text-transform: capitalize;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .active-badge {
    align-self: flex-start;
    font-size: 0.6rem;
    font-weight: var(--font-weight-semibold);
    letter-spacing: 0.06em;
    background: color-mix(in srgb, var(--info) 20%, transparent);
    color: var(--info);
    border: 1px solid color-mix(in srgb, var(--info) 40%, transparent);
    border-radius: var(--radius-sm);
    padding: 1px 6px;
    /* in card header: push to right */
    margin-left: auto;
  }

  .theme-preview {
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    overflow: hidden;
    aspect-ratio: 240 / 130;
    background: var(--bg-tertiary);
  }

  /* ---- General form ---- */
  .subsection-title {
    margin: 0;
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-semibold);
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    border-bottom: 1px solid var(--border);
    padding-bottom: var(--space-xs);
  }

  .form-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-md);
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: var(--space-xs);
  }

  label:not(.checkbox-label):not(.radio-label) {
    font-size: var(--font-size-sm);
    color: var(--text-muted);
  }

  .checkbox-label {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
    cursor: pointer;
    font-size: var(--font-size-sm);
    color: var(--text);
  }

  .checkboxes {
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
  }

  input[type="text"],
  input[type="number"] {
    padding: var(--space-sm) var(--space-md);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    background: var(--bg-secondary);
    color: var(--text);
    font-size: var(--font-size-sm);
  }
  input:focus { outline: none; border-color: var(--primary); }
  input:disabled { opacity: 0.5; cursor: not-allowed; }

  .save-status {
    font-size: var(--font-size-xs, 0.72rem);
    color: var(--text-muted);
    font-style: italic;
    margin: 0;
  }
  .save-status.saved { color: var(--success); }
</style>
