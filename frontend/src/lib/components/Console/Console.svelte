<script lang="ts">
  import { historyStore } from "../../stores/historyStore";
  import ConsoleEntry from "./ConsoleEntry.svelte";
  import Button from "../base/Button.svelte";

  let filterMethod = "";
  let filterStatus = "";
  let filterUrl = "";

  const METHOD_OPTIONS = [
    { value: "", label: "All methods" },
    { value: "GET",    label: "GET"    },
    { value: "POST",   label: "POST"   },
    { value: "PUT",    label: "PUT"    },
    { value: "PATCH",  label: "PATCH"  },
    { value: "DELETE", label: "DELETE" },
  ];

  const STATUS_OPTIONS = [
    { value: "",    label: "All statuses"      },
    { value: "2xx", label: "2xx Success"       },
    { value: "3xx", label: "3xx Redirect"      },
    { value: "4xx", label: "4xx Client error"  },
    { value: "5xx", label: "5xx Server error"  },
    { value: "err", label: "Error"             },
  ];

  $: filtered = $historyStore.filter((e) => {
    if (filterMethod && e.request.method !== filterMethod) return false;
    if (filterUrl && !e.request.url.toLowerCase().includes(filterUrl.toLowerCase())) return false;
    if (filterStatus) {
      if (filterStatus === "err") return !!e.error;
      if (!e.response) return false;
      const s = e.response.status;
      if (filterStatus === "2xx" && !(s >= 200 && s < 300)) return false;
      if (filterStatus === "3xx" && !(s >= 300 && s < 400)) return false;
      if (filterStatus === "4xx" && !(s >= 400 && s < 500)) return false;
      if (filterStatus === "5xx" && !(s >= 500))            return false;
    }
    return true;
  });

  function handleExport() {
    const har = historyStore.exportHAR($historyStore);
    const blob = new Blob([har], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "yapla-export.har";
    a.click();
    URL.revokeObjectURL(url);
  }
</script>

<div class="console">
  <!-- Toolbar -->
  <div class="console-toolbar">
    <span class="console-title">Console</span>

    <input
      class="filter-input"
      type="text"
      placeholder="Filter URL…"
      bind:value={filterUrl}
    />

    <select class="filter-select" bind:value={filterMethod}>
      {#each METHOD_OPTIONS as opt}
        <option value={opt.value}>{opt.label}</option>
      {/each}
    </select>

    <select class="filter-select" bind:value={filterStatus}>
      {#each STATUS_OPTIONS.filter((o, i, arr) => arr.findIndex(x => x.value === o.value) === i) as opt}
        <option value={opt.value}>{opt.label}</option>
      {/each}
    </select>

    <span class="console-count">{filtered.length} / {$historyStore.length}</span>

    <Button variant="secondary" size="small" click={handleExport} disabled={$historyStore.length === 0}>
      Export HAR
    </Button>
    <Button variant="tertiary" size="small" click={() => historyStore.clear()} disabled={$historyStore.length === 0}>
      Clear
    </Button>
  </div>

  <!-- Entry list -->
  <div class="console-list">
    {#if filtered.length === 0}
      <div class="console-empty">
        {#if $historyStore.length === 0}
          No requests yet — send one to see it here
        {:else}
          No entries match the current filters
        {/if}
      </div>
    {:else}
      {#each filtered as entry (entry.id)}
        <ConsoleEntry {entry} />
      {/each}
    {/if}
  </div>
</div>

<style>
  .console {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: var(--bg-primary);
    overflow: hidden;
  }

  .console-toolbar {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
    padding: var(--space-xs) var(--space-md);
    border-bottom: 1px solid var(--border);
    background: var(--bg-primary);
    flex-shrink: 0;
  }

  .console-title {
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-semibold);
    color: var(--text-muted);
    flex-shrink: 0;
  }

  .filter-input {
    flex: 1;
    min-width: 0;
    padding: 2px var(--space-sm);
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    color: var(--text);
    font-size: var(--font-size-sm);
    font-family: var(--font-mono);
  }
  .filter-input:focus {
    outline: none;
    border-color: var(--primary);
  }

  .filter-select {
    padding: 2px var(--space-sm);
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    color: var(--text);
    font-size: var(--font-size-sm);
    cursor: pointer;
  }
  .filter-select:focus {
    outline: none;
    border-color: var(--primary);
  }

  .console-count {
    font-size: 0.7rem;
    color: var(--text-muted);
    font-family: var(--font-mono);
    flex-shrink: 0;
  }

  .console-list {
    flex: 1;
    overflow-y: auto;
  }

  .console-empty {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: var(--text-muted);
    font-size: var(--font-size-sm);
    font-style: italic;
  }
</style>
