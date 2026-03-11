<script lang="ts">
  import { createEventDispatcher } from "svelte";

  export let env: { id: string; name: string };
  export let menuOpen: boolean;
  export let isActive: boolean = false;
  export let isFocused: boolean = false;

  const dispatch = createEventDispatcher<{
    delete: string;
    open: string;
    activate: string;
    toggleMenu: string;
  }>();

  function openEnvironment() {
    dispatch("open", env.name);
  }

  function activateEnvironment() {
    dispatch("activate", env.name);
  }

  function toggleMenu(e: Event) {
    e.stopPropagation();
    dispatch("toggleMenu", env.name);
  }

  function handleDeleteEnvironment(e: Event) {
    e.stopPropagation();
    dispatch("delete", env.name);
  }
</script>

<div class="environment-item" class:active={isActive} class:focused={isFocused}>
  <div class="environment-info">
    <input
      class="active-radio"
      type="radio"
      name="active-environment"
      checked={isActive}
      on:change={activateEnvironment}
      aria-label={`Set ${env.name} as active environment`}
    />
    <button class="environment-name-btn" on:click={openEnvironment}>{env.name}</button>
  </div>

  <div class="environment-actions">
    <button class="icon-btn" on:click={toggleMenu} title="More actions" aria-label="More actions">
      ...
    </button>
  </div>

  {#if menuOpen}
    <div class="environment-menu">
      <button class="menu-item danger" on:click={handleDeleteEnvironment}> Delete </button>
    </div>
  {/if}
</div>

<style>
  .environment-item {
    display: flex;
    align-items: center;
    padding: var(--space-sm) var(--space-md);
    border-radius: var(--radius-md);
    border: 1px solid transparent;
    transition: background-color 0.15s;
    position: relative;
  }

  .environment-item:hover,
  .environment-item.focused {
    background-color: var(--bg-tertiary);
  }

  .environment-item.active {
    border-color: var(--primary);
    box-shadow: 0 0 0 1px color-mix(in srgb, var(--primary) 25%, transparent);
  }

  .environment-item.focused:not(.active) {
    border-color: var(--border);
  }

  .environment-info {
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: center;
    gap: var(--space-xs);
  }

  .active-radio {
    margin: 0;
    accent-color: var(--primary);
    cursor: pointer;
  }

  .environment-name-btn {
    background: none;
    border: none;
    color: var(--text);
    font: inherit;
    font-weight: var(--font-weight-medium);
    cursor: pointer;
    padding: 0;
    text-align: left;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }


  .environment-actions {
    opacity: 0;
    pointer-events: none;
    transition: opacity 0.15s;
  }

  .environment-item:hover .environment-actions,
  .environment-item.focused .environment-actions,
  .environment-item.active .environment-actions {
    opacity: 1;
    pointer-events: auto;
  }

  .icon-btn {
    background: none;
    border: 1px solid transparent;
    cursor: pointer;
    padding: 0 var(--space-xs);
    border-radius: var(--radius-sm);
    color: var(--text-muted);
    transition: all var(--transition-fast);
    font-size: var(--font-size-sm);
    height: 24px;
  }
  .icon-btn:hover {
    background: var(--bg-tertiary);
    color: var(--text);
  }
  
  .environment-menu {
    position: absolute;
    right: var(--space-sm);
    top: calc(100% + 6px);
    background: var(--bg-primary);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-md);
    display: flex;
    flex-direction: column;
    min-width: 140px;
    z-index: var(--z-dropdown);
  }

  .menu-item {
    padding: var(--space-sm) var(--space-md);
    background: none;
    border: none;
    text-align: left;
    font-size: var(--font-size-sm);
    color: var(--text);
    cursor: pointer;
  }

  .menu-item:hover {
    background: var(--bg-tertiary);
  }

  .menu-item.danger {
    color: var(--danger);
  }

  .menu-item.danger:hover {
    background: var(--status-danger-bg);
  }
</style>
