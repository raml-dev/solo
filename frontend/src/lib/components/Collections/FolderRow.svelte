<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import FolderRow from "$src/lib/components/Collections/FolderRow.svelte";
  import RequestRow from "$src/lib/components/Collections/RequestRow.svelte";
  import ContextMenu from "$src/lib/components/common/ContextMenu.svelte";
  import ContextMenuDivider from "$src/lib/components/common/ContextMenuDivider.svelte";
  import ContextMenuItem from "$src/lib/components/common/ContextMenuItem.svelte";
  import {
    collectionTreeUI,
    collectionTreeUIState
  } from "$src/lib/features/collections/collectionTreeUI.svelte";
  import { collectionStore } from "$src/lib/stores/collectionStore.svelte";
  import { clampNumberToMax, getTotalRequestCount } from "$src/lib/utils/helpers";
  import { collection } from "$wails/go/models";
  import AngleDownOutline from "flowbite-svelte-icons/AngleDownOutline.svelte";
  import AngleRightOutline from "flowbite-svelte-icons/AngleRightOutline.svelte";
  import DotsHorizontalOutline from "flowbite-svelte-icons/DotsHorizontalOutline.svelte";
  import FolderOpenOutline from "flowbite-svelte-icons/FolderOpenOutline.svelte";
  import FolderOutline from "flowbite-svelte-icons/FolderOutline.svelte";
  import PlusOutline from "flowbite-svelte-icons/PlusOutline.svelte";
  import Input from "flowbite-svelte/Input.svelte";
  import { tick } from "svelte";
  import { SvelteSet } from "svelte/reactivity";

  const OUTLINE_BUTTON_CLASSES =
    "text-neutral-800/70 hover:text-neutral-800 dark:text-neutral-100/70 dark:hover:text-neutral-100";

  interface Props {
    folder: collection.Folder;
    collectionName: string;
    searchQuery: string;
    isSearching: boolean;
    selectedRequestId: string | null;
    expandedFolders: SvelteSet<string>;
    onAddRequestToFolder: (collectionName: string, folderId: string) => void;
    onCreateFolder: (collectionName: string, parentFolderId: string | null) => void;
    onRenameFolder: (collectionName: string, folder: collection.Folder) => void;
    onDeleteFolder: (collectionName: string, folder: collection.Folder) => void;
    onDeleteRequest: (collectionName: string, requestId: string) => void;
    onRequestSelect?: (requestId: string) => void;
  }

  let {
    folder,
    collectionName,
    searchQuery,
    isSearching,
    selectedRequestId,
    expandedFolders,
    onAddRequestToFolder,
    onCreateFolder,
    onRenameFolder,
    onDeleteFolder,
    onDeleteRequest,
    onRequestSelect = () => {}
  }: Props = $props();

  let editingNameInputEl: HTMLInputElement | undefined = $state();
  let isEditing = $state(false);
  let editingName = $state("");
  let requestDragState = $derived(collectionTreeUIState.requestDrag);
  let folderRenameState = $derived(collectionTreeUIState.folderRename);
  let anyContextMenuOpen = $derived(
    collectionTreeUIState.collectionContextMenu.open ||
      collectionTreeUIState.requestContextMenu.open ||
      collectionTreeUIState.folderContextMenu.open
  );

  function normalize(value: string | undefined | null): string {
    return (value || "").toLowerCase();
  }

  function requestMatches(request: collection.Request, query: string): boolean {
    if (!query) return true;
    return normalize(request.name).includes(query) || normalize(request.url).includes(query);
  }

  function folderMatches(currentFolder: collection.Folder, query: string): boolean {
    if (!query) return true;
    return normalize(currentFolder.name).includes(query);
  }

  function shouldShowFolder(currentFolder: collection.Folder, query: string): boolean {
    if (!query) return true;
    if (folderMatches(currentFolder, query)) return true;
    if ((currentFolder.requests || []).some((request) => requestMatches(request, query)))
      return true;
    return (currentFolder.folders || []).some((subfolder) => shouldShowFolder(subfolder, query));
  }

  function getVisibleFolders(folders: collection.Folder[], query: string): collection.Folder[] {
    if (!query) return folders || [];
    return (folders || []).filter((currentFolder) => shouldShowFolder(currentFolder, query));
  }

  function getVisibleRequests(requests: collection.Request[], query: string): collection.Request[] {
    if (!query) return requests || [];
    return (requests || []).filter((request) => requestMatches(request, query));
  }

  function getVisibleSubfolders(): collection.Folder[] {
    return getVisibleFolders(folder.folders || [], searchQuery);
  }

  function getVisibleFolderRequests(): collection.Request[] {
    return getVisibleRequests(folder.requests || [], searchQuery);
  }

  function isExpanded(): boolean {
    return isSearching || expandedFolders.has(folder.id);
  }

  function toggleFolder(event: Event) {
    event.stopPropagation();

    if (isSearching) return;

    if (expandedFolders.has(folder.id)) {
      expandedFolders.delete(folder.id);
      return;
    }

    expandedFolders.add(folder.id);
  }

  function isContextMenuGesture(event: MouseEvent): boolean {
    return event.button !== 0 || event.ctrlKey;
  }

  function handleFolderActivate(event: MouseEvent) {
    if (isEditing) {
      return;
    }

    if (isContextMenuGesture(event)) {
      return;
    }

    if (anyContextMenuOpen) {
      collectionTreeUI.closeAllContextMenus();
      return;
    }

    toggleFolder(event);
  }

  function isNoDragTarget(target: EventTarget | null): boolean {
    return target instanceof HTMLElement && target.closest('[data-no-drag="true"]') !== null;
  }

  function closeContextMenu() {
    collectionTreeUI.closeFolderContextMenu();
  }

  function getMenuTriggerId(): string {
    return `folder-menu-${folder.id}`;
  }

  async function handleContextMenu(event: MouseEvent) {
    if (requestDragState.sourceRequestId || isNoDragTarget(event.target)) {
      return;
    }

    event.preventDefault();
    event.stopPropagation();
    closeContextMenu();
    collectionTreeUI.stageFolderContextMenu(
      folder,
      folder.id,
      collectionName,
      event.clientX,
      event.clientY
    );
    await tick();
    collectionTreeUI.showFolderContextMenu();
  }

  async function beginRename() {
    isEditing = true;
    editingName = folder.name || "";

    await tick();
    editingNameInputEl?.focus();
    editingNameInputEl?.select();
  }

  async function startRename(event: Event) {
    event.stopPropagation();
    await beginRename();
  }

  function cancelRename() {
    isEditing = false;
    editingName = "";
  }

  async function commitRename() {
    if (!isEditing) return;

    const nextName = editingName.trim();
    if (!nextName || nextName === folder.name) {
      cancelRename();
      return;
    }

    try {
      await collectionStore.updateFolder(
        collectionName,
        collection.Folder.createFrom({
          ...folder,
          name: nextName,
          lastUpdateTimestamp: new Date().toISOString()
        })
      );
    } catch {
      // error already shown by store
    } finally {
      cancelRename();
    }
  }

  function isFolderRequestDropTarget(): boolean {
    return (
      requestDragState.targetMode === "container" &&
      requestDragState.targetCollectionName === collectionName &&
      requestDragState.targetParentFolderId === folder.id
    );
  }

  function handleFolderRequestDragOver(event: DragEvent) {
    if (
      !requestDragState.sourceRequestId ||
      requestDragState.sourceCollectionName !== collectionName ||
      isSearching
    ) {
      return;
    }

    event.preventDefault();
    event.stopPropagation();
    collectionTreeUI.setRequestContainerTarget(collectionName, folder.id);

    if (event.dataTransfer) {
      event.dataTransfer.dropEffect = "move";
    }
  }

  async function handleFolderRequestDrop(event: DragEvent) {
    event.preventDefault();
    event.stopPropagation();

    if (
      !requestDragState.sourceRequestId ||
      requestDragState.sourceCollectionName !== collectionName ||
      requestDragState.targetMode !== "container" ||
      requestDragState.targetParentFolderId !== folder.id
    ) {
      collectionTreeUI.resetRequestDrag();
      return;
    }

    try {
      await collectionStore.moveRequest(
        collectionName,
        requestDragState.sourceRequestId,
        requestDragState.sourceParentFolderId,
        folder.id
      );
    } finally {
      collectionTreeUI.resetRequestDrag();
    }
  }

  $effect(() => {
    if (
      folderRenameState.folderId !== folder.id ||
      folderRenameState.collectionName !== collectionName
    ) {
      return;
    }

    collectionTreeUI.consumeFolderRename();
    void beginRename();
  });
</script>

{#snippet actionsDropdown(triggeredBy: string, isOpen: boolean | undefined, onClose: () => void)}
  <ContextMenu {triggeredBy} {isOpen} menuClass="z-50 w-44" {onClose}>
    <ContextMenuItem
      onclick={() => {
        onAddRequestToFolder(collectionName, folder.id);
        onClose();
      }}
    >
      New request
    </ContextMenuItem>
    <ContextMenuItem
      onclick={() => {
        onCreateFolder(collectionName, folder.id);
        onClose();
      }}
    >
      New subfolder
    </ContextMenuItem>
    <ContextMenuDivider />
    <ContextMenuItem
      onclick={(event) => {
        void startRename(event);
        onClose();
      }}
    >
      Rename
    </ContextMenuItem>
    <ContextMenuItem
      className="text-danger-600 hover:bg-danger-50 dark:text-danger-400 dark:hover:bg-danger-900/20"
      onclick={() => {
        onDeleteFolder(collectionName, folder);
        onClose();
      }}
    >
      Delete
    </ContextMenuItem>
  </ContextMenu>
{/snippet}

<div class="relative space-y-1">
  {#if isFolderRequestDropTarget()}
    <div class="pointer-events-none absolute inset-x-2 top-0 h-0.5 rounded bg-primary-500"></div>
    <div class="pointer-events-none absolute inset-x-2 bottom-0 h-0.5 rounded bg-primary-500"></div>
  {/if}

  <div
    class={`group flex items-center gap-2 rounded px-2 py-1.5 hover:bg-neutral-100 dark:hover:bg-neutral-700/60 ${isFolderRequestDropTarget() ? "bg-primary-50/60 ring-1 ring-primary-400/60 ring-inset dark:bg-primary-900/10 dark:ring-primary-700/70" : ""}`}
    onclick={handleFolderActivate}
    onkeypress={(event) => event.key === "Enter" && toggleFolder(event)}
    oncontextmenu={(event) => void handleContextMenu(event)}
    ondragenter={handleFolderRequestDragOver}
    ondragover={handleFolderRequestDragOver}
    ondrop={(event) => void handleFolderRequestDrop(event)}
    role="button"
    tabindex="0"
  >
    <button
      class="h-6 w-4 p-0 text-xs hover:cursor-pointer dark:text-white"
      onclick={toggleFolder}
      aria-label="Toggle folder"
    >
      {#if isExpanded()}
        <AngleDownOutline class={`h-3 w-3 ${OUTLINE_BUTTON_CLASSES}`} />
      {:else}
        <AngleRightOutline class={`h-3 w-3 ${OUTLINE_BUTTON_CLASSES}`} />
      {/if}
    </button>

    {#if isExpanded()}
      <FolderOpenOutline class={`h-4 w-4 shrink-0 ${OUTLINE_BUTTON_CLASSES}`} />
    {:else}
      <FolderOutline class={`h-4 w-4 shrink-0 ${OUTLINE_BUTTON_CLASSES}`} />
    {/if}

    <span
      class="w-8 rounded bg-neutral-200 px-1.5 py-0.5 text-center font-mono text-xs text-neutral-600 dark:bg-neutral-700 dark:text-neutral-300"
    >
      {clampNumberToMax(getTotalRequestCount(folder))}
    </span>

    {#if isEditing}
      <div data-no-drag="true" class="min-w-0 flex-1">
        <Input
          type="text"
          size="sm"
          class="min-w-0 flex-1"
          bind:value={editingName}
          bind:elementRef={editingNameInputEl}
          autofocus
          onclick={(event) => event.stopPropagation()}
          onmousedown={(event) => event.stopPropagation()}
          onkeydown={(event) => {
            event.stopPropagation();
            if (event.key === "Enter") {
              void commitRename();
            }
            if (event.key === "Escape") {
              cancelRename();
            }
          }}
          onblur={() => void commitRename()}
        />
      </div>
    {:else}
      <span
        class="min-w-0 flex-1 truncate text-sm text-neutral-800 dark:text-neutral-100"
        role="button"
        tabindex="0"
        onkeydown={(event) => event.key === "Enter" && startRename(event)}
      >
        {folder.name}
      </span>
    {/if}

    <button
      data-no-drag="true"
      class="invisible group-hover:visible hover:cursor-pointer"
      onclick={(event: MouseEvent) => {
        event.stopPropagation();
        onAddRequestToFolder(collectionName, folder.id);
      }}
      title="Add request"
      aria-label="Add request"
    >
      <PlusOutline class={`h-3 w-3 ${OUTLINE_BUTTON_CLASSES}`} />
    </button>
    <button
      data-no-drag="true"
      id={getMenuTriggerId()}
      class="invisible group-hover:visible hover:cursor-pointer"
      title="Folder actions"
      aria-label="Folder actions"
      onclick={(event: MouseEvent) => {
        event.stopPropagation();
        closeContextMenu();
      }}
      oncontextmenu={(event: MouseEvent) => {
        event.preventDefault();
        event.stopPropagation();
      }}
    >
      <DotsHorizontalOutline class={`h-3 w-3 ${OUTLINE_BUTTON_CLASSES}`} />
    </button>
    {@render actionsDropdown(`#${getMenuTriggerId()}`, undefined, closeContextMenu)}
  </div>

  {#if isExpanded()}
    <div
      class="space-y-1 border-l border-neutral-200 pl-4 dark:border-neutral-700"
      role="group"
      aria-label={`${folder.name} items`}
    >
      {#each getVisibleSubfolders() as subfolder (subfolder.id)}
        <FolderRow
          folder={subfolder}
          {collectionName}
          {searchQuery}
          {isSearching}
          {selectedRequestId}
          {expandedFolders}
          {onAddRequestToFolder}
          {onCreateFolder}
          {onRenameFolder}
          {onDeleteFolder}
          {onDeleteRequest}
          {onRequestSelect}
        />
      {/each}

      {#each getVisibleFolderRequests() as request (request.id)}
        <RequestRow
          {request}
          {collectionName}
          parentFolderId={folder.id}
          selected={selectedRequestId === request.id}
          {isSearching}
          dragEnabled
          {onDeleteRequest}
          {onRequestSelect}
          onAnyMenuOpen={closeContextMenu}
        />
      {/each}

      {#if getVisibleSubfolders().length === 0 && getVisibleFolderRequests().length === 0}
        <div
          class={`rounded px-2 py-2 text-xs transition-colors ${isFolderRequestDropTarget() ? "bg-primary-50 text-primary-700 dark:bg-primary-900/20 dark:text-primary-300" : "text-neutral-500 dark:text-neutral-400"}`}
          ondragenter={handleFolderRequestDragOver}
          ondragover={handleFolderRequestDragOver}
          ondrop={(event) => void handleFolderRequestDrop(event)}
          role="group"
          aria-label={`Empty state for ${folder.name}`}
        >
          {isSearching ? "No matching items" : "Folder is empty"}
        </div>
      {/if}
    </div>
  {/if}
</div>
