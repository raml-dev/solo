<script lang="ts">
  import { run } from "svelte/legacy";

  import Button from "$src/lib/components/base/Button.svelte";
  import Dropdown from "$src/lib/components/base/Dropdown.svelte";
  import EmptyState from "$src/lib/components/base/EmptyState.svelte";
  import Tab from "$src/lib/components/base/Tab.svelte";
  import Tabs from "$src/lib/components/base/Tabs.svelte";
  import CodeMirrorEditor from "$src/lib/components/RequestBuilder/CodeMirrorEditor.svelte";
  import RequestBody from "$src/lib/components/RequestBuilder/RequestBody.svelte";
  import RequestHeaders from "$src/lib/components/RequestBuilder/RequestHeaders.svelte";
  import RequestRunner from "$src/lib/components/RequestBuilder/RequestRunner.svelte";
  import RequestScripts from "$src/lib/components/RequestBuilder/RequestScripts.svelte";
  import RequestSettings from "$src/lib/components/RequestBuilder/RequestSettings.svelte";
  import TokenInput from "$src/lib/components/RequestBuilder/TokenInput.svelte";
  import TokenTooltip from "$src/lib/components/RequestBuilder/TokenTooltip.svelte";
  import type { InputFormat } from "$src/lib/components/RequestBuilder/types";
  import SaveRequestModal from "$src/lib/components/SaveRequestModal.svelte";
  import { collectionStore } from "$src/lib/stores/collectionStore";
  import { configurationStore } from "$src/lib/stores/configurationStore";
  import { selectedEnvironment } from "$src/lib/stores/environmentStore";
  import { historyStore } from "$src/lib/stores/historyStore";
  import { sessionVarsStore } from "$src/lib/stores/sessionVarsStore";
  import {
    activeTab as activeTabState,
    tabStore,
    type TabResponse
  } from "$src/lib/stores/tabStore";
  import { Execute, GetSessionVars } from "$wails/go/main/App";
  import type { configuration as conf } from "$wails/go/models";
  import { main } from "$wails/go/models";
  import { onMount } from "svelte";

  interface Header {
    id: string;
    key: string;
    value: string;
    enabled: boolean;
  }

  // --- Local form state ---
  let method = $state("GET");
  let url = $state("");
  let requestBody = $state("");
  let requestBodyFormat: InputFormat = $state("none");
  let headers: Header[] = $state([]);
  let requestSettings: conf.RequestSettingsOverride = $state({});
  let requestName = $state("");
  let preRequestScript = $state("");
  let postResponseScript = $state("");

  // Tracks which tab request is currently loaded — prevents re-loading while typing,
  // but still reloads when a preview tab is recycled to another request.
  let activeBuilderTabId: string | null = $state(null);
  let activeBuilderRequestId: string | null = $state(null);

  // UI state
  let requestPaneTab = $state("Body");
  let response: TabResponse | null = $state(null);
  let requestError: string | null = $state(null);
  let loading = $state(false);
  let responseTab = $state("body");
  let showSaveDialog = $state(false);
  let responseHeight = $state(300);
  let isResizing = $state(false);
  let builderElement: HTMLElement | undefined = $state();
  let environmentEntries: { key: string; value: string }[] = $state([]);

  const { config: globalConfig } = configurationStore;

  onMount(() => {
    const handleSaveNew = () => {
      showSaveDialog = true;
    };
    window.addEventListener("yapla:save-request-new", handleSaveNew);
    return () => {
      window.removeEventListener("yapla:save-request-new", handleSaveNew);
    };
  });

  function loadTabIntoForm(tab: NonNullable<typeof $activeTabState>) {
    activeBuilderTabId = tab.id;
    activeBuilderRequestId = tab.requestId;
    method = tab.verb || "GET";
    url = tab.url || "";
    requestBody = tab.body || "";
    requestBodyFormat = (tab.bodyFormat as InputFormat) || "none";
    headers = tab.headers ? [...tab.headers] : [];
    requestSettings = { ...(tab.settings || {}) };
    requestName = tab.label || "";
    preRequestScript = tab.preRequestScript || "";
    postResponseScript = tab.postResponseScript || "";
    response = tab.response ?? null;
    requestError = tab.requestError ?? null;
  }

  function resetForm() {
    activeBuilderTabId = null;
    activeBuilderRequestId = null;
    method = "GET";
    url = "";
    requestBody = "";
    requestBodyFormat = "json";
    headers = [];
    requestSettings = {};
    requestName = "";
    preRequestScript = "";
    postResponseScript = "";
    response = null;
    requestError = null;
  }

  async function handleSave() {
    if (!activeBuilderTabId) return;
    if (!$activeTabState?.requestId) {
      showSaveDialog = true;
      return;
    }
    await tabStore.saveTab(activeBuilderTabId);
  }

  // --- Field change handlers (called from template on user interaction) ---
  function onFieldChange() {
    // Sync in-memory tab state
    if (!activeBuilderTabId) return;
    tabStore.updateTabFormState(activeBuilderTabId, {
      verb: method,
      url,
      body: requestBody,
      bodyFormat: requestBodyFormat,
      headers,
      settings: requestSettings,
      preRequestScript,
      postResponseScript
    });
  }

  function handleMethodChange(value: string) {
    method = value;
    onFieldChange();
  }

  function handleBodyFormatChange(value: string) {
    requestBodyFormat = value as InputFormat;
    // Also update Content-Type header
    const ct = headers.find((h) => h.key.toLowerCase() === "content-type");
    if (ct) {
      ct.value =
        value === "json"
          ? "application/json"
          : value === "xml"
            ? "application/xml"
            : value === "text"
              ? "text/plain"
              : "";
      headers = [...headers];
    }
    onFieldChange();
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
    responseHeight = Math.max(40, Math.min(rect.bottom - e.clientY, rect.height - 150));
  }
  function stopResize() {
    isResizing = false;
    window.removeEventListener("mousemove", handleResize);
    window.removeEventListener("mouseup", stopResize);
    document.body.style.userSelect = "";
  }

  // --- Methods / Body format options ---
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
    { value: "xml", label: "XML" },
    { value: "text", label: "Text" }
  ];

  // --- Beautify ---
  function formatBody() {
    if (!requestBody?.trim()) return;
    if (requestBodyFormat === "json") {
      try {
        requestBody = JSON.stringify(JSON.parse(requestBody), null, 2);
      } catch {
        // do nothing
      }
    } else if (requestBodyFormat === "xml") {
      requestBody = prettifyXml(requestBody);
    }
    onFieldChange();
  }

  function prettifyXml(xmlStr: string): string {
    const INDENT = "  ";
    let formatted = "",
      depth = 0;
    const parts = xmlStr.replace(/>\s*</g, "><").split(/(<[^>]+>)/);
    for (const part of parts) {
      if (!part.trim()) continue;
      if (part.startsWith("</")) {
        depth = Math.max(0, depth - 1);
        formatted += INDENT.repeat(depth) + part + "\n";
      } else if (part.startsWith("<?") || part.startsWith("<!")) {
        formatted += INDENT.repeat(depth) + part + "\n";
      } else if (part.startsWith("<") && part.endsWith("/>")) {
        formatted += INDENT.repeat(depth) + part + "\n";
      } else if (part.startsWith("<")) {
        formatted += INDENT.repeat(depth) + part + "\n";
        depth++;
      } else {
        const lines = formatted.trimEnd().split("\n");
        const last = lines[lines.length - 1];
        if (last?.trimEnd().endsWith(">")) {
          lines[lines.length - 1] = last + part;
          formatted = lines.join("\n") + "\n";
        } else {
          formatted += INDENT.repeat(depth) + part + "\n";
        }
      }
    }
    return formatted.trim();
  }

  // --- Environment token resolution ---
  function resolveEnvironmentTokens(
    value: string,
    sessionVars: Record<string, string> = {}
  ): string {
    if (!value) return value;

    const envMap = new Map(environmentEntries.map((e) => [e.key, e.value]));
    const sessionMap = new Map(Object.entries(sessionVars || {}));

    return value.replace(/\{\{([^{}\r\n]+?)\}\}/g, (_full, key: string) => {
      const k = key.trim();
      if (sessionMap.has(k)) return String(sessionMap.get(k) ?? "");
      if (envMap.has(k)) return String(envMap.get(k) ?? "");
      return _full;
    });
  }

  // --- Send request ---
  async function sendRequest() {
    if (!$activeTabState?.requestId) {
      showSaveDialog = true;
      return;
    }
    loading = true;

    // Keep token resolution aligned with backend env.get precedence:
    // session vars first, then selected environment.
    const sessionVars = await GetSessionVars().catch(() => ({}) as Record<string, string>);

    const resolvedUrl = resolveEnvironmentTokens(url, sessionVars);
    const resolvedBody = resolveEnvironmentTokens(requestBody, sessionVars);
    const resolvedHeaders = headers
      .filter((h) => h.enabled)
      .reduce(
        (acc, { key, value }) => ({
          ...acc,
          [resolveEnvironmentTokens(key, sessionVars)]: resolveEnvironmentTokens(value, sessionVars)
        }),
        {} as Record<string, string>
      );

    const requestOptions = new main.RequestOptions({
      body: resolvedBody,
      headers: resolvedHeaders,
      method,
      url: resolvedUrl,
      settings: requestSettings,
      preRequestScript: preRequestScript || "",
      postResponseScript: postResponseScript || ""
    });

    try {
      const responseData = await Execute(requestOptions);
      requestError = null;
      sessionVarsStore.refresh();
      const rawBody = responseData.body ?? "";
      const fmt = detectResponseFormat(responseData.headers ?? {});
      response = {
        status: responseData.statusCode,
        statusText: "TBD",
        time: responseData.duration,
        headers: responseData.headers,
        body: prettyPrint(rawBody, fmt)
      };
      if (activeBuilderTabId) {
        tabStore.updateTabResponse(activeBuilderTabId, response, null);
      }
      historyStore.push({
        collectionName: $activeTabState?.collectionName ?? null,
        requestName: requestName || null,
        request: { method, url: resolvedUrl, headers: resolvedHeaders, body: resolvedBody },
        response: {
          status: responseData.statusCode,
          time: responseData.duration,
          headers: responseData.headers,
          body: rawBody
        },
        error: null
      });
    } catch (error) {
      response = null;
      requestError = String(error);
      if (activeBuilderTabId) {
        tabStore.updateTabResponse(activeBuilderTabId, null, requestError);
      }
      historyStore.push({
        collectionName: $activeTabState?.collectionName ?? null,
        requestName: requestName || null,
        request: { method, url: resolvedUrl, headers: resolvedHeaders, body: resolvedBody },
        response: null,
        error: requestError
      });
    } finally {
      loading = false;
    }
  }

  function prettyPrint(body: string, fmt: "json" | "xml" | "text"): string {
    if (!body?.trim()) return body;
    if (fmt === "json") {
      try {
        return JSON.stringify(JSON.parse(body), null, 2);
      } catch {
        // do nothing
      }
    }
    if (fmt === "xml") return prettifyXml(body);
    return body;
  }

  function detectResponseFormat(hdrs: Record<string, string>): "json" | "xml" | "text" {
    const ct =
      Object.entries(hdrs ?? {}).find(([k]) => k.toLowerCase() === "content-type")?.[1] ?? "";
    if (ct.includes("json")) return "json";
    if (ct.includes("xml") || ct.includes("html")) return "xml";
    return "text";
  }

  function getResponseFormat(currentResponse: TabResponse | null): "json" | "xml" | "text" {
    if (!currentResponse) return "text";
    return detectResponseFormat(currentResponse.headers);
  }

  async function handleSaveRequest(data: { name: string; collection: string | null }) {
    const { name, collection: targetCollection } = data;
    if (!targetCollection || !activeBuilderTabId) return;
    try {
      const headersObj = headers
        .filter((h) => h.enabled && h.key)
        .reduce((acc, { key, value }) => ({ ...acc, [key]: value }), {} as Record<string, string>);
      await collectionStore.addRequest(targetCollection, {
        name: name || "Untitled Request",
        url,
        verb: method,
        body: requestBody,
        headers: headersObj,
        settings: requestSettings
      });
      showSaveDialog = false;
      await sendRequest();
    } catch {
      /* shown by store */
    }
  }

  function getStatusClass(status: number): string {
    if (status >= 200 && status < 300) return "status-success";
    if (status >= 300 && status < 400) return "status-info";
    if (status >= 400 && status < 500) return "status-warning";
    return "status-error";
  }
  run(() => {
    environmentEntries = Object.entries($selectedEnvironment?.values ?? {}).map(([key, val]) => ({
      key,
      value: String(val?.value ?? "")
    }));
  });
  // --- Tab switching: ONE-WAY, store → form only ---
  // Reload when active tab changes OR when the same preview tab id is recycled
  // to point at a different request.
  run(() => {
    const nextTabId = $activeTabState?.id ?? null;
    const nextRequestId = $activeTabState?.requestId ?? null;
    if (nextTabId !== activeBuilderTabId || nextRequestId !== activeBuilderRequestId) {
      if ($activeTabState) {
        loadTabIntoForm($activeTabState);
      } else {
        resetForm();
      }
    }
  });
  let responseFormat = $derived(getResponseFormat(response));
</script>

{#if $activeTabState}
  <div class="request-builder" bind:this={builderElement}>
    <TokenTooltip />

    <!-- Request Header -->
    <div class="request-header">
      <div class="request-name-container">
        <span class="request-name">{requestName || "New Request"}</span>
        {#if $activeTabState.isDirty}
          <button
            class="save-btn"
            onclick={handleSave}
            title="Save Request (Ctrl+S)"
            aria-label="Save Request"
          >
            <svg
              width="14"
              height="14"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"></path>
              <polyline points="17 21 17 13 7 13 7 21"></polyline>
              <polyline points="7 3 7 8 15 8"></polyline>
            </svg>
            <span>Save</span>
          </button>
        {/if}
      </div>
    </div>

    <!-- Request Line -->
    <div class="request-line">
      <div class="url-bar">
        <div class="url-bar-method">
          <Dropdown
            bind:value={method}
            options={methodOptions}
            change={handleMethodChange}
            variant="url-method"
            square
          />
        </div>
        <div class="url-bar-divider"></div>
        <TokenInput
          bind:value={url}
          placeholder="Enter request URL"
          {environmentEntries}
          wrapperClass="url-bar-input"
          onChange={onFieldChange}
        />
        <div class="url-bar-divider"></div>
        <Button
          variant="primary"
          click={sendRequest}
          disabled={loading}
          square
          style="padding: 0 var(--space-xl); font-weight: var(--font-weight-semibold); align-self: stretch; border-radius: 0 var(--radius-md) var(--radius-md) 0;"
          >{loading ? "Sending…" : "Send"}</Button
        >
      </div>
    </div>

    <div class="request-content-bar">
      <Tabs bind:activeValue={requestPaneTab} variant="minimal">
        <Tab title="Headers" value="Headers" />
        <Tab title="Body" value="Body" />
        <Tab
          title="Scripts"
          value="Scripts"
          badge={preRequestScript.trim() || postResponseScript.trim() ? "●" : undefined}
        />
        <Tab title="Settings" value="Settings" />
        <Tab title="Runner" value="Runner" />
      </Tabs>

      {#if requestPaneTab === "Body"}
        <div class="body-format-selector">
          {#if requestBodyFormat !== "none"}
            <button
              class="beautify-btn"
              title="Prettify / Format body"
              onclick={formatBody}
              disabled={requestBodyFormat === "text"}
            >
              <svg
                width="14"
                height="14"
                viewBox="0 0 16 16"
                fill="none"
                xmlns="http://www.w3.org/2000/svg"
              >
                <path
                  d="M2 4h12M2 8h8M2 12h10"
                  stroke="currentColor"
                  stroke-width="1.5"
                  stroke-linecap="round"
                />
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
      {#if requestPaneTab === "Headers"}
        {#key $activeTabState.id}
          <RequestHeaders {headers} onChange={onFieldChange} />
        {/key}
      {:else if requestPaneTab === "Body"}
        {#key $activeTabState.id}
          {#if requestBodyFormat === "none"}
            <EmptyState message="This request does not have a body" />
          {:else}
            <RequestBody
              bind:requestBody
              bind:format={requestBodyFormat}
              onChange={onFieldChange}
            />
          {/if}
        {/key}
      {:else if requestPaneTab === "Scripts"}
        {#key $activeTabState.id}
          <RequestScripts
            bind:preRequestScript
            bind:postResponseScript
            onPreChange={(val) => {
              preRequestScript = val;
              onFieldChange();
            }}
            onPostChange={(val) => {
              postResponseScript = val;
              onFieldChange();
            }}
          />
        {/key}
      {:else if requestPaneTab === "Settings"}
        {#key $activeTabState.id}
          <RequestSettings
            bind:requestSettings
            globalConfig={$globalConfig}
            onChange={onFieldChange}
          />
        {/key}
      {:else if requestPaneTab === "Runner"}
        {#key $activeTabState.id}
          <RequestRunner
            {method}
            {url}
            body={requestBody}
            {headers}
            settings={requestSettings}
            {preRequestScript}
            {postResponseScript}
          />
        {/key}
      {/if}
    </div>

    <!-- Response Section -->
    <div class="response-section" style="height: {responseHeight}px">
      <div class="resize-handle" class:resizing={isResizing} onmousedown={startResize}></div>

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
              {#each Object.entries(response.headers) as [key, value] (key)}
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
  onSave={handleSaveRequest}
  onCancel={() => (showSaveDialog = false)}
/>

<style>
  .request-builder {
    display: flex;
    flex-direction: column;
    height: 100%;
    width: 100%;
    background: var(--bg-primary);
  }

  .request-header {
    padding: var(--space-sm) var(--space-lg) 0 var(--space-lg);
    background: var(--bg-primary);
  }

  .request-name-container {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
    height: 24px;
  }

  .request-name {
    font-size: var(--font-size-xs);
    font-weight: var(--font-weight-semibold);
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .save-btn {
    display: flex;
    align-items: center;
    gap: 4px;
    background: var(--bg-tertiary);
    border: 1px solid var(--border);
    color: var(--primary);
    font-size: 11px;
    font-family: var(--font-sans);
    font-weight: var(--font-weight-bold);
    padding: 1px 6px;
    border-radius: var(--radius-sm);
    cursor: pointer;
    transition: all var(--transition-fast);
  }

  .save-btn:hover {
    background: var(--primary);
    color: var(--bg-primary);
    border-color: var(--primary);
  }

  .request-line {
    padding: var(--space-xs) var(--space-lg) var(--space-md) var(--space-lg);
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
    transition:
      color var(--transition-fast),
      background var(--transition-fast);
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
