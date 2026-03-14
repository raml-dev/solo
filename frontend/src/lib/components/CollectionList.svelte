<script lang="ts">
  import Button from "$src/lib/components/base/Button.svelte";
  import DropZone from "$src/lib/components/base/DropZone.svelte";
  import Modal from "$src/lib/components/base/Modal.svelte";
  import Tab from "$src/lib/components/base/Tab.svelte";
  import Tabs from "$src/lib/components/base/Tabs.svelte";
  import GitImportView from "$src/lib/components/GitImportView.svelte";
  import GitStatusPanel from "$src/lib/components/GitStatusPanel.svelte";
  import { collectionStore } from "$src/lib/stores/collectionStore";
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

  let importActiveTab = $state("postman");

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

  function selectRequest(requestId: string, collectionName: string) {
    // Find the request data to pass metadata to tabStore
    const coll = $collectionStore.collections.find(
      (c: collection.Collection) => c.name === collectionName
    ) as collection.Collection;
    const req = coll?.requests.find((r) => r.id === requestId);
    if (!req) return;

    const headers = req.headers
      ? Object.entries(req.headers).map(([key, value], i) => ({
          id: `header-${i}`,
          key,
          value: String(value),
          enabled: true
        }))
      : [];

    tabStore.openTab(requestId, collectionName, {
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

    onRequestSelect(requestId);
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
    return `method-${method.toLowerCase()}`;
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
  class="collection-list"
  class:collapsed={isCollapsed}
  style={`width: ${isCollapsed ? "auto" : sidebarWidth + "px"};`}
>
  <div class="resize-handle" onmousedown={startResize}></div>
  <div class="header">
    <div class="header-title">
      {#if !isCollapsed}
        <h3>Collections</h3>
      {/if}
      <div class="header-actions">
        {#if !isCollapsed}
          <Button variant="secondary" size="small" click={openImportModal}>Import</Button>
          <Button variant="primary" size="small" click={() => (showNewCollectionDialog = true)}>
            New
          </Button>
        {/if}
        <span title={isCollapsed ? "Expand sidebar" : "Collapse sidebar"}>
          <Button variant="secondary" click={toggleCollapse}>
            {isCollapsed ? ">" : "<"}
          </Button>
        </span>
      </div>
    </div>

    {#if !isCollapsed}
      <div class="search-row">
        <input
          type="text"
          class="input input-sm search-input"
          placeholder="Search collections or requests"
          bind:value={searchQuery}
        />
        {#if searchQuery}
          <button class="clear-search" onclick={() => (searchQuery = "")} aria-label="Clear search">
            x
          </button>
        {/if}
      </div>
    {/if}
  </div>

  {#if !isCollapsed}
    {#if $collectionStore.loading}
      <div class="loading">Loading collections...</div>
    {/if}

    <div class="collections">
      {#each filteredCollections as collection (collection.id)}
        <div
          class="collection-item"
          class:selected={selectedCollectionName === collection.name}
          class:menu-open={activeMenu === collection.name}
        >
          <div
            class="collection-header"
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
              class="expand-btn"
              onclick={(e) => {
                e.stopPropagation();
                toggleCollection(collection.name);
              }}
              aria-label="Toggle collection"
            >
              <span class="expand-icon" class:expanded={isExpanded(collection.name)}> &gt; </span>
            </button>

            <div class="collection-info">
              {#if collection.gitRemote}
                <svg
                  width="12"
                  height="12"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  class="provider-icon"
                  aria-label={`Git remote: ${collection.gitRemote}`}
                >
                  <path d={getProviderIconPath(collection.gitProvider || "git")} />
                </svg>
              {/if}
              <span class="collection-name">{collection.name}</span>
              <span class="collection-count">{collection.requests?.length || 0}</span>
            </div>

            <div class="collection-actions">
              {#if collection.gitRemote}
                <button
                  class="icon-btn"
                  onclick={(e) => {
                    e.stopPropagation();
                    gitStatusCollectionId = collection.id;
                    gitStatusCollectionName = collection.name;
                  }}
                  title="Git status & actions"
                >
                  <svg
                    width="12"
                    height="12"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                  >
                    <circle cx="12" cy="18" r="3" /><circle cx="6" cy="6" r="3" /><circle
                      cx="18"
                      cy="6"
                      r="3"
                    />
                    <path d="M18 9v2a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2V9" />
                    <line x1="12" y1="12" x2="12" y2="15" />
                  </svg>
                </button>
                <button
                  class="icon-btn"
                  onclick={(e) => {
                    e.stopPropagation();
                    handleSync(collection.id);
                  }}
                  title="Sync with Git remote"
                  disabled={syncingCollections.has(collection.id)}
                >
                  {#if syncingCollections.has(collection.id)}
                    <span class="sync-spinner"></span>
                  {:else}
                    <svg
                      width="12"
                      height="12"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2"
                    >
                      <path
                        d="M21 2v6h-6M3 22v-6h6M21 12c0 4.97-4.03 9-9 9-3.32 0-6.23-1.8-7.81-4.47M3 12c0-4.97 4.03-9 9-9 3.32 0 6.23 1.8 7.81 4.47"
                      ></path>
                    </svg>
                  {/if}
                </button>
              {/if}
              <button
                class="icon-btn"
                onclick={(e) => handleAddRequest(e, collection.name)}
                title="Add request"
                aria-label="Add request"
              >
                +
              </button>
              <button
                class="icon-btn"
                onclick={(e) => toggleMenu(e, collection.name)}
                title="More actions"
                aria-label="More actions"
              >
                ...
              </button>
            </div>

            {#if activeMenu === collection.name}
              <div class="collection-menu">
                <button
                  class="menu-item"
                  onclick={(e) => {
                    e.stopPropagation();
                    openRenameCollection(collection.name);
                  }}
                >
                  Rename
                </button>
                <button
                  class="menu-item danger"
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
            <div class="requests">
              {#if getVisibleRequests(collection, normalizedQuery).length === 0}
                <div class="empty-requests">
                  {isSearching ? "No matching requests" : "No requests yet"}
                </div>
              {:else}
                {#each getVisibleRequests(collection, normalizedQuery) as request (request.id)}
                  <div
                    class="request-item"
                    class:selected={selectedRequestId === request.id}
                    onclick={() => selectRequest(request.id, collection.name)}
                    onkeypress={(e) =>
                      e.key === "Enter" && selectRequest(request.id, collection.name)}
                    role="button"
                    tabindex="0"
                  >
                    <span class={`method-badge ${getMethodClass(request.verb)}`}>
                      {request.verb}
                    </span>
                    <span class="request-name">{request.name}</span>
                    <button
                      class="icon-btn subtle"
                      onclick={() => handleDeleteRequest(collection.name, request.id)}
                      title="Delete request"
                      aria-label="Delete request"
                    >
                      x
                    </button>
                  </div>
                {/each}
              {/if}
            </div>
          {/if}
        </div>
      {/each}
    </div>

    {#if filteredCollections.length === 0 && !$collectionStore.loading}
      <div class="empty-state">
        <p>
          {isSearching ? "No matching collections or requests" : "No collections yet"}
        </p>
        {#if !isSearching}
          <p class="hint">Create your first collection to get started</p>
        {/if}
      </div>
    {/if}
  {/if}
</div>

{#if showNewCollectionDialog}
  <Modal toggleFn={closeNewCollectionDialog}>
    <div class="dialog">
      <h3>New Collection</h3>
      <!-- svelte-ignore a11y_autofocus -->
      <input
        type="text"
        bind:value={newCollectionName}
        placeholder="Collection name"
        onkeydown={(e) => e.key === "Enter" && handleCreateCollection()}
        autofocus
      />
    </div>
    {#snippet additional_buttons()}
      <Button variant="primary" click={handleCreateCollection}>Create</Button>
    {/snippet}
  </Modal>
{/if}

{#if showRenameCollectionDialog}
  <Modal toggleFn={closeRenameDialog}>
    <div class="dialog">
      <h3>Rename Collection</h3>
      <!-- svelte-ignore a11y_autofocus -->
      <input
        type="text"
        bind:value={renameCollectionName}
        placeholder="Collection name"
        onkeydown={(e) => e.key === "Enter" && handleRenameCollection()}
        autofocus
      />
    </div>
    {#snippet additional_buttons()}
      <Button variant="primary" click={handleRenameCollection}>Save</Button>
    {/snippet}
  </Modal>
{/if}

{#if showDeleteConfirmDialog}
  <Modal toggleFn={closeDeleteConfirmDialog}>
    <div class="dialog">
      <h3>Delete Collection</h3>
      <p>Are you sure you want to delete "{deleteTarget}"?</p>
      <p class="warning">This action cannot be undone.</p>
    </div>
    {#snippet additional_buttons()}
      <Button variant="danger" click={confirmDelete}>Delete</Button>
    {/snippet}
  </Modal>
{/if}
{#if showDeleteRequestConfirmDialog}
  <Modal toggleFn={closeDeleteRequestConfirmDialog}>
    <div class="dialog">
      <h3>Delete Request</h3>
      <p>Are you sure you want to delete this request?</p>
      <p class="warning">This action cannot be undone.</p>
    </div>
    {#snippet additional_buttons()}
      <Button variant="danger" click={confirmDeleteRequest}>Delete</Button>
    {/snippet}
  </Modal>
{/if}

{#if showImportSelector}
  <Modal title="Import Collection" toggleFn={() => (showImportSelector = false)}>
    <div class="import-modal-body">
      <Tabs bind:activeValue={importActiveTab}>
        <Tab title="Postman" value="postman">
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
        </Tab>

        <Tab title="Bruno" value="bruno">
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
        </Tab>

        <Tab title="Git" value="git">
          <GitImportView onImported={() => (showImportSelector = false)} />
        </Tab>
      </Tabs>
    </div>

    {#snippet additional_buttons()}
      {#if importActiveTab === "postman"}
        <Button variant="primary" click={() => handleSelectImportFormat("postman")}>
          Select file…
        </Button>
      {:else if importActiveTab === "bruno"}
        <Button variant="primary" click={() => handleSelectImportFormat("bruno")}>
          Select folder…
        </Button>
      {/if}
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

<style>
  .collection-list {
    display: flex;
    flex-direction: column;
    height: 100%;
    /* width is now controlled by a style property */
    flex-shrink: 0;
    background: var(--bg-secondary);
    border-right: 1px solid var(--border);
    position: relative;
  }

  .resize-handle {
    position: absolute;
    top: 0;
    right: -4px;
    bottom: 0;
    width: 8px;
    cursor: col-resize;
    z-index: 50; /* Ensure it's on top */
  }

  .collection-list.collapsed {
    width: auto;
  }

  .header {
    padding: var(--space-md);
    border-bottom: 1px solid var(--border);
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
  }

  .header-title {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .header-actions {
    display: flex;
    align-items: center;
    gap: var(--space-xs);
  }

  .header h3 {
    margin: 0;
    font-size: var(--font-size-lg);
    font-weight: var(--font-weight-semibold);
  }

  .search-row {
    display: flex;
    align-items: center;
    gap: var(--space-xs);
  }

  .search-input {
    flex: 1;
  }

  .clear-search {
    width: 24px;
    height: 24px;
    border-radius: var(--radius-sm);
    border: 1px solid var(--border);
    background: var(--bg-tertiary);
    color: var(--text-muted);
    cursor: pointer;
    padding: 0;
  }

  .clear-search:hover {
    color: var(--text);
    border-color: var(--border-dark);
  }

  .loading {
    padding: var(--space-md);
    text-align: center;
    color: var(--text-muted);
  }

  .error {
    margin: var(--space-md);
    padding: var(--space-sm);
    background: var(--status-danger-bg);
    color: var(--status-danger-text);
    border-radius: var(--radius-md);
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: var(--font-size-sm);
  }

  .collections {
    flex: 1;
    overflow-y: auto;
    padding: var(--space-sm);
    display: flex;
    flex-direction: column;
    gap: var(--space-xs);
  }

  .collection-item {
    border-radius: var(--radius-md);
    background: var(--bg-primary);
    border: 1px solid transparent;
    overflow: visible;
  }

  .collection-item.selected {
    border-color: var(--primary);
    box-shadow: var(--shadow-sm);
  }

  .collection-item.menu-open {
    border-color: var(--border-dark);
  }

  .collection-header {
    display: flex;
    align-items: center;
    padding: var(--space-sm) var(--space-md);
    cursor: pointer;
    gap: var(--space-xs);
    position: relative;
    border-radius: var(--radius-md);
  }

  .collection-header:hover {
    background: var(--bg-tertiary);
  }

  .expand-btn {
    background: none;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    padding: 0;
    width: 20px;
    height: 20px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .expand-icon {
    display: inline-block;
    transition: transform var(--transition-fast);
  }

  .expand-icon.expanded {
    transform: rotate(90deg);
  }

  .collection-info {
    display: flex;
    align-items: center;
    gap: var(--space-xs);
    flex: 1;
    min-width: 0;
  }

  .provider-icon {
    flex-shrink: 0;
    color: var(--text-muted);
  }

  .collection-name {
    font-weight: var(--font-weight-medium);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .collection-count {
    color: var(--text-muted);
    font-size: var(--font-size-xs);
    background: var(--bg-tertiary);
    padding: 0 var(--space-xs);
    border-radius: var(--radius-sm);
  }

  .collection-actions {
    display: flex;
    gap: var(--space-xs);
    opacity: 0;
    pointer-events: none;
    transition: opacity var(--transition-fast);
  }

  .collection-header:hover .collection-actions,
  .collection-item.menu-open .collection-actions {
    opacity: 1;
    pointer-events: auto;
  }

  .icon-btn {
    background: none;
    border: 1px solid transparent;
    cursor: pointer;
    padding: 0 var(--space-xs);
    border-radius: var(--radius-sm);
    color: var(--text-muted);
    transition: all var(--transition-fast);
    font-size: var(--font-size-sm);
    height: 24px;
  }

  .icon-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .sync-spinner {
    display: block;
    width: 10px;
    height: 10px;
    border: 1.5px solid var(--border);
    border-top-color: var(--primary);
    border-radius: 50%;
    animation: spin 0.7s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .icon-btn.subtle {
    opacity: 0;
  }

  .request-item:hover .icon-btn.subtle {
    opacity: 1;
  }

  .collection-menu {
    position: absolute;
    right: var(--space-sm);
    top: calc(100% + 6px);
    background: var(--bg-primary);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-md);
    display: flex;
    flex-direction: column;
    min-width: 140px;
    z-index: var(--z-dropdown);
  }

  .menu-item {
    padding: var(--space-sm) var(--space-md);
    background: none;
    border: none;
    text-align: left;
    font-size: var(--font-size-sm);
    color: var(--text);
    cursor: pointer;
  }

  .menu-item:hover {
    background: var(--bg-tertiary);
  }

  .menu-item.danger {
    color: var(--danger);
  }

  .menu-item.danger:hover {
    background: var(--status-danger-bg);
  }

  .requests {
    background: var(--bg-secondary);
    padding: 0 var(--space-sm) var(--space-sm) calc(var(--space-lg) + 8px);
    display: flex;
    flex-direction: column;
    gap: var(--space-xs);
  }

  .request-item {
    display: flex;
    align-items: center;
    padding: var(--space-xs) var(--space-sm);
    cursor: pointer;
    gap: var(--space-sm);
    border-radius: var(--radius-sm);
    transition: background-color var(--transition-fast);
  }

  .request-item:hover {
    background: var(--bg-tertiary);
  }

  .request-item.selected {
    background: var(--status-info-bg);
  }

  .method-badge {
    font-family: var(--font-mono);
    font-size: var(--font-size-xs);
    font-weight: var(--font-weight-semibold);
    min-width: 48px;
    text-align: center;
    padding: 2px var(--space-xs);
    border-radius: var(--radius-sm);
  }

  .method-badge.method-get {
    background: var(--method-get-bg);
    color: var(--method-get-text);
  }

  .method-badge.method-post {
    background: var(--method-post-bg);
    color: var(--method-post-text);
  }

  .method-badge.method-put {
    background: var(--method-put-bg);
    color: var(--method-put-text);
  }

  .method-badge.method-delete {
    background: var(--method-delete-bg);
    color: var(--method-delete-text);
  }

  .method-badge.method-patch {
    background: var(--method-patch-bg);
    color: var(--method-patch-text);
  }

  .method-badge.method-head,
  .method-badge.method-options {
    background: var(--bg-tertiary);
    color: var(--text-muted);
  }

  .request-name {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .empty-requests {
    color: var(--text-muted);
    font-size: var(--font-size-xs);
    padding: var(--space-xs) var(--space-sm);
  }

  .empty-state {
    padding: var(--space-xl);
    text-align: center;
    color: var(--text-muted);
  }

  .empty-state p {
    margin: var(--space-xs) 0;
  }

  .empty-state .hint {
    font-size: var(--font-size-sm);
  }

  .dialog h3 {
    margin: 0 0 var(--space-md) 0;
    font-size: var(--font-size-xl);
  }

  .dialog p {
    margin: 0 0 var(--space-sm) 0;
    color: var(--text);
  }

  .dialog .warning {
    color: var(--text-muted);
    font-size: var(--font-size-sm);
    margin-bottom: var(--space-md);
  }

  .format-buttons {
    display: flex;
    gap: var(--space-md);
    justify-content: center;
  }

  .dialog input {
    width: 100%;
    padding: var(--space-sm);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    background: var(--bg-secondary);
    color: var(--text);
    font-size: var(--font-size-md);
    margin-bottom: var(--space-md);
  }

  .dialog input:focus {
    outline: none;
    border-color: var(--primary);
  }

  .dialog-actions {
    display: flex;
    gap: var(--space-sm);
    justify-content: flex-end;
  }

  /* ── Import Modal ──────────────────────────────────────────── */

  .import-modal-body {
    /* remove default dialog-content padding so the Tabs bar touches the edges */
    margin: calc(-1 * var(--space-lg));
  }
</style>
