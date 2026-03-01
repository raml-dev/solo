<script lang="ts">
  import { getContext, onMount } from "svelte";
  import type { Writable } from "svelte/store";

  type TabItem = {
    title: string;
    icon?: string;
    value: symbol;
  };

  export let title = "";
  export let icon = "";
  export let value = Symbol();

  const tabsStore = getContext<Writable<TabItem[]>>("tabs");
  const activeTabStore = getContext<Writable<symbol>>("activeTab");

  onMount(() => {
    if (!$activeTabStore) {
      $activeTabStore = value;
    }

    const item = { title, value, icon };
    $tabsStore = [...$tabsStore, item];
    return () => {
      $tabsStore = $tabsStore.splice($tabsStore.indexOf(item), 1);
    };
  });
</script>

{#if $activeTabStore === value}
  <slot />
{/if}
