<script lang="ts">
  import { onMount } from "svelte";
  import MainLayout from "./lib/components/MainLayout.svelte";
  import CollectionList from "./lib/components/CollectionList.svelte";
  import HTTPRequestBuilder from "./lib/components/RequestBuilder/HTTPRequestBuilder.svelte";
  import RequestTabBar from "./lib/components/RequestBuilder/RequestTabBar.svelte";
  import ToastContainer from "./lib/components/base/ToastContainer.svelte";
  import Console from "./lib/components/Console/Console.svelte";
  import Modal from "./lib/components/base/Modal.svelte";
  import Button from "./lib/components/base/Button.svelte";
  import { historyStore } from "./lib/stores/historyStore";
  import { configurationStore } from "./lib/stores/configurationStore";
  import { collectionStore } from "./lib/stores/collectionStore";
  import { environmentStore } from "./lib/stores/environmentStore";
  import { tabStore, activeTab } from "./lib/stores/tabStore";
  import { EventsOn } from "../wailsjs/runtime/runtime";
  import { ForceQuit } from "../wailsjs/go/main/App";

  let consoleOpen = $state(false);
  let consoleHeight = $state(260);
  const MIN_HEIGHT = 120;
  const MAX_HEIGHT = 700;
  let isResizing = $state(false);
  let resizeStartY = 0;
  let resizeStartH = 0;

  let showGlobalUnsavedModal = $state(false);

  async function initializeApp() {
    await Promise.all([
      configurationStore.init(),
      collectionStore.loadCollections(),
      environmentStore.loadEnvironments()
    ]).catch((err) => {
      console.error("Failed to initialize app", err);
    });
  }

  function toggleConsole() {
    consoleOpen = !consoleOpen;
  }

  function startResize(e: MouseEvent) {
    isResizing = true;
    resizeStartY = e.clientY;
    resizeStartH = consoleHeight;
    window.addEventListener("mousemove", onMouseMove);
    window.addEventListener("mouseup", stopResize);
  }

  function onMouseMove(e: MouseEvent) {
    const delta = resizeStartY - e.clientY;
    consoleHeight = Math.min(MAX_HEIGHT, Math.max(MIN_HEIGHT, resizeStartH + delta));
  }

  function stopResize() {
    isResizing = false;
    window.removeEventListener("mousemove", onMouseMove);
    window.removeEventListener("mouseup", stopResize);
  }

  async function handleKeyDown(e: KeyboardEvent) {
    // Ctrl+S or Cmd+S
    if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === "s") {
      e.preventDefault();
      const tab = $activeTab;
      if (tab) {
        if (tab.requestId) {
          await tabStore.saveTab(tab.id);
        } else {
          // Trigger the Save modal in HTTPRequestBuilder for new requests
          window.dispatchEvent(new CustomEvent("yapla:save-request-new"));
        }
      }
    }
  }

  async function handleSaveAllAndQuit() {
    const dirtyTabs = $tabStore.tabs.filter((t) => t.isDirty && t.requestId);
    for (const tab of dirtyTabs) {
      try {
        await tabStore.saveTab(tab.id);
      } catch (err) {
        console.error("Failed to save tab", tab.label, err);
      }
    }
    await ForceQuit();
  }

  onMount(() => {
    (async () => {
      await initializeApp();
    })();

    window.addEventListener("keydown", handleKeyDown);

    EventsOn("app:request-close", () => {
      const dirtyTabs = $tabStore.tabs.filter((t) => t.isDirty);
      if (dirtyTabs.length > 0) {
        showGlobalUnsavedModal = true;
      } else {
        ForceQuit();
      }
    });

    return () => {
      window.removeEventListener("keydown", handleKeyDown);
    };
  });
</script>

<ToastContainer />

<MainLayout title="yapla">
  {#snippet navbar_actions()}
    <div class="nav-actions"></div>
  {/snippet}

  <!-- Main area: sidebar + builder + console panel stacked -->
  <div class="app-body">
    <CollectionList />

    <div class="builder-area">
      <RequestTabBar />
      <HTTPRequestBuilder />
    </div>
  </div>

  <!-- Console panel: expands above the bottom bar -->

  {#snippet bottom_bar()}
    {#if consoleOpen}
      <div class="console-panel" style="height: {consoleHeight}px">
        <div
          class="console-resize-handle"
          class:resizing={isResizing}
          onmousedown={startResize}
        ></div>
        <Console />
      </div>
    {/if}

    <!-- Bottom bar content -->
    <button class="bottom-btn" class:active={consoleOpen} onclick={toggleConsole}>
      <svg
        width="12"
        height="12"
        viewBox="0 0 16 16"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
      >
        <rect x="1" y="1" width="14" height="14" rx="2" stroke="currentColor" stroke-width="1.4" />
        <path
          d="M4 6l3 3-3 3M9 12h4"
          stroke="currentColor"
          stroke-width="1.4"
          stroke-linecap="round"
          stroke-linejoin="round"
        />
      </svg>
      Console
      {#if $historyStore.length > 0}
        <span class="bottom-btn-badge">{$historyStore.length}</span>
      {/if}
    </button>
  {/snippet}
</MainLayout>

{#if showGlobalUnsavedModal}
  <Modal title="Unsaved Changes" toggleFn={() => (showGlobalUnsavedModal = false)} size="default">
    <div class="confirm-modal-body">
      <p>You have unsaved changes in some requests. Do you want to save them before quitting?</p>
      <p class="text-muted">If you don't save, your changes will be permanently lost.</p>

      <div class="confirm-modal-actions">
        <Button variant="secondary" click={() => ForceQuit()}>Discard and Quit</Button>
        <div class="flex-spacer"></div>
        <Button variant="secondary" click={() => (showGlobalUnsavedModal = false)}>Cancel</Button>
        <Button variant="primary" click={handleSaveAllAndQuit}>Save All and Quit</Button>
      </div>
    </div>
  </Modal>
{/if}

<style>
  .confirm-modal-body {
    padding: var(--space-md);
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
  }

  .confirm-modal-body p {
    margin: 0;
    font-size: var(--font-size-sm);
  }

  .text-muted {
    color: var(--text-muted);
  }

  .confirm-modal-actions {
    display: flex;
    gap: var(--space-sm);
    margin-top: var(--space-md);
  }

  .flex-spacer {
    flex: 1;
  }

  :global(body) {
    margin: 0;
    padding: 0;
  }

  .nav-actions {
    display: flex;
    gap: var(--space-sm);
    align-items: center;
  }

  .app-body {
    display: flex;
    flex: 1;
    min-width: 0;
    overflow: hidden;
  }

  .builder-area {
    display: flex;
    flex-direction: column;
    flex: 1;
    min-width: 0;
    overflow: hidden;
  }

  /* Bottom bar button */
  .bottom-btn {
    display: flex;
    align-items: center;
    gap: 5px;
    background: none;
    border: none;
    cursor: pointer;
    color: var(--text-muted);
    font-size: 0.72rem;
    padding: 0 var(--space-sm);
    height: 100%;
    border-right: 1px solid var(--border);
    transition:
      color 0.15s,
      background 0.15s;
  }
  .bottom-btn:hover {
    color: var(--text);
    background: var(--bg-secondary);
  }
  .bottom-btn.active {
    color: var(--primary);
  }

  .bottom-btn-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    background: var(--primary);
    color: var(--bg-primary);
    border-radius: 10px;
    font-size: 0.6rem;
    font-weight: var(--font-weight-semibold);
    padding: 0 4px;
    min-width: 14px;
    height: 14px;
  }

  /* Console panel floats above the bottom bar */
  .console-panel {
    position: absolute;
    bottom: 28px; /* height of bottom_bar */
    left: 0;
    right: 0;
    border-top: 1px solid var(--border);
    background: var(--bg-primary);
    overflow: hidden;
    z-index: 100;
  }

  .console-resize-handle {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 4px;
    cursor: ns-resize;
    z-index: 10;
  }
  .console-resize-handle:hover,
  .console-resize-handle.resizing {
    background: var(--primary);
    opacity: 0.4;
  }
</style>
