<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: GPL-3.0-only
-->

<script lang="ts">
  import EnvAutocompletePopover from "$src/lib/components/RequestBuilder/EnvAutocompletePopover.svelte";
  import {
    forceHideTokenTooltip,
    hideTokenTooltipDelay,
    showTokenTooltip
  } from "$src/lib/stores/tokenTooltipStore";
  import { sessionVarsStore } from "$src/lib/stores/sessionVarsStore";
  import { environmentStoreState } from "$src/lib/stores/environmentStore.svelte";
  import {
    clampActiveIndex,
    createEnvTokenDecorationPlugin,
    createEnvTokenSnippet,
    createTokenizedEditorTheme,
    filterEnvTokenEntries,
    findEnvTokenTriggerContext,
    getTokenizedEditorSizeClass,
    normalizeEnvironmentTokenEntries
  } from "$src/lib/utils/tokens";
  import { defaultKeymap, history, historyKeymap } from "@codemirror/commands";
  import { Annotation, Compartment, EditorSelection, EditorState } from "@codemirror/state";
  import { EditorView, keymap } from "@codemirror/view";
  import { onDestroy, onMount, type Snippet } from "svelte";

  interface Props {
    value?: string;
    placeholder?: string;
    disabled?: boolean;
    size?: "sm" | "md" | "lg";
    class?: string;
    right?: Snippet;
    rightVisible?: boolean;
    onChange?: () => void;
    onEnter?: () => void;
  }

  type EnvEntry = { key: string; value: string; type?: string };

  let {
    value = $bindable(""),
    placeholder = "",
    disabled = false,
    size = "md",
    class: className = "",
    right,
    rightVisible = true,
    onChange,
    onEnter
  }: Props = $props();

  let editorElement: HTMLDivElement | undefined = $state();
  let view: EditorView | undefined = $state();

  const externalUpdate = Annotation.define<boolean>();
  const editableCompartment = new Compartment();
  const tokenDecoratorCompartment = new Compartment();

  let autocompleteOpen = $state(false);
  let autocompleteQuery = $state("");
  let autocompleteActiveIndex = $state(0);
  let autocompleteEntries: EnvEntry[] = $state([]);
  let shouldUseMacOsLayoutNudge = $state(false);
  let editorReady = $state(true);

  function sanitizeSingleLine(input: string): string {
    return (input ?? "").replace(/[\r\n]+/g, " ");
  }

  let selectedEnvironment = $derived(
    environmentStoreState.environments.find(
      (e) => e.name === environmentStoreState.selectedEnvironmentName
    ) || null
  );

  let environmentEntries = $derived(normalizeEnvironmentTokenEntries(selectedEnvironment?.values));
  let knownEnvironmentKeys = $derived(new Set(environmentEntries.map((entry) => entry.key)));
  let sessionKeys = $derived(new Set(Object.keys($sessionVarsStore ?? {})));
  let knownKeysSignature = $derived(
    `${environmentEntries.map((entry) => entry.key).join("\u0000")}|${Object.keys($sessionVarsStore ?? {}).join("\u0000")}`
  );

  let sizeClass = $derived(
    `${getTokenizedEditorSizeClass(size)} ${size === "sm" ? "px-2 py-1" : size === "lg" ? "px-3 py-3" : "px-2.5 py-2.5"}`
  );

  let shellClass = $derived(
    `relative flex w-full items-center rounded-lg border border-gray-300 bg-gray-50 text-gray-900 dark:border-gray-600 dark:bg-gray-700 dark:text-white ${disabled ? "cursor-not-allowed opacity-50" : "focus-within:z-10 focus-within:border-primary-500 focus-within:ring-1 focus-within:ring-primary-500"} ${sizeClass}`
  );

  function createTokenDecorator(_knownKeysSignature?: string) {
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

  function refreshAutocomplete(state: EditorState) {
    if (state.facet(EditorState.readOnly)) {
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
      autocompleteQuery = "";
      autocompleteActiveIndex = 0;
      return;
    }

    autocompleteQuery = triggerContext.normalizedQuery;
    autocompleteEntries = filterEnvTokenEntries(environmentEntries, autocompleteQuery);
    autocompleteActiveIndex = clampActiveIndex(autocompleteActiveIndex, autocompleteEntries.length);
    autocompleteOpen = true;
    forceHideTokenTooltip();
  }

  function applyAutocompleteEntry(entry: EnvEntry) {
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

  const navigationKeymap = keymap.of([
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
  ]);

  const singleLineFilter = EditorState.transactionFilter.of((tr) => {
    if (tr.newDoc.lines > 1) {
      return [];
    }
    return tr;
  });

  const triggerSendKeymap = keymap.of([
    {
      key: "Mod-Enter",
      run: () => {
        onEnter?.();
        return true;
      }
    }
  ]);

  function isMacOs(): boolean {
    const userAgentDataPlatform =
      typeof navigator !== "undefined" && "userAgentData" in navigator
        ? (navigator as Navigator & { userAgentData?: { platform?: string } }).userAgentData
            ?.platform
        : "";

    const platform = navigator.platform || "";
    const userAgent = navigator.userAgent || "";
    const source = `${userAgentDataPlatform} ${platform} ${userAgent}`.toLowerCase();
    return source.includes("mac");
  }

  function forceHorizontalNudgeRelayout() {
    if (!view || !editorElement) {
      editorReady = true;
      return;
    }

    const previousWidth = editorElement.style.width;

    // In WKWebView, this bug is often fixed only after a horizontal resize.
    // Nudge width by a tiny amount and restore it on the next frame.
    editorElement.style.width = "calc(100% - 0.01px)";
    void editorElement.getBoundingClientRect();

    requestAnimationFrame(() => {
      if (!view || !editorElement) {
        editorReady = true;
        return;
      }
      editorElement.style.width = previousWidth || "";
      void editorElement.getBoundingClientRect();
      view.requestMeasure();
      refreshAutocomplete(view.state);
      editorReady = true;
    });
  }

  onMount(() => {
    shouldUseMacOsLayoutNudge = isMacOs();
    editorReady = !shouldUseMacOsLayoutNudge;

    const initialValue = sanitizeSingleLine(value);
    if (value !== initialValue) {
      value = initialValue;
    }

    const extensions = [
      singleLineFilter,
      triggerSendKeymap,
      navigationKeymap,
      keymap.of([...historyKeymap, ...defaultKeymap]),
      history(),
      tokenDecoratorCompartment.of(createTokenDecorator()),
      createTokenizedEditorTheme({ singleLine: true }),
      EditorView.updateListener.of((update) => {
        if (
          update.docChanged &&
          !update.transactions.some((transaction) => transaction.annotation(externalUpdate))
        ) {
          const current = sanitizeSingleLine(update.state.doc.toString());
          if (value !== current) {
            value = current;
            onChange?.();
          }
        }

        if (update.docChanged || update.selectionSet) {
          refreshAutocomplete(update.state);
        }
      }),
      editableCompartment.of(
        disabled
          ? [EditorState.readOnly.of(true), EditorView.editable.of(false)]
          : [EditorState.readOnly.of(false), EditorView.editable.of(true)]
      )
    ];

    const state = EditorState.create({
      doc: initialValue,
      extensions
    });

    view = new EditorView({
      state,
      parent: editorElement
    });

    refreshAutocomplete(state);

    let frame1 = 0;
    let frame2 = 0;

    if (shouldUseMacOsLayoutNudge) {
      frame1 = requestAnimationFrame(() => {
        frame2 = requestAnimationFrame(() => {
          forceHorizontalNudgeRelayout();
        });
      });
    }

    return () => {
      cancelAnimationFrame(frame1);
      cancelAnimationFrame(frame2);
      view?.destroy();
    };
  });

  onDestroy(() => {
    view?.destroy();
  });

  $effect(() => {
    if (!view) return;

    const incoming = sanitizeSingleLine(value);
    const current = view.state.doc.toString();

    if (incoming !== current) {
      view.dispatch({
        changes: { from: 0, to: current.length, insert: incoming },
        annotations: [externalUpdate.of(true)]
      });
    }
  });

  $effect(() => {
    if (!view) return;

    view.dispatch({
      effects: editableCompartment.reconfigure(
        disabled
          ? [EditorState.readOnly.of(true), EditorView.editable.of(false)]
          : [EditorState.readOnly.of(false), EditorView.editable.of(true)]
      )
    });

    refreshAutocomplete(view.state);
  });

  $effect(() => {
    if (!view) return;

    view.dispatch({
      effects: tokenDecoratorCompartment.reconfigure(createTokenDecorator(knownKeysSignature))
    });
  });
</script>

<div class="{shellClass} {className}" data-size={size}>
  <div class="relative min-w-0 flex-1">
    {#if shouldUseMacOsLayoutNudge && !editorReady}
      <span class="invisible block select-none">M</span>
    {/if}

    <div
      bind:this={editorElement}
      class="env-token-input-editor w-full {shouldUseMacOsLayoutNudge && !editorReady
        ? 'pointer-events-none absolute inset-0'
        : ''}"
      style={`visibility: ${shouldUseMacOsLayoutNudge && !editorReady ? "hidden" : "visible"}`}
    ></div>

    {#if !value && placeholder}
      <div
        class="pointer-events-none absolute inset-0 flex items-center text-neutral-400 dark:text-neutral-400"
      >
        <span class="truncate">{placeholder}</span>
      </div>
    {/if}
  </div>

  {#if right && rightVisible}
    <div class="ml-1.5 shrink-0">{@render right()}</div>
  {/if}

  <EnvAutocompletePopover
    open={autocompleteOpen}
    entries={autocompleteEntries}
    activeIndex={autocompleteActiveIndex}
    class="env-token-autocomplete absolute top-full right-0 left-0 z-90 mt-1"
    onHoverIndex={(index) => (autocompleteActiveIndex = index)}
    onSelect={(entry) => applyAutocompleteEntry(entry)}
    onRequestClose={() => (autocompleteOpen = false)}
  />
</div>

<style>
  :global(.env-token-input-editor .cm-editor) {
    color: inherit;
    background: transparent !important;
    font-family: var(--font-mono);
    min-height: 100%;
    height: 100%;
  }

  :global(.env-token-input-editor .cm-scroller) {
    min-height: 100%;
    display: flex;
    align-items: center;
    width: 100%;
    background: transparent !important;
  }

  :global(.env-token-input-editor .cm-content) {
    font-family: inherit;
    line-height: inherit;
    width: 100%;
    padding: 0;
    background: transparent !important;
    min-height: 1em;
    display: flex;
    align-items: center;
    min-height: 100%;
  }

  :global(.env-token-input-editor .cm-line) {
    line-height: inherit;
    padding: 0;
    margin: 0;
    width: 100%;
    background: transparent !important;
  }
</style>
