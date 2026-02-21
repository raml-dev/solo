<script lang="ts">
  export let toggleFn: () => void = null;
  export let title: string = "";
</script>

<div
  class="modal-overlay"
  role="presentation"
  on:click={(event) => {
    if (event.target === event.currentTarget) toggleFn?.();
  }}
>
  <div class="modal-panel">
    <header class="modal-header">
      {#if title}
        <h3 class="modal-title">{title}</h3>
      {/if}
      <div class="modal-header-actions">
        <slot name="additional-buttons" />
        <button class="close-btn" on:click={toggleFn} aria-label="Close modal">&times;</button>
      </div>
    </header>
    <div class="modal-body">
      <slot />
    </div>
  </div>
</div>

<style>
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: var(--z-modal);
  }

  .modal-panel {
    display: flex;
    flex-direction: column;
    background: var(--bg-primary);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    max-width: 800px;
    width: 90%;
    max-height: 90vh;
    box-shadow: var(--shadow-lg);
  }

  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-md) var(--space-lg);
    border-bottom: 1px solid var(--border);
    background: var(--bg-secondary);
    border-top-left-radius: var(--radius-lg);
    border-top-right-radius: var(--radius-lg);
    flex-shrink: 0;
  }

  .modal-title {
    margin: 0;
    font-size: var(--font-size-lg);
    font-weight: var(--font-weight-semibold);
    color: var(--text);
  }

  .modal-header-actions {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
  }

  .close-btn {
    background: none;
    border: none;
    font-size: 24px;
    line-height: 1;
    color: var(--text-muted);
    cursor: pointer;
    padding: 4px 8px;
    border-radius: var(--radius-sm);
    transition: all var(--transition-base);
  }

  .close-btn:hover {
    color: var(--text);
    background: var(--bg-tertiary);
  }

  .modal-body {
    flex: 1;
    overflow-y: auto;
    padding: var(--space-lg);
  }
</style>
