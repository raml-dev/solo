/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: AGPL-3.0-only
 */

import { collectionStore } from "$src/lib/stores/collectionStore.svelte";
import { notifications } from "$src/lib/stores/notificationStore";
import {
  ImportBrunoCollection,
  ImportOpenAPICollection,
  ImportPostmanCollection,
  ImportSoloCollection,
  SelectDirectory,
  SelectFile
} from "$wails/go/main/App";

export type CollectionLocalImportFormat = "postman" | "bruno" | "openapi" | "solo";

export interface PendingLocalImport {
  format: CollectionLocalImportFormat;
  path: string;
}

interface CollectionImportState {
  selectedLocalFormat: CollectionLocalImportFormat;
  pendingLocalImport: PendingLocalImport | null;
  localImportLoading: boolean;
  pendingSoloCollectionPath: string | null;
  soloCollectionOverwriteName: string | null;
}

type LocalImportRunResult = "completed" | "awaiting_overwrite";

interface LocalImportHandler {
  pick: () => Promise<string>;
  run: (path: string) => Promise<LocalImportRunResult>;
  errorTitle: string;
}

const initialState: CollectionImportState = {
  selectedLocalFormat: "postman",
  pendingLocalImport: null,
  localImportLoading: false,
  pendingSoloCollectionPath: null,
  soloCollectionOverwriteName: null
};

function parseCollectionNameFromError(message: string): string | null {
  const match = message.match(/collection\s+(\S+)\s+already exists/i);
  return match ? match[1] : null;
}

async function executePostmanImport(path: string): Promise<LocalImportRunResult> {
  await ImportPostmanCollection(path);
  await collectionStore.loadCollections();
  notifications.success("Postman collection imported");
  return "completed";
}

async function executeBrunoImport(path: string): Promise<LocalImportRunResult> {
  await ImportBrunoCollection(path);
  await collectionStore.loadCollections();
  notifications.success("Bruno collection imported");
  return "completed";
}

async function executeOpenAPIImport(path: string): Promise<LocalImportRunResult> {
  const { warnings } = await ImportOpenAPICollection(path);
  await collectionStore.loadCollections();

  for (const warning of warnings ?? []) {
    notifications.warning(warning);
  }

  notifications.success("OpenAPI collection imported");
  return "completed";
}

async function executeSoloCollectionImport(
  path: string,
  overwrite: boolean
): Promise<LocalImportRunResult> {
  try {
    await ImportSoloCollection(path, overwrite);
    await collectionStore.loadCollections();
    notifications.success("Collection imported successfully");
    return "completed";
  } catch (err) {
    const message = String(err ?? "Failed to import collection");
    const existingName = parseCollectionNameFromError(message);

    if (!overwrite && existingName) {
      collectionImportStoreState.pendingSoloCollectionPath = path;
      collectionImportStoreState.soloCollectionOverwriteName = existingName;
      return "awaiting_overwrite";
    }

    throw new Error(message);
  }
}

const LOCAL_IMPORT_HANDLERS: Record<CollectionLocalImportFormat, LocalImportHandler> = {
  postman: {
    pick: () => SelectFile("Select Postman Collection", "*.json", "JSON Files"),
    run: executePostmanImport,
    errorTitle: "Failed to import Postman collection"
  },
  bruno: {
    pick: () => SelectDirectory("Select Bruno Collection Folder"),
    run: executeBrunoImport,
    errorTitle: "Failed to import Bruno collection"
  },
  openapi: {
    pick: () =>
      SelectFile(
        "Select OpenAPI / Swagger Document",
        "*.json;*.yaml;*.yml",
        "OpenAPI / Swagger Files"
      ),
    run: executeOpenAPIImport,
    errorTitle: "Failed to import OpenAPI collection"
  },
  solo: {
    pick: () => SelectFile("Select Solo Collection", "*.json", "JSON Files"),
    run: (path) => executeSoloCollectionImport(path, false),
    errorTitle: "Failed to import collection"
  }
};

function setPendingLocalImport(format: CollectionLocalImportFormat, path: string) {
  if (!path) {
    return;
  }

  collectionImportStoreState.selectedLocalFormat = format;
  collectionImportStoreState.pendingLocalImport = { format, path };
  collectionImportStoreState.pendingSoloCollectionPath = null;
  collectionImportStoreState.soloCollectionOverwriteName = null;
}

export const collectionImportStoreState = $state<CollectionImportState>({ ...initialState });

export const collectionImportStore = {
  resetLocalImport() {
    collectionImportStoreState.selectedLocalFormat = initialState.selectedLocalFormat;
    collectionImportStoreState.pendingLocalImport = null;
    collectionImportStoreState.localImportLoading = false;
    collectionImportStoreState.pendingSoloCollectionPath = null;
    collectionImportStoreState.soloCollectionOverwriteName = null;
  },

  async pickLocalImportPath() {
    const format = collectionImportStoreState.selectedLocalFormat;
    const path = await LOCAL_IMPORT_HANDLERS[format].pick();

    if (!path) {
      return;
    }

    setPendingLocalImport(format, path);
  },

  async setPendingLocalImportFromDrop(format: CollectionLocalImportFormat, path?: string) {
    if (!path) {
      return;
    }

    setPendingLocalImport(format, path);
  },

  clearPendingLocalImport() {
    collectionImportStoreState.pendingLocalImport = null;
    collectionImportStoreState.pendingSoloCollectionPath = null;
    collectionImportStoreState.soloCollectionOverwriteName = null;
  },

  async runPendingLocalImport() {
    const pendingImport = collectionImportStoreState.pendingLocalImport;
    if (!pendingImport) {
      return;
    }

    const handler = LOCAL_IMPORT_HANDLERS[pendingImport.format];
    let shouldKeepLoadingUntilReset = false;

    collectionImportStoreState.localImportLoading = true;
    collectionImportStoreState.pendingSoloCollectionPath = null;
    collectionImportStoreState.soloCollectionOverwriteName = null;

    try {
      const result = await handler.run(pendingImport.path);
      if (result === "completed") {
        collectionImportStoreState.pendingLocalImport = null;
        shouldKeepLoadingUntilReset = true;
      }
    } catch (err) {
      notifications.error(handler.errorTitle, String(err));
    } finally {
      if (!shouldKeepLoadingUntilReset) {
        collectionImportStoreState.localImportLoading = false;
      }
    }
  },

  cancelSoloCollectionOverwrite() {
    collectionImportStoreState.pendingSoloCollectionPath = null;
    collectionImportStoreState.soloCollectionOverwriteName = null;
  },

  async confirmSoloCollectionOverwrite() {
    const path = collectionImportStoreState.pendingSoloCollectionPath;
    if (!path) {
      return;
    }

    let shouldKeepLoadingUntilReset = false;
    collectionImportStoreState.localImportLoading = true;
    collectionImportStoreState.pendingSoloCollectionPath = null;
    collectionImportStoreState.soloCollectionOverwriteName = null;

    try {
      const result = await executeSoloCollectionImport(path, true);
      if (result === "completed") {
        collectionImportStoreState.pendingLocalImport = null;
        shouldKeepLoadingUntilReset = true;
      }
    } catch (err) {
      notifications.error("Failed to import collection", String(err));
    } finally {
      if (!shouldKeepLoadingUntilReset) {
        collectionImportStoreState.localImportLoading = false;
      }
    }
  }
};
