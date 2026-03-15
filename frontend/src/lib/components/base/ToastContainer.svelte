<script lang="ts">
  import Toast from "$src/lib/components/base/Toast.svelte";
  import { notifications } from "$src/lib/stores/notificationStore";
</script>

{#if $notifications.length > 0}
  <div class="toast-container">
    {#each $notifications as notification (notification.id)}
      <Toast {notification} />
    {/each}
  </div>
{/if}
