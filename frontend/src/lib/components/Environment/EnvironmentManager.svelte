<script lang="ts">
  import { environmentStore } from "../../stores/environmentStore";
  import { notifications } from "../../stores/notificationStore";
  import Button from "../base/Button.svelte";
  import EnvironmentItem from "./EnvironmentItem.svelte";
  import EnvironmentModals from "./EnvironmentModals.svelte";
  import Modal from "../base/Modal.svelte";
  import Tabs from "../base/Tabs.svelte";
  import Tab from "../base/Tab.svelte";
  import DropZone from "../base/DropZone.svelte";
  import {
    ImportPostmanEnvironment,
    ImportBrunoEnvironment,
    SelectFile
  } from "../../../../wailsjs/go/main/App";

  let showNewEnvironmentDialog = false;
  let showDeleteConfirmDialog = false;
  let deleteTarget: string | null = null;
  let activeMenu: string | null = null;
  let showImportSelector = false;
  let importActiveTab = "postman";
  let showOverwriteConfirmDialog = false;
  let pendingImport: { format: "postman" | "bruno"; path: string } | null = null;
  let overwriteTargetName: string | null = null;

  $: environments = $environmentStore.environments;

  function selectEnvironment(name: string) {
    environmentStore.selectEnvironment(name);
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
    } catch (err) {
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
      showDeleteConfirmDialog = false;
      deleteTarget = null;
    } catch (err) {
      // error already shown by store
    }
  }

  function toggleMenu(event: CustomEvent<string>) {
    const environmentName = event.detail;
    activeMenu = activeMenu === environmentName ? null : environmentName;
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
    const filePath =
      path ?? (await SelectFile("Select Bruno Environment", "*.bru", "Bruno Files"));
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

<div class="environment-list">
  <div class="header">
    <div class="header-title">
      <h3>Environments</h3>
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
        on:select={(e) => selectEnvironment(e.detail)}
        on:delete={handleDeleteEnvironment}
        on:toggleMenu={toggleMenu}
      />
    {/each}
  </div>
</div>

{#if environments.length === 0 && !$environmentStore.loading}
  <div class="empty-state">
    <p>No environments yet</p>
    <p class="hint">Create your first environment to get started</p>
  </div>
{/if}

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
            <svelte:fragment slot="icon">
              <svg width="44" height="44" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round">
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                <polyline points="17 8 12 3 7 8"/>
                <line x1="12" y1="3" x2="12" y2="15"/>
              </svg>
            </svelte:fragment>
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
            <svelte:fragment slot="icon">
              <svg width="44" height="44" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round">
                <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>
                <polyline points="9 22 9 12 15 12 15 22"/>
              </svg>
            </svelte:fragment>
          </DropZone>
        </Tab>
      </Tabs>
    </div>

    <svelte:fragment slot="additional-buttons">
      {#if importActiveTab === "postman"}
        <Button variant="primary" click={() => handleSelectImportFormat("postman")}>Select file…</Button>
      {:else}
        <Button variant="primary" click={() => handleSelectImportFormat("bruno")}>Select folder…</Button>
      {/if}
    </svelte:fragment>
  </Modal>
{/if}

{#if showOverwriteConfirmDialog}
  <Modal toggleFn={closeOverwriteConfirmDialog} size="wide">
    <h3>Overwrite environment?</h3>
    <p>Environment "{overwriteTargetName}" already exists.</p>
    <p class="warning">Do you want to overwrite it?</p>
    <svelte:fragment slot="additional-buttons">
      <Button variant="danger" click={confirmOverwrite}>Overwrite</Button>
    </svelte:fragment>
  </Modal>
{/if}

<style>
  .environment-list {
    display: flex;
    flex-direction: column;
    flex: 1;
    overflow: hidden;
    padding: var(--space-md);
    border-bottom: 1px solid var(--border);
    gap: var(--space-sm);
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
  }

  .import-modal-body {
    margin: calc(-1 * var(--space-lg));
  }

  .header h3 {
    margin: 0;
    font-size: var(--font-size-lg);
    font-weight: var(--font-weight-semibold);
  }

  .loading {
    padding: var(--space-md);
    text-align: center;
    color: var(--text-muted);
  }

  .error {
    margin: var(--space-md);
    padding: var(--space-sm);
    background: var(--status-danger-bg);
    color: var(--status-danger-text);
    border-radius: var(--radius-md);
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: var(--font-size-sm);
  }

  .error button {
    background: none;
    border: none;
    color: inherit;
    font-size: var(--font-size-lg);
    cursor: pointer;
    padding: 0 var(--space-xs);
  }

  .success {
    margin: var(--space-md);
    padding: var(--space-sm);
    background: var(--status-success-bg);
    color: var(--status-success-text);
    border-radius: var(--radius-md);
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
    padding: var(--space-sm);
    display: flex;
    flex-direction: column;
    gap: var(--space-xs);
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
