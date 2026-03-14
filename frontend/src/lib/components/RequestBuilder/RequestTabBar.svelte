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
      <p class="text-muted">Your changes will be lost if you don't save them.</p>

      <div class="confirm-modal-actions">
        <Button variant="secondary" click={confirmCloseDiscard}>Don't Save</Button>
        <div class="flex-spacer"></div>
        <Button variant="secondary" click={closeConfirmModal}>Cancel</Button>
        <Button variant="primary" click={confirmCloseSave}>Save</Button>
      </div>
    </div>
  </Modal>
{/if}

<style>
  .confirm-modal-body {
    padding: var(--space-md);
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
  }

  .confirm-modal-body p {
    margin: 0;
    font-size: var(--font-size-sm);
  }

  .text-muted {
    color: var(--text-muted);
  }

  .confirm-modal-actions {
    display: flex;
    gap: var(--space-sm);
    margin-top: var(--space-md);
  }

  .flex-spacer {
    flex: 1;
  }

  .tab-bar {
    display: flex;
    align-items: stretch;
    background: var(--bg-secondary);
    border-bottom: 1px solid var(--border);
    height: 38px;
    flex-shrink: 0;
    overflow: hidden;
  }

  .tab-list {
    display: flex;
    align-items: stretch;
    overflow-x: auto;
    flex: 1;
    scrollbar-width: none;
  }

  .tab-list::-webkit-scrollbar {
    display: none;
  }

  .tab {
    display: flex;
    align-items: center;
    gap: var(--space-xs);
    padding: 0 var(--space-sm) 0 var(--space-md);
    border-right: 1px solid var(--border);
    cursor: pointer;
    white-space: nowrap;
    min-width: 0;
    max-width: 200px;
    position: relative;
    color: var(--text-muted);
    font-size: var(--font-size-sm);
    user-select: none;
    transition: background var(--transition-fast);
    flex-shrink: 0;
  }

  .tab:hover {
    background: var(--bg-tertiary);
    color: var(--text);
  }

  .tab.active {
    background: var(--bg-primary);
    color: var(--text);
  }

  .tab.active::after {
    content: "";
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    height: 2px;
    background: var(--primary);
  }

  .method-badge {
    font-size: var(--font-size-xs);
    font-weight: var(--font-weight-semibold);
    padding: 1px 4px;
    border-radius: var(--radius-sm);
    flex-shrink: 0;
  }

  .method-get {
    background: var(--method-get-bg);
    color: var(--method-get-text);
  }
  .method-post {
    background: var(--method-post-bg);
    color: var(--method-post-text);
  }
  .method-put {
    background: var(--method-put-bg);
    color: var(--method-put-text);
  }
  .method-delete {
    background: var(--method-delete-bg);
    color: var(--method-delete-text);
  }
  .method-patch {
    background: var(--method-patch-bg);
    color: var(--method-patch-text);
  }
  .method-head {
    background: var(--bg-tertiary);
    color: var(--text-muted);
  }
  .method-options {
    background: var(--bg-tertiary);
    color: var(--text-muted);
  }

  .tab-label {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    flex: 1;
    min-width: 0;
  }

  .preview-tab .tab-label {
    font-style: italic;
  }

  .dirty-dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--warning);
    flex-shrink: 0;
  }

  .close-btn {
    background: none;
    border: none;
    cursor: pointer;
    color: var(--text-light);
    font-size: var(--font-size-base);
    padding: 0 2px;
    line-height: 1;
    border-radius: var(--radius-sm);
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 18px;
    height: 18px;
    opacity: 0;
    transition:
      opacity var(--transition-fast),
      background var(--transition-fast);
  }

  .tab:hover .close-btn,
  .tab.active .close-btn {
    opacity: 1;
  }

  .close-btn:hover {
    background: var(--bg-tertiary);
    color: var(--danger);
  }

  .new-tab-btn {
    background: none;
    border: none;
    border-left: 1px solid var(--border);
    cursor: pointer;
    color: var(--text-muted);
    font-size: var(--font-size-lg);
    padding: 0 var(--space-md);
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    transition:
      background var(--transition-fast),
      color var(--transition-fast);
  }

  .new-tab-btn:hover {
    background: var(--bg-tertiary);
    color: var(--primary);
  }
</style>
