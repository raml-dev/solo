<script lang="ts">
  import { run } from "svelte/legacy";

  import { environmentStore } from "../../stores/environmentStore";
  import { notifications } from "../../stores/notificationStore";
  import Button from "../base/Button.svelte";
  import EnvironmentItem from "./EnvironmentItem.svelte";
  import EnvironmentEditor from "./EnvironmentEditor.svelte";
  import EnvironmentModals from "./EnvironmentModals.svelte";
  import Modal from "../base/Modal.svelte";
  import Tabs from "../base/Tabs.svelte";
  import Tab from "../base/Tab.svelte";
  import DropZone from "../base/DropZone.svelte";
  import {
    ImportPostmanEnvironment,
    ImportBrunoEnvironment,
    SelectFile,
    SyncGitEnvironment,
    GetGitEnvironmentStatus,
    GitEnvKeepOurs,
    GitEnvKeepTheirs,
    GitEnvAbortRebase,
    GitEnvDiscardChanges,
    OpenEnvironmentInTerminal
  } from "../../../../wailsjs/go/main/App";
  import type { environment } from "../../../../wailsjs/go/models";
  import GitEnvImportView from "../GitEnvImportView.svelte";
  import GitStatusPanel from "../GitStatusPanel.svelte";
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
  let selectedEnvironment: environment.Environment | null = $state(null);
  let focusedEnvironmentName: string | null = $state(null);
  let syncingEnvironments: Set<string> = $state(new Set());
  let gitStatusEnvId: string | null = $state(null);
  let gitStatusEnvName: string | null = $state(null);

  let environments = $derived($environmentStore.environments);
  run(() => {
    const exists =
      focusedEnvironmentName && environments.some((e) => e.name === focusedEnvironmentName);
    if (!exists) {
      focusedEnvironmentName =
        $environmentStore.selectedEnvironmentName || environments[0]?.name || null;
    }
  });
  run(() => {
    selectedEnvironment = environments.find((env) => env.name === focusedEnvironmentName) || null;
  });

  function openEnvironment(name: string) {
    focusedEnvironmentName = name;
  }

  function activateEnvironment(name: string) {
    environmentStore.selectEnvironment(name);
  }

  async function handleUpdateEnvironment(
    event: CustomEvent<{ name: string; values: Record<string, environment.ValueType> }>
  ) {
    try {
      const { name, values } = event.detail;
      const env = environments.find((e) => e.name === name);
      if (env) {
        env.values = values;
        await environmentStore.updateEnvironment(env);
      }
    } catch (err) {
      notifications.error("Failed to update environment", String(err));
    }
  }

  async function handleCreateEnvironment(event: CustomEvent<string>) {
    const trimmed = event.detail.trim();
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

  function handleDeleteEnvironment(event: CustomEvent<string>) {
    deleteTarget = event.detail;
    showDeleteConfirmDialog = true;
    activeMenu = null;
  }

  async function confirmDelete() {
    if (!deleteTarget) return;

    try {
      await environmentStore.deleteEnvironment(deleteTarget);
      if (selectedEnvironment && selectedEnvironment.name === deleteTarget) {
        selectedEnvironment = null;
      }
      showDeleteConfirmDialog = false;
      deleteTarget = null;
    } catch {
      // error already shown by store
    }
  }

  function toggleMenu(event: CustomEvent<string>) {
    const environmentName = event.detail;
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
          on:open={(e) => openEnvironment(e.detail)}
          on:activate={(e) => activateEnvironment(e.detail)}
          on:delete={handleDeleteEnvironment}
          on:toggleMenu={toggleMenu}
          on:sync={(e) => handleSync(e.detail)}
          on:gitStatus={(e) => handleGitStatus(e.detail)}
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
    <EnvironmentEditor env={selectedEnvironment} on:update={handleUpdateEnvironment} />
  </div>
</div>

<EnvironmentModals
  bind:showNewEnvironmentDialog
  bind:showDeleteConfirmDialog
  bind:deleteTarget
  on:create={handleCreateEnvironment}
  on:confirmDelete={confirmDelete}
  on:closeNew={() => (showNewEnvironmentDialog = false)}
  on:closeDelete={() => (showDeleteConfirmDialog = false)}
/>

{#if showImportSelector}
  <Modal title="Import Environment" toggleFn={() => (showImportSelector = false)} size="wide">
    <div class="import-modal-body">
      <Tabs bind:activeValue={importActiveTab}>
        <Tab title="Postman" value="postman">
          <DropZone
            title="Drop your Postman environment here"
            subtitle="Supports Postman Environment JSON"
            on:drop={async (e) => {
              const paths = e.detail.paths;
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
            on:drop={async (e) => {
              const paths = e.detail.paths;
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
          <GitEnvImportView on:imported={() => (showImportSelector = false)} />
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
    on:close={() => {
      gitStatusEnvId = null;
      gitStatusEnvName = null;
    }}
  />
{/if}

<style>
  .environment-manager-layout {
    display: flex;
    height: 100%;
    overflow: hidden;
    border-radius: var(--radius-lg);
    background: var(--bg-primary);
  }

  .environment-list {
    width: fit-content;
    min-width: 180px;
    max-width: 320px;
    flex-shrink: 0;
    background: var(--bg-secondary);
    border-right: 1px solid var(--border);
    display: flex;
    flex-direction: column;
    overflow: hidden;
    border-radius: var(--radius-lg) 0 0 var(--radius-lg);
    padding: var(--space-lg) var(--space-sm);
    gap: var(--space-sm);
  }

  .environment-editor-pane {
    flex-grow: 1;
    overflow: auto;
    background: var(--bg-primary);
    border-radius: 0 var(--radius-lg) var(--radius-lg) 0;
  }

  .header {
    padding: 0 var(--space-sm);
  }

  .header-title {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .header-actions {
    display: flex;
    gap: var(--space-xs);
    align-items: center;
    width: 100%;
  }

  .header-actions :global(button) {
    flex: 1;
  }

  .import-modal-body {
    margin: calc(-1 * var(--space-lg));
  }

  .loading {
    padding: var(--space-md);
    text-align: center;
    color: var(--text-muted);
    font-size: var(--font-size-sm);
  }

  .warning {
    color: var(--text-muted);
    font-size: var(--font-size-sm);
    margin-bottom: var(--space-md);
  }

  .environments {
    flex: 1;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: var(--space-xs);
    padding: var(--space-xs);
  }

  .empty-state {
    padding: var(--space-xl);
    text-align: center;
    color: var(--text-muted);
  }

  .empty-state p {
    margin: var(--space-xs) 0;
  }

  .empty-state .hint {
    font-size: var(--font-size-sm);
  }
</style>
