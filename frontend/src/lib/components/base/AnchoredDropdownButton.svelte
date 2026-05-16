<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import ChevronDownOutline from "flowbite-svelte-icons/ChevronDownOutline.svelte";
  import Button from "flowbite-svelte/Button.svelte";
  import type { Snippet } from "svelte";
  import { tick } from "svelte";

  type AriaHaspopup = "dialog" | "menu" | "grid" | "listbox" | "tree" | true | false;

  interface Props {
    id: string;
    open?: boolean;
    disabled?: boolean;
    class?: string;
    triggerText?: string;
    triggerTextClass?: string;
    triggerClass?: string;
    panelClass?: string;
    ariaHaspopup?: AriaHaspopup;
    matchTriggerWidth?: boolean;
    children: Snippet;
    trigger?: Snippet;
    onopen?: () => void | Promise<void>;
    onclose?: () => void;
  }

  let {
    id,
    open = $bindable(false),
    disabled = false,
    class: className = "",
    triggerText = "",
    triggerTextClass = "truncate text-sm",
    triggerClass = "w-full justify-between px-3 py-2 text-left",
    panelClass = "p-1",
    ariaHaspopup = "listbox",
    matchTriggerWidth = true,
    children,
    trigger,
    onopen = async () => {},
    onclose = () => {}
  }: Props = $props();

  const PANEL_GAP = 8;

  let triggerWidth = $state(0);
  let viewportHeight = $state(0);
  let boundaryTop = $state(0);
  let boundaryBottom = $state(0);
  let rootTop = $state(0);
  let rootBottom = $state(0);
  let panelRenderedHeight = $state(0);
  let rootElement: HTMLDivElement | undefined = $state();
  let panelElement: HTMLDivElement | undefined = $state();

  const availableSpaceAbove = $derived(Math.max(0, rootTop - boundaryTop - PANEL_GAP));
  const availableSpaceBelow = $derived(Math.max(0, boundaryBottom - rootBottom - PANEL_GAP));
  const preferredPanelHeight = $derived(Math.max(panelRenderedHeight, 1));
  const opensUpward = $derived(
    availableSpaceBelow < preferredPanelHeight && availableSpaceAbove > availableSpaceBelow
  );
  const availablePanelSpace = $derived(opensUpward ? availableSpaceAbove : availableSpaceBelow);
  const shouldConstrainPanelHeight = $derived(
    availablePanelSpace > 0 && availablePanelSpace < preferredPanelHeight
  );
  const panelStyle = $derived.by(() => {
    const styles: string[] = [];

    if (shouldConstrainPanelHeight) {
      styles.push(`max-height: ${availablePanelSpace}px;`);
    }

    if (matchTriggerWidth) {
      styles.push(`width: ${Math.max(triggerWidth, 1)}px;`);
    }

    return styles.join(" ");
  });

  function findBoundaryElement(element: HTMLElement | undefined): HTMLElement | null {
    let current = element?.parentElement;

    while (current) {
      const style = getComputedStyle(current);
      const overflowY = style.overflowY;
      const isScrollable = ["auto", "scroll", "overlay", "hidden", "clip"].includes(overflowY);

      if (isScrollable && current !== document.body && current !== document.documentElement) {
        return current;
      }

      current = current.parentElement;
    }

    return null;
  }

  function measureAnchor() {
    if (!rootElement) return;

    const boundaryElement = findBoundaryElement(rootElement);
    const rect = rootElement.getBoundingClientRect();

    rootTop = rect.top;
    rootBottom = rect.bottom;

    if (boundaryElement) {
      const boundaryRect = boundaryElement.getBoundingClientRect();
      boundaryTop = boundaryRect.top;
      boundaryBottom = boundaryRect.bottom;
      return;
    }

    boundaryTop = 0;
    boundaryBottom = viewportHeight;
  }

  function measurePanel() {
    if (!panelElement) return;
    panelRenderedHeight = panelElement.getBoundingClientRect().height;
  }

  async function setOpen(nextOpen: boolean) {
    if (open === nextOpen) return;

    open = nextOpen;
    if (!open) {
      onclose();
      return;
    }

    measureAnchor();
    await tick();
    measurePanel();
    await onopen();
    await tick();
    measurePanel();
    measureAnchor();
  }

  async function toggleOpen() {
    await setOpen(!open);
  }

  function closeDropdown() {
    void setOpen(false);
  }

  function handleDocumentMouseDown(event: MouseEvent) {
    if (!open) return;
    const target = event.target;
    if (!(target instanceof Node)) return;
    if (rootElement?.contains(target)) return;
    closeDropdown();
  }

  function handleWindowKeydown(event: KeyboardEvent) {
    if (!open || event.key !== "Escape") return;
    closeDropdown();
  }

  function handleDocumentScroll() {
    if (!open) return;
    measureAnchor();
    measurePanel();
  }

  function handleTriggerKeydown(event: KeyboardEvent) {
    if (["ArrowDown", "ArrowUp", "Enter", " "].includes(event.key)) {
      event.preventDefault();
      void setOpen(true);
    }
  }
</script>

<svelte:document
  onmousedown={handleDocumentMouseDown}
  onkeydown={handleWindowKeydown}
  onscroll={handleDocumentScroll}
/>
<svelte:window
  bind:innerHeight={viewportHeight}
  onresize={() => {
    measureAnchor();
    measurePanel();
  }}
/>

<div class={["relative w-full", className]} bind:this={rootElement} bind:offsetWidth={triggerWidth}>
  <Button
    {id}
    color="light"
    class={triggerClass}
    {disabled}
    aria-haspopup={ariaHaspopup}
    aria-expanded={open}
    aria-controls={`${id}-panel`}
    onclick={() => void toggleOpen()}
    onkeydown={handleTriggerKeydown}
  >
    {#if trigger}
      {@render trigger()}
    {:else}
      <span class={triggerTextClass}>{triggerText}</span>
      <ChevronDownOutline class="ms-2 h-4 w-4 shrink-0" />
    {/if}
  </Button>

  {#if open}
    <div
      id={`${id}-panel`}
      bind:this={panelElement}
      class={`absolute left-0 z-30 flex rounded-lg border border-neutral-200 bg-white shadow-lg dark:border-neutral-700 dark:bg-neutral-900 ${panelClass} ${opensUpward ? "bottom-full mb-1" : "top-full mt-1"}`}
      style={panelStyle}
    >
      {@render children()}
    </div>
  {/if}
</div>
