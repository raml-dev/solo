<script lang="ts">
  import Alert from "flowbite-svelte/Alert.svelte";
  import Badge from "flowbite-svelte/Badge.svelte";
  import Button from "flowbite-svelte/Button.svelte";
  import Modal from "flowbite-svelte/Modal.svelte";
  import Spinner from "flowbite-svelte/Spinner.svelte";
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
    <div class="flex items-center justify-between gap-2">
      <div class="flex min-w-0 flex-1 items-center gap-2">
        <svg
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          class="shrink-0 text-neutral-500 dark:text-neutral-400"
        >
          <circle cx="12" cy="18" r="3" /><circle cx="6" cy="6" r="3" /><circle cx="18" cy="6" r="3" />
          <path d="M18 9v2a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2V9" />
          <line x1="12" y1="12" x2="12" y2="15" />
        </svg>
        <div>
          <h2 class="text-sm font-semibold text-neutral-800 dark:text-neutral-100">{entityName}</h2>
          {#if status}
            <span class="font-mono text-xs text-neutral-500 dark:text-neutral-400">⎇ {status.branch}</span>
          {/if}
        </div>
      </div>
      <div class="flex shrink-0 items-center gap-1">
        <Button
          color="light"
          size="xs"
          title="Refresh"
          onclick={refresh}
          disabled={loading || actionInProgress}
        >
          {#if loading}
            <Spinner size="4" />
          {:else}
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
              <path d="M21 2v6h-6M3 22v-6h6M21 12c0 4.97-4.03 9-9 9-3.32 0-6.23-1.8-7.81-4.47M3 12c0-4.97 4.03-9 9-9 3.32 0 6.23 1.8 7.81 4.47" />
            </svg>
          {/if}
        </Button>
        <Button color="light" size="xs" title="Open in Terminal" onclick={handleOpenTerminal}>
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="4 17 10 11 4 5" /><line x1="12" y1="19" x2="20" y2="19" />
          </svg>
        </Button>
        <Button color="light" size="xs" title="Close" onclick={requestClose}>✕</Button>
      </div>
    </div>
  {/snippet}

  <div class="flex flex-col gap-4">
    {#if loading}
      <div class="flex items-center gap-2 py-6 text-sm text-neutral-500 dark:text-neutral-400">
        <Spinner size="4" />
        <span>Loading git status…</span>
      </div>
    {:else if errorMessage}
      <Alert color="red">
        <span>{errorMessage}</span>
      </Alert>
    {:else if status}
      <!-- Status badge -->
      <div class="flex flex-wrap items-center gap-3">
        {#if statusVariant(status) === "ok"}
          <Badge color="green">{statusLabel(status)}</Badge>
        {:else if statusVariant(status) === "warning"}
          <Badge color="yellow">{statusLabel(status)}</Badge>
        {:else}
          <Badge color="red">{statusLabel(status)}</Badge>
        {/if}
        {#if status.ahead > 0 || status.behind > 0}
          <span class="flex items-center gap-3 text-xs">
            {#if status.ahead > 0}<span class="text-primary-600 dark:text-primary-400">↑ {status.ahead} to push</span>{/if}
            {#if status.behind > 0}<span class="text-neutral-500 dark:text-neutral-400">↓ {status.behind} to pull</span>{/if}
          </span>
        {/if}
      </div>

      <!-- Conflict / rebase warning -->
      {#if status.isRebaseInProgress || status.hasConflicts}
        <Alert color="red" class="mt-1">
          <div class="flex items-center gap-2 font-medium">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
              <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
              <line x1="12" y1="9" x2="12" y2="13" /><line x1="12" y1="17" x2="12.01" y2="17" />
            </svg>
            <strong>{status.isRebaseInProgress ? "Rebase in progress" : "Merge conflicts detected"}</strong>
          </div>
          {#if status.conflictFiles?.length > 0}
            <ul class="mt-2 list-none space-y-1 font-mono text-sm">
              {#each status.conflictFiles as f (f)}<li>⚡ {f}</li>{/each}
            </ul>
          {/if}
        </Alert>
      {/if}

      <!-- Changed files -->
      {#if status.statusLines?.length > 0}
        <section class="flex flex-col gap-2">
          <h4 class="text-xs font-semibold uppercase tracking-wide text-neutral-500 dark:text-neutral-400">
            Changed files <Badge color="gray" class="ml-1">{status.statusLines.length}</Badge>
          </h4>
          <div class="space-y-0.5">
            {#each status.statusLines as line (line)}
              <div class="flex items-baseline gap-2 rounded px-1 py-0.5 font-mono text-xs hover:bg-neutral-100 dark:hover:bg-neutral-800">
                <span class="shrink-0">{statusLineIcon(line)}</span>
                <span class="w-5 shrink-0 font-bold text-neutral-400">{line.slice(0, 2)}</span>
                <span class="min-w-0 truncate text-neutral-700 dark:text-neutral-300">{line.slice(3)}</span>
              </div>
            {/each}
          </div>
        </section>
      {/if}

      <!-- Recent log -->
      {#if status.recentLog?.length > 0}
        <section class="flex flex-col gap-2">
          <h4 class="text-xs font-semibold uppercase tracking-wide text-neutral-500 dark:text-neutral-400">Recent commits</h4>
          <div class="space-y-2">
            {#each status.recentLog as entry (entry.hash)}
              <div class="flex flex-col gap-0.5 rounded border border-neutral-200 p-2 dark:border-neutral-700">
                <code class="font-mono text-xs text-neutral-400">{entry.hash}</code>
                <span class="text-sm text-neutral-800 dark:text-neutral-100">{entry.message}</span>
                <span class="text-xs text-neutral-500 dark:text-neutral-400">{entry.author} · {entry.date}</span>
              </div>
            {/each}
          </div>
        </section>
      {/if}

      <!-- Output feedback -->
      {#if lastOutput}
        <Alert color="green" class="mt-1">{lastOutput}</Alert>
      {/if}

      <!-- Discard confirm inline -->
      {#if showDiscardConfirm}
        <div class="rounded-lg border border-warning-200 bg-warning-50 p-3 text-sm dark:border-warning-700 dark:bg-warning-900/20">
          <span>⚠ This will permanently discard all local uncommitted changes. Continue?</span>
          <div class="mt-2 flex items-center gap-2">
            <Button color="red" size="sm" onclick={handleDiscard} disabled={actionInProgress}>Yes, discard</Button>
            <Button color="light" size="sm" onclick={() => (showDiscardConfirm = false)} disabled={actionInProgress}>Cancel</Button>
          </div>
        </div>
      {/if}
    {/if}
  </div>

  {#snippet footer()}
    <div class="flex w-full items-center justify-between gap-2">
      <div class="flex items-center gap-2">
        {#if status?.isRebaseInProgress || status?.hasConflicts}
          <Button color="primary" size="sm" disabled={actionInProgress || !status} onclick={handleKeepOurs}>Keep Ours</Button>
          <Button color="light" size="sm" disabled={actionInProgress || !status} onclick={handleKeepTheirs}>Keep Theirs</Button>
          <Button color="red" size="sm" disabled={actionInProgress || !status} onclick={handleAbortRebase}>Abort Rebase</Button>
        {:else}
          <Button color="primary" size="sm" disabled={actionInProgress || loading || !status} onclick={handleSync}>
            {actionInProgress ? "Syncing…" : "↺ Sync"}
          </Button>
          {#if status?.isDirty}
            <Button color="red" size="sm" disabled={actionInProgress || !status} onclick={() => (showDiscardConfirm = true)}>Discard Changes</Button>
          {/if}
        {/if}
      </div>
      <Button color="light" size="sm" onclick={requestClose}>Close</Button>
    </div>
  {/snippet}
</Modal>
