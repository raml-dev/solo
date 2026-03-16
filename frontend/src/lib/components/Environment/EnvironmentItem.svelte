<script lang="ts">
  import Button from "flowbite-svelte/Button.svelte";
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

<div
  class="relative flex items-center gap-1.5 px-2 py-1.5 {isFocused ? 'bg-primary-50 dark:bg-primary-900/30' : 'hover:bg-neutral-100 dark:hover:bg-neutral-800'}"
>
  <input
    type="radio"
    name="active-environment"
    checked={isActive}
    onchange={activateEnvironment}
    aria-label={`Set ${env.name} as active environment`}
    class="shrink-0 accent-primary-600"
  />
  {#if env.gitRemote}
    <svg
      width="10"
      height="10"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="2"
      class="shrink-0 text-neutral-400"
      aria-label={`Git remote: ${env.gitRemote}`}
    >
      <path d={getProviderIconPath(env.gitProvider || "git")} />
    </svg>
  {/if}
  <Button
    color="light"
    size="sm"
    class="min-w-0 flex-1 justify-start truncate border-0 shadow-none"
    onclick={openEnvironment}
  >
    {env.name}
  </Button>
  <Button
    color="light"
    size="xs"
    class="shrink-0 border-0 shadow-none"
    onclick={toggleMenu}
    title="More actions"
    aria-label="More actions"
  >
    •••
  </Button>

  {#if menuOpen}
    <div
      class="absolute right-0 top-full z-50 min-w-40 rounded-lg border border-neutral-200 bg-white py-1 shadow-lg dark:border-neutral-700 dark:bg-neutral-800"
    >
      {#if env.gitRemote}
        <Button
          color="light"
          size="sm"
          class="w-full justify-start border-0 shadow-none"
          onclick={handleGitStatus}
        >
          Git status
        </Button>
        <Button
          color="light"
          size="sm"
          class="w-full justify-start border-0 shadow-none disabled:opacity-50"
          onclick={handleSync}
          disabled={isSyncing}
        >
          {isSyncing ? "Syncing…" : "Sync with Git"}
        </Button>
        <div class="my-1 border-t border-neutral-200 dark:border-neutral-700"></div>
      {/if}
      <Button
        color="light"
        size="sm"
        class="w-full justify-start border-0 shadow-none text-danger-600 hover:bg-danger-50 dark:text-danger-400 dark:hover:bg-danger-900/20"
        onclick={handleDeleteEnvironment}
      >
        Delete
      </Button>
    </div>
  {/if}
</div>
