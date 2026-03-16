<script lang="ts">
  import Button from "flowbite-svelte/Button.svelte";
  import DropZone from "$src/lib/components/base/DropZone.svelte";
  import Input from "flowbite-svelte/Input.svelte";
  import Label from "flowbite-svelte/Label.svelte";
  import Modal from "flowbite-svelte/Modal.svelte";
  import TabItem from "flowbite-svelte/TabItem.svelte";
  import Tabs from "flowbite-svelte/Tabs.svelte";
  import FeedbackEmptyState from "$src/lib/components/common/FeedbackEmptyState.svelte";
  import GitImportView from "$src/lib/components/GitImportView.svelte";
  import ToastContainer from "$src/lib/components/base/ToastContainer.svelte";
  import GitStatusPanel from "$src/lib/components/GitStatusPanel.svelte";
  import { collectionStore } from "$src/lib/stores/collectionStore";
  import { modalStack, topModalId } from "$src/lib/stores/modalStackStore";
  import { notifications } from "$src/lib/stores/notificationStore";
  import { tabStore } from "$src/lib/stores/tabStore";
  import {
    GetGitCollectionStatus,
    GitAbortRebase,
    GitDiscardChanges,
    GitKeepOurs,
    GitKeepTheirs,
    ImportBrunoCollection,
    ImportPostmanCollection,
    OpenCollectionInTerminal,
    SelectDirectory,
    SelectFile,
    SyncGitCollection
  } from "$wails/go/main/App";
  import { collection } from "$wails/go/models";
  import { onDestroy, onMount } from "svelte";
  import { SvelteSet } from "svelte/reactivity";

  interface Props {
    onRequestSelect?: (requestId: string) => void;
  }

  let { onRequestSelect = () => {} }: Props = $props();

  let showNewCollectionDialog = $state(false);
  let showRenameCollectionDialog = $state(false);
  let showDeleteConfirmDialog = $state(false);
  let showDeleteRequestConfirmDialog = $state(false);
  let showImportSelector = $state(false);

  const collectionModalScope = `collections-${Math.random().toString(36).slice(2)}`;
  const newCollectionModalId = `${collectionModalScope}-new`;
  const renameCollectionModalId = `${collectionModalScope}-rename`;
  const deleteCollectionModalId = `${collectionModalScope}-delete-collection`;
  const deleteRequestModalId = `${collectionModalScope}-delete-request`;
  const importCollectionModalId = `${collectionModalScope}-import`;

  $effect(() => {
    if (showNewCollectionDialog) {
      modalStack.open(newCollectionModalId);
    } else {
      modalStack.close(newCollectionModalId);
    }
  });

  $effect(() => {
    if (showRenameCollectionDialog) {
      modalStack.open(renameCollectionModalId);
    } else {
      modalStack.close(renameCollectionModalId);
    }
  });

  $effect(() => {
    if (showDeleteConfirmDialog) {
      modalStack.open(deleteCollectionModalId);
    } else {
      modalStack.close(deleteCollectionModalId);
    }
  });

  $effect(() => {
    if (showDeleteRequestConfirmDialog) {
      modalStack.open(deleteRequestModalId);
    } else {
      modalStack.close(deleteRequestModalId);
    }
  });

  $effect(() => {
    if (showImportSelector) {
      modalStack.open(importCollectionModalId);
    } else {
      modalStack.close(importCollectionModalId);
    }
  });

  let importActiveTab = $state("postman");
  let gitImportActionState: { loading: boolean; disabled: boolean; submit: () => void } | null =
    $state(null);

  let newCollectionName = $state("");
  let renameCollectionName = $state("");
  let renameTarget: string | null = null;
  let deleteTarget: string | null = $state(null);
  let deleteRequestTarget: string | null = null;
  let deleteRequestCollectionName: string | null = null;
  let expandedCollections: Set<string> = new SvelteSet();
  let searchQuery = $state("");
  let activeMenu: string | null = $state(null);
  let isCollapsed = $state(false);
  let gitStatusCollectionId: string | null = $state(null);
  let gitStatusCollectionName: string | null = $state(null);
  let syncingCollections: Set<string> = $state(new Set());

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
    expandedCollections = new SvelteSet(expandedCollections);
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
      settings: req.settings || {},
      preRequestScript: req.preRequestScript || "",
      postResponseScript: req.postResponseScript || ""
    });

    onRequestSelect(req.id);
  }

  function openRenameCollection(collectionName: string) {
    renameTarget = collectionName;
    renameCollectionName = collectionName;
    showRenameCollectionDialog = true;
    activeMenu = null;
  }

  function closeNewCollectionDialog() {
    showNewCollectionDialog = false;
    newCollectionName = "";
  }

  function closeRenameDialog() {
    showRenameCollectionDialog = false;
    renameTarget = null;
    renameCollectionName = "";
  }

  async function handleCreateCollection() {
    const trimmed = newCollectionName.trim();
    if (!trimmed) return;

    const exists = collections.some(
      (collection) => collection.name.toLowerCase() === trimmed.toLowerCase()
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

    const exists = collections.some(
      (collection) => collection.name.toLowerCase() === trimmed.toLowerCase()
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
        expandedCollections = new SvelteSet(expandedCollections);
      }
      closeRenameDialog();
    } catch {
      // error already shown by store
    }
  }

  function handleDeleteCollection(collectionName: string) {
    deleteTarget = collectionName;
    showDeleteConfirmDialog = true;
    activeMenu = null;
  }

  function closeDeleteConfirmDialog() {
    showDeleteConfirmDialog = false;
    deleteTarget = null;
  }

  async function confirmDelete() {
    if (!deleteTarget) return;

    try {
      await collectionStore.deleteCollection(deleteTarget);
      expandedCollections.delete(deleteTarget);
      expandedCollections = new SvelteSet(expandedCollections);
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
      expandedCollections = new SvelteSet(expandedCollections);

      if (newReq?.id) {
        tabStore.openTab(newReq.id, collectionName, {
          label: "New Request",
          verb: "GET",
          url: "",
          body: "",
          bodyFormat: "json",
          headers: [],
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
    showDeleteRequestConfirmDialog = true;
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
    showDeleteRequestConfirmDialog = false;
    deleteRequestTarget = null;
    deleteRequestCollectionName = null;
  }

  function getMethodClass(method: string): string {
    const m = method.toUpperCase();
    if (m === "GET") return "bg-success-100 text-success-700 dark:bg-success-900/50 dark:text-success-300";
    if (m === "POST") return "bg-primary-100 text-primary-700 dark:bg-primary-900/50 dark:text-primary-300";
    if (m === "PUT" || m === "PATCH") {
      return "bg-warning-100 text-warning-700 dark:bg-warning-900/50 dark:text-warning-300";
    }
    if (m === "DELETE") return "bg-danger-100 text-danger-700 dark:bg-danger-900/50 dark:text-danger-300";
    return "bg-neutral-100 text-neutral-700 dark:bg-neutral-800 dark:text-neutral-300";
  }

  function toggleMenu(e: Event, collectionName: string) {
    e.stopPropagation();
    activeMenu = activeMenu === collectionName ? null : collectionName;
  }

  function clearMenu() {
    if (activeMenu) {
      activeMenu = null;
    }
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

  async function handleSelectImportFormat(format: "postman" | "bruno") {
    showImportSelector = false;
    if (format === "postman") {
      await handleImportPostman();
    } else if (format === "bruno") {
      await handleImportBruno();
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

  function openImportModal() {
    importActiveTab = "postman";
    gitImportActionState = null;
    showImportSelector = true;
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

  onMount(async () => {
    document.addEventListener("click", clearMenu);
  });
  onDestroy(async () => {
    document.removeEventListener("click", clearMenu);
    modalStack.close(newCollectionModalId);
    modalStack.close(renameCollectionModalId);
    modalStack.close(deleteCollectionModalId);
    modalStack.close(deleteRequestModalId);
    modalStack.close(importCollectionModalId);
  });
  let collections = $derived($collectionStore.collections);
  let selectedCollectionName = $derived($collectionStore.selectedCollectionName);
  // Highlight in sidebar is driven by the active tab, not the collectionStore selection
  let selectedRequestId = $derived(
    $tabStore.tabs.find((t) => t.id === $tabStore.activeTabId)?.requestId ?? null
  );
  let normalizedQuery = $derived(searchQuery.trim().toLowerCase());
  let isSearching = $derived(normalizedQuery.length > 0);
  let filteredCollections = $derived(
    collections.filter((collection) => shouldShowCollection(collection, normalizedQuery))
  );
</script>

<div
  class="relative flex h-full flex-col border-r border-neutral-200 bg-white dark:border-neutral-800 dark:bg-neutral-900"
  class:collapsed={isCollapsed}
  style={`width: ${isCollapsed ? "auto" : sidebarWidth + "px"};`}
>
  <div class="absolute right-0 top-0 z-20 h-full w-1 cursor-col-resize" onmousedown={startResize}></div>

  <div class="border-b border-neutral-200 p-3 dark:border-neutral-800">
    <div class="mb-2 flex items-center justify-between gap-2">
      {#if !isCollapsed}
        <h3 class="text-sm font-semibold text-neutral-800 dark:text-neutral-100">Collections</h3>
      {/if}

      <div class="flex items-center gap-1">
        {#if !isCollapsed}
          <Button color="light" size="sm" onclick={openImportModal}>Import</Button>
          <Button color="primary" size="sm" onclick={() => (showNewCollectionDialog = true)}>
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
          <Button color="light" size="sm" onclick={() => (searchQuery = "")} aria-label="Clear search">
            Clear
          </Button>
        {/if}
      </div>
    {/if}
  </div>

  {#if !isCollapsed}
    <div class="min-h-0 flex-1 overflow-y-auto p-2">
      {#if $collectionStore.loading}
        <div class="p-3 text-sm text-neutral-500 dark:text-neutral-400">Loading collections...</div>
      {/if}

      <div class="space-y-2">
        {#each filteredCollections as collection (collection.id)}
          <div
            class="rounded-lg border border-neutral-200 bg-neutral-50 dark:border-neutral-700 dark:bg-neutral-800/40"
            class:ring-1={selectedCollectionName === collection.name}
            class:ring-primary-400={selectedCollectionName === collection.name}
            class:ring-offset-0={selectedCollectionName === collection.name}
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
                class="h-6 w-6 rounded text-xs text-neutral-600 hover:bg-neutral-200 dark:text-neutral-300 dark:hover:bg-neutral-700"
                onclick={(e) => {
                  e.stopPropagation();
                  toggleCollection(collection.name);
                }}
                aria-label="Toggle collection"
              >
                {isExpanded(collection.name) ? "▾" : "▸"}
              </button>

              <div class="min-w-0 flex-1">
                <div class="flex items-center gap-2">
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
                  <span class="truncate text-sm font-medium text-neutral-800 dark:text-neutral-100">
                    {collection.name}
                  </span>
                  <span class="rounded bg-neutral-200 px-1.5 py-0.5 text-xs text-neutral-600 dark:bg-neutral-700 dark:text-neutral-300">
                    {collection.requests?.length || 0}
                  </span>
                </div>
              </div>

              <div class="flex items-center gap-1">
                {#if collection.gitRemote}
                  <Button
                    color="light"
                    size="sm"
                    onclick={(e: MouseEvent) => {
                      e.stopPropagation();
                      gitStatusCollectionId = collection.id;
                      gitStatusCollectionName = collection.name;
                    }}
                    title="Git status & actions"
                  >
                    Git
                  </Button>
                  <Button
                    color="light"
                    size="sm"
                    loading={syncingCollections.has(collection.id)}
                    onclick={(e: MouseEvent) => {
                      e.stopPropagation();
                      handleSync(collection.id);
                    }}
                    title="Sync with Git remote"
                    disabled={syncingCollections.has(collection.id)}
                  >
                    Sync
                  </Button>
                {/if}

                <Button
                  color="light"
                  size="sm"
                  onclick={(e: MouseEvent) => handleAddRequest(e, collection.name)}
                  title="Add request"
                  aria-label="Add request"
                >
                  +
                </Button>
                <Button
                  color="light"
                  size="sm"
                  onclick={(e: MouseEvent) => toggleMenu(e, collection.name)}
                  title="More actions"
                  aria-label="More actions"
                >
                  ...
                </Button>
              </div>

              {#if activeMenu === collection.name}
                <div
                  class="absolute right-2 top-10 z-10 w-36 rounded-lg border border-neutral-200 bg-white p-1 shadow-lg dark:border-neutral-700 dark:bg-neutral-800"
                >
                  <button
                    class="block w-full rounded px-2 py-1.5 text-left text-sm text-neutral-700 hover:bg-neutral-100 dark:text-neutral-200 dark:hover:bg-neutral-700"
                    onclick={(e) => {
                      e.stopPropagation();
                      openRenameCollection(collection.name);
                    }}
                  >
                    Rename
                  </button>
                  <button
                    class="block w-full rounded px-2 py-1.5 text-left text-sm text-danger-700 hover:bg-danger-50 dark:text-danger-300 dark:hover:bg-danger-900/40"
                    onclick={(e) => {
                      e.stopPropagation();
                      handleDeleteCollection(collection.name);
                    }}
                  >
                    Delete
                  </button>
                </div>
              {/if}
            </div>

            {#if isExpanded(collection.name)}
              <div class="space-y-1 border-t border-neutral-200 px-2 pb-2 pt-1 dark:border-neutral-700">
                {#if getVisibleRequests(collection, normalizedQuery).length === 0}
                  <div class="px-1 py-2 text-xs text-neutral-500 dark:text-neutral-400">
                    {isSearching ? "No matching requests" : "No requests yet"}
                  </div>
                {:else}
                  {#each getVisibleRequests(collection, normalizedQuery) as request (request.id)}
                    <div
                      class={`flex items-center gap-2 rounded px-2 py-1.5 hover:bg-neutral-100 dark:hover:bg-neutral-700/60 ${selectedRequestId === request.id ? "bg-neutral-200/70 dark:bg-neutral-700/90" : ""}`}
                      onclick={() => selectRequest(request, collection.name)}
                      onkeypress={(e) =>
                        e.key === "Enter" && selectRequest(request, collection.name)}
                      role="button"
                      tabindex="0"
                    >
                      <span
                        class={`inline-flex min-w-14 justify-center rounded px-1.5 py-0.5 text-[10px] font-semibold ${getMethodClass(request.verb)}`}
                      >
                        {request.verb}
                      </span>
                      <span class="min-w-0 flex-1 truncate text-sm text-neutral-800 dark:text-neutral-100">
                        {request.name}
                      </span>
                      <Button
                        color="light"
                        size="xs"
                        onclick={(e: MouseEvent) => {
                          e.stopPropagation();
                          handleDeleteRequest(collection.name, request.id);
                        }}
                        title="Delete request"
                        aria-label="Delete request"
                      >
                        ×
                      </Button>
                    </div>
                  {/each}
                {/if}
              </div>
            {/if}
          </div>
        {/each}
      </div>

      {#if filteredCollections.length === 0 && !$collectionStore.loading}
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

{#if showNewCollectionDialog}
  <Modal
    bind:open={showNewCollectionDialog}
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

{#if showRenameCollectionDialog}
  <Modal
    bind:open={showRenameCollectionDialog}
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

{#if showDeleteConfirmDialog}
  <Modal
    bind:open={showDeleteConfirmDialog}
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
{#if showDeleteRequestConfirmDialog}
  <Modal
    bind:open={showDeleteRequestConfirmDialog}
    onclose={closeDeleteRequestConfirmDialog}
    title="Delete Request"
  >
    {#if $topModalId === deleteRequestModalId}
      <ToastContainer />
    {/if}
    <div class="space-y-2 text-sm">
      <p class="text-neutral-700 dark:text-neutral-200">Are you sure you want to delete this request?</p>
      <p class="text-danger-600 dark:text-danger-300">This action cannot be undone.</p>
    </div>
    {#snippet footer()}
      <div class="flex w-full justify-end gap-2">
        <Button color="red" onclick={confirmDeleteRequest}>Delete</Button>
      </div>
    {/snippet}
  </Modal>
{/if}

{#if showImportSelector}
  <Modal title="Import Collection" bind:open={showImportSelector} size="xl">
    {#if $topModalId === importCollectionModalId}
      <ToastContainer />
    {/if}
    <div class="import-modal-body">
      <Tabs bind:selected={importActiveTab}>
        <TabItem key="postman" title="Postman">
          <DropZone
            title="Drop your Postman collection here"
            subtitle="Supports Postman Collection v2 / v2.1 (JSON)"
            onDrop={async (e) => {
              const paths = e.paths;
              showImportSelector = false;
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
              showImportSelector = false;
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

        <TabItem key="git" title="Git">
          <GitImportView
            onImported={() => (showImportSelector = false)}
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
