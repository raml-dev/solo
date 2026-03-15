<script lang="ts">
  interface Props {
    title?: string;
    toggleFn: () => void;
    size?: "default" | "wide" | "settings" | "fullpage";
    children?: import("svelte").Snippet;
    additional_buttons?: import("svelte").Snippet;
  }

  let {
    title = "Dialog",
    toggleFn,
    size = "default",
    children,
    additional_buttons
  }: Props = $props();
</script>

<div
  class="dialog-overlay"
  role="presentation"
  onclick={toggleFn}
  onkeydown={(e) => e.key === "Escape" && toggleFn()}
>
  <div
    class="dialog"
    class:wide={size === "wide"}
    class:settings={size === "settings"}
    class:fullpage={size === "fullpage"}
    role="dialog"
    aria-modal="true"
    onclick={(e) => e.stopPropagation()}
  >
    {#if size === "settings" || size === "fullpage"}
      <div class="settings-close-btn-wrapper">
        <button class="btn-close" onclick={toggleFn}>&times;</button>
      </div>
    {:else}
      <header class="dialog-header">
        <h3>{title}</h3>
        <button class="btn-close" onclick={toggleFn}>&times;</button>
      </header>
    {/if}
    <div class="dialog-content" class:no-padding={size === "settings" || size === "fullpage"}>
      {@render children?.()}
    </div>
    {#if size !== "settings" && size !== "fullpage"}
      <footer class="dialog-footer">
        <div class="additional_buttons">
          {@render additional_buttons?.()}
        </div>
        <button class="btn" onclick={toggleFn}>Close</button>
      </footer>
    {/if}
  </div>
</div>
