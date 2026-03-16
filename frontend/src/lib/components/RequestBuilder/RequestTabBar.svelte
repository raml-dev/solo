<script lang="ts">
  import Button from "flowbite-svelte/Button.svelte";
  import CloseButton from "flowbite-svelte/CloseButton.svelte";
  import Modal from "flowbite-svelte/Modal.svelte";
  import ToastContainer from "$src/lib/components/base/ToastContainer.svelte";
  import { tabStore } from "$src/lib/stores/tabStore";
  import { modalStack, topModalId } from "$src/lib/stores/modalStackStore";
  import { HTTP_METHOD_COLOR_MAP, type MethodSemanticFamily } from "$src/lib/theme/themeModel";
  import { onDestroy } from "svelte";

  let tabs = $derived($tabStore.tabs);
  let activeTabId = $derived($tabStore.activeTabId);

  let tabToCloseId: string | null = $state(null);
  let showConfirmClose = $state(false);

  const tabBarModalScope = `tabbar-${Math.random().toString(36).slice(2)}`;
  const confirmCloseModalId = `${tabBarModalScope}-confirm-close`;

  $effect(() => {
    if (showConfirmClose) {
      modalStack.open(confirmCloseModalId);
    } else {
      modalStack.close(confirmCloseModalId);
    }
  });

  onDestroy(() => {
    modalStack.close(confirmCloseModalId);
  });

  let tabToClose = $derived(tabs.find((t) => t.id === tabToCloseId));

  function getMethodBadgeClass(verb: string): string {
    const base = "rounded px-1 py-0.5 text-[10px] font-semibold uppercase";
    const family =
      HTTP_METHOD_COLOR_MAP[(verb || "GET").toUpperCase() as keyof typeof HTTP_METHOD_COLOR_MAP] ||
      ("neutral" as MethodSemanticFamily);

    if (family === "success") {
      return `${base} bg-success-100 text-success-700 dark:bg-success-900 dark:text-success-300`;
    }

    if (family === "primary") {
      return `${base} bg-primary-100 text-primary-700 dark:bg-primary-900 dark:text-primary-300`;
    }

    if (family === "warning") {
      return `${base} bg-warning-100 text-warning-700 dark:bg-warning-900 dark:text-warning-300`;
    }

    if (family === "danger") {
      return `${base} bg-danger-100 text-danger-700 dark:bg-danger-900 dark:text-danger-300`;
    }

    return `${base} bg-neutral-100 text-neutral-700 dark:bg-neutral-800 dark:text-neutral-300`;
  }

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
      showConfirmClose = true;
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
    showConfirmClose = false;
    tabToCloseId = null;
  }
</script>

<div
  class="flex items-center gap-2 border-b border-neutral-200 px-2 py-1 dark:border-neutral-700"
>
  <div
    class="flex min-w-0 flex-1 items-center gap-1 overflow-x-auto"
    role="tablist"
    aria-label="Open request tabs"
  >
    {#each tabs as tab, index (tab.id)}
      <div
        class={`group inline-flex max-w-xs items-center rounded-md border ${
          tab.id === activeTabId
            ? "border-primary-300 bg-primary-50 dark:border-primary-700 dark:bg-primary-900/40"
            : "border-transparent bg-neutral-100/70 hover:bg-neutral-200/70 dark:bg-neutral-800/60 dark:hover:bg-neutral-700/70"
        } ${tab.isPreview ? "italic opacity-85" : ""}`}
      >
        <Button
          color="light"
          size="xs"
          class="inline-flex min-w-0 items-center gap-2 border-0 bg-transparent px-2 py-1.5 text-sm shadow-none hover:bg-transparent dark:bg-transparent dark:hover:bg-transparent"
          role="tab"
          aria-selected={tab.id === activeTabId}
          tabindex={tab.id === activeTabId ? 0 : -1}
          data-tab-id={tab.id}
          onclick={() => tabStore.setActiveTab(tab.id)}
          ondblclick={() => tabStore.fixTab(tab.id)}
          onkeydown={(event: KeyboardEvent) => handleTabKeydown(event, index, tab.id)}
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
    onclick={() => tabStore.newEmptyTab()}
    title="New request"
    aria-label="New request"
  >
    +
  </Button>
</div>

{#if showConfirmClose}
  <Modal title="Unsaved Changes" bind:open={showConfirmClose}>
    {#if $topModalId === confirmCloseModalId}
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
