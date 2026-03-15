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
