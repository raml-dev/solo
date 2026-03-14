<script lang="ts">
  import { onMount, onDestroy, createEventDispatcher } from "svelte";
  import { OnFileDrop, OnFileDropOff } from "../../../../wailsjs/runtime/runtime";

  interface Props {
    /** Titolo grande al centro della zona */
    title?: string;
    /** Sottotitolo descrittivo */
    subtitle?: string;
    icon?: import("svelte").Snippet;
  }

  let { title = "Drop file here", subtitle = "", icon }: Props = $props();

  const dispatch = createEventDispatcher<{ drop: { paths: string[] } }>();

  let dragOver = $state(false);
  let el: HTMLDivElement | undefined = $state();

  // Wails OnFileDrop fornisce i path reali del filesystem.
  // useDropTarget=true limita il callback agli elementi CSS marcati come drop target.
  onMount(() => {
    try {
      OnFileDrop((x, y, paths) => {
        // Verifica che il drop sia avvenuto sopra questo elemento specifico
        if (!el) return;
        const rect = el.getBoundingClientRect();
        if (x >= rect.left && x <= rect.right && y >= rect.top && y <= rect.bottom) {
          dragOver = false;
          dispatch("drop", { paths });
        }
      }, true);
    } catch {
      // Fuori dal runtime Wails (es. dev server browser): no-op
    }
  });

  onDestroy(() => {
    try {
      OnFileDropOff();
    } catch {
      // no-op
    }
  });

  // Visual feedback durante il drag — usiamo gli eventi del webview
  // solo per l'highlight visivo (il drop effettivo viene da Wails)
  function onDragEnter() {
    dragOver = true;
  }
  function onDragLeave() {
    dragOver = false;
  }
  // Preveniamo il comportamento default del webview (apertura file)
  function onDragOver(e: DragEvent) {
    e.preventDefault();
  }
  function onDrop(e: DragEvent) {
    e.preventDefault();
  }
</script>

<!--
  --wails-drop-target: drop  →  Wails riconosce questo elemento come drop target
  e consegna i path reali al callback OnFileDrop registrato in onMount.
-->
<div
  bind:this={el}
  class="dropzone"
  class:drag-over={dragOver}
  role="button"
  tabindex="0"
  style="--wails-drop-target: drop"
  ondragenter={onDragEnter}
  ondragleave={onDragLeave}
  ondragover={onDragOver}
  ondrop={onDrop}
  onkeydown={(e) => e.key === "Enter" && dispatch("drop", { paths: [] })}
>
  <div class="dropzone-icon">
    {#if icon}{@render icon()}{:else}
      <!-- icona upload di default -->
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

  <p class="dropzone-title">{title}</p>
  {#if subtitle}
    <p class="dropzone-subtitle">{subtitle}</p>
  {/if}
</div>

<style>
  .dropzone {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--space-md);
    min-height: 240px;
    margin: var(--space-lg);
    border: 2px dashed var(--border);
    border-radius: var(--radius-lg);
    cursor: pointer;
    transition:
      border-color var(--transition-fast),
      background var(--transition-fast);
    user-select: none;
  }

  .dropzone:hover,
  .dropzone:focus,
  .dropzone.drag-over {
    border-color: var(--primary);
    background: var(--status-info-bg);
    outline: none;
  }

  .dropzone-icon {
    color: var(--text-muted);
    opacity: 0.6;
    transition:
      opacity var(--transition-fast),
      color var(--transition-fast);
  }

  .dropzone:hover .dropzone-icon,
  .dropzone.drag-over .dropzone-icon {
    opacity: 1;
    color: var(--primary);
  }

  .dropzone-title {
    font-size: var(--font-size-lg);
    font-weight: var(--font-weight-semibold);
    color: var(--text);
    margin: 0;
  }

  .dropzone-subtitle {
    font-size: var(--font-size-sm);
    color: var(--text-muted);
    margin: 0;
    text-align: center;
  }
</style>
