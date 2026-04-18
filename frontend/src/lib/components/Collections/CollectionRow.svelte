<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import RequestRow from "$src/lib/components/Collections/RequestRow.svelte";
  import { collection } from "$wails/go/models";
  import AngleDownOutline from "flowbite-svelte-icons/AngleDownOutline.svelte";
  import AngleRightOutline from "flowbite-svelte-icons/AngleRightOutline.svelte";
  import DotsHorizontalOutline from "flowbite-svelte-icons/DotsHorizontalOutline.svelte";
  import PlusOutline from "flowbite-svelte-icons/PlusOutline.svelte";
  import Dropdown from "flowbite-svelte/Dropdown.svelte";
  import DropdownDivider from "flowbite-svelte/DropdownDivider.svelte";
  import DropdownItem from "flowbite-svelte/DropdownItem.svelte";
  import { tick } from "svelte";

  const OUTLINE_BUTTON_CLASSES =
    "text-neutral-800/70 hover:text-neutral-800 dark:text-neutral-100/70 dark:hover:text-neutral-100";

  interface RequestDragState {
    sourceRequestId: string | null;
    sourceCollectionName: string | null;
    targetRequestId: string | null;
    targetCollectionName: string | null;
    position: "before" | "after" | null;
  }

  interface Props {
    collection: collection.Collection;
    expanded: boolean;
    isSearching: boolean;
    selectedRequestId: string | null;
    visibleRequests: collection.Request[];
    dragState: RequestDragState;
    syncing: boolean;
    providerIconPath: string;
    onActivateCollection: (collectionName: string) => void;
    onToggleCollection: (collectionName: string) => void;
    onAddRequest: (collectionName: string) => void;
    onDeleteRequest: (collectionName: string, requestId: string) => void;
    onOpenGitStatus: (currentCollection: collection.Collection) => void;
    onSync: (collectionId: string) => void;
    onExportCollection: (collectionName: string) => void;
    onRenameCollection: (collectionName: string) => void;
    onDeleteCollection: (collectionName: string) => void;
    onRequestSelect?: (requestId: string) => void;
  }

  let {
    collection: currentCollection,
    expanded,
    isSearching,
    selectedRequestId,
    visibleRequests,
    dragState,
    syncing,
    providerIconPath,
    onActivateCollection,
    onToggleCollection,
    onAddRequest,
    onDeleteRequest,
    onOpenGitStatus,
    onSync,
    onExportCollection,
    onRenameCollection,
    onDeleteCollection,
    onRequestSelect = () => {}
  }: Props = $props();

  let contextMenuX = $state(0);
  let contextMenuY = $state(0);
  let contextMenuOpenKey = $state(0);
  let isContextMenuOpen = $state(false);

  function isNoDragTarget(target: EventTarget | null): boolean {
    return target instanceof HTMLElement && target.closest('[data-no-drag="true"]') !== null;
  }

  function closeContextMenu() {
    isContextMenuOpen = false;
  }

  function getMenuTriggerId(): string {
    return `collection-menu-${currentCollection.id}`;
  }

  function getContextMenuTriggerId(): string {
    return `collection-context-menu-${currentCollection.id}-${contextMenuOpenKey}`;
  }

  function getContextMenuPositionStyle(): string {
    return `left: ${contextMenuX + 2}px; top: ${contextMenuY + 2}px;`;
  }

  async function handleContextMenu(event: MouseEvent) {
    if (dragState.sourceRequestId || isNoDragTarget(event.target)) {
      return;
    }

    event.preventDefault();
    event.stopPropagation();

    isContextMenuOpen = false;
    contextMenuX = event.clientX;
    contextMenuY = event.clientY;
    contextMenuOpenKey += 1;

    await tick();
    isContextMenuOpen = true;
  }

  function handleHeaderActivate(event: Event) {
    event.stopPropagation();

    if (isContextMenuOpen) {
      closeContextMenu();
      return;
    }

    onActivateCollection(currentCollection.name);
  }
</script>

{#snippet actionsDropdown(triggeredBy: string, isOpen: boolean | undefined, onClose: () => void)}
  <Dropdown {triggeredBy} {isOpen} class="z-50 w-40" triggerDelay={0} onclose={onClose}>
    {#if currentCollection.gitRemote}
      <DropdownItem
        class="text-gray-900 dark:text-white"
        onclick={() => {
          onOpenGitStatus(currentCollection);
          onClose();
        }}
      >
        Git status
      </DropdownItem>
      <DropdownItem
        class="text-gray-900 dark:text-white"
        disabled={syncing}
        onclick={() => {
          onSync(currentCollection.id);
          onClose();
        }}
      >
        {syncing ? "Syncing..." : "Sync with Git"}
      </DropdownItem>
      <DropdownDivider />
    {/if}
    <DropdownItem
      class="text-gray-900 dark:text-white"
      onclick={() => {
        onExportCollection(currentCollection.name);
        onClose();
      }}
    >
      Export
    </DropdownItem>
    <DropdownDivider />
    <DropdownItem
      class="text-gray-900 dark:text-white"
      onclick={() => {
        onRenameCollection(currentCollection.name);
        onClose();
      }}
    >
      Rename
    </DropdownItem>
    <DropdownItem
      class="text-danger-600 hover:bg-danger-50 dark:text-danger-400 dark:hover:bg-danger-900/20"
      onclick={() => {
        onDeleteCollection(currentCollection.name);
        onClose();
      }}
    >
      Delete
    </DropdownItem>
  </Dropdown>
{/snippet}

<div
  class="rounded-lg border border-neutral-200 bg-neutral-50 dark:border-neutral-700 dark:bg-neutral-800/40"
>
  <div
    class="relative flex items-center px-2 py-2"
    onclick={handleHeaderActivate}
    onkeypress={(event) => {
      if (event.key === "Enter") {
        handleHeaderActivate(event);
      }
    }}
    oncontextmenu={(event) => void handleContextMenu(event)}
    role="button"
    tabindex="0"
  >
    <button
      class="h-6 w-4 p-0 text-xs hover:cursor-pointer dark:text-white"
      onclick={(event: MouseEvent) => {
        event.stopPropagation();
        onToggleCollection(currentCollection.name);
      }}
      aria-label="Toggle collection"
    >
      {#if expanded}
        <AngleDownOutline class={`h-3 w-3 ${OUTLINE_BUTTON_CLASSES}`} />
      {:else}
        <AngleRightOutline class={`h-3 w-3 ${OUTLINE_BUTTON_CLASSES}`} />
      {/if}
    </button>

    <div class="min-w-0 flex-1">
      <div class="flex cursor-pointer items-center gap-2">
        <span
          class="w-6 rounded bg-neutral-200 px-1.5 py-0.5 text-center text-xs text-neutral-600 dark:bg-neutral-700 dark:text-neutral-300"
        >
          {currentCollection.requests?.length || 0}
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
        class="hover:cursor-pointer"
        onclick={(event: MouseEvent) => {
          event.stopPropagation();
          onAddRequest(currentCollection.name);
        }}
        title="Add request"
        aria-label="Add request"
      >
        <PlusOutline class={`h-3 w-3 ${OUTLINE_BUTTON_CLASSES}`} />
      </button>
      <button
        data-no-drag="true"
        id={getMenuTriggerId()}
        class="ml-1 hover:cursor-pointer"
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
        <DotsHorizontalOutline class={`h-3 w-3 ${OUTLINE_BUTTON_CLASSES}`} />
      </button>
      {@render actionsDropdown(`#${getMenuTriggerId()}`, undefined, closeContextMenu)}
    </div>
  </div>

  {#if expanded}
    <div class="space-y-1 border-t border-neutral-200 px-2 pt-1 pb-2 dark:border-neutral-700">
      {#if visibleRequests.length === 0}
        <div class="px-1 py-2 text-xs text-neutral-500 dark:text-neutral-400">
          {isSearching ? "No matching requests" : "No requests yet"}
        </div>
      {:else}
        {#each visibleRequests as request (request.id)}
          <RequestRow
            {request}
            collectionName={currentCollection.name}
            selected={selectedRequestId === request.id}
            requestCount={currentCollection.requests?.length || 0}
            {isSearching}
            {dragState}
            {onDeleteRequest}
            {onRequestSelect}
            onAnyMenuOpen={closeContextMenu}
          />
        {/each}
      {/if}
    </div>
  {/if}
</div>

{#if isContextMenuOpen}
  <button
    id={getContextMenuTriggerId()}
    type="button"
    class="pointer-events-none fixed z-90 h-0 w-0 opacity-0"
    style={getContextMenuPositionStyle()}
    tabindex="-1"
    aria-hidden="true"
  ></button>
  {@render actionsDropdown(`#${getContextMenuTriggerId()}`, isContextMenuOpen, closeContextMenu)}
{/if}
