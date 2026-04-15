<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import ToastContainer from "$src/lib/components/base/ToastContainer.svelte";
  import CollectionList from "$src/lib/components/CollectionList.svelte";
  import Console from "$src/lib/components/Console/Console.svelte";
  import MainLayout from "$src/lib/components/MainLayout.svelte";
  import HTTPRequestBuilder from "$src/lib/components/RequestBuilder/HTTPRequestBuilder.svelte";
  import RequestTabBar from "$src/lib/components/RequestBuilder/RequestTabBar.svelte";
  import { collectionStore } from "$src/lib/stores/collectionStore.svelte";
  import { configurationStore } from "$src/lib/stores/configurationStore.svelte";
  import { environmentStore } from "$src/lib/stores/environmentStore.svelte";
  import { historyStore } from "$src/lib/stores/historyStore";
  import { hasOpenModals } from "$src/lib/stores/modalStackStore.svelte";
  import { notifications } from "$src/lib/stores/notificationStore";
  import { getActiveTab, tabStore } from "$src/lib/stores/tabStore.svelte";
  import { initZoom } from "$src/lib/stores/zoomStore.svelte";
  import { ForceQuit } from "$wails/go/main/App";
  import { EventsOn } from "$wails/runtime/runtime";
  import TerminalOutline from "flowbite-svelte-icons/TerminalOutline.svelte";
  import Badge from "flowbite-svelte/Badge.svelte";
  import Button from "flowbite-svelte/Button.svelte";
  import ThemeProvider from "flowbite-svelte/ThemeProvider.svelte";
  import { onMount } from "svelte";
  import { initWindowDimensions, saveWindowState } from "./lib/stores/windowDimensionsStore.svelte";

  let consoleOpen = $state(false);
  let consoleHeight = $state(260);
  const MIN_HEIGHT = 120;
  const MAX_HEIGHT = 700;
  let isResizing = $state(false);
  let resizeStartY = 0;
  let resizeStartH = 0;
  const appNameAscii = `           █
 ▄▄▄  ▄▄▄  █  ▄▄▄
▀▄▄  █   █ █ █   █
▄▄▄▀ ▀▄▄▄▀ █ ▀▄▄▄▀`;
  // const globalUnsavedModal = modalStack.createModal("app-unsaved");

  const flowbiteTheme = {
    input: {
      input: "placeholder:text-neutral-400 dark:placeholder:text-neutral-400"
    },
    textarea: {
      base: "placeholder:text-neutral-400 dark:placeholder:text-neutral-400"
    },
    search: {
      input: "placeholder:text-neutral-400 dark:placeholder:text-neutral-400"
    },
    fileupload: {
      base: "placeholder:text-neutral-400 dark:placeholder:text-neutral-400"
    },
    phoneInput: {
      input: "placeholder:text-neutral-400 dark:placeholder:text-neutral-400"
    },
    floatingLabelInput: {
      input: "placeholder:text-neutral-400 dark:placeholder:text-neutral-400"
    },
    tags: {
      input: "placeholder:text-neutral-400 dark:placeholder:text-neutral-400"
    },
    multiSelect: {
      placeholder: "text-neutral-400 dark:text-neutral-400"
    },
    dropdown: "overflow-hidden!"
  };

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
      setTimeout(() => {
        ForceQuit();
      }, 150);
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

<ThemeProvider theme={flowbiteTheme}>
  {#if !$hasOpenModals}
    <ToastContainer />
  {/if}

  <MainLayout title={appNameAscii}>
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
          <button
            type="button"
            class="h-1 shrink-0 cursor-row-resize bg-neutral-200 p-0 transition-colors hover:bg-primary-400 dark:bg-neutral-700"
            class:bg-primary-500={isResizing}
            onmousedown={startResize}
            aria-label="Resize console"
          ></button>
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
          class="border-0 shadow-none {consoleOpen
            ? 'text-primary-600 dark:text-primary-400'
            : 'text-neutral-600 dark:text-neutral-400'}"
        >
          <TerminalOutline size="xs" />
          Console
          {#if $historyStore.length > 0}
            <Badge
              color="primary"
              class="ml-1 flex h-4 w-4 items-center justify-center rounded-full p-0 text-[10px]"
            >
              {$historyStore.length}
            </Badge>
          {/if}
        </Button>
      </div>
    {/snippet}
  </MainLayout>
</ThemeProvider>
