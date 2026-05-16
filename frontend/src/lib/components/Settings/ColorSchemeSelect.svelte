<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import AnchoredDropdownButton from "$src/lib/components/base/AnchoredDropdownButton.svelte";
  import ThemePreview from "$src/lib/components/Settings/ThemePreview.svelte";
  import type { theme } from "$wails/go/models";
  import CheckOutline from "flowbite-svelte-icons/CheckOutline.svelte";
  import { tick } from "svelte";

  interface Props {
    id: string;
    value: string;
    themes: theme.Theme[];
    triggerClass?: string;
    onchange?: (themeId: string) => void;
  }

  let {
    id,
    value,
    themes,
    triggerClass = "w-full justify-between px-3 py-2 text-left",
    onchange = () => {}
  }: Props = $props();

  let isOpen = $state(false);
  let themeButtonElements: Array<HTMLButtonElement | undefined> = $state([]);

  const selectedLabel = $derived.by(() => {
    const selectedTheme = themes.find((item) => item.id === value);
    return selectedTheme?.label || "Untitled Theme";
  });

  function formatThemeName(label: string): string {
    return label || "Untitled Theme";
  }

  function handleSelect(themeId: string) {
    isOpen = false;
    onchange(themeId);
  }

  async function handleOpen() {
    await tick();
    const selectedIndex = themes.findIndex((currentTheme) => currentTheme.id === value);
    if (selectedIndex < 0) return;
    themeButtonElements[selectedIndex]?.scrollIntoView({ block: "nearest", inline: "nearest" });
  }
</script>

<AnchoredDropdownButton
  {id}
  bind:open={isOpen}
  {triggerClass}
  triggerText={formatThemeName(selectedLabel)}
  ariaHaspopup="grid"
  matchTriggerWidth={true}
  panelClass="p-3 max-h-96 overflow-auto"
  onopen={handleOpen}
>
  <div class="grid w-full grid-cols-3 gap-2" role="grid" aria-label="Color schemes">
    {#each themes as currentTheme, index (currentTheme.id)}
      <button
        bind:this={themeButtonElements[index]}
        type="button"
        class={`flex flex-col gap-1 rounded-lg border p-2 text-left transition-colors hover:border-primary-400 hover:bg-neutral-50 dark:hover:bg-neutral-800 ${
          value === currentTheme.id
            ? "border-primary-500 bg-primary-50 dark:bg-primary-900/20"
            : "border-neutral-200 dark:border-neutral-700"
        }`}
        onclick={() => handleSelect(currentTheme.id)}
      >
        <div class="overflow-hidden rounded-md border border-neutral-200 dark:border-neutral-700">
          <div class="w-full">
            <ThemePreview seeds={currentTheme.config?.seeds} />
          </div>
        </div>
        <div class="flex items-center justify-between gap-1">
          <span class="truncate text-xs font-medium text-neutral-700 dark:text-neutral-200">
            {formatThemeName(currentTheme.label)}
          </span>
          {#if value === currentTheme.id}
            <CheckOutline class="h-3.5 w-3.5 shrink-0 text-primary-600 dark:text-primary-300" />
          {/if}
        </div>
      </button>
    {/each}
  </div>
</AnchoredDropdownButton>
