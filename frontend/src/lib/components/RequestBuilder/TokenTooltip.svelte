<script lang="ts">
  import { tokenTooltipStore, cancelHideTokenTooltip, tooltipMouseLeave, forceHideTokenTooltip } from "../../stores/tokenTooltipStore";
  import { selectedEnvironment, environmentStore } from "../../stores/environmentStore";
  import { environment } from "../../../../wailsjs/go/models";
  import { tick } from "svelte";

  let isEditing = false;
  let editValue = "";
  let inputElement: HTMLInputElement;

  $: state = $tokenTooltipStore;
  $: tokenKey = state.tokenKey;
  $: visible = state.visible;
  
  $: envValue = $selectedEnvironment?.values[tokenKey]?.value ?? "";
  $: exists = $selectedEnvironment?.values[tokenKey] !== undefined;

  $: if (visible && !isEditing) {
    editValue = envValue;
  }

  $: if (!visible) {
    isEditing = false;
  }

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
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div
    class="token-tooltip"
    style="left: {state.x}px; top: {state.y}px;"
    on:mouseenter={cancelHideTokenTooltip}
    on:mouseleave={tooltipMouseLeave}
  >
    {#if isEditing}
      <div class="edit-mode">
        <input 
          bind:this={inputElement}
          type="text" 
          bind:value={editValue} 
          on:keydown={handleKeydown} 
          class="tooltip-input"
        />
        <div class="edit-actions">
          <button class="btn-save" on:click={save}>Save</button>
          <button class="btn-cancel" on:click={() => isEditing = false}>Cancel</button>
        </div>
      </div>
    {:else}
      <div class="preview-mode">
        <span class="preview-value" class:unresolved={!exists}>
          {exists ? envValue : "Unresolved variable"}
        </span>
        <button class="edit-btn" on:click={handleEditClick} title="Edit value">
          ✎
        </button>
      </div>
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

  .preview-value {
    font-family: var(--font-mono);
    overflow-wrap: anywhere;
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
