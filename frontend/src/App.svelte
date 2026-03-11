<script lang="ts">
  import { onMount } from "svelte";
  import MainLayout from "./lib/components/MainLayout.svelte";
  import CollectionList from "./lib/components/CollectionList.svelte";
  import HTTPRequestBuilder from "./lib/components/RequestBuilder/HTTPRequestBuilder.svelte";
  import RequestTabBar from "./lib/components/RequestBuilder/RequestTabBar.svelte";
  import ToastContainer from "./lib/components/base/ToastContainer.svelte";
  import Console from "./lib/components/Console/Console.svelte";
  import { historyStore } from "./lib/stores/historyStore";
  import { configurationStore } from "./lib/stores/configurationStore";
  import { collectionStore } from "./lib/stores/collectionStore";
  import { environmentStore } from "./lib/stores/environmentStore";

  let consoleOpen = false;
  let consoleHeight = 260;
  const MIN_HEIGHT = 120;
  const MAX_HEIGHT = 700;
  let isResizing = false;
  let resizeStartY = 0;
  let resizeStartH = 0;

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

  onMount(async () => {
    await configurationStore.init();
    await collectionStore.loadCollections();
    await environmentStore.loadEnvironments();
  });
</script>

<ToastContainer />

<MainLayout title="yapla">
  <svelte:fragment slot="navbar-actions">
    <div class="nav-actions">
    </div>
  </svelte:fragment>

  <!-- Main area: sidebar + builder + console panel stacked -->
  <div class="app-body">
    <CollectionList />

    <div class="builder-area">
      <RequestTabBar />
      <HTTPRequestBuilder />
    </div>
  </div>

  <!-- Console panel: expands above the bottom bar -->
  <svelte:fragment slot="bottom-bar">
    {#if consoleOpen}
      <div class="console-panel" style="height: {consoleHeight}px">
        <div
          class="console-resize-handle"
          class:resizing={isResizing}
          on:mousedown={startResize}
        ></div>
        <Console />
      </div>
    {/if}

    <!-- Bottom bar content -->
    <button
      class="bottom-btn"
      class:active={consoleOpen}
      on:click={toggleConsole}
    >
      <svg width="12" height="12" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
        <rect x="1" y="1" width="14" height="14" rx="2" stroke="currentColor" stroke-width="1.4"/>
        <path d="M4 6l3 3-3 3M9 12h4" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round"/>
      </svg>
      Console
      {#if $historyStore.length > 0}
        <span class="bottom-btn-badge">{$historyStore.length}</span>
      {/if}
    </button>
  </svelte:fragment>
</MainLayout>

<style>
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
    transition: color 0.15s, background 0.15s;
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
    bottom: 28px; /* height of bottom-bar */
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
