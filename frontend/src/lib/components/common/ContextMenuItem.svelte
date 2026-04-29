<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import DropdownItem from "flowbite-svelte/DropdownItem.svelte";

  interface Props {
    className?: string;
    disabled?: boolean;
    onclick?: (event: MouseEvent) => void | Promise<void>;
    children?: import("svelte").Snippet;
  }

  let {
    className = "text-gray-900 dark:text-white",
    disabled = false,
    onclick = () => {},
    children
  }: Props = $props();

  function handleClick(event: MouseEvent) {
    event.stopPropagation();
    return onclick(event);
  }
</script>

<DropdownItem class={className} {disabled} onclick={handleClick}>
  {@render children?.()}
</DropdownItem>
