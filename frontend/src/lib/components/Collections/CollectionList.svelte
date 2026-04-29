<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import ToastContainer from "$src/lib/components/base/ToastContainer.svelte";
  import CollectionRow from "$src/lib/components/Collections/CollectionRow.svelte";
  import CollectionSidebarHeader from "$src/lib/components/Collections/CollectionSidebarHeader.svelte";
  import ExportCollectionModal from "$src/lib/components/Collections/ExportCollectionModal.svelte";
  import ContextMenu from "$src/lib/components/common/ContextMenu.svelte";
  import ContextMenuItem from "$src/lib/components/common/ContextMenuItem.svelte";
  import FeedbackEmptyState from "$src/lib/components/common/FeedbackEmptyState.svelte";
  import VariablesTableEditor from "$src/lib/components/common/VariablesTableEditor.svelte";
  import GitImportView from "$src/lib/components/GitImportView.svelte";
  import GitStatusPanel from "$src/lib/components/GitStatusPanel.svelte";
  import ImportModal from "$src/lib/components/imports/ImportModal.svelte";
  import type { LocalImportFormatOption } from "$src/lib/components/imports/importTypes";
  import LocalImportPane from "$src/lib/components/imports/LocalImportPane.svelte";
  import CodeMirrorEditor from "$src/lib/components/RequestBuilder/CodeMirrorEditor.svelte";
  import {
    collectionTreeUI,
    collectionTreeUIState
  } from "$src/lib/features/collections/collectionTreeUI.svelte";
  import {
    collectionImportStore,
    collectionImportStoreState,
    type CollectionLocalImportFormat
  } from "$src/lib/stores/collectionImportStore.svelte";
  import { collectionStore, collectionStoreState } from "$src/lib/stores/collectionStore.svelte";
  import { modalStack, topModalId } from "$src/lib/stores/modalStackStore.svelte";
  import { notifications } from "$src/lib/stores/notificationStore";
  import { getActiveTab, tabStore, tabStoreState } from "$src/lib/stores/tabStore.svelte";
  import {
    ExportCollection,
    ExportOpenAPICollection,
    GetGitCollectionStatus,
    GitAbortRebase,
    GitDiscardChanges,
    GitKeepOurs,
    GitKeepTheirs,
    ImportCurlRequest,
    OpenCollectionInTerminal,
    SyncGitCollection
  } from "$wails/go/main/App";
  import { collection } from "$wails/go/models";
  import Button from "flowbite-svelte/Button.svelte";
  import Input from "flowbite-svelte/Input.svelte";
  import Label from "flowbite-svelte/Label.svelte";
  import Modal from "flowbite-svelte/Modal.svelte";
  import Select from "flowbite-svelte/Select.svelte";
  import Spinner from "flowbite-svelte/Spinner.svelte";
  import { onDestroy, onMount } from "svelte";
  import { SvelteSet } from "svelte/reactivity";

  interface Props {
    onRequestSelect?: (requestId: string) => void;
  }

  const COLLECTION_LOCAL_IMPORT_FORMATS: LocalImportFormatOption<CollectionLocalImportFormat>[] = [
    {
      key: "postman",
      label: "Postman",
      dropTitle: "Drop your Postman collection here",
      dropSubtitle: "Supports Postman Collection v2 / v2.1 (JSON)",
      pickerButtonLabel: "Select file...",
      icon: "upload"
    },
    {
      key: "bruno",
      label: "Bruno",
      dropTitle: "Drop your Bruno collection folder here",
      dropSubtitle: "Supports Bruno collection folders (.bru files)",
      pickerButtonLabel: "Select folder...",
      icon: "folder"
    },
    {
      key: "openapi",
      label: "OpenAPI / Swagger",
      dropTitle: "Drop your OpenAPI or Swagger document here",
      dropSubtitle: "Supports OpenAPI 3.x and Swagger 2.x (JSON or YAML)",
      pickerButtonLabel: "Select file...",
      icon: "document"
    },
    {
      key: "solo",
      label: "solo",
      dropTitle: "Drop your solo collection here",
      dropSubtitle: "Supports solo collection JSON",
      pickerButtonLabel: "Select file...",
      icon: "upload"
    }
  ];

  let { onRequestSelect = () => {} }: Props = $props();

  const newCollectionModal = modalStack.createModal("collections-new");
  const renameCollectionModal = modalStack.createModal("collections-rename");
  const deleteCollectionModal = modalStack.createModal("collections-delete-collection");
  const deleteFolderModal = modalStack.createModal("collections-delete-folder");
  const deleteRequestModal = modalStack.createModal("collections-delete-request");
  const importCollectionModal = modalStack.createModal("collections-import");
  const collectionVariablesModal = modalStack.createModal("collections-variables");
  const soloCollectionOverwriteModal = modalStack.createModal("collections-solo-overwrite");
  const exportCollectionModal = modalStack.createModal("collections-export");
  let exportCollectionTargetName = $state<string | null>(null);
  let gitImportActionState: { loading: boolean; disabled: boolean; submit: () => void } | null =
    $state(null);

  let newCollectionName = $state("");
  let renameCollectionName = $state("");
  let renameTarget: string | null = null;
  let deleteTarget: string | null = $state(null);
  let deleteFolderCollectionName: string | null = $state(null);
  let deleteFolderName: string | null = $state(null);
  let deleteFolderTarget: string | null = $state(null);
  let deleteRequestTarget: string | null = null;
  let deleteRequestCollectionName: string | null = null;
  let collectionVariablesTargetName: string | null = $state(null);
  let expandedCollections = new SvelteSet<string>();
  let expandedFolders = new SvelteSet<string>();
  let searchQuery = $state("");
  let isCollapsed = $state(false);
  let gitStatusCollectionId: string | null = $state(null);
  let gitStatusCollectionName: string | null = $state(null);
  let syncingCollections: Set<string> = $state(new Set());

  // cURL import state
  let curlInput = $state("");
  let curlTargetCollection = $state("");
  let curlNewCollectionName = $state("");
  let curlCreatingNew = $state(false);

  // Pre-select the active collection when the import modal opens
  $effect(() => {
    if (importCollectionModal.open && collectionStoreState.selectedCollectionName) {
      curlTargetCollection = collectionStoreState.selectedCollectionName;
    }
  });

  $effect(() => {
    soloCollectionOverwriteModal.open = !!collectionImportStoreState.soloCollectionOverwriteName;
  });

  $effect(() => {
    if (
      collectionVariablesModal.open &&
      collectionVariablesTargetName &&
      !collectionVariablesTarget
    ) {
      closeCollectionVariablesModal();
    }
  });

  let sidebarWidth = $state(280); // Default width
  let isResizing = false;
  let suppressNextPrimaryClick = false;

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

  function folderMatches(folder: collection.Folder, query: string): boolean {
    if (!query) return true;
    return normalize(folder.name).includes(query);
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
    if (getVisibleRequests(collection, query).length > 0) return true;
    return (collection.folders || []).some((folder) => shouldShowFolder(folder, query));
  }

  function shouldShowFolder(folder: collection.Folder, query: string): boolean {
    if (!query) return true;
    if (folderMatches(folder, query)) return true;
    if ((folder.requests || []).some((request) => requestMatches(request, query))) return true;
    return (folder.folders || []).some((subfolder) => shouldShowFolder(subfolder, query));
  }

  function getVisibleFolders(folders: collection.Folder[], query: string): collection.Folder[] {
    if (!query) return folders || [];
    return (folders || []).filter((folder) => shouldShowFolder(folder, query));
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

  function openRequestTab(req: collection.Request, collectionName: string) {
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
      label: req.name || "New Request",
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

  function handleCollectionActivate(collectionName: string) {
    selectCollection(collectionName);
    toggleCollection(collectionName);
  }

  function handleAddRequestToCollection(
    collectionName: string,
    newRequest: Partial<collection.Request> = {
      name: "New Request",
      url: "",
      verb: "GET"
    }
  ) {
    const syntheticEvent = {
      stopPropagation() {}
    } as Event;
    return handleAddRequest(syntheticEvent, collectionName, newRequest);
  }

  function openGitStatusForCollection(currentCollection: collection.Collection) {
    gitStatusCollectionId = currentCollection.id;
    gitStatusCollectionName = currentCollection.name;
  }

  async function handleAddFolder(collectionName: string, parentFolderId: string | null = null) {
    try {
      const newFolder = await collectionStore.addFolder(
        collectionName,
        parentFolderId,
        "New Folder"
      );

      if (parentFolderId) {
        expandedFolders.add(parentFolderId);
      }
      expandedFolders.add(newFolder.id);
      expandedCollections.add(collectionName);
    } catch {
      // error already shown by store
    }
  }

  function openRenameCollection(collectionName: string) {
    renameTarget = collectionName;
    renameCollectionName = collectionName;
    renameCollectionModal.open = true;
  }

  function openRenameFolder(collectionName: string, folder: collection.Folder) {
    collectionTreeUI.startFolderRename(folder.id, collectionName);
  }

  function closeNewCollectionDialog() {
    newCollectionModal.open = false;
    newCollectionName = "";
  }

  function closeRenameDialog() {
    renameCollectionModal.open = false;
    renameTarget = null;
    renameCollectionName = "";
  }

  function openNewCollectionModal() {
    newCollectionModal.open = true;
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
    deleteCollectionModal.open = true;
  }

  function handleDeleteFolder(collectionName: string, folder: collection.Folder) {
    deleteFolderCollectionName = collectionName;
    deleteFolderTarget = folder.id;
    deleteFolderName = folder.name;
    deleteFolderModal.open = true;
  }

  function closeDeleteConfirmDialog() {
    deleteCollectionModal.open = false;
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

  function closeDeleteFolderConfirmDialog() {
    deleteFolderModal.open = false;
    deleteFolderCollectionName = null;
    deleteFolderName = null;
    deleteFolderTarget = null;
  }

  async function confirmDeleteFolder() {
    if (!deleteFolderCollectionName || !deleteFolderTarget) return;

    try {
      await collectionStore.removeFolder(deleteFolderCollectionName, deleteFolderTarget);
      expandedFolders.delete(deleteFolderTarget);
      closeDeleteFolderConfirmDialog();
    } catch {
      // error already shown by store
    }
  }

  async function handleAddRequest(
    e: Event,
    collectionName: string,
    newRequest: Partial<collection.Request> = {
      name: "New Request",
      url: "",
      verb: "GET"
    }
  ) {
    e.stopPropagation();

    try {
      const newReq = await collectionStore.addRequest(collectionName, newRequest);

      expandedCollections.add(collectionName);

      if (newReq?.id) {
        openRequestTab(newReq, collectionName);
      }
    } catch {
      // error already shown by store
    }
  }

  async function handleAddRequestToFolder(collectionName: string, folderId: string) {
    try {
      const newReq = await collectionStore.addRequestToFolder(collectionName, folderId, {
        name: "New Request",
        url: "",
        verb: "GET"
      });

      expandedCollections.add(collectionName);
      expandedFolders.add(folderId);

      if (newReq?.id) {
        openRequestTab(newReq, collectionName);
      }
    } catch {
      // error already shown by store
    }
  }

  async function handleDeleteRequest(collectionName: string, requestId: string) {
    deleteRequestCollectionName = collectionName;
    deleteRequestTarget = requestId;
    deleteRequestModal.open = true;
  }

  async function confirmDeleteRequest() {
    if (!deleteRequestTarget || !deleteRequestCollectionName) return;

    try {
      await collectionStore.removeRequest(deleteRequestCollectionName, deleteRequestTarget);
      tabStore.removeTabsForRequest(deleteRequestTarget);
      closeDeleteRequestConfirmDialog();
    } catch {
      // error already shown by store
    }
  }

  function closeDeleteRequestConfirmDialog() {
    deleteRequestModal.open = false;
    deleteRequestTarget = null;
    deleteRequestCollectionName = null;
  }

  function toggleCollapse() {
    isCollapsed = !isCollapsed;
    localStorage.setItem("sidebar_collapsed", String(isCollapsed));
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

  function openExportModal(collectionName: string) {
    exportCollectionTargetName = collectionName;
    exportCollectionModal.open = true;
  }

  async function handleExportCollection(format: "solo" | "openapi") {
    if (!exportCollectionTargetName) return;
    const name = exportCollectionTargetName;
    exportCollectionModal.open = false;
    try {
      if (format === "openapi") {
        await ExportOpenAPICollection(name);
      } else {
        await ExportCollection(name);
      }
      notifications.success("Collection exported successfully");
    } catch (err) {
      notifications.error("Failed to export collection", String(err));
    }
  }

  function openCollectionVariables(collectionName: string) {
    collectionVariablesTargetName = collectionName;
    collectionVariablesModal.open = true;
  }

  function closeCollectionVariablesModal() {
    collectionVariablesModal.open = false;
    collectionVariablesTargetName = null;
  }

  async function handleUpdateCollectionVariables(
    values: Record<string, { value: string; type: string }>
  ) {
    if (!collectionVariablesTargetName) {
      return;
    }

    try {
      await collectionStore.updateCollectionVariables(collectionVariablesTargetName, values);
    } catch {
      // error already shown by store
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
    importCollectionModal.open = false;
    collectionImportStore.resetLocalImport();
    curlInput = "";
    curlCreatingNew = false;
    curlNewCollectionName = "";
  }

  function openImportModal() {
    collectionImportStore.resetLocalImport();
    gitImportActionState = null;
    importCollectionModal.open = true;
  }

  async function runPendingLocalImport() {
    await collectionImportStore.runPendingLocalImport();

    if (
      !collectionImportStoreState.pendingLocalImport &&
      !collectionImportStoreState.soloCollectionOverwriteName
    ) {
      importCollectionModal.open = false;
    }
  }

  async function confirmSoloCollectionOverwrite() {
    await collectionImportStore.confirmSoloCollectionOverwrite();

    if (
      !collectionImportStoreState.pendingLocalImport &&
      !collectionImportStoreState.soloCollectionOverwriteName
    ) {
      importCollectionModal.open = false;
    }
  }

  function isContextMenuOpen(): boolean {
    return (
      collectionTreeUIState.collectionContextMenu.open ||
      collectionTreeUIState.requestContextMenu.open ||
      collectionTreeUIState.folderContextMenu.open
    );
  }

  function isClickInsideContextMenu(target: EventTarget | null): boolean {
    return target instanceof Element && target.closest('[popover="manual"]') !== null;
  }

  function handleGlobalMouseDown(event: MouseEvent) {
    if (event.button !== 0 || !isContextMenuOpen() || isClickInsideContextMenu(event.target)) {
      return;
    }

    suppressNextPrimaryClick = true;
    collectionTreeUI.closeAllContextMenus();
  }

  function handleGlobalClick(event: MouseEvent) {
    if (!suppressNextPrimaryClick || event.button !== 0) {
      return;
    }

    suppressNextPrimaryClick = false;
    event.preventDefault();
    event.stopPropagation();
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

    document.addEventListener("mousedown", handleGlobalMouseDown, true);
    document.addEventListener("click", handleGlobalClick, true);

    return () => {
      document.removeEventListener("mousedown", handleGlobalMouseDown, true);
      document.removeEventListener("click", handleGlobalClick, true);
    };
  });

  onDestroy(async () => {
    modalStack.destroyModal(newCollectionModal.id);
    modalStack.destroyModal(renameCollectionModal.id);
    modalStack.destroyModal(deleteCollectionModal.id);
    modalStack.destroyModal(deleteFolderModal.id);
    modalStack.destroyModal(deleteRequestModal.id);
    modalStack.destroyModal(importCollectionModal.id);
    modalStack.destroyModal(collectionVariablesModal.id);
    modalStack.destroyModal(soloCollectionOverwriteModal.id);
    modalStack.destroyModal(exportCollectionModal.id);
    collectionTreeUI.closeCollectionContextMenu();
    collectionTreeUI.closeRequestContextMenu();
    collectionTreeUI.closeFolderContextMenu();
  });
  let collections = $derived(collectionStoreState.collections);
  // Highlight in sidebar is driven by the active tab, not the collectionStore selection
  let selectedRequestId = $derived(
    tabStoreState.tabs.find((t) => t.id === getActiveTab()?.id)?.id ?? null
  );

  let normalizedQuery = $derived(searchQuery.trim().toLowerCase());
  let isSearching = $derived(normalizedQuery.length > 0);
  let filteredCollections = $derived(
    collections.filter((collection) => shouldShowCollection(collection, normalizedQuery))
  );
  let activePendingLocalImport = $derived(
    collectionImportStoreState.pendingLocalImport?.format ===
      collectionImportStoreState.selectedLocalFormat
      ? collectionImportStoreState.pendingLocalImport
      : null
  );
  let collectionVariablesTarget = $derived(
    collectionVariablesTargetName
      ? collections.find(
          (currentCollection) => currentCollection.name === collectionVariablesTargetName
        ) || null
      : null
  );
  let collectionContextMenuState = $derived(collectionTreeUIState.collectionContextMenu);
  let requestContextMenuState = $derived(collectionTreeUIState.requestContextMenu);
  let folderContextMenuState = $derived(collectionTreeUIState.folderContextMenu);
  let collectionContextTarget = $derived(collectionContextMenuState.collection);
  let requestContextTarget = $derived(requestContextMenuState.request);
  let folderContextTarget = $derived(folderContextMenuState.folder);

  function getCollectionContextMenuTriggerId(): string {
    return `collection-context-menu-trigger-${collectionContextMenuState.openKey}`;
  }

  function getRequestContextMenuTriggerId(): string {
    return `request-context-menu-trigger-${requestContextMenuState.openKey}`;
  }

  function getFolderContextMenuTriggerId(): string {
    return `folder-context-menu-trigger-${folderContextMenuState.openKey}`;
  }

  function getCollectionContextMenuPositionStyle(): string {
    return `left: ${collectionContextMenuState.x + 2}px; top: ${collectionContextMenuState.y + 2}px;`;
  }

  function getRequestContextMenuPositionStyle(): string {
    return `left: ${requestContextMenuState.x + 2}px; top: ${requestContextMenuState.y + 2}px;`;
  }

  function getFolderContextMenuPositionStyle(): string {
    return `left: ${folderContextMenuState.x + 2}px; top: ${folderContextMenuState.y + 2}px;`;
  }

  async function handleDuplicateRequestFromContextMenu() {
    if (!requestContextTarget || !requestContextMenuState.collectionName) {
      return;
    }

    const newRequest: Partial<collection.Request> = {
      ...requestContextTarget,
      name: `${requestContextTarget.name} (copy)`
    };
    delete newRequest.id;

    try {
      const duplicatedRequest = requestContextMenuState.parentFolderId
        ? await collectionStore.addRequestToFolder(
            requestContextMenuState.collectionName,
            requestContextMenuState.parentFolderId,
            newRequest
          )
        : await collectionStore.addRequest(requestContextMenuState.collectionName, newRequest);

      if (duplicatedRequest?.id) {
        openRequestTab(duplicatedRequest, requestContextMenuState.collectionName);
      }
    } catch {
      // error already shown by store
    }
  }
</script>

{#snippet requestContextActionsDropdown(triggeredBy: string, isOpen: boolean, onClose: () => void)}
  <ContextMenu {triggeredBy} {isOpen} {onClose}>
    <ContextMenuItem
      onclick={() => {
        if (requestContextMenuState.requestId && requestContextMenuState.collectionName) {
          collectionTreeUI.startRequestRename(
            requestContextMenuState.requestId,
            requestContextMenuState.collectionName,
            requestContextMenuState.parentFolderId
          );
        }
        onClose();
      }}
    >
      Rename
    </ContextMenuItem>
    <ContextMenuItem
      onclick={() => {
        void handleDuplicateRequestFromContextMenu();
        onClose();
      }}
    >
      Duplicate
    </ContextMenuItem>
    <ContextMenuItem
      className="text-danger-600 hover:bg-danger-50 dark:text-danger-400 dark:hover:bg-danger-900/20"
      onclick={() => {
        if (requestContextMenuState.collectionName && requestContextMenuState.requestId) {
          void handleDeleteRequest(
            requestContextMenuState.collectionName,
            requestContextMenuState.requestId
          );
        }
        onClose();
      }}
    >
      Delete
    </ContextMenuItem>
  </ContextMenu>
{/snippet}

{#snippet collectionContextActionsDropdown(
  triggeredBy: string,
  isOpen: boolean,
  onClose: () => void
)}
  <ContextMenu {triggeredBy} {isOpen} {onClose}>
    <ContextMenuItem
      onclick={() => {
        if (collectionContextTarget) {
          handleAddRequestToCollection(collectionContextTarget.name);
        }
        onClose();
      }}
    >
      New request
    </ContextMenuItem>
    <ContextMenuItem
      onclick={() => {
        if (collectionContextTarget) {
          void handleAddFolder(collectionContextTarget.name, null);
        }
        onClose();
      }}
    >
      New folder
    </ContextMenuItem>
    <ContextMenuItem
      onclick={() => {
        if (collectionContextTarget) {
          openCollectionVariables(collectionContextTarget.name);
        }
        onClose();
      }}
    >
      Variables
    </ContextMenuItem>
    {#if collectionContextTarget?.gitRemote}
      <ContextMenuItem
        onclick={() => {
          if (collectionContextTarget) {
            openGitStatusForCollection(collectionContextTarget);
          }
          onClose();
        }}
      >
        Git status
      </ContextMenuItem>
      <ContextMenuItem
        disabled={!!collectionContextTarget && syncingCollections.has(collectionContextTarget.id)}
        onclick={() => {
          if (collectionContextTarget) {
            void handleSync(collectionContextTarget.id);
          }
          onClose();
        }}
      >
        {collectionContextTarget && syncingCollections.has(collectionContextTarget.id)
          ? "Syncing..."
          : "Sync with Git"}
      </ContextMenuItem>
    {/if}
    <ContextMenuItem
      onclick={() => {
        if (collectionContextTarget) {
          openExportModal(collectionContextTarget.name);
        }
        onClose();
      }}
    >
      Export
    </ContextMenuItem>

    <ContextMenuItem
      onclick={() => {
        if (collectionContextTarget) {
          openRenameCollection(collectionContextTarget.name);
        }
        onClose();
      }}
    >
      Rename
    </ContextMenuItem>
    <ContextMenuItem
      className="text-danger-600 hover:bg-danger-50 dark:text-danger-400 dark:hover:bg-danger-900/20"
      onclick={() => {
        if (collectionContextTarget) {
          handleDeleteCollection(collectionContextTarget.name);
        }
        onClose();
      }}
    >
      Delete
    </ContextMenuItem>
  </ContextMenu>
{/snippet}

{#snippet folderContextActionsDropdown(triggeredBy: string, isOpen: boolean, onClose: () => void)}
  <ContextMenu {triggeredBy} {isOpen} menuClass="z-50 w-44" {onClose}>
    <ContextMenuItem
      onclick={() => {
        if (folderContextMenuState.collectionName && folderContextMenuState.folderId) {
          void handleAddRequestToFolder(
            folderContextMenuState.collectionName,
            folderContextMenuState.folderId
          );
        }
        onClose();
      }}
    >
      New request
    </ContextMenuItem>
    <ContextMenuItem
      onclick={() => {
        if (folderContextMenuState.collectionName && folderContextMenuState.folderId) {
          void handleAddFolder(
            folderContextMenuState.collectionName,
            folderContextMenuState.folderId
          );
        }
        onClose();
      }}
    >
      New subfolder
    </ContextMenuItem>

    <ContextMenuItem
      onclick={() => {
        if (folderContextMenuState.collectionName && folderContextTarget) {
          openRenameFolder(folderContextMenuState.collectionName, folderContextTarget);
        }
        onClose();
      }}
    >
      Rename
    </ContextMenuItem>
    <ContextMenuItem
      className="text-danger-600 hover:bg-danger-50 dark:text-danger-400 dark:hover:bg-danger-900/20"
      onclick={() => {
        if (folderContextMenuState.collectionName && folderContextTarget) {
          handleDeleteFolder(folderContextMenuState.collectionName, folderContextTarget);
        }
        onClose();
      }}
    >
      Delete
    </ContextMenuItem>
  </ContextMenu>
{/snippet}

<div
  class="relative flex h-full {!isCollapsed &&
    'min-w-sidebar'} shrink-0 flex-col border-r border-neutral-200 bg-white dark:border-neutral-800 dark:bg-neutral-900"
  class:collapsed={isCollapsed}
  style={`width: ${isCollapsed ? "auto" : sidebarWidth + "px"};`}
>
  <button
    type="button"
    class="absolute top-0 right-0 z-20 h-full w-1 cursor-col-resize bg-transparent p-0"
    onmousedown={startResize}
    aria-label="Resize sidebar"
  ></button>

  <CollectionSidebarHeader
    collapsed={isCollapsed}
    bind:searchQuery
    onToggleCollapse={toggleCollapse}
    onCreateCollection={openNewCollectionModal}
    onOpenImportModal={openImportModal}
  />

  {#if !isCollapsed}
    <div class="min-h-0 flex-1 overflow-y-auto p-2">
      {#if collectionStoreState.loading}
        <div class="p-3 text-sm text-neutral-500 dark:text-neutral-400">Loading collections...</div>
      {/if}

      <div class="space-y-2">
        {#each filteredCollections as collection (collection.id)}
          <CollectionRow
            {collection}
            expanded={isExpanded(collection.name)}
            searchQuery={normalizedQuery}
            {isSearching}
            {selectedRequestId}
            visibleFolders={getVisibleFolders(collection.folders || [], normalizedQuery)}
            visibleRequests={getVisibleRequests(collection, normalizedQuery)}
            {expandedFolders}
            syncing={syncingCollections.has(collection.id)}
            providerIconPath={getProviderIconPath(collection.gitProvider || "git")}
            onActivateCollection={handleCollectionActivate}
            onToggleCollection={toggleCollection}
            onAddRequest={handleAddRequestToCollection}
            onAddRequestToFolder={handleAddRequestToFolder}
            onCreateFolder={handleAddFolder}
            onRenameFolder={openRenameFolder}
            onDeleteFolder={handleDeleteFolder}
            onDeleteRequest={handleDeleteRequest}
            onOpenGitStatus={openGitStatusForCollection}
            onSync={(collectionId) => void handleSync(collectionId)}
            onExportCollection={(collectionName) => openExportModal(collectionName)}
            onOpenVariables={openCollectionVariables}
            onRenameCollection={openRenameCollection}
            onDeleteCollection={handleDeleteCollection}
            {onRequestSelect}
          />
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

{#if collectionContextTarget}
  <button
    id={getCollectionContextMenuTriggerId()}
    type="button"
    class="pointer-events-none fixed z-90 h-0 w-0 opacity-0"
    style={getCollectionContextMenuPositionStyle()}
    tabindex="-1"
    aria-hidden="true"
  ></button>
  {@render collectionContextActionsDropdown(
    `#${getCollectionContextMenuTriggerId()}`,
    collectionContextMenuState.open,
    collectionTreeUI.closeCollectionContextMenu
  )}
{/if}

{#if requestContextTarget && requestContextMenuState.collectionName}
  <button
    id={getRequestContextMenuTriggerId()}
    type="button"
    class="pointer-events-none fixed z-90 h-0 w-0 opacity-0"
    style={getRequestContextMenuPositionStyle()}
    tabindex="-1"
    aria-hidden="true"
  ></button>
  {@render requestContextActionsDropdown(
    `#${getRequestContextMenuTriggerId()}`,
    requestContextMenuState.open,
    collectionTreeUI.closeRequestContextMenu
  )}
{/if}

{#if folderContextTarget && folderContextMenuState.collectionName}
  <button
    id={getFolderContextMenuTriggerId()}
    type="button"
    class="pointer-events-none fixed z-90 h-0 w-0 opacity-0"
    style={getFolderContextMenuPositionStyle()}
    tabindex="-1"
    aria-hidden="true"
  ></button>
  {@render folderContextActionsDropdown(
    `#${getFolderContextMenuTriggerId()}`,
    folderContextMenuState.open,
    collectionTreeUI.closeFolderContextMenu
  )}
{/if}

{#if newCollectionModal.open}
  <Modal
    bind:open={newCollectionModal.open}
    onclose={closeNewCollectionDialog}
    title="New Collection"
  >
    {#if $topModalId === newCollectionModal.id}
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

{#if renameCollectionModal.open}
  <Modal
    bind:open={renameCollectionModal.open}
    onclose={closeRenameDialog}
    title="Rename Collection"
  >
    {#if $topModalId === renameCollectionModal.id}
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

{#if deleteCollectionModal.open}
  <Modal
    bind:open={deleteCollectionModal.open}
    onclose={closeDeleteConfirmDialog}
    title="Delete Collection"
  >
    {#if $topModalId === deleteCollectionModal.id}
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

{#if deleteFolderModal.open}
  <Modal
    bind:open={deleteFolderModal.open}
    onclose={closeDeleteFolderConfirmDialog}
    title="Delete Folder"
  >
    {#if $topModalId === deleteFolderModal.id}
      <ToastContainer />
    {/if}
    <div class="space-y-2 text-sm">
      <p class="text-neutral-700 dark:text-neutral-200">
        Are you sure you want to delete folder "{deleteFolderName}"?
      </p>
      <p class="text-danger-600 dark:text-danger-300">
        This also removes all nested folders and requests.
      </p>
    </div>
    {#snippet footer()}
      <div class="flex w-full justify-end gap-2">
        <Button color="red" onclick={confirmDeleteFolder}>Delete</Button>
      </div>
    {/snippet}
  </Modal>
{/if}

{#if deleteRequestModal.open}
  <Modal
    bind:open={deleteRequestModal.open}
    onclose={closeDeleteRequestConfirmDialog}
    title="Delete Request"
  >
    {#if $topModalId === deleteRequestModal.id}
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

{#if importCollectionModal.open}
  <ImportModal
    title="Import Collection or Request"
    bind:open={importCollectionModal.open}
    onClose={closeImportModal}
    showCurlSection
    localActionLabel={collectionImportStoreState.localImportLoading ? "Importing..." : "Import"}
    localActionDisabled={!activePendingLocalImport || collectionImportStoreState.localImportLoading}
    onLocalAction={runPendingLocalImport}
    curlActionLabel="Import Request"
    curlActionDisabled={!curlInput.trim() || (!curlTargetCollection && !curlCreatingNew)}
    onCurlAction={handleImportCurl}
    gitActionState={gitImportActionState}
  >
    {#snippet localContent()}
      {#if $topModalId === importCollectionModal.id}
        <ToastContainer />
      {/if}
      {#if collectionImportStoreState.localImportLoading}
        <div class="flex h-full items-center justify-center">
          <Spinner type="bars" color="primary" />
        </div>
      {:else}
        <LocalImportPane
          formats={COLLECTION_LOCAL_IMPORT_FORMATS}
          bind:selectedFormat={collectionImportStoreState.selectedLocalFormat}
          onImport={collectionImportStore.setPendingLocalImportFromDrop}
        />
      {/if}
    {/snippet}

    {#snippet curlContent()}
      {#if $topModalId === importCollectionModal.id}
        <ToastContainer />
      {/if}
      <div class="flex flex-col gap-4">
        <div class="flex flex-col gap-1">
          <Label>Paste cURL command</Label>
          <div
            class="h-48 overflow-hidden rounded border border-neutral-200 dark:border-neutral-700"
          >
            <CodeMirrorEditor value={curlInput} language="text" onChange={(v) => (curlInput = v)} />
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
                New...
              </Button>
            </div>
          {/if}
        </div>
      </div>
    {/snippet}

    {#snippet gitContent()}
      {#if $topModalId === importCollectionModal.id}
        <ToastContainer />
      {/if}
      <GitImportView
        onImported={() => (importCollectionModal.open = false)}
        onActionStateChange={(state) => {
          gitImportActionState = state;
        }}
      />
    {/snippet}
  </ImportModal>
{/if}

{#if collectionVariablesModal.open && collectionVariablesTarget}
  <Modal
    bind:open={collectionVariablesModal.open}
    onclose={closeCollectionVariablesModal}
    title="Collection Variables"
    size="xl"
  >
    {#if $topModalId === collectionVariablesModal.id}
      <ToastContainer />
    {/if}
    {#key collectionVariablesTarget.id}
      <VariablesTableEditor
        name={collectionVariablesTarget.name}
        values={collectionVariablesTarget.variables}
        onUpdate={handleUpdateCollectionVariables}
      />
    {/key}
  </Modal>
{/if}

{#if soloCollectionOverwriteModal.open}
  <Modal
    title="Overwrite collection?"
    bind:open={soloCollectionOverwriteModal.open}
    onclose={() => {
      collectionImportStore.cancelSoloCollectionOverwrite();
      soloCollectionOverwriteModal.open = false;
    }}
    size="xl"
  >
    {#if $topModalId === soloCollectionOverwriteModal.id}
      <ToastContainer />
    {/if}
    <p>Collection "{collectionImportStoreState.soloCollectionOverwriteName}" already exists.</p>
    <p class="text-sm text-neutral-600 dark:text-neutral-400">Do you want to overwrite it?</p>
    {#snippet footer()}
      <div class="flex w-full justify-end gap-2">
        <Button color="red" onclick={confirmSoloCollectionOverwrite}>Overwrite</Button>
      </div>
    {/snippet}
  </Modal>
{/if}

<ExportCollectionModal
  bind:open={exportCollectionModal.open}
  onExport={(format) => void handleExportCollection(format)}
  onClose={() => (exportCollectionModal.open = false)}
/>

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
