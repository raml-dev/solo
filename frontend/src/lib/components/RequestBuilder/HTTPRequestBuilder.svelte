<script lang="ts">
  import Alert from "flowbite-svelte/Alert.svelte";
  import Badge from "flowbite-svelte/Badge.svelte";
  import Button from "flowbite-svelte/Button.svelte";
  import Select from "flowbite-svelte/Select.svelte";
  import FeedbackEmptyState from "$src/lib/components/common/FeedbackEmptyState.svelte";
  import TabItem from "flowbite-svelte/TabItem.svelte";
  import Tabs from "flowbite-svelte/Tabs.svelte";
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
  let responseCollapsed = $state(false);
  let isResizing = $state(false);
  let builderElement: HTMLElement | undefined = $state();
  let environmentEntries: { key: string; value: string }[] = $derived(
    Object.entries($selectedEnvironment?.values ?? {}).map(([key, val]) => ({
      key,
      value: String(val?.value ?? "")
    }))
  );

  const { config: globalConfig } = configurationStore;

  onMount(() => {
    if (builderElement) {
      responseHeight = Math.floor(builderElement.clientHeight * 0.35);
    }

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
    { value: "GET", name: "GET" },
    { value: "POST", name: "POST" },
    { value: "PUT", name: "PUT" },
    { value: "DELETE", name: "DELETE" },
    { value: "PATCH", name: "PATCH" },
    { value: "HEAD", name: "HEAD" },
    { value: "OPTIONS", name: "OPTIONS" }
  ];
  const bodyFormatOptions = [
    { value: "none", name: "None" },
    { value: "json", name: "JSON" },
    { value: "xml", name: "XML" },
    { value: "text", name: "Text" }
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

  function getStatusBadgeColor(status: number): "green" | "blue" | "yellow" | "red" {
    if (status >= 200 && status < 300) return "green";
    if (status >= 300 && status < 400) return "blue";
    if (status >= 400 && status < 500) return "yellow";
    return "red";
  }
  // --- Tab switching: ONE-WAY, store → form only ---
  // Reload when active tab changes OR when the same preview tab id is recycled
  // to point at a different request.
  $effect(() => {
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
  <div class="flex h-full flex-col overflow-hidden" bind:this={builderElement}>
    <TokenTooltip />

    <!-- Request Header -->
    <div class="flex shrink-0 items-center justify-between border-b border-neutral-200 px-3 py-2 dark:border-neutral-700">
      <div class="flex items-center gap-2">
        <span class="text-sm font-semibold text-neutral-800 dark:text-neutral-100">{requestName || "New Request"}</span>
        {#if $activeTabState.isDirty}
          <Button
            color="light"
            size="xs"
            onclick={handleSave}
            title="Save Request (Ctrl+S)"
            aria-label="Save Request"
          >
            Save
          </Button>
        {/if}
      </div>
    </div>

    <!-- Request Line -->
    <div class="shrink-0 border-b border-neutral-200 dark:border-neutral-700">
      <div class="flex items-stretch">
        <div class="shrink-0">
          <Select
            bind:value={method}
            items={methodOptions}
            placeholder=""
            size="sm"
            onchange={() => handleMethodChange(method)}
            class="h-full rounded-none border-0 bg-transparent px-3 py-0 font-semibold"
          />
        </div>
        <div class="w-px shrink-0 self-stretch bg-neutral-200 dark:bg-neutral-700"></div>
        <TokenInput
          bind:value={url}
          placeholder="Enter request URL"
          {environmentEntries}
          wrapperClass="min-w-0 flex-1"
          onChange={onFieldChange}
        />
        <div class="w-px shrink-0 self-stretch bg-neutral-200 dark:bg-neutral-700"></div>
        <Button
          color="primary"
          {loading}
          onclick={sendRequest}
          disabled={loading}
          class="self-stretch rounded-l-none px-6"
        >Send</Button>
      </div>
    </div>

    <!-- Request tabs + body format selector -->
    <div class="flex shrink-0 items-center border-b border-neutral-200 dark:border-neutral-700">
      <Tabs bind:selected={requestPaneTab} tabStyle="underline" contentClass="hidden">
        <TabItem key="Headers" title="Headers" />
        <TabItem key="Body" title="Body" />
        <TabItem key="Scripts">
          {#snippet titleSlot()}
            <span class="inline-flex items-center gap-1">
              <span>Scripts</span>
              {#if preRequestScript.trim() || postResponseScript.trim()}
                <span aria-hidden="true">●</span>
              {/if}
            </span>
          {/snippet}
        </TabItem>
        <TabItem key="Settings" title="Settings" />
        <TabItem key="Runner" title="Runner" />
      </Tabs>

      {#if requestPaneTab === "Body"}
        <div class="ml-auto flex items-center gap-1 px-2">
          {#if requestBodyFormat !== "none"}
            <Button
              color="light"
              size="xs"
              title="Prettify / Format body"
              onclick={formatBody}
              disabled={requestBodyFormat === "text"}
            >
              Beautify
            </Button>
            <span class="text-neutral-300 dark:text-neutral-600">|</span>
          {/if}
          <Select
            bind:value={requestBodyFormat}
            items={bodyFormatOptions}
            placeholder=""
            size="sm"
            underline
            onchange={() => handleBodyFormatChange(requestBodyFormat)}
          />
        </div>
      {/if}
    </div>

    <!-- Request tab content -->
    <div class="min-h-0 flex-1 flex flex-col overflow-hidden">
      {#if requestPaneTab === "Headers"}
        {#key $activeTabState.id}
          <RequestHeaders {headers} onChange={onFieldChange} />
        {/key}
      {:else if requestPaneTab === "Body"}
        {#key $activeTabState.id}
          {#if requestBodyFormat === "none"}
            <FeedbackEmptyState variant="info" title="This request does not have a body" compact />
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
    <div
      class="flex shrink-0 flex-col overflow-hidden border-t border-neutral-200 dark:border-neutral-700"
      style={responseCollapsed ? undefined : `height: ${responseHeight}px`}
    >
      {#if !responseCollapsed}
        <Button
          color="light"
          class="h-1 w-full shrink-0 cursor-row-resize rounded-none border-0 bg-neutral-200 p-0 shadow-none transition-colors hover:bg-primary-400 dark:bg-neutral-700 {isResizing ? 'bg-primary-500' : ''}"
          onmousedown={startResize}
          aria-label="Resize response panel"
        />
      {/if}

      <div
        class="flex shrink-0 cursor-pointer items-center border-b border-neutral-200 dark:border-neutral-700"
        role="button"
        tabindex="0"
        onclick={() => (responseCollapsed = !responseCollapsed)}
        onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); responseCollapsed = !responseCollapsed; } }}
        aria-expanded={!responseCollapsed}
        aria-label={responseCollapsed ? "Expand response" : "Collapse response"}
      >
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <div onclick={(e) => e.stopPropagation()}>
          <Tabs bind:selected={responseTab} tabStyle="underline" contentClass="hidden">
            <TabItem key="body" title="Body" />
            <TabItem key="headers" title="Headers" />
          </Tabs>
        </div>
        <div class="ml-auto flex items-center gap-2 px-2">
          {#if response}
            <!-- svelte-ignore a11y_click_events_have_key_events -->
            <!-- svelte-ignore a11y_no_static_element_interactions -->
            <div onclick={(e) => e.stopPropagation()} class="flex items-center gap-2">
              <Badge color={getStatusBadgeColor(response.status)}>{response.status}</Badge>
              <Badge color="gray">{response.time}ms</Badge>
            </div>
          {/if}
          <Button
            color="light"
            size="xs"
            class="border-0 shadow-none"
            onclick={(e: MouseEvent) => { e.stopPropagation(); responseCollapsed = !responseCollapsed; }}
            aria-label={responseCollapsed ? "Expand response" : "Collapse response"}
          >
            <svg
              width="12"
              height="12"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2.5"
              class="transition-transform {responseCollapsed ? 'rotate-180' : ''}"
            >
              <polyline points="6 9 12 15 18 9" />
            </svg>
          </Button>
        </div>
      </div>

      {#if !responseCollapsed}
        {#if requestError}
          <Alert color="red" class="m-3">
            <div class="flex flex-col gap-1">
              <span class="font-medium">Request failed</span>
              <span class="text-sm">{requestError}</span>
            </div>
          </Alert>
        {:else if response}
          <div class="min-h-0 flex-1 overflow-hidden">
            {#if responseTab === "body"}
              <div class="h-full">
                <CodeMirrorEditor
                  value={response.body ?? ""}
                  format={responseFormat}
                  readOnly={true}
                />
              </div>
            {:else}
              <div class="h-full overflow-y-auto">
                {#each Object.entries(response.headers) as [key, value] (key)}
                  <div class="flex items-start gap-2 border-b border-neutral-100 px-3 py-1.5 text-sm dark:border-neutral-800">
                    <span class="shrink-0 font-medium text-neutral-700 dark:text-neutral-300">{key}:</span>
                    <span class="min-w-0 break-all text-neutral-500 dark:text-neutral-400">{value}</span>
                  </div>
                {/each}
              </div>
            {/if}
          </div>
        {:else}
          <div class="flex h-full items-center justify-center">
            <FeedbackEmptyState title="Send a request to see the response" compact />
          </div>
        {/if}
      {/if}
    </div>
  </div>
{:else}
  <div class="flex h-full w-full items-center justify-center">
    <FeedbackEmptyState title="Open a request from the sidebar or press + to start a new one" />
  </div>
{/if}

<SaveRequestModal
  bind:show={showSaveDialog}
  bind:requestName
  onSave={handleSaveRequest}
  onCancel={() => (showSaveDialog = false)}
/>
