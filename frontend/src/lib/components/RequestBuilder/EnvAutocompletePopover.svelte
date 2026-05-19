<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import type { ResolvedVariableEntry, VariableSource } from "$src/lib/utils/variableResolution";
  import { formatVariableSourceLabel } from "$src/lib/utils/variableResolution";
  import ExclamationCircleSolid from "flowbite-svelte-icons/ExclamationCircleSolid.svelte";
  import Badge from "flowbite-svelte/Badge.svelte";
  import Card from "flowbite-svelte/Card.svelte";
  import { onDestroy, onMount, tick } from "svelte";

  interface Props {
    open?: boolean;
    entries?: ResolvedVariableEntry[];
    activeIndex?: number;
    class?: string;
    style?: string;
    emptyLabel?: string;
    onSelect?: (entry: ResolvedVariableEntry, index: number) => void;
    onHoverIndex?: (index: number) => void;
    onRequestClose?: () => void;
  }

  let {
    open = false,
    entries = [],
    activeIndex = 0,
    class: className = "",
    style = "",
    emptyLabel = "No environment variables",
    onSelect,
    onHoverIndex,
    onRequestClose
  }: Props = $props();

  let rootEl: HTMLDivElement | undefined = $state();
  let listEl: HTMLDivElement | undefined = $state();

  let lastPointerX = Number.NaN;
  let lastPointerY = Number.NaN;

  function getSourceBadgeColor(source: VariableSource): "blue" | "gray" | "yellow" {
    return source === "session" ? "gray" : source === "environment" ? "blue" : "yellow";
  }

  function getConflictMessage(entry: ResolvedVariableEntry): string {
    if (!entry.hasConflicts) {
      return "";
    }

    const sources = entry.definedIn
      .filter((source) => source != entry.winningSource)
      .map((source) => formatVariableSourceLabel(source))
      .join(", ");
    return `Also defined in ${sources}`;
  }

  function isRealPointerMove(event: PointerEvent): boolean {
    const moved = event.clientX !== lastPointerX || event.clientY !== lastPointerY;
    lastPointerX = event.clientX;
    lastPointerY = event.clientY;
    return moved;
  }

  function handleListPointerMove(event: PointerEvent) {
    if (!isRealPointerMove(event)) return;

    const target = event.target as HTMLElement;
    const row = target.closest("[data-env-index]") as HTMLElement | null;
    if (!row) return;

    const nextIndex = Number(row.dataset.envIndex);
    if (Number.isNaN(nextIndex)) return;

    onHoverIndex?.(nextIndex);
  }

  function handleDocumentMouseDown(event: MouseEvent) {
    if (!open || !rootEl) return;
    const target = event.target as Node;
    if (rootEl.contains(target)) return;
    onRequestClose?.();
  }

  onMount(() => {
    document.addEventListener("mousedown", handleDocumentMouseDown);
  });

  onDestroy(() => {
    document.removeEventListener("mousedown", handleDocumentMouseDown);
  });

  $effect(() => {
    if (!open || !listEl || entries.length === 0) return;

    const index = activeIndex;

    tick().then(() => {
      if (!open || !listEl) return;

      const selected = listEl.querySelector(`[data-env-index="${index}"]`) as HTMLElement | null;
      if (!selected) return;

      const listRect = listEl.getBoundingClientRect();
      const itemRect = selected.getBoundingClientRect();
      const margin = 4;

      if (itemRect.top < listRect.top) {
        listEl.scrollTop -= listRect.top - itemRect.top + margin;
      } else if (itemRect.bottom > listRect.bottom) {
        listEl.scrollTop += itemRect.bottom - listRect.bottom + margin;
      }
    });
  });
</script>

{#if open}
  <div bind:this={rootEl} class={className} {style}>
    <Card class="w-full p-1 shadow-lg">
      {#if entries.length > 0}
        <div
          bind:this={listEl}
          tabindex={0}
          role="listbox"
          class="max-h-56 space-y-1 overflow-auto"
          onpointermove={handleListPointerMove}
        >
          {#each entries as entry, index (entry.key)}
            <!-- NOTE: the following element was converted to a plain <button> from a flowbite's <Button>
             because Button, when rendered as a child of a ButtonGroup (such as when triggered from HTTPRequestBuilder's EnvTokenInput),
             inherits a particular svelte context, which triggers some automatic class-manipulating logic from flowbite
             to properly not-render borders, which is right if the Button is actually inside a ButtonGroup,
             but not in this situation: this is an _indirect_ child of the ButtonGroup, but it still inherits that context.
            -->
            <button
              type="button"
              role="option"
              aria-selected={index === activeIndex}
              data-env-index={index}
              class={`flex w-full flex-col items-start justify-between gap-1 rounded-lg px-3 py-2 text-left text-xs transition-colors ${
                index === activeIndex
                  ? "bg-primary-700 text-white hover:bg-primary-800 dark:bg-primary-600 dark:hover:bg-primary-700"
                  : "border border-gray-300 bg-white text-gray-900 hover:bg-gray-100 dark:border-gray-600 dark:bg-gray-800 dark:text-white dark:hover:bg-gray-700"
              }`}
              onmousedown={(event: MouseEvent) => {
                event.preventDefault();
                onSelect?.(entry, index);
              }}
            >
              <span class="flex w-full flex-row justify-between gap-2">
                <Badge class="min-w-25" color={getSourceBadgeColor(entry.winningSource)}>
                  {formatVariableSourceLabel(entry.winningSource)}
                </Badge>
                <span class="min-w-0 flex-1 shrink-0 items-center">
                  <span class="block truncate font-mono">{entry.key}</span>
                </span>
                <span class="flex shrink-0 items-center gap-2">
                  <span class="max-w-36 truncate text-xs">{entry.computedValue}</span>
                </span>
              </span>
              {#if entry.hasConflicts}
                <span class="flex gap-1 truncate text-xs opacity-80"
                  ><ExclamationCircleSolid class="h-4 w-4 shrink-0 fill-warning-500" /><span
                    >{getConflictMessage(entry)}</span
                  ></span
                >
              {/if}
            </button>
          {/each}
        </div>
      {:else}
        <div class="px-2 py-1 text-xs text-neutral-500 dark:text-neutral-400">{emptyLabel}</div>
      {/if}
    </Card>
  </div>
{/if}
