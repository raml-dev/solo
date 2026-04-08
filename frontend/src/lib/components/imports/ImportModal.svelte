<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: GPL-3.0-only
-->

<script lang="ts">
  import VerticalSectionLayout from "$src/lib/components/layouts/VerticalSectionLayout.svelte";
  import Button from "flowbite-svelte/Button.svelte";
  import Modal from "flowbite-svelte/Modal.svelte";

  type ImportSection = "local" | "curl" | "git";

  interface GitActionState {
    loading: boolean;
    disabled: boolean;
    submit: () => void;
  }

  interface Props {
    title: string;
    open: boolean;
    showCurlSection?: boolean;

    localActionLabel?: string;
    localActionDisabled?: boolean;
    onLocalAction?: () => void;

    curlActionLabel?: string;
    curlActionDisabled?: boolean;
    onCurlAction?: () => void;

    gitActionState?: GitActionState | null;

    localContent?: import("svelte").Snippet;
    curlContent?: import("svelte").Snippet;
    gitContent?: import("svelte").Snippet;

    onClose?: (event?: Event) => void;
  }

  let {
    title,
    open = $bindable(),
    showCurlSection = false,
    localActionLabel = "Select file...",
    localActionDisabled = false,
    onLocalAction,
    curlActionLabel = "Import Request",
    curlActionDisabled = false,
    onCurlAction,
    gitActionState = null,
    localContent,
    curlContent,
    gitContent,
    onClose
  }: Props = $props();

  let selectedSection = $state<ImportSection>("local");

  let activeSection = $derived(
    !showCurlSection && selectedSection === "curl" ? "local" : selectedSection
  );

  const sectionItems = $derived.by(() => {
    const items: { id: ImportSection; label: string }[] = [{ id: "local", label: "Local" }];

    if (showCurlSection) {
      items.push({ id: "curl", label: "cURL" });
    }

    items.push({ id: "git", label: "Git" });

    return items;
  });
</script>

<Modal
  bind:open
  {title}
  classes={{ body: "h-[400px] overflow-hidden p-4" }}
  size="xl"
  onclose={(event) => {
    selectedSection = "local";
    onClose?.(event);
  }}
>
  <VerticalSectionLayout
    items={sectionItems}
    activeItem={activeSection}
    onChange={(next) => {
      selectedSection = next as ImportSection;
    }}
  >
    {#snippet children(section: string)}
      {#if section === "local"}
        {@render localContent?.()}
      {:else if section === "curl"}
        {@render curlContent?.()}
      {:else if section === "git"}
        {@render gitContent?.()}
      {/if}
    {/snippet}
  </VerticalSectionLayout>

  {#snippet footer()}
    <div class="flex w-full justify-end gap-2">
      {#if activeSection === "local"}
        <Button color="primary" disabled={localActionDisabled} onclick={onLocalAction}>
          {localActionLabel}
        </Button>
      {:else if activeSection === "curl"}
        <Button color="primary" disabled={curlActionDisabled} onclick={onCurlAction}>
          {curlActionLabel}
        </Button>
      {:else if activeSection === "git"}
        <Button
          color="primary"
          loading={gitActionState?.loading ?? false}
          disabled={gitActionState?.disabled ?? true}
          onclick={() => gitActionState?.submit()}
        >
          Import from Git
        </Button>
      {/if}
    </div>
  {/snippet}
</Modal>
