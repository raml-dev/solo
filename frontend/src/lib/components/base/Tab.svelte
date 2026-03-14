<script lang="ts">
  import { run } from "svelte/legacy";

  import { getContext, onMount } from "svelte";
  import type { Writable } from "svelte/store";

  type TabItem = {
    title: string;
    value: string;
    badge?: string;
  };

  interface Props {
    title?: string;
    value: string;
    badge?: string | undefined;
    children?: import("svelte").Snippet;
  }

  let { title = "", value, badge = undefined, children }: Props = $props();

  const tabsStore = getContext<Writable<TabItem[]>>("tabs");
  const activeTabStore = getContext<Writable<string>>("activeTab");

  onMount(() => {
    // If no tab is active yet, make this one the default
    if (!$activeTabStore) {
      $activeTabStore = value;
    }

    const item = { title, value, badge };
    $tabsStore = [...$tabsStore, item];
    return () => {
      // Clean up when the tab is destroyed
      $tabsStore = $tabsStore.filter((t) => t.value !== value);
    };
  });

  // Keep badge in sync if it changes reactively
  run(() => {
    const existing = $tabsStore.find((t) => t.value === value);
    if (existing && existing.badge !== badge) {
      $tabsStore = $tabsStore.map((t) => (t.value === value ? { ...t, badge } : t));
    }
  });
</script>

{#if $activeTabStore === value}
  {@render children?.()}
{/if}
