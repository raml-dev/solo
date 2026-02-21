<script lang="ts">
  import { environmentStore } from "../../stores/environmentStore";
  import type { Environment, EnvironmentValue } from "../../stores/environmentStore";
  import { environment } from "../../../../wailsjs/go/models";
  import { createEventDispatcher } from "svelte";
  import Button from "../base/Button.svelte";

  export let env: Environment;
  export let selected: boolean;
  export let menuOpen: boolean;

  let expanded = false;
  type DraftRow = { id: number; key: string; value: string };
  let draftValues: DraftRow[] = [];
  let nextRowId = 0;

  let dirty = false;

  const dispatch = createEventDispatcher();

  function toggleEnvironment() {
    expanded = !expanded;
    if (expanded && draftValues.length === 0) {
      initDraft();
    }
  }

  function initDraft() {
    const rows: DraftRow[] = Object.entries(env.values || {}).map(
      ([k, v]) => ({
        id: nextRowId++,
        key: k,
        value: v.value,
      })
    );
    rows.push({ id: nextRowId++, key: "", value: "" });
    draftValues = rows;
  }

  function isDirty(): boolean {
    const stored = env.values || {};
    const draftMap: Record<string, string> = {};
    for (const row of draftValues) {
      const k = row.key.trim();
      if (k) draftMap[k] = row.value;
    }
    const draftKeys = Object.keys(draftMap);
    const storedKeys = Object.keys(stored);
    if (draftKeys.length !== storedKeys.length) return true;
    else for (const k of draftKeys) {
      if (!(k in stored)) return true;
      else if (draftMap[k] !== stored[k].value) return true;
    }
    return false;
  }

  function handleUpdateRow(id: number, field: "key" | "value", val: string) {
    const idx = draftValues.findIndex((r) => r.id === id);
    if (idx === -1) return;

    draftValues[idx] = { ...draftValues[idx], [field]: val };

    const isLast = idx === draftValues.length - 1;

    if (isLast && val.trim()) {
      draftValues = [...draftValues, { id: nextRowId++, key: "", value: "" }];
    } else if (!isLast) {
      const updated = draftValues[idx];
      if (!updated.key.trim() && !updated.value.trim()) {
        draftValues.splice(idx, 1);
        draftValues = [...draftValues];
      }
    }
    dirty = isDirty()
  }

  async function handleSaveEnvironment() {
    for (const row of draftValues) {
      if (row.value.trim() && !row.key.trim()) {
        dispatch("error", "A variable has a value but no name. Please add a name or clear the value.");
        return;
      }
    }
    try {
      const processedValues: Record<string, EnvironmentValue> = {};
      for (const row of draftValues) {
        const k = row.key.trim();
        if (k)
          processedValues[k] = new environment.ValueType({
            value: row.value,
            type: "string",
          });
      }
      const envToSave = new environment.Environment({
        ...env,
        values: processedValues,
      });
      await environmentStore.updateEnvironment(envToSave);
      draftValues = [];
      initDraft();
    } catch (err) {
      console.error("Error saving environment:", err);
      dispatch("error", `Error saving environment: ${err}`);
    }
  }

  function selectEnvironment() {
    dispatch("select", env.name);
  }

  function toggleMenu(e: Event) {
    e.stopPropagation();
    dispatch("toggleMenu", env.name);
  }

  function handleDeleteEnvironment() {
    dispatch("delete", env.name);
  }
</script>

<div
  class="environment-item"
  class:selected
  class:menu-open={menuOpen}
>
  <input
    type="checkbox"
    class="env-select"
    checked={selected}
    on:change|stopPropagation={selectEnvironment}
    title="Set as active environment"
    aria-label="Set {env.name} as active environment"
  />
  <div class="environment-accordion">
    <div
      class="environment-header"
      on:click={toggleEnvironment}
      on:keypress={(e) => e.key === "Enter" && toggleEnvironment()}
      role="button"
      tabindex="0"
    >
      <button
        class="expand-btn"
        on:click|stopPropagation={toggleEnvironment}
        aria-label="Toggle environment"
      >
        <span class="expand-icon" class:expanded> &gt; </span>
      </button>

      <div class="environment-info">
        <span class="environment-name">{env.name}</span>
        <span class="environment-count">
          {Object.keys(env.values || {}).length}
        </span>
      </div>

      <div class="environment-actions">
        <button
          class="icon-btn"
          on:click={toggleMenu}
          title="More actions"
          aria-label="More actions"
        >
          ...
        </button>
      </div>

      {#if menuOpen}
        <div class="environment-menu">
          <button
            class="menu-item danger"
            on:click={handleDeleteEnvironment}
          >
            Delete
          </button>
        </div>
      {/if}
    </div>

    {#if expanded}
      <div class="values">
        <div class="values-header">
          <span>Variable</span>
          <span>Value</span>
          <span />
        </div>

        {#each draftValues as row (row.id)}
          <div class="value-row">
            <input
              type="text"
              value={row.key}
              on:input={(e) => handleUpdateRow(row.id, "key", e.currentTarget.value)}
              class="input-sm"
              placeholder="Variable name"
            />
            <input
              type="text"
              value={row.value}
              on:input={(e) => handleUpdateRow(row.id, "value", e.currentTarget.value)}
              class="input-sm"
              placeholder="Value"
            />
          </div>
        {/each}

        <div class="values-footer">
          <Button variant="primary" size="small" disabled={!dirty} click={handleSaveEnvironment}>Save</Button>
        </div>
      </div>
    {/if}
  </div>
</div>
<style>
    .environment-item {
    display: flex;
    align-items: flex-start;
    gap: var(--space-sm);
  }

  .environment-item.selected .environment-accordion {
    border-color: var(--primary);
    box-shadow: var(--shadow-sm);
  }

  .environment-item.menu-open .environment-accordion {
    border-color: var(--border-dark);
  }

  .environment-accordion {
    flex: 1;
    min-width: 0;
    border-radius: var(--radius-md);
    background: var(--bg-primary);
    border: 1px solid transparent;
    overflow: visible;
  }

  .environment-header {
    display: flex;
    align-items: center;
    padding: var(--space-sm) var(--space-md);
    cursor: pointer;
    gap: var(--space-xs);
    position: relative;
    border-radius: var(--radius-md);
  }

  .environment-header:hover {
    background: var(--bg-tertiary);
  }

  .expand-btn {
    background: none;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    padding: 0;
    width: 20px;
    height: 20px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .expand-icon {
    display: inline-block;
    transition: transform var(--transition-fast);
  }

  .expand-icon.expanded {
    transform: rotate(90deg);
  }

  .environment-info {
    display: flex;
    align-items: center;
    gap: var(--space-xs);
    flex: 1;
    min-width: 0;
  }

  .environment-name {
    font-weight: var(--font-weight-medium);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .environment-count {
    color: var(--text-muted);
    font-size: var(--font-size-xs);
    background: var(--bg-tertiary);
    padding: 0 var(--space-xs);
    border-radius: var(--radius-sm);
  }

  .environment-actions {
    display: flex;
    gap: var(--space-xs);
    opacity: 0;
    pointer-events: none;
    transition: opacity var(--transition-fast);
  }

  .environment-header:hover .environment-actions,
  .environment-item.menu-open .environment-actions {
    opacity: 1;
    pointer-events: auto;
  }

  .icon-btn {
    background: none;
    border: 1px solid transparent;
    cursor: pointer;
    padding: 0 var(--space-xs);
    border-radius: var(--radius-sm);
    color: var(--text-muted);
    transition: all var(--transition-fast);
    font-size: var(--font-size-sm);
    height: 24px;
  }

  .icon-btn:hover {
    background: var(--bg-tertiary);
    color: var(--text);
  }

  .env-select {
    margin-top: calc(var(--space-sm) + 2px);
    flex-shrink: 0;
    cursor: pointer;
    accent-color: var(--primary);
    width: 14px;
    height: 14px;
  }

  .environment-menu {
    position: absolute;
    right: var(--space-sm);
    top: calc(100% + 6px);
    background: var(--bg-primary);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-md);
    display: flex;
    flex-direction: column;
    min-width: 140px;
    z-index: var(--z-dropdown);
  }

  .menu-item {
    padding: var(--space-sm) var(--space-md);
    background: none;
    border: none;
    text-align: left;
    font-size: var(--font-size-sm);
    color: var(--text);
    cursor: pointer;
  }

  .menu-item:hover {
    background: var(--bg-tertiary);
  }

  .menu-item.danger {
    color: var(--danger);
  }

  .menu-item.danger:hover {
    background: var(--status-danger-bg);
  }

  .values {
    background: var(--bg-secondary);
    padding: 0 var(--space-sm) var(--space-sm) calc(var(--space-lg) + 8px);
    display: flex;
    flex-direction: column;
    gap: var(--space-xs);
  }

  .values-header {
    display: grid;
    grid-template-columns: 1fr 1fr auto;
    gap: var(--space-sm);
    padding: var(--space-sm);
    font-size: var(--font-size-xs);
    font-weight: var(--font-weight-semibold);
    color: var(--text-muted);
    border-bottom: 1px solid var(--border);
  }

  .value-row {
    display: grid;
    grid-template-columns: 1fr 1fr auto;
    gap: var(--space-sm);
    padding: var(--space-xs) 0;
    align-items: center;
  }

  .input-sm {
    padding: var(--space-xs) var(--space-sm);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    background: var(--bg-secondary);
    color: var(--text);
    font-size: var(--font-size-sm);
    width: 100%;
    margin-bottom: 0;
  }

  .input-sm:focus {
    outline: none;
    border-color: var(--primary);
  }

  .input-sm:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .values-footer {
    display: flex;
    justify-content: flex-end;
    padding-top: var(--space-xs);
  }

</style>
