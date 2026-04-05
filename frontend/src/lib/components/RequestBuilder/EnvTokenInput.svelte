<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: GPL-3.0-only
-->

<script lang="ts">
  import Button from "flowbite-svelte/Button.svelte";
  import Card from "flowbite-svelte/Card.svelte";
  import {
    forceHideTokenTooltip,
    hideTokenTooltipDelay,
    showTokenTooltip
  } from "$src/lib/stores/tokenTooltipStore";
  import { environmentStoreState } from "$src/lib/stores/environmentStore.svelte";
  import {
    clampActiveIndex,
    createEnvTokenDecorationPlugin,
    createEnvTokenSnippet,
    filterEnvTokenEntries,
    findEnvTokenTriggerContext,
    normalizeEnvironmentTokenEntries
  } from "$src/lib/utils/tokens";
  import { defaultKeymap, history, historyKeymap } from "@codemirror/commands";
  import { Annotation, Compartment, EditorSelection, EditorState } from "@codemirror/state";
  import { EditorView, keymap } from "@codemirror/view";
  import { onDestroy, onMount } from "svelte";

  interface Props {
    value?: string;
    placeholder?: string;
    disabled?: boolean;
    size?: "sm" | "md" | "lg";
    class?: string;
    onChange?: () => void;
    onEnter?: () => void;
  }

  type EnvEntry = { key: string; value: string; type: string };

  let {
    value = $bindable(""),
    placeholder = "",
    disabled = false,
    size = "md",
    class: className = "",
    onChange,
    onEnter
  }: Props = $props();

  let editorElement: HTMLDivElement | undefined = $state();
  let view: EditorView | undefined = $state();

  const externalUpdate = Annotation.define<boolean>();
  const editableCompartment = new Compartment();

  let autocompleteOpen = $state(false);
  let autocompleteQuery = $state("");
  let autocompleteActiveIndex = $state(0);
  let autocompleteEntries: EnvEntry[] = $state([]);

  function sanitizeSingleLine(input: string): string {
    return (input ?? "").replace(/[\r\n]+/g, " ");
  }

  let selectedEnvironment = $derived(
    environmentStoreState.environments.find(
      (e) => e.name === environmentStoreState.selectedEnvironmentName
    ) || null
  );

  let environmentEntries = $derived(normalizeEnvironmentTokenEntries(selectedEnvironment?.values));

  let sizeClass = $derived(
    size === "sm"
      ? "text-xs leading-4 px-2 py-1"
      : size === "lg"
        ? "sm:text-base leading-6 px-3 py-3"
        : "text-sm leading-5 px-2.5 py-2.5"
  );

  let shellClass = $derived(
    `relative flex w-full items-center rounded-lg border border-gray-300 bg-gray-50 text-gray-900 dark:border-gray-600 dark:bg-gray-700 dark:text-white ${disabled ? "cursor-not-allowed opacity-50" : "focus-within:z-10 focus-within:border-primary-500 focus-within:ring-1 focus-within:ring-primary-500"} ${sizeClass}`
  );

  const tokenDecorator = createEnvTokenDecorationPlugin({
    tokenClassName: "cm-env-token",
    onTokenMouseOver: (tokenKey, rect) => showTokenTooltip(tokenKey, rect.left, rect.bottom),
    onTokenMouseOut: () => hideTokenTooltipDelay()
  });

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

  onMount(() => {
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
      tokenDecorator,
      EditorView.theme({
        "&": {
          background: "transparent",
          font: "inherit"
        },
        ".cm-scroller": {
          font: "inherit",
          lineHeight: "inherit",
          overflowX: "auto",
          overflowY: "hidden"
        },
        ".cm-content": {
          padding: "0",
          fontFamily: "inherit"
        },
        ".cm-line": {
          padding: "0"
        },
        "&.cm-focused": {
          outline: "none"
        }
      }),
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

    return () => {
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
</script>

<div class="{shellClass} {className}" data-size={size}>
  <div class="relative h-full w-full">
    <div bind:this={editorElement} class="env-token-input-editor h-full w-full"></div>

    {#if !value && placeholder}
      <div
        class="pointer-events-none absolute inset-0 flex items-center text-neutral-400 dark:text-neutral-400"
      >
        <span class="truncate">{placeholder}</span>
      </div>
    {/if}
  </div>

  {#if autocompleteOpen}
    <div class="env-token-autocomplete absolute top-full right-0 left-0 z-90 mt-1">
      <Card class="w-full p-1 shadow-lg">
        {#if autocompleteEntries.length > 0}
          <div role="listbox" class="max-h-56 space-y-1 overflow-auto">
            {#each autocompleteEntries as entry, index (entry.key)}
              <Button
                role="option"
                aria-selected={index === autocompleteActiveIndex}
                color={index === autocompleteActiveIndex ? "primary" : "light"}
                size="xs"
                class="flex w-full items-center justify-between gap-2"
                onmouseenter={() => (autocompleteActiveIndex = index)}
                onmousedown={(event: MouseEvent) => {
                  event.preventDefault();
                  applyAutocompleteEntry(entry);
                }}
              >
                <span class="truncate font-mono">{entry.key}</span>
                <span class="max-w-36 truncate text-xs opacity-80">{entry.value}</span>
              </Button>
            {/each}
          </div>
        {:else}
          <div class="px-2 py-1 text-xs text-neutral-500 dark:text-neutral-400">
            No environment variables
          </div>
        {/if}
      </Card>
    </div>
  {/if}
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
    caret-color: var(--color-neutral-900);
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

  :global(.dark .env-token-input-editor .cm-content) {
    caret-color: var(--color-neutral-100);
  }

  :global(.env-token-input-editor .cm-env-token) {
    cursor: pointer;
    border-radius: 0.25rem;
    background: var(--color-primary-100);
    color: var(--color-primary-700);
  }

  :global(.dark .env-token-input-editor .cm-env-token) {
    background: var(--color-primary-900);
    color: var(--color-primary-300);
  }
</style>
