<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import ToastContainer from "$src/lib/components/base/ToastContainer.svelte";
  import CollectionList from "$src/lib/components/Collections/CollectionList.svelte";
  import ContextMenu from "$src/lib/components/common/ContextMenu.svelte";
  import ContextMenuItem from "$src/lib/components/common/ContextMenuItem.svelte";
  import History from "$src/lib/components/History/History.svelte";
  import MainLayout from "$src/lib/components/MainLayout.svelte";
  import HTTPRequestBuilder from "$src/lib/components/RequestBuilder/HTTPRequestBuilder.svelte";
  import RequestTabBar from "$src/lib/components/RequestBuilder/RequestTabBar.svelte";
  import { collectionStore } from "$src/lib/stores/collectionStore.svelte";
  import { configurationStore } from "$src/lib/stores/configurationStore.svelte";
  import { environmentStore, environmentStoreState } from "$src/lib/stores/environmentStore.svelte";
  import { historyStore } from "$src/lib/stores/historyStore.svelte";
  import { hasOpenModals, modalStack } from "$src/lib/stores/modalStackStore.svelte";
  import { notifications } from "$src/lib/stores/notificationStore";
  import { getActiveTab, tabStore } from "$src/lib/stores/tabStore.svelte";
  import { updateStore } from "$src/lib/stores/updateStore.svelte";
  import {
    initWindowDimensions,
    saveWindowState
  } from "$src/lib/stores/windowDimensionsStore.svelte";
  import { initZoom } from "$src/lib/stores/zoomStore.svelte";
  import { flowbiteTheme } from "$src/lib/theme/flowbiteCustomTheme";
  import { ForceQuit } from "$wails/go/main/App";
  import { EventsOn } from "$wails/runtime/runtime";
  import EditOutline from "flowbite-svelte-icons/EditOutline.svelte";
  import GlobeOutline from "flowbite-svelte-icons/GlobeOutline.svelte";
  import ClockArrowOutline from "flowbite-svelte-icons/ClockArrowOutline.svelte";
  import Badge from "flowbite-svelte/Badge.svelte";
  import Button from "flowbite-svelte/Button.svelte";
  import ThemeProvider from "flowbite-svelte/ThemeProvider.svelte";
  import { onMount } from "svelte";

  let isEnvDropdownOpen = $state(false);

  let environments = $derived(environmentStoreState.environments);
  let selectedEnvironmentName = $derived(environmentStoreState.selectedEnvironmentName);

  const environmentManagerModal = modalStack.createModal("app-environments");

  let historyOpen = $state(false);
  let consoleHeight = $state(260);
  const MIN_HEIGHT = 120;
  const MAX_HEIGHT = 700;
  let isResizing = $state(false);
  let resizeStartY = 0;
  let resizeStartH = 0;

  async function initializeApp() {
    await configurationStore.init();
    await Promise.all([
      updateStore.init(),
      collectionStore.loadCollections(),
      environmentStore.loadEnvironments()
    ]).catch((err) => {
      console.error("Failed to initialize app", err);
    });
  }

  function toggleHistory() {
    historyOpen = !historyOpen;
  }

  function toggleEnvironmentManager() {
    environmentManagerModal.open = !environmentManagerModal.open;
  }

  function closeEnvDropdown() {
    isEnvDropdownOpen = false;
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

      const tab = getActiveTab();

      if (tab) {
        if (tab.requestId) {
          await tabStore.saveTab(tab.id);
        } else {
          // Trigger the Save modal in HTTPRequestBuilder for new requests
          window.dispatchEvent(new CustomEvent("solo:save-request-new"));
        }
      }
    }
  }

  onMount(() => {
    (async () => {
      await initializeApp();
      console.log("Application UI started");
    })();
    window.addEventListener("dragover", (e) => {
      e.preventDefault(); // necessary, otherwise drop won't fire
    });

    window.addEventListener("drop", (e) => {
      e.preventDefault(); // stops WebKitGTK navigation
    });
    window.addEventListener("keydown", handleKeyDown);

    EventsOn("app:request-close", async () => {
      await saveWindowState();

      ForceQuit();
    });

    // TODO any
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    EventsOn("auth:token-refreshed", (data: any) => {
      if (data.error) {
        notifications.error(
          "OAuth2 Token Refresh Failed",
          `Provider returned status ${data.status}`
        );
      } else {
        notifications.success("OAuth2 Token Refreshed", "A new access token has been acquired.");
      }

      // Log to history store (Console)
      historyStore.push({
        collectionName: null,
        requestName: "[Auth] Token Refresh",
        request: {
          method: data.method || "POST",
          url: data.url || "",
          headers: { "Content-Type": "application/x-www-form-urlencoded" },
          body: JSON.stringify(data.params || {}, null, 2)
        },
        response: data.error
          ? null
          : {
              status: data.status,
              time: 0,
              headers: { "Content-Type": "application/json" },
              body: data.body || ""
            },
        error: data.error ? `Auth request failed with status ${data.status}: ${data.body}` : null
      });
    });

    // TODO any
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    EventsOn("request:executed", (data: any) => {
      const opts = data.options || {};
      const resp = data.response;
      const currentTab = getActiveTab();

      historyStore.push({
        collectionName: currentTab?.collectionName ?? null,
        requestName: currentTab?.label ?? null,
        request: {
          method: opts.method || "GET",
          url: opts.url || "",
          headers: Object.entries(opts.headers || {}).reduce(
            (acc, [k, v]) => ({ ...acc, [k]: String(v) }),
            {} as Record<string, string>
          ),
          body: opts.body || ""
        },
        response: resp
          ? {
              status: resp.statusCode,
              time: resp.duration,
              headers: resp.headers,
              body: resp.body
            }
          : null,
        error: data.error || null
      });
    });

    const zoomCleanup = initZoom();
    const windowDimensionsCleanup = initWindowDimensions();

    return () => {
      window.removeEventListener("keydown", handleKeyDown);
      zoomCleanup();
      windowDimensionsCleanup();
    };
  });
</script>

{#snippet envDropdown(triggeredBy: string, isOpen: boolean | undefined, onClose: () => void)}
  <ContextMenu {triggeredBy} {isOpen} menuClass="z-50 w-max max-w-72 min-w-40" {onClose}>
    {#each environments as environment (environment.id)}
      <ContextMenuItem
        onclick={() => {
          void environmentStore.selectEnvironment(environment.name);
          onClose();
        }}
      >
        <div class="flex items-center gap-2">
          <div class="flex flex-1 items-center gap-2">
            <input
              type="radio"
              name="active-environment"
              checked={selectedEnvironmentName === environment.name}
              onchange={() => {
                void environmentStore.selectEnvironment(environment.name);
                onClose();
              }}
              aria-label={`Set ${environment.name} as active environment`}
              class="relative mr-2 flex h-4 w-4 shrink-0 items-center border-gray-300 bg-gray-100 text-primary-600 dark:border-gray-600 dark:bg-gray-700"
            />
            <span>{environment.name}</span>
          </div>
          <div class="flex">
            <EditOutline
              class="h-4 w-4 shrink-0 cursor-pointer"
              onclick={toggleEnvironmentManager}
            />
          </div>
        </div>
      </ContextMenuItem>
    {/each}
    <!-- <ContextMenuItem
      onclick={() => {
        onClose();
      }}
    >
      <div class="flex items-center gap-2">add env</div>
    </ContextMenuItem> -->
  </ContextMenu>
{/snippet}
<ThemeProvider theme={flowbiteTheme}>
  {#if !$hasOpenModals}
    <ToastContainer />
  {/if}

  <MainLayout>
    <!-- Main area: sidebar + builder -->
    <div class="flex min-h-0 flex-1 overflow-hidden">
      <CollectionList />
      <div class="flex min-w-0 flex-1 flex-col overflow-hidden">
        <RequestTabBar />
        <HTTPRequestBuilder />
      </div>
    </div>

    <!-- History panel + status bar -->
    {#snippet bottom_bar()}
      {#if historyOpen}
        <div
          class="flex flex-col overflow-hidden border-t border-neutral-200 dark:border-neutral-700"
          style="height: {consoleHeight}px"
        >
          <button
            type="button"
            class="h-1 shrink-0 cursor-row-resize bg-neutral-200 p-0 transition-colors hover:bg-primary-400 dark:bg-neutral-700"
            class:bg-primary-500={isResizing}
            onmousedown={startResize}
            aria-label="Resize console"
          ></button>
          <div class="min-h-0 flex-1 overflow-hidden">
            <History />
          </div>
        </div>
      {/if}

      <!-- Status bar -->
      <div
        class="flex h-[--spacing-statusbar] shrink-0 items-center justify-between border-t border-neutral-200 bg-neutral-50 px-2 dark:border-neutral-700 dark:bg-neutral-900"
      >
        <Button
          color="alternative"
          size="xs"
          onclick={toggleHistory}
          class="gap-1 border-0 bg-transparent shadow-none hover:bg-transparent focus:ring-0 active:bg-transparent dark:bg-transparent dark:hover:bg-transparent dark:active:bg-transparent
          {historyOpen
            ? 'text-primary-600 dark:text-primary-400'
            : 'text-neutral-600 dark:text-neutral-400'}"
        >
          <ClockArrowOutline class="-ml-1.5 h-4" />
          History
          {#if $historyStore.length > 0}
            <Badge
              color="primary"
              class="ml-1 flex h-4 w-4 items-center justify-center rounded-full p-0 text-[10px]"
            >
              {$historyStore.length}
            </Badge>
          {/if}
        </Button>
        <Button
          id="env-dropdown-button"
          size="xs"
          onclick={() => (isEnvDropdownOpen = true)}
          class="items-center gap-1 border-0 bg-transparent shadow-none hover:bg-transparent focus:ring-0 active:bg-transparent dark:bg-transparent dark:hover:bg-transparent dark:active:bg-transparent {selectedEnvironmentName
            ? 'text-neutral-600 dark:text-neutral-400'
            : 'text-neutral-400 dark:text-neutral-600'}"
          color="alternative"
        >
          <GlobeOutline class="h-4" />
          Active env: {selectedEnvironmentName ?? "No environment"}
          <!-- <AngleUpOutline class="w-3 shrink-0" /> -->
        </Button>

        {@render envDropdown("#env-dropdown-button", isEnvDropdownOpen, closeEnvDropdown)}
      </div>
    {/snippet}
  </MainLayout>
</ThemeProvider>
