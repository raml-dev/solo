<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import AnchoredDropdownButton from "$src/lib/components/base/AnchoredDropdownButton.svelte";
  import { getZoomPercent } from "$src/lib/stores/zoomStore.svelte";
  import CheckOutline from "flowbite-svelte-icons/CheckOutline.svelte";
  import { tick } from "svelte";

  interface Props {
    id: string;
    value: number;
    options: number[];
    class?: string;
    triggerClass?: string;
    onchange?: (value: number) => void | Promise<void>;
  }

  let {
    id,
    value,
    options,
    class: className = "",
    triggerClass = "w-full justify-between px-3 py-2 text-left",
    onchange = async () => {}
  }: Props = $props();

  let isOpen = $state(false);
  let optionElements: Array<HTMLButtonElement | undefined> = $state([]);

  const selectedValue = $derived(value.toFixed(2));
  const selectedLabel = $derived(`${getZoomPercent(value)}%`);

  function getOptionClass(isSelected: boolean): string {
    return [
      "flex w-full items-center justify-between gap-3 rounded-lg px-3 py-2 text-left text-sm transition-colors hover:bg-neutral-100 dark:hover:bg-neutral-800",
      isSelected
        ? "bg-primary-50 text-primary-700 dark:bg-primary-900/30 dark:text-primary-200"
        : ""
    ].join(" ");
  }

  async function handleSelect(nextValue: number) {
    isOpen = false;
    await onchange(nextValue);
  }

  async function handleOpen() {
    await tick();
    const selectedIndex = options.findIndex((option) => option.toFixed(2) === selectedValue);
    if (selectedIndex < 0) return;
    optionElements[selectedIndex]?.scrollIntoView({ block: "nearest" });
  }
</script>

<AnchoredDropdownButton
  {id}
  bind:open={isOpen}
  class={className}
  triggerText={selectedLabel}
  {triggerClass}
  panelClass="h-40 overflow-y-auto"
  onopen={handleOpen}
>
  <div class="flex w-full min-w-40 flex-col gap-1">
    {#each options as option, index (option)}
      <button
        bind:this={optionElements[index]}
        type="button"
        class={getOptionClass(selectedValue === option.toFixed(2))}
        onclick={() => void handleSelect(option)}
      >
        <span>{getZoomPercent(option)}%</span>
        {#if selectedValue === option.toFixed(2)}
          <CheckOutline class="h-4 w-4 shrink-0" />
        {/if}
      </button>
    {/each}
  </div>
</AnchoredDropdownButton>
