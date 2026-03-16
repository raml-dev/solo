<script lang="ts">
  import Dropzone from "flowbite-svelte/Dropzone.svelte";
  import { OnFileDrop, OnFileDropOff } from "$wails/runtime/runtime";
  import { onDestroy, onMount } from "svelte";
  import type { Snippet } from "svelte";

  interface Props {
    title?: string;
    subtitle?: string;
    icon?: Snippet;
    onDrop?: (e: { paths: string[] }) => void;
  }

  let { title = "Drop file here", subtitle = "", icon, onDrop }: Props = $props();

  let dragOver = $state(false);
  let wrapperEl: HTMLDivElement | undefined = $state();

  function handleWailsDrop(x: number, y: number, paths: string[]) {
    if (!wrapperEl) return;

    const rect = wrapperEl.getBoundingClientRect();
    if (x >= rect.left && x <= rect.right && y >= rect.top && y <= rect.bottom) {
      dragOver = false;
      onDrop?.({ paths });
    }
  }

  function onDragEnter(event: DragEvent) {
    event.preventDefault();
    dragOver = true;
  }

  function onDragLeave(event: DragEvent) {
    event.preventDefault();
    dragOver = false;
  }

  function onDragOver(event: DragEvent) {
    event.preventDefault();
    dragOver = true;
  }

  function onDropEvent(event: DragEvent) {
    event.preventDefault();
    dragOver = false;
  }

  function onClick(event: MouseEvent) {
    // Route click to consumer fallback picker logic (same path as "Select file/folder").
    event.preventDefault();
    event.stopPropagation();
    onDrop?.({ paths: [] });
  }

  function onKeyDown(event: KeyboardEvent) {
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      onDrop?.({ paths: [] });
    }
  }

  onMount(() => {
    try {
      OnFileDrop(handleWailsDrop, true);
    } catch {
      // Running outside Wails runtime
    }
  });

  onDestroy(() => {
    try {
      OnFileDropOff();
    } catch {
      // Running outside Wails runtime
    }
  });
</script>

<div
  bind:this={wrapperEl}
  style="--wails-drop-target: drop"
  role="button"
  tabindex="0"
  ondragenter={onDragEnter}
  ondragleave={onDragLeave}
  ondragover={onDragOver}
  ondrop={onDropEvent}
  onclick={onClick}
  onkeydown={onKeyDown}
>
  <Dropzone
    class={(dragOver
      ? "border-primary-500 bg-primary-50 dark:border-primary-400 dark:bg-primary-900/20 "
      : "") + "pointer-events-none"}
  >
    <div class="flex flex-col items-center justify-center py-4 text-center">
      <div class="mb-3 text-gray-500 dark:text-gray-400">
        {#if icon}
          {@render icon()}
        {:else}
          <svg
            width="44"
            height="44"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="1.4"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
            <polyline points="17 8 12 3 7 8" />
            <line x1="12" y1="3" x2="12" y2="15" />
          </svg>
        {/if}
      </div>

      <p class="mb-1 text-base font-medium text-gray-700 dark:text-gray-200">{title}</p>
      {#if subtitle}
        <p class="text-sm text-gray-500 dark:text-gray-400">{subtitle}</p>
      {/if}
    </div>
  </Dropzone>
</div>
