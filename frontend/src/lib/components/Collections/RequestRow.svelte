<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import {
    collectionTreeUI,
    collectionTreeUIState
  } from "$src/lib/features/collections/collectionTreeUI.svelte";
  import { collectionStore } from "$src/lib/stores/collectionStore.svelte";
  import { notifications } from "$src/lib/stores/notificationStore";
  import { tabStore } from "$src/lib/stores/tabStore.svelte";
  import { getMethodBadgeClass } from "$src/lib/utils/http";
  import { collection } from "$wails/go/models";
  import DotsHorizontalOutline from "flowbite-svelte-icons/DotsHorizontalOutline.svelte";
  import Dropdown from "flowbite-svelte/Dropdown.svelte";
  import DropdownItem from "flowbite-svelte/DropdownItem.svelte";
  import Input from "flowbite-svelte/Input.svelte";
  import { tick } from "svelte";

  const OUTLINE_BUTTON_CLASSES =
    "text-neutral-800/70 hover:text-neutral-800 dark:text-neutral-100/70 dark:hover:text-neutral-100";

  interface Props {
    request: collection.Request;
    collectionName: string;
    parentFolderId?: string | null;
    selected: boolean;
    isSearching: boolean;
    dragEnabled?: boolean;
    onDeleteRequest: (collectionName: string, requestId: string) => void;
    onRequestSelect?: (requestId: string) => void;
    onAnyMenuOpen?: () => void;
  }

  let {
    request,
    collectionName,
    parentFolderId = null,
    selected,
    isSearching,
    dragEnabled = true,
    onDeleteRequest,
    onRequestSelect = () => {},
    onAnyMenuOpen = () => {}
  }: Props = $props();

  let editingNameInputEl: HTMLInputElement | undefined = $state();
  let isEditing = $state(false);
  let editingName = $state("");
  let requestDragState = $derived(collectionTreeUIState.requestDrag);
  let requestRenameState = $derived(collectionTreeUIState.requestRename);
  let anyContextMenuOpen = $derived(
    collectionTreeUIState.collectionContextMenu.open ||
      collectionTreeUIState.requestContextMenu.open ||
      collectionTreeUIState.folderContextMenu.open
  );

  function normalizeHeaders(headers: Record<string, unknown> | undefined) {
    return headers
      ? Object.entries(headers).map(([key, value], i) => ({
          id: `header-${i}`,
          key,
          value: String(value),
          enabled: true
        }))
      : [];
  }

  function openRequest() {
    if (!request?.id) {
      notifications.warning("Unable to open request: missing request id");
      return;
    }

    tabStore.openTab(request.id, collectionName, {
      label: request.name || "New Request",
      verb: request.verb || "GET",
      url: request.url || "",
      body: request.body || "",
      bodyFormat: "json",
      headers: normalizeHeaders(request.headers),
      auth: request.auth,
      settings: request.settings || {},
      preRequestScript: request.preRequestScript || "",
      postResponseScript: request.postResponseScript || ""
    });

    onRequestSelect(request.id);
  }

  function isContextMenuGesture(event: MouseEvent): boolean {
    return event.button !== 0 || event.ctrlKey;
  }

  function handleRequestActivate(event: MouseEvent) {
    if (isContextMenuGesture(event)) {
      return;
    }

    if (anyContextMenuOpen) {
      collectionTreeUI.closeAllContextMenus();
      return;
    }

    openRequest();
  }

  async function beginRename() {
    isEditing = true;
    editingName = request.name || "";

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
    if (!nextName || nextName === request.name) {
      cancelRename();
      return;
    }

    try {
      await collectionStore.updateRequest(
        collectionName,
        collection.Request.createFrom({
          ...request,
          name: nextName,
          lastUpdateTimestamp: new Date().toISOString()
        })
      );
      tabStore.renameTabsByRequestId(request.id, nextName);
    } catch {
      // error already shown by store
    } finally {
      cancelRename();
    }
  }

  function resetDragState() {
    collectionTreeUI.resetRequestDrag();
  }

  function canDrag(): boolean {
    return dragEnabled && !isSearching && !isEditing;
  }

  function isDragSource(): boolean {
    return (
      requestDragState.sourceRequestId === request.id &&
      requestDragState.sourceCollectionName === collectionName &&
      requestDragState.sourceParentFolderId === parentFolderId
    );
  }

  function isDropTarget(position: "before" | "after"): boolean {
    return (
      requestDragState.targetMode === "request" &&
      requestDragState.targetRequestId === request.id &&
      requestDragState.targetCollectionName === collectionName &&
      requestDragState.targetParentFolderId === parentFolderId &&
      requestDragState.position === position
    );
  }

  function isNoDragTarget(target: EventTarget | null): boolean {
    return target instanceof HTMLElement && target.closest('[data-no-drag="true"]') !== null;
  }

  function getDragPosition(event: DragEvent): "before" | "after" | null {
    const currentTarget = event.currentTarget;
    if (!(currentTarget instanceof HTMLElement)) {
      return null;
    }

    const bounds = currentTarget.getBoundingClientRect();
    const midpoint = bounds.top + bounds.height / 2;
    return event.clientY < midpoint ? "before" : "after";
  }

  function handleDragStart(event: DragEvent) {
    if (isNoDragTarget(event.target) || !canDrag()) {
      event.preventDefault();
      resetDragState();
      return;
    }

    collectionTreeUI.startRequestDrag(request.id, collectionName, parentFolderId);

    event.dataTransfer?.setData("text/plain", request.id);
    if (event.dataTransfer) {
      event.dataTransfer.effectAllowed = "move";
    }
  }

  function handleDragOver(event: DragEvent) {
    if (!requestDragState.sourceRequestId || !requestDragState.sourceCollectionName) {
      return;
    }

    if (
      requestDragState.sourceCollectionName !== collectionName ||
      requestDragState.sourceRequestId === request.id
    ) {
      return;
    }

    const nextPosition = getDragPosition(event);
    if (!nextPosition) {
      return;
    }

    event.preventDefault();
    event.stopPropagation();
    collectionTreeUI.setRequestTarget(collectionName, parentFolderId, request.id, nextPosition);

    if (event.dataTransfer) {
      event.dataTransfer.dropEffect = "move";
    }
  }

  async function handleDrop(event: DragEvent) {
    event.preventDefault();
    event.stopPropagation();

    if (
      !requestDragState.sourceRequestId ||
      !requestDragState.sourceCollectionName ||
      requestDragState.sourceCollectionName !== collectionName ||
      requestDragState.sourceRequestId === request.id ||
      requestDragState.targetMode !== "request" ||
      requestDragState.targetRequestId !== request.id ||
      requestDragState.targetCollectionName !== collectionName ||
      requestDragState.targetParentFolderId !== parentFolderId ||
      !requestDragState.position
    ) {
      resetDragState();
      return;
    }

    try {
      await collectionStore.moveRequest(
        collectionName,
        requestDragState.sourceRequestId,
        requestDragState.sourceParentFolderId,
        parentFolderId,
        request.id,
        requestDragState.position
      );
    } finally {
      resetDragState();
    }
  }

  function handleDragEnd() {
    resetDragState();
  }

  function closeContextMenu() {
    collectionTreeUI.closeRequestContextMenu();
  }

  function getMenuTriggerId(): string {
    return `request-menu-${request.id}`;
  }

  async function handleContextMenu(event: MouseEvent) {
    if (requestDragState.sourceRequestId || isEditing || isNoDragTarget(event.target)) {
      return;
    }

    event.preventDefault();
    event.stopPropagation();
    onAnyMenuOpen();
    closeContextMenu();
    collectionTreeUI.stageRequestContextMenu(
      request,
      request.id,
      collectionName,
      parentFolderId,
      event.clientX,
      event.clientY
    );
    await tick();
    collectionTreeUI.showRequestContextMenu();
  }

  async function duplicateRequest() {
    const newRequest: Partial<collection.Request> = {
      ...request,
      name: `${request.name} (copy)`
    };
    delete newRequest.id;

    try {
      const duplicatedRequest = parentFolderId
        ? await collectionStore.addRequestToFolder(collectionName, parentFolderId, newRequest)
        : await collectionStore.addRequest(collectionName, newRequest);
      if (duplicatedRequest?.id) {
        tabStore.renameTabsByRequestId(duplicatedRequest.id, duplicatedRequest.name);
        openDuplicatedRequest(duplicatedRequest);
      }
    } catch {
      // error already shown by store
    }
  }

  function openDuplicatedRequest(duplicatedRequest: collection.Request) {
    tabStore.openTab(duplicatedRequest.id, collectionName, {
      label: duplicatedRequest.name || "New Request",
      verb: duplicatedRequest.verb || "GET",
      url: duplicatedRequest.url || "",
      body: duplicatedRequest.body || "",
      bodyFormat: "json",
      headers: normalizeHeaders(duplicatedRequest.headers),
      auth: duplicatedRequest.auth,
      settings: duplicatedRequest.settings || {},
      preRequestScript: duplicatedRequest.preRequestScript || "",
      postResponseScript: duplicatedRequest.postResponseScript || ""
    });
    onRequestSelect(duplicatedRequest.id);
  }

  $effect(() => {
    if (
      requestRenameState.requestId !== request.id ||
      requestRenameState.collectionName !== collectionName ||
      requestRenameState.parentFolderId !== parentFolderId
    ) {
      return;
    }

    collectionTreeUI.consumeRequestRename();
    void beginRename();
  });
</script>

{#snippet actionsDropdown(triggeredBy: string, isOpen: boolean | undefined, onClose: () => void)}
  <Dropdown {triggeredBy} {isOpen} class="z-50 w-40" triggerDelay={0} onclose={onClose}>
    <DropdownItem
      class="text-gray-900 dark:text-white"
      onclick={(event) => {
        void startRename(event);
        onClose();
      }}
    >
      Rename
    </DropdownItem>
    <DropdownItem
      class="text-gray-900 dark:text-white"
      onclick={() => {
        void duplicateRequest();
        onClose();
      }}
    >
      Duplicate
    </DropdownItem>
    <DropdownItem
      class="text-danger-600 hover:bg-danger-50 dark:text-danger-400 dark:hover:bg-danger-900/20"
      onclick={() => {
        onDeleteRequest(collectionName, request.id);
        onClose();
      }}
    >
      Delete
    </DropdownItem>
  </Dropdown>
{/snippet}

<div class="relative">
  {#if isDropTarget("before")}
    <div class="pointer-events-none absolute inset-x-2 top-0 h-0.5 rounded bg-primary-500"></div>
  {/if}

  <div
    class={`group flex cursor-pointer items-center gap-2 rounded px-2 py-1.5 select-none [-webkit-user-drag:element] hover:bg-neutral-100 dark:hover:bg-neutral-700/60 ${selected ? "bg-neutral-200/70 dark:bg-neutral-700/90" : ""} ${isDragSource() ? "opacity-50" : ""}`}
    draggable={canDrag()}
    onclick={handleRequestActivate}
    onkeypress={(event) => event.key === "Enter" && !isEditing && openRequest()}
    oncontextmenu={(event) => void handleContextMenu(event)}
    ondragstart={handleDragStart}
    ondragover={handleDragOver}
    ondrop={(event) => void handleDrop(event)}
    ondragend={handleDragEnd}
    role="button"
    tabindex="0"
  >
    <span class={getMethodBadgeClass(request.verb)}>
      {request.verb}
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
        class="min-w-0 flex-1 truncate text-sm text-neutral-800 select-none dark:text-neutral-100"
        role="button"
        tabindex="0"
        onkeydown={(event) => event.key === "Enter" && startRename(event)}
      >
        {request.name}
      </span>
    {/if}
    <button
      data-no-drag="true"
      draggable={false}
      class="invisible group-hover:visible hover:cursor-pointer"
      id={getMenuTriggerId()}
      onclick={(event: MouseEvent) => {
        event.stopPropagation();
        onAnyMenuOpen();
        closeContextMenu();
      }}
      oncontextmenu={(event: MouseEvent) => {
        event.preventDefault();
        event.stopPropagation();
      }}
      onmousedown={(event: MouseEvent) => {
        event.stopPropagation();
      }}
      ondragstart={(event: DragEvent) => {
        event.preventDefault();
        event.stopPropagation();
      }}
      title="Request actions"
      aria-label="Request actions"
    >
      <DotsHorizontalOutline class={`h-3 w-3 ${OUTLINE_BUTTON_CLASSES}`} />
    </button>
    {@render actionsDropdown(`#${getMenuTriggerId()}`, undefined, closeContextMenu)}
  </div>

  {#if isDropTarget("after")}
    <div class="pointer-events-none absolute inset-x-2 bottom-0 h-0.5 rounded bg-primary-500"></div>
  {/if}
</div>
