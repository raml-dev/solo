<script lang="ts">
  import Button from "./base/Button.svelte";
  import Modal from "./base/Modal.svelte";
  import EnvironmentManager from "./EnvironmentManager.svelte";
  import ThemeSelector from "./ThemeSelector.svelte";

  export let title = "Yapla";
  let showThemeSelector = false;
  let showEnvironmentManager = false;

  function toggleThemeSelector() {
    showThemeSelector = !showThemeSelector;
  }

  function toggleEnvironmentManager() {
    showEnvironmentManager = !showEnvironmentManager;
  }
</script>

<div class="app-container">
  <!-- Top Navbar -->
  <nav class="navbar">
    <div class="flex items-center gap-md">
      <h1 class="text-lg font-semibold">{title}</h1>
    </div>
    <div class="flex items-center gap-sm">
      <slot name="navbar-actions" />
      <Button variant="secondary" click={toggleEnvironmentManager}>Environments</Button>
      <Button variant="secondary" click={toggleThemeSelector}>Theme</Button>
      <Button variant="secondary" click={() => {}}>Settings</Button>
      <Button variant="secondary" click={() => {}}>Help</Button>
    </div>
  </nav>

  <!-- Main ContEnvironmet Area -->
  <div class="main-content">
    <slot />
  </div>
</div>

{#if showThemeSelector}
  <Modal toggleFn={toggleThemeSelector}>
    <ThemeSelector />
  </Modal>
{/if}

{#if showEnvironmentManager}
  <Modal toggleFn={toggleEnvironmentManager}>
    <EnvironmentManager />
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
  }
</style>
