<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: GPL-3.0-only
-->

<script lang="ts">
  import {
    createEnvTokenDecorationPlugin,
    createEnvTokenSnippet,
    filterEnvTokenEntries,
    findEnvTokenTriggerContext
  } from "$src/lib/utils/tokens";
  import { hideTokenTooltipDelay, showTokenTooltip } from "$src/lib/stores/tokenTooltipStore";
  import {
    autocompletion,
    completionKeymap,
    type CompletionContext,
    type CompletionResult
  } from "@codemirror/autocomplete";
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
  import { Annotation, Compartment, EditorState } from "@codemirror/state";
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
    onChange?: (value: string) => void;
    showCopyPaste?: boolean;
  }

  let {
    value = $bindable(""),
    format = $bindable("json"),
    language = "",
    environmentEntries = [],
    readOnly = false,
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
      appTheme,
      // --- edit-only extensions ---
      ...(!readOnly
        ? [
            history(),
            keymap.of([...historyKeymap, ...completionKeymap]),
            autocompletion({ override: [envCompletionSource] }),
            tokenHighlightPlugin,
            EditorView.updateListener.of((update) => {
              if (
                update.docChanged &&
                !update.transactions.some((tr) => tr.annotation(externalUpdate))
              ) {
                onChange?.(update.state.doc.toString());
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

  // Token highlighter — only used in edit mode
  const tokenHighlightPlugin = createEnvTokenDecorationPlugin({
    tokenClassName: "cm-solo-token",
    onTokenMouseOver: (tokenKey, rect) => showTokenTooltip(tokenKey, rect.left, rect.bottom),
    onTokenMouseOut: () => hideTokenTooltipDelay()
  });

  // Autocomplete for {{...}} — only used in edit mode
  function envCompletionSource(context: CompletionContext): CompletionResult | null {
    const triggerContext = findEnvTokenTriggerContext(context.state.doc.toString(), context.pos);
    if (!triggerContext) return null;

    const filteredEntries = filterEnvTokenEntries(
      environmentEntries,
      triggerContext.normalizedQuery
    );

    return {
      // Use query span (after "{{") as completion range so CodeMirror's
      // internal filtering doesn't see the opening braces as prefix text.
      from: triggerContext.from + 2,
      to: triggerContext.to,
      options: filteredEntries.map((entry) => ({
        label: entry.key,
        type: "variable",
        apply: (view) => {
          const inserted = createEnvTokenSnippet(entry.key);
          view.dispatch({
            changes: {
              from: triggerContext.from,
              to: triggerContext.to,
              insert: inserted
            }
          });
        }
      }))
    };
  }

  const appTheme = EditorView.theme({
    "&": {
      height: "100%",
      fontSize: "0.875rem"
    },
    ".cm-foldGutter .cm-gutterElement": {
      cursor: "pointer"
    },
    ".cm-solo-token": {
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
</script>

<div class="relative h-full">
  <div class="editor-container h-full" class:read-only={readOnly} bind:this={editorEl}></div>
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
