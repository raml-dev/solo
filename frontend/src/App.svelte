<script lang="ts">
  import CollectionList from "$src/lib/components/CollectionList.svelte";
  import Console from "$src/lib/components/Console/Console.svelte";
  import MainLayout from "$src/lib/components/MainLayout.svelte";
  import HTTPRequestBuilder from "$src/lib/components/RequestBuilder/HTTPRequestBuilder.svelte";
  import RequestTabBar from "$src/lib/components/RequestBuilder/RequestTabBar.svelte";
  import Button from "flowbite-svelte/Button.svelte";
  import Modal from "flowbite-svelte/Modal.svelte";
  import ToastContainer from "$src/lib/components/base/ToastContainer.svelte";
  import { collectionStore } from "$src/lib/stores/collectionStore";
  import { configurationStore } from "$src/lib/stores/configurationStore";
  import { environmentStore } from "$src/lib/stores/environmentStore";
  import { historyStore } from "$src/lib/stores/historyStore";
  import { hasOpenModals, modalStack, topModalId } from "$src/lib/stores/modalStackStore";
  import { activeTab, tabStore } from "$src/lib/stores/tabStore";
  import { ForceQuit } from "$wails/go/main/App";
  import { EventsOn } from "$wails/runtime/runtime";
  import { onMount } from "svelte";

  let consoleOpen = $state(false);
  let consoleHeight = $state(260);
  const MIN_HEIGHT = 120;
  const MAX_HEIGHT = 700;
  let isResizing = $state(false);
  let resizeStartY = 0;
  let resizeStartH = 0;
  let showGlobalUnsavedModal = $state(false);

  const appModalScope = `app-${Math.random().toString(36).slice(2)}`;
  const globalUnsavedModalId = `${appModalScope}-unsaved`;

  $effect(() => {
    if (showGlobalUnsavedModal) {
      modalStack.open(globalUnsavedModalId);
    } else {
      modalStack.close(globalUnsavedModalId);
    }
  });

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
      modalStack.close(globalUnsavedModalId);
      window.removeEventListener("keydown", handleKeyDown);
    };
  });
</script>

{#if !$hasOpenModals}
  <ToastContainer />
{/if}

<MainLayout title="yapla">
  {#snippet navbar_actions()}
    <div class="nav-actions"></div>
  {/snippet}

  <!-- Main area: sidebar + builder + console panel stacked -->

  <div class="app-body">
    <CollectionList />
    <div class="builder-area"><RequestTabBar /><HTTPRequestBuilder /></div>
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
        <rect x="1" y="1" width="14" height="14" rx="2" stroke="currentColor" stroke-width="1.4"
        ></rect>

        <path
          d="M4 6l3 3-3 3M9 12h4"
          stroke="currentColor"
          stroke-width="1.4"
          stroke-linecap="round"
          stroke-linejoin="round"
        ></path>
      </svg>

      Console

      {#if $historyStore.length > 0}
        <span class="bottom-btn-badge">{$historyStore.length}</span>
      {/if}
    </button>
  {/snippet}
</MainLayout>

{#if showGlobalUnsavedModal}
  <Modal title="Unsaved Changes" bind:open={showGlobalUnsavedModal}>
    {#if $topModalId === globalUnsavedModalId}
      <ToastContainer />
    {/if}
    <div class="confirm-modal-body">
      <p>You have unsaved changes in some requests. Do you want to save them before quitting?</p>
      <p class="text-gray-500 dark:text-gray-400">
        If you don't save, your changes will be permanently lost.
      </p>

      <div class="confirm-modal-actions">
        <Button color="light" onclick={() => ForceQuit()}>Discard and Quit</Button>
        <div class="flex-spacer"></div>

        <Button color="light" onclick={() => (showGlobalUnsavedModal = false)}>Cancel</Button>

        <Button color="primary" onclick={handleSaveAllAndQuit}>Save All and Quit</Button>
      </div>
    </div>
  </Modal>
{/if}
