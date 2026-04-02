<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: GPL-3.0-only
-->

<script lang="ts">
  import DropZone from "$src/lib/components/base/DropZone.svelte";
  import ToastContainer from "$src/lib/components/base/ToastContainer.svelte";
  import FeedbackEmptyState from "$src/lib/components/common/FeedbackEmptyState.svelte";
  import GitImportView from "$src/lib/components/GitImportView.svelte";
  import GitStatusPanel from "$src/lib/components/GitStatusPanel.svelte";
  import CodeMirrorEditor from "$src/lib/components/RequestBuilder/CodeMirrorEditor.svelte";
  import { collectionStoreState, collectionStore } from "$src/lib/stores/collectionStore.svelte";
  import { modalStack, topModalId } from "$src/lib/stores/modalStackStore";
  import { notifications } from "$src/lib/stores/notificationStore";
  import { getActiveTab, tabStore, tabStoreState } from "$src/lib/stores/tabStore.svelte";
  import { getMethodBadgeClass } from "$src/lib/utils/http";
  import {
    ExportCollection,
    GetGitCollectionStatus,
    GitAbortRebase,
    GitDiscardChanges,
    GitKeepOurs,
    GitKeepTheirs,
    ImportBrunoCollection,
    ImportCurlRequest,
    ImportOpenAPICollection,
    ImportPostmanCollection,
    ImportSoloCollection,
    OpenCollectionInTerminal,
    SelectDirectory,
    SelectFile,
    SyncGitCollection
  } from "$wails/go/main/App";
  import { collection } from "$wails/go/models";
  import AngleDownOutline from "flowbite-svelte-icons/AngleDownOutline.svelte";
  import AngleRightOutline from "flowbite-svelte-icons/AngleRightOutline.svelte";
  import DotsHorizontalOutline from "flowbite-svelte-icons/DotsHorizontalOutline.svelte";
  import PlusOutline from "flowbite-svelte-icons/PlusOutline.svelte";
  import Button from "flowbite-svelte/Button.svelte";
  import Dropdown from "flowbite-svelte/Dropdown.svelte";
  import DropdownDivider from "flowbite-svelte/DropdownDivider.svelte";
  import DropdownItem from "flowbite-svelte/DropdownItem.svelte";
  import Input from "flowbite-svelte/Input.svelte";
  import Label from "flowbite-svelte/Label.svelte";
  import Modal from "flowbite-svelte/Modal.svelte";
  import Select from "flowbite-svelte/Select.svelte";
  import TabItem from "flowbite-svelte/TabItem.svelte";
  import Tabs from "flowbite-svelte/Tabs.svelte";
  import { onDestroy, onMount, tick } from "svelte";
  import { SvelteSet } from "svelte/reactivity";

  interface Props {
    onRequestSelect?: (requestId: string) => void;
  }

  let { onRequestSelect = () => {} }: Props = $props();

  let editingRequestNameInputEl: HTMLInputElement | undefined = $state();

  const collectionModalScope = `collections-${Math.random().toString(36).slice(2)}`;
  const newCollectionModalId = `${collectionModalScope}-new`;
  const renameCollectionModalId = `${collectionModalScope}-rename`;
  const deleteCollectionModalId = `${collectionModalScope}-delete-collection`;
  const deleteRequestModalId = `${collectionModalScope}-delete-request`;
  const importCollectionModalId = `${collectionModalScope}-import`;
  const soloCollectionOverwriteModalId = `${collectionModalScope}-solo-overwrite`;

  const showNewCollectionDialog = modalStack.binding(newCollectionModalId);
  const showRenameCollectionDialog = modalStack.binding(renameCollectionModalId);
  const showDeleteConfirmDialog = modalStack.binding(deleteCollectionModalId);
  const showDeleteRequestConfirmDialog = modalStack.binding(deleteRequestModalId);
  const showImportSelector = modalStack.binding(importCollectionModalId);
  const showSoloCollectionOverwriteDialog = modalStack.binding(soloCollectionOverwriteModalId);

  let soloCollectionOverwriteName: string | null = $state(null);
  let pendingSoloCollectionPath: string | null = null;

  let importActiveTab = $state("postman");
  let gitImportActionState: { loading: boolean; disabled: boolean; submit: () => void } | null =
    $state(null);

  let newCollectionName = $state("");
  let renameCollectionName = $state("");
  let renameTarget: string | null = null;
  let deleteTarget: string | null = $state(null);
  let deleteRequestTarget: string | null = null;
  let deleteRequestCollectionName: string | null = null;
  let expandedCollections = new SvelteSet<string>();
  let searchQuery = $state("");
  let isCollapsed = $state(false);
  let gitStatusCollectionId: string | null = $state(null);
  let gitStatusCollectionName: string | null = $state(null);
  let syncingCollections: Set<string> = $state(new Set());
  let editingRequestId: string | null = $state(null);
  let editingRequestCollectionName: string | null = $state(null);
  let editingRequestName = $state("");

  // cURL import state
  let curlInput = $state("");
  let curlTargetCollection = $state("");
  let curlNewCollectionName = $state("");
  let curlCreatingNew = $state(false);

  // Pre-select the active collection when the import modal opens
  $effect(() => {
    if ($showImportSelector && collectionStoreState.selectedCollectionName) {
      curlTargetCollection = collectionStoreState.selectedCollectionName;
    }
  });

  let sidebarWidth = $state(280); // Default width
  let isResizing = false;

  function startResize(e: MouseEvent) {
    e.preventDefault();
    isResizing = true;
    window.addEventListener("mousemove", handleResize);
    window.addEventListener("mouseup", stopResize);
    document.body.style.userSelect = "none";
    document.body.style.cursor = "col-resize";
  }

  function handleResize(e: MouseEvent) {
    if (!isResizing) return;
    const newWidth = e.clientX;
    const minWidth = 200;
    const maxWidth = 600;
    sidebarWidth = Math.max(minWidth, Math.min(newWidth, maxWidth));
  }

  function stopResize() {
    isResizing = false;
    window.removeEventListener("mousemove", handleResize);
    window.removeEventListener("mouseup", stopResize);
    document.body.style.userSelect = "";
    document.body.style.cursor = "";
    localStorage.setItem("sidebar_width", String(sidebarWidth));
  }

  function normalize(value: string | undefined | null): string {
    return (value || "").toLowerCase();
  }

  function requestMatches(request: collection.Request, query: string): boolean {
    if (!query) return true;
    return normalize(request.name).includes(query) || normalize(request.url).includes(query);
  }

  function collectionMatches(collection: collection.Collection, query: string): boolean {
    if (!query) return true;
    return normalize(collection.name).includes(query);
  }

  function getVisibleRequests(
    collection: collection.Collection,
    query: string
  ): collection.Request[] {
    const requests = collection.requests || [];
    if (!query) return requests;
    return requests.filter((request) => requestMatches(request, query));
  }

  function shouldShowCollection(collection: collection.Collection, query: string): boolean {
    if (!query) return true;
    if (collectionMatches(collection, query)) return true;
    return getVisibleRequests(collection, query).length > 0;
  }

  function isExpanded(collectionName: string): boolean {
    return isSearching || expandedCollections.has(collectionName);
  }

  function toggleCollection(collectionName: string) {
    if (isSearching) return;
    if (expandedCollections.has(collectionName)) {
      expandedCollections.delete(collectionName);
    } else {
      expandedCollections.add(collectionName);
    }
  }

  function selectCollection(name: string) {
    collectionStore.selectCollection(name);
  }

  function selectRequest(req: collection.Request, collectionName: string) {
    if (!req?.id) {
      notifications.warning("Unable to open request: missing request id");
      return;
    }

    const headers = req.headers
      ? Object.entries(req.headers).map(([key, value], i) => ({
          id: `header-${i}`,
          key,
          value: String(value),
          enabled: true
        }))
      : [];

    tabStore.openTab(req.id, collectionName, {
      label: req.name || "Request",
      verb: req.verb || "GET",
      url: req.url || "",
      body: req.body || "",
      bodyFormat: "json",
      headers,
      auth: req.auth,
      settings: req.settings || {},
      preRequestScript: req.preRequestScript || "",
      postResponseScript: req.postResponseScript || ""
    });

    onRequestSelect(req.id);
  }

  function isEditingRequest(requestId: string, collectionName: string): boolean {
    return editingRequestId === requestId && editingRequestCollectionName === collectionName;
  }

  async function startRequestRename(e: Event, req: collection.Request, collectionName: string) {
    e.stopPropagation();
    editingRequestId = req.id;
    editingRequestCollectionName = collectionName;
    editingRequestName = req.name || "";

    await tick();
    editingRequestNameInputEl?.focus();
    editingRequestNameInputEl?.select();
  }

  function cancelRequestRename() {
    editingRequestId = null;
    editingRequestCollectionName = null;
    editingRequestName = "";
  }

  async function commitRequestRename(req: collection.Request, collectionName: string) {
    if (!isEditingRequest(req.id, collectionName)) return;

    const nextName = editingRequestName.trim();
    if (!nextName || nextName === req.name) {
      cancelRequestRename();
      return;
    }

    try {
      await collectionStore.updateRequest(
        collectionName,
        collection.Request.createFrom({
          ...req,
          name: nextName,
          lastUpdateTimestamp: new Date().toISOString()
        })
      );
      tabStore.renameTabsByRequestId(req.id, nextName);
    } catch {
      // error already shown by store
    } finally {
      cancelRequestRename();
    }
  }

  function openRenameCollection(collectionName: string) {
    renameTarget = collectionName;
    renameCollectionName = collectionName;
    showRenameCollectionDialog.set(true);
  }

  function closeNewCollectionDialog() {
    showNewCollectionDialog.set(false);
    newCollectionName = "";
  }

  function closeRenameDialog() {
    showRenameCollectionDialog.set(false);
    renameTarget = null;
    renameCollectionName = "";
  }

  async function handleCreateCollection() {
    const trimmed = newCollectionName.trim();
    if (!trimmed) return;

    const exists = collectionStoreState.collections.some(
      (c) => c.name.toLowerCase() === trimmed.toLowerCase()
    );
    if (exists) {
      notifications.warning(`Collection "${trimmed}" already exists`);
      return;
    }

    try {
      await collectionStore.createCollection(trimmed);
      closeNewCollectionDialog();
    } catch {
      // error already shown by store
    }
  }

  async function handleRenameCollection() {
    if (!renameTarget) return;
    const trimmed = renameCollectionName.trim();
    if (!trimmed || trimmed === renameTarget) {
      closeRenameDialog();
      return;
    }

    const exists = collectionStoreState.collections.some(
      (c) => c.name.toLowerCase() === trimmed.toLowerCase()
    );
    if (exists) {
      notifications.warning(`Collection "${trimmed}" already exists`);
      return;
    }

    try {
      await collectionStore.renameCollection(renameTarget, trimmed);
      if (expandedCollections.has(renameTarget)) {
        expandedCollections.delete(renameTarget);
        expandedCollections.add(trimmed);
      }
      closeRenameDialog();
    } catch {
      // error already shown by store
    }
  }

  function handleDeleteCollection(collectionName: string) {
    deleteTarget = collectionName;
    showDeleteConfirmDialog.set(true);
  }

  function closeDeleteConfirmDialog() {
    showDeleteConfirmDialog.set(false);
    deleteTarget = null;
  }

  async function confirmDelete() {
    if (!deleteTarget) return;

    try {
      await collectionStore.deleteCollection(deleteTarget);
      expandedCollections.delete(deleteTarget);
      closeDeleteConfirmDialog();
    } catch (err) {
      console.error("Error deleting collection:", err);
    }
  }

  async function handleAddRequest(e: Event, collectionName: string) {
    e.stopPropagation();

    try {
      const newReq = await collectionStore.addRequest(collectionName, {
        name: "New Request",
        url: "",
        verb: "GET"
      });

      expandedCollections.add(collectionName);

      if (newReq?.id) {
        tabStore.openTab(newReq.id, collectionName, {
          label: "New Request",
          verb: "GET",
          url: "",
          body: "",
          bodyFormat: "json",
          headers: [],
          auth: newReq.auth,
          settings: {}
        });
      }
    } catch {
      // error already shown by store
    }
  }

  async function handleDeleteRequest(collectionName: string, requestId: string) {
    deleteRequestCollectionName = collectionName;
    deleteRequestTarget = requestId;
    showDeleteRequestConfirmDialog.set(true);
  }

  async function confirmDeleteRequest() {
    if (!deleteRequestTarget || !deleteRequestCollectionName) return;

    try {
      await collectionStore.removeRequest(deleteRequestCollectionName, deleteRequestTarget);
      closeDeleteRequestConfirmDialog();
    } catch {
      // error already shown by store
    }
  }

  function closeDeleteRequestConfirmDialog() {
    showDeleteRequestConfirmDialog.set(false);
    deleteRequestTarget = null;
    deleteRequestCollectionName = null;
  }

  function toggleCollapse() {
    isCollapsed = !isCollapsed;
    localStorage.setItem("sidebar_collapsed", String(isCollapsed));
  }

  async function handleImportPostman() {
    try {
      const filePath = await SelectFile("Select Postman Collection", "*.json", "JSON Files");
      if (!filePath) return;
      await ImportPostmanCollection(filePath);
      await collectionStore.loadCollections();
      notifications.success("Postman collection imported");
    } catch (err) {
      notifications.error("Failed to import Postman collection", String(err));
    }
  }

  async function handleImportBruno() {
    try {
      const dirPath = await SelectDirectory("Select Bruno Collection Folder");
      if (!dirPath) return;
      await ImportBrunoCollection(dirPath);
      await collectionStore.loadCollections();
      notifications.success("Bruno collection imported");
    } catch (err) {
      notifications.error("Failed to import Bruno collection", String(err));
    }
  }

  async function handleImportOpenAPI() {
    try {
      const filePath = await SelectFile(
        "Select OpenAPI / Swagger Document",
        "*.json;*.yaml;*.yml",
        "OpenAPI / Swagger Files"
      );
      if (!filePath) return;
      const warnings = await ImportOpenAPICollection(filePath);
      await collectionStore.loadCollections();
      for (const w of warnings) {
        notifications.warning(w);
      }
      notifications.success("OpenAPI collection imported");
    } catch (err) {
      notifications.error("Failed to import OpenAPI collection", String(err));
    }
  }

  async function handleImportCurl() {
    if (!curlInput.trim()) return;

    if (curlCreatingNew) {
      const name = curlNewCollectionName.trim();
      if (!name) {
        notifications.error("Collection name is required");
        return;
      }
      try {
        await collectionStore.createCollection(name);
        curlTargetCollection = name;
        curlCreatingNew = false;
        curlNewCollectionName = "";
      } catch {
        return; // createCollection already shows an error toast
      }
    }

    if (!curlTargetCollection) {
      notifications.error("Select a destination collection");
      return;
    }

    try {
      await ImportCurlRequest(curlInput, curlTargetCollection);
      await collectionStore.loadCollections();
      notifications.success("Request imported from cURL");
      curlInput = "";
    } catch (err) {
      notifications.error("Failed to import cURL request", String(err));
    }
  }

  function parseCollectionNameFromError(message: string): string | null {
    const match = message.match(/collection\s+(\S+)\s+already exists/i);
    return match ? match[1] : null;
  }

  async function executeSoloCollectionImport(path: string, overwrite: boolean) {
    try {
      await ImportSoloCollection(path, overwrite);
      await collectionStore.loadCollections();
      notifications.success("Collection imported successfully");
    } catch (err) {
      const message = String(err ?? "Failed to import collection");
      const existingName = parseCollectionNameFromError(message);
      if (!overwrite && existingName) {
        pendingSoloCollectionPath = path;
        soloCollectionOverwriteName = existingName;
        showSoloCollectionOverwriteDialog.set(true);
        return;
      }
      notifications.error("Failed to import collection", message);
    }
  }

  async function handleImportSoloCollection(path?: string) {
    const filePath = path ?? (await SelectFile("Select Solo Collection", "*.json", "JSON Files"));
    if (!filePath) return;
    showImportSelector.set(false);
    await executeSoloCollectionImport(filePath, false);
  }

  async function confirmSoloCollectionOverwrite() {
    if (!pendingSoloCollectionPath) return;
    const path = pendingSoloCollectionPath;
    pendingSoloCollectionPath = null;
    showSoloCollectionOverwriteDialog.set(false);
    await executeSoloCollectionImport(path, true);
  }

  async function handleExportCollection(collectionName: string) {
    try {
      await ExportCollection(collectionName);
      notifications.success("Collection exported successfully");
    } catch (err) {
      notifications.error("Failed to export collection", String(err));
    }
  }

  async function handleSelectImportFormat(format: "postman" | "bruno" | "openapi") {
    showImportSelector.set(false);
    if (format === "postman") {
      await handleImportPostman();
    } else if (format === "bruno") {
      await handleImportBruno();
    } else if (format === "openapi") {
      await handleImportOpenAPI();
    }
  }

  async function handleSync(collectionId: string) {
    syncingCollections.add(collectionId);
    syncingCollections = new SvelteSet(syncingCollections);
    try {
      await SyncGitCollection(collectionId);
      notifications.success("Git collection synced successfully");
      await collectionStore.loadCollections();
    } catch (err) {
      notifications.error("Sync failed", String(err));
    } finally {
      syncingCollections.delete(collectionId);
      syncingCollections = new SvelteSet(syncingCollections);
    }
  }

  function getProviderIconPath(provider: string) {
    switch (provider) {
      case "github":
        return "M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.003-.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z";
      case "gitlab":
        return "M12 1L9 11h6L12 1zm0 0L3 11l9 12 9-12-9-10z";
      default:
        return "M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5";
    }
  }

  function closeImportModal() {
    showImportSelector.set(false);
    curlInput = "";
    curlCreatingNew = false;
    curlNewCollectionName = "";
  }

  function openImportModal() {
    importActiveTab = "postman";
    gitImportActionState = null;
    showImportSelector.set(true);
  }

  onMount(() => {
    const storedWidth = localStorage.getItem("sidebar_width");
    if (storedWidth) {
      sidebarWidth = parseInt(storedWidth, 10);
    }

    const stored = localStorage.getItem("sidebar_collapsed");
    if (stored !== null) {
      isCollapsed = stored === "true";
    }
  });

  onDestroy(async () => {
    modalStack.close(newCollectionModalId);
    modalStack.close(renameCollectionModalId);
    modalStack.close(deleteCollectionModalId);
    modalStack.close(deleteRequestModalId);
    modalStack.close(importCollectionModalId);
    modalStack.close(soloCollectionOverwriteModalId);
  });
  let collections = $derived(collectionStoreState.collections);
  // Highlight in sidebar is driven by the active tab, not the collectionStore selection
  let selectedRequestId = $derived(
    tabStoreState.tabs.find((t) => t.id === getActiveTab().id)?.id ?? null
  );

  let normalizedQuery = $derived(searchQuery.trim().toLowerCase());
  let isSearching = $derived(normalizedQuery.length > 0);
  let filteredCollections = $derived(
    collections.filter((collection) => shouldShowCollection(collection, normalizedQuery))
  );
</script>

<div
  class="relative flex h-full shrink-0 flex-col border-r border-neutral-200 bg-white dark:border-neutral-800 dark:bg-neutral-900"
  class:collapsed={isCollapsed}
  style={`width: ${isCollapsed ? "auto" : sidebarWidth + "px"};`}
>
  <button
    type="button"
    class="absolute top-0 right-0 z-20 h-full w-1 cursor-col-resize bg-transparent p-0"
    onmousedown={startResize}
    aria-label="Resize sidebar"
  ></button>

  <div class="border-b border-neutral-200 p-3 dark:border-neutral-800">
    <div class="mb-2 flex flex-wrap items-center justify-between gap-1">
      {#if !isCollapsed}
        <h3 class="text-sm font-semibold text-neutral-800 dark:text-neutral-100">Collections</h3>
      {/if}

      <div class="flex items-center gap-1">
        {#if !isCollapsed}
          <Button color="light" size="sm" onclick={openImportModal}>Import</Button>
          <Button color="primary" size="sm" onclick={() => showNewCollectionDialog.set(true)}>
            New
          </Button>
        {/if}
        <Button color="light" size="sm" onclick={toggleCollapse}>
          {isCollapsed ? ">" : "<"}
        </Button>
      </div>
    </div>

    {#if !isCollapsed}
      <div class="flex items-center gap-2">
        <Input
          size="sm"
          class="flex-1"
          type="text"
          placeholder="Search collections or requests"
          bind:value={searchQuery}
        />
        {#if searchQuery}
          <Button
            color="light"
            size="sm"
            onclick={() => (searchQuery = "")}
            aria-label="Clear search"
          >
            Clear
          </Button>
        {/if}
      </div>
    {/if}
  </div>

  {#if !isCollapsed}
    <div class="min-h-0 flex-1 overflow-y-auto p-2">
      {#if collectionStoreState.loading}
        <div class="p-3 text-sm text-neutral-500 dark:text-neutral-400">Loading collections...</div>
      {/if}

      <div class="space-y-2">
        {#each filteredCollections as collection (collection.id)}
          <div
            class="rounded-lg border border-neutral-200 bg-neutral-50 dark:border-neutral-700 dark:bg-neutral-800/40"
          >
            <div
              class="relative flex items-center gap-2 px-2 py-2"
              onclick={(e) => {
                e.stopPropagation();
                selectCollection(collection.name);
                toggleCollection(collection.name);
              }}
              onkeypress={(e) => {
                if (e.key === "Enter") {
                  selectCollection(collection.name);
                  toggleCollection(collection.name);
                }
              }}
              role="button"
              tabindex="0"
            >
              <button
                class="h-6 w-6 p-0 text-xs hover:cursor-pointer dark:text-white"
                onclick={(e: MouseEvent) => {
                  e.stopPropagation();
                  toggleCollection(collection.name);
                }}
                aria-label="Toggle collection"
              >
                {#if isExpanded(collection.name)}
                  <AngleDownOutline />
                {:else}
                  <AngleRightOutline />
                {/if}
              </button>

              <div class="min-w-0 flex-1">
                <div class="flex items-center gap-2 cursor-pointer">
                  {#if collection.gitRemote}
                    <svg
                      width="12"
                      height="12"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2"
                      class="text-neutral-500 dark:text-neutral-400"
                      aria-label={`Git remote: ${collection.gitRemote}`}
                    >
                      <path d={getProviderIconPath(collection.gitProvider || "git")} />
                    </svg>
                  {/if}
                  <span
                    class="truncate text-sm font-medium text-neutral-800 dark:text-neutral-100"
                  >
                    {collection.name}
                  </span>
                  <span
                    class="rounded bg-neutral-200 px-1.5 py-0.5 text-xs text-neutral-600 dark:bg-neutral-700 dark:text-neutral-300"
                  >
                    {collection.requests?.length || 0}
                  </span>
                </div>
              </div>

              <div class="flex items-center gap-1">
                <button
                  class="hover:cursor-pointer"
                  onclick={(e: MouseEvent) => {
                    e.stopPropagation();
                    handleAddRequest(e, collection.name);
                  }}
                  title="Add request"
                  aria-label="Add request"
                >
                  <PlusOutline />
                </button>
                <button
                  id="collection-menu-{collection.id}"
                  class="ml-1 hover:cursor-pointer"
                  title="More actions"
                  aria-label="More actions"
                  onclick={(e: MouseEvent) => e.stopPropagation()}
                >
                  <DotsHorizontalOutline />
                </button>
                <Dropdown
                  triggeredBy="#collection-menu-{collection.id}"
                  class="z-50 w-40"
                  triggerDelay={0}
                >
                  {#if collection.gitRemote}
                    <DropdownItem
                      class="text-gray-900 dark:text-white"
                      onclick={() => {
                        gitStatusCollectionId = collection.id;
                        gitStatusCollectionName = collection.name;
                      }}
                    >
                      Git status
                    </DropdownItem>
                    <DropdownItem
                      class="text-gray-900 dark:text-white"
                      disabled={syncingCollections.has(collection.id)}
                      onclick={() => {
                        handleSync(collection.id);
                      }}
                    >
                      {syncingCollections.has(collection.id) ? "Syncing…" : "Sync with Git"}
                    </DropdownItem>
                    <DropdownDivider />
                  {/if}
                  <DropdownItem
                    class="text-gray-900 dark:text-white"
                    onclick={() => handleExportCollection(collection.name)}
                  >
                    Export
                  </DropdownItem>
                  <DropdownDivider />
                  <DropdownItem
                    class="text-gray-900 dark:text-white"
                    onclick={() => {
                      openRenameCollection(collection.name);
                    }}
                  >
                    Rename
                  </DropdownItem>
                  <DropdownItem
                    class="text-danger-600 hover:bg-danger-50 dark:text-danger-400 dark:hover:bg-danger-900/20"
                    onclick={() => {
                      handleDeleteCollection(collection.name);
                    }}
                  >
                    Delete
                  </DropdownItem>
                </Dropdown>
              </div>
            </div>

            {#if isExpanded(collection.name)}
              <div
                class="space-y-1 border-t border-neutral-200 px-2 pt-1 pb-2 dark:border-neutral-700"
              >
                {#if getVisibleRequests(collection, normalizedQuery).length === 0}
                  <div class="px-1 py-2 text-xs text-neutral-500 dark:text-neutral-400">
                    {isSearching ? "No matching requests" : "No requests yet"}
                  </div>
                {:else}
                  {#each getVisibleRequests(collection, normalizedQuery) as request (request.id)}
                    <div
                      class={`cursor-pointer group flex items-center gap-2 rounded px-2 py-1.5 hover:bg-neutral-100 dark:hover:bg-neutral-700/60 ${selectedRequestId === request.id ? "bg-neutral-200/70 dark:bg-neutral-700/90" : ""}`}
                      onclick={() => selectRequest(request, collection.name)}
                      ondblclick={(e) => startRequestRename(e, request, collection.name)}
                      onkeypress={(e) =>
                        e.key === "Enter" &&
                        !isEditingRequest(request.id, collection.name) &&
                        selectRequest(request, collection.name)}
                      role="button"
                      tabindex="0"
                    >
                      <span class={getMethodBadgeClass(request.verb)}>
                        {request.verb}
                      </span>
                      {#if isEditingRequest(request.id, collection.name)}
                        <Input
                          type="text"
                          size="sm"
                          class="min-w-0 flex-1"
                          bind:value={editingRequestName}
                          bind:elementRef={editingRequestNameInputEl}
                          autofocus
                          onclick={(e) => e.stopPropagation()}
                          onkeydown={(e) => {
                            e.stopPropagation();
                            if (e.key === "Enter") {
                              void commitRequestRename(request, collection.name);
                            }
                            if (e.key === "Escape") {
                              cancelRequestRename();
                            }
                          }}
                          onblur={() => void commitRequestRename(request, collection.name)}
                        />
                      {:else}
                        <span
                          class="min-w-0 flex-1 truncate text-sm text-neutral-800 dark:text-neutral-100"
                          role="button"
                          tabindex="0"
                          ondblclick={(e) => startRequestRename(e, request, collection.name)}
                          onkeydown={(e) =>
                            e.key === "Enter" && startRequestRename(e, request, collection.name)}
                        >
                          {request.name}
                        </span>
                      {/if}
                      <button
                        class="invisible group-hover:visible hover:cursor-pointer"
                        id="request-menu-{request.id}"
                        onclick={(e: MouseEvent) => {
                          e.stopPropagation();
                        }}
                        title="Request actions"
                        aria-label="Request actions"
                      >
                        <DotsHorizontalOutline />
                      </button>
                      <Dropdown
                        triggeredBy="#request-menu-{request.id}"
                        class="z-50 w-40"
                        triggerDelay={0}
                      >
                        <DropdownItem
                          class="text-gray-900 dark:text-white"
                          onclick={(e) => {
                            startRequestRename(e, request, collection.name);
                          }}
                        >
                          Rename
                        </DropdownItem>
                        <DropdownItem
                          class="text-danger-600 hover:bg-danger-50 dark:text-danger-400 dark:hover:bg-danger-900/20"
                          onclick={() => {
                            handleDeleteRequest(collection.name, request.id);
                          }}
                        >
                          Delete
                        </DropdownItem>
                      </Dropdown>
                    </div>
                  {/each}
                {/if}
              </div>
            {/if}
          </div>
        {/each}
      </div>

      {#if filteredCollections.length === 0 && !collectionStoreState.loading}
        <div class="pt-2">
          <FeedbackEmptyState
            title={isSearching ? "No matching collections or requests" : "No collections yet"}
            detail={!isSearching ? "Create your first collection to get started" : undefined}
          />
        </div>
      {/if}
    </div>
  {/if}
</div>

{#if $showNewCollectionDialog}
  <Modal
    bind:open={$showNewCollectionDialog}
    onclose={closeNewCollectionDialog}
    title="New Collection"
  >
    {#if $topModalId === newCollectionModalId}
      <ToastContainer />
    {/if}
    <div class="space-y-2">
      <Label for="new-collection-name">Collection name</Label>
      <Input
        id="new-collection-name"
        type="text"
        bind:value={newCollectionName}
        placeholder="Collection name"
        onkeydown={(e) => e.key === "Enter" && handleCreateCollection()}
        autofocus
      />
    </div>
    {#snippet footer()}
      <div class="flex w-full justify-end gap-2">
        <Button color="primary" onclick={handleCreateCollection}>Create</Button>
      </div>
    {/snippet}
  </Modal>
{/if}

{#if $showRenameCollectionDialog}
  <Modal
    bind:open={$showRenameCollectionDialog}
    onclose={closeRenameDialog}
    title="Rename Collection"
  >
    {#if $topModalId === renameCollectionModalId}
      <ToastContainer />
    {/if}
    <div class="space-y-2">
      <Label for="rename-collection-name">Collection name</Label>
      <Input
        id="rename-collection-name"
        type="text"
        bind:value={renameCollectionName}
        placeholder="Collection name"
        onkeydown={(e) => e.key === "Enter" && handleRenameCollection()}
        autofocus
      />
    </div>
    {#snippet footer()}
      <div class="flex w-full justify-end gap-2">
        <Button color="primary" onclick={handleRenameCollection}>Save</Button>
      </div>
    {/snippet}
  </Modal>
{/if}

{#if $showDeleteConfirmDialog}
  <Modal
    bind:open={$showDeleteConfirmDialog}
    onclose={closeDeleteConfirmDialog}
    title="Delete Collection"
  >
    {#if $topModalId === deleteCollectionModalId}
      <ToastContainer />
    {/if}
    <div class="space-y-2 text-sm">
      <p class="text-neutral-700 dark:text-neutral-200">
        Are you sure you want to delete "{deleteTarget}"?
      </p>
      <p class="text-danger-600 dark:text-danger-300">This action cannot be undone.</p>
    </div>
    {#snippet footer()}
      <div class="flex w-full justify-end gap-2">
        <Button color="red" onclick={confirmDelete}>Delete</Button>
      </div>
    {/snippet}
  </Modal>
{/if}
{#if $showDeleteRequestConfirmDialog}
  <Modal
    bind:open={$showDeleteRequestConfirmDialog}
    onclose={closeDeleteRequestConfirmDialog}
    title="Delete Request"
  >
    {#if $topModalId === deleteRequestModalId}
      <ToastContainer />
    {/if}
    <div class="space-y-2 text-sm">
      <p class="text-neutral-700 dark:text-neutral-200">
        Are you sure you want to delete this request?
      </p>
      <p class="text-danger-600 dark:text-danger-300">This action cannot be undone.</p>
    </div>
    {#snippet footer()}
      <div class="flex w-full justify-end gap-2">
        <Button color="red" onclick={confirmDeleteRequest}>Delete</Button>
      </div>
    {/snippet}
  </Modal>
{/if}

{#if $showImportSelector}
  <Modal
    title="Import Collection or Request"
    bind:open={$showImportSelector}
    onclose={closeImportModal}
    size="xl"
  >
    {#if $topModalId === importCollectionModalId}
      <ToastContainer />
    {/if}
    <div class="flex flex-col gap-4">
      <Tabs bind:selected={importActiveTab}>
        <TabItem key="postman" title="Postman">
          <DropZone
            title="Drop your Postman collection here"
            subtitle="Supports Postman Collection v2 / v2.1 (JSON)"
            onDrop={async (e) => {
              const paths = e.paths;
              showImportSelector.set(false);
              if (paths.length > 0) {
                await ImportPostmanCollection(paths[0]);
                await collectionStore.loadCollections();
              } else {
                await handleImportPostman();
              }
            }}
          >
            {#snippet icon()}
              <svg
                width="44"
                height="44"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="1.4"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                <polyline points="17 8 12 3 7 8" />
                <line x1="12" y1="3" x2="12" y2="15" />
              </svg>
            {/snippet}
          </DropZone>
        </TabItem>

        <TabItem key="bruno" title="Bruno">
          <DropZone
            title="Drop your Bruno collection folder here"
            subtitle="Supports Bruno collection folders (.bru files)"
            onDrop={async (e) => {
              const paths = e.paths;
              showImportSelector.set(false);
              if (paths.length > 0) {
                await ImportBrunoCollection(paths[0]);
                await collectionStore.loadCollections();
              } else {
                await handleImportBruno();
              }
            }}
          >
            {#snippet icon()}
              <svg
                width="44"
                height="44"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="1.4"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
                <polyline points="9 22 9 12 15 12 15 22" />
              </svg>
            {/snippet}
          </DropZone>
        </TabItem>

        <TabItem key="openapi" title="OpenAPI / Swagger">
          <DropZone
            title="Drop your OpenAPI or Swagger document here"
            subtitle="Supports OpenAPI 3.x and Swagger 2.x (JSON or YAML)"
            onDrop={async (e) => {
              const paths = e.paths;
              showImportSelector.set(false);
              if (paths.length > 0) {
                const warnings = await ImportOpenAPICollection(paths[0]);
                await collectionStore.loadCollections();
                for (const w of warnings) {
                  notifications.warning(w);
                }
                notifications.success("OpenAPI collection imported");
              } else {
                await handleImportOpenAPI();
              }
            }}
          >
            {#snippet icon()}
              <svg
                width="44"
                height="44"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="1.4"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
                <polyline points="14 2 14 8 20 8" />
                <line x1="16" y1="13" x2="8" y2="13" />
                <line x1="16" y1="17" x2="8" y2="17" />
                <polyline points="10 9 9 9 8 9" />
              </svg>
            {/snippet}
          </DropZone>
        </TabItem>

        <TabItem key="curl" title="cURL">
          <div class="flex flex-col gap-4">
            <div class="flex flex-col gap-1">
              <Label>Paste cURL command</Label>
              <div
                class="h-48 overflow-hidden rounded border border-neutral-200 dark:border-neutral-700"
              >
                <CodeMirrorEditor
                  value={curlInput}
                  language="text"
                  onChange={(v) => (curlInput = v)}
                />
              </div>
            </div>

            <div class="flex flex-col gap-1">
              <Label>Destination collection</Label>
              {#if curlCreatingNew}
                <div class="flex gap-2">
                  <Input
                    bind:value={curlNewCollectionName}
                    placeholder="New collection name"
                    class="flex-1"
                  />
                  <Button
                    color="alternative"
                    size="sm"
                    onclick={() => {
                      curlCreatingNew = false;
                      curlNewCollectionName = "";
                    }}
                  >
                    Cancel
                  </Button>
                </div>
              {:else}
                <div class="flex gap-2">
                  <Select bind:value={curlTargetCollection} class="flex-1">
                    {#each collectionStoreState.collections as coll (coll.id)}
                      <option value={coll.name}>{coll.name}</option>
                    {/each}
                  </Select>
                  <Button color="alternative" size="sm" onclick={() => (curlCreatingNew = true)}>
                    New…
                  </Button>
                </div>
              {/if}
            </div>
          </div>
        </TabItem>

        <TabItem key="solo" title="Solo">
          <DropZone
            title="Drop your Solo collection here"
            subtitle="Supports Solo collection JSON"
            onDrop={async (e) => {
              const paths = e.paths;
              showImportSelector.set(false);
              if (paths.length > 0) {
                await executeSoloCollectionImport(paths[0], false);
              } else {
                await handleImportSoloCollection();
              }
            }}
          >
            {#snippet icon()}
              <svg
                width="44"
                height="44"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="1.4"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                <polyline points="17 8 12 3 7 8" />
                <line x1="12" y1="3" x2="12" y2="15" />
              </svg>
            {/snippet}
          </DropZone>
        </TabItem>

        <TabItem key="git" title="Git">
          <GitImportView
            onImported={() => showImportSelector.set(false)}
            onActionStateChange={(state) => {
              gitImportActionState = state;
            }}
          />
        </TabItem>
      </Tabs>
    </div>

    {#snippet footer()}
      <div class="flex w-full gap-2">
        {#if importActiveTab === "postman"}
          <Button color="primary" onclick={() => handleSelectImportFormat("postman")}>
            Select file…
          </Button>
        {:else if importActiveTab === "bruno"}
          <Button color="primary" onclick={() => handleSelectImportFormat("bruno")}>
            Select folder…
          </Button>
        {:else if importActiveTab === "openapi"}
          <Button color="primary" onclick={() => handleSelectImportFormat("openapi")}>
            Select file…
          </Button>
        {:else if importActiveTab === "curl"}
          <Button
            color="primary"
            disabled={!curlInput.trim() || (!curlTargetCollection && !curlCreatingNew)}
            onclick={handleImportCurl}
          >
            Import Request
          </Button>
        {:else if importActiveTab === "solo"}
          <Button color="primary" onclick={() => handleImportSoloCollection()}>
            Select file...
          </Button>
        {:else if importActiveTab === "git"}
          <Button
            color="primary"
            loading={gitImportActionState?.loading ?? false}
            disabled={gitImportActionState?.disabled ?? true}
            onclick={() => gitImportActionState?.submit()}
          >
            Import from Git
          </Button>
        {/if}
      </div>
    {/snippet}
  </Modal>
{/if}

{#if $showSoloCollectionOverwriteDialog}
  <Modal
    title="Overwrite collection?"
    bind:open={$showSoloCollectionOverwriteDialog}
    onclose={() => {
      pendingSoloCollectionPath = null;
      soloCollectionOverwriteName = null;
      showSoloCollectionOverwriteDialog.set(false);
    }}
    size="xl"
  >
    {#if $topModalId === soloCollectionOverwriteModalId}
      <ToastContainer />
    {/if}
    <p>Collection "{soloCollectionOverwriteName}" already exists.</p>
    <p class="text-sm text-neutral-600 dark:text-neutral-400">Do you want to overwrite it?</p>
    {#snippet footer()}
      <div class="flex w-full justify-end gap-2">
        <Button color="red" onclick={confirmSoloCollectionOverwrite}>Overwrite</Button>
      </div>
    {/snippet}
  </Modal>
{/if}

{#if gitStatusCollectionId && gitStatusCollectionName}
  <GitStatusPanel
    entityId={gitStatusCollectionId}
    entityName={gitStatusCollectionName}
    fnGetStatus={GetGitCollectionStatus}
    fnSync={SyncGitCollection}
    fnKeepOurs={GitKeepOurs}
    fnKeepTheirs={GitKeepTheirs}
    fnAbortRebase={GitAbortRebase}
    fnDiscard={GitDiscardChanges}
    fnOpenTerminal={OpenCollectionInTerminal}
    onReload={collectionStore.loadCollections}
    onClose={() => {
      gitStatusCollectionId = null;
      gitStatusCollectionName = null;
    }}
  />
{/if}
