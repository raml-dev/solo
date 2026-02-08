<script lang="ts">
  import Button from "./Button.svelte";

  export let toggleFn: () => void = null;
</script>

<div
  class="modal-overlay"
  role="presentation"
  on:click={(event) => {
    if (event.target == event.currentTarget) toggleFn();
  }}
>
  <div class="modal-panel">
    <slot />
    <div class="modal-footer">
      <Button variant="secondary" on:click={toggleFn}>Close</Button>
      <slot name="additional-buttons" />
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
    background: var(--bg-primary);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: var(--space-md);
    max-width: 800px;
    width: 90%;
    max-height: 90vh;
    overflow-y: auto;
    box-shadow: var(--shadow-lg);
  }

  .modal-footer {
    padding: var(--space-lg);
    border-top: 1px solid var(--border);
    display: flex;
    justify-content: flex-end;
    gap: var(--space-xs);
  }
</style>
