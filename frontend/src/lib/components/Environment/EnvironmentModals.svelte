<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import Button from "flowbite-svelte/Button.svelte";
  import Helper from "flowbite-svelte/Helper.svelte";
  import Input from "flowbite-svelte/Input.svelte";
  import Label from "flowbite-svelte/Label.svelte";
  import Modal from "flowbite-svelte/Modal.svelte";
  import ToastContainer from "$src/lib/components/base/ToastContainer.svelte";
  import { modalStack, topModalId } from "$src/lib/stores/modalStackStore.svelte";
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

  const newEnvironmentModal = modalStack.createModal("environment-modals-new");
  const deleteEnvironmentModal = modalStack.createModal("environment-modals-delete");

  $effect(() => {
    if (newEnvironmentModal.open !== showNewEnvironmentDialog) {
      newEnvironmentModal.open = showNewEnvironmentDialog;
    }
  });

  $effect(() => {
    if (deleteEnvironmentModal.open !== showDeleteConfirmDialog) {
      deleteEnvironmentModal.open = showDeleteConfirmDialog;
    }
  });

  onDestroy(() => {
    modalStack.destroyModal(newEnvironmentModal.id);
    modalStack.destroyModal(deleteEnvironmentModal.id);
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

{#if newEnvironmentModal.open}
  <Modal
    bind:open={newEnvironmentModal.open}
    onclose={closeNewEnvironmentDialog}
    title="New Environment"
  >
    {#if $topModalId === newEnvironmentModal.id}
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

{#if deleteEnvironmentModal.open}
  <Modal
    bind:open={deleteEnvironmentModal.open}
    onclose={closeDeleteConfirmDialog}
    title="Delete Environment"
  >
    {#if $topModalId === deleteEnvironmentModal.id}
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
