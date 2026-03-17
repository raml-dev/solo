<script lang="ts">
  import CheckOutline from "flowbite-svelte-icons/CheckOutline.svelte";
  import ClipboardCleanSolid from "flowbite-svelte-icons/ClipboardCleanSolid.svelte";
  import Clipboard from "flowbite-svelte/Clipboard.svelte";
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
    Decoration,
    EditorView,
    highlightActiveLine,
    highlightActiveLineGutter,
    keymap,
    lineNumbers,
    ViewPlugin,
    type DecorationSet
  } from "@codemirror/view";
  import { tags } from "@lezer/highlight";
  import { indentationMarkers } from "@replit/codemirror-indentation-markers";
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
  const tokenHighlightPlugin = ViewPlugin.fromClass(
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
        const re = /\{\{([^{}\r\n]+?)\}\}/g;
        for (const { from, to } of v.visibleRanges) {
          const text = v.state.doc.sliceString(from, to);
          for (const match of text.matchAll(re)) {
            const start = from + match.index!;
            const end = start + match[0].length;
            builder.push(
              Decoration.mark({
                class: "cm-yapla-token",
                attributes: { "data-token-key": match[1].trim() }
              }).range(start, end)
            );
          }
        }
        return Decoration.set(builder);
      }
    },
    {
      decorations: (v) => v.decorations,
      eventHandlers: {
        mouseover: (e) => {
          const target = e.target as HTMLElement;
          if (target.classList.contains("cm-yapla-token")) {
            const tokenKey = target.dataset.tokenKey;
            if (tokenKey) {
              const rect = target.getBoundingClientRect();
              showTokenTooltip(tokenKey, rect.left, rect.bottom);
            }
          }
        },
        mouseout: (e) => {
          const target = e.target as HTMLElement;
          if (target.classList.contains("cm-yapla-token")) hideTokenTooltipDelay();
        },
        mouseleave: () => hideTokenTooltipDelay()
      }
    }
  );

  // Autocomplete for {{...}} — only used in edit mode
  function envCompletionSource(context: CompletionContext): CompletionResult | null {
    const node = context.matchBefore(/\{\{([\w-]*)/);
    if (!node) return null;
    return {
      from: node.from + 2,
      options: environmentEntries.map((e) => ({
        label: e.key,
        type: "variable",
        apply: (v, completion) => {
          v.dispatch({
            changes: { from: node!.from, to: context.pos, insert: `{{${completion.label}}}` }
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
    ".cm-yapla-token": {
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

<div class="relative">
  <div class="editor-container h-full" class:read-only={readOnly} bind:this={editorEl}></div>
  {#if showCopyPaste}
    <Clipboard
      color={success ? "alternative" : "light"}
      bind:success
      {value}
      size="sm"
      class="absolute end-2 top-2 h-8 px-2.5 font-medium focus:ring-0"
    >
      {#if success}
        <CheckOutline class="h-3 w-3" /> Copied
      {:else}
        <ClipboardCleanSolid class="h-3 w-3" /> Copy code
      {/if}
    </Clipboard>
  {/if}
</div>
