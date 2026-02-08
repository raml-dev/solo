<script lang="ts">
  import Button from "./base/Button.svelte";

  export let items: Array<{ id: string; name: string; method?: string }> = [];
  export let onSelect: (id: string) => void = () => {};
  export let selectedId: string | null = null;
</script>

<aside class="sidebar">
  <div class="sidebar-header">
    <h2 class="text-base font-semibold">Collections</h2>
    <Button variant="primary">+ New</Button>
  </div>

  <div class="sidebar-content">
    {#each items as item}
      <button
        class="sidebar-item"
        class:active={item.id === selectedId}
        on:click={() => onSelect(item.id)}
      >
        {#if item.method}
          <span class="method-badge method-{item.method.toLowerCase()}">
            {item.method}
          </span>
        {/if}
        <span class="item-name">{item.name}</span>
      </button>
    {/each}
  </div>
</aside>

<style>
  .sidebar {
    width: var(--sidebar-width);
    background: var(--bg-secondary);
    border-right: 1px solid var(--border);
    display: flex;
    flex-direction: column;
    flex-shrink: 0;
  }

  .sidebar-header {
    padding: var(--space-lg);
    border-bottom: 1px solid var(--border);
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .sidebar-content {
    flex: 1;
    overflow-y: auto;
    padding: var(--space-sm);
  }

  .sidebar-item {
    width: 100%;
    display: flex;
    align-items: center;
    gap: var(--space-sm);
    padding: var(--space-sm) var(--space-md);
    border: none;
    background: none;
    color: var(--text);
    cursor: pointer;
    border-radius: var(--radius-md);
    text-align: left;
    transition: background var(--transition-fast);
    font-family: var(--font-sans);
    font-size: var(--font-size-sm);
  }

  .sidebar-item:hover {
    background: var(--bg-tertiary);
  }

  .sidebar-item.active {
    background: var(--bg-tertiary);
    font-weight: var(--font-weight-medium);
  }

  .item-name {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>
