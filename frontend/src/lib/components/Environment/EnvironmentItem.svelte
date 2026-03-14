<script lang="ts">
  import type { environment } from "$wails/go/models";

  interface Props {
    env: environment.Environment;
    menuOpen: boolean;
    isActive?: boolean;
    isFocused?: boolean;
    isSyncing?: boolean;
    onDelete?: (name: string) => void;
    onOpen?: (name: string) => void;
    onActivate?: (name: string) => void;
    onToggleMenu?: (name: string) => void;
    onSync?: (id: string) => void;
    onGitStatus?: (id: string) => void;
  }

  let {
    env,
    menuOpen,
    isActive = false,
    isFocused = false,
    isSyncing = false,
    onDelete,
    onOpen,
    onActivate,
    onToggleMenu,
    onSync,
    onGitStatus
  }: Props = $props();

  function openEnvironment() {
    onOpen?.(env.name);
  }

  function activateEnvironment() {
    onActivate?.(env.name);
  }

  function toggleMenu(e: Event) {
    e.stopPropagation();
    onToggleMenu?.(env.name);
  }

  function handleDeleteEnvironment(e: Event) {
    e.stopPropagation();
    onDelete?.(env.name);
  }

  function handleSync(e: Event) {
    e.stopPropagation();
    onSync?.(env.id);
  }

  function handleGitStatus(e: Event) {
    e.stopPropagation();
    onGitStatus?.(env.id);
  }

  function getProviderIconPath(provider: string) {
    switch (provider) {
      case "github":
        return "M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.003-.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z";
      case "gitlab":
        return "M12 1L9 11h6L12 1zm0 0L3 11l9 12 9-12-9-10z";
      default:
        return "M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5";
    }
  }
</script>

<div class="environment-item" class:active={isActive} class:focused={isFocused}>
  <div class="environment-info">
    <input
      class="active-radio"
      type="radio"
      name="active-environment"
      checked={isActive}
      onchange={activateEnvironment}
      aria-label={`Set ${env.name} as active environment`}
    />
    {#if env.gitRemote}
      <svg
        width="10"
        height="10"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        class="provider-icon"
        aria-label={`Git remote: ${env.gitRemote}`}
      >
        <path d={getProviderIconPath(env.gitProvider || "git")} />
      </svg>
    {/if}
    <button class="environment-name-btn" onclick={openEnvironment}>{env.name}</button>
  </div>

  <div class="environment-actions">
    {#if env.gitRemote}
      <button class="icon-btn" onclick={handleGitStatus} title="Git status & actions">
        <svg
          width="12"
          height="12"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
        >
          <circle cx="12" cy="18" r="3" /><circle cx="6" cy="6" r="3" /><circle
            cx="18"
            cy="6"
            r="3"
          />
          <path d="M18 9v2a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2V9" />
          <line x1="12" y1="12" x2="12" y2="15" />
        </svg>
      </button>
      <button
        class="icon-btn"
        onclick={handleSync}
        title="Sync with Git remote"
        disabled={isSyncing}
      >
        {#if isSyncing}
          <span class="sync-spinner"></span>
        {:else}
          <svg
            width="12"
            height="12"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <path
              d="M21 2v6h-6M3 22v-6h6M21 12c0 4.97-4.03 9-9 9-3.32 0-6.23-1.8-7.81-4.47M3 12c0-4.97 4.03-9 9-9 3.32 0 6.23 1.8 7.81 4.47"
            ></path>
          </svg>
        {/if}
      </button>
    {/if}
    <button class="icon-btn" onclick={toggleMenu} title="More actions" aria-label="More actions">
      ...
    </button>
  </div>

  {#if menuOpen}
    <div class="environment-menu">
      <button class="menu-item danger" onclick={handleDeleteEnvironment}> Delete </button>
    </div>
  {/if}
</div>

<style>
  .environment-item {
    display: flex;
    align-items: center;
    padding: var(--space-sm) var(--space-md);
    border-radius: var(--radius-md);
    border: 1px solid transparent;
    transition: background-color 0.15s;
    position: relative;
  }

  .environment-item:hover,
  .environment-item.focused {
    background-color: var(--bg-tertiary);
  }

  .environment-item.active {
    border-color: var(--primary);
    box-shadow: 0 0 0 1px color-mix(in srgb, var(--primary) 25%, transparent);
  }

  .environment-item.focused:not(.active) {
    border-color: var(--border);
  }

  .environment-info {
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: center;
    gap: var(--space-xs);
  }

  .provider-icon {
    flex-shrink: 0;
    color: var(--text-muted);
  }

  .active-radio {
    margin: 0;
    accent-color: var(--primary);
    cursor: pointer;
  }

  .environment-name-btn {
    background: none;
    border: none;
    color: var(--text);
    font: inherit;
    font-weight: var(--font-weight-medium);
    cursor: pointer;
    padding: 0;
    text-align: left;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    flex: 1;
  }

  .environment-actions {
    opacity: 0;
    pointer-events: none;
    transition: opacity 0.15s;
    display: flex;
    gap: 2px;
  }

  .environment-item:hover .environment-actions,
  .environment-item.focused .environment-actions,
  .environment-item.active .environment-actions {
    opacity: 1;
    pointer-events: auto;
  }

  .icon-btn {
    background: none;
    border: 1px solid transparent;
    cursor: pointer;
    padding: 0 var(--space-xs);
    border-radius: var(--radius-sm);
    color: var(--text-muted);
    transition: all var(--transition-fast);
    font-size: var(--font-size-sm);
    height: 24px;
  }
  .icon-btn:hover {
    background: var(--bg-tertiary);
    color: var(--text);
  }

  .icon-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .sync-spinner {
    display: block;
    width: 10px;
    height: 10px;
    border: 1.5px solid var(--border);
    border-top-color: var(--primary);
    border-radius: 50%;
    animation: spin 0.7s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .environment-menu {
    position: absolute;
    right: var(--space-sm);
    top: calc(100% + 6px);
    background: var(--bg-primary);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-md);
    display: flex;
    flex-direction: column;
    min-width: 140px;
    z-index: var(--z-dropdown);
  }

  .menu-item {
    padding: var(--space-sm) var(--space-md);
    background: none;
    border: none;
    text-align: left;
    font-size: var(--font-size-sm);
    color: var(--text);
    cursor: pointer;
  }

  .menu-item:hover {
    background: var(--bg-tertiary);
  }

  .menu-item.danger {
    color: var(--danger);
  }

  .menu-item.danger:hover {
    background: var(--status-danger-bg);
  }
</style>
