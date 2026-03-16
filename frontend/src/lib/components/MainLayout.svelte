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

  let { title = "solo", navbar_actions, children, bottom_bar }: Props = $props();
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

<div class="flex h-screen flex-col overflow-hidden">
  <!-- Top Navbar -->
  <nav class="flex h-12 shrink-0 items-center justify-between border-b border-neutral-200 bg-white px-3 dark:border-neutral-700 dark:bg-neutral-800">
    <div class="flex items-center gap-4">
      <h1 aria-label="solo" class="select-none">
        <pre aria-hidden="true" class="pointer-events-none cursor-default select-none text-[0.35rem]/[1.3] font-mono text-primary-700 dark:text-primary-500">{title}</pre></h1>
    </div>
    <div class="flex items-center gap-2">
      {@render navbar_actions?.()}
      <Button size="xs" color="light" onclick={toggleEnvironmentManager}>Environments</Button>
      <Button size="xs" color="light" onclick={toggleMainConfiguration}>Settings</Button>
    </div>
  </nav>

  <!-- Main Content Area -->
  <div class="flex min-h-0 flex-1 overflow-hidden">
    {@render children?.()}
  </div>

  <!-- Bottom Bar (always visible) -->
  <div class="w-full shrink-0">
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
  <Modal title="Settings" bind:open={showMainConfiguration} size="xl" bodyClass="h-[600px] overflow-hidden p-4">
    {#if $topModalId === settingsModalId}
      <ToastContainer />
    {/if}
    <MainConfiguration />
  </Modal>
{/if}
