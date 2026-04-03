<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: GPL-3.0-only
-->

<script lang="ts">
  import Button from "flowbite-svelte/Button.svelte";
  import CloseButton from "flowbite-svelte/CloseButton.svelte";
  import Modal from "flowbite-svelte/Modal.svelte";
  import ToastContainer from "$src/lib/components/base/ToastContainer.svelte";
  import { getActiveTab, tabStore, tabStoreState } from "$src/lib/stores/tabStore.svelte";
  import { modalStack, topModalId } from "$src/lib/stores/modalStackStore.svelte";
  import { getMethodBadgeClass } from "$src/lib/utils/http";
  import { onDestroy } from "svelte";

  let tabs = $derived(tabStoreState.tabs);
  let activeTabId = $derived(getActiveTab().id);

  let tabToCloseId: string | null = $state(null);

  const confirmCloseModal = modalStack.createModal("tabbar-confirm-close");

  onDestroy(() => {
    modalStack.destroyModal(confirmCloseModal.id);
  });

  let tabToClose = $derived(tabs.find((t) => t.id === tabToCloseId));

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

  function handleClose(e: MouseEvent, tabId: string) {
    e.stopPropagation();
    const tab = tabs.find((t) => t.id === tabId);
    if (tab?.isDirty) {
      tabToCloseId = tabId;
      confirmCloseModal.open = true;
    } else {
      tabStore.closeTab(tabId);
    }
  }

  async function confirmCloseSave() {
    if (tabToCloseId) {
      if (tabToClose?.requestId) {
        await tabStore.saveTab(tabToCloseId);
      }
      tabStore.closeTab(tabToCloseId);
    }
    closeConfirmModal();
  }

  function confirmCloseDiscard() {
    if (tabToCloseId) {
      tabStore.closeTab(tabToCloseId);
    }
    closeConfirmModal();
  }

  function closeConfirmModal() {
    confirmCloseModal.open = false;
    tabToCloseId = null;
  }
</script>

<div class="flex items-center gap-2 border-b border-neutral-200 px-2 py-1 dark:border-neutral-700">
  <div
    class="flex min-w-0 flex-1 items-center gap-1 overflow-x-auto"
    role="tablist"
    aria-label="Open request tabs"
  >
    {#each tabs as tab, index (tab.id)}
      <div
        role="tab"
        tabindex="0"
        onclick={() => tabStore.setActiveTab(tab.id)}
        ondblclick={() => (tab.isPreview = false)}
        onkeydown={(event: KeyboardEvent) => handleTabKeydown(event, index, tab.id)}
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
          color="none"
          size="xs"
          class="p-1! opacity-70 transition-opacity group-hover:opacity-100 hover:opacity-100"
          ariaLabel="Close tab"
          onclick={(e: MouseEvent) => handleClose(e, tab.id)}
        />
      </div>
    {/each}
  </div>

  <Button
    color="light"
    size="xs"
    class="shrink-0"
    onclick={() => tabStore.makeEmptyTab()}
    title="New request"
    aria-label="New request"
  >
    +
  </Button>
</div>

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
