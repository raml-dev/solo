<script lang="ts">
  import Button from "$src/lib/components/base/Button.svelte";
  import DropZone from "$src/lib/components/base/DropZone.svelte";
  import Modal from "$src/lib/components/base/Modal.svelte";
  import Tab from "$src/lib/components/base/Tab.svelte";
  import Tabs from "$src/lib/components/base/Tabs.svelte";
  import EnvironmentEditor from "$src/lib/components/Environment/EnvironmentEditor.svelte";
  import EnvironmentItem from "$src/lib/components/Environment/EnvironmentItem.svelte";
  import EnvironmentModals from "$src/lib/components/Environment/EnvironmentModals.svelte";
  import GitEnvImportView from "$src/lib/components/GitEnvImportView.svelte";
  import GitStatusPanel from "$src/lib/components/GitStatusPanel.svelte";
  import { environmentStore } from "$src/lib/stores/environmentStore";
  import { notifications } from "$src/lib/stores/notificationStore";
  import {
    GetGitEnvironmentStatus,
    GitEnvAbortRebase,
    GitEnvDiscardChanges,
    GitEnvKeepOurs,
    GitEnvKeepTheirs,
    ImportBrunoEnvironment,
    ImportPostmanEnvironment,
    OpenEnvironmentInTerminal,
    SelectFile,
    SyncGitEnvironment
  } from "$wails/go/main/App";
  import { environment } from "$wails/go/models";
  import { SvelteSet } from "svelte/reactivity";

  let showNewEnvironmentDialog = $state(false);
  let showDeleteConfirmDialog = $state(false);
  let deleteTarget: string | null = $state(null);
  let activeMenu: string | null = $state(null);
  let showImportSelector = $state(false);
  let importActiveTab = $state("postman");
  let showOverwriteConfirmDialog = $state(false);
  let pendingImport: { format: "postman" | "bruno"; path: string } | null = null;
  let overwriteTargetName: string | null = $state(null);
  let focusedEnvironmentName: string | null = $state(null);
  let syncingEnvironments: Set<string> = $state(new Set());
  let gitStatusEnvId: string | null = $state(null);
  let gitStatusEnvName: string | null = $state(null);

  let environments = $derived($environmentStore.environments);
  $effect(() => {
    const exists =
      focusedEnvironmentName && environments.some((e) => e.name === focusedEnvironmentName);
    if (!exists) {
      focusedEnvironmentName =
        $environmentStore.selectedEnvironmentName || environments[0]?.name || null;
    }
  });
  let selectedEnvironment = $derived(
    environments.find((env) => env.name === focusedEnvironmentName) || null
  );

  function openEnvironment(name: string) {
    focusedEnvironmentName = name;
  }

  function activateEnvironment(name: string) {
    environmentStore.selectEnvironment(name);
  }

  async function handleUpdateEnvironment(data: {
    name: string;
    values: Record<string, environment.ValueType>;
  }) {
    try {
      const { name, values } = data;
      const env = environments.find((e) => e.name === name);
      if (env) {
        const updated = new environment.Environment({
          ...env,
          values
        });
        await environmentStore.updateEnvironment(updated);
      }
    } catch (err) {
      notifications.error("Failed to update environment", String(err));
    }
  }

  async function handleCreateEnvironment(name: string) {
    const trimmed = name.trim();
    if (!trimmed) return;

    const exists = environments.some((env) => env.name.toLowerCase() === trimmed.toLowerCase());
    if (exists) {
      notifications.warning(`Environment "${trimmed}" already exists`);
      return;
    }

    try {
      await environmentStore.createEnvironment(trimmed);
      showNewEnvironmentDialog = false;
    } catch {
      // error already shown by store
    }
  }

  function handleDeleteEnvironment(name: string) {
    deleteTarget = name;
    showDeleteConfirmDialog = true;
    activeMenu = null;
  }

  async function confirmDelete() {
    if (!deleteTarget) return;

    try {
      await environmentStore.deleteEnvironment(deleteTarget);
      showDeleteConfirmDialog = false;
      deleteTarget = null;
    } catch {
      // error already shown by store
    }
  }

  function toggleMenu(environmentName: string) {
    activeMenu = activeMenu === environmentName ? null : environmentName;
  }

  function handleLayoutClick() {
    activeMenu = null;
  }

  async function handleSync(environmentId: string) {
    syncingEnvironments.add(environmentId);
    syncingEnvironments = new SvelteSet(syncingEnvironments);
    try {
      await SyncGitEnvironment(environmentId);
      notifications.success("Git environment synced successfully");
      await environmentStore.loadEnvironments();
    } catch (err) {
      notifications.error("Sync failed", String(err));
    } finally {
      syncingEnvironments.delete(environmentId);
      syncingEnvironments = new SvelteSet(syncingEnvironments);
    }
  }

  function handleGitStatus(environmentId: string) {
    const env = environments.find((e) => e.id === environmentId);
    if (!env) return;
    gitStatusEnvId = environmentId;
    gitStatusEnvName = env.name;
  }

  function openImportModal() {
    importActiveTab = "postman";
    showImportSelector = true;
  }

  function parseExistingNameFromError(message: string): string | null {
    const match = message.match(/environment\s+([^\s]+)\s+already exists/i);
    return match ? match[1] : null;
  }

  async function executeImport(format: "postman" | "bruno", path: string, overwrite: boolean) {
    try {
      if (format === "postman") {
        await ImportPostmanEnvironment(path, overwrite);
      } else {
        await ImportBrunoEnvironment(path, overwrite);
      }
      await environmentStore.loadEnvironments();
      notifications.success("Environment imported successfully");
    } catch (err) {
      const message = String(err ?? "Failed to import environment");
      const existingName = parseExistingNameFromError(message);
      if (!overwrite && existingName) {
        pendingImport = { format, path };
        overwriteTargetName = existingName;
        showOverwriteConfirmDialog = true;
        return;
      }
      notifications.error("Failed to import environment", message);
    }
  }

  async function handleImportPostman(path?: string) {
    const filePath =
      path ?? (await SelectFile("Select Postman Environment", "*.json", "JSON Files"));
    if (!filePath) return;
    await executeImport("postman", filePath, false);
  }

  async function handleImportBruno(path?: string) {
    const filePath = path ?? (await SelectFile("Select Bruno Environment", "*.bru", "Bruno Files"));
    if (!filePath) return;
    await executeImport("bruno", filePath, false);
  }

  async function handleSelectImportFormat(format: "postman" | "bruno") {
    showImportSelector = false;
    if (format === "postman") {
      await handleImportPostman();
    } else if (format === "bruno") {
      await handleImportBruno();
    }
  }

  async function confirmOverwrite() {
    if (!pendingImport) return;
    const { format, path } = pendingImport;
    pendingImport = null;
    showOverwriteConfirmDialog = false;
    await executeImport(format, path, true);
  }

  function closeOverwriteConfirmDialog() {
    pendingImport = null;
    overwriteTargetName = null;
    showOverwriteConfirmDialog = false;
  }
</script>

<div class="environment-manager-layout" onclick={handleLayoutClick}>
  <div class="environment-list">
    <div class="header">
      <div class="header-title">
        <div class="header-actions">
          <Button variant="secondary" size="small" click={openImportModal}>Import</Button>
          <Button variant="primary" size="small" click={() => (showNewEnvironmentDialog = true)}>
            New
          </Button>
        </div>
      </div>
    </div>

    {#if $environmentStore.loading}
      <div class="loading">Loading environments...</div>
    {/if}

    <div class="environments">
      {#each environments as environment (environment.id)}
        <EnvironmentItem
          env={environment}
          menuOpen={activeMenu === environment.name}
          isActive={environment.name === $environmentStore.selectedEnvironmentName}
          isFocused={environment.name === focusedEnvironmentName}
          isSyncing={syncingEnvironments.has(environment.id)}
          onOpen={openEnvironment}
          onActivate={activateEnvironment}
          onDelete={handleDeleteEnvironment}
          onToggleMenu={toggleMenu}
          onSync={handleSync}
          onGitStatus={handleGitStatus}
        />
      {/each}
    </div>
    {#if environments.length === 0 && !$environmentStore.loading}
      <div class="empty-state">
        <p>No environments yet</p>
        <p class="hint">Create your first environment to get started</p>
      </div>
    {/if}
  </div>
  <div class="environment-editor-pane">
    <EnvironmentEditor env={selectedEnvironment} onUpdate={handleUpdateEnvironment} />
  </div>
</div>

<EnvironmentModals
  bind:showNewEnvironmentDialog
  bind:showDeleteConfirmDialog
  bind:deleteTarget
  onCreate={handleCreateEnvironment}
  onConfirmDelete={confirmDelete}
  onCloseNew={() => (showNewEnvironmentDialog = false)}
  onCloseDelete={() => (showDeleteConfirmDialog = false)}
/>

{#if showImportSelector}
  <Modal title="Import Environment" toggleFn={() => (showImportSelector = false)} size="wide">
    <div class="import-modal-body">
      <Tabs bind:activeValue={importActiveTab}>
        <Tab title="Postman" value="postman">
          <DropZone
            title="Drop your Postman environment here"
            subtitle="Supports Postman Environment JSON"
            onDrop={async (e) => {
              const paths = e.paths;
              showImportSelector = false;
              if (paths.length > 0) {
                await handleImportPostman(paths[0]);
              } else {
                await handleImportPostman();
              }
            }}
          >
            {#snippet icon()}
              <svg
                width="44"
                height="44"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="1.4"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                <polyline points="17 8 12 3 7 8" />
                <line x1="12" y1="3" x2="12" y2="15" />
              </svg>
            {/snippet}
          </DropZone>
        </Tab>

        <Tab title="Bruno" value="bruno">
          <DropZone
            title="Drop your Bruno environment here"
            subtitle="Supports Bruno environment .bru files"
            onDrop={async (e) => {
              const paths = e.paths;
              showImportSelector = false;
              if (paths.length > 0) {
                await handleImportBruno(paths[0]);
              } else {
                await handleImportBruno();
              }
            }}
          >
            {#snippet icon()}
              <svg
                width="44"
                height="44"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="1.4"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
                <polyline points="9 22 9 12 15 12 15 22" />
              </svg>
            {/snippet}
          </DropZone>
        </Tab>

        <Tab title="Git" value="git">
          <GitEnvImportView onImported={() => (showImportSelector = false)} />
        </Tab>
      </Tabs>
    </div>

    {#snippet additional_buttons()}
      {#if importActiveTab === "postman"}
        <Button variant="primary" click={() => handleSelectImportFormat("postman")}
          >Select file…</Button
        >
      {:else if importActiveTab === "bruno"}
        <Button variant="primary" click={() => handleSelectImportFormat("bruno")}
          >Select file…</Button
        >
      {/if}
    {/snippet}
  </Modal>
{/if}

{#if showOverwriteConfirmDialog}
  <Modal toggleFn={closeOverwriteConfirmDialog} size="wide">
    <h3>Overwrite environment?</h3>
    <p>Environment "{overwriteTargetName}" already exists.</p>
    <p class="warning">Do you want to overwrite it?</p>
    {#snippet additional_buttons()}
      <Button variant="danger" click={confirmOverwrite}>Overwrite</Button>
    {/snippet}
  </Modal>
{/if}

{#if gitStatusEnvId && gitStatusEnvName}
  <GitStatusPanel
    entityId={gitStatusEnvId}
    entityName={gitStatusEnvName}
    fnGetStatus={GetGitEnvironmentStatus}
    fnSync={SyncGitEnvironment}
    fnKeepOurs={GitEnvKeepOurs}
    fnKeepTheirs={GitEnvKeepTheirs}
    fnAbortRebase={GitEnvAbortRebase}
    fnDiscard={GitEnvDiscardChanges}
    fnOpenTerminal={OpenEnvironmentInTerminal}
    onReload={environmentStore.loadEnvironments}
    onClose={() => {
      gitStatusEnvId = null;
      gitStatusEnvName = null;
    }}
  />
{/if}
