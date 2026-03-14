<script lang="ts">
  interface Props {
    message?: string;
    icon?: string;
    detail?: string;
    type?: "default" | "error";
    children?: import("svelte").Snippet;
  }

  let { message = "", icon = "", detail = "", type = "default", children }: Props = $props();
</script>

<div class="empty-state" class:empty-state-error={type === "error"}>
  {#if icon}
    <span class="empty-state-icon">{icon}</span>
  {/if}
  <p class="empty-state-message">{message}</p>
  {#if detail}
    <p class="empty-state-detail">{detail}</p>
  {/if}
  {@render children?.()}
</div>

<style>
  .empty-state {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: var(--space-sm);
    color: var(--text-muted);
    font-size: var(--font-size-sm);
    background: var(--bg-secondary);
    padding: var(--space-xl);
  }

  .empty-state-error {
    color: var(--danger);
  }

  .empty-state-icon {
    font-size: 1.5rem;
    opacity: 0.6;
  }

  .empty-state-message {
    margin: 0;
    font-style: italic;
    text-align: center;
  }

  .empty-state-detail {
    margin: 0;
    font-size: var(--font-size-xs, 0.72rem);
    color: var(--text-muted);
    text-align: center;
    max-width: 400px;
    word-break: break-word;
    font-family: var(--font-mono);
  }
</style>
