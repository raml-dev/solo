<script lang="ts">
  import { getContext, onMount } from "svelte";
  import type { Writable } from "svelte/store";

  export let title = "";
  export let icon = "";
  export let value = Symbol();

  const itemsStore = getContext<Writable<any[]>>("items");
  const activeTabStore = getContext<Writable<number>>("activeTab");

  onMount(() => {
    if (!$activeTabStore) {
      $activeTabStore = value as unknown as number; // TS brainf*ck
    }

    const item = { title, value: value, icon };
    $itemsStore = [...$itemsStore, item];
    return () => {
      $itemsStore = $itemsStore.splice($itemsStore.indexOf(item), 1);
    };
  });
</script>

{#if $activeTabStore === value}
  <slot />
{/if}
