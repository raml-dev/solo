<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import { collection } from "$wails/go/models";
  import Button from "flowbite-svelte/Button.svelte";
  import Input from "flowbite-svelte/Input.svelte";
  import Label from "flowbite-svelte/Label.svelte";
  import Toggle from "flowbite-svelte/Toggle.svelte";
  import EnvTokenInput from "$src/lib/components/RequestBuilder/EnvTokenInput.svelte";

  interface Props {
    auth: collection.AuthConfiguration;
    onChange?: () => void;
  }

  let { auth = $bindable(), onChange }: Props = $props();

  function handleChange() {
    onChange?.();
  }

  function handleTokenUrlChange(value: string) {
    auth.tokenUrl = value;
    onChange?.();
  }

  function handleTokenPathChange(value: string) {
    auth.tokenPath = value;
    onChange?.();
  }

  // Convert map to array for easier editing in UI
  let templateRows = $state(
    Object.entries(auth.template || {}).map(([key, value], i) => ({
      id: `row-${i}-${Date.now()}`,
      key,
      value: String(value)
    }))
  );

  function syncTemplateToAuth() {
    const newTemplate: Record<string, string> = {};
    templateRows.forEach((row) => {
      if (row.key) {
        newTemplate[row.key] = row.value;
      }
    });
    auth.template = newTemplate;
  }

  function addRow() {
    templateRows = [...templateRows, { id: `row-${Date.now()}`, key: "", value: "" }];
    syncTemplateToAuth();
    handleChange();
  }

  function removeRow(id: string) {
    templateRows = templateRows.filter((r) => r.id !== id);
    syncTemplateToAuth();
    handleChange();
  }

  function updateTemplateRowKey(id: string, value: string) {
    const row = templateRows.find((r) => r.id === id);
    if (!row) return;
    row.key = value;
    syncTemplateToAuth();
    handleChange();
  }

  function updateTemplateRowValue(id: string, value: string) {
    const row = templateRows.find((r) => r.id === id);
    if (!row) return;
    row.value = value;
    syncTemplateToAuth();
    handleChange();
  }
</script>

<div class="flex-1 space-y-6 overflow-y-auto p-4">
  <div
    class="flex items-center justify-between border-b border-neutral-100 pb-4 dark:border-neutral-800"
  >
    <div class="space-y-0.5">
      <Label class="text-base font-semibold">Enable Integrated OAuth2</Label>
      <div class="text-sm text-neutral-500 dark:text-neutral-400">
        Automatically fetch and inject Bearer tokens into your requests.
      </div>
    </div>
    <Toggle bind:checked={auth.enabled} onchange={handleChange} size="default" />
  </div>

  {#if auth.enabled}
    <div class="animate-in fade-in slide-in-from-top-1 space-y-4 duration-200">
      <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
        <div class="space-y-2">
          <Label>Token URL</Label>
          <EnvTokenInput
            bind:value={auth.tokenUrl}
            placeholder="https://auth.example.com/oauth2/token"
            onChange={() => handleTokenUrlChange(auth.tokenUrl)}
            size="sm"
          />
          <div class="text-xs text-neutral-500">
            The endpoint where the POST request for the token will be sent.
          </div>
        </div>

        <div class="space-y-2">
          <Label for="token-path">Token Path (JSONPath-like)</Label>
          <Input
            id="token-path"
            type="text"
            size="sm"
            placeholder="access_token"
            bind:value={auth.tokenPath}
            oninput={() => handleTokenPathChange(auth.tokenPath ?? "")}
          />
          <div class="text-xs text-neutral-500">
            The field in the JSON response containing the access token.
          </div>
        </div>
      </div>

      <div class="space-y-3">
        <div class="flex items-center justify-between">
          <Label class="font-semibold">Token Request Template (POST Body)</Label>
          <Button color="light" size="xs" onclick={addRow}>+ Add Parameter</Button>
        </div>

        <div
          class="rounded-lg border border-neutral-200 bg-neutral-50 p-2 dark:border-neutral-700 dark:bg-neutral-900/50"
        >
          <div
            class="mb-2 grid grid-cols-[1fr_1.5fr_auto] gap-2 px-2 text-xs font-medium text-neutral-500 uppercase"
          >
            <div>Key</div>
            <div>Value</div>
            <div class="w-8"></div>
          </div>

          <div class="space-y-2">
            {#each templateRows as row (row.id)}
              <div class="grid grid-cols-[1fr_1.5fr_auto] items-center gap-2">
                <Input
                  type="text"
                  size="sm"
                  placeholder="grant_type"
                  bind:value={row.key}
                  oninput={() => updateTemplateRowKey(row.id, row.key)}
                  class="bg-white dark:bg-neutral-800"
                />
                <EnvTokenInput
                  bind:value={row.value}
                  placeholder="client_credentials"
                  onChange={() => updateTemplateRowValue(row.id, row.value)}
                  size="sm"
                />
                <Button
                  color="light"
                  size="xs"
                  class="h-8 w-8 p-0"
                  onclick={() => removeRow(row.id)}
                  title="Remove parameter"
                >
                  <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      stroke-width="2"
                      d="M6 18L18 6M6 6l12 12"
                    />
                  </svg>
                </Button>
              </div>
            {/each}

            {#if templateRows.length === 0}
              <div class="py-4 text-center text-sm text-neutral-500">
                No parameters defined. Add common ones like <code>grant_type</code>,
                <code>client_id</code>, etc.
              </div>
            {/if}
          </div>
        </div>
      </div>
    </div>
  {/if}
</div>
