<script lang="ts">
  import {Execute} from '../../../wailsjs/go/main/App';
  import {main} from '../../../wailsjs/go/models';
  import Button from './base/Button.svelte';
  import Dropdown from './base/Dropdown.svelte';

  interface Header {
    id: string;
    key: string;
    value: string;
    enabled: boolean;
  }

  interface HTTPResponse {
    status: number;
    statusText: string;
    time: number;
    headers: Record<string, string>;
    body: string;
  }

  let method = "GET";
  let url = "";
  let activeTab = "headers";
  let requestBody = "";
  let headers: Header[] = [
    { id: '1', key: 'Content-Type', value: 'application/json', enabled: true },
    { id: '2', key: 'test', value: 'test', enabled: true },
  ];
  
  let response: HTTPResponse | null = null;
  let loading = false;
  let responseTab = "body";

  const methodOptions = [
    { value: "GET", label: "GET" },
    { value: "POST", label: "POST" },
    { value: "PUT", label: "PUT" },
    { value: "DELETE", label: "DELETE" },
    { value: "PATCH", label: "PATCH" },
    { value: "HEAD", label: "HEAD" },
    { value: "OPTIONS", label: "OPTIONS" },
  ];

  function handleMethodChange(event: CustomEvent<string>) {
    method = event.detail;
  }

  function addHeader() {
    const newHeader: Header = {
      id: Date.now().toString(),
      key: '',
      value: '',
      enabled: true
    };
    headers = [...headers, newHeader];
  }

  function removeHeader(id: string) {
    headers = headers.filter(h => h.id !== id);
  }

  function toggleHeader(id: string) {
    headers = headers.map(h => 
      h.id === id ? { ...h, enabled: !h.enabled } : h
    );
  }

  async function sendRequest() {
    loading = true;

    const requestOptions: main.RequestOptions = {
      body: requestBody,
      headers: headers.filter(h => h.enabled).reduce((acc, {key, value}) => ({...acc, [key]: value}), {}),
      method,
      url
    }

    try {
      // Simulate API call - replace with actual Wails backend call
      const responseData = await Execute(requestOptions)
      
      response = {
        status: responseData.statusCode,
        statusText: "TBD",
        
        time: responseData.duration,
        headers: responseData.headers,
        body: responseData.body
      };
    } catch (error) {
      console.error('Request failed:', error);
    } finally {
      loading = false;
    }
  }

  function getStatusClass(status: number): string {
    if (status >= 200 && status < 300) return 'status-success';
    if (status >= 300 && status < 400) return 'status-info';
    if (status >= 400 && status < 500) return 'status-warning';
    return 'status-error';
  }
</script>

<div class="request-builder">
  <!-- Request Line -->
  <div class="request-line">
    <div class="method-dropdown">
      <Dropdown 
        bind:value={method}
        options={methodOptions}
        on:change={handleMethodChange}
      />
    </div>
    <input 
      type="text" 
      class="input url-input" 
      placeholder="Enter request URL"
      bind:value={url}
    />
    <Button 
      variant="primary"
      on:click={sendRequest}
      disabled={loading}
      style="min-width: 100px;font-weight: var(--font-weight-semibold);"
      >{loading ? 'SENDING...' : 'SEND'}</Button>
  </div>

  <!-- Request Tabs -->
  <div class="tabs">
    <button class="tab" class:active={activeTab === "headers"} on:click={() => activeTab = "headers"}>
      Headers
    </button>
    <button class="tab" class:active={activeTab === "body"} on:click={() => activeTab = "body"}>
      Body
    </button>
  </div>

  <!-- Request Content -->
  <div class="request-content">
    {#if activeTab === "headers"}
      <div class="headers-editor">
        {#each headers as header (header.id)}
          <div class="header-row">
            <input 
              type="checkbox" 
              class="header-checkbox"
              checked={header.enabled}
              on:change={() => toggleHeader(header.id)}
            />
            <input 
              type="text" 
              class="input header-input"
              placeholder="Header name"
              bind:value={header.key}
              disabled={!header.enabled}
            />
            <input 
              type="text" 
              class="input header-input"
              placeholder="Value"
              bind:value={header.value}
              disabled={!header.enabled}
            />
            <button 
              class="btn-icon btn-remove"
              on:click={() => removeHeader(header.id)}
              title="Remove header"
            >
              ×
            </button>
          </div>
        {/each}
        <button class="btn-add-header" on:click={addHeader}>
          + Add Header
        </button>
      </div>
    {:else if activeTab === "body"}
      <div class="body-editor">
        <textarea 
          class="input code-input" 
          rows="12" 
          placeholder="Request body (JSON, XML, etc.)"
          bind:value={requestBody}
        ></textarea>
      </div>
    {/if}
  </div>

  <!-- Response Section -->
  <div class="response-section">
    <div class="response-header-bar">
      <h3 class="text-base font-semibold">RESPONSE</h3>
      {#if response}
        <div class="response-meta">
          <span class="status-badge {getStatusClass(response.status)}">
            Status: {response.status}
          </span>
          <span class="time-badge">Time: {response.time}ms</span>
        </div>
      {/if}
    </div>

    {#if response}
      <!-- Response Tabs -->
      <div class="tabs response-tabs">
        <button class="tab" class:active={responseTab === "body"} on:click={() => responseTab = "body"}>
          Body
        </button>
        <button class="tab" class:active={responseTab === "headers"} on:click={() => responseTab = "headers"}>
          Headers
        </button>
      </div>

      <!-- Response Content -->
      <div class="response-content">
        {#if responseTab === "body"}
          <pre class="response-body"><code>{response.body}</code></pre>
        {:else}
          <div class="response-headers">
            {#each Object.entries(response.headers) as [key, value]}
              <div class="response-header-row">
                <span class="header-key">{key}:</span>
                <span class="header-value">{value}</span>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    {:else}
      <div class="response-empty">
        <p class="text-muted">Send a request to see the response</p>
      </div>
    {/if}
  </div>
</div>

<style>
  .request-builder {
    display: flex;
    flex-direction: column;
    height: 100%;
    width: 100%;
    background: var(--bg-primary);
  }

  .request-line {
    display: flex;
    gap: var(--space-sm);
    padding: var(--space-lg);
    background: var(--bg-primary);
    border-bottom: 1px solid var(--border);
  }

  .method-dropdown {
    min-width: 120px;
  }

  .url-input {
    flex: 1;
  }

  .btn-send {
    min-width: 100px;
    font-weight: var(--font-weight-semibold);
  }

  .tabs {
    display: flex;
    background: var(--bg-primary);
    border-bottom: 1px solid var(--border);
    padding: 0 var(--space-lg);
  }

  .response-tabs {
    padding: 0 var(--space-md);
  }

  .request-content {
    flex: 1;
    overflow-y: auto;
    background: var(--bg-secondary);
    min-height: 200px;
  }

  .headers-editor {
    padding: var(--space-md);
  }

  .header-row {
    display: flex;
    gap: var(--space-sm);
    margin-bottom: var(--space-sm);
    align-items: center;
  }

  .header-checkbox {
    width: 18px;
    height: 18px;
    cursor: pointer;
  }

  .header-input {
    flex: 1;
  }

  .header-input:disabled {
    opacity: 0.5;
    background: var(--bg-tertiary);
  }

  .btn-icon {
    width: 32px;
    height: 32px;
    border: none;
    background: var(--bg-tertiary);
    color: var(--text-muted);
    border-radius: var(--radius-md);
    cursor: pointer;
    font-size: var(--font-size-xl);
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all var(--transition-fast);
  }

  .btn-remove:hover {
    background: var(--danger);
    color: var(--bg-primary);
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

  .body-editor {
    padding: var(--space-md);
  }

  .code-input {
    width: 100%;
    font-family: var(--font-mono);
    font-size: var(--font-size-sm);
    line-height: var(--line-height-relaxed);
    resize: vertical;
  }

  .response-section {
    border-top: 1px solid var(--border);
    display: flex;
    flex-direction: column;
    max-height: 50%;
    background: var(--bg-primary);
  }

  .response-header-bar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--space-md) var(--space-lg);
    background: var(--bg-secondary);
  }

  .response-meta {
    display: flex;
    gap: var(--space-md);
    align-items: center;
  }

  .status-badge {
    padding: var(--space-xs) var(--space-md);
    border-radius: var(--radius-md);
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-semibold);
  }

  .status-success {
    background: var(--success);
    color: var(--bg-primary);
  }

  .status-info {
    background: var(--info);
    color: var(--bg-primary);
  }

  .status-warning {
    background: var(--warning);
    color: var(--bg-primary);
  }

  .status-error {
    background: var(--danger);
    color: var(--bg-primary);
  }

  .time-badge {
    font-size: var(--font-size-sm);
    color: var(--text-muted);
  }

  .response-content {
    flex: 1;
    overflow-y: auto;
    overflow-x: auto;
    background: var(--bg-tertiary);
    max-width: 100%;
  }

  .response-body {
    padding: var(--space-lg);
    margin: 0;
    font-family: var(--font-mono);
    font-size: var(--font-size-sm);
    line-height: var(--line-height-relaxed);
    color: var(--text);
    white-space: pre-wrap;
    word-wrap: break-word;
    overflow-wrap: break-word;
    max-width: 100%;
  }

  .response-body code {
    font-family: inherit;
  }

  .response-headers {
    padding: var(--space-md);
  }

  .response-header-row {
    display: flex;
    padding: var(--space-sm) var(--space-md);
    border-bottom: 1px solid var(--border);
    font-family: var(--font-mono);
    font-size: var(--font-size-sm);
    gap: var(--space-md);
  }

  .header-key {
    font-weight: var(--font-weight-semibold);
    color: var(--primary);
    min-width: 200px;
    flex-shrink: 0;
  }

  .header-value {
    color: var(--text);
    word-break: break-all;
    overflow-wrap: break-word;
    flex: 1;
  }

  .response-empty {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 200px;
    background: var(--bg-tertiary);
  }
</style>
