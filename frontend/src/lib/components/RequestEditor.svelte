<script lang="ts">
  import { selectedRequest, collectionStore } from '../stores/collectionStore';
  import type { Request } from '../stores/collectionStore';
  import Button from './base/Button.svelte';

  let editedRequest: Partial<Request> | null = null;
  let hasChanges = false;

  $: if ($selectedRequest) {
    editedRequest = { ...$selectedRequest };
    hasChanges = false;
  } else {
    editedRequest = null;
    hasChanges = false;
  }

  function handleChange() {
    hasChanges = true;
  }

  async function handleSave() {
    if (!editedRequest || !$selectedRequest || !$collectionStore.selectedCollectionName) return;
    
    try {
      await collectionStore.updateRequest(
        $collectionStore.selectedCollectionName,
        {
          ...$selectedRequest,
          ...editedRequest,
        } as Request
      );
      hasChanges = false;
    } catch (err) {
      console.error('Error saving request:', err);
    }
  }

  function handleDiscard() {
    if ($selectedRequest) {
      editedRequest = { ...$selectedRequest };
      hasChanges = false;
    }
  }

  function addHeader() {
    if (!editedRequest) return;
    editedRequest.headers = {
      ...editedRequest.headers,
      '': ''
    };
    handleChange();
  }

  function removeHeader(key: string) {
    if (!editedRequest) return;
    const newHeaders = { ...editedRequest.headers };
    delete newHeaders[key];
    editedRequest.headers = newHeaders;
    handleChange();
  }

  function updateHeaderKey(oldKey: string, newKey: string) {
    if (!editedRequest) return;
    const value = editedRequest.headers?.[oldKey];
    const newHeaders = { ...editedRequest.headers };
    delete newHeaders[oldKey];
    if (newKey) {
      newHeaders[newKey] = value;
    }
    editedRequest.headers = newHeaders;
    handleChange();
  }

  function updateHeaderValue(key: string, value: string) {
    if (!editedRequest) return;
    editedRequest.headers = {
      ...editedRequest.headers,
      [key]: value
    };
    handleChange();
  }
</script>

<div class="request-editor">
  {#if editedRequest && $selectedRequest}
    <div class="editor-header">
      <input 
        type="text" 
        class="request-name-input"
        bind:value={editedRequest.name}
        on:input={handleChange}
        placeholder="Request name"
      />
      
      {#if hasChanges}
        <div class="save-actions">
          <Button variant="secondary" size="small" on:click={handleDiscard}>
            Discard
          </Button>
          <Button variant="primary" size="small" on:click={handleSave}>
            Save Changes
          </Button>
        </div>
      {/if}
    </div>

    <div class="request-config">
      <div class="request-line">
        <select 
          bind:value={editedRequest.verb}
          on:change={handleChange}
          class="method-select"
        >
          <option value="GET">GET</option>
          <option value="POST">POST</option>
          <option value="PUT">PUT</option>
          <option value="PATCH">PATCH</option>
          <option value="DELETE">DELETE</option>
          <option value="HEAD">HEAD</option>
          <option value="OPTIONS">OPTIONS</option>
        </select>
        
        <input 
          type="text" 
          class="url-input"
          bind:value={editedRequest.url}
          on:input={handleChange}
          placeholder="https://api.example.com/endpoint"
        />
      </div>
    </div>

    <div class="editor-body">
      <div class="section">
        <div class="section-header">
          <h4>Headers</h4>
          <Button variant="secondary" size="small" on:click={addHeader}>
            + Add Header
          </Button>
        </div>
        
        <div class="headers-list">
          {#if editedRequest.headers && Object.keys(editedRequest.headers).length > 0}
            {#each Object.entries(editedRequest.headers) as [key, value]}
              <div class="header-row">
                <input 
                  type="text" 
                  class="header-key"
                  value={key}
                  on:input={(e) => updateHeaderKey(key, e.currentTarget.value)}
                  placeholder="Header name"
                />
                <input 
                  type="text" 
                  class="header-value"
                  value={value}
                  on:input={(e) => updateHeaderValue(key, e.currentTarget.value)}
                  placeholder="Header value"
                />
                <button 
                  class="remove-btn"
                  on:click={() => removeHeader(key)}
                  title="Remove header"
                >
                  ×
                </button>
              </div>
            {/each}
          {:else}
            <p class="empty-message">No headers defined</p>
          {/if}
        </div>
      </div>

      <div class="section">
        <div class="section-header">
          <h4>Body</h4>
        </div>
        
        <textarea 
          class="body-textarea"
          bind:value={editedRequest.body}
          on:input={handleChange}
          placeholder="Request body (JSON, XML, etc.)"
        />
      </div>

      <div class="section metadata">
        <div class="metadata-item">
          <span class="label">ID:</span>
          <span class="value">{$selectedRequest.id}</span>
        </div>
        <div class="metadata-item">
          <span class="label">Created:</span>
          <span class="value">{new Date($selectedRequest.creationTimestamp).toLocaleString()}</span>
        </div>
        <div class="metadata-item">
          <span class="label">Last Modified:</span>
          <span class="value">{new Date($selectedRequest.lastUpdateTimestamp).toLocaleString()}</span>
        </div>
      </div>
    </div>
  {:else}
    <div class="empty-state">
      <p>Select a request to edit</p>
    </div>
  {/if}
</div>

<style>
  .request-editor {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: var(--color-bg);
    overflow: hidden;
  }

  .editor-header {
    display: flex;
    align-items: center;
    gap: var(--spacing-md);
    padding: var(--spacing-md);
    border-bottom: 1px solid var(--color-border);
  }

  .request-name-input {
    flex: 1;
    padding: var(--spacing-sm);
    border: 1px solid var(--color-border);
    border-radius: var(--border-radius);
    background: var(--color-bg-secondary);
    color: var(--color-text);
    font-size: var(--font-size-lg);
    font-weight: var(--font-weight-semibold);
  }

  .request-name-input:focus {
    outline: none;
    border-color: var(--color-primary);
  }

  .save-actions {
    display: flex;
    gap: var(--spacing-sm);
  }

  .request-config {
    padding: var(--spacing-md);
    border-bottom: 1px solid var(--color-border);
  }

  .request-line {
    display: flex;
    gap: var(--spacing-sm);
  }

  .method-select {
    padding: var(--spacing-sm);
    border: 1px solid var(--color-border);
    border-radius: var(--border-radius);
    background: var(--color-bg-secondary);
    color: var(--color-text);
    font-family: var(--font-mono);
    font-weight: var(--font-weight-bold);
    cursor: pointer;
    min-width: 100px;
  }

  .method-select:focus {
    outline: none;
    border-color: var(--color-primary);
  }

  .url-input {
    flex: 1;
    padding: var(--spacing-sm);
    border: 1px solid var(--color-border);
    border-radius: var(--border-radius);
    background: var(--color-bg-secondary);
    color: var(--color-text);
    font-family: var(--font-mono);
  }

  .url-input:focus {
    outline: none;
    border-color: var(--color-primary);
  }

  .editor-body {
    flex: 1;
    overflow-y: auto;
    padding: var(--spacing-md);
  }

  .section {
    margin-bottom: var(--spacing-lg);
  }

  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: var(--spacing-md);
  }

  .section-header h4 {
    margin: 0;
    font-size: var(--font-size-md);
    font-weight: var(--font-weight-semibold);
  }

  .headers-list {
    display: flex;
    flex-direction: column;
    gap: var(--spacing-sm);
  }

  .header-row {
    display: flex;
    gap: var(--spacing-sm);
    align-items: center;
  }

  .header-key,
  .header-value {
    padding: var(--spacing-sm);
    border: 1px solid var(--color-border);
    border-radius: var(--border-radius);
    background: var(--color-bg-secondary);
    color: var(--color-text);
    font-family: var(--font-mono);
    font-size: var(--font-size-sm);
  }

  .header-key {
    flex: 1;
  }

  .header-value {
    flex: 2;
  }

  .header-key:focus,
  .header-value:focus {
    outline: none;
    border-color: var(--color-primary);
  }

  .remove-btn {
    background: none;
    border: none;
    color: var(--color-text-secondary);
    cursor: pointer;
    font-size: var(--font-size-xl);
    padding: var(--spacing-xs);
    transition: color 0.15s;
  }

  .remove-btn:hover {
    color: var(--color-danger);
  }

  .body-textarea {
    width: 100%;
    min-height: 200px;
    padding: var(--spacing-sm);
    border: 1px solid var(--color-border);
    border-radius: var(--border-radius);
    background: var(--color-bg-secondary);
    color: var(--color-text);
    font-family: var(--font-mono);
    font-size: var(--font-size-sm);
    resize: vertical;
  }

  .body-textarea:focus {
    outline: none;
    border-color: var(--color-primary);
  }

  .metadata {
    padding: var(--spacing-md);
    background: var(--color-bg-secondary);
    border-radius: var(--border-radius);
  }

  .metadata-item {
    display: flex;
    gap: var(--spacing-sm);
    margin-bottom: var(--spacing-xs);
    font-size: var(--font-size-sm);
  }

  .metadata-item .label {
    color: var(--color-text-secondary);
    font-weight: var(--font-weight-medium);
  }

  .metadata-item .value {
    color: var(--color-text);
    font-family: var(--font-mono);
  }

  .empty-state {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: var(--color-text-secondary);
  }

  .empty-message {
    color: var(--color-text-secondary);
    font-size: var(--font-size-sm);
    text-align: center;
    padding: var(--spacing-md);
  }
</style>
