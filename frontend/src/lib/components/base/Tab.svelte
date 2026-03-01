<script lang="ts">
  import Tab from "../../../types/Tab";
  import { getContext, onMount } from "svelte";
  import type { Writable } from "svelte/store";

  export let title = "";
  export let icon = "";
  export let value = Symbol();

  const tabsStore = getContext<Writable<Tab[]>>("tabs");
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
