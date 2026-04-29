<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import EnvironmentManager from "$src/lib/components/Environment/EnvironmentManager.svelte";
  import MainConfiguration from "$src/lib/components/MainConfiguration.svelte";
  import UpdateBanner from "$src/lib/components/UpdateBanner.svelte";
  import ToastContainer from "$src/lib/components/base/ToastContainer.svelte";
  import SoloSvg from "$src/lib/components/common/SoloSvg.svelte";
  import { modalStack, topModalId } from "$src/lib/stores/modalStackStore.svelte";
  import CogSolid from "flowbite-svelte-icons/CogSolid.svelte";
  import GlobeSolid from "flowbite-svelte-icons/GlobeSolid.svelte";
  import Button from "flowbite-svelte/Button.svelte";
  import Modal from "flowbite-svelte/Modal.svelte";
  import { onDestroy } from "svelte";

  interface Props {
    navbar_actions?: import("svelte").Snippet;
    children?: import("svelte").Snippet;
    bottom_bar?: import("svelte").Snippet;
  }

  let { navbar_actions, children, bottom_bar }: Props = $props();
  const environmentManagerModal = modalStack.createModal("layout-environments");
  const settingsModal = modalStack.createModal("layout-settings");

  function toggleEnvironmentManager() {
    environmentManagerModal.open = !environmentManagerModal.open;
  }

  function toggleMainConfiguration() {
    settingsModal.open = !settingsModal.open;
  }

  onDestroy(() => {
    modalStack.destroyModal(environmentManagerModal.id);
    modalStack.destroyModal(settingsModal.id);
  });
</script>

<div class="flex h-screen flex-col overflow-hidden">
  <!-- Top Navbar -->
  <nav
    class="flex h-12 shrink-0 items-center justify-between border-b border-neutral-200 bg-white px-3 dark:border-neutral-700 dark:bg-neutral-800"
  >
    <div class="flex items-center">
      <h1 aria-label="Solo" class="select-none">
        <SoloSvg
          aria-hidden="true"
          size={32}
          class="pointer-events-none cursor-default text-primary-700 select-none dark:text-primary-500"
        />
      </h1>
    </div>
    <div class="flex items-center gap-2">
      {@render navbar_actions?.()}
      <Button class="h-8 gap-1" color="light" onclick={toggleEnvironmentManager}
        ><GlobeSolid class="-ml-1.5 h-4" />Environments</Button
      >
      <Button class="h-8 gap-1" color="light" onclick={toggleMainConfiguration}
        ><CogSolid class="-ml-1.5 h-4" />Settings</Button
      >
    </div>
  </nav>

  <UpdateBanner />

  <!-- Main Content Area -->
  <div class="flex min-h-0 flex-1 overflow-hidden">
    {@render children?.()}
  </div>

  <!-- Bottom Bar (always visible) -->
  <div class="w-full shrink-0">
    {@render bottom_bar?.()}
  </div>
</div>

{#if environmentManagerModal.open}
  <Modal
    bind:open={environmentManagerModal.open}
    classes={{ body: "h-full overflow-y-auto" }}
    fullscreen
    size="none"
  >
    {#if $topModalId === environmentManagerModal.id}
      <ToastContainer />
    {/if}
    <EnvironmentManager />
  </Modal>
{/if}

{#if settingsModal.open}
  <Modal
    title="Settings"
    bind:open={settingsModal.open}
    size="xl"
    classes={{ body: "h-[600px] overflow-hidden p-4" }}
  >
    {#if $topModalId === settingsModal.id}
      <ToastContainer />
    {/if}
    <MainConfiguration />
  </Modal>
{/if}
