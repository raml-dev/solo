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

<style>
  .dropdown {
    position: relative;
    width: 100%;
  }

  .dropdown.variant-minimal-wrapper {
    width: auto;
  }

  .dropdown-trigger {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-sm) var(--space-md);
    background: var(--bg-secondary);
    color: var(--text);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-semibold);
    font-family: var(--font-sans);
    cursor: pointer;
    transition: all var(--transition-fast);
  }

  .dropdown-trigger:hover {
    background: var(--bg-tertiary);
  }

  .dropdown-trigger.open {
    border-color: var(--primary);
  }

  .dropdown-trigger.variant-minimal {
    background: none;
    border: none;
    font-size: var(--font-size-xs);
    font-weight: var(--font-weight-semibold);
    color: var(--text);
    padding: 2px var(--space-xs);
    border-radius: var(--radius-sm);
    gap: 2px;
    width: auto;
  }
  .dropdown-trigger.variant-minimal:hover {
    background: var(--bg-tertiary);
  }
  .dropdown-trigger.variant-minimal .dropdown-value {
    text-transform: uppercase;
    letter-spacing: 0.03em;
  }
  .dropdown-trigger.variant-minimal .dropdown-arrow {
    fill: var(--text);
  }

  /* When variant is minimal, align menu to the right of the trigger */
  .dropdown:has(:global(.variant-minimal)) .dropdown-menu {
    left: auto;
    right: 0;
    min-width: 120px;
  }

  /* square: no border-radius (for use inside a composed bar) */
  .dropdown-trigger.square {
    border-radius: 0;
  }

  /* url-method variant: compact, colored by HTTP method, no outer border */
  .dropdown-trigger.variant-url-method {
    background: transparent;
    border: none;
    border-radius: 0;
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-semibold);
    padding: 0 var(--space-md);
    min-width: 90px;
    height: 100%;
    color: var(--text);
  }
  .dropdown-trigger.variant-url-method:hover {
    background: var(--bg-tertiary);
  }
  .dropdown-trigger.variant-url-method .dropdown-value {
    font-weight: var(--font-weight-semibold);
  }
  .dropdown:has(:global(.variant-url-method)) .dropdown-menu {
    min-width: 130px;
    left: 0;
    right: auto;
  }
  .dropdown.variant-url-method-wrapper {
    width: auto;
    height: 100%;
    display: flex;
    align-items: stretch;
  }

  /* HTTP method colors */
  .dropdown-trigger.variant-url-method[data-method="GET"] {
    color: #4ec9a4;
  }
  .dropdown-trigger.variant-url-method[data-method="POST"] {
    color: #f5a623;
  }
  .dropdown-trigger.variant-url-method[data-method="PUT"] {
    color: #6c9ef8;
  }
  .dropdown-trigger.variant-url-method[data-method="PATCH"] {
    color: #b57bee;
  }
  .dropdown-trigger.variant-url-method[data-method="DELETE"] {
    color: #e06c75;
  }
  .dropdown-trigger.variant-url-method[data-method="HEAD"] {
    color: #56b6c2;
  }
  .dropdown-trigger.variant-url-method[data-method="OPTIONS"] {
    color: var(--text-muted);
  }

  .dropdown-trigger.disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .dropdown-value {
    flex: 1;
    text-align: left;
  }

  .dropdown-arrow {
    fill: var(--text-muted);
    transition: transform var(--transition-fast);
    flex-shrink: 0;
  }

  .dropdown-arrow.rotated {
    transform: rotate(180deg);
  }

  .dropdown-menu {
    position: absolute;
    top: calc(100% + var(--space-xs));
    left: 0;
    right: 0;
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-lg);
    z-index: var(--z-dropdown);
    max-height: 300px;
    overflow-y: auto;
    padding: var(--space-xs);
  }

  .dropdown-option {
    width: 100%;
    padding: var(--space-sm) var(--space-md);
    background: none;
    color: var(--text);
    border: none;
    border-radius: var(--radius-sm);
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-medium);
    font-family: var(--font-sans);
    cursor: pointer;
    transition: all var(--transition-fast);
    text-align: left;
  }

  .dropdown-option:hover {
    background: var(--bg-tertiary);
  }

  .dropdown-option.selected {
    background: var(--primary);
    color: var(--bg-primary);
  }

  .option-badge {
    display: inline-block;
    padding: var(--space-xs) var(--space-sm);
    border-radius: var(--radius-sm);
    font-weight: var(--font-weight-semibold);
    text-transform: uppercase;
    font-size: var(--font-size-xs);
  }

  /* Scrollbar styling */
  .dropdown-menu::-webkit-scrollbar {
    width: 8px;
  }

  .dropdown-menu::-webkit-scrollbar-track {
    background: var(--bg-secondary);
  }

  .dropdown-menu::-webkit-scrollbar-thumb {
    background: var(--border-dark);
    border-radius: var(--radius-sm);
  }

  .dropdown-menu::-webkit-scrollbar-thumb:hover {
    background: var(--text-muted);
  }
</style>
