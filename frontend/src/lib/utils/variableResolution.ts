/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: AGPL-3.0-only
 */

export type VariableSource = "session" | "environment" | "collection";

export interface VariableValueLike {
  value?: unknown;
  type?: unknown;
}

export interface NormalizedVariableValue {
  value: string;
  type: string;
}

export interface ResolvedVariableEntry {
  key: string;
  computedValue: string;
  winningSource: VariableSource;
  winningType: string;
  hasConflicts: boolean;
  definedIn: VariableSource[];
  sourceValues: Partial<Record<VariableSource, NormalizedVariableValue>>;
}

interface ResolveVariableEntriesOptions {
  sessionValues?: Record<string, string> | null;
  environmentValues?: Record<string, VariableValueLike> | null;
  collectionValues?: Record<string, VariableValueLike> | null;
}

function hasOwnKey(record: Record<string, unknown>, key: string): boolean {
  return Object.prototype.hasOwnProperty.call(record, key);
}

export function normalizeVariableValues(
  values: Record<string, VariableValueLike> | null | undefined
): Record<string, NormalizedVariableValue> {
  return Object.fromEntries(
    Object.entries(values ?? {}).map(([key, entry]) => [
      key,
      {
        value: String(entry?.value ?? ""),
        type: String(entry?.type ?? "default")
      }
    ])
  );
}

export function resolveVariableEntries({
  sessionValues = {},
  environmentValues,
  collectionValues
}: ResolveVariableEntriesOptions): ResolvedVariableEntry[] {
  const normalizedSessionValues = sessionValues ?? {};
  const normalizedEnvironmentValues = normalizeVariableValues(environmentValues);
  const normalizedCollectionValues = normalizeVariableValues(collectionValues);

  const keys = new Set<string>([
    ...Object.keys(normalizedSessionValues),
    ...Object.keys(normalizedEnvironmentValues),
    ...Object.keys(normalizedCollectionValues)
  ]);

  return [...keys]
    .sort((left, right) => left.localeCompare(right))
    .map((key) => {
      const sourceValues: ResolvedVariableEntry["sourceValues"] = {};
      const definedIn: VariableSource[] = [];

      if (hasOwnKey(normalizedSessionValues, key)) {
        definedIn.push("session");
        sourceValues.session = {
          value: String(normalizedSessionValues[key] ?? ""),
          type: "session"
        };
      }

      if (hasOwnKey(normalizedEnvironmentValues, key)) {
        definedIn.push("environment");
        sourceValues.environment = normalizedEnvironmentValues[key];
      }

      if (hasOwnKey(normalizedCollectionValues, key)) {
        definedIn.push("collection");
        sourceValues.collection = normalizedCollectionValues[key];
      }

      const winningSource =
        definedIn.find((source) => source === "session") ??
        definedIn.find((source) => source === "environment") ??
        "collection";
      const winningValue = sourceValues[winningSource];

      return {
        key,
        computedValue: winningValue?.value ?? "",
        winningSource,
        winningType: winningValue?.type ?? "default",
        hasConflicts: definedIn.length > 1,
        definedIn,
        sourceValues
      };
    });
}

export function createResolvedVariableEntryMap(
  entries: ResolvedVariableEntry[]
): Map<string, ResolvedVariableEntry> {
  return new Map(entries.map((entry) => [entry.key, entry]));
}

export function resolveVariableTokens(
  value: string,
  entriesOrMap: ResolvedVariableEntry[] | Map<string, ResolvedVariableEntry>
): string {
  if (!value) {
    return value;
  }

  const entryMap =
    entriesOrMap instanceof Map ? entriesOrMap : createResolvedVariableEntryMap(entriesOrMap);

  return value.replace(/\{\{([^{}\r\n]+?)\}\}/g, (full, rawKey: string) => {
    const key = rawKey.trim();
    const entry = entryMap.get(key);
    return entry ? entry.computedValue : full;
  });
}

export function formatVariableSourceLabel(source: VariableSource): string {
  return source === "session" ? "Session" : source === "environment" ? "Environment" : "Collection";
}
