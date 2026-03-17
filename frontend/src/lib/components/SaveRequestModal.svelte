<script lang="ts">
  import Button from "flowbite-svelte/Button.svelte";
  import Helper from "flowbite-svelte/Helper.svelte";
  import Input from "flowbite-svelte/Input.svelte";
  import Label from "flowbite-svelte/Label.svelte";
  import Select from "flowbite-svelte/Select.svelte";
  import Modal from "flowbite-svelte/Modal.svelte";
  import ToastContainer from "$src/lib/components/base/ToastContainer.svelte";
  import { notifications } from "$src/lib/stores/notificationStore";
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
  let creatingNew = $state(false);
  let newCollectionName = $state("");
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

  async function handleSave() {
    if (creatingNew) {
      const trimmed = newCollectionName.trim();
      if (!trimmed) return;
      try {
        await collectionStore.createCollection(trimmed);
        selectedCollectionName = trimmed;
        creatingNew = false;
        newCollectionName = "";
      } catch (err) {
        notifications.error("Failed to create collection", String(err));
        return;
      }
    }
    onSave?.({
      name: requestName,
      collection: selectedCollectionName || null
    });
  }

  function handleCancel() {
    onCancel?.();
  }
</script>

{#if show}
  <Modal title="Save Request" bind:open={show} size="lg" onclose={handleCancel}>
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
        {#if creatingNew}
          <div class="flex gap-2">
            <Input
              bind:value={newCollectionName}
              placeholder="New collection name"
              class="flex-1"
              onkeydown={(e) => e.key === "Enter" && handleSave()}
            />
            <Button
              color="alternative"
              size="sm"
              onclick={() => {
                creatingNew = false;
                newCollectionName = "";
              }}
            >
              Cancel
            </Button>
          </div>
        {:else}
          <div class="flex gap-2">
            <Select
              id="request-collection"
              bind:value={selectedCollectionName}
              items={collectionOptions}
              placeholder="Select a collection"
              size="sm"
              class="flex-1"
            />
            <Button color="alternative" size="sm" onclick={() => (creatingNew = true)}>New…</Button>
          </div>
        {/if}
        <Helper>Select the target collection for this request.</Helper>
      </div>
    </div>

    {#snippet footer()}
      <div class="flex w-full justify-end gap-2">
        <Button color="alternative" onclick={handleCancel}>Cancel</Button>
        <Button
          color="primary"
          disabled={creatingNew ? !newCollectionName.trim() : !selectedCollectionName}
          onclick={handleSave}>Save</Button
        >
      </div>
    {/snippet}
  </Modal>
{/if}
