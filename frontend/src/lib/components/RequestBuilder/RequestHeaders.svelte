<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import type { InputFormat } from "$src/lib/components/RequestBuilder/types";
  import EnvTokenInput from "$src/lib/components/RequestBuilder/EnvTokenInput.svelte";
  import {
    createResolvedVariableEntryMap,
    resolveVariableTokens,
    type ResolvedVariableEntry
  } from "$src/lib/utils/variableResolution";
  import ExclamationCircleSolid from "flowbite-svelte-icons/ExclamationCircleSolid.svelte";
  import PlusOutline from "flowbite-svelte-icons/PlusOutline.svelte";
  import TrashBinOutline from "flowbite-svelte-icons/TrashBinOutline.svelte";
  import Button from "flowbite-svelte/Button.svelte";
  import Checkbox from "flowbite-svelte/Checkbox.svelte";
  import Input from "flowbite-svelte/Input.svelte";

  interface Props {
    headers: Header[];
    body: string;
    bodyFormat: InputFormat;
    variableEntries?: ResolvedVariableEntry[];
    onChange?: () => void;
  }

  let { headers = $bindable(), body, bodyFormat, variableEntries = [], onChange }: Props = $props();

  type Header = {
    id: string;
    key: string;
    value: string;
    enabled: boolean;
    autoInjectedContentType?: boolean;
    injectedContentTypeValue?: string;
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

  function isContentTypeHeader(header: Header): boolean {
    return header.key.trim().toLowerCase() === "content-type";
  }

  let variableEntryMap = $derived(createResolvedVariableEntryMap(variableEntries));

  function resolveTemplateTokens(value: string): string {
    return resolveVariableTokens(value, variableEntryMap);
  }

  function getExpectedContentTypeValue(): string | null {
    if (bodyFormat === "none") return null;
    if (!body.trim()) return null;
    if (bodyFormat === "json") return "application/json";
    if (bodyFormat === "xml") return "application/xml";
    return null;
  }

  function shouldShowContentTypeWarning(header: Header): boolean {
    if (!header.enabled) return false;

    const resolvedValue = resolveTemplateTokens(header.value).trim().toLowerCase();
    const expectedValue = getExpectedContentTypeValue();

    if (!expectedValue) {
      return resolvedValue.length > 0;
    }

    return resolvedValue !== expectedValue;
  }

  function getContentTypeWarningMessage(): string {
    const hasBody = bodyFormat !== "none" && body.trim().length > 0;
    if (!hasBody) return "There is no Body in this request";
    return "The Content-Type value is different from the selected Body type";
  }
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
        {#snippet contentTypeWarning()}
          <span
            class="inline-flex"
            title={getContentTypeWarningMessage()}
            aria-label={getContentTypeWarningMessage()}
          >
            <ExclamationCircleSolid
              class="h-3.5 w-4 cursor-pointer text-warning-500 dark:text-warning-400"
            />
          </span>
        {/snippet}

        <EnvTokenInput
          bind:value={header.value}
          placeholder="Value"
          disabled={!header.enabled}
          class="w-full"
          size="sm"
          {variableEntries}
          right={contentTypeWarning}
          rightVisible={isContentTypeHeader(header) && shouldShowContentTypeWarning(header)}
          onChange={() => onChange?.()}
        />
      </div>
      <Button
        color="light"
        size="xs"
        class="shrink-0"
        onclick={() => removeHeader(header.id)}
        aria-label="Remove header"
      >
        <TrashBinOutline class="h-4 w-4" />
      </Button>
    </div>
  {/each}

  <div class="pt-1">
    <Button color="light" size="sm" onclick={addHeader}>
      <PlusOutline class="mr-1 h-4 w-4" />
      <span>Add Header</span>
    </Button>
  </div>
</div>
