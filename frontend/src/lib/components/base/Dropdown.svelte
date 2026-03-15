<script lang="ts">
  import { onDestroy, onMount } from "svelte";

  interface Props {
    value?: string;
    options?: Array<{ value: string; label: string; color?: string }>;
    placeholder?: string;
    disabled?: boolean;
    variant?: "default" | "minimal" | "url-method";
    square?: boolean;
    change: (value: string) => void;
  }

  let {
    value = $bindable(""),
    options = [],
    placeholder = "Select...",
    disabled = false,
    variant = "default",
    square = false,
    change
  }: Props = $props();

  let isOpen = $state(false);
  let dropdownElement: HTMLDivElement | undefined = $state();
  let selectedLabel = $derived(options.find((opt) => opt.value === value)?.label ?? placeholder);

  function toggleDropdown() {
    if (!disabled) {
      isOpen = !isOpen;
    }
  }

  function selectOption(option: { value: string; label: string; color?: string }) {
    value = option.value;
    isOpen = false;
    change(option.value);
  }

  function handleClickOutside(event: MouseEvent) {
    if (dropdownElement && !dropdownElement.contains(event.target as Node)) {
      isOpen = false;
    }
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === "Escape") {
      isOpen = false;
    }
  }

  onMount(async () => {
    document.addEventListener("click", handleClickOutside);
    document.addEventListener("keydown", handleKeydown);
  });

  onDestroy(async () => {
    document.removeEventListener("click", handleClickOutside);
    document.removeEventListener("keydown", handleKeydown);
  });
</script>

<div
  class="dropdown"
  class:variant-minimal-wrapper={variant === "minimal"}
  class:variant-url-method-wrapper={variant === "url-method"}
  bind:this={dropdownElement}
>
  <button
    class="dropdown-trigger"
    class:variant-minimal={variant === "minimal"}
    class:variant-url-method={variant === "url-method"}
    class:square
    class:disabled
    class:open={isOpen}
    data-method={variant === "url-method" ? value : undefined}
    onclick={toggleDropdown}
    type="button"
  >
    <span class="dropdown-value">{selectedLabel}</span>
    <svg class="dropdown-arrow" class:rotated={isOpen} width="12" height="12" viewBox="0 0 12 12">
      <path d="M6 9L1 4h10z" />
    </svg>
  </button>

  {#if isOpen}
    <div class="dropdown-menu">
      {#each options as option (option.value)}
        <button
          class="dropdown-option"
          class:selected={option.value === value}
          onclick={() => selectOption(option)}
          type="button"
        >
          {#if option.color}
            <span class="option-badge" style="background: {option.color}">
              {option.label}
            </span>
          {:else}
            {option.label}
          {/if}
        </button>
      {/each}
    </div>
  {/if}
</div>
