<script lang="ts">
  import Button from "flowbite-svelte/Button.svelte";
  import Modal from "flowbite-svelte/Modal.svelte";
  import ToastContainer from "$src/lib/components/base/ToastContainer.svelte";
  import { modalStack, topModalId } from "$src/lib/stores/modalStackStore";
  import { notifications } from "$src/lib/stores/notificationStore";
  import { onDestroy, onMount } from "svelte";

  interface Props {
    // ── Props ────────────────────────────────────────────────────────────────
    entityId: string;
    entityName: string;
    // Caller provides the backend functions — works for both collections and environments.
    fnGetStatus: (id: string) => Promise<CollectionStatus>;
    fnSync: (id: string) => Promise<void>;
    fnKeepOurs: (id: string) => Promise<void>;
    fnKeepTheirs: (id: string) => Promise<void>;
    fnAbortRebase: (id: string) => Promise<void>;
    fnDiscard: (id: string) => Promise<void>;
    fnOpenTerminal: (id: string) => Promise<void>;
    // Called after a successful action so the parent can reload its store.
    onReload?: () => Promise<void>;
    onClose?: () => void;
  }

  let {
    entityId,
    entityName,
    fnGetStatus,
    fnSync,
    fnKeepOurs,
    fnKeepTheirs,
    fnAbortRebase,
    fnDiscard,
    fnOpenTerminal,
    onReload = async () => {},
    onClose
  }: Props = $props();

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

  // ── State ────────────────────────────────────────────────────────────────
  let status: CollectionStatus | null = $state(null);
  let loading = $state(true);
  let actionInProgress = $state(false);
  let lastOutput = $state("");
  let showDiscardConfirm = $state(false);
  let errorMessage = $state("");
  let open = $state(true);

  const gitStatusPanelModalId = `git-status-${Math.random().toString(36).slice(2)}`;

  // ── Lifecycle ────────────────────────────────────────────────────────────
  onMount(() => {
    refresh();
  });

  $effect(() => {
    if (open) {
      modalStack.open(gitStatusPanelModalId);
    } else {
      modalStack.close(gitStatusPanelModalId);
      onClose?.();
    }
  });

  onDestroy(() => {
    modalStack.close(gitStatusPanelModalId);
  });

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

  const handleSync = () => runAction("Sync", () => fnSync(entityId));
  const handleKeepOurs = () => runAction("Keep Ours", () => fnKeepOurs(entityId));
  const handleKeepTheirs = () => runAction("Keep Theirs", () => fnKeepTheirs(entityId));
  const handleAbortRebase = () => runAction("Abort Rebase", () => fnAbortRebase(entityId));
  const handleDiscard = () => {
    showDiscardConfirm = false;
    runAction("Discard Changes", () => fnDiscard(entityId));
  };

  async function handleOpenTerminal() {
    try {
      await fnOpenTerminal(entityId);
    } catch (err) {
      notifications.error("Failed to open terminal", String(err));
    }
  }

  function requestClose() {
    open = false;
  }

  // ── Derived helpers ──────────────────────────────────────────────────────
  function statusLabel(s: CollectionStatus): string {
    if (s.isRebaseInProgress) return "Rebase in progress";
    if (s.hasConflicts) return "Conflict";
    if (s.ahead > 0 && s.behind > 0) return `↑${s.ahead} ↓${s.behind}`;
    if (s.ahead > 0) return `↑${s.ahead} ahead`;
    if (s.behind > 0) return `↓${s.behind} behind`;
    if (s.isDirty) return "Uncommitted changes";
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

<Modal bind:open title="Git Status" size="xl" onclose={requestClose}>
  {#if $topModalId === gitStatusPanelModalId}
    <ToastContainer />
  {/if}

  {#snippet header()}
    <div class="panel-header">
      <div class="header-left">
        <svg
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          class="git-icon"
        >
          <circle cx="12" cy="18" r="3" /><circle cx="6" cy="6" r="3" /><circle
            cx="18"
            cy="6"
            r="3"
          />
          <path d="M18 9v2a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2V9" />
          <line x1="12" y1="12" x2="12" y2="15" />
        </svg>
        <div>
          <h2 class="panel-title">{entityName}</h2>
          {#if status}
            <span class="branch-label">⎇ {status.branch}</span>
          {/if}
        </div>
      </div>
      <div class="header-right">
        <button
          class="icon-btn"
          title="Refresh"
          onclick={refresh}
          disabled={loading || actionInProgress}
        >
          <svg
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2.5"
            class:spinning={loading}
          >
            <path
              d="M21 2v6h-6M3 22v-6h6M21 12c0 4.97-4.03 9-9 9-3.32 0-6.23-1.8-7.81-4.47M3 12c0-4.97 4.03-9 9-9 3.32 0 6.23 1.8 7.81 4.47"
            />
          </svg>
        </button>
        <button class="icon-btn" title="Open in Terminal" onclick={handleOpenTerminal}>
          <svg
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <polyline points="4 17 10 11 4 5" /><line x1="12" y1="19" x2="20" y2="19" />
          </svg>
        </button>
        <button class="icon-btn close-btn" onclick={requestClose} title="Close">✕</button>
      </div>
    </div>
  {/snippet}

  <div class="panel-body">
    {#if loading}
      <div class="loading-state">
        <div class="spinner"></div>
        <span>Loading git status…</span>
      </div>
    {:else if errorMessage}
      <div class="error-box">
        <svg
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
        >
          <circle cx="12" cy="12" r="10" /><line x1="12" y1="8" x2="12" y2="12" /><line
            x1="12"
            y1="16"
            x2="12.01"
            y2="16"
          />
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
            <svg
              width="14"
              height="14"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2.5"
            >
              <path
                d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"
              />
              <line x1="12" y1="9" x2="12" y2="13" /><line x1="12" y1="17" x2="12.01" y2="17" />
            </svg>
            <strong
              >{status.isRebaseInProgress
                ? "Rebase in progress"
                : "Merge conflicts detected"}</strong
            >
          </div>
          {#if status.conflictFiles?.length > 0}
            <ul class="conflict-files">
              {#each status.conflictFiles as f (f)}<li>⚡ {f}</li>{/each}
            </ul>
          {/if}
        </div>
      {/if}

      <!-- Changed files -->
      {#if status.statusLines?.length > 0}
        <section class="section">
          <h4 class="section-title">
            Changed files <span class="count">{status.statusLines.length}</span>
          </h4>
          <div class="file-list">
            {#each status.statusLines as line (line)}
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
            {#each status.recentLog as entry (entry.hash)}
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
            <Button color="red" size="sm" onclick={handleDiscard} disabled={actionInProgress}
              >Yes, discard</Button
            >
            <Button
              color="light"
              size="sm"
              onclick={() => (showDiscardConfirm = false)}
              disabled={actionInProgress}>Cancel</Button
            >
          </div>
        </div>
      {/if}
    {/if}
  </div>

  {#snippet footer()}
    <div class="panel-footer">
      <div class="actions-left">
        {#if status?.isRebaseInProgress || status?.hasConflicts}
          <Button
            color="primary"
            size="sm"
            disabled={actionInProgress || !status}
            onclick={handleKeepOurs}>Keep Ours</Button
          >
          <Button
            color="light"
            size="sm"
            disabled={actionInProgress || !status}
            onclick={handleKeepTheirs}>Keep Theirs</Button
          >
          <Button
            color="red"
            size="sm"
            disabled={actionInProgress || !status}
            onclick={handleAbortRebase}>Abort Rebase</Button
          >
        {:else}
          <Button
            color="primary"
            size="sm"
            disabled={actionInProgress || loading || !status}
            onclick={handleSync}
          >
            {actionInProgress ? "Syncing…" : "↺ Sync"}
          </Button>
          {#if status?.isDirty}
            <Button
              color="red"
              size="sm"
              disabled={actionInProgress || !status}
              onclick={() => (showDiscardConfirm = true)}
            >
              Discard Changes
            </Button>
          {/if}
        {/if}
      </div>
      <Button color="light" size="sm" onclick={requestClose}>Close</Button>
    </div>
  {/snippet}
</Modal>
