<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import VariablesTableEditor from "$src/lib/components/common/VariablesTableEditor.svelte";
  import { environment } from "$wails/go/models";

  interface Props {
    env?: environment.Environment | null;
    onUpdate?: (data: { name: string; values: Record<string, environment.ValueType> }) => void;
  }

  let { env = null, onUpdate }: Props = $props();

  function handleUpdate(values: Record<string, { value: string; type: string }>) {
    if (!env) {
      return;
    }

    const updatedValues = Object.fromEntries(
      Object.entries(values).map(([key, value]) => [key, new environment.ValueType(value)])
    ) as Record<string, environment.ValueType>;
    onUpdate?.({ name: env.name, values: updatedValues });
  }
</script>

<div class="space-y-4">
  {#if env}
    {#key env.id}
      <VariablesTableEditor name={env.name} values={env.values} onUpdate={handleUpdate} />
    {/key}
  {:else}
    <div
      class="rounded-lg border border-dashed border-neutral-300 p-4 text-sm text-neutral-500 dark:border-neutral-700 dark:text-neutral-400"
    >
      Select an environment to view its variables.
    </div>
  {/if}
</div>
