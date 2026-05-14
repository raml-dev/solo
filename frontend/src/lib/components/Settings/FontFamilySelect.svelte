<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import CheckOutline from "flowbite-svelte-icons/CheckOutline.svelte";
  import ChevronDownOutline from "flowbite-svelte-icons/ChevronDownOutline.svelte";
  import Button from "flowbite-svelte/Button.svelte";
  import Input from "flowbite-svelte/Input.svelte";
  import { tick } from "svelte";

  interface Props {
    id: string;
    value?: string;
    families: string[];
    placeholder: string;
    searchPlaceholder: string;
    previewKind: "sans" | "mono";
    disabled?: boolean;
    onchange?: () => void;
  }

  const LIST_ROW_HEIGHT = 36;
  const LIST_OVERSCAN = 5;
  const PANEL_GAP = 8;
  const PREFERRED_PANEL_HEIGHT = 360;

  let {
    id,
    value = $bindable(""),
    families,
    placeholder,
    searchPlaceholder,
    previewKind,
    disabled = false,
    onchange = () => {}
  }: Props = $props();

  let isOpen = $state(false);
  let filter = $state("");
  let activeIndex = $state(0);
  let triggerWidth = $state(0);
  let viewportHeight = $state(0);
  let rootTop = $state(0);
  let rootBottom = $state(0);
  let listScrollTop = $state(0);
  let listViewportHeight = $state(0);
  let rootElement: HTMLDivElement | undefined = $state();
  let filterInput: HTMLInputElement | undefined = $state();
  let listElement: HTMLDivElement | undefined = $state();

  const visibleFamilies = $derived.by(() => {
    const query = filter.trim().toLowerCase();
    if (!query) return families;
    return families.filter((family) => family.toLowerCase().includes(query));
  });

  const selectedLabel = $derived(value || "Default");
  const availableSpaceAbove = $derived(Math.max(0, rootTop - PANEL_GAP));
  const availableSpaceBelow = $derived(Math.max(0, viewportHeight - rootBottom - PANEL_GAP));
  const opensUpward = $derived(
    availableSpaceBelow < PREFERRED_PANEL_HEIGHT && availableSpaceAbove > availableSpaceBelow
  );
  const panelMaxHeight = $derived(
    Math.max(
      180,
      Math.min(PREFERRED_PANEL_HEIGHT, opensUpward ? availableSpaceAbove : availableSpaceBelow)
    )
  );
  const visibleRowCount = $derived(
    Math.max(1, Math.ceil(Math.max(listViewportHeight, LIST_ROW_HEIGHT) / LIST_ROW_HEIGHT))
  );
  const visibleStartIndex = $derived(Math.floor(listScrollTop / LIST_ROW_HEIGHT));
  const renderStartIndex = $derived(Math.max(0, visibleStartIndex - LIST_OVERSCAN));
  const renderEndIndex = $derived(
    Math.min(visibleFamilies.length, visibleStartIndex + visibleRowCount + LIST_OVERSCAN)
  );
  const renderedFamilies = $derived(visibleFamilies.slice(renderStartIndex, renderEndIndex));
  const topSpacerPx = $derived(renderStartIndex * LIST_ROW_HEIGHT);
  const bottomSpacerPx = $derived((visibleFamilies.length - renderEndIndex) * LIST_ROW_HEIGHT);

  function quoteCssFontFamily(family: string) {
    return `"${family.replaceAll("\\", "\\\\").replaceAll('"', '\\"')}"`;
  }

  function getFontPreviewStyle(family: string) {
    const fallback =
      previewKind === "sans"
        ? "var(--font-sans), ui-sans-serif, system-ui, sans-serif"
        : "var(--font-mono), ui-monospace, SFMono-Regular, monospace";

    if (!family) return fallback;
    return `${quoteCssFontFamily(family)}, ${fallback}`;
  }

  function getOptionClass(isActive: boolean) {
    return [
      "flex h-9 w-full items-center justify-between gap-3 px-3 text-left text-sm hover:bg-neutral-100 dark:hover:bg-neutral-800",
      isActive ? "bg-primary-50 text-primary-700 dark:bg-primary-900/30 dark:text-primary-200" : ""
    ].join(" ");
  }

  function getSelectedIndex() {
    if (!value) return 0;
    const index = visibleFamilies.findIndex((family) => family === value);
    return index >= 0 ? index + 1 : 0;
  }

  function measureAnchor() {
    if (!rootElement) return;
    const rect = rootElement.getBoundingClientRect();
    rootTop = rect.top;
    rootBottom = rect.bottom;
  }

  function resetSearch() {
    filter = "";
    activeIndex = getSelectedIndex();
    listScrollTop = 0;
    if (listElement) {
      listElement.scrollTop = 0;
    }
  }

  async function scrollActiveOptionIntoView() {
    await tick();
    if (!listElement || activeIndex === 0) {
      if (listElement && activeIndex === 0) {
        listElement.scrollTop = 0;
        listScrollTop = 0;
      }
      return;
    }

    const optionTop = (activeIndex - 1) * LIST_ROW_HEIGHT;
    const optionBottom = optionTop + LIST_ROW_HEIGHT;
    const viewportTop = listElement.scrollTop;
    const viewportBottom = viewportTop + listViewportHeight;

    if (optionTop < viewportTop) {
      listElement.scrollTop = optionTop;
      listScrollTop = optionTop;
      return;
    }

    if (optionBottom > viewportBottom) {
      const nextScrollTop = optionBottom - listViewportHeight;
      listElement.scrollTop = nextScrollTop;
      listScrollTop = nextScrollTop;
    }
  }

  async function focusFilter() {
    await tick();
    filterInput?.focus();
    filterInput?.select();
    await scrollActiveOptionIntoView();
  }

  async function setOpen(open: boolean) {
    isOpen = open;
    if (!open) {
      resetSearch();
      return;
    }

    measureAnchor();
    activeIndex = getSelectedIndex();
    await focusFilter();
  }

  async function setActiveIndex(nextIndex: number) {
    activeIndex = Math.max(0, Math.min(visibleFamilies.length, nextIndex));
    await scrollActiveOptionIntoView();
  }

  function commitSelection(nextValue: string) {
    value = nextValue;
    onchange();
    isOpen = false;
    resetSearch();
  }

  async function handleFilterInput() {
    activeIndex = getSelectedIndex();
    listScrollTop = 0;
    if (listElement) {
      listElement.scrollTop = 0;
    }
    await scrollActiveOptionIntoView();
  }

  function closeDropdown() {
    isOpen = false;
    resetSearch();
  }

  async function handleKeydown(event: KeyboardEvent) {
    if (!isOpen && ["ArrowDown", "ArrowUp", "Enter", " "].includes(event.key)) {
      event.preventDefault();
      await setOpen(true);
      return;
    }

    if (!isOpen) return;

    switch (event.key) {
      case "ArrowDown":
        event.preventDefault();
        await setActiveIndex(activeIndex + 1);
        break;
      case "ArrowUp":
        event.preventDefault();
        await setActiveIndex(activeIndex - 1);
        break;
      case "Enter":
        event.preventDefault();
        if (activeIndex === 0) {
          commitSelection("");
          return;
        }
        commitSelection(visibleFamilies[activeIndex - 1] || "");
        break;
      case "Escape":
        closeDropdown();
        break;
      default:
        break;
    }
  }

  function handleDocumentMouseDown(event: MouseEvent) {
    if (!isOpen) return;
    const target = event.target;
    if (!(target instanceof Node)) return;
    if (rootElement?.contains(target)) return;
    closeDropdown();
  }

  function handleListScroll(event: Event) {
    const currentTarget = event.currentTarget;
    if (!(currentTarget instanceof HTMLDivElement)) return;
    listScrollTop = currentTarget.scrollTop;
  }
</script>

<svelte:document onmousedown={handleDocumentMouseDown} />
<svelte:window bind:innerHeight={viewportHeight} onresize={measureAnchor} />

<div class="relative w-full" bind:this={rootElement} bind:offsetWidth={triggerWidth}>
  <Button
    {id}
    color="light"
    class="relative w-full justify-between px-3 py-2 text-left"
    {disabled}
    aria-haspopup="listbox"
    aria-expanded={isOpen}
    aria-controls={`${id}-listbox`}
    onclick={() => void setOpen(!isOpen)}
    onkeydown={(event: KeyboardEvent) => void handleKeydown(event)}
  >
    <span class="truncate text-sm" style:font-family={getFontPreviewStyle(value)}>
      {selectedLabel}
    </span>
    <ChevronDownOutline class="ms-2 h-4 w-4 shrink-0" />
  </Button>

  {#if isOpen}
    <div
      class={`absolute left-0 z-30 flex rounded-lg border border-neutral-200 bg-white p-3 shadow-lg dark:border-neutral-700 dark:bg-neutral-900 ${opensUpward ? "bottom-full mb-1" : "top-full mt-1"}`}
      style={`width: ${Math.max(triggerWidth, 1)}px; max-height: ${panelMaxHeight}px;`}
    >
      <div class="flex min-h-0 flex-1 flex-col gap-3">
        <div class="flex items-center justify-between gap-2">
          <Button size="xs" color="light" onclick={() => commitSelection("")}>Default</Button>
        </div>

        <Input
          bind:elementRef={filterInput}
          bind:value={filter}
          size="sm"
          clearable
          placeholder={searchPlaceholder}
          oninput={() => void handleFilterInput()}
          onkeydown={(event) => void handleKeydown(event)}
        />

        <div
          class="flex min-h-0 flex-1 flex-col overflow-hidden rounded-lg border border-neutral-200 dark:border-neutral-700"
        >
          <button
            type="button"
            class={`${getOptionClass(activeIndex === 0)} border-b border-neutral-200 dark:border-neutral-700`}
            role="option"
            aria-selected={value === ""}
            onclick={() => commitSelection("")}
            onpointerenter={() => (activeIndex = 0)}
          >
            <span class="truncate">{placeholder}</span>
            {#if value === ""}
              <CheckOutline class="h-4 w-4 shrink-0" />
            {/if}
          </button>

          <div
            id={`${id}-listbox`}
            bind:this={listElement}
            bind:clientHeight={listViewportHeight}
            class="min-h-0 flex-1 overflow-y-auto"
            role="listbox"
            aria-label={placeholder}
            onscroll={handleListScroll}
          >
            {#if visibleFamilies.length === 0}
              <p class="px-3 py-2 text-sm text-neutral-500 dark:text-neutral-400">
                No fonts match your search.
              </p>
            {:else}
              <div style={`height: ${topSpacerPx}px;`}></div>

              {#each renderedFamilies as family, localIndex (family)}
                {@const actualIndex = renderStartIndex + localIndex}
                <button
                  id={`${id}-option-${actualIndex + 1}`}
                  type="button"
                  class={getOptionClass(activeIndex === actualIndex + 1)}
                  role="option"
                  aria-selected={value === family}
                  style:font-family={getFontPreviewStyle(family)}
                  onclick={() => commitSelection(family)}
                  onpointerenter={() => (activeIndex = actualIndex + 1)}
                >
                  <span class="truncate">{family}</span>
                  {#if value === family}
                    <CheckOutline class="h-4 w-4 shrink-0" />
                  {/if}
                </button>
              {/each}

              <div style={`height: ${bottomSpacerPx}px;`}></div>
            {/if}
          </div>
        </div>
      </div>
    </div>
  {/if}
</div>
