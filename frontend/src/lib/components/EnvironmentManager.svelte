<script lang="ts">
  import { environmentStore } from "../stores/environmentStore";
  import type { Environment, EnvironmentValue } from "../stores/environmentStore";
  import { environment } from "../../../wailsjs/go/models";
  import Button from "./base/Button.svelte";
  import Modal from "./base/Modal.svelte";

  let showNewEnvironmentDialog = false;
  let showDeleteConfirmDialog = false;
  let newEnvironmentName = "";
  let deleteTarget: string | null = null;
  let expandedEnvironments: Set<string> = new Set();
  let activeMenu: string | null = null;

  // For editing values
  let editingEnvironment: Environment | null = null;
  let newValueKey = "";
  let newValueValue = "";
  let newValueType = "string";

  $: environments = $environmentStore.environments;
  $: selectedEnvironmentName = $environmentStore.selectedEnvironmentName;

  function isExpanded(environmentName: string): boolean {
    return expandedEnvironments.has(environmentName);
  }

  function toggleEnvironment(environmentName: string) {
    if (expandedEnvironments.has(environmentName)) {
      expandedEnvironments.delete(environmentName);
    } else {
      expandedEnvironments.add(environmentName);
    }
    expandedEnvironments = new Set(expandedEnvironments);
  }

  function selectEnvironment(name: string) {
    environmentStore.selectEnvironment(name);
  }

  function closeNewEnvironmentDialog() {
    showNewEnvironmentDialog = false;
    newEnvironmentName = "";
  }

  async function handleCreateEnvironment() {
    const trimmed = newEnvironmentName.trim();
    if (!trimmed) {
      return;
    }

    const exists = environments.some((env) => env.name.toLowerCase() === trimmed.toLowerCase());
    if (exists) {
      alert(`Environment "${trimmed}" already exists.`);
      return;
    }

    try {
      await environmentStore.createEnvironment(trimmed);
      closeNewEnvironmentDialog();
    } catch (err) {
      console.error("Error creating environment:", err);
      alert(`Error creating environment: ${err}`);
    }
  }

  function handleDeleteEnvironment(environmentName: string) {
    deleteTarget = environmentName;
    showDeleteConfirmDialog = true;
    activeMenu = null;
  }

  async function confirmDelete() {
    if (!deleteTarget) return;

    try {
      await environmentStore.deleteEnvironment(deleteTarget);
      closeDeleteConfirmDialog();
    } catch (err) {
      console.error("Error deleting environment:", err);
      alert(`Error deleting environment: ${err}`);
    }
  }

  function closeDeleteConfirmDialog() {
    showDeleteConfirmDialog = false;
    deleteTarget = null;
  }

  function toggleMenu(e: Event, environmentName: string) {
    e.stopPropagation();
    activeMenu = activeMenu === environmentName ? null : environmentName;
  }

  function openEditEnvironment(env: Environment) {
    // Deep copy the environment using the Environment class
    const copiedValues: Record<string, EnvironmentValue> = {};
    for (const [key, val] of Object.entries(env.values || {})) {
      copiedValues[key] = new environment.ValueType(val);
    }
    editingEnvironment = new environment.Environment({
      ...env,
      values: copiedValues
    });
    activeMenu = null;
  }

  function closeEditDialog() {
    editingEnvironment = null;
    newValueKey = "";
    newValueValue = "";
    newValueType = "default";
  }

  async function handleAddValue() {
    if (!editingEnvironment) return;
    const key = newValueKey.trim();
    const value = newValueValue.trim();

    if (!key) {
      alert("Value name is required");
      return;
    }

    if (editingEnvironment.values[key]) {
      alert(`Value "${key}" already exists`);
      return;
    }

    // Create a proper ValueType instance
    const valueType = new environment.ValueType({
      value,
      type: newValueType
    });
    editingEnvironment.values[key] = valueType;
    // Trigger reactivity with new Environment instance
    editingEnvironment = new environment.Environment(editingEnvironment);

    newValueKey = "";
    newValueValue = "";
    newValueType = "string";
  }

  async function handleRemoveValue(key: string) {
    if (!editingEnvironment) return;
    delete editingEnvironment.values[key];
    // Trigger reactivity with new Environment instance
    editingEnvironment = new environment.Environment(editingEnvironment);
  }

  async function handleUpdateValue(key: string, value: string) {
    if (!editingEnvironment) return;
    // Create a new ValueType instance with updated value
    const existingType = editingEnvironment.values[key];
    editingEnvironment.values[key] = new environment.ValueType({
      value,
      type: existingType.type
    });
    // Trigger reactivity with new Environment instance
    editingEnvironment = new environment.Environment(editingEnvironment);
  }

  async function handleSaveEnvironment() {
    if (!editingEnvironment) return;

    try {
      // Ensure all values are proper ValueType instances
      const processedValues: Record<string, EnvironmentValue> = {};
      for (const [key, val] of Object.entries(editingEnvironment.values || {})) {
        processedValues[key] = new environment.ValueType(val);
      }

      // Create a new Environment instance with processed values
      const envToSave = new environment.Environment({
        ...editingEnvironment,
        values: processedValues
      });

      await environmentStore.updateEnvironment(envToSave);
      closeEditDialog();
    } catch (err) {
      console.error("Error saving environment:", err);
      alert(`Error saving environment: ${err}`);
    }
  }
</script>

<div class="environment-list">
  <div class="header">
    <div class="header-title">
      <h3>Environments</h3>
      <Button variant="primary" size="small" click={() => (showNewEnvironmentDialog = true)}>
        New
      </Button>
    </div>
  </div>

  {#if $environmentStore.loading}
    <div class="loading">Loading environments...</div>
  {/if}

  {#if $environmentStore.error}
    <div class="error">
      {$environmentStore.error}
      <button on:click={() => environmentStore.clearError()}>x</button>
    </div>
  {/if}

  <div class="environments">
    {#each environments as environment (environment.id)}
      <div
        class="environment-item"
        class:selected={selectedEnvironmentName === environment.name}
        class:menu-open={activeMenu === environment.name}
      >
        <div
          class="environment-header"
          on:click={() => {
            selectEnvironment(environment.name);
            toggleEnvironment(environment.name);
          }}
          on:keypress={(e) => e.key === "Enter" && selectEnvironment(environment.name)}
          role="button"
          tabindex="0"
        >
          <button
            class="expand-btn"
            on:click={(e) => {
              e.stopPropagation();
              toggleEnvironment(environment.name);
            }}
            aria-label="Toggle environment"
          >
            <span class="expand-icon" class:expanded={isExpanded(environment.name)}> &gt; </span>
          </button>

          <div class="environment-info">
            <span class="environment-name">{environment.name}</span>
            <span class="environment-count">
              {Object.keys(environment.values || {}).length}
            </span>
          </div>

          <div class="environment-actions">
            <button
              class="icon-btn"
              on:click={(e) => {
                e.stopPropagation();
                openEditEnvironment(environment);
              }}
              title="Edit environment"
              aria-label="Edit environment"
            >
              ✎
            </button>
            <button
              class="icon-btn"
              on:click={(e) => toggleMenu(e, environment.name)}
              title="More actions"
              aria-label="More actions"
            >
              ...
            </button>
          </div>

          {#if activeMenu === environment.name}
            <div class="environment-menu">
              <button
                class="menu-item danger"
                on:click={() => handleDeleteEnvironment(environment.name)}
              >
                Delete
              </button>
            </div>
          {/if}
        </div>

        {#if isExpanded(environment.name)}
          <div class="values">
            {#if Object.keys(environment.values || {}).length === 0}
              <div class="empty-values">No variables yet</div>
            {:else}
              {#each Object.entries(environment.values || {}) as [key, val] ([key])}
                <div class="value-item">
                  <span class="value-key">{key}</span>
                  <span class="value-value">{val.value || "(empty)"}</span>
                </div>
              {/each}
            {/if}
          </div>
        {/if}
      </div>
    {/each}
  </div>
</div>

{#if environments.length === 0 && !$environmentStore.loading}
  <div class="empty-state">
    <p>No environments yet</p>
    <p class="hint">Create your first environment to get started</p>
  </div>
{/if}

{#if showNewEnvironmentDialog}
  <Modal toggleFn={() => (showNewEnvironmentDialog = false)}>
    <h3>New Environment</h3>
    <!-- svelte-ignore a11y-autofocus -->
    <input
      type="text"
      bind:value={newEnvironmentName}
      placeholder="Environment name"
      on:keydown={(e) => e.key === "Enter" && handleCreateEnvironment()}
      autofocus
    />
    <svelte:fragment slot="additional-buttons">
      <Button variant="primary" click={handleCreateEnvironment}>Create</Button>
    </svelte:fragment>
  </Modal>
{/if}

{#if showDeleteConfirmDialog}
  <Modal toggleFn={closeDeleteConfirmDialog}>
    <h3>Delete Environment</h3>
    <p>Are you sure you want to delete "{deleteTarget}"?</p>
    <p class="warning">This action cannot be undone.</p>
    <svelte:fragment slot="additional-buttons">
      <Button variant="danger" click={confirmDelete}>Delete</Button>
    </svelte:fragment>
  </Modal>
{/if}

{#if editingEnvironment}
  <Modal toggleFn={closeEditDialog}>
    <h3>Edit Environment: {editingEnvironment.name}</h3>

    <div class="values-editor">
      <div class="values-header">
        <span>Variable</span>
        <span>Value</span>
        <span>Type</span>
        <span></span>
      </div>

      {#each Object.entries(editingEnvironment.values || {}) as [key, val] ([key])}
        <div class="value-row">
          <input type="text" value={key} disabled class="input-sm" />
          <input
            type="text"
            value={val.value}
            on:input={(e) => handleUpdateValue(key, e.currentTarget.value)}
            class="input-sm"
            placeholder="Value"
          />
          <input type="text" value={val.type} disabled class="input-sm input-type" />
          <button
            class="icon-btn danger"
            on:click={() => handleRemoveValue(key)}
            title="Remove variable"
          >
            x
          </button>
        </div>
      {/each}

      <div class="value-row new-value">
        <input type="text" bind:value={newValueKey} placeholder="Variable name" class="input-sm" />
        <input type="text" bind:value={newValueValue} placeholder="Value" class="input-sm" />
        <input
          type="text"
          bind:value={newValueType}
          placeholder="Type"
          class="input-sm input-type"
        />
        <button class="icon-btn" on:click={handleAddValue} title="Add variable"> + </button>
      </div>
    </div>

    <svelte:fragment slot="additional-buttons">
      <Button variant="primary" click={handleSaveEnvironment}>Save</Button>
    </svelte:fragment>
  </Modal>
{/if}

<style>
  .environment-list {
    display: flex;
    flex-direction: column;
    flex: 1;
    overflow: hidden;
    padding: var(--space-md);
    border-bottom: 1px solid var(--border);
    display: flex;
    flex-direction: column;
    gap: var(--space-sm);
  }
  .header-title {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .header h3 {
    margin: 0;
    font-size: var(--font-size-lg);
    font-weight: var(--font-weight-semibold);
  }

  .loading {
    padding: var(--space-md);
    text-align: center;
    color: var(--text-muted);
  }

  .error {
    margin: var(--space-md);
    padding: var(--space-sm);
    background: var(--status-danger-bg);
    color: var(--status-danger-text);
    border-radius: var(--radius-md);
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: var(--font-size-sm);
  }

  .error button {
    background: none;
    border: none;
    color: inherit;
    font-size: var(--font-size-lg);
    cursor: pointer;
    padding: 0 var(--space-xs);
  }
  .environments {
    flex: 1;
    overflow-y: auto;
    padding: var(--space-sm);
    display: flex;
    flex-direction: column;
    gap: var(--space-xs);
  }

  .environment-item {
    border-radius: var(--radius-md);
    background: var(--bg-primary);
    border: 1px solid transparent;
    overflow: visible;
  }

  .environment-item.selected {
    border-color: var(--primary);
    box-shadow: var(--shadow-sm);
  }

  .environment-item.menu-open {
    border-color: var(--border-dark);
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

  .icon-btn.danger {
    color: var(--danger);
  }

  .icon-btn.danger:hover {
    background: var(--status-danger-bg);
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

  .value-item {
    display: flex;
    align-items: center;
    padding: var(--space-xs) var(--space-sm);
    gap: var(--space-sm);
    border-radius: var(--radius-sm);
    background: var(--bg-primary);
    border: 1px solid var(--border);
  }

  .value-key {
    font-family: var(--font-mono);
    font-size: var(--font-size-xs);
    font-weight: var(--font-weight-semibold);
    color: var(--primary);
    min-width: 100px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .value-value {
    flex: 1;
    font-size: var(--font-size-xs);
    color: var(--text-muted);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .empty-values {
    color: var(--text-muted);
    font-size: var(--font-size-xs);
    padding: var(--space-xs) var(--space-sm);
  }

  .empty-state {
    padding: var(--space-xl);
    text-align: center;
    color: var(--text-muted);
  }

  .empty-state p {
    margin: var(--space-xs) 0;
  }

  .empty-state .hint {
    font-size: var(--font-size-sm);
  }

  .warning {
    color: var(--text-muted);
    font-size: var(--font-size-sm);
    margin-bottom: var(--space-md);
  }

  input {
    width: 100%;
    padding: var(--space-sm);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    background: var(--bg-secondary);
    color: var(--text);
    font-size: var(--font-size-md);
    margin-bottom: var(--space-md);
  }

  input:focus {
    outline: none;
    border-color: var(--primary);
  }

  .values-editor {
    margin-bottom: var(--space-md);
  }

  .values-header {
    display: grid;
    grid-template-columns: 1fr 2fr 100px 40px;
    gap: var(--space-sm);
    padding: var(--space-sm);
    font-size: var(--font-size-xs);
    font-weight: var(--font-weight-semibold);
    color: var(--text-muted);
    border-bottom: 1px solid var(--border);
  }

  .value-row {
    display: grid;
    grid-template-columns: 1fr 2fr 100px 40px;
    gap: var(--space-sm);
    padding: var(--space-xs) 0;
    align-items: center;
  }

  .value-row.new-value {
    margin-top: var(--space-sm);
    padding-top: var(--space-sm);
    border-top: 1px solid var(--border);
  }

  .input-sm {
    padding: var(--space-xs) var(--space-sm);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    background: var(--bg-secondary);
    color: var(--text);
    font-size: var(--font-size-sm);
    width: 100%;
  }

  .input-sm:focus {
    outline: none;
    border-color: var(--primary);
  }

  .input-sm:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .input-type {
    font-size: var(--font-size-xs);
  }
</style>
