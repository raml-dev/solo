<script lang="ts">
  import Button from "flowbite-svelte/Button.svelte";
  import Card from "flowbite-svelte/Card.svelte";
  import Input from "flowbite-svelte/Input.svelte";
  import {
    envAutocomplete,
    type EnvAutocompleteRenderState
  } from "$src/lib/actions/envAutocomplete";
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
  let autocompleteMenuEl: HTMLDivElement | undefined = $state();
  let scrollLeft = $state(0);

  let autocompleteState: EnvAutocompleteRenderState | null = $state(null);
  let autocompleteHandle: ReturnType<typeof envAutocomplete> | null = null;
  const autocompleteListboxId = `token-autocomplete-${Math.random().toString(36).slice(2)}`;

  let segments = $derived(splitTextSegments(value));

  function handleAutocompleteStateChange(state: EnvAutocompleteRenderState) {
    autocompleteState = state;
  }

  $effect(() => {
    if (!inputEl) return;

    autocompleteHandle = envAutocomplete(inputEl, {
      entries: [],
      insertMode: "token",
      onStateChange: handleAutocompleteStateChange
    });

    return () => {
      autocompleteHandle?.destroy();
      autocompleteHandle = null;
      autocompleteState = null;
    };
  });

  $effect(() => {
    if (!autocompleteHandle) return;
    autocompleteHandle.update({
      entries: environmentEntries,
      insertMode: "token",
      menuElement: autocompleteMenuEl ?? null,
      onStateChange: handleAutocompleteStateChange
    });
  });

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
        {#each segments as segment, index (`${segment.isToken ? segment.tokenKey : "text"}-${index}`)}
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
  <Input
    bind:elementRef={inputEl}
    type="text"
    class={`real-input ${inputClass}`}
    bind:value
    {disabled}
    role="combobox"
    aria-autocomplete="list"
    aria-expanded={autocompleteState?.open === true}
    aria-controls={autocompleteListboxId}
    aria-activedescendant={autocompleteState?.open && autocompleteState.items.length > 0
      ? `${autocompleteListboxId}-option-${autocompleteState.activeIndex}`
      : undefined}
    oninput={() => {
      scrollLeft = inputEl?.scrollLeft ?? 0;
      onChange?.();
    }}
    onscroll={() => (scrollLeft = inputEl?.scrollLeft ?? 0)}
  />
</div>

{#if autocompleteState?.open && autocompleteState.items.length > 0}
  <div
    bind:this={autocompleteMenuEl}
    class="fixed z-90"
    style={`left:${autocompleteState.left}px;top:${autocompleteState.top}px;min-width:${autocompleteState.minWidth}px;max-width:${autocompleteState.maxWidth}px;`}
  >
    <Card class="w-full p-2 shadow-lg">
      <div class="mb-1 px-1 text-[11px] text-gray-500 dark:text-gray-400">
        Environment variables
      </div>
      <div id={autocompleteListboxId} role="listbox" class="max-h-56 space-y-1 overflow-auto">
        {#each autocompleteState.items as entry, index (entry.key)}
          <Button
            id={`${autocompleteListboxId}-option-${index}`}
            role="option"
            aria-selected={index === autocompleteState.activeIndex}
            color={index === autocompleteState.activeIndex ? "primary" : "light"}
            size="xs"
            class="flex w-full items-center justify-between gap-2"
            onmouseenter={() => autocompleteState?.setActive(index)}
            onmousedown={(event: MouseEvent) => {
              event.preventDefault();
              autocompleteState?.select(index);
            }}
          >
            <span class="truncate font-mono">{entry.key}</span>
            <span class="max-w-36 truncate text-xs opacity-80">{entry.value}</span>
          </Button>
        {/each}
      </div>
    </Card>
  </div>
{/if}
