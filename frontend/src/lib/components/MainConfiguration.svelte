<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import ToastContainer from "$src/lib/components/base/ToastContainer.svelte";
  import FeedbackEmptyState from "$src/lib/components/common/FeedbackEmptyState.svelte";
  import About from "$src/lib/components/Settings/About.svelte";
  import AppearancePane from "$src/lib/components/Settings/AppearancePane.svelte";
  import {
    configurationStore,
    configurationStoreState,
    saveConfig
  } from "$src/lib/stores/configurationStore.svelte";
  import { fontListsStore, fontListsStoreState } from "$src/lib/stores/fontListsStore.svelte";
  import { modalStack, topModalId } from "$src/lib/stores/modalStackStore.svelte";
  import { notifications } from "$src/lib/stores/notificationStore";
  import { updateStore } from "$src/lib/stores/updateStore.svelte";
  import { resetZoom, setZoomLevel, zoomState } from "$src/lib/stores/zoomStore.svelte";
  import type { ThemeMode } from "$src/lib/theme/themeModel";
  import { createStableId, mapRecordToRowsWithStableIds } from "$src/lib/utils/stableKeyValueRows";
  import {
    DeleteHost,
    ExportLogsZip,
    GetAllHosts,
    GetDefaultConfiguration,
    SelectFile,
    SetDebugMode,
    UpsertHost
  } from "$wails/go/main/App";
  import type { theme } from "$wails/go/models";
  import { configuration, host } from "$wails/go/models";
  import TrashBinOutline from "flowbite-svelte-icons/TrashBinOutline.svelte";
  import Badge from "flowbite-svelte/Badge.svelte";
  import Button from "flowbite-svelte/Button.svelte";
  import Helper from "flowbite-svelte/Helper.svelte";
  import Input from "flowbite-svelte/Input.svelte";
  import Label from "flowbite-svelte/Label.svelte";
  import Modal from "flowbite-svelte/Modal.svelte";
  import Table from "flowbite-svelte/Table.svelte";
  import TableBody from "flowbite-svelte/TableBody.svelte";
  import TableBodyCell from "flowbite-svelte/TableBodyCell.svelte";
  import TableBodyRow from "flowbite-svelte/TableBodyRow.svelte";
  import TableHead from "flowbite-svelte/TableHead.svelte";
  import TableHeadCell from "flowbite-svelte/TableHeadCell.svelte";
  import Toggle from "flowbite-svelte/Toggle.svelte";
  import { onMount } from "svelte";

  // 1) custom types
  type SettingsSection = "general" | "appearance" | "troubleshooting" | "hosts" | "about";
  type HostCookieRow = { id: string; key: string; value: string };

  // 2) props
  // no props in this component

  // constants
  const NAV_ITEMS: { id: SettingsSection; label: string }[] = [
    { id: "general", label: "General" },
    { id: "appearance", label: "Appearance" },
    { id: "hosts", label: "Hosts" },
    { id: "troubleshooting", label: "Troubleshooting" },
    { id: "about", label: "About" }
  ];
  const deleteHostModal = modalStack.createModal("settings-delete-host");

  // 3) $state vars
  let activeSection: SettingsSection = $state("general");

  let defaultConfig = $state(createEmptyConfig());

  let isExportingLogs = $state(false);

  let hostsList: host.Host[] = $state([]);
  let editingHost: host.Host | null = $state(null);
  let editingHostName = $state("");
  let editingCookies: HostCookieRow[] = $state([]);
  let customTlsEnabled = $state(false);
  let hostToDelete = $state("");

  // 4) $derived vars
  const themesState = $derived(configurationStoreState.allThemes);
  const configSaveStatus = $derived(configurationStoreState.saveStatus);
  const fontLists = $derived(fontListsStoreState.lists);
  const fontsLoading = $derived(fontListsStoreState.loading);
  const fontsReady = $derived(fontListsStoreState.ready);
  const fontsError = $derived(fontListsStoreState.error);
  const zoomLevel = $derived(zoomState.level);

  // 5) helper functions
  function findTheme(id: string) {
    return (themesState || []).find((t) => t.id === id) || null;
  }

  function ensureTypographyDefaults(settings?: configuration.GeneralSettings | null) {
    if (!settings) return;
    settings.defaultFontFamily ??= "";
    settings.monoFontFamily ??= "";
  }

  function createEmptyConfig() {
    const cfg = new configuration.Configuration();
    cfg.general = new configuration.GeneralSettings({ debugMode: true });
    ensureTypographyDefaults(cfg.general);
    cfg.request = new configuration.RequestSettings();
    cfg.customThemes = [] as theme.Theme[];
    return cfg;
  }

  async function handleThemeSelect(themeId: string) {
    if (!findTheme(themeId)) return;
    try {
      await configurationStore.changeTheme(themeId);
    } catch {
      /* shown by store */
    }
  }

  function handleTextSettingChange() {
    saveConfig();
  }

  function handleToggleSettingChange() {
    saveConfig();
  }

  function handleUpdateToggleSettingChange() {
    saveConfig();
    void updateStore.syncWithConfiguration();
  }

  function handleThemeModeChange(mode: ThemeMode) {
    configurationStore.applyThemeMode(mode);
    saveConfig();
  }

  async function handleZoomLevelChange(level: number) {
    await setZoomLevel(level);
  }

  async function handleSansFontChange(fontFamily: string) {
    try {
      await configurationStore.changeDefaultFontFamily(fontFamily);
    } catch {
      /* shown by store */
    }
  }

  async function handleMonoFontChange(fontFamily: string) {
    try {
      await configurationStore.changeMonoFontFamily(fontFamily);
    } catch {
      /* shown by store */
    }
  }

  async function handleRefreshFonts() {
    try {
      await fontListsStore.refresh();
    } catch {
      /* shown by store */
    }
  }

  async function handleSansFontReset() {
    await handleSansFontChange("");
  }

  async function handleMonoFontReset() {
    await handleMonoFontChange("");
  }

  async function handleZoomLevelReset() {
    await resetZoom();
  }

  async function handleLogsExport() {
    if (isExportingLogs) return;

    try {
      isExportingLogs = true;
      const saved = await ExportLogsZip();
      if (saved) {
        notifications.success("Logs archive exported");
      }
    } catch (err) {
      notifications.error("Failed to export logs", String(err));
    } finally {
      isExportingLogs = false;
    }
  }

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
    editingHostName = "";
    editingCookies = [];
    customTlsEnabled = newHost.tlsConfig.enabled;
  }

  function editExistingHost(h: host.Host) {
    editingHost = JSON.parse(JSON.stringify(h)) as host.Host;
    if (!editingHost.cookies) editingHost.cookies = {};
    if (!editingHost.tlsConfig) {
      editingHost.tlsConfig = new host.TLSConfig({
        enabled: false,
        insecureSkipVerify: false,
        publicCertificateFilePath: "",
        privateKeyFilePath: "",
        caCertificateFilePath: ""
      });
    }
    customTlsEnabled = Boolean(editingHost.tlsConfig.enabled);
    editingHostName = editingHost.name ?? "";
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
    if (!editingHost || !editingHostName.trim()) return;
    try {
      editingHost.name = editingHostName.trim();
      // Map cookie array back to object
      const cookieMap: Record<string, string> = {};
      editingCookies.forEach((c) => {
        if (c.key.trim()) cookieMap[c.key.trim()] = c.value;
      });
      editingHost.cookies = cookieMap;

      if (!editingHost.tlsConfig) {
        editingHost.tlsConfig = new host.TLSConfig({
          enabled: false,
          insecureSkipVerify: false,
          publicCertificateFilePath: "",
          privateKeyFilePath: "",
          caCertificateFilePath: ""
        });
      }
      editingHost.tlsConfig.enabled = customTlsEnabled;

      await UpsertHost(editingHost);
      await fetchHosts();
      editingHost = null;
      editingHostName = "";
      notifications.success("Host configuration saved");
    } catch (err) {
      notifications.error("Failed to save host", String(err));
    }
  }

  function handleDeleteHost(name: string) {
    hostToDelete = name;
    deleteHostModal.open = true;
  }

  async function confirmDeleteHost() {
    deleteHostModal.open = false;
    try {
      await DeleteHost(hostToDelete);
      await fetchHosts();
      notifications.success("Host deleted");
    } catch (err) {
      notifications.error("Failed to delete host", String(err));
    }
  }

  function handleSectionChange(section: SettingsSection) {
    activeSection = section;
    if (section === "hosts" && hostsList.length === 0) {
      void fetchHosts();
    }
  }

  // 6) onMount and onDestroy
  onMount(() => {
    let disposed = false;

    configurationStore.applyThemeMode(
      (configurationStoreState.config.general.themeMode as ThemeMode) || "system"
    );
    void (async () => {
      try {
        await SetDebugMode(true);
      } catch (err) {
        console.warn("Failed to enable debug mode at runtime", err);
      }

      try {
        const loaded = await GetDefaultConfiguration();
        if (disposed) return;
        defaultConfig = new configuration.Configuration(loaded);
        if (!defaultConfig.general) defaultConfig.general = new configuration.GeneralSettings();
        ensureTypographyDefaults(defaultConfig.general);
        if (!defaultConfig.request) defaultConfig.request = new configuration.RequestSettings();
      } catch (err) {
        if (disposed) return;
        notifications.error("Failed to load default configuration", String(err));
      }
    })();

    return () => {
      disposed = true;
      modalStack.destroyModal(deleteHostModal.id);
    };
  });
</script>

<div class="flex h-full gap-6">
  <!-- Sidebar nav -->
  <nav class="flex shrink-0 flex-col gap-1">
    {#each NAV_ITEMS as item (item.id)}
      <Button
        color={activeSection === item.id ? "primary" : "light"}
        class="justify-start"
        onclick={() => handleSectionChange(item.id)}
      >
        {item.label}
      </Button>
    {/each}
  </nav>

  <!-- Content -->
  <div class="min-w-0 flex-1 overflow-y-auto pr-3 pl-1">
    {#if activeSection === "appearance"}
      <AppearancePane
        themeMode={(configurationStoreState.config.general.themeMode as ThemeMode) || "system"}
        activeTheme={configurationStoreState.config.general.activeTheme}
        defaultFontFamily={configurationStoreState.config.general.defaultFontFamily}
        monoFontFamily={configurationStoreState.config.general.monoFontFamily}
        themes={themesState || []}
        {fontLists}
        {fontsLoading}
        {fontsReady}
        {fontsError}
        {zoomLevel}
        onThemeModeChange={handleThemeModeChange}
        onThemeSelect={(themeId) => void handleThemeSelect(themeId)}
        onSansFontChange={(fontFamily) => void handleSansFontChange(fontFamily)}
        onMonoFontChange={(fontFamily) => void handleMonoFontChange(fontFamily)}
        onZoomLevelChange={(level) => void handleZoomLevelChange(level)}
        onRefreshFonts={() => void handleRefreshFonts()}
        onResetSansFont={() => void handleSansFontReset()}
        onResetMonoFont={() => void handleMonoFontReset()}
        onResetZoomLevel={() => void handleZoomLevelReset()}
      />
    {:else if activeSection === "general"}
      <div class="flex flex-col gap-5">
        <div>
          <h2 class="text-base font-semibold text-neutral-900 dark:text-neutral-100">General</h2>
          <p class="mt-1 text-sm text-neutral-500 dark:text-neutral-400">
            Configure general application behavior and request defaults.
          </p>
        </div>

        <Toggle
          bind:checked={configurationStoreState.config.general.checkForUpdates}
          onchange={handleUpdateToggleSettingChange}
        >
          Check for updates on startup
        </Toggle>

        <Toggle
          bind:checked={configurationStoreState.config.general.includePrereleaseUpdates}
          onchange={handleUpdateToggleSettingChange}
          disabled={!configurationStoreState.config.general.checkForUpdates}
        >
          Include release candidates in update discovery
        </Toggle>

        <div>
          <h3 class="mb-3 text-sm font-semibold text-neutral-700 dark:text-neutral-300">
            Request Defaults
          </h3>
          <div class="flex flex-col gap-4">
            <div class="grid grid-cols-2 gap-3">
              <div class="flex flex-col gap-1">
                <Label for="timeout">Timeout (seconds)</Label>
                <Input
                  id="timeout"
                  type="number"
                  size="sm"
                  bind:value={configurationStoreState.config.request.timeoutSeconds}
                  min="0"
                  step="1"
                  placeholder="Default: {defaultConfig.request.timeoutSeconds}"
                  oninput={handleTextSettingChange}
                />
              </div>
              <div class="flex flex-col gap-1">
                <Label for="max-redirects">Max Redirects</Label>
                <Input
                  id="max-redirects"
                  type="number"
                  size="sm"
                  bind:value={configurationStoreState.config.request.maxRedirects}
                  min="0"
                  step="1"
                  placeholder="Default: {defaultConfig.request.maxRedirects}"
                  disabled={!configurationStoreState.config.request.followRedirects}
                  oninput={handleTextSettingChange}
                />
              </div>
            </div>

            <div class="flex flex-col gap-1">
              <Label for="user-agent">Default User Agent</Label>
              <Input
                id="user-agent"
                type="text"
                size="sm"
                bind:value={configurationStoreState.config.request.defaultUserAgent}
                placeholder="Default: {defaultConfig.request.defaultUserAgent}"
                oninput={handleTextSettingChange}
              />
            </div>

            <div class="flex flex-col gap-1">
              <Label for="proxy">Proxy URL</Label>
              <Input
                id="proxy"
                type="text"
                size="sm"
                bind:value={configurationStoreState.config.request.proxyUrl}
                placeholder="http://user:pass@host:port (optional)"
                oninput={handleTextSettingChange}
              />
            </div>

            <div class="flex flex-col gap-3">
              <Toggle
                bind:checked={configurationStoreState.config.request.followRedirects}
                onchange={handleToggleSettingChange}
              >
                Follow Redirects
              </Toggle>
              <Toggle
                bind:checked={configurationStoreState.config.request.validateSSL}
                onchange={handleToggleSettingChange}
              >
                Validate SSL Certificates
              </Toggle>
            </div>
          </div>
        </div>

        {#if configSaveStatus === "saving"}
          <Helper color="gray">Saving...</Helper>
        {:else if configSaveStatus === "saved"}
          <Helper color="green">Saved</Helper>
        {/if}
      </div>
    {:else if activeSection === "troubleshooting"}
      <div class="flex flex-col gap-6">
        <div>
          <h2 class="text-base font-semibold text-neutral-900 dark:text-neutral-100">
            Troubleshooting
          </h2>
          <p class="mt-1 text-sm text-neutral-500 dark:text-neutral-400">
            Collect diagnostics and enable verbose logging when investigating issues.
          </p>
        </div>

        <div
          class="flex flex-col gap-2 rounded-lg border border-neutral-200 p-4 dark:border-neutral-700"
        >
          <h3 class="text-sm font-semibold text-neutral-700 dark:text-neutral-300">Debug mode</h3>
          <p class="text-sm text-neutral-500 dark:text-neutral-400">
            Enable detailed runtime logs. This can increase log volume.
          </p>
          <Toggle bind:checked={configurationStoreState.config.general.debugMode}>
            Enable debug mode
          </Toggle>
        </div>

        <div
          class="flex flex-col gap-2 rounded-lg border border-neutral-200 p-4 dark:border-neutral-700"
        >
          <h3 class="text-sm font-semibold text-neutral-700 dark:text-neutral-300">Logs export</h3>
          <p class="text-sm text-neutral-500 dark:text-neutral-400">
            Download a ZIP archive with all application log files, including rotated logs.
          </p>
          <div>
            <Button color="light" onclick={handleLogsExport} disabled={isExportingLogs}>
              {isExportingLogs ? "Preparing archive..." : "Download logs (.zip)"}
            </Button>
          </div>
        </div>
      </div>
    {:else if activeSection === "hosts"}
      <div class="flex flex-col gap-4">
        <div>
          <h2 class="text-base font-semibold text-neutral-900 dark:text-neutral-100">Hosts</h2>
          <p class="mt-1 text-sm text-neutral-500 dark:text-neutral-400">
            Manage specific TLS and certificate configurations for your target hosts.
          </p>
        </div>

        {#if !editingHost}
          <div class="flex justify-end">
            <Button color="primary" onclick={startAddHost}>Add Host</Button>
          </div>

          {#if hostsList.length === 0}
            <FeedbackEmptyState
              variant="info"
              title="No specific host configuration found."
              compact
            />
          {:else}
            <Table>
              <TableHead>
                <TableHeadCell>Hostname</TableHeadCell>
                <TableHeadCell>TLS</TableHeadCell>
                <TableHeadCell>Skip Verify</TableHeadCell>
                <TableHeadCell>Actions</TableHeadCell>
              </TableHead>
              <TableBody>
                {#each hostsList as h (h.id)}
                  <TableBodyRow>
                    <TableBodyCell>
                      <span class="font-medium">{h.name}</span>
                    </TableBodyCell>
                    <TableBodyCell>
                      <Badge color={h.tlsConfig.enabled ? "green" : "gray"}>
                        {h.tlsConfig.enabled ? "Enabled" : "Disabled"}
                      </Badge>
                    </TableBodyCell>
                    <TableBodyCell>
                      <Badge color={h.tlsConfig.insecureSkipVerify ? "yellow" : "gray"}>
                        {h.tlsConfig.insecureSkipVerify ? "Yes" : "No"}
                      </Badge>
                    </TableBodyCell>
                    <TableBodyCell>
                      <div class="flex items-center gap-2">
                        <Button size="xs" color="light" onclick={() => editExistingHost(h)}>
                          Edit
                        </Button>
                        <Button size="xs" color="red" onclick={() => handleDeleteHost(h.name)}>
                          Delete
                        </Button>
                      </div>
                    </TableBodyCell>
                  </TableBodyRow>
                {/each}
              </TableBody>
            </Table>
          {/if}
        {:else}
          <!-- Host edit form -->
          <div class="flex flex-col gap-4">
            <h3 class="text-sm font-semibold text-neutral-700 dark:text-neutral-300">
              {editingHostName ? "Edit Host" : "New Host"}
            </h3>

            <div class="flex flex-col gap-1">
              <Label for="h-name">Hostname (e.g. api.company.com or localhost:8443)</Label>
              <Input
                id="h-name"
                type="text"
                size="sm"
                bind:value={editingHostName}
                placeholder="Hostname"
              />
            </div>

            <Toggle bind:checked={customTlsEnabled}>Enable Custom TLS Configuration</Toggle>

            {#if customTlsEnabled}
              <div
                class="flex flex-col gap-4 rounded-lg border border-neutral-200 p-4 dark:border-neutral-700"
              >
                <div class="flex flex-col gap-1">
                  <Toggle bind:checked={editingHost.tlsConfig.insecureSkipVerify}>
                    Insecure Skip Verify
                  </Toggle>
                  <Helper color="red">Not recommended for production use.</Helper>
                </div>

                <div class="flex flex-col gap-1">
                  <Label>Public Certificate (.crt, .pem)</Label>
                  <div class="flex items-center gap-2">
                    <Input
                      type="text"
                      size="sm"
                      value={editingHost.tlsConfig.publicCertificateFilePath}
                      readonly
                      class="flex-1"
                    />
                    <Button
                      size="sm"
                      color="light"
                      onclick={() => pickCertFile("publicCertificateFilePath")}
                    >
                      Browse
                    </Button>
                  </div>
                </div>

                <div class="flex flex-col gap-1">
                  <Label>Private Key (.key)</Label>
                  <div class="flex items-center gap-2">
                    <Input
                      type="text"
                      size="sm"
                      value={editingHost.tlsConfig.privateKeyFilePath}
                      readonly
                      class="flex-1"
                    />
                    <Button
                      size="sm"
                      color="light"
                      onclick={() => pickCertFile("privateKeyFilePath")}
                    >
                      Browse
                    </Button>
                  </div>
                </div>

                <div class="flex flex-col gap-1">
                  <Label>CA Certificate (Root CA)</Label>
                  <div class="flex items-center gap-2">
                    <Input
                      type="text"
                      size="sm"
                      value={editingHost.tlsConfig.caCertificateFilePath}
                      readonly
                      class="flex-1"
                    />
                    <Button
                      size="sm"
                      color="light"
                      onclick={() => pickCertFile("caCertificateFilePath")}
                    >
                      Browse
                    </Button>
                  </div>
                </div>
              </div>
            {/if}

            <div class="flex flex-col gap-3">
              <h4 class="text-sm font-semibold text-neutral-700 dark:text-neutral-300">Cookies</h4>
              <p class="text-sm text-neutral-500 dark:text-neutral-400">
                These cookies will be automatically added to all requests sent to this host.
              </p>
              <div class="flex flex-col gap-2">
                {#each editingCookies as cookie (cookie.id)}
                  <div class="flex items-center gap-2">
                    <Input
                      size="sm"
                      bind:value={cookie.key}
                      placeholder="Cookie Name"
                      class="flex-1"
                    />
                    <Input size="sm" bind:value={cookie.value} placeholder="Value" class="flex-1" />
                    <Button
                      size="xs"
                      color="red"
                      aria-label="Remove cookie"
                      onclick={() => removeCookieRow(cookie.id)}
                    >
                      <TrashBinOutline size="xs" />
                    </Button>
                  </div>
                {/each}
                <div>
                  <Button size="sm" color="alternative" onclick={addCookieRow}>+ Add Cookie</Button>
                </div>
              </div>
            </div>

            <div
              class="flex items-center justify-end gap-2 border-t border-neutral-200 pt-4 dark:border-neutral-700"
            >
              <Button color="light" onclick={() => (editingHost = null)}>Cancel</Button>
              <Button color="primary" onclick={handleSaveHost} disabled={!editingHostName.trim()}>
                Save Host
              </Button>
            </div>
          </div>
        {/if}
      </div>
    {:else if activeSection === "about"}
      <About />
    {/if}
  </div>
</div>

<!-- Delete host confirm modal -->
{#if deleteHostModal.open}
  <Modal title="Delete Host" bind:open={deleteHostModal.open} size="sm">
    {#if $topModalId === deleteHostModal.id}
      <ToastContainer />
    {/if}
    <p class="text-neutral-700 dark:text-neutral-300">
      Are you sure you want to delete the host configuration for
      <strong>{hostToDelete}</strong>?
    </p>
    {#snippet footer()}
      <div class="flex w-full items-center justify-end gap-2">
        <Button color="light" onclick={() => (deleteHostModal.open = false)}>Cancel</Button>
        <Button color="red" onclick={confirmDeleteHost}>Delete</Button>
      </div>
    {/snippet}
  </Modal>
{/if}
