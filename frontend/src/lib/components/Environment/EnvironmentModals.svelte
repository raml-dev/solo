<script lang="ts">
  import Button from "flowbite-svelte/Button.svelte";
  import Helper from "flowbite-svelte/Helper.svelte";
  import Input from "flowbite-svelte/Input.svelte";
  import Label from "flowbite-svelte/Label.svelte";
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
  <Modal
    bind:open={showNewEnvironmentDialog}
    onclose={closeNewEnvironmentDialog}
    title="New Environment"
  >
    {#if $topModalId === newEnvironmentModalId}
      <ToastContainer />
    {/if}
    <div class="space-y-2">
      <Label for="new-environment-name">Environment name</Label>
      <Input
        id="new-environment-name"
        type="text"
        bind:value={newEnvironmentName}
        placeholder="Environment name"
        onkeydown={(e) => e.key === "Enter" && handleCreateEnvironment()}
        autofocus
      />
      <Helper>Use a unique name for the environment.</Helper>
    </div>
    {#snippet footer()}
      <div class="flex w-full justify-end gap-2">
        <Button color="light" onclick={closeNewEnvironmentDialog}>Cancel</Button>
        <Button
          color="primary"
          disabled={!newEnvironmentName.trim()}
          onclick={handleCreateEnvironment}>Create</Button
        >
      </div>
    {/snippet}
  </Modal>
{/if}

{#if showDeleteConfirmDialog}
  <Modal
    bind:open={showDeleteConfirmDialog}
    onclose={closeDeleteConfirmDialog}
    title="Delete Environment"
  >
    {#if $topModalId === deleteEnvironmentModalId}
      <ToastContainer />
    {/if}
    <div class="space-y-2">
      <p>Are you sure you want to delete "{deleteTarget}"?</p>
      <p class="text-sm text-warning-700 dark:text-warning-300">This action cannot be undone.</p>
    </div>
    {#snippet footer()}
      <div class="flex w-full justify-end gap-2">
        <Button color="light" onclick={closeDeleteConfirmDialog}>Cancel</Button>
        <Button color="red" onclick={confirmDelete}>Delete</Button>
      </div>
    {/snippet}
  </Modal>
{/if}
