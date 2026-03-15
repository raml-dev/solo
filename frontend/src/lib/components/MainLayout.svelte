<script lang="ts">
  import Button from "$src/lib/components/base/Button.svelte";
  import Modal from "$src/lib/components/base/Modal.svelte";
  import EnvironmentManager from "$src/lib/components/Environment/EnvironmentManager.svelte";
  import MainConfiguration from "$src/lib/components/MainConfiguration.svelte";

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
    <div class="flex items-center gap-4">
      <h1 class="text-lg font-semibold">{title}</h1>
    </div>
    <div class="flex items-center gap-2">
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
