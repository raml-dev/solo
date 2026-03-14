<script lang="ts">
  export let title: string = "Dialog";
  export let toggleFn: () => void;
  export let size: "default" | "wide" | "settings" | "fullpage" = "default";
</script>

<div
  class="dialog-overlay"
  role="presentation"
  on:click={toggleFn}
  on:keydown={(e) => e.key === "Escape" && toggleFn()}
>
  <div class="dialog" class:wide={size === "wide"} class:settings={size === "settings"} class:fullpage={size === "fullpage"} role="dialog" aria-modal="true" on:click|stopPropagation>
    {#if size === "settings" || size === "fullpage"}
      <div class="settings-close-btn-wrapper">
        <button class="btn-close" on:click={toggleFn}>&times;</button>
      </div>
    {:else}
      <header class="dialog-header">
        <h3>{title}</h3>
        <button class="btn-close" on:click={toggleFn}>&times;</button>
      </header>
    {/if}
    <div class="dialog-content" class:no-padding={size === "settings" || size === "fullpage"}>
      <slot />
    </div>
    {#if size !== "settings" && size !== "fullpage"}
      <footer class="dialog-footer">
        <div class="additional-buttons">
          <slot name="additional-buttons" />
        </div>
        <button class="btn" on:click={toggleFn}>Close</button>
      </footer>
    {/if}
  </div>
</div>

<style>
  .dialog-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .dialog {
    background: var(--bg-secondary);
    border-radius: var(--radius-lg);
    min-width: 500px;
    max-width: 80%;
    max-height: 80vh;
    box-shadow: var(--shadow-xl);
    display: flex;
    flex-direction: column;
    position: relative;
  }

  .dialog.wide {
    min-width: 900px;
    max-width: 90vw;
    max-height: 90vh;
  }

  .dialog.settings {
    min-width: 760px;
    max-width: 90vw;
    width: 860px;
    height: 600px;
    max-height: 90vh;
  }

  .dialog.fullpage {
    width: 100vw;
    height: 100vh;
    max-width: 100vw;
    max-height: 100vh;
    border-radius: 0;
  }

  .settings-close-btn-wrapper {
    position: absolute;
    top: var(--space-sm);
    right: var(--space-sm);
    z-index: 2;
  }

  .dialog-content.no-padding {
    padding: 0;
    overflow: hidden;
  }

  .dialog-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--space-lg);
    border-bottom: 1px solid var(--border);
  }

  .dialog-header h3 {
    margin: 0;
    font-size: var(--font-size-lg);
    font-weight: var(--font-weight-semibold);
  }

  .btn-close {
    background: none;
    border: none;
    font-size: 1.5rem;
    cursor: pointer;
    color: var(--text-muted);
  }

  .dialog-content {
    padding: var(--space-lg);
    overflow-y: auto;
    flex: 1;
  }

  .dialog-footer {
    display: flex;
    justify-content: flex-end;
    align-items: center;
    gap: var(--space-md);
    padding: var(--space-lg);
    border-top: 1px solid var(--border);
  }

  .additional-buttons {
    margin-right: auto;
    display: flex;
    gap: var(--space-sm);
  }

  .btn {
    padding: var(--space-sm) var(--space-md);
    border-radius: var(--radius-md);
    border: 1px solid var(--border);
    background: var(--bg-tertiary);
    color: var(--text);
    cursor: pointer;
  }
</style>
