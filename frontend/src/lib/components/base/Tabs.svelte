<script lang="ts">
  import { setContext } from "svelte";
  import type { Writable } from "svelte/store";
  import { writable } from "svelte/store";

  type TabItem = {
    title: string;
    value: string;
    badge?: string;
  };

  interface Props {
    activeValue: string;
    variant?: "default" | "minimal";
    children?: import("svelte").Snippet;
  }

  let { activeValue = $bindable(), variant = "default", children }: Props = $props();

  const tabs: Writable<TabItem[]> = writable([]);
  const activeTabStore = writable(activeValue);
  setContext("tabs", tabs);
  setContext("activeTab", activeTabStore);

  // Two-way binding.
  $effect(() => {
    $activeTabStore = activeValue;
  });
</script>

<div class="tabs" class:variant-minimal={variant === "minimal"}>
  {#each $tabs as tab (tab.value)}
    <button
      class="tab"
      class:active={activeValue === tab.value}
      onclick={() => (activeValue = tab.value)}
    >
      <span>{tab.title}</span>
      {#if tab.badge}
        <span class="tab-badge">{tab.badge}</span>
      {/if}
    </button>
  {/each}
</div>

{@render children?.()}

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

  .tab-badge {
    font-size: 0.6rem;
    color: var(--success);
    margin-left: 2px;
    line-height: 1;
  }

  .tabs.variant-minimal {
    padding: 0 0 0 var(--space-xs);
    border-bottom: none;
    flex: 1;
  }
  .tabs.variant-minimal .tab {
    padding: var(--space-sm) var(--space-md);
    font-size: var(--font-size-sm);
  }
</style>
