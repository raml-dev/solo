/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: AGPL-3.0-only
 */

import { Decoration, EditorView, ViewPlugin, type DecorationSet } from "@codemirror/view";

export interface TextSegment {
  text: string;
  isToken: boolean;
  tokenKey?: string;
}

export interface EnvTokenMatch {
  full: string;
  key: string;
  from: number;
  to: number;
}

export interface EnvTokenTriggerContext {
  from: number;
  to: number;
  rawQuery: string;
  normalizedQuery: string;
}

export type TokenizedEditorSize = "sm" | "md" | "lg";

export interface EnvTokenEntry {
  key: string;
  value?: string;
  type?: string;
}

export interface EnvironmentTokenValueLike {
  value?: unknown;
  type?: unknown;
}

export type EnvTokenStatus = "known" | "unknown" | "session";

export function getTokenizedEditorSizeClass(size: TokenizedEditorSize = "md"): string {
  return size === "sm"
    ? "text-xs leading-4"
    : size === "lg"
      ? "sm:text-base leading-6"
      : "text-sm leading-5";
}

export function normalizeEnvironmentTokenEntries(
  values: Record<string, EnvironmentTokenValueLike> | null | undefined
): Required<EnvTokenEntry>[] {
  return Object.entries(values ?? {})
    .map(([key, valueType]) => ({
      key,
      value: String(valueType?.value ?? ""),
      type: String(valueType?.type ?? "")
    }))
    .sort((a, b) => a.key.localeCompare(b.key));
}

interface EnvTokenDecorationPluginOptions {
  tokenClassName: string;
  resolveTokenStatus?: (tokenKey: string) => EnvTokenStatus;
  onTokenMouseOver?: (tokenKey: string, rect: DOMRect) => void;
  onTokenMouseOut?: () => void;
}

interface TokenizedEditorThemeOptions {
  singleLine?: boolean;
}

const ENV_TOKEN_REGEX = /\{\{([^{}\r\n]+?)\}\}/g;

export function createEnvTokenSnippet(key: string): string {
  return `{{${key}}}`;
}

export function extractEnvTokenMatches(text: string): EnvTokenMatch[] {
  if (!text) return [];

  const matches: EnvTokenMatch[] = [];

  for (const match of text.matchAll(ENV_TOKEN_REGEX)) {
    const from = match.index ?? 0;
    const full = match[0];
    matches.push({
      full,
      key: match[1].trim(),
      from,
      to: from + full.length
    });
  }

  return matches;
}

export function findEnvTokenTriggerContext(
  text: string,
  cursor: number
): EnvTokenTriggerContext | null {
  const safeText = text ?? "";
  const safeCursor = Math.max(0, Math.min(cursor, safeText.length));
  const beforeCursor = safeText.slice(0, safeCursor);
  const from = beforeCursor.lastIndexOf("{{");

  if (from === -1) return null;

  const rawQuery = beforeCursor.slice(from + 2);

  // Trigger is invalid if a close sequence already exists in the active fragment
  // or if the fragment contains unsupported braces/newlines.
  if (rawQuery.includes("}}") || /[{}\r\n]/.test(rawQuery)) return null;

  return {
    from,
    to: safeCursor,
    rawQuery,
    normalizedQuery: rawQuery.trim().toLowerCase()
  };
}

export function filterEnvTokenEntries<T extends EnvTokenEntry>(
  entries: T[],
  normalizedQuery: string,
  maxItems?: number
): T[] {
  const safeQuery = (normalizedQuery ?? "").trim().toLowerCase();

  const filtered = (entries ?? [])
    .filter((entry) => !safeQuery || entry.key.toLowerCase().includes(safeQuery))
    .sort((a, b) => a.key.localeCompare(b.key));

  if (!maxItems || maxItems <= 0) return filtered;
  return filtered.slice(0, maxItems);
}

export function clampActiveIndex(index: number, listLength: number): number {
  if (listLength <= 0) return 0;
  return Math.min(Math.max(0, index), listLength - 1);
}

export function createTokenizedEditorTheme(options: TokenizedEditorThemeOptions = {}) {
  const { singleLine = false } = options;

  // Keep only mode-dependent behavior here; shared visual styling lives in
  // src/assets/styles/codemirror-theme.css.
  return EditorView.theme({
    ".cm-scroller": {
      overflowY: singleLine ? "hidden" : "auto"
    }
  });
}

export function createEnvTokenDecorationPlugin(options: EnvTokenDecorationPluginOptions) {
  const { tokenClassName, resolveTokenStatus, onTokenMouseOver, onTokenMouseOut } = options;
  const tokenSelector = `.${tokenClassName}`;

  return ViewPlugin.fromClass(
    class {
      decorations: DecorationSet;

      constructor(v: EditorView) {
        this.decorations = this.build(v);
      }

      update(u: { view: EditorView; docChanged: boolean }) {
        if (u.docChanged) this.decorations = this.build(u.view);
      }

      build(v: EditorView) {
        const builder: import("@codemirror/state").Range<Decoration>[] = [];
        const text = v.state.doc.toString();

        for (const match of extractEnvTokenMatches(text)) {
          const status = resolveTokenStatus?.(match.key) ?? "known";
          const tokenStatusClass = `${tokenClassName}--${status}`;

          builder.push(
            Decoration.mark({
              class: `${tokenClassName} ${tokenStatusClass}`,
              attributes: { "data-token-key": match.key }
            }).range(match.from, match.to)
          );
        }

        return Decoration.set(builder);
      }
    },
    {
      decorations: (v) => v.decorations,
      eventHandlers: {
        mouseover: (event: MouseEvent) => {
          if (!onTokenMouseOver) return;

          const target = event.target as HTMLElement;
          const tokenEl = target.closest(tokenSelector) as HTMLElement | null;
          if (!tokenEl) return;

          const tokenKey = tokenEl.dataset.tokenKey;
          if (!tokenKey) return;

          onTokenMouseOver(tokenKey, tokenEl.getBoundingClientRect());
        },
        mouseout: (event: MouseEvent) => {
          if (!onTokenMouseOut) return;

          const target = event.target as HTMLElement;
          const tokenEl = target.closest(tokenSelector) as HTMLElement | null;
          if (!tokenEl) return;

          onTokenMouseOut();
        },
        mouseleave: () => {
          onTokenMouseOut?.();
        }
      }
    }
  );
}

export function splitTextSegments(value: string): TextSegment[] {
  if (!value) return [];

  const segments: TextSegment[] = [];
  let cursor = 0;

  for (const match of extractEnvTokenMatches(value)) {
    if (match.from > cursor) {
      segments.push({ text: value.slice(cursor, match.from), isToken: false });
    }

    segments.push({ text: match.full, isToken: true, tokenKey: match.key });
    cursor = match.to;
  }

  if (cursor < value.length) {
    segments.push({ text: value.slice(cursor), isToken: false });
  }

  return segments;
}
