<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import Dropdown from "flowbite-svelte/Dropdown.svelte";

  interface Props {
    triggeredBy: string;
    isOpen?: boolean;
    menuClass?: string;
    onClose?: () => void;
    children?: import("svelte").Snippet;
  }

  let {
    triggeredBy,
    isOpen = undefined,
    menuClass = "z-50 w-40",
    onClose = () => {},
    children
  }: Props = $props();
</script>

<Dropdown {triggeredBy} {isOpen} class={menuClass} triggerDelay={0} onclose={onClose}>
  {@render children?.()}
</Dropdown>
