<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import type { TabState } from "$src/lib/stores/tabStore.svelte";
  import { getMethodBadgeClass } from "$src/lib/utils/http";
  import ChevronDownOutline from "flowbite-svelte-icons/ChevronDownOutline.svelte";
  import Button from "flowbite-svelte/Button.svelte";
  import Dropdown from "flowbite-svelte/Dropdown.svelte";
  import DropdownItem from "flowbite-svelte/DropdownItem.svelte";

  interface Props {
    tabs: TabState[];
    activeTabId: string | null;
    onselect: (tabId: string) => void;
  }

  let { tabs, activeTabId, onselect }: Props = $props();

  let isOpen = $state(false);

  function handleSelect(tabId: string) {
    isOpen = false;
    onselect(tabId);
  }
</script>

<Button
  id="tab-list-dropdown-btn"
  color="light"
  size="xs"
  class="h-8 shrink-0 border-none bg-transparent inset-ring-primary-500 focus-within:inset-ring-1 focus-within:outline-hidden hover:bg-neutral-200 focus:ring-0 focus:outline-hidden disabled:cursor-auto dark:border-none dark:bg-transparent"
  title="Open tab list"
  aria-label="Open tab list"
  disabled={tabs.length === 0}
>
  <ChevronDownOutline
    class="h-3 w-3 text-neutral-800/70 hover:text-neutral-800 dark:text-neutral-100/70 dark:hover:text-neutral-100"
  />
</Button>

<Dropdown
  triggeredBy="#tab-list-dropdown-btn"
  bind:isOpen
  triggerDelay={0}
  class="z-50 w-64 overflow-visible! p-0 dark:bg-neutral-700"
>
  <div class="max-h-64 overflow-y-auto py-1">
    {#each tabs as tab (tab.id)}
      <DropdownItem
        class={tab.id === activeTabId ? "bg-primary-50 dark:bg-primary-900/40" : ""}
        onclick={() => handleSelect(tab.id)}
      >
        <div class="flex items-center gap-2 text-neutral-800 dark:text-neutral-100">
          <span class={getMethodBadgeClass(tab.verb)}>{tab.verb}</span>
          <span class="min-w-0 truncate">{tab.label}</span>
          {#if tab.isDirty}
            <span class="h-2 w-2 shrink-0 rounded-full bg-warning-500" title="Unsaved changes"
            ></span>
          {/if}
        </div>
      </DropdownItem>
    {/each}
  </div>
</Dropdown>
