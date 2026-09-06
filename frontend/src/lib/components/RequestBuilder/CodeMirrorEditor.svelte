<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import { jsonVars } from "$src/lib/codemirror/lang-json-vars";
  import { vsCodeDarkHighlightStyle } from "$src/lib/codemirror/themes/vscode/dark";
  import { vsCodeLightHighlightStyle } from "$src/lib/codemirror/themes/vscode/light";
  import EnvAutocompletePopover from "$src/lib/components/RequestBuilder/EnvAutocompletePopover.svelte";
  import { configurationStoreState } from "$src/lib/stores/configurationStore.svelte";
  import { hideTokenTooltipDelay, showTokenTooltip } from "$src/lib/stores/tokenTooltipStore";
  import {
    clampActiveIndex,
    createEnvTokenAutocompleteChange,
    createEnvTokenDecorationPlugin,
    createTokenizedEditorTheme,
    filterEnvTokenEntries,
    findEnvTokenTriggerContext,
    getTokenizedEditorSizeClass
  } from "$src/lib/utils/tokens";
  import {
    createResolvedVariableEntryMap,
    type ResolvedVariableEntry
  } from "$src/lib/utils/variableResolution";
  import { closeBrackets } from "@codemirror/autocomplete";
  import { defaultKeymap, history, historyKeymap } from "@codemirror/commands";
  import { xml } from "@codemirror/lang-xml";
  import {
    foldGutter,
    foldKeymap,
    HighlightStyle,
    StreamLanguage,
    syntaxHighlighting
  } from "@codemirror/language";
  import { lua } from "@codemirror/legacy-modes/mode/lua";
  import {
    closeSearchPanel,
    findNext,
    findPrevious,
    getSearchQuery,
    highlightSelectionMatches,
    openSearchPanel,
    replaceAll,
    replaceNext,
    search,
    SearchQuery,
    setSearchQuery
  } from "@codemirror/search";
  import { Annotation, Compartment, EditorSelection, EditorState, Prec } from "@codemirror/state";
  import {
    EditorView,
    highlightActiveLine,
    highlightActiveLineGutter,
    keymap,
    lineNumbers
  } from "@codemirror/view";
  import { tags } from "@lezer/highlight";
  import { indentationMarkers } from "@replit/codemirror-indentation-markers";
  import ArrowDownOutline from "flowbite-svelte-icons/ArrowDownOutline.svelte";
  import ArrowUpOutline from "flowbite-svelte-icons/ArrowUpOutline.svelte";
  import CheckOutline from "flowbite-svelte-icons/CheckOutline.svelte";
  import ClipboardCleanSolid from "flowbite-svelte-icons/ClipboardCleanSolid.svelte";
  import Button from "flowbite-svelte/Button.svelte";
  import Clipboard from "flowbite-svelte/Clipboard.svelte";
  import CloseButton from "flowbite-svelte/CloseButton.svelte";
  import Input from "flowbite-svelte/Input.svelte";
  import Toolbar from "flowbite-svelte/Toolbar.svelte";
  import ToolbarButton from "flowbite-svelte/ToolbarButton.svelte";
  import ToolbarGroup from "flowbite-svelte/ToolbarGroup.svelte";
  import { onDestroy, onMount } from "svelte";

  interface Props {
    value: string;
    format?: "none" | "json" | "xml" | "text";
    language?: "json" | "xml" | "lua" | "text" | "none" | "";
    variableEntries?: ResolvedVariableEntry[];
    readOnly?: boolean;
    size?: "sm" | "md" | "lg";
    onChange?: (value: string) => void;
    showCopyPaste?: boolean;
  }

  let {
    value = $bindable(""),
    format = $bindable("json"),
    language = "",
    variableEntries = [],
    readOnly = false,
    size = "md",
    onChange,
    showCopyPaste = false
  }: Props = $props();

  let success = $state(false);

  const config = $derived(configurationStoreState);

  // Annotation to mark transactions initiated by us (external sync) so the
  // updateListener does NOT re-emit "change" and cause an infinite loop.
  const externalUpdate = Annotation.define<boolean>();

  let editorEl: HTMLDivElement | undefined = $state();
  let view: EditorView | undefined = $state();

  let languageCompartment = new Compartment();
  let highlightThemeCompartment = new Compartment();
  let tokenDecoratorCompartment = new Compartment();

  let autocompleteOpen = $state(false);
  let autocompleteActiveIndex = $state(0);
  let autocompleteEntries: ResolvedVariableEntry[] = $state([]);
  let autocompleteLeft = $state(8);
  let autocompleteTop = $state(8);
  let autocompleteMaxWidth = $state(320);

  let autocompletePopoverStyle = $derived(
    `left: ${autocompleteLeft}px; top: ${autocompleteTop}px; min-width: 220px; max-width: ${autocompleteMaxWidth}px;`
  );

  let sizeClass = $derived(getTokenizedEditorSizeClass(size));
  let variableEntryMap = $derived(createResolvedVariableEntryMap(variableEntries));
  let knownVariableKeys = $derived(new Set(variableEntries.map((entry) => entry.key)));
  let knownKeysSignature = $derived(
    variableEntries.map((entry) => `${entry.key}:${entry.winningSource}`).join("\u0000")
  );

  let searchToolbarOpen = $state(false);
  let searchQuery = $state("");
  let replaceQuery = $state("");
  let searchCaseSensitive = $state(false);
  let searchRegexp = $state(false);
  let searchWholeWord = $state(false);
  let searchInputEl: HTMLInputElement | undefined = $state();
  let replaceInputEl: HTMLInputElement | undefined = $state();

  // Custom search panel factory that syncs with our Svelte UI
  function createSearchPanel() {
    return {
      dom: createSearchPanelDOM(),
      mount() {
        // Panel mounted - sync initial state
        syncPanelToEditor();
      }
    };
  }

  // Create a minimal panel DOM element (will be hidden, we use our Svelte UI)
  function createSearchPanelDOM(): HTMLElement {
    const dom = document.createElement("div");
    dom.style.display = "none"; // Hide the default panel, we use our custom UI
    return dom;
  }

  function applySearchQuery() {
    if (!view) return;

    view.dispatch({
      effects: setSearchQuery.of(
        new SearchQuery({
          search: searchQuery,
          replace: replaceQuery,
          caseSensitive: searchCaseSensitive,
          regexp: searchRegexp,
          wholeWord: searchWholeWord
        })
      )
    });
  }

  function syncPanelToEditor() {
    queueMicrotask(() => {
      applySearchQuery();
    });
  }

  function initializeSearchToolbarFromEditor() {
    if (!view) return;

    const currentQuery = getSearchQuery(view.state);
    searchQuery = currentQuery.search;
    replaceQuery = currentQuery.replace;
    searchCaseSensitive = currentQuery.caseSensitive;
    searchRegexp = currentQuery.regexp;
    searchWholeWord = currentQuery.wholeWord;

    if (!searchQuery) {
      const selectedText = view.state.sliceDoc(
        view.state.selection.main.from,
        view.state.selection.main.to
      );
      if (selectedText) {
        searchQuery = selectedText;
      }
    }

    applySearchQuery();
  }

  function focusSearchInput() {
    queueMicrotask(() => {
      searchInputEl?.focus();
      searchInputEl?.select();
    });
  }

  function openSearchToolbar() {
    if (!view) return false;

    searchToolbarOpen = true;
    initializeSearchToolbarFromEditor();
    focusSearchInput();

    // Open the search panel to enable highlighting (deferred to avoid update-in-progress error)
    queueMicrotask(() => {
      if (view) {
        openSearchPanel(view);
      }
    });
    return true;
  }

  function closeSearchToolbar() {
    searchToolbarOpen = false;
    view?.focus();

    // Close the search panel
    if (view) {
      closeSearchPanel(view);
    }
  }

  function runFindNext() {
    if (!view) return;
    findNext(view);
  }

  function runFindPrevious() {
    if (!view) return;
    findPrevious(view);
  }

  function runReplaceNext() {
    if (!view) return;
    replaceNext(view);
  }

  function runReplaceAll() {
    if (!view) return;
    replaceAll(view);
  }

  const searchToolbarKeymap = Prec.highest(
    keymap.of([
      {
        key: "Mod-f",
        run: () => openSearchToolbar()
      },
      {
        key: "F3",
        run: () => {
          runFindNext();
          return true;
        }
      },
      {
        key: "Shift-F3",
        run: () => {
          runFindPrevious();
          return true;
        }
      },
      {
        key: "Mod-h",
        run: () => {
          if (!openSearchToolbar()) return false;
          queueMicrotask(() => replaceInputEl?.focus());
          return true;
        }
      }
    ])
  );

  function estimatePopoverHeight(entryCount: number): number {
    if (entryCount <= 0) return 40;

    // Mirror visual constraints from EnvAutocompletePopover:
    // - max-h-56 list (~224px)
    // - per-row button height roughly ~30px including spacing
    const visibleRows = Math.min(entryCount, 7);
    const estimatedRowsHeight = visibleRows * 30;
    return Math.min(224, estimatedRowsHeight) + 16;
  }

  function updateAutocompletePosition(cursorPos: number, entryCount: number) {
    if (!view || !editorEl) return;

    const caretRect = view.coordsAtPos(cursorPos);
    if (!caretRect) return;

    const containerRect = editorEl.getBoundingClientRect();
    const left = caretRect.left - containerRect.left;
    const belowTop = caretRect.bottom - containerRect.top + 6;
    const estimatedPopoverHeight = estimatePopoverHeight(entryCount);

    const belowSpace = containerRect.height - belowTop - 8;
    const aboveSpace = caretRect.top - containerRect.top - 8;

    // Prefer below when there is enough room, or when below has more room than above.
    const placeBelow = belowSpace >= estimatedPopoverHeight || belowSpace >= aboveSpace;

    const aboveTop = caretRect.top - containerRect.top - estimatedPopoverHeight - 6;
    const top = placeBelow ? belowTop : aboveTop;

    autocompleteLeft = Math.max(8, Math.min(left, Math.max(8, containerRect.width - 240)));
    autocompleteTop = Math.max(8, top);
    autocompleteMaxWidth = Math.max(220, containerRect.width - 16);
  }

  function refreshAutocomplete(state: EditorState) {
    if (readOnly) {
      autocompleteOpen = false;
      return;
    }

    const triggerContext = findEnvTokenTriggerContext(
      state.doc.toString(),
      state.selection.main.head
    );
    if (!triggerContext) {
      autocompleteOpen = false;
      autocompleteEntries = [];
      autocompleteActiveIndex = 0;
      return;
    }

    const normalizedQuery = triggerContext.normalizedQuery;
    autocompleteEntries = filterEnvTokenEntries(variableEntries, normalizedQuery);
    autocompleteActiveIndex = clampActiveIndex(autocompleteActiveIndex, autocompleteEntries.length);
    autocompleteOpen = true;
    updateAutocompletePosition(triggerContext.to, autocompleteEntries.length);
  }

  function applyAutocompleteEntry(entry: ResolvedVariableEntry) {
    if (!view) return;

    const completionChange = createEnvTokenAutocompleteChange(
      view.state.doc.toString(),
      view.state.selection.main.head,
      entry.key
    );
    if (!completionChange) return;

    view.dispatch({
      changes: completionChange,
      selection: EditorSelection.cursor(completionChange.from + completionChange.insert.length)
    });

    autocompleteOpen = false;
    autocompleteEntries = [];
    autocompleteActiveIndex = 0;
  }

  const autocompleteNavigationKeymap = Prec.highest(
    keymap.of([
      {
        key: "ArrowDown",
        run: () => {
          if (!autocompleteOpen || autocompleteEntries.length === 0) return false;
          autocompleteActiveIndex = (autocompleteActiveIndex + 1) % autocompleteEntries.length;
          return true;
        }
      },
      {
        key: "ArrowUp",
        run: () => {
          if (!autocompleteOpen || autocompleteEntries.length === 0) return false;
          autocompleteActiveIndex =
            (autocompleteActiveIndex - 1 + autocompleteEntries.length) % autocompleteEntries.length;
          return true;
        }
      },
      {
        key: "Enter",
        run: () => {
          if (!autocompleteOpen || autocompleteEntries.length === 0) return false;
          const entry = autocompleteEntries[autocompleteActiveIndex];
          if (!entry) return false;
          applyAutocompleteEntry(entry);
          return true;
        }
      },
      {
        key: "Tab",
        run: () => {
          if (!autocompleteOpen || autocompleteEntries.length === 0) return false;
          const entry = autocompleteEntries[autocompleteActiveIndex];
          if (!entry) return false;
          applyAutocompleteEntry(entry);
          return true;
        }
      },
      {
        key: "Escape",
        run: () => {
          if (!autocompleteOpen) return false;
          autocompleteOpen = false;
          return true;
        }
      }
    ])
  );

  // --- Lifecycle ---
  onMount(() => {
    const extensions = [
      lineNumbers(),
      highlightActiveLineGutter(),
      foldGutter(),
      indentationMarkers(),
      highlightActiveLine(),
      search({ createPanel: createSearchPanel }),
      highlightSelectionMatches(),
      closeBrackets(),
      searchToolbarKeymap,
      keymap.of([...defaultKeymap, ...foldKeymap]),
      languageCompartment.of(getLangExtension()),
      syntaxHighlighting(customHighlightStyle),
      highlightThemeCompartment.of(getHighlightThemeExtension()),
      createTokenizedEditorTheme(),
      appTheme,
      tokenDecoratorCompartment.of(readOnly ? [] : createTokenHighlightPlugin()),
      // --- edit-only extensions ---
      ...(!readOnly
        ? [
            history(),
            autocompleteNavigationKeymap,
            keymap.of([...historyKeymap]),
            EditorView.updateListener.of((update) => {
              if (
                update.docChanged &&
                !update.transactions.some((tr) => tr.annotation(externalUpdate))
              ) {
                onChange?.(update.state.doc.toString());
              }

              if (update.docChanged || update.selectionSet) {
                refreshAutocomplete(update.state);
              }
            })
          ]
        : [
            EditorState.readOnly.of(true),
            EditorView.editable.of(false),
            // Make readOnly editor focusable so keymaps work
            EditorView.contentAttributes.of({ tabindex: "0" })
          ])
    ];

    const state = EditorState.create({
      doc: value ?? "",
      extensions
    });

    view = new EditorView({ state, parent: editorEl });
    refreshAutocomplete(state);

    return () => view?.destroy();
  });

  onDestroy(() => view?.destroy());

  // When format/language changes, reconfigure language extension (guarded)
  let _lastLang = $state("");

  function getLangExtension() {
    // explicit language prop takes priority
    const lang = language || format;
    if (lang === "json") return jsonVars();
    if (lang === "xml") return xml();
    if (lang === "lua") return StreamLanguage.define(lua);
    return [];
  }

  function getHighlightThemeExtension() {
    return syntaxHighlighting(
      config.appliedThemeMode === "dark" ? vsCodeDarkHighlightStyle : vsCodeLightHighlightStyle
    );
  }

  // --- Extensions ---

  const customHighlightStyle = HighlightStyle.define([
    { tag: tags.propertyName, color: "var(--color-primary-600)" }
  ]);

  // Token highlighter — active in edit mode via tokenDecoratorCompartment
  function createTokenHighlightPlugin(_knownKeysSignature?: string) {
    void _knownKeysSignature;

    return createEnvTokenDecorationPlugin({
      tokenClassName: "cm-env-token",
      resolveTokenStatus: (tokenKey) =>
        variableEntryMap.get(tokenKey)?.winningSource === "session"
          ? "session"
          : variableEntryMap.get(tokenKey)?.winningSource === "collection"
            ? "collection"
            : knownVariableKeys.has(tokenKey)
              ? "known"
              : "unknown",
      onTokenMouseOver: (tokenKey, rect) => showTokenTooltip(tokenKey, rect.left, rect.bottom),
      onTokenMouseOut: () => hideTokenTooltipDelay()
    });
  }

  const appTheme = EditorView.theme({
    "&": {
      height: "100%"
    },
    ".cm-foldGutter .cm-gutterElement": {
      cursor: "pointer"
    }
  });
  // When value prop changes from outside, sync to editor WITHOUT triggering change event
  $effect(() => {
    if (view && value !== view.state.doc.toString()) {
      view.dispatch({
        changes: { from: 0, to: view.state.doc.length, insert: value ?? "" },
        annotations: [externalUpdate.of(true)]
      });
    }
  });
  $effect(() => {
    const effectiveLang = language || format;
    if (view && effectiveLang !== _lastLang) {
      _lastLang = effectiveLang;
      view.dispatch({
        effects: languageCompartment.reconfigure(getLangExtension())
      });
    }
  });

  $effect(() => {
    if (!view) return;

    view.dispatch({
      effects: highlightThemeCompartment.reconfigure(getHighlightThemeExtension())
    });
  });

  $effect(() => {
    if (!view) return;
    refreshAutocomplete(view.state);
  });

  $effect(() => {
    if (!view) return;

    view.dispatch({
      effects: tokenDecoratorCompartment.reconfigure(
        readOnly ? [] : createTokenHighlightPlugin(knownKeysSignature)
      )
    });
  });

  // Sync search query state with editor - REMOVED: handled by oninput={applySearchQuery} on Input
</script>

<div class="relative h-full {sizeClass}">
  {#if searchToolbarOpen}
    <div class="absolute top-2 right-2 z-30 max-w-[calc(100%-1rem)]">
      <Toolbar
        color="default"
        class="flex-wrap gap-2 rounded-lg border border-neutral-200 p-2 shadow-sm dark:border-neutral-700"
      >
        <div class="flex flex-col">
          <ToolbarGroup class="min-w-0 flex-1" padding="none">
            <Input
              placeholder="Find"
              bind:value={searchQuery}
              bind:elementRef={searchInputEl}
              oninput={applySearchQuery}
              onkeydown={(event) => {
                if (event.key === "Escape") {
                  event.preventDefault();
                  closeSearchToolbar();
                  return;
                }

                if (event.key === "Enter") {
                  event.preventDefault();
                  if (event.shiftKey) runFindPrevious();
                  else runFindNext();
                }
              }}
              class="h-6 min-w-40 flex-1"
            >
              {#snippet right()}
                <Button
                  color="light"
                  class="m-0.5 shrink-0 rounded-lg border-none bg-transparent p-1 whitespace-normal hover:bg-transparent focus:ring-0 focus:ring-transparent focus:outline-hidden dark:bg-transparent dark:text-gray-400 dark:hover:bg-transparent
                    {searchCaseSensitive
                    ? ' text-primary-700  dark:text-primary-300'
                    : 'dark:hover:text-gray-50'}"
                  aria-label="Toggle case sensitive"
                  onclick={() => {
                    searchCaseSensitive = !searchCaseSensitive;
                    applySearchQuery();
                  }}
                >
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    // width="24"
                    // height="24"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    class="h-4 w-4"
                    ><path d="m2 16 4.039-9.69a.5.5 0 0 1 .923 0L11 16" /><path d="M22 9v7" /><path
                      d="M3.304 13h6.392"
                    /><circle cx="18.5" cy="12.5" r="3.5" /></svg
                  >
                </Button>
                <Button
                  color="light"
                  class="m-0.5 shrink-0 rounded-lg border-none bg-transparent p-1 whitespace-normal hover:bg-transparent focus:ring-0 focus:ring-transparent focus:outline-hidden dark:bg-transparent dark:text-gray-400 dark:hover:bg-transparent
                    {searchWholeWord
                    ? ' text-primary-700  dark:text-primary-300'
                    : 'dark:hover:text-gray-50'}"
                  aria-label="Toggle case sensitive"
                  onclick={() => {
                    searchWholeWord = !searchWholeWord;
                    applySearchQuery();
                  }}
                >
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    // width="24"
                    // height="24"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    class="h-4 w-4"
                    ><circle cx="7" cy="12" r="3" /><path d="M10 9v6" /><circle
                      cx="17"
                      cy="12"
                      r="3"
                    /><path d="M14 7v8" /><path
                      d="M22 17v1c0 .5-.5 1-1 1H3c-.5 0-1-.5-1-1v-1"
                    /></svg
                  >
                </Button>
                <Button
                  color="light"
                  class="m-0.5 shrink-0 rounded-lg border-none bg-transparent p-1 whitespace-normal hover:bg-transparent focus:ring-0 focus:ring-transparent focus:outline-hidden dark:bg-transparent dark:text-gray-400 dark:hover:bg-transparent
                    {searchRegexp
                    ? ' text-primary-700  dark:text-primary-300'
                    : 'dark:hover:text-gray-50'}"
                  aria-label="Toggle case sensitive"
                  onclick={() => {
                    searchRegexp = !searchRegexp;
                    applySearchQuery();
                  }}
                >
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    // width="24"
                    // height="24"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    class="h-4 w-4"
                    ><path d="M17 3v10" /><path d="m12.67 5.5 8.66 5" /><path
                      d="m12.67 10.5 8.66-5"
                    /><path
                      d="M9 17a2 2 0 0 0-2-2H5a2 2 0 0 0-2 2v2a2 2 0 0 0 2 2h2a2 2 0 0 0 2-2v-2z"
                    /></svg
                  >
                </Button>
              {/snippet}
            </Input>
            <div class="flex gap-1">
              <ToolbarButton
                class="p-1"
                color="default"
                aria-label="Find previous"
                onclick={runFindPrevious}><ArrowUpOutline class="h-4 w-4 shrink-0" /></ToolbarButton
              >
              <ToolbarButton
                class="p-1"
                color="default"
                aria-label="Find next"
                onclick={runFindNext}><ArrowDownOutline class="h-4 w-4 shrink-0" /></ToolbarButton
              >
            </div>
          </ToolbarGroup>

          <ToolbarGroup class="min-w-0 flex-1" padding="none">
            <Input
              placeholder="Replace"
              bind:value={replaceQuery}
              bind:elementRef={replaceInputEl}
              oninput={applySearchQuery}
              onkeydown={(event) => {
                if (event.key === "Escape") {
                  event.preventDefault();
                  closeSearchToolbar();
                  return;
                }

                if (event.key === "Enter") {
                  event.preventDefault();
                  runReplaceNext();
                }
              }}
              class="h-6 min-w-40 flex-1"
            />
            <div class="flex gap-1">
              <ToolbarButton
                class="p-1"
                color="default"
                aria-label="Replace next"
                onclick={runReplaceNext}
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  // width="24"
                  // height="24"
                  class="h-4 w-4"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  ><path d="M14 4a1 1 0 0 1 1-1" /><path d="M15 10a1 1 0 0 1-1-1" /><path
                    d="M21 4a1 1 0 0 0-1-1"
                  /><path d="M21 9a1 1 0 0 1-1 1" /><path d="m3 7 3 3 3-3" /><path
                    d="M6 10V5a2 2 0 0 1 2-2h2"
                  /><rect x="3" y="14" width="7" height="7" rx="1" /></svg
                ></ToolbarButton
              >
              <ToolbarButton
                class="p-1"
                color="default"
                aria-label="Replace all"
                onclick={runReplaceAll}
                ><svg
                  xmlns="http://www.w3.org/2000/svg"
                  // width="24"
                  // height="24"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  class="h-4 w-4"
                  ><path d="M14 14a1 1 0 0 1 1 1v5a1 1 0 0 1-1 1" /><path
                    d="M14 4a1 1 0 0 1 1-1"
                  /><path d="M15 10a1 1 0 0 1-1-1" /><path
                    d="M19 14a1 1 0 0 1 1 1v5a1 1 0 0 1-1 1"
                  /><path d="M21 4a1 1 0 0 0-1-1" /><path d="M21 9a1 1 0 0 1-1 1" /><path
                    d="m3 7 3 3 3-3"
                  /><path d="M6 10V5a2 2 0 0 1 2-2h2" /><rect
                    x="3"
                    y="14"
                    width="7"
                    height="7"
                    rx="1"
                  /></svg
                ></ToolbarButton
              >
            </div>
          </ToolbarGroup>
        </div>
        <ToolbarGroup spacing="tight" padding="none" class="ml-auto">
          <ToolbarButton color="default" aria-label="Close search" onclick={closeSearchToolbar}
            ><CloseButton
              tabindex={0}
              color="none"
              size="xs"
              class="p-1! opacity-70 inset-ring-primary-500 transition-opacity group-hover:opacity-100 focus-within:inset-ring-1 focus-within:outline-hidden hover:opacity-100 focus:outline-hidden"
            /></ToolbarButton
          >
        </ToolbarGroup>
      </Toolbar>
    </div>
  {/if}

  <div class="h-full">
    <div class="editor-container h-full" class:read-only={readOnly} bind:this={editorEl}></div>
  </div>

  <EnvAutocompletePopover
    open={autocompleteOpen}
    entries={autocompleteEntries}
    activeIndex={autocompleteActiveIndex}
    class="absolute z-20"
    style={autocompletePopoverStyle}
    onHoverIndex={(index) => (autocompleteActiveIndex = index)}
    onSelect={(entry) => applyAutocompleteEntry(entry)}
    onRequestClose={() => (autocompleteOpen = false)}
  />

  {#if showCopyPaste}
    <Clipboard
      color={success ? "alternative" : "light"}
      bind:success
      {value}
      size="sm"
      class={`absolute inset-e-2 h-8 px-2.5 font-medium focus:ring-0 ${searchToolbarOpen ? "top-24" : "top-2"}`}
    >
      {#if success}
        <CheckOutline class="h-3 w-3" /> Copied
      {:else}
        <ClipboardCleanSolid class="h-3 w-3" /> Copy code
      {/if}
    </Clipboard>
  {/if}
</div>
