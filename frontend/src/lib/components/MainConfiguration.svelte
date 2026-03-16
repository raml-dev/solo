<script lang="ts">
  import ToastContainer from "$src/lib/components/base/ToastContainer.svelte";
  import FeedbackEmptyState from "$src/lib/components/common/FeedbackEmptyState.svelte";
  import ThemePreview from "$src/lib/components/Settings/ThemePreview.svelte";
  import { configurationStore } from "$src/lib/stores/configurationStore";
  import { modalStack, topModalId } from "$src/lib/stores/modalStackStore";
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
  import TrashBinOutline from "flowbite-svelte-icons/TrashBinOutline.svelte";
  import Badge from "flowbite-svelte/Badge.svelte";
  import Button from "flowbite-svelte/Button.svelte";
  import Helper from "flowbite-svelte/Helper.svelte";
  import Input from "flowbite-svelte/Input.svelte";
  import Label from "flowbite-svelte/Label.svelte";
  import Modal from "flowbite-svelte/Modal.svelte";
  import Radio from "flowbite-svelte/Radio.svelte";
  import Table from "flowbite-svelte/Table.svelte";
  import TableBody from "flowbite-svelte/TableBody.svelte";
  import TableBodyCell from "flowbite-svelte/TableBodyCell.svelte";
  import TableBodyRow from "flowbite-svelte/TableBodyRow.svelte";
  import TableHead from "flowbite-svelte/TableHead.svelte";
  import TableHeadCell from "flowbite-svelte/TableHeadCell.svelte";
  import Toggle from "flowbite-svelte/Toggle.svelte";
  import { onMount } from "svelte";

  // --- Nav ---
  type SettingsSection = "general" | "themes" | "hosts";
  let activeSection: SettingsSection = $state("general");

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
  let selectedThemeMode = $state("system");

  let initialized = $state(false);
  $effect(() => {
    if (!initialized && $config?.general?.activeTheme && ($allThemes || []).length > 0) {
      activeThemeId = $config.general.activeTheme;
      selectedThemeMode = $config.general.themeMode || "system";
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
      themeMode: cfg.general?.themeMode ?? "light",
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
      current.general.themeMode = editableConfig.general.themeMode;
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

  $effect(() => {
    if (!initialized) return;
    configurationStore.applyThemeMode(selectedThemeMode);
    editableConfig.general = new configuration.GeneralSettings({
      ...editableConfig.general,
      themeMode: selectedThemeMode
    });
    void persistRequestSettings();
  });

  // --- Hosts ---
  type HostCookieRow = { id: string; key: string; value: string };

  let hostsList: host.Host[] = $state([]);
  let editingHost: host.Host | null = $state(null);
  let editingCookies: HostCookieRow[] = $state([]);

  // Delete confirm modal
  const deleteHostModalId = `settings-delete-host-${Math.random().toString(36).slice(2)}`;
  let showDeleteHostModal = $state(false);
  let hostToDelete = $state("");

  $effect(() => {
    if (showDeleteHostModal) modalStack.open(deleteHostModalId);
    else modalStack.close(deleteHostModalId);
  });

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

  function handleDeleteHost(name: string) {
    hostToDelete = name;
    showDeleteHostModal = true;
  }

  async function confirmDeleteHost() {
    showDeleteHostModal = false;
    try {
      await DeleteHost(hostToDelete);
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

<div class="flex h-full gap-6">
  <!-- Sidebar nav -->
  <nav class="flex w-36 shrink-0 flex-col gap-1">
    {#each NAV_ITEMS as item (item.id)}
      <Button
        color={activeSection === item.id ? "primary" : "light"}
        class="justify-start"
        onclick={() => (activeSection = item.id)}
      >
        {item.label}
      </Button>
    {/each}
  </nav>

  <!-- Content -->
  <div class="min-w-0 flex-1 overflow-y-auto">
    {#if activeSection === "themes"}
      <div class="flex flex-col gap-4">
        <div>
          <h2 class="text-base font-semibold text-neutral-900 dark:text-neutral-100">Themes</h2>
          <p class="mt-1 text-sm text-neutral-500 dark:text-neutral-400">
            Personalize your experience with themes that match your style.
          </p>
        </div>

        <!-- Display mode selector -->
        <div class="flex flex-col gap-2">
          <p class="text-sm font-medium text-neutral-700 dark:text-neutral-300">Display mode</p>
          <div class="flex gap-4">
            {#each [{ value: "light", label: "Light" }, { value: "dark", label: "Dark" }, { value: "system", label: "System" }] as mode (mode.value)}
              <Radio name="themeMode" bind:group={selectedThemeMode} value={mode.value}
                >{mode.label}</Radio
              >
            {/each}
          </div>
        </div>

        <div class="grid grid-cols-3 gap-3">
          {#each $allThemes || [] as t (t.id)}
            <button
              class="flex cursor-pointer flex-col items-center rounded-lg border p-3 text-left transition-all hover:border-primary-400"
              class:border-primary-500={activeThemeId === t.id}
              class:ring-2={activeThemeId === t.id}
              class:ring-primary-500={activeThemeId === t.id}
              class:border-neutral-200={activeThemeId !== t.id}
              class:dark:border-neutral-700={activeThemeId !== t.id}
              onclick={() => handleThemeSelect(t.id)}
            >
              <div class="mb-2 w-full overflow-hidden rounded">
                <ThemePreview seeds={t.config?.seeds} />
              </div>
              <p class="text-sm font-medium text-neutral-700 dark:text-neutral-200">
                {formatThemeName(t.label)}
              </p>
              {#if activeThemeId === t.id}
                <Badge color="primary" class="mt-1">Active</Badge>
              {/if}
            </button>
          {/each}
        </div>
      </div>
    {:else if activeSection === "general"}
      <div class="flex flex-col gap-5">
        <div>
          <h2 class="text-base font-semibold text-neutral-900 dark:text-neutral-100">General</h2>
          <p class="mt-1 text-sm text-neutral-500 dark:text-neutral-400">
            Configure general application behavior and request defaults.
          </p>
        </div>

        <Toggle bind:checked={editableConfig.general.checkForUpdates}>
          Check for updates on startup
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
                  bind:value={editableConfig.request.timeoutSeconds}
                  min="0"
                  step="1"
                  placeholder="Default: {defaultConfig.request.timeoutSeconds}"
                />
              </div>
              <div class="flex flex-col gap-1">
                <Label for="max-redirects">Max Redirects</Label>
                <Input
                  id="max-redirects"
                  type="number"
                  size="sm"
                  bind:value={editableConfig.request.maxRedirects}
                  min="0"
                  step="1"
                  placeholder="Default: {defaultConfig.request.maxRedirects}"
                  disabled={!editableConfig.request.followRedirects}
                />
              </div>
            </div>

            <div class="flex flex-col gap-1">
              <Label for="user-agent">Default User Agent</Label>
              <Input
                id="user-agent"
                type="text"
                size="sm"
                bind:value={editableConfig.request.defaultUserAgent}
                placeholder="Default: {defaultConfig.request.defaultUserAgent}"
              />
            </div>

            <div class="flex flex-col gap-1">
              <Label for="proxy">Proxy URL</Label>
              <Input
                id="proxy"
                type="text"
                size="sm"
                bind:value={editableConfig.request.proxyUrl}
                placeholder="http://user:pass@host:port (optional)"
              />
            </div>

            <div class="flex flex-col gap-3">
              <Toggle bind:checked={editableConfig.request.followRedirects}>
                Follow Redirects
              </Toggle>
              <Toggle bind:checked={editableConfig.request.validateSSL}>
                Validate SSL Certificates
              </Toggle>
            </div>
          </div>
        </div>

        {#if saveStatus === "saving"}
          <Helper color="gray">Saving…</Helper>
        {:else if saveStatus === "saved"}
          <Helper color="green">Saved</Helper>
        {/if}
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
              {editingHost.name ? "Edit Host" : "New Host"}
            </h3>

            <div class="flex flex-col gap-1">
              <Label for="h-name">Hostname (e.g. api.company.com or localhost:8443)</Label>
              <Input
                id="h-name"
                type="text"
                size="sm"
                bind:value={editingHost.name}
                placeholder="Hostname"
              />
            </div>

            <Toggle bind:checked={editingHost.tlsConfig.enabled}>
              Enable Custom TLS Configuration
            </Toggle>

            {#if editingHost.tlsConfig.enabled}
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
              <Button color="primary" onclick={handleSaveHost} disabled={!editingHost.name}>
                Save Host
              </Button>
            </div>
          </div>
        {/if}
      </div>
    {/if}
  </div>
</div>

<!-- Delete host confirm modal -->
{#if showDeleteHostModal}
  <Modal title="Delete Host" bind:open={showDeleteHostModal} size="sm">
    {#if $topModalId === deleteHostModalId}
      <ToastContainer />
    {/if}
    <p class="text-neutral-700 dark:text-neutral-300">
      Are you sure you want to delete the host configuration for
      <strong>{hostToDelete}</strong>?
    </p>
    {#snippet footer()}
      <div class="flex w-full items-center justify-end gap-2">
        <Button color="light" onclick={() => (showDeleteHostModal = false)}>Cancel</Button>
        <Button color="red" onclick={confirmDeleteHost}>Delete</Button>
      </div>
    {/snippet}
  </Modal>
{/if}
