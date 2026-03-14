<script lang="ts">
  import { notifications } from "../../stores/notificationStore";
  import Toast from "./Toast.svelte";
</script>

{#if $notifications.length > 0}
  <div class="toast-container">
    {#each $notifications as notification (notification.id)}
      <Toast {notification} />
    {/each}
  </div>
{/if}

<style>
  .toast-container {
    position: fixed;
    bottom: var(--space-xl);
    right: var(--space-xl);
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
    z-index: 9999;
    pointer-events: none;
  }

  .toast-container :global(.toast) {
    pointer-events: all;
    animation: slide-in 0.2s ease-out;
  }

  @keyframes slide-in {
    from {
      transform: translateX(110%);
      opacity: 0;
    }
    to {
      transform: translateX(0);
      opacity: 1;
    }
  }
</style>
