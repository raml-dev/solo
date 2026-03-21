/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: GPL-3.0-only
 */

export interface EnvAutocompleteEntry {
  key: string;
  value: string;
}

export interface EnvAutocompleteRenderState {
  open: boolean;
  items: EnvAutocompleteEntry[];
  activeIndex: number;
  left: number;
  top: number;
  minWidth: number;
  maxWidth: number;
  query: string;
  select: (index: number) => void;
  setActive: (index: number) => void;
  close: () => void;
}

export interface EnvAutocompleteOptions {
  entries: EnvAutocompleteEntry[];
  trigger?: string;
  maxItems?: number;
  insertMode?: "value" | "token";
  menuElement?: HTMLElement | null;
  onStateChange?: (state: EnvAutocompleteRenderState) => void;
}

type TextFieldElement = HTMLInputElement | HTMLTextAreaElement;

interface MatchContext {
  start: number;
  end: number;
  query: string;
}

const DEFAULT_TRIGGER = "{{";
const DEFAULT_MAX_ITEMS = 8;
const DEFAULT_INSERT_MODE = "value";

function getCaretCoordinates(
  node: TextFieldElement,
  caretIndex: number
): { left: number; top: number } {
  const div = document.createElement("div");
  const span = document.createElement("span");
  const styles = window.getComputedStyle(node);

  div.style.position = "absolute";
  div.style.visibility = "hidden";
  div.style.whiteSpace = "pre-wrap";
  div.style.overflowWrap = "break-word";
  div.style.wordBreak = "break-word";
  div.style.pointerEvents = "none";
  div.style.left = "-9999px";
  div.style.top = "0";
  div.style.font = styles.font;
  div.style.fontFamily = styles.fontFamily;
  div.style.fontSize = styles.fontSize;
  div.style.fontWeight = styles.fontWeight;
  div.style.fontStyle = styles.fontStyle;
  div.style.letterSpacing = styles.letterSpacing;
  div.style.textTransform = styles.textTransform;
  div.style.textIndent = styles.textIndent;
  div.style.padding = styles.padding;
  div.style.border = styles.border;
  div.style.boxSizing = styles.boxSizing;
  div.style.lineHeight = styles.lineHeight;
  div.style.width = `${node.clientWidth}px`;

  const textBeforeCaret = node.value.slice(0, caretIndex);
  div.textContent = textBeforeCaret;

  // Keep width in place when caret is at end of line.
  span.textContent = "\u200b";
  div.appendChild(span);
  document.body.appendChild(div);

  const nodeRect = node.getBoundingClientRect();
  const caretRect = span.getBoundingClientRect();

  const left =
    nodeRect.left + (caretRect.left - div.getBoundingClientRect().left) - node.scrollLeft;
  const top = nodeRect.top + (caretRect.top - div.getBoundingClientRect().top) - node.scrollTop;

  div.remove();

  return { left, top };
}

export function envAutocomplete(node: TextFieldElement, options: EnvAutocompleteOptions) {
  let currentOptions = options;
  let open = false;
  let filtered: EnvAutocompleteEntry[] = [];
  let activeIndex = 0;
  let matchContext: MatchContext | null = null;
  let menuLeft = 0;
  let menuTop = 0;
  let menuMinWidth = 220;
  let menuMaxWidth = 220;

  function emitState() {
    currentOptions.onStateChange?.({
      open,
      items: filtered,
      activeIndex,
      left: menuLeft,
      top: menuTop,
      minWidth: menuMinWidth,
      maxWidth: menuMaxWidth,
      query: matchContext?.query ?? "",
      select: (index: number) => {
        const entry = filtered[index];
        if (entry) applySelection(entry);
      },
      setActive: (index: number) => {
        if (!filtered.length) return;
        activeIndex = Math.max(0, Math.min(index, filtered.length - 1));
        emitState();
      },
      close: closeMenu
    });
  }

  function getMatchContext(): MatchContext | null {
    const caret = node.selectionStart;
    if (caret === null) return null;

    const trigger = currentOptions.trigger ?? DEFAULT_TRIGGER;
    const beforeCaret = node.value.slice(0, caret);
    const openIndex = beforeCaret.lastIndexOf(trigger);

    if (openIndex === -1) return null;

    const query = beforeCaret.slice(openIndex + trigger.length);

    if (query.includes("}}") || query.includes("\n") || query.includes("\r")) {
      return null;
    }

    return {
      start: openIndex,
      end: caret,
      query
    };
  }

  function updateFiltered(context: MatchContext) {
    const query = context.query.trim().toLowerCase();
    const entries = currentOptions.entries ?? [];

    filtered = entries
      .filter((entry) => !query || entry.key.toLowerCase().includes(query))
      .sort((a, b) => a.key.localeCompare(b.key))
      .slice(0, currentOptions.maxItems ?? DEFAULT_MAX_ITEMS);
  }

  function setMenuPosition() {
    const caret = node.selectionStart ?? node.value.length;
    const { left, top } = getCaretCoordinates(node, caret);
    const rect = node.getBoundingClientRect();
    const rawLineHeight = window.getComputedStyle(node).lineHeight;
    const lineHeight = Number.parseFloat(rawLineHeight);
    const safeLineHeight = Number.isFinite(lineHeight) ? lineHeight : 18;

    menuLeft = window.scrollX + Math.max(rect.left, left - 8);
    menuTop = window.scrollY + top + safeLineHeight + 6;
    menuMaxWidth = Math.max(220, rect.width);
    menuMinWidth = 220;
  }

  function closeMenu() {
    open = false;
    matchContext = null;
    filtered = [];
    activeIndex = 0;
    emitState();
  }

  function applySelection(entry: EnvAutocompleteEntry) {
    if (!matchContext) return;

    const insertMode = currentOptions.insertMode ?? DEFAULT_INSERT_MODE;
    const replacement = insertMode === "token" ? `{{${entry.key}}}` : entry.value;

    node.setRangeText(replacement, matchContext.start, matchContext.end, "end");
    node.dispatchEvent(new Event("input", { bubbles: true }));
    closeMenu();
  }

  function refreshFromCaret() {
    const context = getMatchContext();

    if (!context) {
      closeMenu();
      return;
    }

    matchContext = context;
    updateFiltered(context);

    if (filtered.length === 0) {
      closeMenu();
      return;
    }

    open = true;
    activeIndex = Math.max(0, Math.min(activeIndex, filtered.length - 1));
    setMenuPosition();
    emitState();
  }

  function onInput() {
    refreshFromCaret();
  }

  function onClick() {
    refreshFromCaret();
  }

  function onKeyUp() {
    refreshFromCaret();
  }

  function onSelect() {
    refreshFromCaret();
  }

  function onKeyDown(event: Event) {
    if (!open || filtered.length === 0) return;
    if (!(event instanceof KeyboardEvent)) return;

    if (event.key === "ArrowDown") {
      event.preventDefault();
      activeIndex = (activeIndex + 1) % filtered.length;
      emitState();
      return;
    }

    if (event.key === "ArrowUp") {
      event.preventDefault();
      activeIndex = (activeIndex - 1 + filtered.length) % filtered.length;
      emitState();
      return;
    }

    if (event.key === "Enter" || event.key === "Tab") {
      event.preventDefault();
      applySelection(filtered[activeIndex]);
      return;
    }

    if (event.key === "Escape") {
      event.preventDefault();
      closeMenu();
    }
  }

  function onDocumentClick(event: MouseEvent) {
    if (!open) return;
    const target = event.target as Node;
    const menuEl = currentOptions.menuElement;

    if (target === node || node.contains(target) || (menuEl && menuEl.contains(target))) {
      return;
    }

    closeMenu();
  }

  function onWindowResizeOrScroll() {
    if (open) {
      setMenuPosition();
      emitState();
    }
  }

  node.addEventListener("input", onInput);
  node.addEventListener("click", onClick);
  node.addEventListener("keyup", onKeyUp);
  node.addEventListener("keydown", onKeyDown);
  node.addEventListener("select", onSelect);
  node.addEventListener("scroll", onWindowResizeOrScroll);
  document.addEventListener("mousedown", onDocumentClick);
  window.addEventListener("resize", onWindowResizeOrScroll);
  window.addEventListener("scroll", onWindowResizeOrScroll, true);

  return {
    update(nextOptions: EnvAutocompleteOptions) {
      const prev = currentOptions;
      currentOptions = nextOptions;

      const entriesChanged = prev.entries !== nextOptions.entries;
      const triggerChanged = prev.trigger !== nextOptions.trigger;
      const maxItemsChanged = prev.maxItems !== nextOptions.maxItems;
      const insertModeChanged = prev.insertMode !== nextOptions.insertMode;

      if (entriesChanged || triggerChanged || maxItemsChanged || insertModeChanged) {
        activeIndex = 0;
        refreshFromCaret();
        return;
      }

      const menuElementChanged = prev.menuElement !== nextOptions.menuElement;
      if (menuElementChanged && open) {
        emitState();
      }
    },
    destroy() {
      closeMenu();
      node.removeEventListener("input", onInput);
      node.removeEventListener("click", onClick);
      node.removeEventListener("keyup", onKeyUp);
      node.removeEventListener("keydown", onKeyDown);
      node.removeEventListener("select", onSelect);
      node.removeEventListener("scroll", onWindowResizeOrScroll);
      document.removeEventListener("mousedown", onDocumentClick);
      window.removeEventListener("resize", onWindowResizeOrScroll);
      window.removeEventListener("scroll", onWindowResizeOrScroll, true);
    }
  };
}
