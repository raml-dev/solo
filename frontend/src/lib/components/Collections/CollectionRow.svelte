<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import FolderRow from "$src/lib/components/Collections/FolderRow.svelte";
  import RequestRow from "$src/lib/components/Collections/RequestRow.svelte";
  import ContextMenu from "$src/lib/components/common/ContextMenu.svelte";
  import ContextMenuItem from "$src/lib/components/common/ContextMenuItem.svelte";
  import {
    collectionTreeUI,
    collectionTreeUIState
  } from "$src/lib/features/collections/collectionTreeUI.svelte";
  import { collectionStore } from "$src/lib/stores/collectionStore.svelte";
  import { COLLECTION_OUTLINE_BUTTON_CLASSES } from "$src/lib/utils/constants";
  import { clampNumberToMax, getTotalRequestCount } from "$src/lib/utils/helpers";
  import { collection } from "$wails/go/models";
  import AdjustmentsVerticalOutline from "flowbite-svelte-icons/AdjustmentsVerticalOutline.svelte";
  import AngleDownOutline from "flowbite-svelte-icons/AngleDownOutline.svelte";
  import AngleRightOutline from "flowbite-svelte-icons/AngleRightOutline.svelte";
  import DotsHorizontalOutline from "flowbite-svelte-icons/DotsHorizontalOutline.svelte";
  import PlusOutline from "flowbite-svelte-icons/PlusOutline.svelte";
  import { tick } from "svelte";
  import { SvelteSet } from "svelte/reactivity";

  interface Props {
    collection: collection.Collection;
    expanded: boolean;
    searchQuery: string;
    isSearching: boolean;
    selectedRequestId: string | null;
    visibleFolders: collection.Folder[];
    visibleRequests: collection.Request[];
    expandedFolders: SvelteSet<string>;
    syncing: boolean;
    providerIconPath: string;
    onActivateCollection: (collectionName: string) => void;
    onToggleCollection: (collectionName: string) => void;
    onAddRequest: (collectionName: string) => void;
    onAddRequestToFolder: (collectionName: string, folderId: string) => void;
    onCreateFolder: (collectionName: string, parentFolderId: string | null) => void;
    onRenameFolder: (collectionName: string, folder: collection.Folder) => void;
    onDeleteFolder: (collectionName: string, folder: collection.Folder) => void;
    onDeleteRequest: (collectionName: string, requestId: string) => void;
    onOpenGitStatus: (currentCollection: collection.Collection) => void;
    onSync: (collectionId: string) => void;
    onExportCollection: (collectionName: string) => void;
    onOpenVariables: (collectionName: string) => void;
    onRenameCollection: (collectionName: string) => void;
    onDeleteCollection: (collectionName: string) => void;
    onRequestSelect?: (requestId: string) => void;
  }

  let {
    collection: currentCollection,
    expanded,
    searchQuery,
    isSearching,
    selectedRequestId,
    visibleFolders,
    visibleRequests,
    expandedFolders,
    syncing,
    providerIconPath,
    onActivateCollection,
    onToggleCollection,
    onAddRequest,
    onAddRequestToFolder,
    onCreateFolder,
    onRenameFolder,
    onDeleteFolder,
    onDeleteRequest,
    onOpenGitStatus,
    onSync,
    onExportCollection,
    onOpenVariables,
    onRenameCollection,
    onDeleteCollection,
    onRequestSelect = () => {}
  }: Props = $props();

  let requestDragState = $derived(collectionTreeUIState.requestDrag);
  let hasCollectionVariables = $derived(Object.keys(currentCollection.variables ?? {}).length > 0);
  let anyContextMenuOpen = $derived(
    collectionTreeUIState.requestContextMenu.open ||
      collectionTreeUIState.folderContextMenu.open ||
      collectionTreeUIState.collectionContextMenu.open
  );

  function isNoDragTarget(target: EventTarget | null): boolean {
    return target instanceof HTMLElement && target.closest('[data-no-drag="true"]') !== null;
  }

  function closeContextMenu() {
    collectionTreeUI.closeCollectionContextMenu();
  }

  function getMenuTriggerId(): string {
    return `collection-menu-${currentCollection.id}`;
  }

  async function handleContextMenu(event: MouseEvent) {
    if (requestDragState.sourceRequestId || isNoDragTarget(event.target)) {
      return;
    }

    event.preventDefault();
    event.stopPropagation();
    closeContextMenu();
    collectionTreeUI.stageCollectionContextMenu(
      currentCollection,
      currentCollection.id,
      event.clientX,
      event.clientY
    );
    await tick();
    collectionTreeUI.showCollectionContextMenu();
  }

  function handleHeaderActivate(event: Event) {
    event.stopPropagation();

    onActivateCollection(currentCollection.name);
  }

  function isContextMenuGesture(event: MouseEvent): boolean {
    return event.button !== 0 || event.ctrlKey;
  }

  function handleHeaderClick(event: MouseEvent) {
    if (isContextMenuGesture(event)) {
      return;
    }

    if (anyContextMenuOpen) {
      collectionTreeUI.closeAllContextMenus();
      return;
    }

    handleHeaderActivate(event);
  }

  function isRootRequestDropTarget(): boolean {
    return (
      requestDragState.targetMode === "container" &&
      requestDragState.targetCollectionName === currentCollection.name &&
      requestDragState.targetParentFolderId === null
    );
  }

  function handleRootRequestDragOver(event: DragEvent) {
    if (
      !requestDragState.sourceRequestId ||
      requestDragState.sourceCollectionName !== currentCollection.name ||
      isSearching
    ) {
      return;
    }

    event.preventDefault();
    event.stopPropagation();
    collectionTreeUI.setRequestContainerTarget(currentCollection.name, null);

    if (event.dataTransfer) {
      event.dataTransfer.dropEffect = "move";
    }
  }

  async function handleRootRequestDrop(event: DragEvent) {
    event.preventDefault();
    event.stopPropagation();

    if (
      !requestDragState.sourceRequestId ||
      requestDragState.sourceCollectionName !== currentCollection.name ||
      requestDragState.targetMode !== "container" ||
      requestDragState.targetParentFolderId !== null
    ) {
      collectionTreeUI.resetRequestDrag();
      return;
    }

    try {
      await collectionStore.moveRequest(
        currentCollection.name,
        requestDragState.sourceRequestId,
        requestDragState.sourceParentFolderId,
        null
      );
    } finally {
      collectionTreeUI.resetRequestDrag();
    }
  }
</script>

{#snippet actionsDropdown(triggeredBy: string, isOpen: boolean | undefined, onClose: () => void)}
  <ContextMenu {triggeredBy} {isOpen} {onClose}>
    <ContextMenuItem
      onclick={() => {
        onCreateFolder(currentCollection.name, null);
        onClose();
      }}
    >
      New folder
    </ContextMenuItem>

    {#if currentCollection.gitRemote}
      <ContextMenuItem
        onclick={() => {
          onOpenGitStatus(currentCollection);
          onClose();
        }}
      >
        Git status
      </ContextMenuItem>
      <ContextMenuItem
        disabled={syncing}
        onclick={() => {
          onSync(currentCollection.id);
          onClose();
        }}
      >
        {syncing ? "Syncing..." : "Sync with Git"}
      </ContextMenuItem>
    {/if}
    <ContextMenuItem
      onclick={() => {
        onOpenVariables(currentCollection.name);
        onClose();
      }}
    >
      Variables
    </ContextMenuItem>

    <ContextMenuItem
      onclick={() => {
        onExportCollection(currentCollection.name);
        onClose();
      }}
    >
      Export
    </ContextMenuItem>

    <ContextMenuItem
      onclick={() => {
        onRenameCollection(currentCollection.name);
        onClose();
      }}
    >
      Rename
    </ContextMenuItem>
    <ContextMenuItem
      className="text-danger-600 hover:bg-danger-50 dark:text-danger-400 dark:hover:bg-danger-900/20"
      onclick={() => {
        onDeleteCollection(currentCollection.name);
        onClose();
      }}
    >
      Delete
    </ContextMenuItem>
  </ContextMenu>
{/snippet}

<div
  class="rounded-lg border border-neutral-200 bg-neutral-50 dark:border-neutral-700 dark:bg-neutral-800/40"
>
  <div
    class={`relative flex items-center px-2 py-2 ${isRootRequestDropTarget() ? "bg-primary-50/60 dark:bg-primary-900/10" : ""}`}
    onclick={handleHeaderClick}
    onkeypress={(event) => {
      if (event.key === "Enter") {
        handleHeaderActivate(event);
      }
    }}
    oncontextmenu={(event) => void handleContextMenu(event)}
    ondragenter={handleRootRequestDragOver}
    ondragover={handleRootRequestDragOver}
    ondrop={(event) => void handleRootRequestDrop(event)}
    role="button"
    tabindex="0"
  >
    <button
      class="h-6 w-4 p-0 text-xs dark:text-white"
      onclick={(event: MouseEvent) => {
        event.stopPropagation();
        onToggleCollection(currentCollection.name);
      }}
      aria-label="Toggle collection"
    >
      {#if expanded}
        <AngleDownOutline class={`h-3 w-3 ${COLLECTION_OUTLINE_BUTTON_CLASSES}`} />
      {:else}
        <AngleRightOutline class={`h-3 w-3 ${COLLECTION_OUTLINE_BUTTON_CLASSES}`} />
      {/if}
    </button>

    <div class="min-w-0 flex-1">
      <div class="flex cursor-pointer items-center gap-2">
        <span
          class="w-8 rounded bg-neutral-200 px-1.5 py-0.5 text-center font-mono text-xs text-neutral-600 dark:bg-neutral-700 dark:text-neutral-300"
        >
          {clampNumberToMax(getTotalRequestCount(currentCollection))}
        </span>

        {#if currentCollection.gitRemote}
          <svg
            width="12"
            height="12"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            class="text-neutral-500 dark:text-neutral-400"
            aria-label={`Git remote: ${currentCollection.gitRemote}`}
          >
            <path d={providerIconPath} />
          </svg>
        {/if}
        <span class="truncate text-sm font-medium text-neutral-800 dark:text-neutral-100">
          {currentCollection.name}
        </span>
      </div>
    </div>

    <div class="flex items-center gap-1">
      <button
        data-no-drag="true"
        type="button"
        class="h-6 shrink-0 {hasCollectionVariables
          ? 'text-warning-500 hover:text-warning-600 dark:text-warning-400 dark:hover:text-warning-300'
          : COLLECTION_OUTLINE_BUTTON_CLASSES}"
        title="Collection variables"
        aria-label={`Open variables for ${currentCollection.name}`}
        onclick={(event: MouseEvent) => {
          event.stopPropagation();
          onOpenVariables(currentCollection.name);
        }}
      >
        <AdjustmentsVerticalOutline class="h-4 w-4" />
      </button>
      <button
        data-no-drag="true"
        class="{COLLECTION_OUTLINE_BUTTON_CLASSES} h-6"
        onclick={(event: MouseEvent) => {
          event.stopPropagation();
          onAddRequest(currentCollection.name);
        }}
        title="Add request"
        aria-label="Add request"
      >
        <PlusOutline class="h-4 w-4" />
      </button>
      <button
        data-no-drag="true"
        id={getMenuTriggerId()}
        class="{COLLECTION_OUTLINE_BUTTON_CLASSES} ml-1 h-6"
        title="More actions"
        aria-label="More actions"
        onclick={(event: MouseEvent) => {
          event.stopPropagation();
          closeContextMenu();
        }}
        oncontextmenu={(event: MouseEvent) => {
          event.preventDefault();
          event.stopPropagation();
        }}
      >
        <DotsHorizontalOutline class="h-4 w-4" />
      </button>
      {@render actionsDropdown(`#${getMenuTriggerId()}`, undefined, closeContextMenu)}
    </div>
  </div>

  {#if expanded}
    <div
      class="space-y-1 border-t border-neutral-200 px-2 pt-1 pb-2 dark:border-neutral-700"
      role="list"
      aria-label={`${currentCollection.name} items`}
    >
      {#if visibleFolders.length === 0 && visibleRequests.length === 0}
        <div class="px-1 py-2 text-xs text-neutral-500 dark:text-neutral-400">
          {isSearching ? "No matching items" : "No items yet"}
        </div>
      {:else}
        {#each visibleFolders as folder (folder.id)}
          <FolderRow
            {folder}
            collectionName={currentCollection.name}
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
        {#each visibleRequests as request (request.id)}
          <RequestRow
            {request}
            collectionName={currentCollection.name}
            selected={selectedRequestId === request.id}
            {isSearching}
            {onDeleteRequest}
            {onRequestSelect}
            onAnyMenuOpen={closeContextMenu}
          />
        {/each}
      {/if}
    </div>
  {/if}
</div>
