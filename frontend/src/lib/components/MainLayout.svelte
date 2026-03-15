<script lang="ts">
  import Button from "flowbite-svelte/Button.svelte";
  import Modal from "flowbite-svelte/Modal.svelte";
  import EnvironmentManager from "$src/lib/components/Environment/EnvironmentManager.svelte";
  import MainConfiguration from "$src/lib/components/MainConfiguration.svelte";
  import ToastContainer from "$src/lib/components/base/ToastContainer.svelte";
  import { modalStack, topModalId } from "$src/lib/stores/modalStackStore";
  import { onDestroy } from "svelte";

  interface Props {
    title?: string;
    navbar_actions?: import("svelte").Snippet;
    children?: import("svelte").Snippet;
    bottom_bar?: import("svelte").Snippet;
  }

  let { title = "Yapla", navbar_actions, children, bottom_bar }: Props = $props();
  let showEnvironmentManager = $state(false);
  let showMainConfiguration = $state(false);

  const layoutModalScope = `layout-${Math.random().toString(36).slice(2)}`;
  const environmentManagerModalId = `${layoutModalScope}-environments`;
  const settingsModalId = `${layoutModalScope}-settings`;

  $effect(() => {
    if (showEnvironmentManager) {
      modalStack.open(environmentManagerModalId);
    } else {
      modalStack.close(environmentManagerModalId);
    }
  });

  $effect(() => {
    if (showMainConfiguration) {
      modalStack.open(settingsModalId);
    } else {
      modalStack.close(settingsModalId);
    }
  });

  function toggleEnvironmentManager() {
    showEnvironmentManager = !showEnvironmentManager;
  }

  function toggleMainConfiguration() {
    showMainConfiguration = !showMainConfiguration;
  }

  onDestroy(() => {
    modalStack.close(environmentManagerModalId);
    modalStack.close(settingsModalId);
  });
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
    {#if $topModalId === environmentManagerModalId}
      <ToastContainer />
    {/if}
    <EnvironmentManager />
  </Modal>
{/if}

{#if showMainConfiguration}
  <Modal title="Settings" bind:open={showMainConfiguration} size="xl">
    {#if $topModalId === settingsModalId}
      <ToastContainer />
    {/if}
    <MainConfiguration />
  </Modal>
{/if}
