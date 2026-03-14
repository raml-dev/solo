<script lang="ts">
  import { onMount, createEventDispatcher } from "svelte";
  import { notifications } from "../stores/notificationStore";
  import Button from "./base/Button.svelte";

  // ── Props ────────────────────────────────────────────────────────────────
  export let entityId: string;
  export let entityName: string;

  // Caller provides the backend functions — works for both collections and environments.
  export let fnGetStatus:    (id: string) => Promise<CollectionStatus>;
  export let fnSync:         (id: string) => Promise<void>;
  export let fnKeepOurs:     (id: string) => Promise<void>;
  export let fnKeepTheirs:   (id: string) => Promise<void>;
  export let fnAbortRebase:  (id: string) => Promise<void>;
  export let fnDiscard:      (id: string) => Promise<void>;
  export let fnOpenTerminal: (id: string) => Promise<void>;
  // Called after a successful action so the parent can reload its store.
  export let onReload: () => Promise<void> = async () => {};

  // ── Types ────────────────────────────────────────────────────────────────
  interface GitLogEntry {
    hash: string;
    author: string;
    message: string;
    date: string;
  }

  interface CollectionStatus {
    branch: string;
    isRebaseInProgress: boolean;
    hasConflicts: boolean;
    conflictFiles: string[];
    statusLines: string[];
    isDirty: boolean;
    ahead: number;
    behind: number;
    recentLog: GitLogEntry[];
  }

  const dispatch = createEventDispatcher();

  // ── State ────────────────────────────────────────────────────────────────
  let status: CollectionStatus | null = null;
  let loading = true;
  let actionInProgress = false;
  let lastOutput = "";
  let showDiscardConfirm = false;
  let errorMessage = "";

  // ── Lifecycle ────────────────────────────────────────────────────────────
  onMount(() => { refresh(); });

  async function refresh() {
    loading = true;
    errorMessage = "";
    try {
      status = await fnGetStatus(entityId);
    } catch (err) {
      errorMessage = String(err);
    } finally {
      loading = false;
    }
  }

  // ── Actions ──────────────────────────────────────────────────────────────
  async function runAction(label: string, fn: () => Promise<void>) {
    actionInProgress = true;
    lastOutput = "";
    errorMessage = "";
    try {
      await fn();
      lastOutput = `✓ ${label} completed successfully.`;
      await onReload();
      await refresh();
    } catch (err) {
      errorMessage = String(err);
      lastOutput = "";
    } finally {
      actionInProgress = false;
    }
  }

  const handleSync        = () => runAction("Sync",          () => fnSync(entityId));
  const handleKeepOurs    = () => runAction("Keep Ours",     () => fnKeepOurs(entityId));
  const handleKeepTheirs  = () => runAction("Keep Theirs",   () => fnKeepTheirs(entityId));
  const handleAbortRebase = () => runAction("Abort Rebase",  () => fnAbortRebase(entityId));
  const handleDiscard     = () => { showDiscardConfirm = false; runAction("Discard Changes", () => fnDiscard(entityId)); };

  async function handleOpenTerminal() {
    try {
      await fnOpenTerminal(entityId);
    } catch (err) {
      notifications.error("Failed to open terminal", String(err));
    }
  }

  // ── Derived helpers ──────────────────────────────────────────────────────
  function statusLabel(s: CollectionStatus): string {
    if (s.isRebaseInProgress) return "Rebase in progress";
    if (s.hasConflicts)       return "Conflict";
    if (s.ahead > 0 && s.behind > 0) return `↑${s.ahead} ↓${s.behind}`;
    if (s.ahead > 0)   return `↑${s.ahead} ahead`;
    if (s.behind > 0)  return `↓${s.behind} behind`;
    if (s.isDirty)     return "Uncommitted changes";
    return "In sync";
  }

  function statusVariant(s: CollectionStatus): string {
    if (s.isRebaseInProgress || s.hasConflicts) return "danger";
    if (s.ahead > 0 || s.behind > 0 || s.isDirty) return "warning";
    return "ok";
  }

  function statusLineIcon(line: string): string {
    if (!line || line.length < 2) return "·";
    const xy = line.slice(0, 2);
    if (xy === "UU" || xy === "AA" || xy === "DD" || xy === "AU" || xy === "UA") return "⚡";
    if (xy[0] === "M" || xy[1] === "M") return "✎";
    if (xy[0] === "A" || xy[0] === "?") return "+";
    if (xy[0] === "D") return "−";
    return "·";
  }
</script>

<!-- ── Overlay ──────────────────────────────────────────────────────────── -->
<div class="overlay" role="presentation" on:click|self={() => dispatch("close")}>
  <div class="panel" role="dialog" aria-modal="true" aria-label="Git Status">

    <!-- Header -->
    <header class="panel-header">
      <div class="header-left">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="git-icon">
          <circle cx="12" cy="18" r="3"/><circle cx="6" cy="6" r="3"/><circle cx="18" cy="6" r="3"/>
          <path d="M18 9v2a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2V9"/>
          <line x1="12" y1="12" x2="12" y2="15"/>
        </svg>
        <div>
          <h2 class="panel-title">{entityName}</h2>
          {#if status}
            <span class="branch-label">⎇ {status.branch}</span>
          {/if}
        </div>
      </div>
      <div class="header-right">
        <button class="icon-btn" title="Refresh" on:click={refresh} disabled={loading || actionInProgress}>
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" class:spinning={loading}>
            <path d="M21 2v6h-6M3 22v-6h6M21 12c0 4.97-4.03 9-9 9-3.32 0-6.23-1.8-7.81-4.47M3 12c0-4.97 4.03-9 9-9 3.32 0 6.23 1.8 7.81 4.47"/>
          </svg>
        </button>
        <button class="icon-btn" title="Open in Terminal" on:click={handleOpenTerminal}>
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/>
          </svg>
        </button>
        <button class="icon-btn close-btn" on:click={() => dispatch("close")} title="Close">✕</button>
      </div>
    </header>

    <div class="panel-body">
      {#if loading}
        <div class="loading-state">
          <div class="spinner" />
          <span>Loading git status…</span>
        </div>

      {:else if errorMessage}
        <div class="error-box">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
          </svg>
          <span>{errorMessage}</span>
        </div>

      {:else if status}
        <!-- Status badge -->
        <div class="status-row">
          <span class="status-badge status-{statusVariant(status)}">{statusLabel(status)}</span>
          {#if status.ahead > 0 || status.behind > 0}
            <span class="ahead-behind-detail">
              {#if status.ahead > 0}<span class="ahead">↑ {status.ahead} to push</span>{/if}
              {#if status.behind > 0}<span class="behind">↓ {status.behind} to pull</span>{/if}
            </span>
          {/if}
        </div>

        <!-- Conflict / rebase warning -->
        {#if status.isRebaseInProgress || status.hasConflicts}
          <div class="conflict-box">
            <div class="conflict-header">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>
                <line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/>
              </svg>
              <strong>{status.isRebaseInProgress ? "Rebase in progress" : "Merge conflicts detected"}</strong>
            </div>
            {#if status.conflictFiles?.length > 0}
              <ul class="conflict-files">
                {#each status.conflictFiles as f}<li>⚡ {f}</li>{/each}
              </ul>
            {/if}
          </div>
        {/if}

        <!-- Changed files -->
        {#if status.statusLines?.length > 0}
          <section class="section">
            <h4 class="section-title">Changed files <span class="count">{status.statusLines.length}</span></h4>
            <div class="file-list">
              {#each status.statusLines as line}
                <div class="file-line">
                  <span class="file-icon">{statusLineIcon(line)}</span>
                  <span class="file-xy">{line.slice(0, 2)}</span>
                  <span class="file-name">{line.slice(3)}</span>
                </div>
              {/each}
            </div>
          </section>
        {/if}

        <!-- Recent log -->
        {#if status.recentLog?.length > 0}
          <section class="section">
            <h4 class="section-title">Recent commits</h4>
            <div class="log-list">
              {#each status.recentLog as entry}
                <div class="log-entry">
                  <code class="log-hash">{entry.hash}</code>
                  <span class="log-msg">{entry.message}</span>
                  <span class="log-meta">{entry.author} · {entry.date}</span>
                </div>
              {/each}
            </div>
          </section>
        {/if}

        <!-- Output feedback -->
        {#if lastOutput}
          <div class="output-box ok">{lastOutput}</div>
        {/if}

        <!-- Discard confirm inline -->
        {#if showDiscardConfirm}
          <div class="confirm-box">
            <span>⚠ This will permanently discard all local uncommitted changes. Continue?</span>
            <div class="confirm-actions">
              <Button variant="danger" size="sm" click={handleDiscard} disabled={actionInProgress}>Yes, discard</Button>
              <Button variant="secondary" size="sm" click={() => (showDiscardConfirm = false)} disabled={actionInProgress}>Cancel</Button>
            </div>
          </div>
        {/if}
      {/if}
    </div>

    <!-- Footer actions -->
    <footer class="panel-footer">
      <div class="actions-left">
        {#if status?.isRebaseInProgress || status?.hasConflicts}
          <Button variant="primary"   size="sm" disabled={actionInProgress || !status} click={handleKeepOurs}>Keep Ours</Button>
          <Button variant="secondary" size="sm" disabled={actionInProgress || !status} click={handleKeepTheirs}>Keep Theirs</Button>
          <Button variant="danger"    size="sm" disabled={actionInProgress || !status} click={handleAbortRebase}>Abort Rebase</Button>
        {:else}
          <Button variant="primary" size="sm" disabled={actionInProgress || loading || !status} click={handleSync}>
            {actionInProgress ? "Syncing…" : "↺ Sync"}
          </Button>
          {#if status?.isDirty}
            <Button variant="danger" size="sm" disabled={actionInProgress || !status} click={() => (showDiscardConfirm = true)}>
              Discard Changes
            </Button>
          {/if}
        {/if}
      </div>
      <Button variant="secondary" size="sm" click={() => dispatch("close")}>Close</Button>
    </footer>

  </div>
</div>

<style>
  .overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.55);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1100;
    backdrop-filter: blur(2px);
  }

  .panel {
    background: var(--bg-primary);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    width: 100%;
    max-width: 520px;
    max-height: 80vh;
    display: flex;
    flex-direction: column;
    box-shadow: var(--shadow-xl);
    overflow: hidden;
  }

  /* Header */
  .panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-md) var(--space-lg);
    border-bottom: 1px solid var(--border);
    background: var(--bg-secondary);
    gap: var(--space-sm);
  }

  .header-left  { display: flex; align-items: center; gap: var(--space-sm); min-width: 0; }
  .header-right { display: flex; align-items: center; gap: var(--space-xs); flex-shrink: 0; }

  .git-icon   { flex-shrink: 0; color: var(--text-muted); }
  .panel-title {
    margin: 0;
    font-size: 0.95rem;
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .branch-label { font-size: 0.7rem; color: var(--text-muted); font-family: var(--font-mono); }

  .icon-btn {
    background: none;
    border: none;
    cursor: pointer;
    padding: 4px;
    border-radius: var(--radius-sm);
    color: var(--text-muted);
    display: flex;
    align-items: center;
    justify-content: center;
    transition: background 0.15s;
  }
  .icon-btn:hover    { background: var(--bg-tertiary); color: var(--text); }
  .icon-btn:disabled { opacity: 0.4; cursor: not-allowed; }
  .close-btn         { font-size: 1rem; }

  /* Body */
  .panel-body {
    flex: 1;
    overflow-y: auto;
    padding: var(--space-md) var(--space-lg);
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
  }

  .loading-state {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
    color: var(--text-muted);
    padding: var(--space-xl) 0;
    justify-content: center;
    font-size: 0.875rem;
  }

  .spinner {
    width: 16px; height: 16px;
    border: 2px solid var(--border);
    border-top-color: var(--primary);
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
    flex-shrink: 0;
  }

  @keyframes spin { to { transform: rotate(360deg); } }
  .spinning { animation: spin 0.8s linear infinite; }

  /* Status badge */
  .status-row { display: flex; align-items: center; gap: var(--space-sm); flex-wrap: wrap; }

  .status-badge {
    font-size: 0.7rem;
    font-weight: 700;
    padding: 3px 8px;
    border-radius: 20px;
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }
  .status-ok      { background: color-mix(in srgb, var(--success) 15%, transparent); color: var(--success); border: 1px solid color-mix(in srgb, var(--success) 30%, transparent); }
  .status-warning { background: color-mix(in srgb, var(--warning, #f59e0b) 15%, transparent); color: var(--warning, #f59e0b); border: 1px solid color-mix(in srgb, var(--warning, #f59e0b) 30%, transparent); }
  .status-danger  { background: color-mix(in srgb, var(--danger) 15%, transparent); color: var(--danger); border: 1px solid color-mix(in srgb, var(--danger) 30%, transparent); }

  .ahead-behind-detail { display: flex; gap: var(--space-sm); font-size: 0.75rem; }
  .ahead  { color: var(--success); }
  .behind { color: var(--warning, #f59e0b); }

  /* Conflict box */
  .conflict-box {
    background: color-mix(in srgb, var(--danger) 8%, transparent);
    border: 1px solid color-mix(in srgb, var(--danger) 25%, transparent);
    border-radius: var(--radius-md);
    padding: var(--space-sm) var(--space-md);
  }
  .conflict-header { display: flex; align-items: center; gap: var(--space-xs); color: var(--danger); font-size: 0.8rem; }
  .conflict-files {
    margin: var(--space-xs) 0 0 var(--space-md);
    padding: 0;
    list-style: none;
    font-size: 0.75rem;
    font-family: var(--font-mono);
    color: var(--text-muted);
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  /* Sections */
  .section { display: flex; flex-direction: column; gap: var(--space-xs); }
  .section-title {
    margin: 0;
    font-size: 0.7rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-muted);
    display: flex;
    align-items: center;
    gap: var(--space-xs);
  }
  .count { background: var(--bg-tertiary); border-radius: 10px; padding: 0 5px; font-size: 0.65rem; }

  /* File list */
  .file-list { display: flex; flex-direction: column; gap: 2px; max-height: 120px; overflow-y: auto; }
  .file-line {
    display: flex;
    align-items: center;
    gap: var(--space-xs);
    font-family: var(--font-mono);
    font-size: 0.72rem;
    padding: 2px 4px;
    border-radius: var(--radius-sm);
  }
  .file-line:hover { background: var(--bg-secondary); }
  .file-icon { width: 14px; text-align: center; flex-shrink: 0; }
  .file-xy   { color: var(--text-muted); min-width: 18px; font-weight: 600; }
  .file-name { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: var(--text); }

  /* Log list */
  .log-list { display: flex; flex-direction: column; gap: 6px; max-height: 140px; overflow-y: auto; }
  .log-entry {
    display: grid;
    grid-template-columns: auto 1fr;
    grid-template-rows: auto auto;
    column-gap: var(--space-sm);
    row-gap: 1px;
    font-size: 0.75rem;
    padding: 4px 6px;
    border-radius: var(--radius-sm);
  }
  .log-entry:hover { background: var(--bg-secondary); }
  .log-hash {
    grid-row: 1 / 3;
    font-family: var(--font-mono);
    font-size: 0.7rem;
    background: var(--bg-tertiary);
    padding: 1px 5px;
    border-radius: var(--radius-sm);
    color: var(--text-muted);
    align-self: center;
    white-space: nowrap;
  }
  .log-msg  { color: var(--text); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .log-meta { color: var(--text-muted); font-size: 0.68rem; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

  /* Feedback boxes */
  .error-box {
    display: flex;
    align-items: flex-start;
    gap: var(--space-sm);
    background: color-mix(in srgb, var(--danger) 10%, transparent);
    border: 1px solid color-mix(in srgb, var(--danger) 25%, transparent);
    border-radius: var(--radius-md);
    padding: var(--space-sm) var(--space-md);
    font-size: 0.75rem;
    color: var(--danger);
    word-break: break-word;
  }

  .output-box { font-size: 0.75rem; padding: var(--space-sm) var(--space-md); border-radius: var(--radius-md); font-family: var(--font-mono); }
  .output-box.ok {
    background: color-mix(in srgb, var(--success) 10%, transparent);
    border: 1px solid color-mix(in srgb, var(--success) 25%, transparent);
    color: var(--success);
  }

  /* Discard confirm */
  .confirm-box {
    background: color-mix(in srgb, var(--danger) 8%, transparent);
    border: 1px solid color-mix(in srgb, var(--danger) 25%, transparent);
    border-radius: var(--radius-md);
    padding: var(--space-sm) var(--space-md);
    font-size: 0.8rem;
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
    color: var(--text);
  }
  .confirm-actions { display: flex; gap: var(--space-sm); }

  /* Footer */
  .panel-footer {
    padding: var(--space-md) var(--space-lg);
    border-top: 1px solid var(--border);
    background: var(--bg-secondary);
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: var(--space-sm);
  }
  .actions-left { display: flex; gap: var(--space-sm); flex-wrap: wrap; }
</style>
