<script lang="ts">
  import { Execute } from "../../../../wailsjs/go/main/App";
  import { collection, main } from "../../../../wailsjs/go/models";
  import Button from "../base/Button.svelte";
  import Dropdown from "../base/Dropdown.svelte";
  import EmptyState from "../base/EmptyState.svelte";
  import { collectionStore } from "../../stores/collectionStore";
  import { tabStore, activeTab as activeTabState } from "../../stores/tabStore";
  import { selectedEnvironment } from "../../stores/environmentStore";
  import Tabs from "../base/Tabs.svelte";
  import Tab from "../base/Tab.svelte";
  import RequestHeaders from "./RequestHeaders.svelte";
  import RequestBody from "./RequestBody.svelte";
  import CodeMirrorEditor from "./CodeMirrorEditor.svelte";
  import RequestSettings from "./RequestSettings.svelte";
  import TokenInput from "./TokenInput.svelte";
  import TokenTooltip from "./TokenTooltip.svelte";
  import { configurationStore } from "../../stores/configurationStore";
  import type { configuration as conf } from "../../../../wailsjs/go/models";
  import type { InputFormat } from "./types";
  import SaveRequestModal from "../SaveRequestModal.svelte";
  import { historyStore } from "../../stores/historyStore";

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

  // --- Per-tab local form state (mirrors the active tab's stored state) ---
  let method = "GET";
  let url = "";
  let requestBody = "";
  let requestBodyFormat: InputFormat = "none";
  let headers: Header[] = [];
  let requestSettings: conf.RequestSettingsOverride = {};
  let requestName = "";

  // Internal tab tracking
  let activeBuilderTabId: string | null = null;
  let isLoadingTab = false;

  // UI state (not per-tab, stays across switches)
  let requestPaneTab = "Body"; // Headers, Body, Settings
  let response: HTTPResponse | null = null;
  let requestError: string | null = null;
  let loading = false;
  let responseTab = "body";
  let showSaveDialog = false;
  let responseHeight = 300;
  let isResizing = false;
  let builderElement: HTMLElement;
  let environmentEntries: { key: string; value: string }[] = [];

  const { config: globalConfig } = configurationStore;

  $: environmentEntries = Object.entries($selectedEnvironment?.values ?? {}).map(([key, val]) => ({
    key,
    value: String(val?.value ?? "")
  }));

  // --- Autosync: When the request body format changes, update Content-Type header ---
  $: {
    const contentTypeHeader = headers.find((h) => h.key.toLowerCase() === "content-type");
    if (contentTypeHeader) {
      switch (requestBodyFormat) {
        case "json":  contentTypeHeader.value = "application/json"; break;
        case "xml":   contentTypeHeader.value = "application/xml";  break;
        case "text":  contentTypeHeader.value = "text/plain";       break;
        case "none":  contentTypeHeader.value = "";                 break;
      }
      headers = [...headers];
    }
  }

  // --- Tab switching: load form state from tab store ---
  $: if ($activeTabState?.id && $activeTabState.id !== activeBuilderTabId) {
    loadTabIntoForm($activeTabState);
    activeBuilderTabId = $activeTabState.id;
  }

  $: if (!$activeTabState && activeBuilderTabId) {
    // No tab open: reset form
    activeBuilderTabId = null;
    method = "GET";
    url = "";
    requestBody = "";
    requestBodyFormat = "json";
    headers = [];
    requestSettings = {};
    requestName = "";
    response = null;
  }

  function loadTabIntoForm(tab: typeof $activeTabState) {
    if (!tab) return;
    isLoadingTab = true;
    method = tab.verb || "GET";
    url = tab.url || "";
    requestBody = tab.body || "";
    requestBodyFormat = (tab.bodyFormat as InputFormat) || "none";
    headers = tab.headers ? [...tab.headers] : [];
    requestSettings = { ...(tab.settings || {}) };
    requestName = tab.label || "";
    response = null; // clear response on tab switch
    setTimeout(() => { isLoadingTab = false; }, 0);
  }

  // --- Persist form state back to tab store on every change ---
  $: if (!isLoadingTab && activeBuilderTabId && (url !== undefined || method || requestBody !== undefined || headers || requestSettings)) {
    tabStore.updateTabFormState(activeBuilderTabId, {
      verb: method,
      url,
      body: requestBody,
      bodyFormat: requestBodyFormat,
      headers,
      settings: requestSettings
    });
  }

  // --- Autosave to backend for saved requests ---
  const lastPersistedSignatureByTabId: Record<string, string> = {};
  const persistTimers = new Map<string, ReturnType<typeof setTimeout>>();

  function normalizeSettingsSignature(settings: conf.RequestSettingsOverride = {}) {
    return {
      timeoutSeconds: settings.timeoutSeconds ?? null,
      defaultUserAgent: settings.defaultUserAgent ?? null,
      proxyUrl: settings.proxyUrl ?? null,
      followRedirects: settings.followRedirects ?? null,
      validateSSL: settings.validateSSL ?? null,
      maxRedirects: settings.maxRedirects ?? null
    };
  }

  function buildSignature(payload: {
    id: string; name: string; url: string; verb: string;
    body: string; headers: Record<string, string>; settings: conf.RequestSettingsOverride;
  }) {
    return JSON.stringify({
      ...payload,
      settings: normalizeSettingsSignature(payload.settings)
    });
  }

  function findRequestInStore(requestId: string): collection.Request | null {
    for (const coll of $collectionStore.collections) {
      const found = coll.requests.find((r) => r.id === requestId);
      if (found) return found;
    }
    return null;
  }

  $: if (!isLoadingTab && $activeTabState?.requestId && $activeTabState.collectionName) {
    const tab = $activeTabState;
    const headersObj = headers
      .filter((h) => h.enabled && h.key)
      .reduce((acc, { key, value }) => ({ ...acc, [key]: value }), {});

    const sig = buildSignature({
      id: tab.requestId!,
      name: requestName || tab.label,
      url, verb: method, body: requestBody,
      headers: headersObj,
      settings: requestSettings
    });

    const tabId = tab.id;
    const collName = tab.collectionName;
    const reqId = tab.requestId!;

    const pending = persistTimers.get(tabId);
    if (pending) clearTimeout(pending);

    const snapshot = {
      collectionName: collName,
      requestId: reqId,
      name: requestName || tab.label,
      url, verb: method, body: requestBody,
      headersObj, settings: { ...requestSettings },
      signature: sig
    };

    const timer = setTimeout(async () => {
      persistTimers.delete(tabId);
      await persistRequest(tabId, snapshot);
    }, 800);
    persistTimers.set(tabId, timer);
  }

  async function persistRequest(tabId: string, snapshot: {
    collectionName: string;
    requestId: string;
    name: string; url: string; verb: string; body: string;
    headersObj: Record<string, string>;
    settings: conf.RequestSettingsOverride;
    signature: string;
  }) {
    const last = lastPersistedSignatureByTabId[tabId];
    if (snapshot.signature === last) return;

    const stored = findRequestInStore(snapshot.requestId);
    if (!stored) return;

    try {
      await collectionStore.updateRequest(
        snapshot.collectionName,
        collection.Request.createFrom({
          ...stored,
          name: snapshot.name,
          url: snapshot.url,
          verb: snapshot.verb,
          body: snapshot.body,
          headers: snapshot.headersObj,
          settings: snapshot.settings,
          lastUpdateTimestamp: new Date().toISOString()
        })
      );
      lastPersistedSignatureByTabId[tabId] = snapshot.signature;
      tabStore.markDirty(tabId, false);
    } catch (err) {
      // autosave failure is silent — store already notifies
    }
  }

  // --- Resize ---
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
    responseHeight = Math.max(40, Math.min(newHeight, rect.height - 150));
  }

  function stopResize() {
    isResizing = false;
    window.removeEventListener("mousemove", handleResize);
    window.removeEventListener("mouseup", stopResize);
    document.body.style.userSelect = "";
  }

  // --- Methods ---
  const methodOptions = [
    { value: "GET", label: "GET" },
    { value: "POST", label: "POST" },
    { value: "PUT", label: "PUT" },
    { value: "DELETE", label: "DELETE" },
    { value: "PATCH", label: "PATCH" },
    { value: "HEAD", label: "HEAD" },
    { value: "OPTIONS", label: "OPTIONS" }
  ];

  const bodyFormatOptions = [
    { value: "none", label: "None" },
    { value: "json", label: "JSON" },
    { value: "xml",  label: "XML"  },
    { value: "text", label: "Text" }
  ];

  function handleMethodChange(value: string) {
    method = value;
    if (activeBuilderTabId) {
      tabStore.updateTabFormState(activeBuilderTabId, { verb: value });
    }
  }

  function handleBodyFormatChange(value: string) {
    requestBodyFormat = value as InputFormat;
    if (activeBuilderTabId) {
      tabStore.updateTabFormState(activeBuilderTabId, { bodyFormat: value });
    }
  }

  function formatBody() {
    if (!requestBody?.trim()) return;

    if (requestBodyFormat === "json") {
      try {
        const parsed = JSON.parse(requestBody);
        requestBody = JSON.stringify(parsed, null, 2);
      } catch {
        // invalid JSON — do nothing
      }
    } else if (requestBodyFormat === "xml") {
      requestBody = prettifyXml(requestBody);
    }
    // text: nothing to format
  }

  function prettifyXml(xml: string): string {
    const INDENT = "  ";
    let formatted = "";
    let depth = 0;
    // split on tags
    const parts = xml.replace(/>\s*</g, "><").split(/(<[^>]+>)/);
    for (const part of parts) {
      if (!part.trim()) continue;
      if (part.startsWith("</")) {
        // closing tag — dedent first
        depth = Math.max(0, depth - 1);
        formatted += INDENT.repeat(depth) + part + "\n";
      } else if (part.startsWith("<?") || part.startsWith("<!")) {
        // declaration / comment — same level
        formatted += INDENT.repeat(depth) + part + "\n";
      } else if (part.startsWith("<") && !part.endsWith("/>")) {
        // opening tag
        formatted += INDENT.repeat(depth) + part + "\n";
        depth++;
      } else if (part.startsWith("<") && part.endsWith("/>")) {
        // self-closing tag
        formatted += INDENT.repeat(depth) + part + "\n";
      } else {
        // text content — attach to previous line if possible
        const lines = formatted.trimEnd().split("\n");
        const last = lines[lines.length - 1];
        if (last && last.trimEnd().endsWith(">")) {
          lines[lines.length - 1] = last + part;
          formatted = lines.join("\n") + "\n";
        } else {
          formatted += INDENT.repeat(depth) + part + "\n";
        }
      }
    }
    return formatted.trim();
  }

  function resolveEnvironmentTokens(value: string): string {
    if (!value) return value;
    const envMap = new Map(environmentEntries.map((e) => [e.key, e.value]));
    return value.replace(/\{\{([^{}\r\n]+?)\}\}/g, (_full, key: string) => {
      const k = key.trim();
      return envMap.has(k) ? envMap.get(k)! : _full;
    });
  }

  async function sendRequest() {
    if (!$activeTabState?.requestId) {
      showSaveDialog = true;
      return;
    }
    loading = true;
    const resolvedUrl = resolveEnvironmentTokens(url);
    const resolvedBody = resolveEnvironmentTokens(requestBody);
    const resolvedHeaders = headers
      .filter((h) => h.enabled)
      .reduce((acc, { key, value }) => ({
        ...acc,
        [resolveEnvironmentTokens(key)]: resolveEnvironmentTokens(value)
      }), {});

    const requestOptions = new main.RequestOptions({
      body: resolvedBody,
      headers: resolvedHeaders,
      method,
      url: resolvedUrl,
      settings: requestSettings
    });

    try {
      const responseData = await Execute(requestOptions);
      requestError = null;
      const rawBody = responseData.body ?? "";
      const fmt = detectResponseFormat(responseData.headers ?? {});
      const prettyBody = prettyPrint(rawBody, fmt);
      response = {
        status: responseData.statusCode,
        statusText: "TBD",
        time: responseData.duration,
        headers: responseData.headers,
        body: prettyBody
      };
      historyStore.push({
        collectionName: $activeTabState?.collectionName ?? null,
        requestName: requestName || null,
        request: { method, url: resolvedUrl, headers: resolvedHeaders, body: resolvedBody },
        response: { status: responseData.statusCode, time: responseData.duration, headers: responseData.headers, body: rawBody },
        error: null,
      });
    } catch (error) {
      response = null;
      requestError = String(error);
      historyStore.push({
        collectionName: $activeTabState?.collectionName ?? null,
        requestName: requestName || null,
        request: { method, url: resolvedUrl, headers: resolvedHeaders, body: resolvedBody },
        response: null,
        error: requestError,
      });
    } finally {
      loading = false;
    }
  }

  function prettyPrint(body: string, fmt: "json" | "xml" | "text"): string {
    if (!body?.trim()) return body;
    if (fmt === "json") {
      try { return JSON.stringify(JSON.parse(body), null, 2); } catch { return body; }
    }
    if (fmt === "xml") {
      return prettifyXml(body);
    }
    return body;
  }

  function detectResponseFormat(headers: Record<string, string>): "json" | "xml" | "text" {
    const ct = Object.entries(headers ?? {})
      .find(([k]) => k.toLowerCase() === "content-type")?.[1] ?? "";
    if (ct.includes("json")) return "json";
    if (ct.includes("xml") || ct.includes("html")) return "xml";
    return "text";
  }

  $: responseFormat = response ? detectResponseFormat(response.headers) : "text";

  async function handleSaveRequest(event: CustomEvent<{ name: string; collection: string }>) {
    const { name, collection: targetCollection } = event.detail;
    if (!targetCollection || !activeBuilderTabId) return;

    try {
      const headersObj = headers
        .filter((h) => h.enabled && h.key)
        .reduce((acc, { key, value }) => ({ ...acc, [key]: value }), {});

      await collectionStore.addRequest(targetCollection, {
        name: name || "Untitled Request",
        url, verb: method, body: requestBody,
        headers: headersObj, settings: requestSettings
      });

      // The new request is now the selected one in collectionStore
      // CollectionList.openTab will fire; bind this tab to it
      showSaveDialog = false;
      await sendRequest();
    } catch (err) {
      // error already shown by store
    }
  }

  function getStatusClass(status: number): string {
    if (status >= 200 && status < 300) return "status-success";
    if (status >= 300 && status < 400) return "status-info";
    if (status >= 400 && status < 500) return "status-warning";
    return "status-error";
  }
</script>

{#if $activeTabState}
  <div class="request-builder" bind:this={builderElement}>
    <TokenTooltip />
    <!-- Request Line -->
    <div class="request-line">
      <div class="url-bar">
        <div class="url-bar-method">
          <Dropdown bind:value={method} options={methodOptions} change={handleMethodChange} variant="url-method" square />
        </div>
        <div class="url-bar-divider" />
        <TokenInput
          bind:value={url}
          placeholder="Enter request URL"
          {environmentEntries}
          wrapperClass="url-bar-input"
        />
        <div class="url-bar-divider" />
        <Button
          variant="primary"
          click={sendRequest}
          disabled={loading}
          square
          style="padding: 0 var(--space-xl); font-weight: var(--font-weight-semibold); align-self: stretch; border-radius: 0 var(--radius-md) var(--radius-md) 0;"
        >{loading ? "Sending…" : "Send"}</Button>
      </div>
    </div>

    <div class="request-content-bar">
      <Tabs bind:activeValue={requestPaneTab} variant="minimal">
        <Tab title="Headers" value="Headers" />
        <Tab title="Body" value="Body" />
        <Tab title="Settings" value="Settings" />
      </Tabs>

      {#if requestPaneTab === 'Body'}
        <div class="body-format-selector">
          {#if requestBodyFormat !== 'none'}
            <button
              class="beautify-btn"
              title="Prettify / Format body"
              on:click={formatBody}
              disabled={requestBodyFormat === 'text'}
            >
              <svg width="14" height="14" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M2 4h12M2 8h8M2 12h10" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
              </svg>
              Beautify
            </button>
            <span class="format-separator">|</span>
          {/if}
          <Dropdown
            bind:value={requestBodyFormat}
            options={bodyFormatOptions}
            change={handleBodyFormatChange}
            variant="minimal"
          />
        </div>
      {/if}
    </div>

    <div class="request-content-body">
      {#if requestPaneTab === 'Headers'}
        {#key $activeTabState.id}
          <RequestHeaders {headers} />
        {/key}
      {:else if requestPaneTab === 'Body'}
        {#key $activeTabState.id}
          {#if requestBodyFormat === 'none'}
            <EmptyState message="This request does not have a body" />
          {:else}
            <RequestBody bind:requestBody bind:format={requestBodyFormat} />
          {/if}
        {/key}
      {:else if requestPaneTab === 'Settings'}
        {#key $activeTabState.id}
          <RequestSettings bind:requestSettings globalConfig={$globalConfig} />
        {/key}
      {/if}
    </div>

    <!-- Response Section -->
    <div class="response-section" style="height: {responseHeight}px">
      <div class="resize-handle" class:resizing={isResizing} on:mousedown={startResize}></div>

      <div class="response-content-bar">
        <Tabs bind:activeValue={responseTab} variant="minimal">
          <Tab title="Body" value="body" />
          <Tab title="Headers" value="headers" />
        </Tabs>
        {#if response}
          <div class="response-meta">
            <span class="status-badge {getStatusClass(response.status)}">{response.status}</span>
            <span class="time-badge">{response.time}ms</span>
          </div>
        {/if}
      </div>

      {#if requestError}
        <div class="response-error">
          <span class="response-error-icon">✕</span>
          <div class="response-error-body">
            <span class="response-error-title">Request failed</span>
            <span class="response-error-detail">{requestError}</span>
          </div>
        </div>
      {:else if response}
        <div class="response-content">
          {#if responseTab === "body"}
            <div class="response-body-editor">
              <CodeMirrorEditor
                value={response.body ?? ""}
                format={responseFormat}
                readOnly={true}
              />
            </div>
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
        <EmptyState message="Send a request to see the response" />
      {/if}
    </div>
  </div>
{:else}
  <EmptyState message="Open a request from the sidebar or press + to start a new one" />
{/if}

<SaveRequestModal
  bind:show={showSaveDialog}
  bind:requestName
  on:save={handleSaveRequest}
  on:cancel={() => (showSaveDialog = false)}
/>

<style>
  .request-builder {
    display: flex;
    flex-direction: column;
    height: 100%;
    width: 100%;
    background: var(--bg-primary);
  }

  .request-line {
    padding: var(--space-md) var(--space-lg);
    background: var(--bg-primary);
    border-bottom: 1px solid var(--border);
  }

  .url-bar {
    display: flex;
    align-items: stretch;
    height: 38px;
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    background: var(--bg-secondary);
    transition: border-color var(--transition-fast);
  }

  .url-bar:focus-within {
    border-color: var(--primary);
  }

  .url-bar-divider {
    width: 1px;
    background: var(--border);
    flex-shrink: 0;
  }

  .url-bar-method {
    display: flex;
    align-items: stretch;
    flex-shrink: 0;
  }

  .url-bar-method :global(.dropdown-trigger.variant-url-method) {
    border-radius: var(--radius-md) 0 0 var(--radius-md);
  }

  :global(.url-bar-input) {
    flex: 1 !important;
    min-width: 0;
    background: transparent !important;
  }

  .response-tabs {
    padding: 0 var(--space-md);
  }

  .response-content-bar {
    display: flex;
    align-items: stretch;
    justify-content: space-between;
    background: var(--bg-primary);
    border-bottom: 1px solid var(--border);
    padding: 0 var(--space-md) 0 0;
    flex-shrink: 0;
  }

  .response-meta {
    display: flex;
    gap: var(--space-sm);
    align-items: center;
    padding: 0 var(--space-xs);
  }

  .request-content {
    display: flex;
    flex-direction: column;
    flex: 1;
    overflow-y: auto;
    background: var(--bg-secondary);
    min-height: 200px;
  }

  .request-content-bar {
    display: flex;
    align-items: stretch;
    justify-content: space-between;
    background: var(--bg-primary);
    border-bottom: 1px solid var(--border);
    padding: 0 var(--space-md) 0 0;
  }

  .body-format-selector {
    display: flex;
    align-items: center;
    gap: var(--space-xs);
    font-size: var(--font-size-xs);
    color: var(--text-muted);
    padding-right: var(--space-xs);
  }

  .beautify-btn {
    display: flex;
    align-items: center;
    gap: 4px;
    background: none;
    border: none;
    color: var(--text-muted);
    font-size: var(--font-size-xs);
    font-family: var(--font-sans);
    font-weight: var(--font-weight-medium);
    padding: 2px var(--space-xs);
    border-radius: var(--radius-sm);
    cursor: pointer;
    transition: color var(--transition-fast), background var(--transition-fast);
    white-space: nowrap;
  }
  .beautify-btn:hover:not(:disabled) {
    color: var(--text);
    background: var(--bg-tertiary);
  }
  .beautify-btn:disabled {
    opacity: 0.35;
    cursor: default;
  }

  .format-separator {
    color: var(--border-dark);
    font-size: var(--font-size-xs);
    user-select: none;
  }

  .format-label {
    font-size: var(--font-size-xs);
    color: var(--text-muted);
    user-select: none;
  }

  .request-content-body {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-height: 0;
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

  .status-badge {
    padding: var(--space-xs) var(--space-md);
    border-radius: var(--radius-md);
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-semibold);
  }

  .status-success { background: var(--success); color: var(--bg-primary); }
  .status-info    { background: var(--info);    color: var(--bg-primary); }
  .status-warning { background: var(--warning); color: var(--bg-primary); }
  .status-error   { background: var(--danger);  color: var(--bg-primary); }

  .time-badge {
    font-size: var(--font-size-sm);
    color: var(--text-muted);
  }

  .response-content {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    background: var(--bg-secondary);
  }

  .response-error {
    flex: 1;
    display: flex;
    align-items: flex-start;
    gap: var(--space-md);
    padding: var(--space-lg) var(--space-xl);
    background: var(--bg-secondary);
    border-top: 2px solid var(--danger);
  }

  .response-error-icon {
    flex-shrink: 0;
    width: 22px;
    height: 22px;
    border-radius: 50%;
    background: var(--danger);
    color: var(--bg-primary);
    font-size: 0.65rem;
    font-weight: var(--font-weight-semibold);
    display: flex;
    align-items: center;
    justify-content: center;
    margin-top: 2px;
  }

  .response-error-body {
    display: flex;
    flex-direction: column;
    gap: var(--space-xs);
  }

  .response-error-title {
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-semibold);
    color: var(--danger);
  }

  .response-error-detail {
    font-size: var(--font-size-sm);
    font-family: var(--font-mono);
    color: var(--text-muted);
    word-break: break-word;
  }

  .response-body-editor {
    flex: 1;
    overflow: hidden;
    display: flex;
    flex-direction: column;
  }

  .response-body code { font-family: inherit; }

  .response-headers { padding: var(--space-md); }

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
