<script lang="ts">
  import Button from "flowbite-svelte/Button.svelte";
  import Helper from "flowbite-svelte/Helper.svelte";
  import Input from "flowbite-svelte/Input.svelte";
  import Label from "flowbite-svelte/Label.svelte";
  import Select from "flowbite-svelte/Select.svelte";
  import ToastContainer from "$src/lib/components/base/ToastContainer.svelte";
  import { collectionStore } from "$src/lib/stores/collectionStore";
  import { modalStack, topModalId } from "$src/lib/stores/modalStackStore";
  import { onDestroy, onMount } from "svelte";

  interface Props {
    show?: boolean;
    requestName?: string;
    onSave?: (data: { name: string; collection: string | null }) => void;
    onCancel?: () => void;
  }

  let { show = $bindable(false), requestName = $bindable(""), onSave, onCancel }: Props = $props();

  let selectedCollectionName = $state("");
  const collectionOptions = $derived(
    $collectionStore.collections.map((collection) => ({
      value: collection.name,
      name: collection.name
    }))
  );

  const saveRequestModalId = `save-request-${Math.random().toString(36).slice(2)}`;

  $effect(() => {
    if (show) {
      modalStack.open(saveRequestModalId);
    } else {
      modalStack.close(saveRequestModalId);
    }
  });

  // If a collection is already selected in the sidebar, use it by default
  onMount(() => {
    if ($collectionStore.selectedCollectionName) {
      selectedCollectionName = $collectionStore.selectedCollectionName;
    }
  });

  onDestroy(() => {
    modalStack.close(saveRequestModalId);
  });

  function handleSave() {
    onSave?.({
      name: requestName,
      collection: selectedCollectionName || null
    });
  }

  function handleCancel() {
    onCancel?.();
  }

  function handleCollectionChange() {
    if (selectedCollectionName) {
      // Preserve existing behavior: autosave when a collection is selected.
      handleSave();
    }
  }

  function handleOverlayClick(event: MouseEvent) {
    if (event.target === event.currentTarget) {
      handleCancel();
    }
  }

  function handleKeyDown(event: KeyboardEvent) {
    if (event.key === "Escape") {
      handleCancel();
    }
  }
</script>

<svelte:window onkeydown={handleKeyDown} />

{#if show}
  <div
    class="fixed inset-0 z-[70] flex items-center justify-center bg-black/50 p-4"
    role="presentation"
    onclick={handleOverlayClick}
  >
    <div
      class="w-full max-w-lg rounded-lg border border-gray-200 bg-white p-5 shadow-xl dark:border-gray-700 dark:bg-gray-800"
      role="dialog"
      aria-modal="true"
      aria-label="Save request"
    >
      {#if $topModalId === saveRequestModalId}
        <ToastContainer />
      {/if}

      <div class="space-y-4">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">Save Request</h2>

        <div class="space-y-2">
          <Label for="request-name">Request name</Label>
          <Input
            id="request-name"
            type="text"
            bind:value={requestName}
            placeholder="Untitled Request"
          />
        </div>

        <div class="space-y-2">
          <Label for="request-collection">Collection</Label>
          <Select
            id="request-collection"
            bind:value={selectedCollectionName}
            items={collectionOptions}
            placeholder="Select a collection"
            size="sm"
            onchange={handleCollectionChange}
          />
          <Helper>Select the target collection for this request.</Helper>
        </div>

        <div class="flex justify-end gap-2 pt-2">
          <Button color="alternative" onclick={handleCancel}>Cancel</Button>
          <Button color="primary" disabled={!selectedCollectionName} onclick={handleSave}>Save</Button>
        </div>
      </div>
    </div>
  </div>
{/if}
