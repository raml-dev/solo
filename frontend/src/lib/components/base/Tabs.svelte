<script lang="ts">
  import { setContext } from "svelte";
  import { writable } from "svelte/store";
  import type { Writable } from "svelte/store";

  // FIXME: use a shared type for this
  // when doing that, currently the app does not start properly - find a fix
  type TabItem = {
    title: string;
    icon?: string;
    value: symbol;
  };

  export let activeTab: symbol;

  const tabs: Writable<TabItem[]> = writable([]);

  const activeTabStore = writable(activeTab);

  setContext("tabs", tabs);
  setContext("activeTab", activeTabStore);

  $: activeTab = $activeTabStore;
</script>

<div class="tabs">
  {#each $tabs as tab (tab.value)}
    <button
      class="tab"
      class:active={$activeTabStore === tab.value}
      on:click={() => ($activeTabStore = tab.value)}
    >
      <span>
        {#if tab.icon}
          <i class={tab.icon}></i>
        {/if}
        {tab.title}
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
    position: sticky;
    top: 0;
    z-index: 1;
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
