<script lang="ts">
  import { run } from "svelte/legacy";

  import { createEventDispatcher } from "svelte";
  import { environment } from "../../../../wailsjs/go/models";
  import { debounce } from "../../utils/debounce";

  interface Props {
    env?: environment.Environment | null;
  }

  let { env = null }: Props = $props();

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

  run(() => {
    if (!env) {
      variables = [];
      lastLoadedSignature = null;
    } else {
      const sig = `${env.name}::${computeValuesSignature(env.values ?? {})}`;
      if (sig !== lastLoadedSignature) {
        variables = Object.entries(env.values ?? {}).map(([key, value], i) => ({
          id: `${i}-${key}`,
          key,
          value: value?.value ?? "",
          enabled: true
        }));
        lastLoadedSignature = sig;
      }
    }
  });

  const dispatch = createEventDispatcher<{
    update: { name: string; values: Record<string, environment.ValueType> };
  }>();

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
    dispatch("update", { name: env.name, values: updatedValues });
  }, 500);

  function addVariable() {
    variables = [...variables, { id: Date.now().toString(), key: "", value: "", enabled: true }];
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

<style>
  .variable-editor {
    padding: var(--space-lg);
    background: var(--bg-secondary);
    height: 100%;
    overflow-y: auto;
  }
  .editor-header h2 {
    margin: 0 0 var(--space-md) 0;
    font-size: var(--font-size-xl);
  }
  .variable-grid {
    display: grid;
    grid-template-columns: 1fr 1fr 30px;
    gap: var(--space-sm);
    align-items: center;
  }
  .grid-header {
    font-weight: bold;
    font-size: var(--font-size-sm);
    color: var(--text-muted);
  }
  input[type="text"] {
    width: 100%;
    padding: var(--space-sm);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    background: var(--bg-tertiary);
    color: var(--text);
  }
  .remove-btn {
    background: none;
    border: none;
    cursor: pointer;
    color: var(--text-muted);
    padding: var(--space-sm);
  }
  .actions {
    margin-top: var(--space-md);
  }
  .add-btn {
    background: var(--primary);
    color: var(--bg-primary);
    border: none;
    padding: var(--space-sm) var(--space-md);
    border-radius: var(--radius-md);
    cursor: pointer;
  }
  .no-selection {
    display: flex;
    justify-content: center;
    align-items: center;
    height: 100%;
    color: var(--text-muted);
  }
</style>
