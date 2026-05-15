<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import { getActiveTab, tabStore } from "$src/lib/stores/tabStore.svelte";
  import { getHttpStatusString, getStatusBadgeColor } from "$src/lib/utils/http";
  import StopSolid from "flowbite-svelte-icons/StopSolid.svelte";
  import Badge from "flowbite-svelte/Badge.svelte";
  import Button from "flowbite-svelte/Button.svelte";
  import Input from "flowbite-svelte/Input.svelte";
  import Label from "flowbite-svelte/Label.svelte";
  import Progressbar from "flowbite-svelte/Progressbar.svelte";
  import Table from "flowbite-svelte/Table.svelte";
  import TableBody from "flowbite-svelte/TableBody.svelte";
  import TableBodyCell from "flowbite-svelte/TableBodyCell.svelte";
  import TableBodyRow from "flowbite-svelte/TableBodyRow.svelte";
  import TableHead from "flowbite-svelte/TableHead.svelte";
  import TableHeadCell from "flowbite-svelte/TableHeadCell.svelte";
  import Toggle from "flowbite-svelte/Toggle.svelte";

  let tab = $derived(getActiveTab());

  function startRun() {
    if (tab) {
      tabStore.startRunnerActiveTab();
    }
  }

  function stopRun() {
    if (tab) {
      tabStore.cancelRunnerActiveTab();
    }
  }

  function saveConfig() {
    tabStore.storeTabsInLocalStorage();
  }
</script>

{#if tab}
  <div class="min-h-0 flex-1 space-y-4 overflow-y-auto p-3">
    <div
      class="rounded-lg border border-neutral-200 bg-white p-3 dark:border-neutral-700 dark:bg-neutral-800"
    >
      <div class="flex flex-wrap items-end gap-4">
        <div class="flex flex-col gap-1">
          <Label for="concurrency">Concurrency</Label>
          <Input
            id="concurrency"
            type="number"
            bind:value={tab.runnerConfig.concurrency}
            onchange={saveConfig}
            min="1"
            max="100"
            size="sm"
            disabled={tab.runnerState.running}
            aria-label="Number of parallel workers"
          />
        </div>

        <div class="flex flex-col gap-1">
          <Label for="iterations">Iterations</Label>
          <Input
            id="iterations"
            type="number"
            bind:value={tab.runnerConfig.iterations}
            onchange={saveConfig}
            min="1"
            size="sm"
            disabled={tab.runnerState.running}
            aria-label="Total number of requests to perform"
          />
        </div>

        <div class="flex flex-col gap-1 md:pt-6">
          <Toggle
            bind:checked={tab.runnerConfig.stopOnError}
            onchange={saveConfig}
            size="small"
            disabled={tab.runnerState.running}>Stop on error</Toggle
          >
        </div>

        <div class="ml-auto flex gap-2">
          <Button
            color="primary"
            onclick={startRun}
            disabled={tab.runnerState.running}
            loading={tab.runnerState.running}
          >
            {tab.runnerState.running ? "Running..." : "Start Run"}
          </Button>
          {#if tab.runnerState.running}
            <Button color="red" onclick={stopRun}>
              <StopSolid class="me-2 h-4 w-4" />
              Stop
            </Button>
          {/if}
        </div>
      </div>
    </div>

    {#if tab.runnerState.running || tab.runnerState.stats || tab.runnerState.lastResults.length > 0}
      <div class="space-y-3">
        {#if tab.runnerState.running}
          <Progressbar
            progress={tab.runnerState.progress}
            size="h-2"
            color="blue"
            labelInside={false}
          />
        {/if}

        {#if tab.runnerState.stats}
          <div class="grid grid-cols-2 gap-2 sm:grid-cols-4">
            <div
              class="flex flex-col gap-0.5 rounded-lg border border-neutral-200 p-3 dark:border-neutral-700"
            >
              <span class="text-xs text-neutral-500 dark:text-neutral-400">Requests</span>
              <span class="text-base font-semibold text-neutral-900 dark:text-neutral-100"
                >{tab.runnerState.stats.successCount} / {tab.runnerState.stats.totalRequests}</span
              >
            </div>
            <div
              class="flex flex-col gap-0.5 rounded-lg border border-neutral-200 p-3 dark:border-neutral-700"
            >
              <span class="text-xs text-neutral-500 dark:text-neutral-400">Avg Latency</span>
              <span class="text-base font-semibold text-neutral-900 dark:text-neutral-100"
                >{tab.runnerState.stats.avgLatency}ms</span
              >
            </div>
            <div
              class="flex flex-col gap-0.5 rounded-lg border border-neutral-200 p-3 dark:border-neutral-700"
            >
              <span class="text-xs text-neutral-500 dark:text-neutral-400">P95</span>
              <span class="text-base font-semibold text-neutral-900 dark:text-neutral-100"
                >{tab.runnerState.stats.p95Latency}ms</span
              >
            </div>
            <div
              class="flex flex-col gap-0.5 rounded-lg border border-neutral-200 p-3 dark:border-neutral-700"
            >
              <span class="text-xs text-neutral-500 dark:text-neutral-400">Throughput</span>
              <span class="text-base font-semibold text-neutral-900 dark:text-neutral-100"
                >{tab.runnerState.stats.requestsPerSec.toFixed(2)} req/s</span
              >
            </div>
          </div>
        {/if}

        <div class="overflow-x-auto">
          <Table>
            <TableHead>
              <TableHeadCell>#</TableHeadCell>
              <TableHeadCell>Status</TableHeadCell>
              <TableHeadCell>Latency</TableHeadCell>
              <TableHeadCell>Result</TableHeadCell>
            </TableHead>
            <TableBody>
              {#each tab.runnerState.lastResults as res (res.index)}
                <TableBodyRow>
                  <TableBodyCell>{res.index + 1}</TableBodyCell>
                  <TableBodyCell>
                    {#if res.error}
                      <Badge color="red">ERROR</Badge>
                    {:else if res.response}
                      <Badge color={getStatusBadgeColor(res.response.statusCode)}>
                        {getHttpStatusString(res.response.statusCode)}
                      </Badge>
                    {/if}
                  </TableBodyCell>
                  <TableBodyCell>{res.response?.duration ?? "-"} ms</TableBodyCell>
                  <TableBodyCell>
                    <span class="max-w-xs truncate text-xs text-neutral-500 dark:text-neutral-400"
                      >{res.error || "Success"}</span
                    >
                  </TableBodyCell>
                </TableBodyRow>
              {/each}
            </TableBody>
          </Table>
        </div>
      </div>
    {:else}
      <div
        class="rounded-lg border border-neutral-200 bg-white p-6 text-center dark:border-neutral-700 dark:bg-neutral-800"
      >
        <h3 class="text-base font-semibold">Parallel Runner</h3>
        <p>Configure concurrency and iterations to perform load testing on this request.</p>
      </div>
    {/if}
  </div>
{/if}
