<script lang="ts">
  import { envAutocomplete } from "../../actions/envAutocomplete";
  import { splitTextSegments } from "../../utils/tokens";
  import { showTokenTooltip, hideTokenTooltipDelay } from "../../stores/tokenTooltipStore";
  import type { TextSegment } from "../../utils/tokens";

  interface Props {
    value?: string;
    placeholder?: string;
    disabled?: boolean;
    environmentEntries?: { key: string; value: string }[];
    inputClass?: string;
    wrapperClass?: string;
    onChange?: () => void;
  }

  let {
    value = $bindable(""),
    placeholder = "",
    disabled = false,
    environmentEntries = [],
    inputClass = "",
    wrapperClass = "",
    onChange
  }: Props = $props();

  let inputEl: HTMLInputElement | undefined = $state();
  let scrollLeft = $state(0);

  let segments = $derived(splitTextSegments(value));

  function handleTokenEnter(e: MouseEvent, segment: TextSegment) {
    if (!segment.tokenKey) return;
    const target = e.currentTarget as HTMLElement;
    const rect = target.getBoundingClientRect();
    showTokenTooltip(segment.tokenKey, rect.left, rect.bottom);
  }

  function handleTokenLeave() {
    hideTokenTooltipDelay();
  }

  function focusInput() {
    inputEl?.focus();
  }
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="token-input-wrapper {wrapperClass}" onclick={focusInput}>
  <div class="token-input-overlay" aria-hidden="true">
    <div class="token-input-overlay-content" style={`transform: translateX(-${scrollLeft}px);`}>
      {#if !value && placeholder}
        <span class="placeholder">{placeholder}</span>
      {:else}
        {#each segments as segment (segment.tokenKey)}
          {#if segment.isToken}
            <span
              class="token"
              onmouseenter={(e) => handleTokenEnter(e, segment)}
              onmouseleave={handleTokenLeave}>{segment.text}</span
            >
          {:else}
            <span class="text">{segment.text}</span>
          {/if}
        {/each}
      {/if}
    </div>
  </div>
  <input
    bind:this={inputEl}
    type="text"
    class="real-input {inputClass}"
    bind:value
    {disabled}
    oninput={() => {
      scrollLeft = inputEl?.scrollLeft ?? 0;
      onChange?.();
    }}
    onscroll={() => (scrollLeft = inputEl?.scrollLeft ?? 0)}
    use:envAutocomplete={{ entries: environmentEntries, insertMode: "token" }}
  />
</div>

<style>
  .token-input-wrapper {
    flex: 1;
    min-width: 0;
    position: relative;
    display: flex;
    align-items: center;
    background: var(--bg-secondary);
  }

  .token-input-wrapper.input {
    padding: 0;
    min-height: 0;
    align-items: stretch;
  }

  .token-input-wrapper.input .real-input,
  .token-input-wrapper.input .token-input-overlay {
    padding: var(--space-sm) var(--space-md);
    line-height: normal;
  }

  .token-input-overlay {
    position: absolute;
    inset: 0;
    pointer-events: none;
    z-index: 2;
    white-space: pre;
    overflow: hidden;
    padding: var(--space-sm) var(--space-md);
    font: inherit;
  }

  .token-input-overlay-content {
    min-width: 100%;
    color: var(--text);
  }

  .placeholder {
    color: var(--text-light);
  }

  .token {
    color: var(--primary);
    font-weight: var(--font-weight-semibold);
    pointer-events: auto;
    cursor: text;
  }

  .text {
    pointer-events: none;
  }

  .real-input {
    position: relative;
    z-index: 1;
    background: transparent;
    color: transparent;
    caret-color: var(--text);
    width: 100%;
    border: none;
    outline: none;
    padding: var(--space-sm) var(--space-md);
    margin: 0;
    font: inherit;
    line-height: normal;
  }

  .real-input::selection {
    background: rgba(74, 158, 255, 0.25);
    color: transparent;
  }
</style>
