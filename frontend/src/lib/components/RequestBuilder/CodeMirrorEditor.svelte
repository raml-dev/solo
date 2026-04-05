<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: GPL-3.0-only
-->

<script lang="ts">
  import EnvAutocompletePopover from "$src/lib/components/RequestBuilder/EnvAutocompletePopover.svelte";
  import {
    clampActiveIndex,
    createEnvTokenDecorationPlugin,
    createEnvTokenSnippet,
    createTokenizedEditorTheme,
    filterEnvTokenEntries,
    findEnvTokenTriggerContext,
    getTokenizedEditorSizeClass
  } from "$src/lib/utils/tokens";
  import { hideTokenTooltipDelay, showTokenTooltip } from "$src/lib/stores/tokenTooltipStore";
  import { sessionVarsStore } from "$src/lib/stores/sessionVarsStore";
  import { defaultKeymap, history, historyKeymap } from "@codemirror/commands";
  import { json } from "@codemirror/lang-json";
  import { xml } from "@codemirror/lang-xml";
  import {
    defaultHighlightStyle,
    foldGutter,
    foldKeymap,
    HighlightStyle,
    StreamLanguage,
    syntaxHighlighting
  } from "@codemirror/language";
  import { lua } from "@codemirror/legacy-modes/mode/lua";
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
  import CheckOutline from "flowbite-svelte-icons/CheckOutline.svelte";
  import ClipboardCleanSolid from "flowbite-svelte-icons/ClipboardCleanSolid.svelte";
  import Clipboard from "flowbite-svelte/Clipboard.svelte";
  import { onDestroy, onMount } from "svelte";

  interface Props {
    value: string;
    format?: "none" | "json" | "xml" | "text";
    language?: "json" | "xml" | "lua" | "text" | "none" | "";
    environmentEntries?: { key: string; value: string }[];
    readOnly?: boolean;
    size?: "sm" | "md" | "lg";
    onChange?: (value: string) => void;
    showCopyPaste?: boolean;
  }

  let {
    value = $bindable(""),
    format = $bindable("json"),
    language = "",
    environmentEntries = [],
    readOnly = false,
    size = "md",
    onChange,
    showCopyPaste = false
  }: Props = $props();

  let success = $state(false);

  // Annotation to mark transactions initiated by us (external sync) so the
  // updateListener does NOT re-emit "change" and cause an infinite loop.
  const externalUpdate = Annotation.define<boolean>();

  let editorEl: HTMLDivElement | undefined = $state();
  let view: EditorView | undefined = $state();

  let languageCompartment = new Compartment();
  let tokenDecoratorCompartment = new Compartment();

  let autocompleteOpen = $state(false);
  let autocompleteActiveIndex = $state(0);
  let autocompleteEntries: { key: string; value: string }[] = $state([]);
  let autocompleteLeft = $state(8);
  let autocompleteTop = $state(8);
  let autocompleteMaxWidth = $state(320);

  let autocompletePopoverStyle = $derived(
    `left: ${autocompleteLeft}px; top: ${autocompleteTop}px; min-width: 220px; max-width: ${autocompleteMaxWidth}px;`
  );

  let sizeClass = $derived(getTokenizedEditorSizeClass(size));
  let knownEnvironmentKeys = $derived(new Set(environmentEntries.map((entry) => entry.key)));
  let sessionKeys = $derived(new Set(Object.keys($sessionVarsStore ?? {})));
  let knownKeysSignature = $derived(
    `${environmentEntries.map((entry) => entry.key).join("\u0000")}|${Object.keys($sessionVarsStore ?? {}).join("\u0000")}`
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
    autocompleteEntries = filterEnvTokenEntries(environmentEntries, normalizedQuery);
    autocompleteActiveIndex = clampActiveIndex(autocompleteActiveIndex, autocompleteEntries.length);
    autocompleteOpen = true;
    updateAutocompletePosition(triggerContext.to, autocompleteEntries.length);
  }

  function applyAutocompleteEntry(entry: { key: string; value: string }) {
    if (!view) return;

    const triggerContext = findEnvTokenTriggerContext(
      view.state.doc.toString(),
      view.state.selection.main.head
    );
    if (!triggerContext) return;

    const inserted = createEnvTokenSnippet(entry.key);
    view.dispatch({
      changes: { from: triggerContext.from, to: triggerContext.to, insert: inserted },
      selection: EditorSelection.cursor(triggerContext.from + inserted.length)
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
      keymap.of([...defaultKeymap, ...foldKeymap]),
      languageCompartment.of(getLangExtension()),
      syntaxHighlighting(customHighlightStyle),
      syntaxHighlighting(defaultHighlightStyle, { fallback: true }),
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
        : [EditorState.readOnly.of(true), EditorView.editable.of(false)])
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
    if (lang === "json") return json();
    if (lang === "xml") return xml();
    if (lang === "lua") return StreamLanguage.define(lua);
    return [];
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
        sessionKeys.has(tokenKey)
          ? "session"
          : knownEnvironmentKeys.has(tokenKey)
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
</script>

<div class="relative h-full {sizeClass}">
  <div class="editor-container h-full" class:read-only={readOnly} bind:this={editorEl}></div>

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
      class="absolute inset-e-2 top-2 h-8 px-2.5 font-medium focus:ring-0"
    >
      {#if success}
        <CheckOutline class="h-3 w-3" /> Copied
      {:else}
        <ClipboardCleanSolid class="h-3 w-3" /> Copy code
      {/if}
    </Clipboard>
  {/if}
</div>
