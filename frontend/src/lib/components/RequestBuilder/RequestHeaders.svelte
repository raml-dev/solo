<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import type { InputFormat } from "$src/lib/components/RequestBuilder/types";
  import EnvTokenInput from "$src/lib/components/RequestBuilder/EnvTokenInput.svelte";
  import { environmentStoreState } from "$src/lib/stores/environmentStore.svelte";
  import { sessionVarsStore } from "$src/lib/stores/sessionVarsStore";
  import ExclamationCircleSolid from "flowbite-svelte-icons/ExclamationCircleSolid.svelte";
  import Button from "flowbite-svelte/Button.svelte";
  import Checkbox from "flowbite-svelte/Checkbox.svelte";
  import Input from "flowbite-svelte/Input.svelte";

  interface Props {
    headers: Header[];
    body: string;
    bodyFormat: InputFormat;
    onChange?: () => void;
  }

  let { headers = $bindable(), body, bodyFormat, onChange }: Props = $props();

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

  let selectedEnvironment = $derived(
    environmentStoreState.environments.find(
      (environment) => environment.name === environmentStoreState.selectedEnvironmentName
    ) || null
  );

  function resolveTemplateTokens(value: string): string {
    if (!value) return value;

    const envValues = selectedEnvironment?.values ?? {};
    const envMap = new Map(
      Object.entries(envValues).map(([key, entry]) => [key, String(entry?.value ?? "")])
    );
    const sessionMap = new Map(Object.entries($sessionVarsStore ?? {}));

    return value.replace(/\{\{([^{}\r\n]+?)\}\}/g, (full, key: string) => {
      const normalizedKey = key.trim();
      if (sessionMap.has(normalizedKey)) return String(sessionMap.get(normalizedKey) ?? "");
      if (envMap.has(normalizedKey)) return String(envMap.get(normalizedKey) ?? "");
      return full;
    });
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
        aria-label="Remove header">×</Button
      >
    </div>
  {/each}

  <div class="pt-1">
    <Button color="light" size="sm" onclick={addHeader}>+ Add Header</Button>
  </div>
</div>
