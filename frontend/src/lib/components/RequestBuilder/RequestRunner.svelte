<script lang="ts">
  import Button from "flowbite-svelte/Button.svelte";
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
      // Set up event listener for real-time results
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

  function getStatusClass(status: number): string {
    if (status >= 200 && status < 300) return "status-success";
    if (status >= 400) return "status-error";
    return "status-info";
  }

  onMount(() => {
    return () => {
      EventsOff("runner:result");
    };
  });
</script>

<div class="runner-container">
  <div class="runner-config">
    <div class="config-group">
      <label for="concurrency">
        <svg
          width="14"
          height="14"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          ><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon></svg
        >
        <span>Concurrency</span>
      </label>
      <input
        id="concurrency"
        type="number"
        bind:value={concurrency}
        min="1"
        max="100"
        disabled={running}
        aria-label="Number of parallel workers"
      />
    </div>

    <div class="config-group">
      <label for="iterations">
        <svg
          width="14"
          height="14"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          ><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"></polyline></svg
        >
        <span>Iterations</span>
      </label>
      <input
        id="iterations"
        type="number"
        bind:value={iterations}
        min="1"
        disabled={running}
        aria-label="Total number of requests to perform"
      />
    </div>

    <div class="config-group checkbox">
      <label>
        <input type="checkbox" bind:checked={stopOnError} disabled={running} />
        <span>Stop on error</span>
      </label>
    </div>

    <div class="flex-spacer"></div>

    <Button color="primary" onclick={startRun} disabled={running}>
      {#if running}
        <svg
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          class="spin"
          ><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"
          ></polyline></svg
        >
        <span>Running...</span>
      {:else}
        <svg
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"><polygon points="5 3 19 12 5 21 5 3"></polygon></svg
        >
        <span>Start Run</span>
      {/if}
    </Button>
  </div>

  {#if running || stats || lastResults.length > 0}
    <div class="runner-results">
      {#if running}
        <div class="progress-bar">
          <div class="progress-fill" style="width: {progress}%"></div>
        </div>
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
                    <span class="badge badge-error">ERROR</span>
                  {:else if res.response}
                    <span class="badge {getStatusClass(res.response.statusCode)}">
                      {res.response.statusCode}
                    </span>
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
    <div class="empty-runner">
      <svg
        width="48"
        height="48"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
        ><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon></svg
      >
      <h3>Parallel Runner</h3>
      <p>Configure concurrency and iterations to perform load testing on this request.</p>
    </div>
  {/if}
</div>
