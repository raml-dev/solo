<script lang="ts">
  import type { HistoryEntry } from "../../stores/historyStore";

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

<style>
  .entry {
    border-bottom: 1px solid var(--border);
  }
  .entry.expanded {
    background: var(--bg-tertiary, var(--bg-secondary));
  }

  .entry-row {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
    width: 100%;
    padding: var(--space-xs) var(--space-md);
    background: none;
    border: none;
    cursor: pointer;
    text-align: left;
    color: var(--text);
    font-size: var(--font-size-sm);
    min-height: 30px;
  }
  .entry-row:hover {
    background: var(--bg-tertiary, var(--bg-secondary));
  }

  .entry-chevron {
    color: var(--text-muted);
    font-size: 0.65rem;
    width: 10px;
    flex-shrink: 0;
  }

  .entry-time {
    color: var(--text-muted);
    font-family: var(--font-mono);
    font-size: 0.7rem;
    flex-shrink: 0;
    width: 72px;
  }

  .badge {
    flex-shrink: 0;
    padding: 1px var(--space-xs);
    border-radius: var(--radius-sm);
    font-size: 0.68rem;
    font-weight: var(--font-weight-semibold);
    font-family: var(--font-mono);
  }

  .badge-method.method-get {
    color: var(--success);
  }
  .badge-method.method-post {
    color: var(--warning);
  }
  .badge-method.method-put {
    color: var(--info);
  }
  .badge-method.method-patch {
    color: var(--primary);
  }
  .badge-method.method-delete {
    color: var(--danger);
  }

  .badge-status.status-success {
    background: color-mix(in srgb, var(--success) 15%, transparent);
    color: var(--success);
  }
  .badge-status.status-info {
    background: color-mix(in srgb, var(--info) 15%, transparent);
    color: var(--info);
  }
  .badge-status.status-warning {
    background: color-mix(in srgb, var(--warning) 15%, transparent);
    color: var(--warning);
  }
  .badge-status.status-error {
    background: color-mix(in srgb, var(--danger) 15%, transparent);
    color: var(--danger);
  }

  .entry-url {
    flex: 1;
    font-family: var(--font-mono);
    font-size: 0.75rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: var(--text);
  }

  .entry-duration {
    flex-shrink: 0;
    font-family: var(--font-mono);
    font-size: 0.7rem;
    color: var(--text-muted);
  }

  .entry-collection {
    flex-shrink: 0;
    font-size: 0.68rem;
    color: var(--text-muted);
    max-width: 160px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  /* Detail panel */
  .entry-detail {
    padding: var(--space-sm) var(--space-md) var(--space-md);
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
  }

  .detail-tabs {
    display: flex;
    gap: 2px;
    border-bottom: 1px solid var(--border);
    margin-bottom: var(--space-sm);
  }

  .detail-tab {
    background: none;
    border: none;
    border-bottom: 2px solid transparent;
    margin-bottom: -1px;
    padding: var(--space-xs) var(--space-md);
    font-size: var(--font-size-sm);
    color: var(--text-muted);
    cursor: pointer;
  }
  .detail-tab.active {
    color: var(--text);
    border-bottom-color: var(--primary);
  }

  .detail-section {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .detail-error .detail-pre {
    color: var(--danger);
  }

  .detail-label {
    margin: 0;
    font-size: 0.68rem;
    font-weight: var(--font-weight-semibold);
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .detail-pre {
    margin: 0;
    font-family: var(--font-mono);
    font-size: 0.72rem;
    color: var(--text);
    white-space: pre-wrap;
    word-break: break-word;
    background: var(--bg-primary);
    padding: var(--space-sm) var(--space-md);
    border-radius: var(--radius-sm);
    border: 1px solid var(--border);
    max-height: 200px;
    overflow-y: auto;
  }
</style>
