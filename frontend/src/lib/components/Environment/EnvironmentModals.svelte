<script lang="ts">
  import Button from "../base/Button.svelte";
  import Modal from "../base/Modal.svelte";

  interface Props {
    showNewEnvironmentDialog?: boolean;
    showDeleteConfirmDialog?: boolean;
    deleteTarget: string | null;
    onCloseDelete?: () => void;
    onCloseNew?: () => void;
    onConfirmDelete?: (name: string) => void;
    onCreate?: (name: string) => void;
  }

  let {
    showNewEnvironmentDialog = $bindable(false),
    showDeleteConfirmDialog = $bindable(false),
    deleteTarget = $bindable(null),
    onCloseDelete,
    onCloseNew,
    onConfirmDelete,
    onCreate
  }: Props = $props();

  let newEnvironmentName = $state("");

  function closeNewEnvironmentDialog() {
    showNewEnvironmentDialog = false;
    newEnvironmentName = "";
    onCloseNew?.();
  }

  function handleCreateEnvironment() {
    onCreate?.(newEnvironmentName);
  }

  function closeDeleteConfirmDialog() {
    showDeleteConfirmDialog = false;
    onCloseDelete?.();
  }

  function confirmDelete() {
    onConfirmDelete?.(deleteTarget || "");
  }
</script>

{#if showNewEnvironmentDialog}
  <Modal toggleFn={closeNewEnvironmentDialog}>
    <h3>New Environment</h3>
    <!-- svelte-ignore a11y_autofocus -->
    <input
      type="text"
      bind:value={newEnvironmentName}
      placeholder="Environment name"
      onkeydown={(e) => e.key === "Enter" && handleCreateEnvironment()}
      autofocus
    />
    {#snippet additional_buttons()}
      <Button variant="primary" click={handleCreateEnvironment}>Create</Button>
    {/snippet}
  </Modal>
{/if}

{#if showDeleteConfirmDialog}
  <Modal toggleFn={closeDeleteConfirmDialog}>
    <h3>Delete Environment</h3>
    <p>Are you sure you want to delete "{deleteTarget}"?</p>
    <p class="warning">This action cannot be undone.</p>
    {#snippet additional_buttons()}
      <Button variant="danger" click={confirmDelete}>Delete</Button>
    {/snippet}
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
