<script lang="ts">
  import Badge from "flowbite-svelte/Badge.svelte";
  import Button from "flowbite-svelte/Button.svelte";
  import Card from "flowbite-svelte/Card.svelte";
  import Input from "flowbite-svelte/Input.svelte";
  import Label from "flowbite-svelte/Label.svelte";
  import Progressbar from "flowbite-svelte/Progressbar.svelte";
  import Toggle from "flowbite-svelte/Toggle.svelte";
  import { selectedEnvironment } from "$src/lib/stores/environmentStore";
  import { GetSessionVars, RunParallel } from "$wails/go/main/App";
  import type { configuration as conf } from "$wails/go/models";
  import { main, runner } from "$wails/go/models";
  import { EventsOff, EventsOn } from "$wails/runtime";
  import { onMount } from "svelte";

  interface Header {
    id: string;
    key: string;
    value: string;
    enabled: boolean;
  }

  interface Props {
    method: string;
    url: string;
    body: string;
    headers: Header[];
    settings: conf.RequestSettingsOverride;
    preRequestScript: string;
    postResponseScript: string;
  }

  let { method, url, body, headers, settings, preRequestScript, postResponseScript }: Props =
    $props();

  let concurrency = $state(5);
  let iterations = $state(20);
  let stopOnError = $state(false);
  let running = $state(false);
  let progress = $state(0);

  let stats: runner.RunnerStats | null = $state(null);
  let lastResults: runner.RunnerResult[] = $state([]);
  const MAX_VISIBLE_RESULTS = 50;

  let environmentEntries = $derived(
    Object.entries($selectedEnvironment?.values ?? {}).map(([key, val]) => ({
      key,
      value: String(val?.value ?? "")
    }))
  );

  function resolveEnvironmentTokens(
    value: string,
    sessionVars: Record<string, string> = {}
  ): string {
    if (!value) return value;
    const envMap = new Map(environmentEntries.map((e) => [e.key, e.value]));
    const sessionMap = new Map(Object.entries(sessionVars || {}));
    return value.replace(/\{\{([^{}\r\n]+?)\}\}/g, (_full, key: string) => {
      const k = key.trim();
      if (sessionMap.has(k)) return String(sessionMap.get(k) ?? "");
      if (envMap.has(k)) return String(envMap.get(k) ?? "");
      return _full;
    });
  }

  async function startRun() {
    if (running) return;

    running = true;
    progress = 0;
    stats = null;
    lastResults = [];

    const sessionVars = await GetSessionVars().catch(() => ({}) as Record<string, string>);

    const resolvedUrl = resolveEnvironmentTokens(url, sessionVars);
    const resolvedBody = resolveEnvironmentTokens(body, sessionVars);
    const resolvedHeaders = headers
      .filter((h) => h.enabled)
      .reduce(
        (acc, { key, value }) => ({
          ...acc,
          [resolveEnvironmentTokens(key, sessionVars)]: resolveEnvironmentTokens(value, sessionVars)
        }),
        {} as Record<string, string>
      );

    const requestOptions = new main.RequestOptions({
      body: resolvedBody,
      headers: resolvedHeaders,
      method,
      url: resolvedUrl,
      settings: settings,
      preRequestScript: preRequestScript || "",
      postResponseScript: postResponseScript || ""
    });

    try {
      EventsOn("runner:result", (res: runner.RunnerResult) => {
        lastResults = [res, ...lastResults].slice(0, MAX_VISIBLE_RESULTS);
        progress = Math.min(100, (lastResults.length / iterations) * 100);
      });

      stats = await RunParallel(requestOptions, concurrency, iterations, stopOnError);
    } catch (err) {
      console.error("Runner failed", err);
    } finally {
      running = false;
      EventsOff("runner:result");
    }
  }

  function getStatusBadgeColor(status: number): "green" | "blue" | "yellow" | "red" {
    if (status >= 200 && status < 300) return "green";
    if (status >= 300 && status < 400) return "blue";
    if (status >= 400 && status < 500) return "yellow";
    return "red";
  }

  onMount(() => {
    return () => {
      EventsOff("runner:result");
    };
  });
</script>

<div class="runner-container space-y-4">
  <Card class="runner-config p-3">
    <div class="config-group">
      <Label for="concurrency">Concurrency</Label>
      <Input
        id="concurrency"
        type="number"
        bind:value={concurrency}
        min="1"
        max="100"
        size="sm"
        disabled={running}
        aria-label="Number of parallel workers"
      />
    </div>

    <div class="config-group">
      <Label for="iterations">Iterations</Label>
      <Input
        id="iterations"
        type="number"
        bind:value={iterations}
        min="1"
        size="sm"
        disabled={running}
        aria-label="Total number of requests to perform"
      />
    </div>

    <div class="config-group checkbox md:pt-6">
      <Toggle bind:checked={stopOnError} size="small" disabled={running}>Stop on error</Toggle>
    </div>

    <div class="flex-spacer"></div>

    <Button color="primary" class="ml-3" onclick={startRun} disabled={running} loading={running}>
      {running ? "Running..." : "Start Run"}
    </Button>
  </Card>

  {#if running || stats || lastResults.length > 0}
    <div class="runner-results space-y-3">
      {#if running}
        <Progressbar {progress} size="h-2" color="blue" labelInside={false} />
      {/if}

      {#if stats}
        <div class="stats-grid">
          <div class="stat-card">
            <span class="stat-label">Requests</span>
            <span class="stat-value">{stats.successCount} / {stats.totalRequests}</span>
          </div>
          <div class="stat-card">
            <span class="stat-label">Avg Latency</span>
            <span class="stat-value">{stats.avgLatency}ms</span>
          </div>
          <div class="stat-card">
            <span class="stat-label">P95</span>
            <span class="stat-value">{stats.p95Latency}ms</span>
          </div>
          <div class="stat-card">
            <span class="stat-label">Throughput</span>
            <span class="stat-value">{stats.requestsPerSec.toFixed(2)} req/s</span>
          </div>
        </div>
      {/if}

      <div class="results-table-container">
        <table class="results-table">
          <thead>
            <tr>
              <th>#</th>
              <th>Status</th>
              <th>Latency</th>
              <th>Result</th>
            </tr>
          </thead>
          <tbody>
            {#each lastResults as res (res.index)}
              <tr>
                <td>{res.index + 1}</td>
                <td>
                  {#if res.error}
                    <Badge color="red">ERROR</Badge>
                  {:else if res.response}
                    <Badge color={getStatusBadgeColor(res.response.statusCode)}>
                      {res.response.statusCode}
                    </Badge>
                  {/if}
                </td>
                <td>{res.response?.duration ?? "-"} ms</td>
                <td class="error-cell">{res.error || "Success"}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
  {:else}
    <Card class="empty-runner p-6">
      <h3 class="text-base font-semibold">Parallel Runner</h3>
      <p>Configure concurrency and iterations to perform load testing on this request.</p>
    </Card>
  {/if}
</div>
