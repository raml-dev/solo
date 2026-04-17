<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts" generics="TFormat extends string">
  import DropZone from "$src/lib/components/base/DropZone.svelte";
  import type { LocalImportFormatOption } from "$src/lib/components/imports/importTypes";
  import Label from "flowbite-svelte/Label.svelte";
  import Select from "flowbite-svelte/Select.svelte";

  interface Props {
    formats: LocalImportFormatOption<TFormat>[];
    selectedFormat: TFormat;
    onImport: (format: TFormat, droppedPath?: string) => Promise<void>;
  }

  let { formats, selectedFormat = $bindable(), onImport }: Props = $props();

  const selectedOption = $derived(formats.find((format) => format.key === selectedFormat));

  async function handleDrop(paths: string[]) {
    if (!selectedOption) return;

    if (paths.length > 0) {
      await onImport(selectedOption.key, paths[0]);
    } else {
      await onImport(selectedOption.key);
    }
  }
</script>

<div class="flex flex-col gap-4">
  <div class="flex flex-col gap-1">
    <Label for="import-format-select">Import format</Label>
    <Select id="import-format-select" bind:value={selectedFormat}>
      {#each formats as format (format.key)}
        <option value={format.key}>{format.label}</option>
      {/each}
    </Select>
  </div>

  {#if selectedOption}
    <DropZone
      title={selectedOption.dropTitle}
      subtitle={selectedOption.dropSubtitle}
      onDrop={async (event) => {
        await handleDrop(event.paths);
      }}
    >
      {#snippet icon()}
        {#if selectedOption.icon === "folder"}
          <svg
            width="44"
            height="44"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="1.4"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
            <polyline points="9 22 9 12 15 12 15 22" />
          </svg>
        {:else if selectedOption.icon === "document"}
          <svg
            width="44"
            height="44"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="1.4"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
            <polyline points="14 2 14 8 20 8" />
            <line x1="16" y1="13" x2="8" y2="13" />
            <line x1="16" y1="17" x2="8" y2="17" />
            <polyline points="10 9 9 9 8 9" />
          </svg>
        {:else}
          <svg
            width="44"
            height="44"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="1.4"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
            <polyline points="17 8 12 3 7 8" />
            <line x1="12" y1="3" x2="12" y2="15" />
          </svg>
        {/if}
      {/snippet}
    </DropZone>
  {/if}
</div>
