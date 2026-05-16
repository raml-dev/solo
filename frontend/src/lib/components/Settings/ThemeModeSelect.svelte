<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import AnchoredDropdownButton from "$src/lib/components/base/AnchoredDropdownButton.svelte";
  import type { ThemeMode } from "$src/lib/theme/themeModel";
  import CheckOutline from "flowbite-svelte-icons/CheckOutline.svelte";

  interface Props {
    id: string;
    value: ThemeMode;
    triggerClass?: string;
    onchange?: (value: ThemeMode) => void;
  }

  const OPTIONS: { value: ThemeMode; label: string }[] = [
    { value: "light", label: "Light" },
    { value: "dark", label: "Dark" },
    { value: "system", label: "System" }
  ];

  let {
    id,
    value,
    triggerClass = "w-full justify-between px-3 py-2 text-left",
    onchange = () => {}
  }: Props = $props();

  let isOpen = $state(false);

  const selectedLabel = $derived(
    OPTIONS.find((option) => option.value === value)?.label || "System"
  );

  function getOptionClass(isSelected: boolean): string {
    return [
      "flex w-full items-center justify-between gap-3 rounded-lg px-3 py-2 text-left text-sm transition-colors hover:bg-neutral-100 dark:hover:bg-neutral-800",
      isSelected
        ? "bg-primary-50 text-primary-700 dark:bg-primary-900/30 dark:text-primary-200"
        : ""
    ].join(" ");
  }

  function handleSelect(nextValue: ThemeMode) {
    isOpen = false;
    onchange(nextValue);
  }
</script>

<AnchoredDropdownButton {id} bind:open={isOpen} {triggerClass} triggerText={selectedLabel}>
  <div class="flex w-full min-w-48 flex-col gap-1">
    {#each OPTIONS as option (option.value)}
      <button
        type="button"
        class={getOptionClass(value === option.value)}
        onclick={() => handleSelect(option.value)}
      >
        <span>{option.label}</span>
        {#if value === option.value}
          <CheckOutline class="h-4 w-4 shrink-0" />
        {/if}
      </button>
    {/each}
  </div>
</AnchoredDropdownButton>
