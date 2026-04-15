<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import Button from "flowbite-svelte/Button.svelte";
  import Card from "flowbite-svelte/Card.svelte";
  import { onDestroy, onMount, tick } from "svelte";

  export interface EnvAutocompletePopoverEntry {
    key: string;
    value: string;
    type?: string;
  }

  interface Props {
    open?: boolean;
    entries?: EnvAutocompletePopoverEntry[];
    activeIndex?: number;
    class?: string;
    style?: string;
    emptyLabel?: string;
    onSelect?: (entry: EnvAutocompletePopoverEntry, index: number) => void;
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
            <Button
              role="option"
              aria-selected={index === activeIndex}
              data-env-index={index}
              color={index === activeIndex ? "primary" : "light"}
              size="xs"
              class="flex w-full items-center justify-between gap-2"
              onmousedown={(event: MouseEvent) => {
                event.preventDefault();
                onSelect?.(entry, index);
              }}
            >
              <span class="truncate font-mono">{entry.key}</span>
              <span class="max-w-36 truncate text-xs opacity-80">{entry.value}</span>
            </Button>
          {/each}
        </div>
      {:else}
        <div class="px-2 py-1 text-xs text-neutral-500 dark:text-neutral-400">{emptyLabel}</div>
      {/if}
    </Card>
  </div>
{/if}
