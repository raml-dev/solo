<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import AngleDownOutline from "flowbite-svelte-icons/AngleDownOutline.svelte";
  import CloseSidebarSolid from "flowbite-svelte-icons/CloseSidebarSolid.svelte";
  import OpenSidebarSolid from "flowbite-svelte-icons/OpenSidebarSolid.svelte";
  import SearchOutline from "flowbite-svelte-icons/SearchOutline.svelte";
  import Button from "flowbite-svelte/Button.svelte";
  import ButtonGroup from "flowbite-svelte/ButtonGroup.svelte";
  import Dropdown from "flowbite-svelte/Dropdown.svelte";
  import DropdownItem from "flowbite-svelte/DropdownItem.svelte";
  import Input from "flowbite-svelte/Input.svelte";

  const OUTLINE_BUTTON_CLASSES =
    "text-neutral-800/70 hover:text-neutral-800 dark:text-neutral-100/70 dark:hover:text-neutral-100";

  interface CollectionSidebarHeaderProps {
    collapsed: boolean;
    searchQuery: string;
    onToggleCollapse: () => void;
    onCreateCollection: () => void;
    onOpenImportModal: () => void;
  }

  let {
    collapsed,
    searchQuery = $bindable(""),
    onToggleCollapse,
    onCreateCollection,
    onOpenImportModal
  }: CollectionSidebarHeaderProps = $props();

  let isImportMenuOpen: boolean = $state(false);

  function onCloseImportMenu() {
    isImportMenuOpen = false;
  }
</script>

<div class="border-b border-neutral-200 p-3 dark:border-neutral-800">
  <div class="flex items-center justify-between">
    <div class="flex h-10 items-center gap-2">
      <button
        class="h-6 w-6 flex-1 p-0 text-xs hover:cursor-pointer dark:text-white"
        onclick={onToggleCollapse}
        aria-label="Toggle collection list sidebar"
      >
        {#if collapsed}
          <OpenSidebarSolid class={`h-6 w-6 ${OUTLINE_BUTTON_CLASSES}`} />
        {:else}
          <CloseSidebarSolid class={`h-6 w-6 ${OUTLINE_BUTTON_CLASSES}`} />
        {/if}
      </button>
      {#if !collapsed}
        <h3 class="flex-1 text-sm font-semibold text-neutral-800 dark:text-neutral-100">
          Collections
        </h3>
      {/if}
    </div>
    <div class="flex items-center gap-1">
      {#if !collapsed}
        <ButtonGroup>
          <Button
            color="primary"
            class="shrink-0 cursor-pointer border-none inset-ring-primary-500 focus-within:inset-ring-1 focus-within:outline-hidden focus:ring-0 focus:outline-hidden dark:border-none"
            size="xs"
            onclick={onCreateCollection}
            >New
          </Button>
          <Button
            color="primary"
            class="w-0.5 shrink-0 cursor-pointer border-l px-2.5 inset-ring-primary-500 focus-within:inset-ring-1 focus-within:outline-hidden focus:ring-0 focus:outline-hidden dark:border-l-primary-900"
            size="xs"
            id="import-dropdown-button"
            onclick={() => {
              isImportMenuOpen = true;
            }}><AngleDownOutline class="w-4 shrink-0" /></Button
          >
        </ButtonGroup>
        <!-- Import dropdown -->
        <Dropdown
          triggeredBy="#import-dropdown-button"
          isOpen={isImportMenuOpen}
          class="z-50 w-50"
          triggerDelay={0}
          onclose={onCloseImportMenu}
        >
          <DropdownItem
            class="text-gray-900 dark:text-white"
            onclick={() => {
              onOpenImportModal();
              onCloseImportMenu();
            }}
          >
            Import collection...
          </DropdownItem>
        </Dropdown>
      {/if}
    </div>
  </div>

  {#if !collapsed}
    <div class="mt-2 flex items-center gap-2">
      <Input
        size="sm"
        class="flex-1 ps-8"
        type="text"
        clearable
        classes={{ svg: "h-3 w-3" }}
        placeholder="Search collections or requests"
        bind:value={searchQuery}
      >
        {#snippet left()}
          <SearchOutline class="h-4 w-4" />
        {/snippet}
      </Input>
    </div>
  {/if}
</div>
