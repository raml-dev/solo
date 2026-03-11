export interface EnvAutocompleteEntry {
  key: string;
  value: string;
}

export interface EnvAutocompleteOptions {
  entries: EnvAutocompleteEntry[];
  trigger?: string;
  maxItems?: number;
  insertMode?: "value" | "token";
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
  let menu: HTMLDivElement | null = null;
  let open = false;
  let filtered: EnvAutocompleteEntry[] = [];
  let activeIndex = 0;
  let matchContext: MatchContext | null = null;

  function getMatchContext(): MatchContext | null {
    const caret = node.selectionStart;
    if (caret === null) return null;

    const trigger = currentOptions.trigger ?? DEFAULT_TRIGGER;
    const beforeCaret = node.value.slice(0, caret);
    const openIndex = beforeCaret.lastIndexOf(trigger);

    if (openIndex === -1) return null;

    const query = beforeCaret.slice(openIndex + trigger.length);

    // Ignore invalid token contexts to avoid noisy suggestions.
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
    if (!menu) return;

    const caret = node.selectionStart ?? node.value.length;
    const { left, top } = getCaretCoordinates(node, caret);
    const rect = node.getBoundingClientRect();
    const rawLineHeight = window.getComputedStyle(node).lineHeight;
    const lineHeight = Number.parseFloat(rawLineHeight);
    const safeLineHeight = Number.isFinite(lineHeight) ? lineHeight : 18;

    const menuLeft = Math.max(rect.left, left - 8);
    const menuTop = top + safeLineHeight + 6;
    const maxWidth = Math.max(220, rect.width);

    menu.style.left = `${window.scrollX + menuLeft}px`;
    menu.style.top = `${window.scrollY + menuTop}px`;
    menu.style.maxWidth = `${maxWidth}px`;
    menu.style.minWidth = "220px";
  }

  function closeMenu() {
    open = false;
    matchContext = null;

    if (menu) {
      menu.remove();
      menu = null;
    }
  }

  function applySelection(entry: EnvAutocompleteEntry) {
    if (!matchContext) return;

    const insertMode = currentOptions.insertMode ?? DEFAULT_INSERT_MODE;
    const replacement = insertMode === "token" ? `{{${entry.key}}}` : entry.value;

    node.setRangeText(replacement, matchContext.start, matchContext.end, "end");
    node.dispatchEvent(new Event("input", { bubbles: true }));
    closeMenu();
  }

  function renderMenu() {
    if (!open || !matchContext || filtered.length === 0) {
      closeMenu();
      return;
    }

    if (!menu) {
      menu = document.createElement("div");
      menu.className = "env-autocomplete-menu";
      document.body.appendChild(menu);
    }

    setMenuPosition();
    menu.innerHTML = "";

    filtered.forEach((entry, index) => {
      const item = document.createElement("button");
      item.type = "button";
      item.className = "env-autocomplete-item";
      if (index === activeIndex) {
        item.classList.add("active");
      }

      const keyLabel = document.createElement("span");
      keyLabel.className = "env-key";
      keyLabel.textContent = entry.key;

      item.appendChild(keyLabel);
      item.addEventListener("mousedown", (event) => {
        event.preventDefault();
        applySelection(entry);
      });
      menu?.appendChild(item);
    });
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
    activeIndex = Math.min(activeIndex, filtered.length - 1);
    renderMenu();
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

  function onKeyDown(event: KeyboardEvent) {
    if (!open || filtered.length === 0) return;

    if (event.key === "ArrowDown") {
      event.preventDefault();
      activeIndex = (activeIndex + 1) % filtered.length;
      renderMenu();
      return;
    }

    if (event.key === "ArrowUp") {
      event.preventDefault();
      activeIndex = (activeIndex - 1 + filtered.length) % filtered.length;
      renderMenu();
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
    if (target === node || node.contains(target) || (menu && menu.contains(target))) {
      return;
    }
    closeMenu();
  }

  function onWindowResizeOrScroll() {
    if (open) {
      setMenuPosition();
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
      currentOptions = nextOptions;
      activeIndex = 0;
      refreshFromCaret();
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
