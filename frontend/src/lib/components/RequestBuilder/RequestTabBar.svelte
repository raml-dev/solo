<script lang="ts">
  import Button from "$src/lib/components/base/Button.svelte";
  import Modal from "$src/lib/components/base/Modal.svelte";
  import { tabStore } from "$src/lib/stores/tabStore";

  let tabs = $derived($tabStore.tabs);
  let activeTabId = $derived($tabStore.activeTabId);

  let tabToCloseId: string | null = $state(null);
  let showConfirmClose = $state(false);

  let tabToClose = $derived(tabs.find((t) => t.id === tabToCloseId));

  function getMethodClass(verb: string): string {
    return `method-${(verb || "get").toLowerCase()}`;
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
        <span class="method-badge {getMethodClass(tab.verb)}">{tab.verb}</span>
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
  <Modal title="Unsaved Changes" toggleFn={closeConfirmModal} size="default">
    <div class="confirm-modal-body">
      <p>Do you want to save the changes to <strong>{tabToClose?.label}</strong>?</p>
      <p class="text-gray-500 dark:text-gray-400">
        Your changes will be lost if you don't save them.
      </p>

      <div class="confirm-modal-actions">
        <Button variant="secondary" click={confirmCloseDiscard}>Don't Save</Button>
        <div class="flex-spacer"></div>
        <Button variant="secondary" click={closeConfirmModal}>Cancel</Button>
        <Button variant="primary" click={confirmCloseSave}>Save</Button>
      </div>
    </div>
  </Modal>
{/if}
