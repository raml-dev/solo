<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import Dropdown from "flowbite-svelte/Dropdown.svelte";
  import { onMount } from "svelte";

  const CONTEXT_MENU_OPEN_EVENT = "solo:context-menu-open";

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

  const menuId = `context-menu-${crypto.randomUUID()}`;
  let menuOpen = $state(false);

  $effect(() => {
    if (isOpen === undefined) {
      return;
    }

    menuOpen = isOpen;
  });

  $effect(() => {
    if (!menuOpen || typeof document === "undefined") {
      return;
    }

    document.dispatchEvent(
      new CustomEvent(CONTEXT_MENU_OPEN_EVENT, {
        detail: { menuId }
      })
    );
  });

  function closeMenu() {
    if (!menuOpen) {
      return;
    }

    menuOpen = false;
    onClose();
  }

  onMount(() => {
    function handleContextMenuOpen(event: Event) {
      const customEvent = event as CustomEvent<{ menuId: string }>;

      if (customEvent.detail.menuId === menuId || !menuOpen) {
        return;
      }

      closeMenu();
    }

    document.addEventListener(CONTEXT_MENU_OPEN_EVENT, handleContextMenuOpen);

    return () => {
      document.removeEventListener(CONTEXT_MENU_OPEN_EVENT, handleContextMenuOpen);
    };
  });
</script>

<Dropdown
  {triggeredBy}
  bind:isOpen={menuOpen}
  class={menuClass}
  triggerDelay={10}
  onclose={closeMenu}
>
  {@render children?.()}
</Dropdown>
