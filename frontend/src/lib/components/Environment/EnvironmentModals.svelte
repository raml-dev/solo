<script lang="ts">
  import Button from "flowbite-svelte/Button.svelte";
  import Modal from "flowbite-svelte/Modal.svelte";

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
  <Modal bind:open={showNewEnvironmentDialog} onclose={closeNewEnvironmentDialog} title="New Environment">
    <!-- svelte-ignore a11y_autofocus -->
    <input
      type="text"
      bind:value={newEnvironmentName}
      placeholder="Environment name"
      onkeydown={(e) => e.key === "Enter" && handleCreateEnvironment()}
      autofocus
    />
    {#snippet footer()}
      <Button color="primary" onclick={handleCreateEnvironment}>Create</Button>
    {/snippet}
  </Modal>
{/if}

{#if showDeleteConfirmDialog}
  <Modal bind:open={showDeleteConfirmDialog} onclose={closeDeleteConfirmDialog} title="Delete Environment">
    <p>Are you sure you want to delete "{deleteTarget}"?</p>
    <p class="warning">This action cannot be undone.</p>
    {#snippet footer()}
      <Button color="red" onclick={confirmDelete}>Delete</Button>
    {/snippet}
  </Modal>
{/if}
