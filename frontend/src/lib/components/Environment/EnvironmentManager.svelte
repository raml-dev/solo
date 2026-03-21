<script lang="ts">
  import DropZone from "$src/lib/components/base/DropZone.svelte";
  import ToastContainer from "$src/lib/components/base/ToastContainer.svelte";
  import FeedbackEmptyState from "$src/lib/components/common/FeedbackEmptyState.svelte";
  import EnvironmentEditor from "$src/lib/components/Environment/EnvironmentEditor.svelte";
  import EnvironmentItem from "$src/lib/components/Environment/EnvironmentItem.svelte";
  import EnvironmentModals from "$src/lib/components/Environment/EnvironmentModals.svelte";
  import GitEnvImportView from "$src/lib/components/GitEnvImportView.svelte";
  import GitStatusPanel from "$src/lib/components/GitStatusPanel.svelte";
  import { environmentStore } from "$src/lib/stores/environmentStore";
  import { modalStack, topModalId } from "$src/lib/stores/modalStackStore";
  import { notifications } from "$src/lib/stores/notificationStore";
  import {
    ExportEnvironment,
    GetGitEnvironmentStatus,
    GitEnvAbortRebase,
    GitEnvDiscardChanges,
    GitEnvKeepOurs,
    GitEnvKeepTheirs,
    ImportBrunoEnvironment,
    ImportPostmanEnvironment,
    ImportSoloEnvironment,
    OpenEnvironmentInTerminal,
    SelectFile,
    SyncGitEnvironment
  } from "$wails/go/main/App";
  import { environment } from "$wails/go/models";
  import Button from "flowbite-svelte/Button.svelte";
  import Modal from "flowbite-svelte/Modal.svelte";
  import TabItem from "flowbite-svelte/TabItem.svelte";
  import Tabs from "flowbite-svelte/Tabs.svelte";
  import { onDestroy } from "svelte";
  import { SvelteSet } from "svelte/reactivity";

  let showNewEnvironmentDialog = $state(false);
  let showDeleteConfirmDialog = $state(false);
  let deleteTarget: string | null = $state(null);
  let activeMenu: string | null = $state(null);
  let showImportSelector = $state(false);
  let importActiveTab = $state("postman");
  let gitImportActionState: { loading: boolean; disabled: boolean; submit: () => void } | null =
    $state(null);
  let showOverwriteConfirmDialog = $state(false);
  let pendingImport: { format: "postman" | "bruno" | "solo"; path: string } | null = null;
  let overwriteTargetName: string | null = $state(null);
  let focusedEnvironmentName: string | null = $state(null);
  let syncingEnvironments: Set<string> = $state(new Set());
  let gitStatusEnvId: string | null = $state(null);
  let gitStatusEnvName: string | null = $state(null);

  const environmentManagerModalScope = `environment-manager-${Math.random().toString(36).slice(2)}`;
  const importEnvironmentModalId = `${environmentManagerModalScope}-import`;
  const overwriteEnvironmentModalId = `${environmentManagerModalScope}-overwrite`;

  $effect(() => {
    if (showImportSelector) {
      modalStack.open(importEnvironmentModalId);
    } else {
      modalStack.close(importEnvironmentModalId);
    }
  });

  $effect(() => {
    if (showOverwriteConfirmDialog) {
      modalStack.open(overwriteEnvironmentModalId);
    } else {
      modalStack.close(overwriteEnvironmentModalId);
    }
  });

  onDestroy(() => {
    modalStack.close(importEnvironmentModalId);
    modalStack.close(overwriteEnvironmentModalId);
  });

  let environments = $derived($environmentStore.environments);
  $effect(() => {
    const exists =
      focusedEnvironmentName && environments.some((e) => e.name === focusedEnvironmentName);
    if (!exists) {
      focusedEnvironmentName =
        $environmentStore.selectedEnvironmentName || environments[0]?.name || null;
    }
  });
  let selectedEnvironment = $derived(
    environments.find((env) => env.name === focusedEnvironmentName) || null
  );

  function openEnvironment(name: string) {
    focusedEnvironmentName = name;
  }

  function activateEnvironment(name: string) {
    environmentStore.selectEnvironment(name);
  }

  async function handleUpdateEnvironment(data: {
    name: string;
    values: Record<string, environment.ValueType>;
  }) {
    try {
      const { name, values } = data;
      const env = environments.find((e) => e.name === name);
      if (env) {
        const updated = new environment.Environment({
          ...env,
          values
        });
        await environmentStore.updateEnvironment(updated);
      }
    } catch (err) {
      notifications.error("Failed to update environment", String(err));
    }
  }

  async function handleCreateEnvironment(name: string) {
    const trimmed = name.trim();
    if (!trimmed) return;

    const exists = environments.some((env) => env.name.toLowerCase() === trimmed.toLowerCase());
    if (exists) {
      notifications.warning(`Environment "${trimmed}" already exists`);
      return;
    }

    try {
      await environmentStore.createEnvironment(trimmed);
      showNewEnvironmentDialog = false;
    } catch {
      // error already shown by store
    }
  }

  function handleDeleteEnvironment(name: string) {
    deleteTarget = name;
    showDeleteConfirmDialog = true;
    activeMenu = null;
  }

  async function confirmDelete() {
    if (!deleteTarget) return;

    try {
      await environmentStore.deleteEnvironment(deleteTarget);
      showDeleteConfirmDialog = false;
      deleteTarget = null;
    } catch {
      // error already shown by store
    }
  }

  function toggleMenu(environmentName: string) {
    activeMenu = activeMenu === environmentName ? null : environmentName;
  }

  function closeActiveMenu() {
    activeMenu = null;
  }

  async function handleSync(environmentId: string) {
    syncingEnvironments.add(environmentId);
    syncingEnvironments = new SvelteSet(syncingEnvironments);
    try {
      await SyncGitEnvironment(environmentId);
      notifications.success("Git environment synced successfully");
      await environmentStore.loadEnvironments();
    } catch (err) {
      notifications.error("Sync failed", String(err));
    } finally {
      syncingEnvironments.delete(environmentId);
      syncingEnvironments = new SvelteSet(syncingEnvironments);
    }
  }

  function handleGitStatus(environmentId: string) {
    const env = environments.find((e) => e.id === environmentId);
    if (!env) return;
    gitStatusEnvId = environmentId;
    gitStatusEnvName = env.name;
  }

  async function handleExportEnvironment(name: string) {
    try {
      await ExportEnvironment(name);
      notifications.success("Environment exported successfully");
    } catch (err) {
      notifications.error("Failed to export environment", String(err));
    }
  }

  function openImportModal() {
    importActiveTab = "postman";
    gitImportActionState = null;
    showImportSelector = true;
  }

  function parseExistingNameFromError(message: string): string | null {
    const match = message.match(/environment\s+([^\s]+)\s+already exists/i);
    return match ? match[1] : null;
  }

  async function executeImport(
    format: "postman" | "bruno" | "solo",
    path: string,
    overwrite: boolean
  ) {
    try {
      if (format === "postman") {
        await ImportPostmanEnvironment(path, overwrite);
      } else if (format === "bruno") {
        await ImportBrunoEnvironment(path, overwrite);
      } else {
        await ImportSoloEnvironment(path, overwrite);
      }
      await environmentStore.loadEnvironments();
      notifications.success("Environment imported successfully");
    } catch (err) {
      const message = String(err ?? "Failed to import environment");
      const existingName = parseExistingNameFromError(message);
      if (!overwrite && existingName) {
        pendingImport = { format, path };
        overwriteTargetName = existingName;
        showOverwriteConfirmDialog = true;
        return;
      }
      notifications.error("Failed to import environment", message);
    }
  }

  async function handleImportPostman(path?: string) {
    const filePath =
      path ?? (await SelectFile("Select Postman Environment", "*.json", "JSON Files"));
    if (!filePath) return;
    await executeImport("postman", filePath, false);
  }

  async function handleImportBruno(path?: string) {
    const filePath = path ?? (await SelectFile("Select Bruno Environment", "*.bru", "Bruno Files"));
    if (!filePath) return;
    await executeImport("bruno", filePath, false);
  }

  async function handleImportSolo(path?: string) {
    const filePath = path ?? (await SelectFile("Select solo Environment", "*.json", "JSON Files"));
    if (!filePath) return;
    await executeImport("solo", filePath, false);
  }

  async function handleSelectImportFormat(format: "postman" | "bruno" | "solo") {
    showImportSelector = false;
    if (format === "postman") {
      await handleImportPostman();
    } else if (format === "bruno") {
      await handleImportBruno();
    } else {
      await handleImportSolo();
    }
  }

  async function confirmOverwrite() {
    if (!pendingImport) return;
    const { format, path } = pendingImport;
    pendingImport = null;
    showOverwriteConfirmDialog = false;
    await executeImport(format, path, true);
  }

  function closeOverwriteConfirmDialog() {
    pendingImport = null;
    overwriteTargetName = null;
    showOverwriteConfirmDialog = false;
  }
</script>

<svelte:window onclick={closeActiveMenu} />

<div class="flex h-full overflow-hidden">
  <div class="flex w-56 shrink-0 flex-col border-r border-neutral-200 dark:border-neutral-700">
    <div
      class="flex shrink-0 items-center justify-end gap-2 border-b border-neutral-200 p-3 dark:border-neutral-700"
    >
      <Button color="light" size="sm" onclick={openImportModal}>Import</Button>
      <Button color="primary" size="sm" onclick={() => (showNewEnvironmentDialog = true)}>
        New
      </Button>
    </div>

    {#if $environmentStore.loading}
      <div class="px-3 py-2 text-sm text-neutral-500 dark:text-neutral-400">
        Loading environments...
      </div>
    {/if}

    <div class="flex-1 overflow-y-auto">
      {#each environments as environment (environment.id)}
        <EnvironmentItem
          env={environment}
          menuOpen={activeMenu === environment.name}
          isActive={environment.name === $environmentStore.selectedEnvironmentName}
          isFocused={environment.name === focusedEnvironmentName}
          isSyncing={syncingEnvironments.has(environment.id)}
          onOpen={openEnvironment}
          onActivate={activateEnvironment}
          onDelete={handleDeleteEnvironment}
          onExport={handleExportEnvironment}
          onToggleMenu={toggleMenu}
          onSync={handleSync}
          onGitStatus={handleGitStatus}
        />
      {/each}
    </div>
    {#if environments.length === 0 && !$environmentStore.loading}
      <FeedbackEmptyState
        title="No environments yet"
        detail="Create your first environment to get started"
      />
    {/if}
  </div>
  <div class="min-w-0 flex-1 overflow-y-auto p-4">
    <EnvironmentEditor env={selectedEnvironment} onUpdate={handleUpdateEnvironment} />
  </div>
</div>

<EnvironmentModals
  bind:showNewEnvironmentDialog
  bind:showDeleteConfirmDialog
  bind:deleteTarget
  onCreate={handleCreateEnvironment}
  onConfirmDelete={confirmDelete}
  onCloseNew={() => (showNewEnvironmentDialog = false)}
  onCloseDelete={() => (showDeleteConfirmDialog = false)}
/>

{#if showImportSelector}
  <Modal title="Import Environment" bind:open={showImportSelector} size="xl">
    {#if $topModalId === importEnvironmentModalId}
      <ToastContainer />
    {/if}
    <div class="flex flex-col gap-4">
      <Tabs bind:selected={importActiveTab}>
        <TabItem key="postman" title="Postman">
          <DropZone
            title="Drop your Postman environment here"
            subtitle="Supports Postman Environment JSON"
            onDrop={async (e) => {
              const paths = e.paths;
              showImportSelector = false;
              if (paths.length > 0) {
                await handleImportPostman(paths[0]);
              } else {
                await handleImportPostman();
              }
            }}
          >
            {#snippet icon()}
              <svg
                width="44"
                height="44"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="1.4"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                <polyline points="17 8 12 3 7 8" />
                <line x1="12" y1="3" x2="12" y2="15" />
              </svg>
            {/snippet}
          </DropZone>
        </TabItem>

        <TabItem key="bruno" title="Bruno">
          <DropZone
            title="Drop your Bruno environment here"
            subtitle="Supports Bruno environment .bru files"
            onDrop={async (e) => {
              const paths = e.paths;
              showImportSelector = false;
              if (paths.length > 0) {
                await handleImportBruno(paths[0]);
              } else {
                await handleImportBruno();
              }
            }}
          >
            {#snippet icon()}
              <svg
                width="44"
                height="44"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="1.4"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
                <polyline points="9 22 9 12 15 12 15 22" />
              </svg>
            {/snippet}
          </DropZone>
        </TabItem>

        <TabItem key="solo" title="solo">
          <DropZone
            title="Drop your solo environment here"
            subtitle="Supports solo environment JSON"
            onDrop={async (e) => {
              const paths = e.paths;
              showImportSelector = false;
              if (paths.length > 0) {
                await executeImport("solo", paths[0], false);
              } else {
                await handleImportSolo();
              }
            }}
          >
            {#snippet icon()}
              <svg
                width="44"
                height="44"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="1.4"
                stroke-linecap="round"
                stroke-linejoin="round"
              >
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                <polyline points="17 8 12 3 7 8" />
                <line x1="12" y1="3" x2="12" y2="15" />
              </svg>
            {/snippet}
          </DropZone>
        </TabItem>

        <TabItem key="git" title="Git">
          <GitEnvImportView
            onImported={() => (showImportSelector = false)}
            onActionStateChange={(state) => {
              gitImportActionState = state;
            }}
          />
        </TabItem>
      </Tabs>
    </div>

    {#snippet footer()}
      <div class="flex w-full justify-end gap-2">
        {#if importActiveTab === "postman"}
          <Button color="primary" onclick={() => handleSelectImportFormat("postman")}
            >Select file…</Button
          >
        {:else if importActiveTab === "bruno"}
          <Button color="primary" onclick={() => handleSelectImportFormat("bruno")}
            >Select file…</Button
          >
        {:else if importActiveTab === "solo"}
          <Button color="primary" onclick={() => handleSelectImportFormat("solo")}
            >Select file…</Button
          >
        {:else if importActiveTab === "git"}
          <Button
            color="primary"
            loading={gitImportActionState?.loading ?? false}
            disabled={gitImportActionState?.disabled ?? true}
            onclick={() => gitImportActionState?.submit()}
          >
            Import from Git
          </Button>
        {/if}
      </div>
    {/snippet}
  </Modal>
{/if}

{#if showOverwriteConfirmDialog}
  <Modal
    title="Overwrite environment?"
    bind:open={showOverwriteConfirmDialog}
    onclose={closeOverwriteConfirmDialog}
    size="xl"
  >
    {#if $topModalId === overwriteEnvironmentModalId}
      <ToastContainer />
    {/if}
    <p>Environment "{overwriteTargetName}" already exists.</p>
    <p class="text-sm text-neutral-600 dark:text-neutral-400">Do you want to overwrite it?</p>
    {#snippet footer()}
      <div class="flex w-full justify-end gap-2">
        <Button color="red" onclick={confirmOverwrite}>Overwrite</Button>
      </div>
    {/snippet}
  </Modal>
{/if}

{#if gitStatusEnvId && gitStatusEnvName}
  <GitStatusPanel
    entityId={gitStatusEnvId}
    entityName={gitStatusEnvName}
    fnGetStatus={GetGitEnvironmentStatus}
    fnSync={SyncGitEnvironment}
    fnKeepOurs={GitEnvKeepOurs}
    fnKeepTheirs={GitEnvKeepTheirs}
    fnAbortRebase={GitEnvAbortRebase}
    fnDiscard={GitEnvDiscardChanges}
    fnOpenTerminal={OpenEnvironmentInTerminal}
    onReload={environmentStore.loadEnvironments}
    onClose={() => {
      gitStatusEnvId = null;
      gitStatusEnvName = null;
    }}
  />
{/if}
