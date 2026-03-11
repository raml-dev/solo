<script lang="ts">
  import { tabStore } from "../../stores/tabStore";

  $: tabs = $tabStore.tabs;
  $: activeTabId = $tabStore.activeTabId;

  function getMethodClass(verb: string): string {
    return `method-${(verb || "get").toLowerCase()}`;
  }

  function handleClose(e: MouseEvent, tabId: string) {
    e.stopPropagation();
    tabStore.closeTab(tabId);
  }
</script>

<div class="tab-bar">
  <div class="tab-list">
    {#each tabs as tab (tab.id)}
      <!-- svelte-ignore a11y-click-events-have-key-events -->
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="tab"
        class:active={tab.id === activeTabId}
        on:click={() => tabStore.setActiveTab(tab.id)}
        title={tab.label}
      >
        <span class="method-badge {getMethodClass(tab.verb)}">{tab.verb}</span>
        <span class="tab-label">{tab.label}</span>
        {#if tab.isDirty}
          <span class="dirty-dot" title="Unsaved changes"></span>
        {/if}
        <button
          class="close-btn"
          on:click={(e) => handleClose(e, tab.id)}
          aria-label="Close tab"
        >×</button>
      </div>
    {/each}
  </div>

  <button
    class="new-tab-btn"
    on:click={() => tabStore.newEmptyTab()}
    title="New request"
    aria-label="New request"
  >+</button>
</div>

<style>
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

  .method-get    { background: var(--method-get-bg);    color: var(--method-get-text); }
  .method-post   { background: var(--method-post-bg);   color: var(--method-post-text); }
  .method-put    { background: var(--method-put-bg);    color: var(--method-put-text); }
  .method-delete { background: var(--method-delete-bg); color: var(--method-delete-text); }
  .method-patch  { background: var(--method-patch-bg);  color: var(--method-patch-text); }
  .method-head   { background: var(--bg-tertiary);      color: var(--text-muted); }
  .method-options{ background: var(--bg-tertiary);      color: var(--text-muted); }

  .tab-label {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    flex: 1;
    min-width: 0;
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
    transition: opacity var(--transition-fast), background var(--transition-fast);
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
    transition: background var(--transition-fast), color var(--transition-fast);
  }

  .new-tab-btn:hover {
    background: var(--bg-tertiary);
    color: var(--primary);
  }
</style>
