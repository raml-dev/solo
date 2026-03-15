<script lang="ts">
  import Button from "flowbite-svelte/Button.svelte";
  import { collectionStore } from "$src/lib/stores/collectionStore";
  import { onMount } from "svelte";
  import { slide } from "svelte/transition";

  interface Props {
    show?: boolean;
    requestName?: string;
    onSave?: (data: { name: string; collection: string | null }) => void;
    onCancel?: () => void;
  }

  let { show = $bindable(false), requestName = $bindable(""), onSave, onCancel }: Props = $props();

  let selectedCollectionName: string | null = $state(null);
  let showCollectionList = $state(false);

  // If a collection is already selected in the sidebar, use it by default
  onMount(() => {
    if ($collectionStore.selectedCollectionName) {
      selectedCollectionName = $collectionStore.selectedCollectionName;
    }
  });

  function handleSave() {
    onSave?.({
      name: requestName,
      collection: selectedCollectionName
    });
  }

  function handleCancel() {
    onCancel?.();
  }

  function handleSelectCollection(name: string) {
    selectedCollectionName = name;
    showCollectionList = false;
    // Autosave: dispatch immediately once a collection is chosen
    handleSave();
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
  <div class="modal-overlay" role="presentation" onclick={handleOverlayClick}>
    <div class="modal-content" role="dialog" aria-modal="true">
      <header class="modal-header">
        <input
          type="text"
          bind:value={requestName}
          placeholder="Untitled Request"
          class="request-name-input"
        />
      </header>

      <div class="content-area">
        {#if !selectedCollectionName}
          <div class="collection-drop-zone">
            <svg
              width="48"
              height="48"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="1"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="icon"
            >
              <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"
              ></path>
            </svg>
            <h2 class="drop-zone-title">Save to a Collection</h2>
            <p class="drop-zone-subtitle">
              Select a collection to organize and share your request.
            </p>
            <Button color="light" onclick={() => (showCollectionList = !showCollectionList)}>
              Choose Collection
            </Button>
          </div>
        {:else}
          <div class="selected-collection-view">
            <span class="save-to-text">Saving to:</span>
            <div class="selected-collection-badge">
              <span>{selectedCollectionName}</span>
              <button class="change-btn" onclick={() => (showCollectionList = !showCollectionList)}>
                (change)
              </button>
            </div>
          </div>
        {/if}

        {#if showCollectionList}
          <div class="collection-list-container" transition:slide>
            <ul class="collection-list">
              {#each $collectionStore.collections as collection (collection.name)}
                <li>
                  <button
                    class="collection-item"
                    onclick={() => handleSelectCollection(collection.name)}
                  >
                    {collection.name}
                  </button>
                </li>
              {/each}
            </ul>
          </div>
        {/if}
      </div>

      <footer class="modal-footer">
        <Button color="alternative" onclick={handleCancel}>Cancel</Button>
      </footer>
    </div>
  </div>
{/if}
