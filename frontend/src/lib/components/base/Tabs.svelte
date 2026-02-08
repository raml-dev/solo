<script lang="ts">
  import { setContext } from "svelte";
  import { writable } from "svelte/store";

  export let activeTab: number;
  const items = writable([]);

  const activeTabStore = writable(activeTab);

  setContext("items", items);
  setContext("activeTab", activeTabStore);

  $: activeTab = $activeTabStore;
</script>

<div class="tabs">
  {#each $items as item}
    <button
      class="tab"
      class:active={$activeTabStore === item.value}
      on:click={() => ($activeTabStore = item.value)}
    >
      <span>
        {#if item.icon}
          <i class={item.icon}></i>
        {/if}
        {item.title}
      </span>
    </button>
  {/each}
</div>

<slot />

<style>
  .tabs {
    display: flex;
    background: var(--bg-primary);
    border-bottom: 1px solid var(--border);
    padding: 0 var(--space-lg);
  }

  .tab {
    padding: var(--space-sm) var(--space-md);
    border: none;
    background: none;
    cursor: pointer;
    border-bottom: 2px solid transparent;
    transition: all var(--transition-fast);
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-medium);
    font-family: var(--font-sans);
  }

  .tab > span {
    color: var(--text-muted);
  }

  .tab:hover {
    color: var(--text);
  }

  .tab.active {
    border-bottom-color: var(--primary);
  }

  .tab.active > span {
    color: var(--primary);
  }
</style>
