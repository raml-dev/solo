<script lang="ts">
  import Button from "flowbite-svelte/Button.svelte";
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
      // If it's a new unsaved request, we can't save it directly via saveTab
      // because it needs a collection first. For now, let's just close it or
      // we could trigger the Save modal from HTTPRequestBuilder.
      // But tabStore.saveTab handles the existing request case.
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

<div class="tab-bar">
  <div class="tab-list">
    {#each tabs as tab (tab.id)}
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <div
        class="tab"
        class:active={tab.id === activeTabId}
        class:preview-tab={tab.isPreview}
        onclick={() => tabStore.setActiveTab(tab.id)}
        ondblclick={() => tabStore.fixTab(tab.id)}
        title={tab.label}
      >
        <span class={getMethodBadgeClass(tab.verb)}>{tab.verb}</span>
        <span class="tab-label">{tab.label}</span>
        {#if tab.isDirty}
          <span class="dirty-dot" title="Unsaved changes"></span>
        {/if}
        <button class="close-btn" onclick={(e) => handleClose(e, tab.id)} aria-label="Close tab"
          >×</button
        >
      </div>
    {/each}
  </div>

  <button
    class="new-tab-btn"
    onclick={() => tabStore.newEmptyTab()}
    title="New request"
    aria-label="New request">+</button
  >
</div>

{#if showConfirmClose}
  <Modal title="Unsaved Changes" bind:open={showConfirmClose}>
    {#if $topModalId === confirmCloseModalId}
      <ToastContainer />
    {/if}
    <div class="confirm-modal-body">
      <p>Do you want to save the changes to <strong>{tabToClose?.label}</strong>?</p>
      <p class="text-gray-500 dark:text-gray-400">
        Your changes will be lost if you don't save them.
      </p>

      <div class="confirm-modal-actions mt-4 flex items-center gap-2">
        <Button color="red" onclick={confirmCloseDiscard}>Don't Save</Button>
        <div class="ml-auto flex items-center gap-2">
          <Button color="light" onclick={closeConfirmModal}>Cancel</Button>
          <Button color="primary" onclick={confirmCloseSave}>Save</Button>
        </div>
      </div>
    </div>
  </Modal>
{/if}
