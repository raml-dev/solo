<script lang="ts">
  import Button from "./base/Button.svelte";
  import Modal from "./base/Modal.svelte";
  import EnvironmentManager from "./EnvironmentManager.svelte";
  import MainConfiguration from "./MainConfiguration.svelte";

  export let title = "Yapla";
  let showEnvironmentManager = false;
  let showMainConfiguration = false;

  // Bound from MainConfiguration
  let configIsDirty = false;
  let configIsLoading = false;
  let saveConfig: () => Promise<void> = async () => {};
  let revertConfig: () => void = () => {};

  function toggleEnvironmentManager() {
    showEnvironmentManager = !showEnvironmentManager;
  }

  function toggleMainConfiguration() {
    // If closing, revert any unsaved changes (like theme preview)
    if (showMainConfiguration) {
      revertConfig();
    }
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
      <slot name="navbar-actions" />
      <Button variant="secondary" click={toggleEnvironmentManager}>Environments</Button>
      <Button variant="secondary" click={toggleMainConfiguration}>Settings</Button>
      <Button variant="secondary" click={() => {}}>Help</Button>
    </div>
  </nav>

  <!-- Main ContEnvironmet Area -->
  <div class="main-content">
    <slot />
  </div>
</div>

{#if showEnvironmentManager}
  <Modal title="Environments" toggleFn={toggleEnvironmentManager}>
    <EnvironmentManager />
  </Modal>
{/if}

{#if showMainConfiguration}
  <Modal title="Settings" toggleFn={toggleMainConfiguration}>
    <svelte:fragment slot="additional-buttons">
      <Button variant="primary" click={saveConfig} disabled={!configIsDirty || configIsLoading}>
        {configIsLoading ? "Saving..." : "Save"}
      </Button>
    </svelte:fragment>
    <MainConfiguration
      bind:isDirty={configIsDirty}
      bind:isLoading={configIsLoading}
      bind:save={saveConfig}
      bind:revert={revertConfig}
    />
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
