<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';

  export let value: string = '';
  export let options: Array<{ value: string; label: string; color?: string }> = [];
  export let placeholder: string = 'Select...';
  export let disabled: boolean = false;

  const dispatch = createEventDispatcher();
  
  let isOpen = false;
  let dropdownElement: HTMLDivElement;
  let selectedLabel = '';

  $: {
    const selected = options.find(opt => opt.value === value);
    selectedLabel = selected ? selected.label : placeholder;
  }

  function toggleDropdown() {
    if (!disabled) {
      isOpen = !isOpen;
    }
  }

  function selectOption(option: { value: string; label: string; color?: string }) {
    value = option.value;
    selectedLabel = option.label;
    isOpen = false;
    dispatch('change', option.value);
  }

  function handleClickOutside(event: MouseEvent) {
    if (dropdownElement && !dropdownElement.contains(event.target as Node)) {
      isOpen = false;
    }
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Escape') {
      isOpen = false;
    }
  }

  onMount(() => {
    document.addEventListener('click', handleClickOutside);
    document.addEventListener('keydown', handleKeydown);

    return () => {
      document.removeEventListener('click', handleClickOutside);
      document.removeEventListener('keydown', handleKeydown);
    };
  });
</script>

<div class="dropdown" bind:this={dropdownElement}>
  <button 
    class="dropdown-trigger"
    class:disabled
    class:open={isOpen}
    on:click={toggleDropdown}
    type="button"
  >
    <span class="dropdown-value">{selectedLabel}</span>
    <svg 
      class="dropdown-arrow" 
      class:rotated={isOpen}
      width="12" 
      height="12" 
      viewBox="0 0 12 12"
    >
      <path d="M6 9L1 4h10z"/>
    </svg>
  </button>

  {#if isOpen}
    <div class="dropdown-menu">
      {#each options as option}
        <button
          class="dropdown-option"
          class:selected={option.value === value}
          on:click={() => selectOption(option)}
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
    background: var(--bg-tertiary);
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