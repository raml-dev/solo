<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import HistoryEntry from "$src/lib/components/History/HistoryEntry.svelte";
  import FeedbackEmptyState from "$src/lib/components/common/FeedbackEmptyState.svelte";
  import { historyStore, historyStoreState } from "$src/lib/stores/historyStore.svelte";
  import Button from "flowbite-svelte/Button.svelte";
  import Input from "flowbite-svelte/Input.svelte";
  import Select from "flowbite-svelte/Select.svelte";

  let filterMethod = $state("");
  let filterStatus = $state("");
  let filterUrl = $state("");

  const METHOD_OPTIONS = [
    { value: "", name: "All" },
    { value: "GET", name: "GET" },
    { value: "POST", name: "POST" },
    { value: "PUT", name: "PUT" },
    { value: "PATCH", name: "PATCH" },
    { value: "DELETE", name: "DELETE" }
  ];

  const STATUS_OPTIONS = [
    { value: "", name: "All" },
    { value: "2xx", name: "2xx Success" },
    { value: "3xx", name: "3xx Redirect" },
    { value: "4xx", name: "4xx Client error" },
    { value: "5xx", name: "5xx Server error" },
    { value: "err", name: "Error" }
  ];

  let filtered = $derived(
    $historyStore.filter((e) => {
      if (filterMethod && e.request.method !== filterMethod) return false;
      if (filterUrl && !e.request.url.toLowerCase().includes(filterUrl.toLowerCase())) return false;
      if (filterStatus) {
        if (filterStatus === "err") return !!e.error;
        if (!e.response) return false;
        const s = e.response.status;
        if (filterStatus === "2xx" && !(s >= 200 && s < 300)) return false;
        if (filterStatus === "3xx" && !(s >= 300 && s < 400)) return false;
        if (filterStatus === "4xx" && !(s >= 400 && s < 500)) return false;
        if (filterStatus === "5xx" && !(s >= 500)) return false;
      }
      return true;
    })
  );
</script>

<div class="flex h-full min-h-0 w-full flex-col gap-3 overflow-hidden p-3">
  <div
    class="flex flex-wrap items-center gap-4 border-b border-neutral-200 pb-3 dark:border-neutral-700"
  >
    <span class="text-xs font-semibold text-neutral-700 dark:text-neutral-200">History</span>

    <div class="flex flex-1 items-center gap-2">
      <span class="text-xs font-bold whitespace-nowrap text-neutral-500 uppercase">URL</span>
      <Input
        size="sm"
        class="h-7 w-full min-w-40"
        type="text"
        placeholder="Filter URL..."
        classes={{ input: "h-7" }}
        bind:value={filterUrl}
      />
    </div>

    <div class="flex items-center gap-2">
      <span class="text-xs font-bold whitespace-nowrap text-neutral-500 uppercase">Method</span>
      <Select
        size="sm"
        bind:value={filterMethod}
        items={METHOD_OPTIONS}
        class="h-7 w-24"
        classes={{ select: "h-7 py-0" }}
        placeholder=""
      />
    </div>

    <div class="flex items-center gap-2">
      <span class="text-xs font-bold whitespace-nowrap text-neutral-500 uppercase">Status</span>
      <Select
        size="sm"
        bind:value={filterStatus}
        items={STATUS_OPTIONS}
        class="h-7 w-32"
        classes={{ select: "h-7 py-0" }}
        placeholder=""
      />
    </div>

    <div class="ml-auto flex h-6 items-center gap-3">
      <span
        class="text-xs font-bold whitespace-nowrap text-neutral-400 uppercase dark:text-neutral-500"
      >
        {filtered.length} / {$historyStore.length}
      </span>

      <Button
        color="light"
        class="h-7"
        size="sm"
        onclick={() => void historyStore.exportToHarFile($historyStore)}
        disabled={$historyStore.length === 0 || historyStoreState.exporting}
      >
        {historyStoreState.exporting ? "Exporting..." : "Export HAR"}
      </Button>
      <Button
        color="alternative"
        size="sm"
        class="h-7"
        onclick={() => historyStore.clear()}
        disabled={$historyStore.length === 0}
      >
        Clear
      </Button>
    </div>
  </div>

  <div class="min-h-0 flex-1 overflow-y-auto pr-1">
    {#if filtered.length === 0}
      <div class="flex h-full items-center justify-center">
        {#if $historyStore.length === 0}
          <FeedbackEmptyState
            compact
            title="No requests yet"
            detail="Send a request to see it in the history"
          />
        {:else}
          <FeedbackEmptyState compact title="No entries match the current filters" />
        {/if}
      </div>
    {:else}
      <div class="space-y-2">
        {#each filtered as entry (entry.id)}
          <HistoryEntry {entry} />
        {/each}
      </div>
    {/if}
  </div>
</div>
