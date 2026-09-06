<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import EnvTokenInput from "$src/lib/components/RequestBuilder/EnvTokenInput.svelte";
  import { notifications } from "$src/lib/stores/notificationStore";
  import { tabStore } from "$src/lib/stores/tabStore.svelte";
  import { extractEnvTokenMatches } from "$src/lib/utils/tokens";
  import type { ResolvedVariableEntry } from "$src/lib/utils/variableResolution";
  import { collection } from "$wails/go/models";
  import EyeOutline from "flowbite-svelte-icons/EyeOutline.svelte";
  import EyeSlashOutline from "flowbite-svelte-icons/EyeSlashOutline.svelte";
  import PlusOutline from "flowbite-svelte-icons/PlusOutline.svelte";
  import TrashBinOutline from "flowbite-svelte-icons/TrashBinOutline.svelte";
  import Alert from "flowbite-svelte/Alert.svelte";
  import Button from "flowbite-svelte/Button.svelte";
  import Input from "flowbite-svelte/Input.svelte";
  import Label from "flowbite-svelte/Label.svelte";
  import Radio from "flowbite-svelte/Radio.svelte";

  interface Props {
    auth: collection.AuthConfiguration;
    variableEntries?: ResolvedVariableEntry[];
    hasAuthorizationHeader?: boolean;
    onChange?: () => void;
  }

  type TemplateRow = {
    id: string;
    key: string;
    value: string;
  };

  let {
    auth = $bindable(),
    variableEntries = [],
    hasAuthorizationHeader = false,
    onChange
  }: Props = $props();

  let tokenVisible = $state(false);
  let revealingToken = $state(false);
  let bearerVariableEntries = $derived(
    variableEntries.map((entry) => ({
      ...entry,
      computedValue: entry.computedValue ? "••••••••" : ""
    }))
  );
  let bearerTokenIsPlaceholder = $derived(isBearerPlaceholder(auth.bearerToken));
  let bearerTokenVisibilityAvailable = $derived(
    !bearerTokenIsPlaceholder && Boolean(auth.bearerToken || auth.bearerTokenId)
  );

  function isBearerPlaceholder(value?: string): boolean {
    const trimmedValue = (value ?? "").trim();
    const matches = extractEnvTokenMatches(trimmedValue);
    return matches.length === 1 && matches[0].full === trimmedValue;
  }

  function handleChange() {
    onChange?.();
  }

  function selectAuthMode(mode: "none" | "bearer" | "oauth2") {
    auth.mode = mode;
    auth.enabled = auth.mode === "oauth2";
    tokenVisible = false;
    handleChange();
  }

  function handleBearerTokenChange() {
    if (auth.bearerTokenId) {
      auth.bearerTokenId = "";
    }
    if (isBearerPlaceholder(auth.bearerToken)) {
      tokenVisible = false;
    }
    handleChange();
  }

  async function toggleTokenVisibility() {
    if (tokenVisible) {
      tokenVisible = false;
      return;
    }
    if (!auth.bearerToken && auth.bearerTokenId) {
      revealingToken = true;
      try {
        auth.bearerToken = await tabStore.revealBearerToken(auth.bearerTokenId);
      } catch (error) {
        notifications.error("Failed to reveal bearer token", String(error));
        return;
      } finally {
        revealingToken = false;
      }
    }
    if (isBearerPlaceholder(auth.bearerToken)) {
      tokenVisible = false;
      return;
    }
    tokenVisible = true;
  }

  function clearBearerToken() {
    auth.bearerToken = "";
    auth.bearerTokenId = "";
    tokenVisible = false;
    handleChange();
  }

  function handleTokenUrlChange(value: string) {
    auth.tokenUrl = value;
    onChange?.();
  }

  function handleTokenPathChange(value: string) {
    auth.tokenPath = value;
    onChange?.();
  }

  function createTemplateRows(template: Record<string, string> = {}): TemplateRow[] {
    return Object.entries(template).map(([key, value], index) => ({
      id: `row-${index}-${Date.now()}`,
      key,
      value: String(value)
    }));
  }

  function initializeTemplateRows() {
    return createTemplateRows(auth.template || {});
  }

  let templateRows = $state<TemplateRow[]>([]);
  templateRows = initializeTemplateRows();

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

{#snippet bearerTokenVisibilityControl()}
  <button
    type="button"
    class="inline-flex rounded p-0.5 text-neutral-500 hover:text-neutral-800 focus:ring-2 focus:ring-primary-500 focus:outline-hidden disabled:cursor-wait disabled:opacity-50 dark:text-neutral-400 dark:hover:text-white"
    onclick={toggleTokenVisibility}
    disabled={revealingToken}
    aria-label={tokenVisible ? "Hide bearer token" : "Show bearer token"}
    title={tokenVisible ? "Hide bearer token" : "Show bearer token"}
  >
    {#if tokenVisible}
      <EyeSlashOutline class="h-4 w-4" />
    {:else}
      <EyeOutline class="h-4 w-4" />
    {/if}
  </button>
{/snippet}

<div class="flex-1 space-y-6 overflow-y-auto p-4">
  <div class="space-y-3 border-b border-neutral-100 pb-4 dark:border-neutral-800">
    <div class="space-y-0.5">
      <Label class="text-base font-semibold">Authentication</Label>
      <div class="text-sm text-neutral-500 dark:text-neutral-400">
        Select an authentication method for this request.
      </div>
    </div>

    <div class="space-y-2">
      <Button
        onclick={() => selectAuthMode("none")}
        class="focus:ring-0 flex w-full justify-between rounded-lg border border-neutral-200 bg-white p-3 hover:bg-white dark:border-neutral-700 dark:bg-neutral-800 dark:hover:bg-neutral-800"
      >
        <div class="flex min-w-0 flex-col items-start justify-between gap-1">
          <Label for="auth-none-radio" class="font-medium">No Auth</Label>
          <div class="text-xs text-neutral-500 dark:text-neutral-400">
            Do not use native authentication for this request.
          </div>
        </div>
        <Radio
          id="auth-none-radio"
          name="auth-mode"
          value="none"
          group={auth.mode}
          onchange={() => selectAuthMode("none")}
          aria-label="No Auth"
        />
      </Button>

      <Button
        onclick={() => selectAuthMode("bearer")}
        class="focus:ring-0 flex w-full justify-between rounded-lg border border-neutral-200 bg-white p-3 hover:bg-white dark:border-neutral-700 dark:bg-neutral-800 dark:hover:bg-neutral-800"
      >
        <div class="flex min-w-0 flex-col items-start justify-between gap-1">
          <Label for="auth-bearer-radio" class="font-medium">Bearer Token</Label>
          <div class="text-xs text-neutral-500 dark:text-neutral-400">
            Send a static token in the Authorization header.
          </div>
        </div>
        <Radio
          id="auth-bearer-radio"
          name="auth-mode"
          value="bearer"
          group={auth.mode}
          onchange={() => selectAuthMode("bearer")}
          aria-label="Bearer Token"
        />
      </Button>

      <Button
        onclick={() => selectAuthMode("oauth2")}
        class="focus:ring-0 flex w-full justify-between rounded-lg border border-neutral-200 bg-white p-3 hover:bg-white dark:border-neutral-700 dark:bg-neutral-800 dark:hover:bg-neutral-800"
      >
        <div class="flex min-w-0 flex-col items-start justify-between gap-1">
          <Label for="auth-oauth2-radio" class="font-medium">OAuth 2.0</Label>
          <div class="text-xs text-neutral-500 dark:text-neutral-400">
            Fetch and refresh a token automatically.
          </div>
        </div>
        <Radio
          id="auth-oauth2-radio"
          name="auth-mode"
          value="oauth2"
          group={auth.mode}
          onchange={() => selectAuthMode("oauth2")}
          aria-label="OAuth 2.0"
        />
      </Button>
    </div>
  </div>

  {#if auth.mode !== "none" && hasAuthorizationHeader}
    <Alert color="yellow">
      The enabled Authorization header takes precedence over native authentication.
    </Alert>
  {/if}

  {#if auth.mode === "bearer"}
    <div class="animate-in fade-in slide-in-from-top-1 space-y-3 duration-200">
      <div class="space-y-2">
        <Label>Token</Label>
        <div class="flex items-center gap-2">
          <EnvTokenInput
            bind:value={auth.bearerToken}
            placeholder={auth.bearerTokenId ? "Stored token" : "Enter bearer token"}
            class="bearer-token-input {tokenVisible ? '' : 'bearer-token-masked'}"
            variableEntries={bearerVariableEntries}
            right={bearerTokenVisibilityControl}
            rightVisible={bearerTokenVisibilityAvailable}
            onChange={handleBearerTokenChange}
          />
          <Button
            color="light"
            size="sm"
            onclick={clearBearerToken}
            disabled={!auth.bearerToken && !auth.bearerTokenId}>Clear</Button
          >
        </div>
        <div class="text-xs text-neutral-500">
          Enter the token or use <code>{"{{variable}}"}</code> from session, environment, or
          collection variables. Solo adds the <code>Bearer</code> prefix automatically.
        </div>
      </div>
    </div>
  {:else if auth.mode === "oauth2"}
    <div class="animate-in fade-in slide-in-from-top-1 space-y-4 duration-200">
      <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
        <div class="space-y-2">
          <Label>Token URL</Label>
          <EnvTokenInput
            bind:value={auth.tokenUrl}
            placeholder="https://auth.example.com/oauth2/token"
            onChange={() => handleTokenUrlChange(auth.tokenUrl ?? "")}
            size="sm"
            {variableEntries}
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
          <Button color="light" size="xs" onclick={addRow}>
            <PlusOutline class="mr-1 h-4 w-4" />
            <span>Add Parameter</span>
          </Button>
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
                  {variableEntries}
                />
                <Button
                  color="light"
                  size="xs"
                  class="h-8 w-8 p-0"
                  onclick={() => removeRow(row.id)}
                  title="Remove parameter"
                >
                  <TrashBinOutline class="h-4 w-4" />
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

<style>
  :global(.bearer-token-masked .cm-content) {
    -webkit-text-security: disc;
  }

  :global(.bearer-token-input .cm-env-token) {
    -webkit-text-security: none;
    pointer-events: none;
  }
</style>
