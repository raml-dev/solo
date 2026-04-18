/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: AGPL-3.0-only
 */

import { collection } from "$wails/go/models";

export type RequestDropMode = "request" | "container" | null;

export interface RequestContextMenuState {
  requestId: string | null;
  collectionName: string | null;
  parentFolderId: string | null;
  request: collection.Request | null;
  x: number;
  y: number;
  openKey: number;
  open: boolean;
}

export interface FolderContextMenuState {
  folderId: string | null;
  collectionName: string | null;
  folder: collection.Folder | null;
  x: number;
  y: number;
  openKey: number;
  open: boolean;
}

export interface CollectionContextMenuState {
  collectionId: string | null;
  collection: collection.Collection | null;
  x: number;
  y: number;
  openKey: number;
  open: boolean;
}

export interface RequestRenameState {
  requestId: string | null;
  collectionName: string | null;
  parentFolderId: string | null;
}

export interface FolderRenameState {
  folderId: string | null;
  collectionName: string | null;
}

export interface RequestDragState {
  sourceRequestId: string | null;
  sourceCollectionName: string | null;
  sourceParentFolderId: string | null;
  targetRequestId: string | null;
  targetCollectionName: string | null;
  targetParentFolderId: string | null;
  targetMode: RequestDropMode;
  position: "before" | "after" | null;
}

const initialRequestDragState: RequestDragState = {
  sourceRequestId: null,
  sourceCollectionName: null,
  sourceParentFolderId: null,
  targetRequestId: null,
  targetCollectionName: null,
  targetParentFolderId: null,
  targetMode: null,
  position: null
};

const initialRequestContextMenuState: RequestContextMenuState = {
  requestId: null,
  collectionName: null,
  parentFolderId: null,
  request: null,
  x: 0,
  y: 0,
  openKey: 0,
  open: false
};

const initialFolderContextMenuState: FolderContextMenuState = {
  folderId: null,
  collectionName: null,
  folder: null,
  x: 0,
  y: 0,
  openKey: 0,
  open: false
};

const initialCollectionContextMenuState: CollectionContextMenuState = {
  collectionId: null,
  collection: null,
  x: 0,
  y: 0,
  openKey: 0,
  open: false
};

const initialRequestRenameState: RequestRenameState = {
  requestId: null,
  collectionName: null,
  parentFolderId: null
};

const initialFolderRenameState: FolderRenameState = {
  folderId: null,
  collectionName: null
};

export const collectionTreeUIState = $state({
  requestDrag: { ...initialRequestDragState },
  requestContextMenu: { ...initialRequestContextMenuState },
  folderContextMenu: { ...initialFolderContextMenuState },
  collectionContextMenu: { ...initialCollectionContextMenuState },
  requestRename: { ...initialRequestRenameState },
  folderRename: { ...initialFolderRenameState }
});

export const collectionTreeUI = {
  resetRequestDrag() {
    collectionTreeUIState.requestDrag = { ...initialRequestDragState };
  },

  startRequestDrag(requestId: string, collectionName: string, parentFolderId: string | null) {
    collectionTreeUIState.requestDrag = {
      ...initialRequestDragState,
      sourceRequestId: requestId,
      sourceCollectionName: collectionName,
      sourceParentFolderId: parentFolderId
    };
  },

  setRequestTarget(
    collectionName: string,
    parentFolderId: string | null,
    targetRequestId: string,
    position: "before" | "after"
  ) {
    collectionTreeUIState.requestDrag.targetCollectionName = collectionName;
    collectionTreeUIState.requestDrag.targetParentFolderId = parentFolderId;
    collectionTreeUIState.requestDrag.targetRequestId = targetRequestId;
    collectionTreeUIState.requestDrag.targetMode = "request";
    collectionTreeUIState.requestDrag.position = position;
  },

  setRequestContainerTarget(collectionName: string, parentFolderId: string | null) {
    collectionTreeUIState.requestDrag.targetCollectionName = collectionName;
    collectionTreeUIState.requestDrag.targetParentFolderId = parentFolderId;
    collectionTreeUIState.requestDrag.targetRequestId = null;
    collectionTreeUIState.requestDrag.targetMode = "container";
    collectionTreeUIState.requestDrag.position = null;
  },

  stageRequestContextMenu(
    request: collection.Request,
    requestId: string,
    collectionName: string,
    parentFolderId: string | null,
    x: number,
    y: number
  ) {
    collectionTreeUIState.collectionContextMenu = {
      ...initialCollectionContextMenuState
    };
    collectionTreeUIState.folderContextMenu = {
      ...initialFolderContextMenuState
    };
    collectionTreeUIState.requestContextMenu = {
      requestId,
      collectionName,
      parentFolderId,
      request,
      x,
      y,
      openKey: collectionTreeUIState.requestContextMenu.openKey + 1,
      open: false
    };
  },

  showRequestContextMenu() {
    collectionTreeUIState.requestContextMenu.open = true;
  },

  closeRequestContextMenu() {
    collectionTreeUIState.requestContextMenu.open = false;
  },

  stageFolderContextMenu(
    folder: collection.Folder,
    folderId: string,
    collectionName: string,
    x: number,
    y: number
  ) {
    collectionTreeUIState.collectionContextMenu = {
      ...initialCollectionContextMenuState
    };
    collectionTreeUIState.requestContextMenu = {
      ...initialRequestContextMenuState
    };
    collectionTreeUIState.folderContextMenu = {
      folderId,
      collectionName,
      folder,
      x,
      y,
      openKey: collectionTreeUIState.folderContextMenu.openKey + 1,
      open: false
    };
  },

  showFolderContextMenu() {
    collectionTreeUIState.folderContextMenu.open = true;
  },

  closeFolderContextMenu() {
    collectionTreeUIState.folderContextMenu.open = false;
  },

  stageCollectionContextMenu(
    currentCollection: collection.Collection,
    collectionId: string,
    x: number,
    y: number
  ) {
    collectionTreeUIState.requestContextMenu = {
      ...initialRequestContextMenuState
    };
    collectionTreeUIState.folderContextMenu = {
      ...initialFolderContextMenuState
    };
    collectionTreeUIState.collectionContextMenu = {
      collectionId,
      collection: currentCollection,
      x,
      y,
      openKey: collectionTreeUIState.collectionContextMenu.openKey + 1,
      open: false
    };
  },

  showCollectionContextMenu() {
    collectionTreeUIState.collectionContextMenu.open = true;
  },

  closeCollectionContextMenu() {
    collectionTreeUIState.collectionContextMenu.open = false;
  },

  closeAllContextMenus() {
    collectionTreeUIState.requestContextMenu.open = false;
    collectionTreeUIState.folderContextMenu.open = false;
    collectionTreeUIState.collectionContextMenu.open = false;
  },

  startRequestRename(requestId: string, collectionName: string, parentFolderId: string | null) {
    collectionTreeUIState.requestRename = {
      requestId,
      collectionName,
      parentFolderId
    };
  },

  consumeRequestRename() {
    collectionTreeUIState.requestRename = {
      ...initialRequestRenameState
    };
  },

  startFolderRename(folderId: string, collectionName: string) {
    collectionTreeUIState.folderRename = {
      folderId,
      collectionName
    };
  },

  consumeFolderRename() {
    collectionTreeUIState.folderRename = {
      ...initialFolderRenameState
    };
  }
};
