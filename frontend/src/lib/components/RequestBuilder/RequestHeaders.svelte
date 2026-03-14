<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { selectedEnvironment } from "../../../lib/stores/environmentStore";
  import Button from "../base/Button.svelte";
  import TokenInput from "./TokenInput.svelte";

  interface Props {
    headers: Header[];
  }

  let { headers = $bindable() }: Props = $props();

  const dispatch = createEventDispatcher();

  type Header = {
    id: string;
    key: string;
    value: string;
    enabled: boolean;
  };

  function toggleHeader(id: string) {
    headers = headers.map((h) => (h.id === id ? { ...h, enabled: !h.enabled } : h));
    dispatch("change");
  }

  function removeHeader(id: string) {
    headers = headers.filter((h) => h.id !== id);
    dispatch("change");
  }

  function addHeader() {
    const newHeader: Header = {
      id: new Date(Date.now()).toUTCString(),
      key: "",
      value: "",
      enabled: true
    };
    headers = [...headers, newHeader];
    dispatch("change");
  }

  let environmentEntries = $derived(
    Object.entries($selectedEnvironment?.values ?? {}).map(([key, val]) => ({
      key,
      value: String(val?.value ?? "")
    }))
  );
</script>

<div class="headers-editor">
  {#each headers as header (header.id)}
    <div class="header-row">
      <input
        type="checkbox"
        class="header-checkbox"
        checked={header.enabled}
        onchange={() => toggleHeader(header.id)}
      />
      <input
        type="text"
        class="input header-input"
        placeholder="Header name"
        bind:value={header.key}
        disabled={!header.enabled}
        oninput={() => dispatch("change")}
      />
      <div class="header-value-wrapper">
        <TokenInput
          bind:value={header.value}
          placeholder="Value"
          disabled={!header.enabled}
          {environmentEntries}
          wrapperClass="input header-input"
          on:change={() => dispatch("change")}
        />
      </div>
      <Button variant="secondary" click={() => removeHeader(header.id)}>×</Button>
    </div>
  {/each}
  <button class="btn-add-header" onclick={addHeader}> + Add Header </button>
</div>

<style>
  .headers-editor {
    padding: var(--space-md);
  }

  .header-row {
    display: flex;
    gap: var(--space-sm);
    margin-bottom: var(--space-sm);
    align-items: stretch;
  }

  .header-row .input {
    height: 34px;
  }

  .header-checkbox {
    width: 18px;
    height: 18px;
    cursor: pointer;
  }

  .header-input {
    flex: 1;
  }

  .header-value-wrapper {
    flex: 1;
    display: flex;
    min-width: 0;
  }

  .header-input:disabled {
    opacity: 0.5;
    background: var(--bg-tertiary);
  }
  .btn-add-header {
    padding: var(--space-sm) var(--space-md);
    background: none;
    border: 2px dashed var(--border-dark);
    color: var(--text-muted);
    border-radius: var(--radius-md);
    cursor: pointer;
    width: 100%;
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-medium);
    transition: all var(--transition-fast);
  }

  .btn-add-header:hover {
    border-color: var(--primary);
    color: var(--primary);
    background: var(--bg-tertiary);
  }
</style>
