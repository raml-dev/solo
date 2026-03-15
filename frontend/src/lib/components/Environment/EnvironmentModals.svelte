<script lang="ts">
  import Button from "flowbite-svelte/Button.svelte";
  import Modal from "flowbite-svelte/Modal.svelte";
  import ToastContainer from "$src/lib/components/base/ToastContainer.svelte";
  import { modalStack, topModalId } from "$src/lib/stores/modalStackStore";
  import { onDestroy } from "svelte";

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

  const environmentModalScope = `environment-modals-${Math.random().toString(36).slice(2)}`;
  const newEnvironmentModalId = `${environmentModalScope}-new`;
  const deleteEnvironmentModalId = `${environmentModalScope}-delete`;

  $effect(() => {
    if (showNewEnvironmentDialog) {
      modalStack.open(newEnvironmentModalId);
    } else {
      modalStack.close(newEnvironmentModalId);
    }
  });

  $effect(() => {
    if (showDeleteConfirmDialog) {
      modalStack.open(deleteEnvironmentModalId);
    } else {
      modalStack.close(deleteEnvironmentModalId);
    }
  });

  onDestroy(() => {
    modalStack.close(newEnvironmentModalId);
    modalStack.close(deleteEnvironmentModalId);
  });

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
    {#if $topModalId === newEnvironmentModalId}
      <ToastContainer />
    {/if}
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
    {#if $topModalId === deleteEnvironmentModalId}
      <ToastContainer />
    {/if}
    <p>Are you sure you want to delete "{deleteTarget}"?</p>
    <p class="warning">This action cannot be undone.</p>
    {#snippet footer()}
      <Button color="red" onclick={confirmDelete}>Delete</Button>
    {/snippet}
  </Modal>
{/if}
