<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: GPL-3.0-only
-->

<script lang="ts">
  import Button from "flowbite-svelte/Button.svelte";
  import Checkbox from "flowbite-svelte/Checkbox.svelte";
  import Input from "flowbite-svelte/Input.svelte";
  import TokenInput from "$src/lib/components/RequestBuilder/TokenInput.svelte";
  import { selectedEnvironment } from "$src/lib/stores/environmentStore";

  interface Props {
    headers: Header[];
    onChange?: () => void;
  }

  let { headers = $bindable(), onChange }: Props = $props();

  type Header = {
    id: string;
    key: string;
    value: string;
    enabled: boolean;
  };

  function toggleHeader(id: string) {
    headers = headers.map((h) => (h.id === id ? { ...h, enabled: !h.enabled } : h));
    onChange?.();
  }

  function removeHeader(id: string) {
    headers = headers.filter((h) => h.id !== id);
    onChange?.();
  }

  function addHeader() {
    const newHeader: Header = {
      id: crypto.randomUUID(),
      key: "",
      value: "",
      enabled: true
    };
    headers = [...headers, newHeader];
    onChange?.();
  }

  let environmentEntries = $derived(
    Object.entries($selectedEnvironment?.values ?? {}).map(([key, val]) => ({
      key,
      value: String(val?.value ?? "")
    }))
  );
</script>

<div class="min-h-0 flex-1 space-y-2 overflow-y-auto p-3">
  {#each headers as header (header.id)}
    <div class="flex flex-nowrap items-center gap-2">
      <div class="shrink-0">
        <Checkbox
          checked={header.enabled}
          onchange={() => toggleHeader(header.id)}
          aria-label={`Enable header ${header.key || "row"}`}
        />
      </div>
      <div class="min-w-0 flex-1">
        <Input
          type="text"
          size="sm"
          placeholder="Header name"
          bind:value={header.key}
          disabled={!header.enabled}
          oninput={() => onChange?.()}
        />
      </div>
      <div class="min-w-0 flex-1">
        <TokenInput
          bind:value={header.value}
          placeholder="Value"
          disabled={!header.enabled}
          {environmentEntries}
          wrapperClass="w-full"
          size="sm"
          onChange={() => onChange?.()}
        />
      </div>
      <Button
        color="light"
        size="xs"
        class="shrink-0"
        onclick={() => removeHeader(header.id)}
        aria-label="Remove header">×</Button
      >
    </div>
  {/each}

  <div class="pt-1">
    <Button color="light" size="sm" onclick={addHeader}>+ Add Header</Button>
  </div>
</div>
