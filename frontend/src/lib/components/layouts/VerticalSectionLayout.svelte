<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import Button from "flowbite-svelte/Button.svelte";

  interface VerticalSectionItem {
    id: string;
    label: string;
    disabled?: boolean;
  }

  interface Props {
    items: VerticalSectionItem[];
    activeItem: string;
    navWidthClass?: string;
    children?: import("svelte").Snippet<[string]>;
    onChange?: (next: string) => void;
  }

  let {
    items,
    activeItem = $bindable(),
    navWidthClass = "w-36",
    children,
    onChange
  }: Props = $props();

  function selectItem(itemId: string, disabled = false) {
    if (disabled) return;
    activeItem = itemId;
    onChange?.(itemId);
  }
</script>

<div class="flex h-full gap-6">
  <nav class={`flex ${navWidthClass} shrink-0 flex-col gap-1`}>
    {#each items as item (item.id)}
      <Button
        color={activeItem === item.id ? "primary" : "light"}
        class="justify-start"
        disabled={item.disabled}
        onclick={() => selectItem(item.id, item.disabled)}
      >
        {item.label}
      </Button>
    {/each}
  </nav>

  <div class="min-w-0 flex-1 overflow-y-auto">
    {@render children?.(activeItem)}
  </div>
</div>
