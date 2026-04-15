<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import Toast from "flowbite-svelte/Toast.svelte";
  import ToastContainer from "flowbite-svelte/ToastContainer.svelte";
  import { notifications, type NotificationType } from "$src/lib/stores/notificationStore";

  function mapColor(type: NotificationType): "green" | "red" | "yellow" | "blue" {
    if (type === "success") return "green";
    if (type === "error") return "red";
    if (type === "warning") return "yellow";
    return "blue";
  }

  const icons: Record<NotificationType, string> = {
    success: "✓",
    error: "✕",
    warning: "⚠",
    info: "ℹ"
  };
</script>

{#if $notifications.length > 0}
  <ToastContainer position="top-right" class="z-50">
    {#each $notifications as notification (notification.id)}
      <Toast
        color={mapColor(notification.type)}
        dismissable={true}
        onclose={() => notifications.dismiss(notification.id)}
        class="w-96"
      >
        {#snippet icon()}
          <span class="inline-flex h-5 w-5 items-center justify-center"
            >{icons[notification.type]}</span
          >
        {/snippet}

        <div class="flex flex-col">
          <span>{notification.message}</span>
          {#if notification.detail}
            <span class="text-xs opacity-90">{notification.detail}</span>
          {/if}
        </div>
      </Toast>
    {/each}
  </ToastContainer>
{/if}
