<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import { collectionStore, collectionStoreState } from "$src/lib/stores/collectionStore.svelte";
  import { environmentStore, environmentStoreState } from "$src/lib/stores/environmentStore.svelte";
  import { sessionVarsStore } from "$src/lib/stores/sessionVarsStore";
  import { getActiveTab } from "$src/lib/stores/tabStore.svelte";
  import {
    cancelHideTokenTooltip,
    forceHideTokenTooltip,
    tokenTooltipStore,
    tooltipMouseLeave
  } from "$src/lib/stores/tokenTooltipStore";
  import {
    formatVariableSourceLabel,
    resolveVariableEntries,
    type ResolvedVariableEntry,
    type VariableSource
  } from "$src/lib/utils/variableResolution";
  import { environment } from "$wails/go/models";
  import EditOutline from "flowbite-svelte-icons/EditOutline.svelte";
  import ExclamationCircleSolid from "flowbite-svelte-icons/ExclamationCircleSolid.svelte";
  import Badge from "flowbite-svelte/Badge.svelte";
  import Button from "flowbite-svelte/Button.svelte";
  import Input from "flowbite-svelte/Input.svelte";
  import { tick } from "svelte";

  type EditableVariableSource = "environment" | "collection";

  let editingSource: EditableVariableSource | null = $state(null);
  let editValue = $state("");
  let inputElement: HTMLInputElement | undefined = $state();

  let tooltipState = $derived($tokenTooltipStore);
  let tokenKey = $derived(tooltipState.tokenKey);
  let visible = $derived(tooltipState.visible);
  let selectedEnvironment = $derived(
    environmentStoreState.environments.find(
      (currentEnvironment) =>
        currentEnvironment.name === environmentStoreState.selectedEnvironmentName
    ) || null
  );
  let activeCollection = $derived(
    collectionStoreState.collections.find(
      (currentCollection) => currentCollection.name === getActiveTab()?.collectionName
    ) || null
  );
  let resolvedEntry = $derived.by<ResolvedVariableEntry | null>(() => {
    if (!tokenKey) {
      return null;
    }

    return (
      resolveVariableEntries({
        sessionValues: $sessionVarsStore,
        environmentValues: selectedEnvironment?.values,
        collectionValues: activeCollection?.variables
      }).find((entry) => entry.key === tokenKey) || null
    );
  });
  let sessionValue = $derived(tokenKey ? $sessionVarsStore[tokenKey] : undefined);
  let environmentValue = $derived(
    tokenKey && selectedEnvironment?.values[tokenKey]
      ? String(selectedEnvironment.values[tokenKey]?.value ?? "")
      : ""
  );
  let collectionValue = $derived(
    tokenKey && activeCollection?.variables?.[tokenKey]
      ? String(activeCollection.variables[tokenKey]?.value ?? "")
      : ""
  );
  let hasSessionValue = $derived(sessionValue !== undefined);
  let hasEnvironmentValue = $derived(
    !!tokenKey && selectedEnvironment?.values[tokenKey] !== undefined
  );
  let hasCollectionValue = $derived(
    !!tokenKey && activeCollection?.variables?.[tokenKey] !== undefined
  );
  let isEditing = $derived(editingSource !== null);

  function getBadgeColor(source: VariableSource): "gray" | "blue" | "yellow" {
    return source === "session" ? "gray" : source === "environment" ? "blue" : "yellow";
  }

  function getConflictMessage(entry: ResolvedVariableEntry): string {
    if (!entry.hasConflicts) {
      return `Using ${formatVariableSourceLabel(entry.winningSource)}`;
    }

    const sources = entry.definedIn.map((source) => formatVariableSourceLabel(source)).join(", ");
    return `Defined in ${sources}.`;
  }

  function getEditButtonLabel(source: EditableVariableSource): string {
    if (source === "environment") {
      return hasEnvironmentValue ? "Edit Environment" : "Add Environment";
    }

    return hasCollectionValue ? "Edit Collection" : "Add Collection";
  }

  function getEditSourceValue(source: EditableVariableSource): string {
    return source === "environment" ? environmentValue : collectionValue;
  }

  function getEditingOverrideMessage(
    entry: ResolvedVariableEntry | null,
    source: EditableVariableSource | null
  ): string | null {
    if (!entry || !source) {
      return null;
    }

    if (source === "environment" && entry.winningSource === "session") {
      return "Session currently overrides this variable. Saving here updates the environment value, but the computed value stays on Session until the session variable is cleared.";
    }

    if (source === "collection") {
      if (entry.winningSource === "session") {
        return "Session currently overrides this variable. Saving here updates the collection value, but the computed value stays on Session until the session variable is cleared.";
      }

      if (entry.winningSource === "environment") {
        return "Environment currently overrides this variable. Saving here updates the collection value, but the computed value stays on Environment until the environment value is removed or changed.";
      }
    }

    return null;
  }

  async function beginEdit(source: EditableVariableSource) {
    editValue = getEditSourceValue(source);
    editingSource = source;
    await tick();
    inputElement?.focus();
  }

  function handleTooltipMouseLeave() {
    if (!isEditing) {
      tooltipMouseLeave();
    }
  }

  async function save() {
    if (!tokenKey || !editingSource) {
      return;
    }

    if (editingSource === "environment") {
      if (!selectedEnvironment) {
        return;
      }

      const processedValues = { ...selectedEnvironment.values };
      processedValues[tokenKey] = new environment.ValueType({
        value: editValue,
        type: "string"
      });

      const environmentToSave = new environment.Environment({
        ...selectedEnvironment,
        values: processedValues
      });

      try {
        await environmentStore.updateEnvironment(environmentToSave);
        editingSource = null;
      } catch {
        // error already shown by store
      }
      return;
    }

    if (!activeCollection) {
      return;
    }

    const processedValues = {
      ...(activeCollection.variables ?? {}),
      [tokenKey]: {
        value: editValue,
        type: "string"
      }
    };

    try {
      await collectionStore.updateCollectionVariables(activeCollection.name, processedValues);
      editingSource = null;
    } catch {
      // error already shown by store
    }
  }

  async function clearSessionOverride() {
    if (!tokenKey) {
      return;
    }

    try {
      await sessionVarsStore.remove(tokenKey);
    } catch {
      // session store fails silently
    }
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === "Enter") {
      void save();
      return;
    }

    if (event.key === "Escape") {
      editingSource = null;
      forceHideTokenTooltip();
    }
  }

  $effect(() => {
    if (!visible) {
      editingSource = null;
    }
  });
</script>

{#if visible}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="fixed z-50 max-w-120 rounded-lg border border-neutral-200 bg-white p-3 shadow-lg dark:border-neutral-700 dark:bg-neutral-800"
    style="left: {tooltipState.x}px; top: {tooltipState.y}px;"
    onmouseenter={cancelHideTokenTooltip}
    onmouseleave={handleTooltipMouseLeave}
  >
    {#if isEditing}
      <div class="flex flex-col gap-3">
        <div class="flex items-center gap-2">
          <span class="font-mono text-xs text-neutral-500 dark:text-neutral-400">{tokenKey}</span>
          {#if editingSource}
            <Badge color={getBadgeColor(editingSource)}>
              {formatVariableSourceLabel(editingSource)}
            </Badge>
          {/if}
        </div>

        <Input
          bind:elementRef={inputElement}
          type="text"
          size="sm"
          bind:value={editValue}
          onkeydown={handleKeydown}
        />

        {#if getEditingOverrideMessage(resolvedEntry, editingSource)}
          <p class="text-xs text-neutral-500 dark:text-neutral-400">
            {getEditingOverrideMessage(resolvedEntry, editingSource)}
          </p>
        {/if}

        <div class="flex items-center gap-2">
          <Button color="primary" size="xs" onclick={save}>Save</Button>
          <Button color="light" size="xs" onclick={() => (editingSource = null)}>Cancel</Button>
        </div>
      </div>
    {:else}
      <div class="flex flex-col gap-3">
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <div class="font-mono text-sm text-neutral-800 dark:text-neutral-100">{tokenKey}</div>
            {#if resolvedEntry}
              <div class="mt-1 text-sm text-neutral-600 dark:text-neutral-300">
                {resolvedEntry.computedValue}
              </div>
            {:else}
              <div class="mt-1 text-sm text-neutral-400 italic dark:text-neutral-500">
                Unresolved variable
              </div>
            {/if}
          </div>

          {#if resolvedEntry}
            <Badge color={getBadgeColor(resolvedEntry.winningSource)}>
              {formatVariableSourceLabel(resolvedEntry.winningSource)}
            </Badge>
          {/if}
        </div>

        {#if resolvedEntry}
          <p class="flex flex-col gap-1 text-xs text-neutral-500 dark:text-neutral-400">
            <span class="flex flex-row items-center gap-1">
              <ExclamationCircleSolid class="h-4 w-4 shrink-0 fill-warning-500" /><span
                >{getConflictMessage(resolvedEntry)}</span
              ></span
            >
            <span
              >Using {formatVariableSourceLabel(resolvedEntry.winningSource)} by priority: Session > Environment
              > Collection.</span
            >
          </p>
        {:else}
          <p class="text-xs text-neutral-500 dark:text-neutral-400">
            This variable is not defined in Session, Environment, or Collection.
          </p>
        {/if}

        <div class="flex flex-wrap items-center gap-2">
          {#if selectedEnvironment}
            <Button
              color="light"
              size="xs"
              onclick={() => void beginEdit("environment")}
              title={getEditButtonLabel("environment")}
            >
              <EditOutline class="mr-1 h-4 w-4" />
              <span>{getEditButtonLabel("environment")}</span>
            </Button>
          {/if}

          {#if activeCollection}
            <Button
              color="light"
              size="xs"
              onclick={() => void beginEdit("collection")}
              title={getEditButtonLabel("collection")}
            >
              <EditOutline class="mr-1 h-4 w-4" />
              <span>{getEditButtonLabel("collection")}</span>
            </Button>
          {/if}

          {#if hasSessionValue}
            <Button color="light" size="xs" onclick={clearSessionOverride}>Clear Session</Button>
          {/if}
        </div>
      </div>
    {/if}
  </div>
{/if}
