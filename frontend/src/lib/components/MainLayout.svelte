<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import EnvironmentManager from "$src/lib/components/Environment/EnvironmentManager.svelte";
  import MainConfiguration from "$src/lib/components/MainConfiguration.svelte";
  import UpdateBanner from "$src/lib/components/UpdateBanner.svelte";
  import ToastContainer from "$src/lib/components/base/ToastContainer.svelte";
  import { environmentStore, environmentStoreState } from "$src/lib/stores/environmentStore.svelte";
  import { modalStack, topModalId } from "$src/lib/stores/modalStackStore.svelte";
  import AngleDownOutline from "flowbite-svelte-icons/AngleDownOutline.svelte";
  import EditOutline from "flowbite-svelte-icons/EditOutline.svelte";
  import Button from "flowbite-svelte/Button.svelte";
  import ButtonGroup from "flowbite-svelte/ButtonGroup.svelte";
  import Dropdown from "flowbite-svelte/Dropdown.svelte";
  import DropdownItem from "flowbite-svelte/DropdownItem.svelte";
  import Modal from "flowbite-svelte/Modal.svelte";
  import { onDestroy } from "svelte";

  interface Props {
    title?: string;
    navbar_actions?: import("svelte").Snippet;
    children?: import("svelte").Snippet;
    bottom_bar?: import("svelte").Snippet;
  }

  let { title = "solo", navbar_actions, children, bottom_bar }: Props = $props();
  const environmentManagerModal = modalStack.createModal("layout-environments");
  const settingsModal = modalStack.createModal("layout-settings");
  let isEnvDropdownOpen = $state(false);
  let environments = $derived(environmentStoreState.environments);
  let selectedEnvironmentName = $derived(environmentStoreState.selectedEnvironmentName);

  function closeEnvDropdown() {
    isEnvDropdownOpen = false;
  }

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

{#snippet envDropdown(triggeredBy: string, isOpen: boolean | undefined, onClose: () => void)}
  <Dropdown
    {triggeredBy}
    {isOpen}
    class="z-50 w-max max-w-72 min-w-40"
    triggerDelay={0}
    onclose={onClose}
  >
    {#each environments as environment (environment.id)}
      <DropdownItem
        class="text-gray-900 dark:text-white"
        onclick={() => {
          void environmentStore.selectEnvironment(environment.name);
          onClose();
        }}
      >
        <div class="flex items-center gap-2">
          <div class="flex flex-1 items-center gap-2">
            <input
              type="radio"
              name="active-environment"
              checked={selectedEnvironmentName === environment.name}
              onchange={() => {
                void environmentStore.selectEnvironment(environment.name);
                onClose();
              }}
              aria-label={`Set ${environment.name} as active environment`}
              class="relative mr-2 flex h-4 w-4 shrink-0 items-center border-gray-300 bg-gray-100 text-primary-600 dark:border-gray-600 dark:bg-gray-700"
            />
            <span>{environment.name}</span>
          </div>
          <div class="flex">
            <EditOutline
              class="h-4 w-4 shrink-0 cursor-pointer"
              onclick={toggleEnvironmentManager}
            />
          </div>
        </div>
      </DropdownItem>
    {/each}
    <!-- <DropdownItem
      class="text-gray-900 dark:text-white"
      onclick={() => {
        onClose();
      }}
    >
      <div class="flex items-center gap-2">add env</div>
    </DropdownItem> -->
  </Dropdown>
{/snippet}

<div class="flex h-screen flex-col overflow-hidden">
  <!-- Top Navbar -->
  <nav
    class="flex h-12 shrink-0 items-center justify-between border-b border-neutral-200 bg-white px-3 dark:border-neutral-700 dark:bg-neutral-800"
  >
    <div class="flex items-center gap-4">
      <h1 aria-label="solo" class="select-none">
        <pre
          aria-hidden="true"
          class="pointer-events-none cursor-default font-mono text-[0.35rem]/[1.3] text-primary-700 select-none dark:text-primary-500">{title}</pre>
      </h1>
    </div>
    <div class="flex items-center gap-2">
      {@render navbar_actions?.()}
      <ButtonGroup class="h-8">
        <Button size="xs" color="light" onclick={toggleEnvironmentManager}>Environments</Button>
        <Button
          size="xs"
          color="light"
          id="env-dropdown-button"
          class="w-1"
          onclick={() => (isEnvDropdownOpen = true)}
          ><AngleDownOutline class="w-3 shrink-0" /></Button
        >
      </ButtonGroup>
      <Button size="xs" color="light" onclick={toggleMainConfiguration}>Settings</Button>

      {@render envDropdown("#env-dropdown-button", isEnvDropdownOpen, closeEnvDropdown)}
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
