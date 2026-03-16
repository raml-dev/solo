<script lang="ts">
  import Button from "flowbite-svelte/Button.svelte";
  import Input from "flowbite-svelte/Input.svelte";
  import { debounce } from "$src/lib/utils/debounce";
  import { createStableId, mapRecordToRowsWithStableIds } from "$src/lib/utils/stableKeyValueRows";
  import { environment } from "$wails/go/models";

  interface Props {
    env?: environment.Environment | null;
    onUpdate?: (data: { name: string; values: Record<string, environment.ValueType> }) => void;
  }

  let { env = null, onUpdate }: Props = $props();

  type EnvVar = {
    id: string;
    key: string;
    value: string;
    enabled: boolean;
  };

  let variables: EnvVar[] = $state([]);
  let lastLoadedSignature: string | null = $state(null);

  function computeValuesSignature(values: Record<string, environment.ValueType> = {}) {
    const entries = Object.entries(values)
      .map(([k, v]) => [k, String(v?.value ?? "")])
      .sort((a, b) => a[0].localeCompare(b[0]));
    return JSON.stringify(entries);
  }

  $effect(() => {
    if (!env) {
      variables = [];
      lastLoadedSignature = null;
    } else {
      const valuesRecord = Object.fromEntries(
        Object.entries(env.values ?? {}).map(([key, val]) => [key, String(val?.value ?? "")])
      );
      const sig = `${env.name}::${computeValuesSignature(env.values ?? {})}`;
      if (sig !== lastLoadedSignature) {
        variables = mapRecordToRowsWithStableIds(valuesRecord, variables).map((row) => ({
          ...row,
          enabled: true
        }));
        lastLoadedSignature = sig;
      }
    }
  });

  const debouncedUpdate = debounce(() => {
    if (!env) return;
    const updatedValues = variables.reduce(
      (acc, v) => {
        if (v.key.trim()) {
          acc[v.key.trim()] = new environment.ValueType({
            value: v.value,
            type: "default"
          });
        }
        return acc;
      },
      {} as Record<string, environment.ValueType>
    );
    onUpdate?.({ name: env.name, values: updatedValues });
  }, 500);

  function addVariable() {
    variables = [...variables, { id: createStableId(), key: "", value: "", enabled: true }];
  }

  function removeVariable(id: string) {
    variables = variables.filter((v) => v.id !== id);
    debouncedUpdate();
  }
</script>

<div class="space-y-4">
  {#if env}
    <div>
      <h2 class="text-base font-semibold text-gray-900 dark:text-white">{env.name}</h2>
    </div>

    <div class="space-y-2">
      <div
        class="grid grid-cols-[1fr_1fr_auto] gap-2 px-1 text-xs font-semibold text-gray-500 uppercase"
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
            aria-label={`Remove variable ${variable.key || "row"}`}
            onclick={() => removeVariable(variable.id)}
          >
            ✕
          </Button>
        </div>
      {/each}
    </div>

    <div class="flex justify-end">
      <Button color="light" size="sm" onclick={addVariable}>Add Variable</Button>
    </div>
  {:else}
    <div
      class="rounded-lg border border-dashed border-gray-300 p-4 text-sm text-gray-500 dark:border-gray-700 dark:text-gray-400"
    >
      Select an environment to view its variables.
    </div>
  {/if}
</div>
