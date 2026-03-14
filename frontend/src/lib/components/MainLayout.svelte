<script lang="ts">
  import Button from "./base/Button.svelte";
  import Modal from "./base/Modal.svelte";
  import EnvironmentManager from "./Environment/EnvironmentManager.svelte";
  import MainConfiguration from "./MainConfiguration.svelte";

  interface Props {
    title?: string;
    navbar_actions?: import("svelte").Snippet;
    children?: import("svelte").Snippet;
    bottom_bar?: import("svelte").Snippet;
  }

  let { title = "Yapla", navbar_actions, children, bottom_bar }: Props = $props();
  let showEnvironmentManager = $state(false);
  let showMainConfiguration = $state(false);

  function toggleEnvironmentManager() {
    showEnvironmentManager = !showEnvironmentManager;
  }

  function toggleMainConfiguration() {
    showMainConfiguration = !showMainConfiguration;
  }
</script>

<div class="app-container">
  <!-- Top Navbar -->
  <nav class="navbar">
    <div class="flex items-center gap-md">
      <h1 class="text-lg font-semibold">{title}</h1>
    </div>
    <div class="flex items-center gap-sm">
      {@render navbar_actions?.()}
      <Button variant="secondary" click={toggleEnvironmentManager}>Environments</Button>
      <Button variant="secondary" click={toggleMainConfiguration}>Settings</Button>
      <Button variant="secondary" click={() => {}}>Help</Button>
    </div>
  </nav>

  <!-- Main Content Area -->
  <div class="main-content">
    {@render children?.()}
  </div>

  <!-- Bottom Bar (always visible) -->
  <div class="bottom_bar">
    {@render bottom_bar?.()}
  </div>
</div>

{#if showEnvironmentManager}
  <Modal toggleFn={toggleEnvironmentManager} size="fullpage">
    <EnvironmentManager />
  </Modal>
{/if}

{#if showMainConfiguration}
  <Modal title="Settings" toggleFn={toggleMainConfiguration} size="settings">
    <MainConfiguration />
  </Modal>
{/if}

<style>
  .app-container {
    height: 100vh;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .navbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--space-md) var(--space-lg);
    background: var(--bg-primary);
    border-bottom: 1px solid var(--border);
    height: var(--navbar-height);
    flex-shrink: 0;
  }

  .main-content {
    flex: 1;
    overflow: hidden;
    display: flex;
    width: 100%;
    min-height: 0;
  }

  .bottom_bar {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    height: 28px;
    background: var(--bg-primary);
    border-top: 1px solid var(--border);
    padding: 0 var(--space-md);
    gap: 2px;
    position: relative; /* anchor for console-panel */
  }
</style>
