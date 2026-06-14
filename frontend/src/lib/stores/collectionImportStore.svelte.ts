/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: AGPL-3.0-only
 */

import { collectionStore } from "$src/lib/stores/collectionStore.svelte";
import { notifications } from "$src/lib/stores/notificationStore";
import {
  ImportBrunoCollections,
  ImportOpenAPICollections,
  ImportPostmanCollections,
  ImportSoloCollections,
  SelectDirectory,
  SelectFiles
} from "$wails/go/main/App";
import type { collection } from "$wails/go/models";

export type CollectionLocalImportFormat = "postman" | "bruno" | "openapi" | "solo";

export interface PendingLocalImport {
  format: CollectionLocalImportFormat;
  paths: string[];
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
  pick: () => Promise<string[]>;
  run: (paths: string[]) => Promise<LocalImportRunResult>;
  errorTitle: string;
}

const maxFailureDetailToasts = 3;
const maxWarningDetailToasts = 3;

const initialState: CollectionImportState = {
  selectedLocalFormat: "postman",
  pendingLocalImport: null,
  localImportLoading: false,
  pendingSoloCollectionPath: null,
  soloCollectionOverwriteName: null
};

function getImportSummaryLabel(format: CollectionLocalImportFormat): string {
  switch (format) {
    case "postman":
      return "Postman collection import";
    case "bruno":
      return "Bruno collection import";
    case "openapi":
      return "OpenAPI collection import";
    case "solo":
      return "Solo collection import";
  }
}

function getItemLabel(item: collection.BatchImportItemResult): string {
  return item.name || item.path || "Unknown source";
}

async function finishBatchImport(
  format: CollectionLocalImportFormat,
  result: collection.BatchImportResult
): Promise<LocalImportRunResult> {
  const items = result.results ?? [];
  const successes = items.filter((item) => item.success);
  const failures = items.filter((item) => !item.success);
  const warnings = items.flatMap((item) =>
    (item.warnings ?? []).map((warning) => ({
      source: getItemLabel(item),
      warning
    }))
  );
  const total = items.length;

  if (successes.length > 0) {
    await collectionStore.loadCollections();
  }

  const summaryLabel = getImportSummaryLabel(format);
  const detailParts = [
    `${successes.length}/${total} imported`,
    failures.length > 0 ? `${failures.length} failed` : null,
    warnings.length > 0 ? `${warnings.length} warning${warnings.length === 1 ? "" : "s"}` : null
  ].filter(Boolean);

  if (failures.length === 0) {
    notifications.success(summaryLabel, detailParts.join(", "));
  } else if (successes.length > 0) {
    notifications.warning(`${summaryLabel} finished with issues`, detailParts.join(", "));
  } else {
    notifications.error(`${summaryLabel} failed`, detailParts.join(", "));
  }

  for (const item of failures.slice(0, maxFailureDetailToasts)) {
    notifications.error(`Failed to import ${getItemLabel(item)}`, item.error || "Unknown error");
  }

  if (failures.length > maxFailureDetailToasts) {
    notifications.error(
      `${failures.length - maxFailureDetailToasts} more import failures`,
      "Review the source files and try again."
    );
  }

  for (const { source, warning } of warnings.slice(0, maxWarningDetailToasts)) {
    notifications.warning(`${source}: ${warning}`);
  }

  if (warnings.length > maxWarningDetailToasts) {
    notifications.warning(
      `${warnings.length - maxWarningDetailToasts} more import warnings`,
      "Additional warnings were omitted from notifications."
    );
  }

  return "completed";
}

async function executePostmanImport(paths: string[]): Promise<LocalImportRunResult> {
  return finishBatchImport("postman", await ImportPostmanCollections(paths));
}

async function executeBrunoImport(paths: string[]): Promise<LocalImportRunResult> {
  return finishBatchImport("bruno", await ImportBrunoCollections(paths));
}

async function executeOpenAPIImport(paths: string[]): Promise<LocalImportRunResult> {
  return finishBatchImport("openapi", await ImportOpenAPICollections(paths));
}

async function executeSoloCollectionImport(paths: string[]): Promise<LocalImportRunResult> {
  return finishBatchImport("solo", await ImportSoloCollections(paths));
}

const LOCAL_IMPORT_HANDLERS: Record<CollectionLocalImportFormat, LocalImportHandler> = {
  postman: {
    pick: () => SelectFiles("Select Postman Collections", "*.json", "JSON Files"),
    run: executePostmanImport,
    errorTitle: "Failed to import Postman collection"
  },
  bruno: {
    pick: async () => {
      const path = await SelectDirectory("Select Bruno Collection Folder");
      return path ? [path] : [];
    },
    run: executeBrunoImport,
    errorTitle: "Failed to import Bruno collection"
  },
  openapi: {
    pick: () =>
      SelectFiles(
        "Select OpenAPI / Swagger Documents",
        "*.json;*.yaml;*.yml",
        "OpenAPI / Swagger Files"
      ),
    run: executeOpenAPIImport,
    errorTitle: "Failed to import OpenAPI collection"
  },
  solo: {
    pick: () => SelectFiles("Select Solo Collections", "*.json", "JSON Files"),
    run: executeSoloCollectionImport,
    errorTitle: "Failed to import collection"
  }
};

function normalizeImportPaths(paths?: string | string[]): string[] {
  if (!paths) {
    return [];
  }

  return (Array.isArray(paths) ? paths : [paths]).filter((path) => path.length > 0);
}

function setPendingLocalImport(format: CollectionLocalImportFormat, paths: string[]) {
  if (paths.length === 0) {
    return;
  }

  collectionImportStoreState.selectedLocalFormat = format;
  collectionImportStoreState.pendingLocalImport = { format, paths };
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
    const paths = await LOCAL_IMPORT_HANDLERS[format].pick();

    if (paths.length === 0) {
      return;
    }

    setPendingLocalImport(format, paths);
  },

  async setPendingLocalImportFromDrop(
    format: CollectionLocalImportFormat,
    pathOrPaths?: string | string[]
  ) {
    const paths = normalizeImportPaths(pathOrPaths);
    if (paths.length === 0) {
      return;
    }

    setPendingLocalImport(format, paths);
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
      const result = await handler.run(pendingImport.paths);
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
    const paths = collectionImportStoreState.pendingLocalImport?.paths;
    if (!paths || paths.length === 0) {
      return;
    }

    let shouldKeepLoadingUntilReset = false;
    collectionImportStoreState.localImportLoading = true;
    collectionImportStoreState.pendingSoloCollectionPath = null;
    collectionImportStoreState.soloCollectionOverwriteName = null;

    try {
      const result = await executeSoloCollectionImport(paths);
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
