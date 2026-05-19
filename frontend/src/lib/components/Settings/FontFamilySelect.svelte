<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import AnchoredDropdownButton from "$src/lib/components/base/AnchoredDropdownButton.svelte";
  import {
    FONT_FAMILY_SELECT_LIST_OVERSCAN,
    FONT_FAMILY_SELECT_LIST_ROW_HEIGHT
  } from "$src/lib/utils/constants";
  import CheckOutline from "flowbite-svelte-icons/CheckOutline.svelte";
  import Input from "flowbite-svelte/Input.svelte";
  import { tick } from "svelte";

  type FontOption = {
    id: string;
    value: string;
    label: string;
    isDefault?: boolean;
  };

  interface Props {
    id: string;
    value?: string;
    families: string[];
    placeholder: string;
    searchPlaceholder: string;
    previewKind: "sans" | "mono";
    disabled?: boolean;
    class?: string;
    triggerClass?: string;
    onchange?: (value: string) => void;
  }

  let {
    id,
    value = "",
    families,
    placeholder,
    searchPlaceholder,
    previewKind,
    disabled = false,
    class: className = "",
    triggerClass = "w-full justify-between px-3 py-2 text-left",
    onchange = () => {}
  }: Props = $props();

  let isOpen = $state(false);
  let filter = $state("");
  let activeIndex = $state(0);
  let listScrollTop = $state(0);
  let listViewportHeight = $state(0);
  let filterInput: HTMLInputElement | undefined = $state();
  let listElement: HTMLDivElement | undefined = $state();

  const allOptions = $derived<FontOption[]>([
    {
      id: "__default__",
      value: "",
      label: placeholder,
      isDefault: true
    },
    ...families.map((family) => ({
      id: family,
      value: family,
      label: family
    }))
  ]);

  const visibleOptions = $derived.by(() => {
    const query = filter.trim().toLowerCase();
    if (!query) return allOptions;
    return allOptions.filter((option) => option.label.toLowerCase().includes(query));
  });

  const selectedLabel = $derived(value || "Default");
  const visibleRowCount = $derived(
    Math.max(
      1,
      Math.ceil(
        Math.max(listViewportHeight, FONT_FAMILY_SELECT_LIST_ROW_HEIGHT) /
          FONT_FAMILY_SELECT_LIST_ROW_HEIGHT
      )
    )
  );
  const visibleStartIndex = $derived(
    Math.floor(listScrollTop / FONT_FAMILY_SELECT_LIST_ROW_HEIGHT)
  );
  const renderStartIndex = $derived(
    Math.max(0, visibleStartIndex - FONT_FAMILY_SELECT_LIST_OVERSCAN)
  );
  const renderEndIndex = $derived(
    Math.min(
      visibleOptions.length,
      visibleStartIndex + visibleRowCount + FONT_FAMILY_SELECT_LIST_OVERSCAN
    )
  );
  const renderedOptions = $derived(visibleOptions.slice(renderStartIndex, renderEndIndex));
  const topSpacerPx = $derived(renderStartIndex * FONT_FAMILY_SELECT_LIST_ROW_HEIGHT);
  const bottomSpacerPx = $derived(
    (visibleOptions.length - renderEndIndex) * FONT_FAMILY_SELECT_LIST_ROW_HEIGHT
  );

  function quoteCssFontFamily(family: string) {
    return `"${family.replaceAll("\\", "\\\\").replaceAll('"', '\\"')}"`;
  }

  function getFontPreviewStyle(option: FontOption) {
    const fallback =
      previewKind === "sans" ? "var(--font-sans-default)" : "var(--font-mono-default)";

    if (option.isDefault || !option.value) return fallback;
    return `${quoteCssFontFamily(option.value)}, ${fallback}`;
  }

  function getOptionClass(isActive: boolean) {
    return [
      "flex h-9 w-full items-center justify-between gap-3 px-3 text-left text-sm hover:bg-neutral-100 dark:hover:bg-neutral-800",
      isActive ? "bg-primary-50 text-primary-700 dark:bg-primary-900/30 dark:text-primary-200" : ""
    ].join(" ");
  }

  function getSelectedIndex() {
    return visibleOptions.findIndex((option) => option.value === value);
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
    if (!listElement || activeIndex < 0) {
      if (listElement && activeIndex < 0) {
        listElement.scrollTop = 0;
        listScrollTop = 0;
      }
      return;
    }

    const optionTop = activeIndex * FONT_FAMILY_SELECT_LIST_ROW_HEIGHT;
    const optionBottom = optionTop + FONT_FAMILY_SELECT_LIST_ROW_HEIGHT;
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

  async function handleOpen() {
    activeIndex = getSelectedIndex();
    await focusFilter();
  }

  function handleClose() {
    resetSearch();
  }

  async function setActiveIndex(nextIndex: number) {
    activeIndex = Math.max(0, Math.min(visibleOptions.length - 1, nextIndex));
    await scrollActiveOptionIntoView();
  }

  function commitSelection(nextValue: string) {
    onchange(nextValue);
    isOpen = false;
  }

  async function handleFilterInput() {
    activeIndex = getSelectedIndex();
    listScrollTop = 0;
    if (listElement) {
      listElement.scrollTop = 0;
    }
    await scrollActiveOptionIntoView();
  }

  async function handleKeydown(event: KeyboardEvent) {
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
        if (activeIndex < 0 || activeIndex >= visibleOptions.length) return;
        commitSelection(visibleOptions[activeIndex]?.value || "");
        break;
      default:
        break;
    }
  }

  function handleListScroll(event: Event) {
    const currentTarget = event.currentTarget;
    if (!(currentTarget instanceof HTMLDivElement)) return;
    listScrollTop = currentTarget.scrollTop;
  }
</script>

<AnchoredDropdownButton
  {id}
  bind:open={isOpen}
  {disabled}
  triggerText={selectedLabel}
  {triggerClass}
  panelClass="max-h-65 p-3"
  onopen={handleOpen}
  onclose={handleClose}
  class={className}
>
  <div class="flex min-h-0 flex-1 flex-col gap-3">
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
      id={`${id}-listbox`}
      bind:this={listElement}
      bind:clientHeight={listViewportHeight}
      class="min-h-0 flex-1 overflow-y-auto rounded-lg border border-neutral-200 dark:border-neutral-700"
      role="listbox"
      aria-label={placeholder}
      onscroll={handleListScroll}
    >
      {#if visibleOptions.length === 0}
        <p class="px-3 py-2 text-sm text-neutral-500 dark:text-neutral-400">
          No fonts match your search.
        </p>
      {:else}
        <div style={`height: ${topSpacerPx}px;`}></div>

        {#each renderedOptions as option, localIndex (option.id)}
          {@const actualIndex = renderStartIndex + localIndex}
          <button
            id={`${id}-option-${actualIndex}`}
            type="button"
            class={getOptionClass(activeIndex === actualIndex)}
            role="option"
            aria-selected={value === option.value}
            style:font-family={getFontPreviewStyle(option)}
            onclick={() => commitSelection(option.value)}
            onpointerenter={() => (activeIndex = actualIndex)}
          >
            <span class="truncate">{option.label}</span>
            {#if value === option.value}
              <CheckOutline class="h-4 w-4 shrink-0" />
            {/if}
          </button>
        {/each}

        <div style={`height: ${bottomSpacerPx}px;`}></div>
      {/if}
    </div>
  </div>
</AnchoredDropdownButton>
