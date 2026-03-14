<script lang="ts">
  import { run } from "svelte/legacy";

  import ThemePreview from "$src/lib/components/Settings/ThemePreview.svelte";
  import { configurationStore } from "$src/lib/stores/configurationStore";
  import { notifications } from "$src/lib/stores/notificationStore";
  import { debounce } from "$src/lib/utils/debounce";
  import { createStableId, mapRecordToRowsWithStableIds } from "$src/lib/utils/stableKeyValueRows";
  import {
    DeleteHost,
    GetAllHosts,
    GetDefaultConfiguration,
    SelectFile,
    UpsertHost
  } from "$wails/go/main/App";
  import type { theme } from "$wails/go/models";
  import { configuration, host } from "$wails/go/models";
  import { onMount } from "svelte";

  // --- Nav ---
  type SettingsSection = "general" | "themes" | "hosts";
  let activeSection: SettingsSection = $state("themes");

  const NAV_ITEMS: { id: SettingsSection; label: string }[] = [
    { id: "general", label: "General" },
    { id: "themes", label: "Themes" },
    { id: "hosts", label: "Hosts" }
  ];

  // --- Store ---
  const { config, allThemes } = configurationStore;

  function findTheme(name: string) {
    return ($allThemes || []).find((t) => t.name === name) || null;
  }

  function formatThemeName(name: string): string {
    return name.replace(/-/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
  }

  // --- Theme UI state (letti dal config al mount) ---
  let themeMode: "sync" | "manual" = $state("manual");
  let activeThemeName = $state("");
  let dayTheme = $state("zinc-light");
  let nightTheme = $state("zinc-dark");

  // Inizializza da config (una volta sola, appena config+temi sono pronti)
  let initialized = $state(false);
  run(() => {
    if (!initialized && $config?.general?.activeTheme && ($allThemes || []).length > 0) {
      activeThemeName = $config.general.activeTheme;
      themeMode = ($config.general.themeMode as "sync" | "manual") || "manual";
      dayTheme = $config.general.dayTheme || "zinc-light";
      nightTheme = $config.general.nightTheme || "zinc-dark";
      initialized = true;
    }
  });

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
      current.general.themeMode = themeMode;
      current.general.dayTheme = dayTheme;
      current.general.nightTheme = nightTheme;
      await configurationStore.save(current);
      // Applica colori al DOM (changeTheme li applica già via derived store)
      await configurationStore.changeTheme(themeName);
    } catch {
      /* shown by store */
    }
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
    const listener = () => {
      if (themeMode === "sync") applySync();
    };
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

  let defaultConfig = $state(createEmptyConfig());
  let editableConfig = $state(createEmptyConfig());
  let saveStatus: "idle" | "saving" | "saved" = $state("idle");
  let lastPersistedSignature: string | null = null;

  function toSignature(cfg: configuration.Configuration): string {
    return JSON.stringify({
      checkForUpdates: cfg.general?.checkForUpdates ?? false,
      request: {
        timeoutSeconds: cfg.request?.timeoutSeconds ?? 0,
        defaultUserAgent: cfg.request?.defaultUserAgent ?? "",
        followRedirects: cfg.request?.followRedirects ?? true,
        maxRedirects: cfg.request?.maxRedirects ?? 0,
        validateSSL: cfg.request?.validateSSL ?? true,
        proxyUrl: cfg.request?.proxyUrl ?? ""
      }
    });
  }

  onMount(() => {
    (async () => {
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
          const sig = toSignature(copy);
          if (sig !== lastPersistedSignature) {
            editableConfig = copy;
            lastPersistedSignature = sig;
          }
        }
      });
      return () => unsub();
    })();
  });

  async function persistRequestSettings() {
    try {
      saveStatus = "saving";
      const timeoutSeconds = parseInt(String(editableConfig.request.timeoutSeconds), 10) || 0;
      const maxRedirects = parseInt(String(editableConfig.request.maxRedirects), 10) || 0;

      // Work on a detached copy to avoid mutating $config in-place (deep proxy).
      const current = new configuration.Configuration(JSON.parse(JSON.stringify($config)));
      if (!current.general) current.general = new configuration.GeneralSettings();
      if (!current.request) current.request = new configuration.RequestSettings();

      current.general.checkForUpdates = editableConfig.general.checkForUpdates;
      current.request = new configuration.RequestSettings({
        ...editableConfig.request,
        timeoutSeconds,
        maxRedirects
      });

      const sig = toSignature(current);
      if (sig === lastPersistedSignature) {
        saveStatus = "idle";
        return;
      }
      await configurationStore.save(current);
      lastPersistedSignature = sig;
      saveStatus = "saved";
      setTimeout(() => {
        saveStatus = "idle";
      }, 2000);
    } catch {
      saveStatus = "idle";
    }
  }

  const debouncedSave = debounce(persistRequestSettings, 800);
  run(() => {
    if (editableConfig.request || editableConfig.general) debouncedSave();
  });

  // --- Hosts ---
  type HostCookieRow = { id: string; key: string; value: string };

  let hostsList: host.Host[] = $state([]);
  let editingHost: host.Host | null = $state(null);
  let editingCookies: HostCookieRow[] = $state([]);

  async function fetchHosts() {
    try {
      hostsList = await GetAllHosts();
    } catch (err) {
      notifications.error("Failed to load hosts", String(err));
    }
  }

  function startAddHost() {
    const newHost = new host.Host();
    newHost.id = crypto.randomUUID();
    newHost.name = "";
    newHost.tlsConfig = new host.TLSConfig({
      enabled: false,
      insecureSkipVerify: false,
      publicCertificateFilePath: "",
      privateKeyFilePath: "",
      caCertificateFilePath: ""
    });
    newHost.cookies = {};
    editingHost = newHost;
    editingCookies = [];
  }

  function editExistingHost(h: host.Host) {
    editingHost = JSON.parse(JSON.stringify(h)) as host.Host;
    if (!editingHost.cookies) editingHost.cookies = {};
    const cookieRecord = Object.fromEntries(
      Object.entries(editingHost.cookies).map(([key, value]) => [key, String(value ?? "")])
    );
    editingCookies = mapRecordToRowsWithStableIds(cookieRecord);
  }

  function addCookieRow() {
    editingCookies = [...editingCookies, { id: createStableId(), key: "", value: "" }];
  }

  function removeCookieRow(id: string) {
    editingCookies = editingCookies.filter((cookie) => cookie.id !== id);
  }

  async function pickCertFile(
    field: "publicCertificateFilePath" | "privateKeyFilePath" | "caCertificateFilePath"
  ) {
    if (!editingHost) return;
    try {
      const path = await SelectFile("Select Certificate/Key File", "*.*", "All Files");
      if (path) {
        editingHost.tlsConfig[field] = path;
      }
    } catch (err) {
      notifications.error("File selection failed", String(err));
    }
  }

  async function handleSaveHost() {
    if (!editingHost || !editingHost.name) return;
    try {
      // Map cookie array back to object
      const cookieMap: Record<string, string> = {};
      editingCookies.forEach((c) => {
        if (c.key.trim()) cookieMap[c.key.trim()] = c.value;
      });
      editingHost.cookies = cookieMap;

      await UpsertHost(editingHost);
      await fetchHosts();
      editingHost = null;
      notifications.success("Host configuration saved");
    } catch (err) {
      notifications.error("Failed to save host", String(err));
    }
  }

  async function handleDeleteHost(name: string) {
    if (!confirm(`Are you sure you want to delete host config for "${name}"?`)) return;
    try {
      await DeleteHost(name);
      await fetchHosts();
      notifications.success("Host deleted");
    } catch (err) {
      notifications.error("Failed to delete host", String(err));
    }
  }

  run(() => {
    if (activeSection === "hosts" && hostsList.length === 0) {
      fetchHosts();
    }
  });
</script>

<div class="settings-modal">
  <!-- Sidebar -->
  <nav class="settings-nav">
    {#each NAV_ITEMS as item (item.id)}
      <button
        class="nav-item"
        class:active={activeSection === item.id}
        onclick={() => (activeSection = item.id)}
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
        <p class="section-desc">
          Personalize your experience with themes that match your style. Manually select a theme or
          sync with system settings and let the machine set your day and night themes.
        </p>

        <!-- Mode selector -->
        <div class="theme-mode-row">
          <span class="theme-mode-label">Theme selection</span>
          <div class="radio-group">
            <label class="radio-label">
              <input
                type="radio"
                bind:group={themeMode}
                value="sync"
                onchange={() => handleThemeModeChange("sync")}
              />
              Sync with system
            </label>
            <label class="radio-label">
              <input
                type="radio"
                bind:group={themeMode}
                value="manual"
                onchange={() => handleThemeModeChange("manual")}
              />
              Manual
            </label>
          </div>
        </div>

        {#if themeMode === "manual"}
          <!-- Manual: griglia di tutti i temi, click = applica -->
          <div class="theme-grid-manual">
            {#each $allThemes || [] as t (t.name)}
              <button
                class="theme-tile"
                class:active-tile={activeThemeName === t.name}
                onclick={() => handleManualThemeSelect(t.name)}
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
          <label for="proxy">Proxy URL</label>
          <input
            id="proxy"
            type="text"
            bind:value={editableConfig.request.proxyUrl}
            placeholder="http://user:pass@host:port (optional)"
          />
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
    {:else if activeSection === "hosts"}
      <div class="section-body">
        <h2 class="section-title">Hosts</h2>
        <p class="section-desc">
          Manage specific TLS and certificate configurations for your target hosts.
        </p>

        {#if !editingHost}
          <div class="hosts-header">
            <button class="btn btn-primary" onclick={startAddHost}>Add Host</button>
          </div>

          <div class="hosts-list">
            {#if hostsList.length === 0}
              <div class="empty-state">No specific host configuration found.</div>
            {:else}
              <div class="table-container">
                <table class="hosts-table">
                  <thead>
                    <tr>
                      <th>Hostname</th>
                      <th>TLS</th>
                      <th>Skip Verify</th>
                      <th>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {#each hostsList as h (h.id)}
                      <tr>
                        <td><strong>{h.name}</strong></td>
                        <td>
                          <span class="badge" class:badge-success={h.tlsConfig.enabled}>
                            {h.tlsConfig.enabled ? "Enabled" : "Disabled"}
                          </span>
                        </td>
                        <td>
                          <span class="badge" class:badge-warning={h.tlsConfig.insecureSkipVerify}>
                            {h.tlsConfig.insecureSkipVerify ? "Yes" : "No"}
                          </span>
                        </td>
                        <td class="actions-cell">
                          <button class="btn-icon" onclick={() => editExistingHost(h)}>Edit</button>
                          <button
                            class="btn-icon btn-danger"
                            onclick={() => handleDeleteHost(h.name)}>Delete</button
                          >
                        </td>
                      </tr>
                    {/each}
                  </tbody>
                </table>
              </div>
            {/if}
          </div>
        {:else}
          <!-- Edit Form -->
          <div class="host-form">
            <h3 class="subsection-title">{editingHost.name ? "Edit Host" : "New Host"}</h3>

            <div class="form-group">
              <label for="h-name">Hostname (e.g. api.company.com or localhost:8443)</label>
              <input id="h-name" type="text" bind:value={editingHost.name} placeholder="Hostname" />
            </div>

            <div class="checkbox-group">
              <label class="checkbox-label">
                <input type="checkbox" bind:checked={editingHost.tlsConfig.enabled} />
                Enable Custom TLS Configuration
              </label>
            </div>

            {#if editingHost.tlsConfig.enabled}
              <div class="tls-details">
                <div class="checkbox-group">
                  <label class="checkbox-label">
                    <input
                      type="checkbox"
                      bind:checked={editingHost.tlsConfig.insecureSkipVerify}
                    />
                    Insecure Skip Verify (not recommended)
                  </label>
                </div>

                <div class="form-group file-picker">
                  <label>Public Certificate (.crt, .pem)</label>
                  <div class="input-with-action">
                    <input
                      type="text"
                      bind:value={editingHost.tlsConfig.publicCertificateFilePath}
                      readonly
                    />
                    <button
                      class="btn btn-secondary"
                      onclick={() => pickCertFile("publicCertificateFilePath")}>Browse</button
                    >
                  </div>
                </div>

                <div class="form-group file-picker">
                  <label>Private Key (.key)</label>
                  <div class="input-with-action">
                    <input
                      type="text"
                      bind:value={editingHost.tlsConfig.privateKeyFilePath}
                      readonly
                    />
                    <button
                      class="btn btn-secondary"
                      onclick={() => pickCertFile("privateKeyFilePath")}>Browse</button
                    >
                  </div>
                </div>

                <div class="form-group file-picker">
                  <label>CA Certificate (Root CA)</label>
                  <div class="input-with-action">
                    <input
                      type="text"
                      bind:value={editingHost.tlsConfig.caCertificateFilePath}
                      readonly
                    />
                    <button
                      class="btn btn-secondary"
                      onclick={() => pickCertFile("caCertificateFilePath")}>Browse</button
                    >
                  </div>
                </div>
              </div>
            {/if}

            <div class="cookies-section">
              <h3 class="subsection-title">Cookies</h3>
              <p class="section-desc">
                These cookies will be automatically added to all requests sent to this host.
              </p>

              <div class="cookies-list">
                {#each editingCookies as cookie (cookie.id)}
                  <div class="cookie-row">
                    <input type="text" bind:value={cookie.key} placeholder="Cookie Name" />
                    <input type="text" bind:value={cookie.value} placeholder="Value" />
                    <button class="btn-icon btn-danger" onclick={() => removeCookieRow(cookie.id)}
                      >Remove</button
                    >
                  </div>
                {/each}
                <button class="btn btn-secondary btn-sm" onclick={addCookieRow}>+ Add Cookie</button
                >
              </div>
            </div>

            <div class="form-actions">
              <button class="btn btn-secondary" onclick={() => (editingHost = null)}>Cancel</button>
              <button class="btn btn-primary" onclick={handleSaveHost} disabled={!editingHost.name}
                >Save Host</button
              >
            </div>
          </div>
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
    transition:
      background 0.15s,
      color 0.15s;
  }
  .nav-item:hover {
    background: var(--bg-tertiary);
    color: var(--text);
  }
  .nav-item.active {
    background: var(--bg-tertiary);
    color: var(--text);
    font-weight: var(--font-weight-semibold);
  }

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
    transition:
      border-color 0.15s,
      box-shadow 0.15s;
    text-align: left;
    font-family: inherit;
  }
  .theme-tile:hover {
    border-color: var(--border-dark);
  }
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
  input:focus {
    outline: none;
    border-color: var(--primary);
  }
  input:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .save-status {
    font-size: var(--font-size-xs, 0.72rem);
    color: var(--text-muted);
    font-style: italic;
    margin: 0;
  }
  .save-status.saved {
    color: var(--success);
  }

  /* ---- Hosts section ---- */
  .hosts-header {
    display: flex;
    justify-content: flex-end;
    margin-bottom: var(--space-md);
  }

  .table-container {
    overflow-x: auto;
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
  }

  .hosts-table {
    width: 100%;
    border-collapse: collapse;
    font-size: var(--font-size-sm);
  }

  .hosts-table th,
  .hosts-table td {
    padding: var(--space-sm) var(--space-md);
    text-align: left;
    border-bottom: 1px solid var(--border);
  }

  .hosts-table th {
    background: var(--bg-secondary);
    font-weight: var(--font-weight-semibold);
    color: var(--text-muted);
  }

  .hosts-table tr:last-child td {
    border-bottom: none;
  }

  .badge {
    padding: 2px 6px;
    border-radius: var(--radius-sm);
    font-size: 0.7rem;
    font-weight: var(--font-weight-semibold);
    text-transform: uppercase;
    background: var(--bg-tertiary);
    color: var(--text-muted);
  }
  .badge-success {
    background: color-mix(in srgb, var(--success) 20%, transparent);
    color: var(--success);
  }
  .badge-warning {
    background: color-mix(in srgb, var(--warning) 20%, transparent);
    color: var(--warning);
  }

  .actions-cell {
    display: flex;
    gap: var(--space-sm);
  }

  /* ---- Host form ---- */
  .host-form {
    display: flex;
    flex-direction: column;
    gap: var(--space-lg);
    background: var(--bg-secondary);
    padding: var(--space-lg);
    border-radius: var(--radius-lg);
    border: 1px solid var(--border);
  }

  .tls-details {
    padding-left: var(--space-lg);
    border-left: 2px solid var(--primary);
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
    margin-top: var(--space-xs);
  }

  .input-with-action {
    display: flex;
    gap: var(--space-sm);
  }

  .input-with-action input {
    flex: 1;
    background: var(--bg-tertiary);
  }

  .form-actions {
    display: flex;
    justify-content: flex-end;
    gap: var(--space-md);
    margin-top: var(--space-md);
    padding-top: var(--space-md);
    border-top: 1px solid var(--border);
  }

  .checkbox-group {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
  }

  .cookies-section {
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
    margin-top: var(--space-md);
  }

  .cookie-row {
    display: grid;
    grid-template-columns: 1fr 1.5fr auto;
    gap: var(--space-sm);
    margin-bottom: var(--space-xs);
    align-items: center;
  }

  .cookie-row input {
    padding: var(--space-xs) var(--space-sm);
    font-size: var(--font-size-xs);
  }

  .btn-sm {
    padding: var(--space-xs) var(--space-sm);
    font-size: var(--font-size-xs);
    align-self: flex-start;
  }

  /* ---- Buttons & Utils ---- */
  .btn {
    padding: var(--space-sm) var(--space-md);
    border-radius: var(--radius-md);
    border: 1px solid var(--border);
    cursor: pointer;
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-medium);
    transition: all 0.15s;
    background: var(--bg-secondary);
    color: var(--text);
  }
  .btn:hover:not(:disabled) {
    background: var(--bg-tertiary);
  }
  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-primary {
    background: var(--primary);
    color: white;
    border-color: var(--primary);
  }
  .btn-primary:hover:not(:disabled) {
    filter: brightness(1.1);
  }

  .btn-secondary {
    background: var(--bg-tertiary);
    border-color: var(--border);
  }

  .btn-icon {
    background: none;
    border: none;
    cursor: pointer;
    font-size: var(--font-size-xs);
    color: var(--primary);
    padding: 0;
  }
  .btn-icon:hover {
    text-decoration: underline;
  }
  .btn-icon.btn-danger {
    color: var(--danger);
  }

  .empty-state {
    padding: var(--space-xl);
    text-align: center;
    color: var(--text-muted);
    font-style: italic;
    background: var(--bg-secondary);
    border-radius: var(--radius-md);
    border: 1px dashed var(--border);
  }
</style>
