<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import { environmentStore, environmentStoreState } from "$src/lib/stores/environmentStore.svelte";
  import { COLLECTION_OUTLINE_BUTTON_CLASSES } from "$src/lib/utils/constants";
  import type { environment } from "$wails/go/models";
  import DotsHorizontalOutline from "flowbite-svelte-icons/DotsHorizontalOutline.svelte";
  import Input from "flowbite-svelte/Input.svelte";
  import { tick } from "svelte";

  interface Props {
    env: environment.Environment;
    isActive?: boolean;
    isFocused?: boolean;
    isMenuOpen?: boolean;
    onOpen?: (name: string) => void;
    onActivate?: (name: string) => void;
    onOpenMenu?: (name: string, event: MouseEvent) => void;
  }

  let {
    env,
    isActive = false,
    isFocused = false,
    isMenuOpen = false,
    onOpen,
    onActivate,
    onOpenMenu
  }: Props = $props();

  let editingNameInputEl: HTMLInputElement | undefined = $state();
  let isEditing = $state(false);
  let editingName = $state("");
  let renameEnvironmentName = $derived(environmentStoreState.renameEnvironmentName);

  function openEnvironment() {
    onActivate?.(env.name);
    onOpen?.(env.name);
  }

  function activateEnvironment() {
    onActivate?.(env.name);
  }

  function openMenu(event: MouseEvent) {
    onOpenMenu?.(env.name, event);
  }

  async function beginRename() {
    isEditing = true;
    editingName = env.name || "";

    await tick();
    editingNameInputEl?.focus();
    editingNameInputEl?.select();
  }

  function cancelRename() {
    isEditing = false;
    editingName = "";
  }

  async function commitRename() {
    if (!isEditing) return;

    const nextName = editingName.trim();
    if (!nextName || nextName === env.name) {
      cancelRename();
      return;
    }

    try {
      await environmentStore.renameEnvironment(env.name, nextName);
    } catch {
      // error already shown by store
    } finally {
      cancelRename();
    }
  }

  function getProviderIconPath(provider: string) {
    switch (provider) {
      case "github":
        return "M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.003-.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z";
      case "gitlab":
        return "M12 1L9 11h6L12 1zm0 0L3 11l9 12 9-12-9-10z";
      default:
        return "M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5";
    }
  }

  $effect(() => {
    if (renameEnvironmentName !== env.name) {
      return;
    }

    environmentStore.consumeRenameEnvironment();
    void beginRename();
  });
</script>

<div
  class="rounded-lg border border-neutral-200 bg-neutral-50 dark:border-neutral-700 dark:bg-neutral-800/40"
>
  <div
    class={`group relative flex items-center gap-2 px-2 py-2 ${isFocused ? "bg-primary-50/60 ring-1 ring-primary-300/70 ring-inset dark:bg-primary-900/10 dark:ring-primary-700/70" : "hover:bg-neutral-100 dark:hover:bg-neutral-700/60"}`}
    role="button"
    tabindex="0"
    aria-label={`${env.name} environment row`}
    onclick={openEnvironment}
    onkeypress={(event) => {
      if (event.key === "Enter") {
        openEnvironment();
      }
    }}
    oncontextmenu={openMenu}
  >
    <div class="flex h-6 w-4 items-center justify-center">
      <input
        type="radio"
        name="active-environment"
        checked={isActive}
        onchange={activateEnvironment}
        onclick={(event) => event.stopPropagation()}
        aria-label={`Set ${env.name} as active environment`}
        class="relative h-4 w-4 shrink-0 border-gray-300 bg-gray-100 text-primary-600 dark:border-gray-600 dark:bg-gray-700"
      />
    </div>

    <div class="min-w-0 flex-1">
      {#if isEditing}
        <div class="pr-2">
          <Input
            type="text"
            size="sm"
            class="min-w-0 flex-1"
            bind:value={editingName}
            bind:elementRef={editingNameInputEl}
            autofocus
            onclick={(event) => event.stopPropagation()}
            onmousedown={(event) => event.stopPropagation()}
            onkeydown={(event) => {
              event.stopPropagation();
              if (event.key === "Enter") {
                void commitRename();
              }
              if (event.key === "Escape") {
                cancelRename();
              }
            }}
            onblur={() => void commitRename()}
          />
        </div>
      {:else}
        <div class="flex items-center gap-2">
          {#if env.gitRemote}
            <svg
              width="12"
              height="12"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              class="text-neutral-500 dark:text-neutral-400"
              aria-label={`Git remote: ${env.gitRemote}`}
            >
              <path d={getProviderIconPath(env.gitProvider || "git")} />
            </svg>
          {/if}
          <span class="truncate text-sm font-medium text-neutral-800 dark:text-neutral-100">
            {env.name}
          </span>
        </div>
      {/if}
    </div>

    <button
      type="button"
      class="visible ml-1 h-6 {COLLECTION_OUTLINE_BUTTON_CLASSES}"
      onclick={openMenu}
      title="More actions"
      aria-label="More actions"
      aria-expanded={isMenuOpen}
    >
      <DotsHorizontalOutline class="h-4 w-4" />
    </button>
  </div>
</div>
