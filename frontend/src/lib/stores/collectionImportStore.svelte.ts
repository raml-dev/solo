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
  lastLocalImportResult: collection.BatchImportResult | null;
  selectedOverwriteConflictPaths: string[];
}

type LocalImportRunResult = "completed" | "needs_review";

interface LocalImportHandler {
  pick: () => Promise<string[]>;
  run: (paths: string[], overwriteExisting: boolean) => Promise<LocalImportRunResult>;
  errorTitle: string;
}

const maxFailureDetailToasts = 3;
const maxWarningDetailToasts = 3;

const initialState: CollectionImportState = {
  selectedLocalFormat: "postman",
  pendingLocalImport: null,
  localImportLoading: false,
  lastLocalImportResult: null,
  selectedOverwriteConflictPaths: []
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
  result: collection.BatchImportResult,
  overwriteExisting: boolean
): Promise<LocalImportRunResult> {
  const items = result.results ?? [];
  const successes = items.filter((item) => item.success);
  const conflicts = items.filter((item) => !item.success && item.conflict);
  const failures = items.filter((item) => !item.success && !item.conflict);
  const warnings = items.flatMap((item) =>
    (item.warnings ?? []).map((warning) => ({
      source: getItemLabel(item),
      warning
    }))
  );
  const total = items.length;

  collectionImportStoreState.lastLocalImportResult = result;
  collectionImportStoreState.selectedOverwriteConflictPaths = conflicts.map((item) => item.path);

  if (successes.length > 0) {
    await collectionStore.loadCollections();
  }

  const summaryLabel = getImportSummaryLabel(format);
  const detailParts = [
    `${successes.length}/${total} imported`,
    conflicts.length > 0
      ? `${conflicts.length} conflict${conflicts.length === 1 ? "" : "s"}`
      : null,
    failures.length > 0 ? `${failures.length} failed` : null,
    warnings.length > 0 ? `${warnings.length} warning${warnings.length === 1 ? "" : "s"}` : null
  ].filter(Boolean);

  if (conflicts.length > 0 && !overwriteExisting) {
    notifications.warning(`${summaryLabel} needs review`, detailParts.join(", "));
  } else if (failures.length === 0) {
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

  return conflicts.length > 0 || failures.length > 0 || warnings.length > 0
    ? "needs_review"
    : "completed";
}

async function executePostmanImport(
  paths: string[],
  overwriteExisting: boolean
): Promise<LocalImportRunResult> {
  return finishBatchImport(
    "postman",
    await ImportPostmanCollections(paths, overwriteExisting),
    overwriteExisting
  );
}

async function executeBrunoImport(
  paths: string[],
  overwriteExisting: boolean
): Promise<LocalImportRunResult> {
  return finishBatchImport(
    "bruno",
    await ImportBrunoCollections(paths, overwriteExisting),
    overwriteExisting
  );
}

async function executeOpenAPIImport(
  paths: string[],
  overwriteExisting: boolean
): Promise<LocalImportRunResult> {
  return finishBatchImport(
    "openapi",
    await ImportOpenAPICollections(paths, overwriteExisting),
    overwriteExisting
  );
}

async function executeSoloCollectionImport(
  paths: string[],
  overwriteExisting: boolean
): Promise<LocalImportRunResult> {
  return finishBatchImport(
    "solo",
    await ImportSoloCollections(paths, overwriteExisting),
    overwriteExisting
  );
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
  collectionImportStoreState.lastLocalImportResult = null;
  collectionImportStoreState.selectedOverwriteConflictPaths = [];
}

export const collectionImportStoreState = $state<CollectionImportState>({ ...initialState });

export const collectionImportStore = {
  resetLocalImport() {
    collectionImportStoreState.selectedLocalFormat = initialState.selectedLocalFormat;
    collectionImportStoreState.pendingLocalImport = null;
    collectionImportStoreState.localImportLoading = false;
    collectionImportStoreState.lastLocalImportResult = null;
    collectionImportStoreState.selectedOverwriteConflictPaths = [];
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
    collectionImportStoreState.lastLocalImportResult = null;
    collectionImportStoreState.selectedOverwriteConflictPaths = [];
  },

  setConflictOverwriteSelection(path: string, selected: boolean) {
    const selectedPaths = collectionImportStoreState.selectedOverwriteConflictPaths;
    const alreadySelected = selectedPaths.includes(path);

    if (selected && !alreadySelected) {
      collectionImportStoreState.selectedOverwriteConflictPaths = [...selectedPaths, path];
      return;
    }

    if (!selected && alreadySelected) {
      collectionImportStoreState.selectedOverwriteConflictPaths = selectedPaths.filter(
        (selectedPath) => selectedPath !== path
      );
    }
  },

  selectAllConflictOverwrites() {
    collectionImportStoreState.selectedOverwriteConflictPaths = (
      collectionImportStoreState.lastLocalImportResult?.results ?? []
    )
      .filter((item) => item.conflict)
      .map((item) => item.path);
  },

  clearConflictOverwriteSelection() {
    collectionImportStoreState.selectedOverwriteConflictPaths = [];
  },

  async runPendingLocalImport() {
    const pendingImport = collectionImportStoreState.pendingLocalImport;
    if (!pendingImport) {
      return;
    }

    const handler = LOCAL_IMPORT_HANDLERS[pendingImport.format];
    let shouldKeepLoadingUntilReset = false;

    collectionImportStoreState.localImportLoading = true;
    collectionImportStoreState.lastLocalImportResult = null;
    collectionImportStoreState.selectedOverwriteConflictPaths = [];

    try {
      const result = await handler.run(pendingImport.paths, false);
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

  async runSelectedConflictOverwrites() {
    const pendingImport = collectionImportStoreState.pendingLocalImport;
    if (!pendingImport || collectionImportStoreState.selectedOverwriteConflictPaths.length === 0) {
      return;
    }

    const handler = LOCAL_IMPORT_HANDLERS[pendingImport.format];
    let shouldKeepLoadingUntilReset = false;
    collectionImportStoreState.localImportLoading = true;

    try {
      const result = await handler.run(
        collectionImportStoreState.selectedOverwriteConflictPaths,
        true
      );
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
  }
};
