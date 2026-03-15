<script lang="ts">
  import Button from "flowbite-svelte/Button.svelte";
  import Modal from "flowbite-svelte/Modal.svelte";
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
      <Button color="light" onclick={toggleEnvironmentManager}>Environments</Button>
      <Button color="light" onclick={toggleMainConfiguration}>Settings</Button>
      <Button color="light" onclick={() => {}}>Help</Button>
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
  <Modal bind:open={showEnvironmentManager} fullscreen size="none">
    <EnvironmentManager />
  </Modal>
{/if}

{#if showMainConfiguration}
  <Modal title="Settings" bind:open={showMainConfiguration} size="xl">
    <MainConfiguration />
  </Modal>
{/if}
