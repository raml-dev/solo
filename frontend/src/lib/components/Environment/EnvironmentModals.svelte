<script lang="ts">
  import Button from "../base/Button.svelte";
  import Modal from "../base/Modal.svelte";
  import { createEventDispatcher } from "svelte";

  export let showNewEnvironmentDialog: boolean = false;
  export let showDeleteConfirmDialog: boolean = false;
  export let showErrorDialog: boolean = false;
  export let errorMessage: string = "";
  export let deleteTarget: string | null = null;

  let newEnvironmentName = "";

  const dispatch = createEventDispatcher<{
    closeDelete: null;
    closeError: null;
    closeNew: null;
    confirmDelete: string;
    create: string;
  }>();

  function closeNewEnvironmentDialog() {
    showNewEnvironmentDialog = false;
    newEnvironmentName = "";
    dispatch("closeNew");
  }

  function handleCreateEnvironment() {
    dispatch("create", newEnvironmentName);
  }

  function closeDeleteConfirmDialog() {
    showDeleteConfirmDialog = false;
    dispatch("closeDelete");
  }

  function confirmDelete() {
    dispatch("confirmDelete", deleteTarget);
  }
</script>

{#if showNewEnvironmentDialog}
  <Modal toggleFn={closeNewEnvironmentDialog}>
    <h3>New Environment</h3>
    <!-- svelte-ignore a11y-autofocus -->
    <input
      type="text"
      bind:value={newEnvironmentName}
      placeholder="Environment name"
      on:keydown={(e) => e.key === "Enter" && handleCreateEnvironment()}
      autofocus
    />
    <svelte:fragment slot="additional-buttons">
      <Button variant="primary" click={handleCreateEnvironment}>Create</Button>
    </svelte:fragment>
  </Modal>
{/if}

{#if showDeleteConfirmDialog}
  <Modal toggleFn={closeDeleteConfirmDialog}>
    <h3>Delete Environment</h3>
    <p>Are you sure you want to delete "{deleteTarget}"?</p>
    <p class="warning">This action cannot be undone.</p>
    <svelte:fragment slot="additional-buttons">
      <Button variant="danger" click={confirmDelete}>Delete</Button>
    </svelte:fragment>
  </Modal>
{/if}

{#if showErrorDialog}
  <Modal toggleFn={() => dispatch("closeError")}>
    <h3>Error</h3>
    <p>{errorMessage}</p>
  </Modal>
{/if}

<style>
  .warning {
    color: var(--text-muted);
    font-size: var(--font-size-sm);
    margin-bottom: var(--space-md);
  }

  input:not([type="checkbox"]) {
    width: 100%;
    padding: var(--space-sm);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    background: var(--bg-secondary);
    color: var(--text);
    font-size: var(--font-size-md);
  }

  input:not([type="checkbox"]):focus {
    outline: none;
    border-color: var(--primary);
  }
</style>
