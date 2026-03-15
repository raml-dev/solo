<script lang="ts">
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

<div class="variable-editor">
  {#if env}
    <div class="editor-header">
      <h2>{env.name}</h2>
    </div>
    <div class="variable-grid">
      <div class="grid-header">Key</div>
      <div class="grid-header">Value</div>
      <div class="grid-header"></div>

      {#each variables as variable (variable.id)}
        <input type="text" placeholder="Key" bind:value={variable.key} oninput={debouncedUpdate} />
        <input
          type="text"
          placeholder="Value"
          bind:value={variable.value}
          oninput={debouncedUpdate}
        />
        <button class="remove-btn" onclick={() => removeVariable(variable.id)}>
          <svg
            viewBox="0 0 24 24"
            width="18"
            height="18"
            stroke="currentColor"
            stroke-width="2"
            fill="none"
            stroke-linecap="round"
            stroke-linejoin="round"><path d="M18 6L6 18M6 6l12 12" /></svg
          >
        </button>
      {/each}
    </div>
    <div class="actions">
      <button class="add-btn" onclick={addVariable}>Add Variable</button>
    </div>
  {:else}
    <div class="no-selection">
      <p>Select an environment to view its variables.</p>
    </div>
  {/if}
</div>
