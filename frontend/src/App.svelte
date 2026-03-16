<script lang="ts">
  import ToastContainer from "$src/lib/components/base/ToastContainer.svelte";
  import CollectionList from "$src/lib/components/CollectionList.svelte";
  import Console from "$src/lib/components/Console/Console.svelte";
  import MainLayout from "$src/lib/components/MainLayout.svelte";
  import HTTPRequestBuilder from "$src/lib/components/RequestBuilder/HTTPRequestBuilder.svelte";
  import RequestTabBar from "$src/lib/components/RequestBuilder/RequestTabBar.svelte";
  import { collectionStore } from "$src/lib/stores/collectionStore";
  import { configurationStore } from "$src/lib/stores/configurationStore";
  import { environmentStore } from "$src/lib/stores/environmentStore";
  import { historyStore } from "$src/lib/stores/historyStore";
  import { hasOpenModals, modalStack, topModalId } from "$src/lib/stores/modalStackStore";
  import { activeTab, tabStore } from "$src/lib/stores/tabStore";
  import { ForceQuit } from "$wails/go/main/App";
  import { EventsOn } from "$wails/runtime/runtime";
  import TerminalOutline from "flowbite-svelte-icons/TerminalOutline.svelte";
  import Badge from "flowbite-svelte/Badge.svelte";
  import Button from "flowbite-svelte/Button.svelte";
  import Modal from "flowbite-svelte/Modal.svelte";
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
  <!-- Main area: sidebar + builder -->
  <div class="flex min-h-0 flex-1 overflow-hidden">
    <CollectionList />
    <div class="flex min-w-0 flex-1 flex-col overflow-hidden">
      <RequestTabBar />
      <HTTPRequestBuilder />
    </div>
  </div>

  <!-- Console panel + status bar -->
  {#snippet bottom_bar()}
    {#if consoleOpen}
      <div
        class="flex flex-col overflow-hidden border-t border-neutral-200 dark:border-neutral-700"
        style="height: {consoleHeight}px"
      >
        <div
          class="h-1 shrink-0 cursor-row-resize bg-neutral-200 transition-colors hover:bg-primary-400 dark:bg-neutral-700"
          class:bg-primary-500={isResizing}
          onmousedown={startResize}
        ></div>
        <div class="min-h-0 flex-1 overflow-hidden">
          <Console />
        </div>
      </div>
    {/if}

    <!-- Status bar -->
    <div
      class="flex h-[--spacing-statusbar] shrink-0 items-center border-t border-neutral-200 bg-neutral-50 px-2 dark:border-neutral-700 dark:bg-neutral-900"
    >
      <Button
        color="light"
        size="xs"
        onclick={toggleConsole}
        class="border-0 shadow-none {consoleOpen ? 'text-primary-600 dark:text-primary-400' : 'text-neutral-600 dark:text-neutral-400'}"
      >
        <TerminalOutline size="xs" />
        Console
        {#if $historyStore.length > 0}
          <Badge color="primary">{$historyStore.length}</Badge>
        {/if}
      </Button>
    </div>
  {/snippet}
</MainLayout>

{#if showGlobalUnsavedModal}
  <Modal title="Unsaved Changes" bind:open={showGlobalUnsavedModal}>
    {#if $topModalId === globalUnsavedModalId}
      <ToastContainer />
    {/if}
    <div class="flex flex-col gap-2">
      <p>You have unsaved changes in some requests. Do you want to save them before quitting?</p>
      <p class="text-neutral-500 dark:text-neutral-400">
        If you don't save, your changes will be permanently lost.
      </p>
    </div>

    {#snippet footer()}
      <div class="flex w-full items-center gap-2">
        <Button color="red" onclick={() => ForceQuit()}>Discard and Quit</Button>
        <div class="ml-auto flex items-center gap-2">
          <Button color="light" onclick={() => (showGlobalUnsavedModal = false)}>Cancel</Button>
          <Button color="primary" onclick={handleSaveAllAndQuit}>Save All and Quit</Button>
        </div>
      </div>
    {/snippet}
  </Modal>
{/if}
