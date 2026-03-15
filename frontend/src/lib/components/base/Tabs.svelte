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
