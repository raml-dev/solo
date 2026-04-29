<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import ToastContainer from "$src/lib/components/base/ToastContainer.svelte";
  import ContextMenu from "$src/lib/components/common/ContextMenu.svelte";
  import ContextMenuItem from "$src/lib/components/common/ContextMenuItem.svelte";
  import { modalStack, topModalId } from "$src/lib/stores/modalStackStore.svelte";
  import { getActiveTab, tabStore, tabStoreState } from "$src/lib/stores/tabStore.svelte";
  import { getMethodBadgeClass } from "$src/lib/utils/http";
  import PlusOutline from "flowbite-svelte-icons/PlusOutline.svelte";
  import Button from "flowbite-svelte/Button.svelte";
  import CloseButton from "flowbite-svelte/CloseButton.svelte";
  import Modal from "flowbite-svelte/Modal.svelte";
  import { onDestroy, tick } from "svelte";
  import RequestTabDropdown from "./RequestTabDropdown.svelte";

  // --- Derived state ---
  let tabs = $derived(tabStoreState.tabs);
  let activeTabId = $derived(getActiveTab()?.id ?? null);
  let tabToClose = $derived(tabs.find((t) => t.id === tabToCloseId));

  // --- Scroll state ---
  let scrollEl = $state<HTMLElement | null>(null);
  let fadeLeft = $state(false);
  let fadeRight = $state(false);

  // --- Close confirmation state ---
  let tabToCloseId: string | null = $state(null);
  const confirmCloseModal = modalStack.createModal("tabbar-confirm-close");

  // --- Context menu state ---
  let contextMenuTabId: string | null = $state(null);
  let contextMenuX = $state(0);
  let contextMenuY = $state(0);
  let contextMenuOpen = $state(false);
  let contextMenuKey = $state(0);

  // --- Close All flow ---
  let isClosingAll = $state(false);

  function updateFades() {
    if (!scrollEl) return;
    fadeLeft = scrollEl.scrollLeft > 0;
    fadeRight = scrollEl.scrollLeft + scrollEl.clientWidth < scrollEl.scrollWidth - 1;
  }

  $effect(() => {
    if (!scrollEl) return;
    updateFades();
    const ro = new ResizeObserver(updateFades);
    ro.observe(scrollEl);
    return () => ro.disconnect();
  });

  // Auto-scroll to the active tab whenever it changes
  $effect(() => {
    const id = activeTabId;
    if (!id || !scrollEl) return;
    // Run after DOM update
    tick().then(() => {
      if (!scrollEl) return;
      const el = scrollEl.querySelector<HTMLElement>(`[data-tab-id="${id}"]`);
      if (!el) return;
      const tabEl = el.closest<HTMLElement>('[role="tab"]') ?? el;
      const containerWidth = scrollEl.clientWidth;
      const elLeft = tabEl.offsetLeft;
      const elWidth = tabEl.offsetWidth;
      const targetScroll = elLeft - containerWidth / 2 + elWidth / 2;
      const maxScroll = scrollEl.scrollWidth - containerWidth;
      scrollEl.scrollTo({
        left: Math.max(0, Math.min(targetScroll, maxScroll)),
        behavior: "smooth"
      });
    });
  });

  onDestroy(() => {
    modalStack.destroyModal(confirmCloseModal.id);
  });

  // --- Scroll functions ---

  // --- Tab navigation ---
  function focusTabByIndex(index: number) {
    if (!tabs.length) return;
    const normalized = ((index % tabs.length) + tabs.length) % tabs.length;
    const targetTabId = tabs[normalized]?.id;
    if (!targetTabId) return;
    tabStore.setActiveTab(targetTabId);
    const button = document.querySelector<HTMLButtonElement>(`[data-tab-id="${targetTabId}"]`);
    button?.focus();
  }

  function handleTabKeydown(event: KeyboardEvent, index: number, tabId: string) {
    if (event.key === "ArrowRight") {
      event.preventDefault();
      focusTabByIndex(index + 1);
      return;
    }

    if (event.key === "ArrowLeft") {
      event.preventDefault();
      focusTabByIndex(index - 1);
      return;
    }

    if (event.key === "Home") {
      event.preventDefault();
      focusTabByIndex(0);
      return;
    }

    if (event.key === "End") {
      event.preventDefault();
      focusTabByIndex(tabs.length - 1);
      return;
    }

    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      tabStore.setActiveTab(tabId);
      return;
    }

    if (event.key === "Delete") {
      event.preventDefault();
      const synthetic = new MouseEvent("click");
      handleClose(synthetic, tabId);
    }
  }

  // --- Close single tab ---
  function handleClose(e: MouseEvent | KeyboardEvent, tabId: string) {
    e.stopPropagation();
    if (e instanceof MouseEvent || e.key === "Space") {
      const tab = tabs.find((t) => t.id === tabId);
      if (tab?.isDirty) {
        tabToCloseId = tabId;
        confirmCloseModal.open = true;
      } else {
        tabStore.closeTab(tabId);
      }
    }
  }

  // --- Close confirmation modal ---
  async function confirmCloseSave() {
    if (tabToCloseId) {
      if (tabToClose?.requestId) {
        await tabStore.saveTab(tabToCloseId);
      }
      tabStore.closeTab(tabToCloseId);
    }
    confirmCloseModal.open = false;
    tabToCloseId = null;
    if (isClosingAll) {
      closeNextDirtyTab();
    }
  }

  function confirmCloseDiscard() {
    if (tabToCloseId) {
      tabStore.closeTab(tabToCloseId);
    }
    confirmCloseModal.open = false;
    tabToCloseId = null;
    if (isClosingAll) {
      closeNextDirtyTab();
    }
  }

  function closeConfirmModal() {
    confirmCloseModal.open = false;
    tabToCloseId = null;
    isClosingAll = false;
  }

  // --- Context menu ---
  function handleTabContextMenu(event: MouseEvent, tabId: string) {
    event.preventDefault();
    event.stopPropagation();
    contextMenuOpen = false;
    contextMenuTabId = tabId;
    contextMenuX = event.clientX;
    contextMenuY = event.clientY;
    contextMenuKey++;
    tick().then(() => {
      contextMenuOpen = true;
    });
  }

  function closeContextMenu() {
    contextMenuOpen = false;
    contextMenuTabId = null;
  }

  function getContextMenuTriggerId(): string {
    return `tab-context-menu-trigger-${contextMenuKey}`;
  }

  function getContextMenuPositionStyle(): string {
    return `left: ${contextMenuX + 2}px; top: ${contextMenuY + 2}px;`;
  }

  // --- Close All flow ---
  function handleCloseAll() {
    closeContextMenu();
    isClosingAll = true;
    closeNextDirtyTab();
  }

  function closeNextDirtyTab() {
    const dirtyTab = tabs.find((t) => t.isDirty);
    if (dirtyTab) {
      tabToCloseId = dirtyTab.id;
      tabStore.setActiveTab(dirtyTab.id);
      confirmCloseModal.open = true;
    } else {
      // All dirty tabs resolved — close the remaining non-dirty ones
      isClosingAll = false;
      tabStore.closeNonDirtyTabs();
    }
  }

  function handleContextClose() {
    if (!contextMenuTabId) return;
    const tabId = contextMenuTabId;
    closeContextMenu();
    const synthetic = new MouseEvent("click");
    handleClose(synthetic, tabId);
  }
</script>

<div class="flex items-center gap-2 border-b border-neutral-200 px-2 py-1 dark:border-neutral-700">
  <div class="relative min-w-0 flex-1">
    {#if fadeLeft}
      <div
        class="pointer-events-none absolute inset-y-0 left-0 z-10 w-10 bg-gradient-to-r from-white to-transparent dark:from-neutral-900"
      ></div>
    {/if}
    {#if fadeRight}
      <div
        class="pointer-events-none absolute inset-y-0 right-0 z-10 w-10 bg-gradient-to-l from-white to-transparent dark:from-neutral-900"
      ></div>
    {/if}
    <div
      bind:this={scrollEl}
      onscroll={updateFades}
      class="flex items-center gap-1 overflow-x-auto [&::-webkit-scrollbar]:hidden"
      role="tablist"
      aria-label="Open request tabs"
    >
      {#each tabs as tab, index (tab.id)}
        <div
          role="tab"
          tabindex="0"
          onclick={() => tabStore.setActiveTab(tab.id)}
          ondblclick={() => (tab.isPreview = false)}
          onmouseup={(e: MouseEvent) => {
            e.preventDefault();
            // button 1 is middle click
            if (e.button === 1) handleClose(e, tab.id);
          }}
          onkeydown={(event: KeyboardEvent) => handleTabKeydown(event, index, tab.id)}
          oncontextmenu={(event: MouseEvent) => handleTabContextMenu(event, tab.id)}
          class={`group inline-flex max-w-xs items-center rounded-md border inset-ring-primary-500 focus-within:inset-ring-1 focus-within:outline-hidden focus:outline-hidden ${
            tab.id === activeTabId
              ? "border-primary-300 bg-primary-50 dark:border-primary-700 dark:bg-primary-900/40"
              : "border-transparent bg-neutral-100/70 hover:bg-neutral-200/70 dark:bg-neutral-800/60 dark:hover:bg-neutral-700/70"
          } ${tab.isPreview ? "italic opacity-85" : ""}`}
        >
          <Button
            color="light"
            size="xs"
            tabindex={-1}
            class="inline-flex min-w-0 items-center gap-2 border-0 bg-transparent px-2 py-1.5 text-sm shadow-none hover:bg-transparent focus:ring-0 focus:outline-none dark:bg-transparent dark:hover:bg-transparent"
            aria-selected={tab.id === activeTabId}
            data-tab-id={tab.id}
            title={tab.label}
          >
            <span class={getMethodBadgeClass(tab.verb)}>{tab.verb}</span>
            <span class="max-w-48 truncate">{tab.label}</span>
            {#if tab.isDirty}
              <span class="h-2 w-2 rounded-full bg-warning-500" title="Unsaved changes"></span>
            {/if}
          </Button>

          <CloseButton
            tabindex={0}
            color="none"
            size="xs"
            class="p-1! opacity-70 inset-ring-primary-500 transition-opacity group-hover:opacity-100 focus-within:inset-ring-1 focus-within:outline-hidden hover:opacity-100 focus:outline-hidden"
            ariaLabel="Close tab"
            onclick={(e: MouseEvent) => handleClose(e, tab.id)}
            onkeydown={(e: KeyboardEvent) => {
              handleClose(e, tab.id);
            }}
          />
        </div>
      {/each}
    </div>
  </div>
  <Button
    color="light"
    size="xs"
    class="h-8 shrink-0 border-none bg-transparent inset-ring-primary-500 focus-within:inset-ring-1 focus-within:outline-hidden hover:bg-neutral-200 focus:ring-0 focus:outline-hidden dark:border-none dark:bg-transparent"
    onclick={() => tabStore.makeEmptyTab()}
    title="New request"
    aria-label="New request"
  >
    <PlusOutline
      class="h-3 w-3 text-neutral-800/70 hover:text-neutral-800 dark:text-neutral-100/70 dark:hover:text-neutral-100"
    />
  </Button>
  <RequestTabDropdown {tabs} {activeTabId} onselect={tabStore.setActiveTab} />
</div>

{#if contextMenuTabId}
  <button
    id={getContextMenuTriggerId()}
    type="button"
    class="pointer-events-none fixed z-90 h-0 w-0 opacity-0"
    style={getContextMenuPositionStyle()}
    tabindex="-1"
    aria-hidden="true"
  ></button>
  <ContextMenu
    triggeredBy={`#${getContextMenuTriggerId()}`}
    isOpen={contextMenuOpen}
    onClose={closeContextMenu}
  >
    <ContextMenuItem onclick={handleContextClose}>Close</ContextMenuItem>
    <ContextMenuItem onclick={handleCloseAll}>Close All</ContextMenuItem>
  </ContextMenu>
{/if}

{#if confirmCloseModal.open}
  <Modal title="Unsaved Changes" bind:open={confirmCloseModal.open}>
    {#if $topModalId === confirmCloseModal.id}
      <ToastContainer />
    {/if}
    <div class="flex flex-col gap-2">
      <p>Do you want to save the changes to <strong>{tabToClose?.label}</strong>?</p>
      <p class="text-neutral-500 dark:text-neutral-400">
        Your changes will be lost if you don't save them.
      </p>
    </div>

    {#snippet footer()}
      <div class="flex flex-2 items-center gap-2">
        <Button color="red" onclick={confirmCloseDiscard}>Don't Save</Button>
        <div class="ml-auto flex items-center gap-2">
          <Button color="light" onclick={closeConfirmModal}>Cancel</Button>
          <Button color="primary" onclick={confirmCloseSave}>Save</Button>
        </div>
      </div>
    {/snippet}
  </Modal>
{/if}
