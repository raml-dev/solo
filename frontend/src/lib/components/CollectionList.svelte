<script lang="ts">
  import { onDestroy, onMount } from "svelte";
  import { collectionStore } from "../stores/collectionStore";
  import Button from "./base/Button.svelte";
  import Modal from "./base/Modal.svelte";
  import type { collection } from "../../../wailsjs/go/models";

  export let onRequestSelect: (requestId: string) => void = () => {};

  let showNewCollectionDialog = false;
  let showRenameCollectionDialog = false;
  let showDeleteConfirmDialog = false;
  let showDeleteRequestConfirmDialog = false;
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

  $: collections = $collectionStore.collections;
  $: selectedCollectionName = $collectionStore.selectedCollectionName;
  $: selectedRequestId = $collectionStore.selectedRequestId;
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

  function selectRequest(requestId: string) {
    collectionStore.selectRequest(requestId);
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
    if (!trimmed) {
      return;
    }

    const exists = collections.some(
      (collection) => collection.name.toLowerCase() === trimmed.toLowerCase()
    );
    if (exists) {
      alert(`Collection "${trimmed}" already exists.`);
      return;
    }

    try {
      await collectionStore.createCollection(trimmed);
      closeNewCollectionDialog();
    } catch (err) {
      console.error("Error creating collection:", err);
      alert(`Error creating collection: ${err}`);
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
      alert(`Collection "${trimmed}" already exists.`);
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
      console.error("Error renaming collection:", err);
      alert(`Error renaming collection: ${err}`);
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
      await collectionStore.addRequest(collectionName, {
        name: "New Request",
        url: "",
        verb: "GET"
      });

      expandedCollections.add(collectionName);
      expandedCollections = new Set(expandedCollections);
      collectionStore.selectCollection(collectionName);
    } catch (err) {
      console.error("Error adding request:", err);
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
      console.error("Error deleting request:", err);
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

  onMount(() => {
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

<div class="collection-list" class:collapsed={isCollapsed}>
  <div class="header">
    <div class="header-title">
      {#if !isCollapsed}
        <h3>Collections</h3>
      {/if}
      <div class="header-actions">
        {#if !isCollapsed}
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

    {#if $collectionStore.error}
      <div class="error">
        {$collectionStore.error}
        <button on:click={() => collectionStore.clearError()}>x</button>
      </div>
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
                    on:click={() => selectRequest(request.id)}
                    on:keypress={(e) => e.key === "Enter" && selectRequest(request.id)}
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

<style>
  .collection-list {
    display: flex;
    flex-direction: column;
    height: 100%;
    width: var(--sidebar-width);
    flex-shrink: 0;
    background: var(--bg-secondary);
    border-right: 1px solid var(--border);
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
</style>
