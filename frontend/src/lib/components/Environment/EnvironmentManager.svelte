<script lang="ts">
  import { environmentStore } from "../../stores/environmentStore";
  import Button from "../base/Button.svelte";
  import EnvironmentItem from "./EnvironmentItem.svelte";
  import EnvironmentModals from "./EnvironmentModals.svelte";

  let showNewEnvironmentDialog = false;
  let showDeleteConfirmDialog = false;
  let showErrorDialog = false;
  let errorMessage: string = "";
  let deleteTarget: string | null = null;
  let activeMenu: string | null = null;

  $: environments = $environmentStore.environments;
  $: selectedEnvironmentName = $environmentStore.selectedEnvironmentName;

  function showError(msg: string) {
    errorMessage = msg;
    showErrorDialog = true;
  }

  function selectEnvironment(name: string) {
    environmentStore.selectEnvironment(name);
  }

  async function handleCreateEnvironment(event: CustomEvent<string>) {
    const trimmed = event.detail.trim();
    if (!trimmed) {
      return;
    }

    const exists = environments.some((env) => env.name.toLowerCase() === trimmed.toLowerCase());
    if (exists) {
      showError(`Environment "${trimmed}" already exists.`);
      return;
    }

    try {
      await environmentStore.createEnvironment(trimmed);
      showNewEnvironmentDialog = false;
    } catch (err) {
      console.error("Error creating environment:", err);
      showError(`Error creating environment: ${err}`);
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
      console.error("Error deleting environment:", err);
      showError(`Error deleting environment: ${err}`);
    }
  }

  function toggleMenu(event: CustomEvent<string>) {
    const environmentName = event.detail;
    activeMenu = activeMenu === environmentName ? null : environmentName;
  }
</script>

<div class="environment-list">
  <div class="header">
    <div class="header-title">
      <h3>Environments</h3>
      <Button variant="primary" size="small" click={() => (showNewEnvironmentDialog = true)}>
        New
      </Button>
    </div>
  </div>

  {#if $environmentStore.loading}
    <div class="loading">Loading environments...</div>
  {/if}

  {#if $environmentStore.error}
    <div class="error">
      {$environmentStore.error}
      <button on:click={() => environmentStore.clearError()}>x</button>
    </div>
  {/if}

  <div class="environments">
    {#each environments as environment (environment.id)}
      <EnvironmentItem
        env={environment}
        selected={selectedEnvironmentName === environment.name}
        menuOpen={activeMenu === environment.name}
        on:select={(e) => selectEnvironment(e.detail)}
        on:delete={handleDeleteEnvironment}
        on:toggleMenu={toggleMenu}
        on:error={(e) => showError(e.detail)}
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
  bind:showErrorDialog
  bind:errorMessage
  bind:deleteTarget
  on:create={handleCreateEnvironment}
  on:confirmDelete={confirmDelete}
  on:closeNew={() => (showNewEnvironmentDialog = false)}
  on:closeDelete={() => (showDeleteConfirmDialog = false)}
  on:closeError={() => (showErrorDialog = false)}
/>

<style>
  .environment-list {
    display: flex;
    flex-direction: column;
    flex: 1;
    overflow: hidden;
    padding: var(--space-md);
    border-bottom: 1px solid var(--border);
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
  }
  .header-title {
    display: flex;
    justify-content: space-between;
    align-items: center;
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
