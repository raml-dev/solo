<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import CodeMirrorEditor from "$src/lib/components/RequestBuilder/CodeMirrorEditor.svelte";
  import { sessionVarsStore } from "$src/lib/stores/sessionVarsStore";
  import TrashBinOutline from "flowbite-svelte-icons/TrashBinOutline.svelte";
  import Badge from "flowbite-svelte/Badge.svelte";
  import Button from "flowbite-svelte/Button.svelte";

  interface Props {
    preRequestScript?: string;
    postResponseScript?: string;
    onPreChange?: (val: string) => void;
    onPostChange?: (val: string) => void;
  }

  let {
    preRequestScript = $bindable(""),
    postResponseScript = $bindable(""),
    onPreChange = () => {},
    onPostChange = () => {}
  }: Props = $props();

  type ScriptSection = "pre" | "post";
  let activeSection: ScriptSection = $state("pre");

  let sessionEntries: [string, string][] = $derived(
    Object.entries($sessionVarsStore) as [string, string][]
  );

  const LUA_HINT = `-- Available globals:
-- request.method, request.url, request.headers, request.body  (pre only, mutable)
-- response.status, response.headers, response.body, response.time  (post only)
-- env.get("key")  /  env.set("key", "value")  /  env.key = "value"
-- json.parse(s)  /  json.stringify(t)  /  xml.parse(s)
-- env.log("message")`;
</script>

<div class="flex min-h-0 flex-1 overflow-hidden">
  <nav
    class="flex w-44 shrink-0 flex-col gap-1 overflow-y-auto border-r border-neutral-200 p-2 dark:border-neutral-700"
  >
    <Button
      color={activeSection === "pre" ? "primary" : "light"}
      size="sm"
      class="w-full justify-start"
      onclick={() => (activeSection = "pre")}
    >
      <span>Pre-request</span>
      {#if preRequestScript.trim()}
        <span class="ml-1 h-1.5 w-1.5 shrink-0 rounded-full bg-primary-500" title="Script active"
        ></span>
      {/if}
    </Button>

    <Button
      color={activeSection === "post" ? "primary" : "light"}
      size="sm"
      class="w-full justify-start"
      onclick={() => (activeSection = "post")}
    >
      <span>Post-response</span>
      {#if postResponseScript.trim()}
        <span class="ml-1 h-1.5 w-1.5 shrink-0 rounded-full bg-primary-500" title="Script active"
        ></span>
      {/if}
    </Button>

    <div class="mt-2 rounded-lg border border-neutral-200 p-2 dark:border-neutral-700">
      <div class="mb-1 flex items-center justify-between gap-1">
        <span class="text-xs font-semibold text-neutral-500 dark:text-neutral-400"
          >Session Vars</span
        >
        {#if sessionEntries.length > 0}
          <Button
            color="light"
            size="xs"
            class="h-8 shrink-0 border-none bg-transparent text-neutral-800 inset-ring-primary-500 focus-within:inset-ring-1 focus-within:outline-hidden hover:bg-neutral-200 hover:text-neutral-800 focus:ring-0 focus:outline-hidden dark:border-none dark:bg-transparent dark:text-neutral-100 dark:hover:text-neutral-100"
            onclick={() => sessionVarsStore.clear()}
            title="Clear all session variables"
          >
            <TrashBinOutline class="h-4 w-4 shrink-0" />
          </Button>
        {/if}
      </div>
      {#if sessionEntries.length === 0}
        <p class="text-xs text-neutral-400 dark:text-neutral-500">
          No session vars yet.<br />Use <code>env.set()</code> in a script.
        </p>
      {:else}
        <ul class="space-y-1">
          {#each sessionEntries as [key, value] (key)}
            <li class="flex items-baseline gap-1">
              <span
                class="truncate font-mono text-xs font-medium text-neutral-700 dark:text-neutral-300"
                >{key}</span
              >
              <span class="truncate font-mono text-xs text-neutral-500 dark:text-neutral-400"
                >{value}</span
              >
            </li>
          {/each}
        </ul>
      {/if}
    </div>
  </nav>

  <div class="flex min-w-0 flex-1 flex-col overflow-hidden">
    <div
      class="flex shrink-0 items-center justify-between border-b border-neutral-200 px-3 py-2 dark:border-neutral-700"
    >
      <span class="text-xs font-semibold text-neutral-700 dark:text-neutral-300">
        {activeSection === "pre" ? "Pre-request Script" : "Post-response Script"}
      </span>
      <Badge color="blue">Lua</Badge>
    </div>

    <pre
      class="shrink-0 overflow-x-auto border-b border-neutral-200 bg-neutral-50 px-3 py-2 font-mono text-xs text-neutral-500 dark:border-neutral-700 dark:bg-neutral-900 dark:text-neutral-400">{LUA_HINT}</pre>

    {#if activeSection === "pre"}
      <div class="min-h-0 flex-1 overflow-hidden">
        <CodeMirrorEditor value={preRequestScript} language="lua" onChange={onPreChange} />
      </div>
    {:else}
      <div class="min-h-0 flex-1 overflow-hidden">
        <CodeMirrorEditor value={postResponseScript} language="lua" onChange={onPostChange} />
      </div>
    {/if}
  </div>
</div>
