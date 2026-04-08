<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: GPL-3.0-only
-->

<script lang="ts">
  import ToastContainer from "$src/lib/components/base/ToastContainer.svelte";
  import FeedbackEmptyState from "$src/lib/components/common/FeedbackEmptyState.svelte";
  import EnvironmentEditor from "$src/lib/components/Environment/EnvironmentEditor.svelte";
  import EnvironmentItem from "$src/lib/components/Environment/EnvironmentItem.svelte";
  import EnvironmentModals from "$src/lib/components/Environment/EnvironmentModals.svelte";
  import GitEnvImportView from "$src/lib/components/GitEnvImportView.svelte";
  import GitStatusPanel from "$src/lib/components/GitStatusPanel.svelte";
  import ImportModal from "$src/lib/components/imports/ImportModal.svelte";
  import LocalImportPane from "$src/lib/components/imports/LocalImportPane.svelte";
  import { environmentStore, environmentStoreState } from "$src/lib/stores/environmentStore.svelte";
  import { modalStack, topModalId } from "$src/lib/stores/modalStackStore.svelte";
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
  import AngleDownOutline from "flowbite-svelte-icons/AngleDownOutline.svelte";
  import Button from "flowbite-svelte/Button.svelte";
  import ButtonGroup from "flowbite-svelte/ButtonGroup.svelte";
  import Dropdown from "flowbite-svelte/Dropdown.svelte";
  import DropdownItem from "flowbite-svelte/DropdownItem.svelte";
  import Modal from "flowbite-svelte/Modal.svelte";
  import type { LocalImportFormatOption } from "$src/lib/components/imports/importTypes";
  import { onDestroy } from "svelte";
  import { SvelteSet } from "svelte/reactivity";

  type EnvironmentLocalImportFormat = "postman" | "bruno" | "solo";

  const ENVIRONMENT_LOCAL_IMPORT_FORMATS: LocalImportFormatOption<EnvironmentLocalImportFormat>[] =
    [
      {
        key: "postman",
        label: "Postman",
        dropTitle: "Drop your Postman environment here",
        dropSubtitle: "Supports Postman Environment JSON",
        pickerButtonLabel: "Select file…",
        icon: "upload"
      },
      {
        key: "bruno",
        label: "Bruno",
        dropTitle: "Drop your Bruno environment here",
        dropSubtitle: "Supports Bruno environment .bru files",
        pickerButtonLabel: "Select file…",
        icon: "folder"
      },
      {
        key: "solo",
        label: "solo",
        dropTitle: "Drop your solo environment here",
        dropSubtitle: "Supports solo environment JSON",
        pickerButtonLabel: "Select file…",
        icon: "upload"
      }
    ];

  let showNewEnvironmentDialog = $state(false);
  let showDeleteConfirmDialog = $state(false);
  let deleteTarget: string | null = $state(null);
  let activeMenu: string | null = $state(null);
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

  const importEnvironmentModal = modalStack.createModal("environment-manager-import");
  const overwriteEnvironmentModal = modalStack.createModal("environment-manager-overwrite");

  onDestroy(() => {
    modalStack.destroyModal(importEnvironmentModal.id);
    modalStack.destroyModal(overwriteEnvironmentModal.id);
  });

  let environments = $derived(environmentStoreState.environments);
  $effect(() => {
    const exists =
      focusedEnvironmentName && environments.some((e) => e.name === focusedEnvironmentName);
    if (!exists) {
      focusedEnvironmentName =
        environmentStoreState.selectedEnvironmentName || environments[0]?.name || null;
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
    const filePath = path ?? (await SelectFile("Select solo Environment", "*.json", "JSON Files"));
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
  <Dropdown {triggeredBy} {isOpen} class="z-50 w-50" triggerDelay={0} onclose={onClose}>
    <DropdownItem
      class="text-gray-900 dark:text-white"
      onclick={() => {
        openImportModal();
        onClose();
      }}
    >
      Import environment...
    </DropdownItem>
  </Dropdown>
{/snippet}

<svelte:window onclick={closeActiveMenu} />

<div class="flex h-full">
  <div class="flex w-56 shrink-0 flex-col border-r border-neutral-200 dark:border-neutral-700">
    <div
      class="flex shrink-0 items-center justify-end gap-2 border-b border-neutral-200 pr-4 pb-3 md:pr-5 dark:border-neutral-700"
    >
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
      {@render importDropdown("#import-env-dropdown-button", isImportMenuOpen, closeImportMenu)}
    </div>

    {#if environmentStoreState.loading}
      <div class="px-3 py-2 text-sm text-neutral-500 dark:text-neutral-400">
        Loading environments...
      </div>
    {:else if environments.length > 0}
      <div class="flex-1 overflow-y-auto pt-3 pr-4 md:pr-5">
        {#each environments as environment (environment.id)}
          <EnvironmentItem
            env={environment}
            menuOpen={activeMenu === environment.name}
            isActive={environment.name === environmentStoreState.selectedEnvironmentName}
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
    {:else}
      <div class="pt-3 pr-4 md:pr-5">
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
    localActionLabel={selectedLocalImportOption?.pickerButtonLabel || "Select file…"}
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
