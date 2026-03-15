<script lang="ts">
  import Button from "flowbite-svelte/Button.svelte";
  import TokenInput from "$src/lib/components/RequestBuilder/TokenInput.svelte";
  import { selectedEnvironment } from "$src/lib/stores/environmentStore";

  interface Props {
    headers: Header[];
    onChange?: () => void;
  }

  let { headers = $bindable(), onChange }: Props = $props();

  type Header = {
    id: string;
    key: string;
    value: string;
    enabled: boolean;
  };

  function toggleHeader(id: string) {
    headers = headers.map((h) => (h.id === id ? { ...h, enabled: !h.enabled } : h));
    onChange?.();
  }

  function removeHeader(id: string) {
    headers = headers.filter((h) => h.id !== id);
    onChange?.();
  }

  function addHeader() {
    const newHeader: Header = {
      id: new Date(Date.now()).toUTCString(),
      key: "",
      value: "",
      enabled: true
    };
    headers = [...headers, newHeader];
    onChange?.();
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
        oninput={() => onChange?.()}
      />
      <div class="header-value-wrapper">
        <TokenInput
          bind:value={header.value}
          placeholder="Value"
          disabled={!header.enabled}
          {environmentEntries}
          wrapperClass="input header-input"
          onChange={() => onChange?.()}
        />
      </div>
      <Button color="light" onclick={() => removeHeader(header.id)} aria-label="Remove header"
        >×</Button
      >
    </div>
  {/each}
  <button class="btn-add-header" onclick={addHeader}> + Add Header </button>
</div>
