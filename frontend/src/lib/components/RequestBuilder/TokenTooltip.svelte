<!-- @migration-task Error while migrating Svelte code: can't migrate `let isEditing = false;` to `$state` because there's a variable named state.
     Rename the variable and try again or migrate by hand. -->
<script lang="ts">
  import { run } from "svelte/legacy";

  import { environmentStore, selectedEnvironment } from "$src/lib/stores/environmentStore";
  import { sessionVarsStore } from "$src/lib/stores/sessionVarsStore";
  import {
    cancelHideTokenTooltip,
    forceHideTokenTooltip,
    tokenTooltipStore,
    tooltipMouseLeave
  } from "$src/lib/stores/tokenTooltipStore";
  import { environment } from "$wails/go/models";
  import { tick } from "svelte";

  let isEditing = $state(false);
  let editValue = $state("");
  let inputElement: HTMLInputElement | undefined = $state();

  let tooltipState = $derived($tokenTooltipStore);
  let tokenKey = $derived(tooltipState.tokenKey);
  let visible = $derived(tooltipState.visible);

  let sessionValue = $derived(tokenKey ? $sessionVarsStore[tokenKey] : undefined);
  let envValue = $derived($selectedEnvironment?.values[tokenKey]?.value ?? "");
  let hasSessionValue = $derived(sessionValue !== undefined);
  let hasEnvValue = $derived($selectedEnvironment?.values[tokenKey] !== undefined);
  let displayValue = $derived(hasSessionValue ? String(sessionValue) : envValue);
  let exists = $derived(hasSessionValue || hasEnvValue);
  let valueSource = $derived(hasSessionValue ? "session" : hasEnvValue ? "environment" : "none");

  run(() => {
    if (visible && !isEditing) {
      // Editing always targets the persisted environment value, not the session override.
      editValue = envValue;
    }
  });

  run(() => {
    if (!visible) {
      isEditing = false;
    }
  });

  async function handleEditClick() {
    isEditing = true;
    await tick();
    inputElement?.focus();
  }

  async function save() {
    if (!$selectedEnvironment) return;
    const processedValues = { ...$selectedEnvironment.values };
    processedValues[tokenKey] = new environment.ValueType({
      value: editValue,
      type: "string"
    });
    const envToSave = new environment.Environment({
      ...$selectedEnvironment,
      values: processedValues
    });
    await environmentStore.updateEnvironment(envToSave);
    isEditing = false;
  }

  async function clearSessionOverride() {
    if (!tokenKey) return;
    await sessionVarsStore.remove(tokenKey);
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === "Enter") {
      save();
    } else if (e.key === "Escape") {
      isEditing = false;
      forceHideTokenTooltip();
    }
  }
</script>

{#if visible}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="token-tooltip"
    style="left: {tooltipState.x}px; top: {tooltipState.y}px;"
    onmouseenter={cancelHideTokenTooltip}
    onmouseleave={tooltipMouseLeave}
  >
    {#if isEditing}
      <div class="edit-mode">
        <input
          bind:this={inputElement}
          type="text"
          bind:value={editValue}
          onkeydown={handleKeydown}
          class="tooltip-input"
        />
        <div class="edit-actions">
          <button class="btn-save" onclick={save}>Save</button>
          <button class="btn-cancel" onclick={() => (isEditing = false)}>Cancel</button>
        </div>
      </div>
    {:else}
      <div class="preview-mode">
        <div class="preview-main">
          <span class="preview-value" class:unresolved={!exists}>
            {exists ? displayValue : "Unresolved variable"}
          </span>
          {#if valueSource !== "none"}
            <span class="source-label" class:session={valueSource === "session"}>
              {valueSource}
            </span>
          {/if}
        </div>
        <button class="edit-btn" onclick={handleEditClick} title="Edit environment value">
          ✎
        </button>
      </div>
      {#if valueSource === "session"}
        <div class="session-hint-wrap">
          <div class="session-hint">
            Session override attivo: modificando qui cambi l'env salvato, ma il valore effettivo
            resta quello di sessione finché non la svuoti.
          </div>
          <button class="session-clear-btn" onclick={clearSessionOverride}>Use env value</button>
        </div>
      {/if}
    {/if}
  </div>
{/if}

<style>
  .token-tooltip {
    position: fixed;
    z-index: 9999;
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-lg);
    padding: var(--space-xs) var(--space-sm);
    display: inline-block;
    width: max-content;
    max-width: 90vw;
    font-family: var(--font-sans);
    font-size: var(--font-size-sm);
    color: var(--text);
    transform: translateY(8px);
  }

  .preview-mode {
    display: inline-flex;
    align-items: center;
    gap: var(--space-sm);
    max-width: 90vw;
  }

  .session-hint-wrap {
    margin-top: var(--space-xs);
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-sm);
    max-width: 70ch;
  }

  .session-hint {
    color: var(--text-muted);
    font-size: var(--font-size-xs);
    line-height: 1.35;
  }

  .session-clear-btn {
    flex-shrink: 0;
    background: none;
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    color: var(--text);
    font-size: var(--font-size-xs);
    padding: 2px 8px;
    cursor: pointer;
  }

  .session-clear-btn:hover {
    border-color: var(--primary);
    color: var(--primary);
  }

  .preview-main {
    display: inline-flex;
    align-items: center;
    gap: var(--space-xs);
    min-width: 0;
  }

  .preview-value {
    font-family: var(--font-mono);
    overflow-wrap: anywhere;
  }

  .source-label {
    font-size: 0.62rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--text-muted);
    border: 1px solid var(--border);
    background: var(--bg-primary);
    border-radius: var(--radius-sm);
    padding: 1px 6px;
    flex-shrink: 0;
  }

  .source-label.session {
    color: var(--primary);
    border-color: color-mix(in srgb, var(--primary) 35%, transparent);
    background: color-mix(in srgb, var(--primary) 14%, transparent);
  }

  .preview-value.unresolved {
    color: var(--danger);
    font-style: italic;
    font-family: var(--font-sans);
  }

  .edit-btn {
    background: none;
    border: none;
    cursor: pointer;
    color: var(--text-muted);
    font-size: var(--font-size-md);
    padding: var(--space-xs);
    border-radius: var(--radius-sm);
    flex-shrink: 0;
  }

  .edit-btn:hover {
    color: var(--primary);
    background: var(--bg-tertiary);
  }

  .edit-mode {
    display: flex;
    flex-direction: column;
    gap: var(--space-xs);
  }

  .tooltip-input {
    width: 100%;
    padding: var(--space-xs);
    border: 1px solid var(--primary);
    border-radius: var(--radius-sm);
    background: var(--bg-primary);
    color: var(--text);
    font-family: var(--font-mono);
    font-size: var(--font-size-sm);
  }

  .tooltip-input:focus {
    outline: none;
  }

  .edit-actions {
    display: flex;
    justify-content: flex-end;
    gap: var(--space-xs);
  }

  .edit-actions button {
    background: none;
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    cursor: pointer;
    padding: 2px var(--space-sm);
    font-size: var(--font-size-xs);
    color: var(--text);
  }

  .edit-actions .btn-save {
    background: var(--primary);
    color: var(--bg-primary);
    border-color: var(--primary);
  }
</style>
