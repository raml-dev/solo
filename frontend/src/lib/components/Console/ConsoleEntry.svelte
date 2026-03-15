<script lang="ts">
  import type { HistoryEntry } from "$src/lib/stores/historyStore";

  interface Props {
    entry: HistoryEntry;
  }

  let { entry }: Props = $props();

  let expanded = $state(false);
  let detailTab: "request" | "response" = $state("response");

  function getStatusClass(status: number): string {
    if (status >= 200 && status < 300) return "status-success";
    if (status >= 300 && status < 400) return "status-info";
    if (status >= 400 && status < 500) return "status-warning";
    return "status-error";
  }

  function getMethodClass(method: string): string {
    return `method-${method.toLowerCase()}`;
  }

  function formatTime(ts: string): string {
    const d = new Date(ts);
    return d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", second: "2-digit" });
  }

  function truncateUrl(url: string, max = 60): string {
    return url.length > max ? url.slice(0, max) + "…" : url;
  }

  function formatHeaders(headers: Record<string, string>): string {
    return Object.entries(headers)
      .map(([k, v]) => `${k}: ${v}`)
      .join("\n");
  }
</script>

<div class="entry" class:expanded>
  <!-- Summary row -->
  <button class="entry-row" onclick={() => (expanded = !expanded)}>
    <span class="entry-chevron">{expanded ? "▾" : "▸"}</span>
    <span class="entry-time">{formatTime(entry.timestamp)}</span>
    <span class="badge badge-method {getMethodClass(entry.request.method)}"
      >{entry.request.method}</span
    >

    {#if entry.error}
      <span class="badge badge-status status-error">ERR</span>
    {:else if entry.response}
      <span class="badge badge-status {getStatusClass(entry.response.status)}"
        >{entry.response.status}</span
      >
    {/if}

    <span class="entry-url" title={entry.request.url}>{truncateUrl(entry.request.url)}</span>

    {#if entry.response}
      <span class="entry-duration">{entry.response.time}ms</span>
    {/if}

    {#if entry.collectionName}
      <span class="entry-collection"
        >{entry.collectionName}{entry.requestName ? ` / ${entry.requestName}` : ""}</span
      >
    {/if}
  </button>

  <!-- Detail panel -->
  {#if expanded}
    <div class="entry-detail">
      <div class="detail-tabs">
        <button
          class="detail-tab"
          class:active={detailTab === "request"}
          onclick={() => (detailTab = "request")}>Request</button
        >
        <button
          class="detail-tab"
          class:active={detailTab === "response"}
          onclick={() => (detailTab = "response")}>Response</button
        >
      </div>

      {#if detailTab === "request"}
        <div class="detail-section">
          <p class="detail-label">URL</p>
          <pre class="detail-pre">{entry.request.method} {entry.request.url}</pre>
        </div>
        {#if Object.keys(entry.request.headers).length > 0}
          <div class="detail-section">
            <p class="detail-label">Headers</p>
            <pre class="detail-pre">{formatHeaders(entry.request.headers)}</pre>
          </div>
        {/if}
        {#if entry.request.body}
          <div class="detail-section">
            <p class="detail-label">Body</p>
            <pre class="detail-pre">{entry.request.body}</pre>
          </div>
        {/if}
      {:else if entry.error}
        <div class="detail-section detail-error">
          <p class="detail-label">Error</p>
          <pre class="detail-pre">{entry.error}</pre>
        </div>
      {:else if entry.response}
        <div class="detail-section">
          <p class="detail-label">Status</p>
          <pre class="detail-pre">{entry.response.status} — {entry.response.time}ms</pre>
        </div>
        {#if Object.keys(entry.response.headers).length > 0}
          <div class="detail-section">
            <p class="detail-label">Headers</p>
            <pre class="detail-pre">{formatHeaders(entry.response.headers)}</pre>
          </div>
        {/if}
        {#if entry.response.body}
          <div class="detail-section">
            <p class="detail-label">Body</p>
            <pre class="detail-pre">{entry.response.body}</pre>
          </div>
        {/if}
      {/if}
    </div>
  {/if}
</div>
