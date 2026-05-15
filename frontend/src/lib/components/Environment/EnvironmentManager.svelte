<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import ToastContainer from "$src/lib/components/base/ToastContainer.svelte";
  import ContextMenu from "$src/lib/components/common/ContextMenu.svelte";
  import ContextMenuItem from "$src/lib/components/common/ContextMenuItem.svelte";
  import FeedbackEmptyState from "$src/lib/components/common/FeedbackEmptyState.svelte";
  import EnvironmentEditor from "$src/lib/components/Environment/EnvironmentEditor.svelte";
  import EnvironmentItem from "$src/lib/components/Environment/EnvironmentItem.svelte";
  import EnvironmentModals from "$src/lib/components/Environment/EnvironmentModals.svelte";
  import GitEnvImportView from "$src/lib/components/GitEnvImportView.svelte";
  import GitStatusPanel from "$src/lib/components/GitStatusPanel.svelte";
  import ImportModal from "$src/lib/components/imports/ImportModal.svelte";
  import type { LocalImportFormatOption } from "$src/lib/components/imports/importTypes";
  import LocalImportPane from "$src/lib/components/imports/LocalImportPane.svelte";
  import { environmentStore, environmentStoreState } from "$src/lib/stores/environmentStore.svelte";
  import { modalStack, topModalId } from "$src/lib/stores/modalStackStore.svelte";
  import { notifications } from "$src/lib/stores/notificationStore";
  import { DEFAULT_ENV_NAME } from "$src/lib/utils/constants";
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
  import AngleDownOutline from "flowbite-svelte-icons/AngleDownOutline.svelte";
  import Button from "flowbite-svelte/Button.svelte";
  import ButtonGroup from "flowbite-svelte/ButtonGroup.svelte";
  import Modal from "flowbite-svelte/Modal.svelte";
  import { onDestroy, onMount, tick } from "svelte";
  import { SvelteSet } from "svelte/reactivity";

  type EnvironmentLocalImportFormat = "postman" | "bruno" | "solo";
  interface EnvironmentContextMenuState {
    environmentName: string | null;
    x: number;
    y: number;
    openKey: number;
    open: boolean;
  }

  const initialEnvironmentContextMenuState: EnvironmentContextMenuState = {
    environmentName: null,
    x: 0,
    y: 0,
    openKey: 0,
    open: false
  };

  const ENVIRONMENT_LOCAL_IMPORT_FORMATS: LocalImportFormatOption<EnvironmentLocalImportFormat>[] =
    [
      {
        key: "solo",
        label: "Solo",
        dropTitle: "Drop your Solo environment here",
        dropSubtitle: "Supports Solo environment JSON",
        pickerButtonLabel: "Select file...",
        icon: "upload"
      },
      {
        key: "postman",
        label: "Postman",
        dropTitle: "Drop your Postman environment here",
        dropSubtitle: "Supports Postman Environment JSON",
        pickerButtonLabel: "Select file...",
        icon: "upload"
      },
      {
        key: "bruno",
        label: "Bruno",
        dropTitle: "Drop your Bruno environment here",
        dropSubtitle: "Supports Bruno environment .bru files",
        pickerButtonLabel: "Select file...",
        icon: "folder"
      }
    ];

  let showNewEnvironmentDialog = $state(false);
  let showDeleteConfirmDialog = $state(false);
  let deleteTarget: string | null = $state(null);
  let localImportFormat = $state<EnvironmentLocalImportFormat>("postman");
  let gitImportActionState: { loading: boolean; disabled: boolean; submit: () => void } | null =
    $state(null);
  let pendingImport: { format: "postman" | "bruno" | "solo"; path: string } | null = null;
  let overwriteTargetName: string | null = $state(null);
  let focusedEnvironmentName: string | null = $state(null);
  let syncingEnvironments: Set<string> = $state(new Set());
  let gitStatusEnvId: string | null = $state(null);
  let gitStatusEnvName: string | null = $state(null);
  let isImportMenuOpen = $state(false);
  let environmentContextMenu = $state({ ...initialEnvironmentContextMenuState });
  let suppressNextPrimaryClick = false;

  const importEnvironmentModal = modalStack.createModal("environment-manager-import");
  const overwriteEnvironmentModal = modalStack.createModal("environment-manager-overwrite");

  onMount(() => {
    document.addEventListener("mousedown", handleGlobalMouseDown, true);
    document.addEventListener("click", handleGlobalClick, true);

    return () => {
      document.removeEventListener("mousedown", handleGlobalMouseDown, true);
      document.removeEventListener("click", handleGlobalClick, true);
    };
  });

  onDestroy(() => {
    modalStack.destroyModal(importEnvironmentModal.id);
    modalStack.destroyModal(overwriteEnvironmentModal.id);
  });

  let environments = $derived(environmentStoreState.environments);
  $effect(() => {
    const focusedEnvironmentExists =
      focusedEnvironmentName && environments.some((e) => e.name === focusedEnvironmentName);
    if (!focusedEnvironmentExists) {
      focusedEnvironmentName =
        environmentStoreState.selectedEnvironmentName || environments[0]?.name || null;
    }

    const contextEnvironmentExists =
      environmentContextMenu.environmentName &&
      environments.some((e) => e.name === environmentContextMenu.environmentName);
    if (environmentContextMenu.environmentName && !contextEnvironmentExists) {
      closeEnvironmentContextMenu();
    }
  });
  let selectedEnvironment = $derived(
    environments.find((env) => env.name === focusedEnvironmentName) || null
  );
  let environmentContextTarget = $derived(
    environmentContextMenu.environmentName
      ? environments.find((env) => env.name === environmentContextMenu.environmentName) || null
      : null
  );

  function openEnvironment(name: string) {
    if (environmentContextMenu.open) {
      closeEnvironmentContextMenu();
      return;
    }

    focusedEnvironmentName = name;
  }

  function activateEnvironment(name: string) {
    if (environmentContextMenu.open) {
      closeEnvironmentContextMenu();
      return;
    }

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
    closeEnvironmentContextMenu();
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

  function isContextMenuOpen(): boolean {
    return environmentContextMenu.open;
  }

  function isClickInsideContextMenu(target: EventTarget | null): boolean {
    return target instanceof Element && target.closest('[popover="manual"]') !== null;
  }

  function closeEnvironmentContextMenu() {
    environmentContextMenu.open = false;
  }

  function getContextMenuPosition(event: MouseEvent) {
    if (event.type === "contextmenu") {
      return { x: event.clientX, y: event.clientY };
    }

    if (event.currentTarget instanceof HTMLElement) {
      const bounds = event.currentTarget.getBoundingClientRect();
      return { x: bounds.left, y: bounds.bottom };
    }

    return { x: event.clientX, y: event.clientY };
  }

  async function openEnvironmentContextMenu(environmentName: string, event: MouseEvent) {
    event.preventDefault();
    event.stopPropagation();

    const { x, y } = getContextMenuPosition(event);
    environmentContextMenu = {
      environmentName,
      x,
      y,
      openKey: environmentContextMenu.openKey + 1,
      open: false
    };
    await tick();
    environmentContextMenu.open = true;
  }

  function handleGlobalMouseDown(event: MouseEvent) {
    if (event.button !== 0 || !isContextMenuOpen() || isClickInsideContextMenu(event.target)) {
      return;
    }

    suppressNextPrimaryClick = true;
    closeEnvironmentContextMenu();
  }

  function handleGlobalClick(event: MouseEvent) {
    if (!suppressNextPrimaryClick || event.button !== 0) {
      return;
    }

    suppressNextPrimaryClick = false;
    event.preventDefault();
    event.stopPropagation();
  }

  function getEnvironmentContextMenuTriggerId(): string {
    return `environment-context-menu-trigger-${environmentContextMenu.openKey}`;
  }

  function getEnvironmentContextMenuPositionStyle(): string {
    return `left: ${environmentContextMenu.x + 2}px; top: ${environmentContextMenu.y + 2}px;`;
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

  async function handleDuplicateEnvironment(name: string) {
    try {
      const duplicateName = await environmentStore.duplicateEnvironment(name);
      focusedEnvironmentName = duplicateName;
      notifications.success("Environment duplicated successfully");
    } catch {
      // error already shown by store
    } finally {
      closeEnvironmentContextMenu();
    }
  }

  function handleRenameEnvironment(name: string) {
    focusedEnvironmentName = name;
    environmentStore.startRenameEnvironment(name);
    closeEnvironmentContextMenu();
  }

  function openImportModal() {
    localImportFormat = "postman";
    gitImportActionState = null;
    importEnvironmentModal.open = true;
  }

  function closeImportMenu() {
    isImportMenuOpen = false;
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
        overwriteEnvironmentModal.open = true;
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
    const filePath = path ?? (await SelectFile("Select Solo Environment", "*.json", "JSON Files"));
    if (!filePath) return;
    await executeImport("solo", filePath, false);
  }

  async function handleLocalEnvironmentImport(format: EnvironmentLocalImportFormat, path?: string) {
    importEnvironmentModal.open = false;

    if (format === "postman") {
      if (path) {
        await executeImport("postman", path, false);
      } else {
        await handleImportPostman();
      }
      return;
    }

    if (format === "bruno") {
      if (path) {
        await executeImport("bruno", path, false);
      } else {
        await handleImportBruno();
      }
      return;
    }

    if (path) {
      await executeImport("solo", path, false);
    } else {
      await handleImportSolo();
    }
  }

  async function confirmOverwrite() {
    if (!pendingImport) return;
    const { format, path } = pendingImport;
    pendingImport = null;
    overwriteEnvironmentModal.open = false;
    await executeImport(format, path, true);
  }

  function closeOverwriteConfirmDialog() {
    pendingImport = null;
    overwriteTargetName = null;
    overwriteEnvironmentModal.open = false;
  }

  let selectedLocalImportOption = $derived(
    ENVIRONMENT_LOCAL_IMPORT_FORMATS.find((format) => format.key === localImportFormat)
  );
</script>

{#snippet importDropdown(triggeredBy: string, isOpen: boolean | undefined, onClose: () => void)}
  <ContextMenu {triggeredBy} {isOpen} menuClass="z-50 w-50" {onClose}>
    <ContextMenuItem
      onclick={() => {
        openImportModal();
        onClose();
      }}
    >
      Import environment...
    </ContextMenuItem>
  </ContextMenu>
{/snippet}

<div class="flex h-full">
  <div class="flex w-56 shrink-0 flex-col border-r border-neutral-200 dark:border-neutral-700">
    <div
      class="flex shrink-0 items-center justify-between border-b border-neutral-200 px-4 py-3 md:px-5 dark:border-neutral-700"
    >
      <h3 class="text-sm font-semibold text-neutral-800 dark:text-neutral-100">Environments</h3>
      <div class="flex items-center gap-1">
        <ButtonGroup>
          <Button
            color="primary"
            class="shrink-0 cursor-pointer border-none inset-ring-primary-500 focus-within:inset-ring-1 focus-within:outline-hidden focus:ring-0 focus:outline-hidden dark:border-none"
            size="xs"
            onclick={() => (showNewEnvironmentDialog = true)}
            >New
          </Button>
          <Button
            color="primary"
            class="w-0.5 shrink-0 cursor-pointer border-l px-2.5 inset-ring-primary-500 focus-within:inset-ring-1 focus-within:outline-hidden focus:ring-0 focus:outline-hidden dark:border-l-primary-900"
            size="xs"
            id="import-env-dropdown-button"
            onclick={() => (isImportMenuOpen = true)}
            ><AngleDownOutline class="w-4 shrink-0" /></Button
          >
        </ButtonGroup>
      </div>
      {@render importDropdown("#import-env-dropdown-button", isImportMenuOpen, closeImportMenu)}
    </div>

    {#if environmentStoreState.loading}
      <div class="px-3 py-2 text-sm text-neutral-500 dark:text-neutral-400">
        Loading environments...
      </div>
    {:else if environments.length > 0}
      <div class="flex-1 overflow-y-auto p-2">
        <div class="space-y-2">
          {#each environments as environment (environment.id)}
            <EnvironmentItem
              env={environment}
              isActive={environment.name === environmentStoreState.selectedEnvironmentName}
              isFocused={environment.name === focusedEnvironmentName}
              isMenuOpen={environmentContextMenu.open &&
                environment.name === environmentContextMenu.environmentName}
              onOpen={openEnvironment}
              onActivate={activateEnvironment}
              onOpenMenu={openEnvironmentContextMenu}
            />
          {/each}
        </div>
      </div>
    {:else}
      <div class="p-2">
        <FeedbackEmptyState
          title="No environments yet"
          detail="Create your first environment to get started"
        />
      </div>
    {/if}
  </div>
  <div class="min-w-0 flex-1 overflow-y-auto p-4">
    <EnvironmentEditor env={selectedEnvironment} onUpdate={handleUpdateEnvironment} />
  </div>
</div>

{#if environmentContextTarget}
  <div data-environment-context-menu="true">
    <button
      id={getEnvironmentContextMenuTriggerId()}
      type="button"
      class="pointer-events-none fixed z-90 h-0 w-0 opacity-0"
      style={getEnvironmentContextMenuPositionStyle()}
      tabindex="-1"
      aria-hidden="true"
    ></button>
    <ContextMenu
      triggeredBy={`#${getEnvironmentContextMenuTriggerId()}`}
      isOpen={environmentContextMenu.open}
      onClose={closeEnvironmentContextMenu}
    >
      {#if environmentContextTarget.gitRemote}
        <ContextMenuItem
          onclick={() => {
            handleGitStatus(environmentContextTarget.id);
            closeEnvironmentContextMenu();
          }}
        >
          Git status
        </ContextMenuItem>
        <ContextMenuItem
          disabled={syncingEnvironments.has(environmentContextTarget.id)}
          onclick={() => {
            void handleSync(environmentContextTarget.id);
            closeEnvironmentContextMenu();
          }}
        >
          {syncingEnvironments.has(environmentContextTarget.id) ? "Syncing..." : "Sync with Git"}
        </ContextMenuItem>
      {/if}
      <ContextMenuItem
        onclick={() => {
          handleRenameEnvironment(environmentContextTarget.name);
        }}
      >
        Rename
      </ContextMenuItem>
      <ContextMenuItem
        onclick={() => {
          void handleDuplicateEnvironment(environmentContextTarget.name);
        }}
      >
        Duplicate
      </ContextMenuItem>
      <ContextMenuItem
        onclick={() => {
          void handleExportEnvironment(environmentContextTarget.name);
          closeEnvironmentContextMenu();
        }}
      >
        Export
      </ContextMenuItem>
      {#if environmentContextTarget.name !== DEFAULT_ENV_NAME}
        <ContextMenuItem
          className="text-danger-600 hover:bg-danger-50 dark:text-danger-400 dark:hover:bg-danger-900/20"
          onclick={() => {
            handleDeleteEnvironment(environmentContextTarget.name);
          }}
        >
          Delete
        </ContextMenuItem>
      {/if}
    </ContextMenu>
  </div>
{/if}

<EnvironmentModals
  bind:showNewEnvironmentDialog
  bind:showDeleteConfirmDialog
  bind:deleteTarget
  onCreate={handleCreateEnvironment}
  onConfirmDelete={confirmDelete}
  onCloseNew={() => (showNewEnvironmentDialog = false)}
  onCloseDelete={() => (showDeleteConfirmDialog = false)}
/>

{#if importEnvironmentModal.open}
  <ImportModal
    title="Import Environment"
    bind:open={importEnvironmentModal.open}
    localActionLabel={selectedLocalImportOption?.pickerButtonLabel || "Select file..."}
    onLocalAction={() => handleLocalEnvironmentImport(localImportFormat)}
    gitActionState={gitImportActionState}
  >
    {#snippet localContent()}
      {#if $topModalId === importEnvironmentModal.id}
        <ToastContainer />
      {/if}
      <LocalImportPane
        formats={ENVIRONMENT_LOCAL_IMPORT_FORMATS}
        bind:selectedFormat={localImportFormat}
        onImport={handleLocalEnvironmentImport}
      />
    {/snippet}

    {#snippet gitContent()}
      {#if $topModalId === importEnvironmentModal.id}
        <ToastContainer />
      {/if}
      <GitEnvImportView
        onImported={() => (importEnvironmentModal.open = false)}
        onActionStateChange={(state) => {
          gitImportActionState = state;
        }}
      />
    {/snippet}
  </ImportModal>
{/if}

{#if overwriteEnvironmentModal.open}
  <Modal
    title="Overwrite environment?"
    bind:open={overwriteEnvironmentModal.open}
    onclose={closeOverwriteConfirmDialog}
    size="xl"
  >
    {#if $topModalId === overwriteEnvironmentModal.id}
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
