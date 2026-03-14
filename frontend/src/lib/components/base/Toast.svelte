<script lang="ts">
  import type { Notification } from "$src/lib/stores/notificationStore";
  import { notifications } from "$src/lib/stores/notificationStore";

  interface Props {
    notification: Notification;
  }

  let { notification }: Props = $props();

  const icons: Record<string, string> = {
    success: "✓",
    error: "✕",
    warning: "⚠",
    info: "ℹ"
  };
</script>

<div class="toast toast-{notification.type}" role="alert">
  <span class="toast-icon">{icons[notification.type]}</span>
  <div class="toast-body">
    <span class="toast-message">{notification.message}</span>
    {#if notification.detail}
      <span class="toast-detail">{notification.detail}</span>
    {/if}
  </div>
  <button class="toast-close" onclick={() => notifications.dismiss(notification.id)}>✕</button>
  {#if !notification.persistent}
    <div class="toast-progress"></div>
  {/if}
</div>

<style>
  .toast {
    display: flex;
    align-items: flex-start;
    gap: var(--space-sm);
    padding: var(--space-md) var(--space-md);
    border-radius: var(--radius-md);
    border: 1px solid transparent;
    background: var(--bg-primary);
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.25);
    min-width: 280px;
    max-width: 420px;
    position: relative;
    overflow: hidden;
  }

  .toast-success {
    border-color: var(--success);
  }
  .toast-error {
    border-color: var(--danger);
  }
  .toast-warning {
    border-color: var(--warning);
  }
  .toast-info {
    border-color: var(--info);
  }

  .toast-icon {
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-semibold);
    flex-shrink: 0;
    width: 18px;
    height: 18px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-top: 1px;
  }

  .toast-success .toast-icon {
    color: var(--success);
  }
  .toast-error .toast-icon {
    color: var(--danger);
  }
  .toast-warning .toast-icon {
    color: var(--warning);
  }
  .toast-info .toast-icon {
    color: var(--info);
  }

  .toast-body {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }

  .toast-message {
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-semibold);
    color: var(--text);
    line-height: 1.4;
  }

  .toast-detail {
    font-size: var(--font-size-xs, 0.72rem);
    color: var(--text-muted);
    line-height: 1.4;
    word-break: break-word;
  }

  .toast-close {
    background: none;
    border: none;
    cursor: pointer;
    color: var(--text-muted);
    font-size: 0.7rem;
    padding: 0;
    line-height: 1;
    flex-shrink: 0;
    opacity: 0.6;
    transition: opacity 0.15s;
  }
  .toast-close:hover {
    opacity: 1;
  }

  .toast-progress {
    position: absolute;
    bottom: 0;
    left: 0;
    height: 2px;
    width: 100%;
    animation: shrink 5s linear forwards;
  }

  .toast-success .toast-progress {
    background: var(--success);
  }
  .toast-error .toast-progress {
    background: var(--danger);
  }
  .toast-warning .toast-progress {
    background: var(--warning);
  }
  .toast-info .toast-progress {
    background: var(--info);
  }

  @keyframes shrink {
    from {
      width: 100%;
    }
    to {
      width: 0%;
    }
  }
</style>
