<script lang="ts">
  import Button from "flowbite-svelte/Button.svelte";
  import Helper from "flowbite-svelte/Helper.svelte";
  import Input from "flowbite-svelte/Input.svelte";
  import Label from "flowbite-svelte/Label.svelte";
  import Select from "flowbite-svelte/Select.svelte";
  import Modal from "flowbite-svelte/Modal.svelte";
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

</script>

{#if show}
  <Modal
    title="Save Request"
    bind:open={show}
    size="lg"
    onclose={handleCancel}
  >
    {#if $topModalId === saveRequestModalId}
      <ToastContainer />
    {/if}

    <div class="space-y-4">
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
    </div>

    {#snippet footer()}
      <div class="flex w-full justify-end gap-2">
        <Button color="alternative" onclick={handleCancel}>Cancel</Button>
        <Button color="primary" disabled={!selectedCollectionName} onclick={handleSave}>Save</Button>
      </div>
    {/snippet}
  </Modal>
{/if}
