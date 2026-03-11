<script lang="ts">
  import { getContext, onMount } from "svelte";
  import type { Writable } from "svelte/store";

  type TabItem = {
    title: string;
    value: string;
  };

  export let title = "";
  export let value: string;

  const tabsStore = getContext<Writable<TabItem[]>>("tabs");
  const activeTabStore = getContext<Writable<string>>("activeTab");

  onMount(() => {
    // If no tab is active yet, make this one the default
    if (!$activeTabStore) {
      $activeTabStore = value;
    }

    const item = { title, value };
    $tabsStore = [...$tabsStore, item];
    return () => {
      // Clean up when the tab is destroyed
      $tabsStore = $tabsStore.filter(t => t.value !== value);
    };
  });
</script>

{#if $activeTabStore === value}
  <slot />
{/if}
