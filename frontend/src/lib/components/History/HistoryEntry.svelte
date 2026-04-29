<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import type { HistoryEntry } from "$src/lib/stores/historyStore.svelte";
  import { formatTime, truncateString } from "$src/lib/utils/helpers";
  import {
    getHttpStatusString,
    getMethodBadgeClass,
    getStatusBadgeColor
  } from "$src/lib/utils/http";
  import Badge from "flowbite-svelte/Badge.svelte";
  import Button from "flowbite-svelte/Button.svelte";

  interface Props {
    entry: HistoryEntry;
  }

  let { entry }: Props = $props();

  let expanded = $state(false);
  let detailTab: "request" | "response" = $state("response");

  function formatHeaders(headers: Record<string, string>): string {
    return Object.entries(headers)
      .map(([k, v]) => `${k}: ${v}`)
      .join("\n");
  }
</script>

<div class="border-b border-neutral-200 dark:border-neutral-700">
  <Button
    color="light"
    class="w-full justify-start gap-2 rounded-none border-0 px-3 py-2 text-xs shadow-none"
    onclick={() => (expanded = !expanded)}
  >
    <span class="shrink-0 text-neutral-400">{expanded ? "▾" : "▸"}</span>
    <div class="text-left">
      <span class="shrink-0 align-middle font-mono text-neutral-500 dark:text-neutral-400"
        >{formatTime(entry.timestamp)}</span
      >
      <span class={getMethodBadgeClass(entry.request.method)}>
        {entry.request.method}
      </span>
    </div>

    {#if entry.response}
      <div class="w-90 text-left">
        <span
          class="min-w-0 truncate text-left text-neutral-700 dark:text-neutral-300"
          title={entry.request.url}>{truncateString(entry.request.url)}</span
        >
      </div>
    {/if}

    <div class="">
      {#if entry.error}
        <Badge color="red">ERR</Badge>
      {:else if entry.response}
        <div class="flex">
          <span class="w-15 shrink-0 text-left text-neutral-500 dark:text-neutral-400"
            >{entry.response.time}ms</span
          >
          <Badge color={getStatusBadgeColor(entry.response.status)}>
            {getHttpStatusString(entry.response.status)}
          </Badge>
        </div>
      {/if}
    </div>

    {#if entry.collectionName}
      <!-- trick to make truncation happen on the _beginning_ of the string :
      make the container right-to-left, and the actual text again left-to-right
    -->
      <div
        dir="rtl"
        class="min-w-0 flex-1 shrink-0 truncate text-right text-neutral-400 dark:text-neutral-500"
      >
        <span dir="ltr"
          >{entry.collectionName}{entry.requestName ? ` / ${entry.requestName}` : ""}</span
        >
      </div>
    {/if}
  </Button>

  {#if expanded}
    <div
      class="border-t border-neutral-100 bg-neutral-50 p-3 dark:border-neutral-800 dark:bg-neutral-900/50"
    >
      <div class="flex items-center gap-2">
        <Button
          color={detailTab === "request" ? "primary" : "light"}
          size="xs"
          onclick={() => (detailTab = "request")}>Request</Button
        >
        <Button
          color={detailTab === "response" ? "primary" : "light"}
          size="xs"
          onclick={() => (detailTab = "response")}>Response</Button
        >
      </div>

      {#if detailTab === "request"}
        <div class="flex flex-col gap-1">
          <p
            class="text-xs font-semibold tracking-wide text-neutral-500 uppercase dark:text-neutral-400"
          >
            URL
          </p>
          <pre
            class="overflow-x-auto rounded bg-neutral-100 p-2 font-mono text-xs text-neutral-800 dark:bg-neutral-800 dark:text-neutral-100">{entry
              .request.method} {entry.request.url}</pre>
        </div>
        {#if Object.keys(entry.request.headers).length > 0}
          <div class="flex flex-col gap-1">
            <p
              class="text-xs font-semibold tracking-wide text-neutral-500 uppercase dark:text-neutral-400"
            >
              Headers
            </p>
            <pre
              class="overflow-x-auto rounded bg-neutral-100 p-2 font-mono text-xs text-neutral-800 dark:bg-neutral-800 dark:text-neutral-100">{formatHeaders(
                entry.request.headers
              )}</pre>
          </div>
        {/if}
        {#if entry.request.body}
          <div class="flex flex-col gap-1">
            <p
              class="text-xs font-semibold tracking-wide text-neutral-500 uppercase dark:text-neutral-400"
            >
              Body
            </p>
            <pre
              class="overflow-x-auto rounded bg-neutral-100 p-2 font-mono text-xs text-neutral-800 dark:bg-neutral-800 dark:text-neutral-100">{entry
                .request.body}</pre>
          </div>
        {/if}
      {:else if entry.error}
        <div class="flex flex-col gap-1 text-danger-600 dark:text-danger-400">
          <p
            class="text-xs font-semibold tracking-wide text-neutral-500 uppercase dark:text-neutral-400"
          >
            Error
          </p>
          <pre
            class="overflow-x-auto rounded bg-neutral-100 p-2 font-mono text-xs text-neutral-800 dark:bg-neutral-800 dark:text-neutral-100">{entry.error}</pre>
        </div>
      {:else if entry.response}
        <div class="flex flex-col gap-1">
          <p
            class="text-xs font-semibold tracking-wide text-neutral-500 uppercase dark:text-neutral-400"
          >
            Status
          </p>
          <pre
            class="overflow-x-auto rounded bg-neutral-100 p-2 font-mono text-xs text-neutral-800 dark:bg-neutral-800 dark:text-neutral-100">{entry
              .response.status} — {entry.response.time}ms</pre>
        </div>
        {#if Object.keys(entry.response.headers).length > 0}
          <div class="flex flex-col gap-1">
            <p
              class="text-xs font-semibold tracking-wide text-neutral-500 uppercase dark:text-neutral-400"
            >
              Headers
            </p>
            <pre
              class="overflow-x-auto rounded bg-neutral-100 p-2 font-mono text-xs text-neutral-800 dark:bg-neutral-800 dark:text-neutral-100">{formatHeaders(
                entry.response.headers
              )}</pre>
          </div>
        {/if}
        {#if entry.response.body}
          <div class="flex flex-col gap-1">
            <p
              class="text-xs font-semibold tracking-wide text-neutral-500 uppercase dark:text-neutral-400"
            >
              Body
            </p>
            <pre
              class="overflow-x-auto rounded bg-neutral-100 p-2 font-mono text-xs text-neutral-800 dark:bg-neutral-800 dark:text-neutral-100">{entry
                .response.body}</pre>
          </div>
        {/if}
      {/if}
    </div>
  {/if}
</div>
