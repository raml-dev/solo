<script lang="ts">
  import { Execute } from "../../../../wailsjs/go/main/App";
  import { collection, main } from "../../../../wailsjs/go/models";
  import Button from "../base/Button.svelte";
  import Dropdown from "../base/Dropdown.svelte";
  import { collectionStore, selectedRequest } from "../../stores/collectionStore";
  import { selectedEnvironment } from "../../stores/environmentStore";
  import Modal from "../base/Modal.svelte";
  import { envAutocomplete } from "../../actions/envAutocomplete";
  import Tabs from "../base/Tabs.svelte";
  import Tab from "../base/Tab.svelte";
  import RequestHeaders from "./RequestHeaders.svelte";
  import RequestBody from "./RequestBody.svelte";
  import RequestSettings from "./RequestSettings.svelte";
  import { configurationStore } from "../../stores/configurationStore";
  import type { configuration as conf } from "../../../../wailsjs/go/models";

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

  interface TextSegment {
    text: string;
    isToken: boolean;
  }

  let method = "GET";
  let url = "";
  let activeTab = Symbol();
  let requestBody = "";
  let headers: Header[] = [
    { id: "1", key: "Content-Type", value: "application/json", enabled: true },
    { id: "2", key: "test", value: "test", enabled: true }
  ];

  let response: HTTPResponse | null = null;
  let loading = false;
  let responseTab = "body";
  let showSaveDialog = false;
  let requestName = "";
  let responseHeight = 300;
  let isResizing = false;
  let builderElement: HTMLElement;
  let urlInputElement: HTMLInputElement;
  let urlScrollLeft = 0;
  let urlSegments: TextSegment[] = [];
  let environmentEntries: { key: string; value: string }[] = [];

  $: environmentEntries = Object.entries($selectedEnvironment?.values ?? {}).map(([key, val]) => ({
    key,
    value: String(val?.value ?? "")
  }));
  $: urlSegments = splitTextSegments(url);

  function splitTextSegments(value: string): TextSegment[] {
    if (!value) return [];

    const segments: TextSegment[] = [];
    const tokenRegex = /(\{\{[^{}\r\n]+?\}\})/g;
    let cursor = 0;

    for (const match of value.matchAll(tokenRegex)) {
      const index = match.index ?? 0;
      const token = match[0];

      if (index > cursor) {
        segments.push({ text: value.slice(cursor, index), isToken: false });
      }

      segments.push({ text: token, isToken: true });
      cursor = index + token.length;
    }

    if (cursor < value.length) {
      segments.push({ text: value.slice(cursor), isToken: false });
    }

    return segments;
  }

  function resolveEnvironmentTokens(value: string): string {
    if (!value) return value;

    const envMap = new Map(environmentEntries.map((entry) => [entry.key, entry.value]));
    return value.replace(/\{\{([^{}\r\n]+?)\}\}/g, (_fullMatch, key: string) => {
      const cleanKey = key.trim();
      return envMap.has(cleanKey) ? envMap.get(cleanKey)! : _fullMatch;
    });
  }

  // --- Settings Override State ---
  let requestSettings: conf.RequestSettingsOverride = {};
  const { config: globalConfig } = configurationStore;

  function startResize() {
    isResizing = true;
    window.addEventListener("mousemove", handleResize);
    window.addEventListener("mouseup", stopResize);
    document.body.style.userSelect = "none";
  }

  function handleResize(e: MouseEvent) {
    if (!isResizing || !builderElement) return;

    const rect = builderElement.getBoundingClientRect();
    const newHeight = rect.bottom - e.clientY;

    const minHeight = 40;
    const maxHeight = rect.height - 150;

    responseHeight = Math.max(minHeight, Math.min(newHeight, maxHeight));
  }

  function stopResize() {
    isResizing = false;
    window.removeEventListener("mousemove", handleResize);
    window.removeEventListener("mouseup", stopResize);
    document.body.style.userSelect = "";
  }

  // Load request from collection when selected
  $: if ($selectedRequest) {
    loadRequestData($selectedRequest);
  }

  function loadRequestData(request: collection.Request) {
    method = request.verb || "GET";
    url = request.url || "";
    requestBody = request.body || "";
    requestName = request.name || "";

    // Initialize per-request settings, falling back to global defaults for unset boolean values
    const settings = request.settings || {};
    const globalDefaults = $globalConfig.request;

    requestSettings = {
      timeoutSeconds: settings.timeoutSeconds, // Let placeholder handle visual default for numbers/strings
      defaultUserAgent: settings.defaultUserAgent,
      proxyUrl: settings.proxyUrl,
      // For booleans, we must resolve to a concrete true/false for the checkbox
      followRedirects: settings.followRedirects ?? globalDefaults.followRedirects,
      validateSSL: settings.validateSSL ?? globalDefaults.validateSSL,
      // maxRedirects depends on followRedirects, so we handle it in the form
      maxRedirects: settings.maxRedirects
    };

    // Convert headers object to array
    if (request.headers && typeof request.headers === "object") {
      headers = Object.entries(request.headers).map(([key, value], index) => ({
        id: `header-${index}`,
        key,
        value: String(value),
        enabled: true
      }));
    } else {
      headers = [];
    }
  }

  const methodOptions = [
    { value: "GET", label: "GET" },
    { value: "POST", label: "POST" },
    { value: "PUT", label: "PUT" },
    { value: "DELETE", label: "DELETE" },
    { value: "PATCH", label: "PATCH" },
    { value: "HEAD", label: "HEAD" },
    { value: "OPTIONS", label: "OPTIONS" }
  ];

  function handleMethodChange(value: string) {
    method = value;
  }

  async function sendRequest() {
    loading = true;
    const resolvedUrl = resolveEnvironmentTokens(url);
    const resolvedRequestBody = resolveEnvironmentTokens(requestBody);
    const resolvedHeaders = headers
      .filter((h) => h.enabled)
      .reduce(
        (acc, { key, value }) => ({
          ...acc,
          [resolveEnvironmentTokens(key)]: resolveEnvironmentTokens(value)
        }),
        {}
      );

    const requestOptions = new main.RequestOptions({
      body: resolvedRequestBody,
      headers: resolvedHeaders,
      method,
      url: resolvedUrl,
      settings: requestSettings // Pass per-request settings
    });

    try {
      const responseData = await Execute(requestOptions);

      response = {
        status: responseData.statusCode,
        statusText: "TBD",

        time: responseData.duration,
        headers: responseData.headers,
        body: responseData.body
      };
    } catch (error) {
      console.error("Request failed:", error);
    } finally {
      loading = false;
    }
  }

  async function handleSaveToCollection() {
    if (!$collectionStore.selectedCollectionName) {
      alert("Please select a collection first");
      return;
    }

    try {
      const headersObj = headers
        .filter((h) => h.enabled && h.key)
        .reduce((acc, { key, value }) => ({ ...acc, [key]: value }), {});

      // Check if we're updating an existing request or creating a new one
      if ($selectedRequest && $selectedRequest.id) {
        // Update existing request
        await collectionStore.updateRequest(
          $collectionStore.selectedCollectionName,
          collection.Request.createFrom({
            ...$selectedRequest,
            name: requestName || $selectedRequest.name,
            url,
            verb: method,
            body: requestBody,
            headers: headersObj,
            settings: requestSettings, // Save settings with the request
            lastUpdateTimestamp: new Date().toISOString()
          })
        );
      } else {
        // Create new request
        await collectionStore.addRequest($collectionStore.selectedCollectionName, {
          name: requestName || "Untitled Request",
          url,
          verb: method,
          body: requestBody,
          headers: headersObj,
          settings: requestSettings // Save settings with the request
        });
      }

      showSaveDialog = false;
      requestName = "";
    } catch (err) {
      console.error("Error saving request:", err);
      alert("Failed to save request");
    }
  }

  function getStatusClass(status: number): string {
    if (status >= 200 && status < 300) return "status-success";
    if (status >= 300 && status < 400) return "status-info";
    if (status >= 400 && status < 500) return "status-warning";
    return "status-error";
  }
</script>

<div class="request-builder" bind:this={builderElement}>
  <!-- Request Line -->
  <div class="request-line">
    <div class="method-dropdown">
      <Dropdown bind:value={method} options={methodOptions} change={handleMethodChange} />
    </div>
    <div class="url-input-wrapper">
      <div class="url-input-overlay" aria-hidden="true">
        <div
          class="url-input-overlay-content"
          style={`transform: translateX(-${urlScrollLeft}px);`}
        >
          {#if !url}
            <span class="url-placeholder">Enter request URL</span>
          {:else}
            {#each urlSegments as segment (segment)}
              <span class:url-token={segment.isToken}>{segment.text}</span>
            {/each}
          {/if}
        </div>
      </div>
      <input
        bind:this={urlInputElement}
        type="text"
        class="input url-input token-input"
        bind:value={url}
        on:input={() => (urlScrollLeft = urlInputElement?.scrollLeft ?? 0)}
        on:scroll={() => (urlScrollLeft = urlInputElement?.scrollLeft ?? 0)}
        use:envAutocomplete={{ entries: environmentEntries, insertMode: "token" }}
      />
    </div>
    <Button variant="secondary" click={() => (showSaveDialog = true)} style="min-width: 80px;"
      >{$selectedRequest && $selectedRequest.id ? "UPDATE" : "SAVE"}</Button
    >
    <Button
      variant="primary"
      click={sendRequest}
      disabled={loading}
      style="min-width: 100px;font-weight: var(--font-weight-semibold);"
      >{loading ? "SENDING..." : "SEND"}</Button
    >
  </div>

  <div class="request-content">
    <Tabs {activeTab}>
      <Tab title="Headers" value={activeTab}>
        <RequestHeaders {headers} />
      </Tab>
      <Tab title="Body">
        <RequestBody {requestBody} />
      </Tab>
      <Tab title="Settings">
        <RequestSettings bind:requestSettings globalConfig={$globalConfig} />
      </Tab>
    </Tabs>
  </div>

  <!-- Response Section -->
  <div class="response-section" style="height: {responseHeight}px">
    <div class="resize-handle" class:resizing={isResizing} on:mousedown={startResize}></div>
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
        <button
          class="tab"
          class:active={responseTab === "body"}
          on:click={() => (responseTab = "body")}
        >
          Body
        </button>
        <button
          class="tab"
          class:active={responseTab === "headers"}
          on:click={() => (responseTab = "headers")}
        >
          Headers
        </button>
      </div>

      <!-- Response Content -->
      <div class="response-content">
        {#if responseTab === "body"}
          <pre class="response-body"><code>{response.body}</code></pre>
        {:else}
          <div class="response-headers">
            {#each Object.entries(response.headers) as [key, value] ([key])}
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

{#if showSaveDialog}
  <div class="dialog">
    <Modal toggleFn={() => (showSaveDialog = false)}>
      {#if $selectedRequest && $selectedRequest.id}
        <!-- Update existing request -->
        <h3>Update Request</h3>
        {#if !$collectionStore.selectedCollectionName}
          <p class="warning">Please select a collection from the sidebar first!</p>
        {:else}
          <p class="info">
            Do you want to update <strong>{$selectedRequest.name}</strong> in
            <strong>{$collectionStore.selectedCollectionName}</strong>?
          </p>
        {/if}
      {:else}
        <!-- Create new request -->
        <h3>Save Request to Collection</h3>
        {#if !$collectionStore.selectedCollectionName}
          <p class="warning">Please select a collection from the sidebar first!</p>
        {:else}
          <p class="info">
            Saving to collection: <strong>{$collectionStore.selectedCollectionName}</strong>
          </p>
          <!-- svelte-ignore a11y-autofocus -->
          <input
            type="text"
            bind:value={requestName}
            placeholder="Request name"
            on:keydown={(e) => e.key === "Enter" && handleSaveToCollection()}
            autofocus
          />
        {/if}
      {/if}
      <svelte:fragment slot="additional-buttons">
        {#if $collectionStore.selectedCollectionName}
          {#if $selectedRequest && $selectedRequest.id}
            <Button variant="primary" click={handleSaveToCollection}>Update</Button>
          {:else}
            <Button variant="primary" click={handleSaveToCollection}>Save</Button>
          {/if}
        {/if}
      </svelte:fragment>
    </Modal>
  </div>
{/if}

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
    min-width: 100px;
  }

  .url-input {
    flex: 1;
  }

  .url-input-wrapper {
    flex: 1;
    min-width: 0;
    position: relative;
  }

  .url-input-overlay {
    position: absolute;
    inset: 0;
    pointer-events: none;
    z-index: 1;
    padding: var(--space-sm) var(--space-md);
    border: 1px solid transparent;
    font-size: var(--font-size-sm);
    font-family: var(--font-sans);
    white-space: pre;
    overflow: hidden;
  }

  .url-input-overlay-content {
    min-width: 100%;
    color: var(--text);
  }

  .url-placeholder {
    color: var(--text-light);
  }

  .url-token {
    color: var(--primary);
    font-weight: var(--font-weight-semibold);
  }

  .token-input {
    position: relative;
    z-index: 2;
    background: transparent;
    color: transparent;
    caret-color: var(--text);
  }

  .token-input::selection {
    background: rgba(74, 158, 255, 0.25);
    color: transparent;
  }

  .btn-send {
    min-width: 100px;
    font-weight: var(--font-weight-semibold);
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

  .response-section {
    border-top: 1px solid var(--border);
    display: flex;
    flex-direction: column;
    background: var(--bg-primary);
    position: relative;
    flex-shrink: 0;
  }

  .resize-handle {
    position: absolute;
    top: -4px;
    left: 0;
    right: 0;
    height: 8px;
    cursor: ns-resize;
    z-index: 10;
  }

  .resize-handle::after {
    content: "";
    position: absolute;
    left: 0;
    right: 0;
    top: 3px;
    height: 2px;
    background: transparent;
    transition: background-color 0.2s;
  }

  .resize-handle:hover::after,
  .resize-handle.resizing::after {
    background: var(--primary);
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

  .dialog-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .dialog {
    background: var(--color-bg);
    padding: var(--spacing-xl);
    border-radius: var(--border-radius-lg);
    min-width: 400px;
    box-shadow: 0 4px 24px rgba(0, 0, 0, 0.2);
  }

  .dialog h3 {
    margin: 0 0 var(--spacing-md) 0;
    font-size: var(--font-size-xl);
  }

  .dialog .warning {
    color: var(--color-warning);
    margin-bottom: var(--spacing-md);
  }

  .dialog .info {
    color: var(--color-text-secondary);
    margin-bottom: var(--spacing-md);
  }

  .dialog input {
    width: 100%;
    padding: var(--spacing-sm);
    border: 1px solid var(--color-border);
    border-radius: var(--border-radius);
    background: var(--color-bg-secondary);
    color: var(--color-text);
    font-size: var(--font-size-md);
    margin-bottom: var(--spacing-md);
  }

  .dialog input:focus {
    outline: none;
    border-color: var(--color-primary);
  }

  .dialog-actions {
    display: flex;
    gap: var(--spacing-sm);
    justify-content: flex-end;
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
    height: 100%;
    background: var(--bg-tertiary);
  }

  :global(.env-autocomplete-menu) {
    position: absolute;
    z-index: 2000;
    background: var(--bg-primary);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-lg);
    max-height: 280px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
  }

  :global(.env-autocomplete-item) {
    appearance: none;
    border: none;
    background: transparent;
    color: inherit;
    text-align: left;
    cursor: pointer;
    width: 100%;
    padding: var(--space-sm) var(--space-md);
    display: flex;
    justify-content: space-between;
    gap: var(--space-md);
    font-size: var(--font-size-sm);
  }

  :global(.env-autocomplete-item:hover),
  :global(.env-autocomplete-item.active) {
    background: var(--bg-tertiary);
  }

  :global(.env-autocomplete-item .env-key) {
    color: var(--text);
    font-weight: var(--font-weight-semibold);
    white-space: nowrap;
  }

  :global(.env-autocomplete-item .env-value) {
    color: var(--text-muted);
    font-family: var(--font-mono);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 60%;
  }
</style>
