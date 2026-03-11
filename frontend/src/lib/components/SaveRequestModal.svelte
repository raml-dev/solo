<script lang="ts">
  import { createEventDispatcher, onMount } from "svelte";
  import { collectionStore } from "../stores/collectionStore";
  import Button from "./base/Button.svelte";
  import { slide } from "svelte/transition";

  export let show = false;
  export let requestName = "";

  const dispatch = createEventDispatcher();

  let selectedCollectionName: string | null = null;
  let showCollectionList = false;

  // If a collection is already selected in the sidebar, use it by default
  onMount(() => {
    if ($collectionStore.selectedCollectionName) {
      selectedCollectionName = $collectionStore.selectedCollectionName;
    }
  });

  function handleSave() {
    dispatch("save", {
      name: requestName,
      collection: selectedCollectionName
    });
  }

  function handleCancel() {
    dispatch("cancel");
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

<svelte:window on:keydown={handleKeyDown} />

{#if show}
  <div class="modal-overlay" role="presentation" on:click={handleOverlayClick}>
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
            <Button variant="secondary" click={() => (showCollectionList = !showCollectionList)}>
              Choose Collection
            </Button>
          </div>
        {:else}
          <div class="selected-collection-view">
            <span class="save-to-text">Saving to:</span>
            <div class="selected-collection-badge">
              <span>{selectedCollectionName}</span>
              <button
                class="change-btn"
                on:click={() => (showCollectionList = !showCollectionList)}
              >
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
                    on:click={() => handleSelectCollection(collection.name)}
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
        <Button variant="tertiary" click={handleCancel}>Cancel</Button>
      </footer>
    </div>
  </div>
{/if}

<style>
  :global(body) {
    --modal-width: 640px;
    --modal-height: 480px;
  }
  .modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    backdrop-filter: blur(4px);
  }

  .modal-content {
    background: var(--bg-primary);
    border-radius: var(--radius-xl);
    width: var(--modal-width);
    height: var(--modal-height);
    display: flex;
    flex-direction: column;
    overflow: hidden;
    box-shadow: var(--shadow-xl);
    border: 1px solid var(--border);
  }

  .modal-header {
    padding: var(--space-xl);
    border-bottom: 1px solid var(--border);
  }

  .request-name-input {
    width: 100%;
    background: transparent;
    border: none;
    outline: none;
    font-size: var(--font-size-xl);
    font-weight: var(--font-weight-semibold);
    color: var(--text);
    padding: var(--space-sm) 0;
  }

  .content-area {
    flex: 1;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    padding: var(--space-2xl);
    position: relative;
    overflow-y: auto;
  }

  .collection-drop-zone,
  .selected-collection-view {
    text-align: center;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-md);
  }

  .icon {
    color: var(--text-muted);
    margin-bottom: var(--space-md);
  }

  .drop-zone-title {
    font-size: var(--font-size-lg);
    font-weight: var(--font-weight-semibold);
    margin: 0;
  }

  .drop-zone-subtitle {
    font-size: var(--font-size-md);
    color: var(--text-muted);
    margin: 0 0 var(--space-lg) 0;
    max-width: 300px;
  }

  .selected-collection-view {
    gap: var(--space-sm);
  }

  .save-to-text {
    color: var(--text-muted);
  }

  .selected-collection-badge {
    display: flex;
    align-items: center;
    gap: var(--space-md);
    background: var(--bg-secondary);
    padding: var(--space-sm) var(--space-lg);
    border-radius: var(--radius-full);
    font-size: var(--font-size-md);
    border: 1px solid var(--border);
  }

  .change-btn {
    background: none;
    border: none;
    color: var(--primary);
    cursor: pointer;
    font-size: var(--font-size-sm);
  }

  .collection-list-container {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: 80%;
    max-height: 80%;
    background: var(--bg-secondary);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-xl);
    overflow: hidden;
    border: 1px solid var(--border);
  }

  .collection-list {
    list-style: none;
    padding: 0;
    margin: 0;
    max-height: 300px;
    overflow-y: auto;
  }

  .collection-item {
    display: block;
    width: 100%;
    text-align: left;
    padding: var(--space-md) var(--space-lg);
    background: transparent;
    border: none;
    border-bottom: 1px solid var(--border);
    color: var(--text);
    cursor: pointer;
    font-size: var(--font-size-sm);
  }
  .collection-item:last-child {
    border-bottom: none;
  }

  .collection-item:hover {
    background: var(--bg-tertiary);
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: var(--space-md);
    padding: var(--space-lg);
    border-top: 1px solid var(--border);
    background: var(--bg-secondary);
  }
</style>
