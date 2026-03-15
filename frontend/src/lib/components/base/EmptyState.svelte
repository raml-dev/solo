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
