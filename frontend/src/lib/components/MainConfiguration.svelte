<script lang="ts">
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

  function findTheme(id: string) {
    return ($allThemes || []).find((t) => t.id === id) || null;
  }

  function formatThemeName(label: string): string {
    return label || "Untitled Theme";
  }

  // --- Theme UI state ---
  let activeThemeId = $state("");

  let initialized = $state(false);
  $effect(() => {
    if (!initialized && $config?.general?.activeTheme && ($allThemes || []).length > 0) {
      activeThemeId = $config.general.activeTheme;
      initialized = true;
    }
  });

  async function applyAndSaveTheme(themeId: string) {
    if (!findTheme(themeId)) return;
    activeThemeId = themeId;
    try {
      const current = new configuration.Configuration(JSON.parse(JSON.stringify($config)));
      if (!current.general) current.general = new configuration.GeneralSettings();
      current.general.activeTheme = themeId;
      await configurationStore.save(current);
      await configurationStore.changeTheme(themeId);
    } catch {
      /* shown by store */
    }
  }

  async function handleThemeSelect(themeId: string) {
    await applyAndSaveTheme(themeId);
  }

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
    let disposed = false;

    const unsub = config.subscribe((value) => {
      if (disposed) return;
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

    void (async () => {
      try {
        const loaded = await GetDefaultConfiguration();
        if (disposed) return;
        defaultConfig = new configuration.Configuration(loaded);
        if (!defaultConfig.general) defaultConfig.general = new configuration.GeneralSettings();
        if (!defaultConfig.request) defaultConfig.request = new configuration.RequestSettings();
      } catch (err) {
        if (disposed) return;
        notifications.error("Failed to load default configuration", String(err));
      }
    })();

    return () => {
      disposed = true;
      unsub();
    };
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
  $effect(() => {
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

  $effect(() => {
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

        <div class="theme-grid-manual">
          {#each $allThemes || [] as t (t.id)}
            <button
              class="theme-tile"
              class:active-tile={activeThemeId === t.id}
              onclick={() => handleThemeSelect(t.id)}
            >
              <div class="theme-tile-preview">
                <ThemePreview seeds={t.config?.seeds} />
              </div>
              <span class="theme-tile-name">{formatThemeName(t.label)}</span>
              {#if activeThemeId === t.id}
                <span class="active-badge">ACTIVE</span>
              {/if}
            </button>
          {/each}
        </div>
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
