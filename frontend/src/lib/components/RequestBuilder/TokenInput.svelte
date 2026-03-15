<script lang="ts">
  import { envAutocomplete } from "$src/lib/actions/envAutocomplete";
  import { hideTokenTooltipDelay, showTokenTooltip } from "$src/lib/stores/tokenTooltipStore";
  import type { TextSegment } from "$src/lib/utils/tokens";
  import { splitTextSegments } from "$src/lib/utils/tokens";

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
