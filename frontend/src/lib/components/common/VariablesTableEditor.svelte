<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import Button from "flowbite-svelte/Button.svelte";
  import Input from "flowbite-svelte/Input.svelte";
  import PlusOutline from "flowbite-svelte-icons/PlusOutline.svelte";
  import TrashBinOutline from "flowbite-svelte-icons/TrashBinOutline.svelte";
  import { debounce } from "$src/lib/utils/debounce";
  import { createStableId, mapRecordToRowsWithStableIds } from "$src/lib/utils/stableKeyValueRows";
  import type {
    NormalizedVariableValue,
    VariableValueLike
  } from "$src/lib/utils/variableResolution";

  interface Props {
    name: string;
    values?: Record<string, VariableValueLike> | null;
    onUpdate?: (values: Record<string, NormalizedVariableValue>) => void;
  }

  type VariableRow = {
    id: string;
    key: string;
    value: string;
    type: string;
    enabled: boolean;
  };

  let { name, values = {}, onUpdate }: Props = $props();

  let variables: VariableRow[] = $state([]);
  let lastLoadedSignature: string | null = $state(null);

  function computeValuesSignature(currentValues: Record<string, VariableValueLike> = {}) {
    const entries = Object.entries(currentValues)
      .map(([key, value]) => [key, String(value?.value ?? ""), String(value?.type ?? "default")])
      .sort((left, right) => left[0].localeCompare(right[0]));

    return JSON.stringify(entries);
  }

  $effect(() => {
    const signature = `${name}::${computeValuesSignature(values ?? {})}`;
    if (signature === lastLoadedSignature) {
      return;
    }

    const normalizedValues = Object.fromEntries(
      Object.entries(values ?? {}).map(([key, value]) => [key, String(value?.value ?? "")])
    );
    const valueTypes = Object.fromEntries(
      Object.entries(values ?? {}).map(([key, value]) => [key, String(value?.type ?? "default")])
    );

    variables = mapRecordToRowsWithStableIds(normalizedValues, variables).map((row) => ({
      ...row,
      type: valueTypes[row.key] ?? "default",
      enabled: true
    }));
    lastLoadedSignature = signature;
  });

  const debouncedUpdate = debounce(() => {
    const updatedValues = variables.reduce(
      (acc, variable) => {
        const trimmedKey = variable.key.trim();
        if (!trimmedKey) {
          return acc;
        }

        acc[trimmedKey] = {
          value: variable.value,
          type: variable.type || "default"
        };
        return acc;
      },
      {} as Record<string, NormalizedVariableValue>
    );

    onUpdate?.(updatedValues);
  }, 500);

  function addVariable() {
    variables = [
      ...variables,
      { id: createStableId(), key: "", value: "", type: "default", enabled: true }
    ];
  }

  function removeVariable(id: string) {
    variables = variables.filter((variable) => variable.id !== id);
    debouncedUpdate();
  }
</script>

<div class="space-y-4">
  <div>
    <h2 class="text-base font-semibold text-neutral-900 dark:text-white">{name}</h2>
  </div>

  <div class="space-y-2">
    <div
      class="grid grid-cols-[1fr_1fr_auto] gap-2 px-1 text-xs font-semibold text-neutral-500 uppercase"
    >
      <span>Key</span>
      <span>Value</span>
      <span class="sr-only">Actions</span>
    </div>

    {#each variables as variable (variable.id)}
      <div class="grid grid-cols-[1fr_1fr_auto] items-center gap-2">
        <Input
          type="text"
          size="sm"
          placeholder="Key"
          bind:value={variable.key}
          oninput={debouncedUpdate}
        />
        <Input
          type="text"
          size="sm"
          placeholder="Value"
          bind:value={variable.value}
          oninput={debouncedUpdate}
        />
        <Button
          color="light"
          size="xs"
          class="shrink-0"
          aria-label={`Remove variable ${variable.key || "row"}`}
          onclick={() => removeVariable(variable.id)}
        >
          <TrashBinOutline class="h-4 w-4" />
        </Button>
      </div>
    {/each}
  </div>

  <div class="flex justify-end">
    <Button color="light" size="sm" onclick={addVariable}>
      <PlusOutline class="mr-1 h-4 w-4" />
      <span>Add Variable</span>
    </Button>
  </div>
</div>
