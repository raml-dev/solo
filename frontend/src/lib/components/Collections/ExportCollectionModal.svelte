<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import Button from "flowbite-svelte/Button.svelte";
  import Modal from "flowbite-svelte/Modal.svelte";

  type ExportFormat = "solo" | "openapi";

  interface FormatOption {
    id: ExportFormat;
    label: string;
    description: string;
  }

  interface Props {
    open: boolean;
    onExport: (format: ExportFormat) => void;
    onClose: () => void;
  }

  const FORMAT_OPTIONS: FormatOption[] = [
    {
      id: "solo",
      label: "Solo (JSON)",
      description: "Native Solo format, preserves all collection data"
    },
    {
      id: "openapi",
      label: "OpenAPI 3.1 (YAML)",
      description: "Standard OpenAPI format, compatible with other tools"
    }
  ];

  let { open = $bindable(), onExport, onClose }: Props = $props();

  let selectedFormat = $state<ExportFormat>("solo");
</script>

<Modal
  bind:open
  title="Export Collection"
  size="sm"
  onclose={onClose}
>
  <div class="flex flex-col gap-3">
    {#each FORMAT_OPTIONS as option (option.id)}
      <button
        type="button"
        class="flex cursor-pointer items-start gap-3 rounded-lg border p-3 text-left transition-colors
          {selectedFormat === option.id
          ? 'border-primary-500 bg-primary-50 dark:border-primary-400 dark:bg-primary-900/20'
          : 'border-gray-200 hover:border-gray-300 hover:bg-gray-50 dark:border-gray-600 dark:hover:border-gray-500 dark:hover:bg-gray-700/30'}"
        onclick={() => (selectedFormat = option.id)}
      >
        <div class="mt-0.5 flex h-4 w-4 shrink-0 items-center justify-center rounded-full border-2
          {selectedFormat === option.id
          ? 'border-primary-500 dark:border-primary-400'
          : 'border-gray-400 dark:border-gray-500'}">
          {#if selectedFormat === option.id}
            <div class="h-2 w-2 rounded-full bg-primary-500 dark:bg-primary-400"></div>
          {/if}
        </div>
        <div>
          <p class="text-sm font-medium text-gray-900 dark:text-white">{option.label}</p>
          <p class="text-xs text-gray-500 dark:text-gray-400">{option.description}</p>
        </div>
      </button>
    {/each}
  </div>

  {#snippet footer()}
    <div class="flex w-full justify-end gap-2">
      <Button color="alternative" onclick={onClose}>Cancel</Button>
      <Button color="primary" onclick={() => onExport(selectedFormat)}>Export</Button>
    </div>
  {/snippet}
</Modal>
