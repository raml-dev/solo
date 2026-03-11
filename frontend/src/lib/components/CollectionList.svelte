<script lang="ts">
  import { onDestroy, onMount } from "svelte";
  import { collectionStore } from "../stores/collectionStore";
  import { tabStore } from "../stores/tabStore";
  import { notifications } from "../stores/notificationStore";
  import Button from "./base/Button.svelte";
  import Modal from "./base/Modal.svelte";
  import Tabs from "./base/Tabs.svelte";
  import Tab from "./base/Tab.svelte";
  import DropZone from "./base/DropZone.svelte";
  import type { collection } from "../../../wailsjs/go/models";
  import {
    ImportPostmanCollection,
    SelectFile,
    ImportBrunoCollection,
    SelectDirectory
  } from "../../../wailsjs/go/main/App";

  export let onRequestSelect: (requestId: string) => void = () => {};

  let showNewCollectionDialog = false;
  let showRenameCollectionDialog = false;
  let showDeleteConfirmDialog = false;
  let showDeleteRequestConfirmDialog = false;
  let showImportSelector = false;

  let importActiveTab = "postman";


  let newCollectionName = "";
  let renameCollectionName = "";
  let renameTarget: string | null = null;
  let deleteTarget: string | null = null;
  let deleteRequestTarget: string | null = null;
  let deleteRequestCollectionName: string | null = null;
  let expandedCollections: Set<string> = new Set();
  let searchQuery = "";
  let activeMenu: string | null = null;
  let isCollapsed = false;

  let sidebarWidth = 280; // Default width
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

  $: collections = $collectionStore.collections;
  $: selectedCollectionName = $collectionStore.selectedCollectionName;
  // Highlight in sidebar is driven by the active tab, not the collectionStore selection
  $: selectedRequestId = $tabStore.tabs.find((t) => t.id === $tabStore.activeTabId)?.requestId ?? null;
  $: normalizedQuery = searchQuery.trim().toLowerCase();
  $: isSearching = normalizedQuery.length > 0;
  $: filteredCollections = collections.filter((collection) =>
    shouldShowCollection(collection, normalizedQuery)
  );

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
    expandedCollections = new Set(expandedCollections);
  }

  function selectCollection(name: string) {
    collectionStore.selectCollection(name);
  }

  function selectRequest(requestId: string, collectionName: string) {
    // Find the request data to pass metadata to tabStore
    const coll = $collectionStore.collections.find((c) => c.name === collectionName);
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
      settings: req.settings || {}
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
    } catch (err) {
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
        expandedCollections = new Set(expandedCollections);
      }
      closeRenameDialog();
    } catch (err) {
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
      expandedCollections = new Set(expandedCollections);
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
      expandedCollections = new Set(expandedCollections);

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
    } catch (err) {
      // error already shown by store
    }
  }

  async function handleDeleteRequest(collectionName: string, requestId: string) {
    deleteRequestCollectionName = collectionName;
    deleteRequestTarget = requestId;
    showDeleteRequestConfirmDialog = true;
  }

  async function confirmDeleteRequest() {
    if (!deleteRequestTarget) return;

    try {
      await collectionStore.removeRequest(deleteRequestCollectionName, deleteRequestTarget);
      closeDeleteRequestConfirmDialog();
    } catch (err) {
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
</script>

<div class="collection-list" class:collapsed={isCollapsed} style={`width: ${isCollapsed ? 'auto' : sidebarWidth + 'px'};`}>
  <div class="resize-handle" on:mousedown={startResize} />
  <div class="header">
    <div class="header-title">
      {#if !isCollapsed}
        <h3>Collections</h3>
      {/if}
      <div class="header-actions">
        {#if !isCollapsed}
          <Button variant="secondary" size="small" click={openImportModal}>
            Import
          </Button>
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
          <button
            class="clear-search"
            on:click={() => (searchQuery = "")}
            aria-label="Clear search"
          >
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
            on:click={(e) => {
              e.stopPropagation();
              selectCollection(collection.name);
              toggleCollection(collection.name);
            }}
            on:keypress={(e) => {
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
              on:click={(e) => {
                e.stopPropagation();
                toggleCollection(collection.name);
              }}
              aria-label="Toggle collection"
            >
              <span class="expand-icon" class:expanded={isExpanded(collection.name)}> &gt; </span>
            </button>

            <div class="collection-info">
              <span class="collection-name">{collection.name}</span>
              <span class="collection-count">{collection.requests?.length || 0}</span>
            </div>

            <div class="collection-actions">
              <button
                class="icon-btn"
                on:click={(e) => handleAddRequest(e, collection.name)}
                title="Add request"
                aria-label="Add request"
              >
                +
              </button>
              <button
                class="icon-btn"
                on:click={(e) => toggleMenu(e, collection.name)}
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
                  on:click={(e) => {
                    e.stopPropagation();
                    openRenameCollection(collection.name);
                  }}
                >
                  Rename
                </button>
                <button
                  class="menu-item danger"
                  on:click={(e) => {
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
                    on:click={() => selectRequest(request.id, collection.name)}
                    on:keypress={(e) => e.key === "Enter" && selectRequest(request.id, collection.name)}
                    role="button"
                    tabindex="0"
                  >
                    <span class={`method-badge ${getMethodClass(request.verb)}`}>
                      {request.verb}
                    </span>
                    <span class="request-name">{request.name}</span>
                    <button
                      class="icon-btn subtle"
                      on:click={() => handleDeleteRequest(collection.name, request.id)}
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
      <!-- svelte-ignore a11y-autofocus -->
      <input
        type="text"
        bind:value={newCollectionName}
        placeholder="Collection name"
        on:keydown={(e) => e.key === "Enter" && handleCreateCollection()}
        autofocus
      />
    </div>
    <svelte:fragment slot="additional-buttons">
      <Button variant="primary" click={handleCreateCollection}>Create</Button>
    </svelte:fragment>
  </Modal>
{/if}

{#if showRenameCollectionDialog}
  <Modal toggleFn={closeRenameDialog}>
    <div class="dialog">
      <h3>Rename Collection</h3>
      <!-- svelte-ignore a11y-autofocus -->
      <input
        type="text"
        bind:value={renameCollectionName}
        placeholder="Collection name"
        on:keydown={(e) => e.key === "Enter" && handleRenameCollection()}
        autofocus
      />
    </div>
    <svelte:fragment slot="additional-buttons">
      <Button variant="primary" click={handleRenameCollection}>Save</Button>
    </svelte:fragment>
  </Modal>
{/if}

{#if showDeleteConfirmDialog}
  <Modal toggleFn={closeDeleteConfirmDialog}>
    <div class="dialog">
      <h3>Delete Collection</h3>
      <p>Are you sure you want to delete "{deleteTarget}"?</p>
      <p class="warning">This action cannot be undone.</p>
    </div>
    <svelte:fragment slot="additional-buttons">
      <Button variant="danger" click={confirmDelete}>Delete</Button>
    </svelte:fragment>
  </Modal>
{/if}
{#if showDeleteRequestConfirmDialog}
  <Modal toggleFn={closeDeleteRequestConfirmDialog}>
    <div class="dialog">
      <h3>Delete Request</h3>
      <p>Are you sure you want to delete this request?</p>
      <p class="warning">This action cannot be undone.</p>
    </div>
    <svelte:fragment slot="additional-buttons">
      <Button variant="danger" click={confirmDeleteRequest}>Delete</Button>
    </svelte:fragment>
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
            on:drop={async (e) => {
              const paths = e.detail.paths;
              showImportSelector = false;
              if (paths.length > 0) {
                await ImportPostmanCollection(paths[0]);
                await collectionStore.loadCollections();
              } else {
                await handleImportPostman();
              }
            }}
          >
            <svelte:fragment slot="icon">
              <svg width="44" height="44" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round">
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                <polyline points="17 8 12 3 7 8"/>
                <line x1="12" y1="3" x2="12" y2="15"/>
              </svg>
            </svelte:fragment>
          </DropZone>
        </Tab>

        <Tab title="Bruno" value="bruno">
          <DropZone
            title="Drop your Bruno collection folder here"
            subtitle="Supports Bruno collection folders (.bru files)"
            on:drop={async (e) => {
              const paths = e.detail.paths;
              showImportSelector = false;
              if (paths.length > 0) {
                await ImportBrunoCollection(paths[0]);
                await collectionStore.loadCollections();
              } else {
                await handleImportBruno();
              }
            }}
          >
            <svelte:fragment slot="icon">
              <svg width="44" height="44" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round">
                <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>
                <polyline points="9 22 9 12 15 12 15 22"/>
              </svg>
            </svelte:fragment>
          </DropZone>
        </Tab>
      </Tabs>
    </div>

    <svelte:fragment slot="additional-buttons">
      {#if importActiveTab === "postman"}
        <Button variant="primary" click={() => handleSelectImportFormat("postman")}>
          Select file…
        </Button>
      {:else}
        <Button variant="primary" click={() => handleSelectImportFormat("bruno")}>
          Select folder…
        </Button>
      {/if}
    </svelte:fragment>
  </Modal>
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

  .error button {
    background: none;
    border: none;
    color: inherit;
    font-size: var(--font-size-lg);
    cursor: pointer;
    padding: 0 var(--space-xs);
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

  .icon-btn:hover {
    background: var(--bg-tertiary);
    color: var(--text);
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
