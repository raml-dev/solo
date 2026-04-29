<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import FeedbackEmptyState from "$src/lib/components/common/FeedbackEmptyState.svelte";
  import CodeMirrorEditor from "$src/lib/components/RequestBuilder/CodeMirrorEditor.svelte";
  import EnvTokenInput from "$src/lib/components/RequestBuilder/EnvTokenInput.svelte";
  import RequestAuth from "$src/lib/components/RequestBuilder/RequestAuth.svelte";
  import RequestBody from "$src/lib/components/RequestBuilder/RequestBody.svelte";
  import RequestHeaders from "$src/lib/components/RequestBuilder/RequestHeaders.svelte";
  import RequestRunner from "$src/lib/components/RequestBuilder/RequestRunner.svelte";
  import RequestScripts from "$src/lib/components/RequestBuilder/RequestScripts.svelte";
  import RequestSettings from "$src/lib/components/RequestBuilder/RequestSettings.svelte";
  import TokenTooltip from "$src/lib/components/RequestBuilder/TokenTooltip.svelte";
  import type { InputFormat } from "$src/lib/components/RequestBuilder/types";
  import SaveRequestModal from "$src/lib/components/SaveRequestModal.svelte";
  import { collectionStore, collectionStoreState } from "$src/lib/stores/collectionStore.svelte";
  import { configurationStoreState } from "$src/lib/stores/configurationStore.svelte";
  import { environmentStoreState } from "$src/lib/stores/environmentStore.svelte";
  import { modalStack } from "$src/lib/stores/modalStackStore.svelte";
  import { notifications } from "$src/lib/stores/notificationStore";
  import { sessionVarsStore } from "$src/lib/stores/sessionVarsStore";
  import {
    getActiveTab,
    tabStore,
    tabStoreState,
    type TabResponse
  } from "$src/lib/stores/tabStore.svelte";
  import {
    buildResolvedRequestPayload,
    getHttpStatusString,
    getStatusBadgeColor
  } from "$src/lib/utils/http";
  import {
    resolveVariableEntries,
    resolveVariableTokens,
    type ResolvedVariableEntry
  } from "$src/lib/utils/variableResolution";
  import { Execute, GenerateCurl, GetSessionVars, SaveCurlFile } from "$wails/go/main/App";
  import { collection, main } from "$wails/go/models";
  import TerminalSolid from "flowbite-svelte-icons/TerminalSolid.svelte";
  import FloppyDiskSolid from "flowbite-svelte-icons/FloppyDiskSolid.svelte";
  import Alert from "flowbite-svelte/Alert.svelte";
  import Badge from "flowbite-svelte/Badge.svelte";
  import Button from "flowbite-svelte/Button.svelte";
  import ButtonGroup from "flowbite-svelte/ButtonGroup.svelte";
  import Label from "flowbite-svelte/Label.svelte";
  import Modal from "flowbite-svelte/Modal.svelte";
  import Select from "flowbite-svelte/Select.svelte";
  import Spinner from "flowbite-svelte/Spinner.svelte";
  import TabItem from "flowbite-svelte/TabItem.svelte";
  import Tabs from "flowbite-svelte/Tabs.svelte";
  import { onDestroy, onMount, tick } from "svelte";

  // UI-only local state (not tab data)
  let requestPaneTab = $state("Body");
  let response = $derived(getActiveTab()?.response ?? null);
  let requestError = $derived(getActiveTab()?.requestError ?? null);
  let loading = $state(false);
  let responseTab = $state("body");
  let responseHeaderView = $state("received");
  let showSaveDialog = modalStack.createModal("save-request");
  let showCurlModal = $state(false);
  let curlPreview = $state("");
  let responseHeight = $state(300);
  let responseCollapsed = $state(false);
  let isResizing = $state(false);
  let builderElement: HTMLElement | undefined = $state();
  let selectedEnvironment = $derived(
    environmentStoreState.environments.find(
      (e) => e.name === environmentStoreState.selectedEnvironmentName
    ) || null
  );
  let activeCollection = $derived(
    collectionStoreState.collections.find(
      (collection) => collection.name === getActiveTab()?.collectionName
    ) || null
  );
  let resolvedVariableEntries = $derived(
    resolveVariableEntries({
      sessionValues: $sessionVarsStore,
      environmentValues: selectedEnvironment?.values,
      collectionValues: activeCollection?.variables
    })
  );

  const globalConfig = $derived(configurationStoreState.config);

  let requestName = $derived(tabStoreState.tabs[tabStoreState.activeTabIndex]?.label || "");

  // --- Inline rename state ---
  let isEditingName = $state(false);
  let editingName = $state("");
  let nameInputEl = $state<HTMLInputElement | null>(null);

  function startEditingName() {
    editingName = requestName || "New Request";
    isEditingName = true;
    tick().then(() => {
      nameInputEl?.focus();
      nameInputEl?.select();
    });
  }

  function confirmEditingName() {
    if (!isEditingName) return;
    isEditingName = false;
    const tab = getActiveTab();
    if (!tab) return;
    const trimmed = editingName.trim();
    const newName = trimmed || "New Request";
    if (newName !== tab.label) {
      tab.label = newName;
      onFieldChange();
    }
  }

  function cancelEditingName() {
    isEditingName = false;
  }

  onMount(() => {
    if (builderElement) {
      responseHeight = Math.floor(builderElement.clientHeight * 0.35);
    }

    const handleSaveNew = () => {
      showSaveDialog.open = true;
    };
    window.addEventListener("solo:save-request-new", handleSaveNew);
    return () => {
      window.removeEventListener("solo:save-request-new", handleSaveNew);
    };
  });

  onDestroy(() => {
    modalStack.destroyModal("save-request");
  });

  $effect(() => {
    const tab = getActiveTab();
    if (!tab) return;

    const expectedContentType = getExpectedContentType(tab.body, tab.bodyFormat);
    syncInjectedContentTypeHeader(tab, expectedContentType);
  });

  // Field change handler - mutation already happened via bind:, just update metadata
  function onFieldChange() {
    const tab = getActiveTab();
    if (!tab) return;
    tab.isDirty = true;
    tab.isPreview = false;
    tabStore.storeTabsInLocalStorage();
  }

  async function handleSave() {
    const tab = getActiveTab();
    if (!tab?.id || !tab.requestId) {
      showSaveDialog.open = true;
      return;
    }
    await tabStore.saveTab(tab.id);
  }

  function handleMethodChange(value: string) {
    const tab = getActiveTab();
    if (tab) {
      tab.verb = value;
    }
    onFieldChange();
  }

  function handleBodyFormatChange(value: string) {
    const tab = getActiveTab();
    if (!tab) return;
    tab.bodyFormat = value as InputFormat;
    onFieldChange();
  }

  function getExpectedContentType(body: string, bodyFormat: InputFormat): string | null {
    if (bodyFormat === "none") return null;
    if (!body.trim()) return null;
    if (bodyFormat === "json") return "application/json";
    if (bodyFormat === "xml") return "application/xml";
    return null;
  }

  function syncInjectedContentTypeHeader(
    tab: ReturnType<typeof getActiveTab>,
    expectedContentType: string | null
  ) {
    if (!tab) return;
    const contentTypeHeader = tab.headers.find(
      (header) => header.key.trim().toLowerCase() === "content-type"
    );

    if (!contentTypeHeader) {
      if (!expectedContentType) return;
      tab.headers = [
        ...tab.headers,
        {
          id: crypto.randomUUID(),
          key: "Content-Type",
          value: expectedContentType,
          enabled: true,
          autoInjectedContentType: true,
          injectedContentTypeValue: expectedContentType
        }
      ];
      return;
    }

    if (!contentTypeHeader.autoInjectedContentType || !expectedContentType) return;

    const injectedValue = contentTypeHeader.injectedContentTypeValue ?? "";
    if (contentTypeHeader.value === injectedValue && injectedValue !== expectedContentType) {
      contentTypeHeader.value = expectedContentType;
      contentTypeHeader.injectedContentTypeValue = expectedContentType;
      tab.headers = [...tab.headers];
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
    const tab = getActiveTab();
    if (!tab || !tab.body?.trim()) return;
    if (tab.bodyFormat === "json") {
      try {
        tab.body = JSON.stringify(JSON.parse(tab.body), null, 2);
      } catch {
        // do nothing
      }
    } else if (tab.bodyFormat === "xml") {
      tab.body = prettifyXml(tab.body);
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

  function getResolvedVariableEntries(
    sessionValues: Record<string, string> = $sessionVarsStore
  ): ResolvedVariableEntry[] {
    return resolveVariableEntries({
      sessionValues,
      environmentValues: selectedEnvironment?.values,
      collectionValues: activeCollection?.variables
    });
  }

  function resolveRequestTokens(
    value: string,
    sessionValues: Record<string, string> = $sessionVarsStore
  ): string {
    return resolveVariableTokens(value, getResolvedVariableEntries(sessionValues));
  }

  function showUnresolvedVariableWarnings(values: string[]) {
    const unresolved = values.flatMap((value) =>
      [...value.matchAll(/\{\{([^{}\r\n]+?)\}\}/g)].map((match) => match[1].trim())
    );

    for (const key of [...new Set(unresolved)]) {
      notifications.warning(
        `Placeholder "{{${key}}}" not resolved — no value in Session, Environment, or Collection`
      );
    }
  }

  // --- Send request ---
  async function sendRequest() {
    loading = true;

    const tab = getActiveTab();
    if (!tab) {
      loading = false;
      return;
    }

    const requestHeaders = tab.headers
      .filter((h) => h.enabled)
      .reduce((acc, { key, value }) => ({ ...acc, [key]: value }), {} as Record<string, string>);

    const authConfig = collection.AuthConfiguration.createFrom({
      ...tab.auth,
      template: { ...(tab.auth.template || {}) }
    });

    const requestOptions = new main.RequestOptions({
      body: tab.body,
      headers: requestHeaders,
      method: tab.verb,
      url: tab.url,
      collectionName: tab.collectionName || "",
      auth: authConfig,
      settings: tab.settings,
      preRequestScript: tab.preRequestScript || "",
      postResponseScript: tab.postResponseScript || ""
    });

    try {
      const responseData = await Execute(requestOptions);
      sessionVarsStore.refresh();
      const rawBody = responseData.body ?? "";
      const fmt = detectResponseFormat(responseData.headers ?? {});
      const newResponse: TabResponse = {
        status: responseData.statusCode,
        statusText: getHttpStatusString(responseData.statusCode),
        time: responseData.duration,
        headers: responseData.headers,
        requestHeaders: responseData.requestHeaders,
        body: prettyPrint(rawBody, fmt)
      };
      tabStore.updateTabResponse(tab.id, newResponse, null);
    } catch (error) {
      const errorMsg = String(error);
      tabStore.updateTabResponse(tab.id, null, errorMsg);
    } finally {
      loading = false;
    }
  }

  async function handleExportCurl() {
    const tab = getActiveTab();
    if (!tab) return;

    const sessionVars = await GetSessionVars().catch(() => ({}) as Record<string, string>);

    const resolvedUrl = resolveRequestTokens(tab.url, sessionVars);
    const { body: resolvedBody, headers: resolvedHeaders } = buildResolvedRequestPayload({
      body: tab.body,
      headers: tab.headers,
      resolveTokens: (value) => resolveRequestTokens(value, sessionVars)
    });

    // Add cookies from the saved request as Cookie header if present
    if (tab.collectionName && tab.requestId) {
      const colls = collectionStoreState.collections;
      const savedReq = colls
        .find((c) => c.name === tab.collectionName)
        ?.requests.find((r) => r.id === tab.requestId);
      const cookieEntries = Object.entries(savedReq?.cookies ?? {});
      if (cookieEntries.length > 0) {
        resolvedHeaders["Cookie"] = cookieEntries.map(([k, v]) => `${k}=${v}`).join("; ");
      }
    }

    // Warn about unresolved placeholders
    showUnresolvedVariableWarnings([resolvedUrl, resolvedBody, ...Object.values(resolvedHeaders)]);

    try {
      const curl = await GenerateCurl({
        method: tab.verb,
        url: resolvedUrl,
        headers: resolvedHeaders,
        body: resolvedBody
      });
      curlPreview = curl;
      showCurlModal = true;
    } catch (err) {
      notifications.error("Failed to generate cURL", String(err));
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
    const tab = getActiveTab();
    if (!tab || !data.collection) return;
    try {
      const headersObj = tab.headers
        .filter((h) => h.enabled && h.key)
        .reduce((acc, { key, value }) => ({ ...acc, [key]: value }), {} as Record<string, string>);
      const newReq = await collectionStore.addRequest(data.collection, {
        name: data.name,
        url: tab.url,
        verb: tab.verb,
        body: tab.body,
        headers: headersObj,
        auth: tab.auth,
        settings: tab.settings
      });
      tabStore.bindTabToRequest(tab.id, newReq.id, data.collection, data.name);
      showSaveDialog.open = false;
    } catch {
      /* shown by store */
    }
  }

  let responseFormat = $derived(getResponseFormat(response));
</script>

<svelte:window
  onkeydown={(e) => {
    if (e.ctrlKey && e.key === "s") {
      e.preventDefault();
      handleSave();
    }
  }}
/>

{#if tabStoreState.tabs[tabStoreState.activeTabIndex]}
  {@const tab = tabStoreState.tabs[tabStoreState.activeTabIndex]}
  <div class="flex h-full flex-col overflow-hidden" bind:this={builderElement}>
    <TokenTooltip />

    <!-- Request Header -->
    <div
      class="flex shrink-0 items-center justify-between border-b border-neutral-200 px-3 py-1 dark:border-neutral-700"
    >
      <div class="flex items-center gap-2">
        {#if tab.collectionName}
          <span class="text-sm font-semibold text-neutral-800 dark:text-neutral-100"
            >{tab.collectionName} /&nbsp;</span
          >
        {/if}
        {#if isEditingName}
          <input
            bind:this={nameInputEl}
            bind:value={editingName}
            class="max-w-64 rounded-md border border-primary-300 bg-white px-2 py-0.5 text-sm font-semibold text-neutral-800 focus:ring-1 focus:ring-primary-500 focus:outline-none dark:border-primary-700 dark:bg-neutral-800 dark:text-neutral-100"
            onblur={confirmEditingName}
            onkeydown={(e: KeyboardEvent) => {
              if (e.key === "Enter") {
                e.preventDefault();
                confirmEditingName();
              } else if (e.key === "Escape") {
                e.preventDefault();
                cancelEditingName();
              }
            }}
          />
        {:else}
          <span
            class="cursor-pointer text-sm font-semibold text-neutral-800 hover:underline dark:text-neutral-100"
            role="button"
            tabindex="0"
            title="Double-click to rename"
            ondblclick={startEditingName}
            onkeydown={(e: KeyboardEvent) => {
              if (e.key === "Enter") startEditingName();
            }}
          >
            {requestName || "New Request"}
          </span>
        {/if}
      </div>
      <div>
        {#if !tab.requestId || tab.isDirty}
          <Button
            color="light"
            size="xs"
            class="h-8 shrink-0 border-none bg-transparent text-neutral-800 inset-ring-primary-500 focus-within:inset-ring-1 focus-within:outline-hidden hover:bg-neutral-200 hover:text-neutral-800 focus:ring-0 focus:outline-hidden dark:border-none dark:bg-transparent dark:text-neutral-100 dark:hover:text-neutral-100"
            onclick={handleSave}
            title="Save Request (Ctrl+S)"
            aria-label="Save Request"
          >
            <FloppyDiskSolid class="m-1 h-4 w-4 shrink-0" /><span>Save</span>
          </Button>
        {/if}
        <Button
          color="light"
          class="h-8 shrink-0 border-none bg-transparent text-neutral-800 inset-ring-primary-500 focus-within:inset-ring-1 focus-within:outline-hidden hover:bg-neutral-200 hover:text-neutral-800 focus:ring-0 focus:outline-hidden dark:border-none dark:bg-transparent dark:text-neutral-100 dark:hover:text-neutral-100"
          size="xs"
          title="Export as cURL"
          onclick={handleExportCurl}
        >
          <TerminalSolid class="m-1 h-4 w-4 shrink-0" /><span>Export curl</span>
        </Button>
      </div>
    </div>

    <!-- Request Line -->
    <div class="shrink-0 border-b border-neutral-200 p-3 dark:border-neutral-700">
      <ButtonGroup class="w-full">
        <Select
          bind:value={tab.verb}
          items={methodOptions}
          placeholder=""
          size="sm"
          onchange={() => handleMethodChange(tab.verb)}
          class="w-24 font-semibold"
        />
        <EnvTokenInput
          bind:value={tab.url}
          placeholder="Enter request URL"
          class="-ms-px min-w-0 flex-1 rounded-none"
          size="sm"
          variableEntries={resolvedVariableEntries}
          onChange={onFieldChange}
          onEnter={sendRequest}
        />
        <Button
          color="primary"
          class="w-24 px-6"
          size="sm"
          onclick={sendRequest}
          disabled={loading}
        >
          {loading ? "Sending..." : "Send"}
        </Button>
      </ButtonGroup>
    </div>

    <!-- Request tabs + body format selector -->
    <div class="flex shrink-0 items-center border-b border-neutral-200 dark:border-neutral-700">
      <Tabs bind:selected={requestPaneTab} tabStyle="underline" classes={{ content: "hidden" }}>
        <TabItem key="Headers" title="Headers" />
        <TabItem key="Body" title="Body" />
        <TabItem key="Auth">
          {#snippet titleSlot()}
            <span class="inline-flex items-center gap-1">
              <span>OAuth</span>
              {#if tab.auth.enabled}
                <span class="h-1.5 w-1.5 rounded-full bg-primary-500" aria-hidden="true"></span>
              {/if}
            </span>
          {/snippet}
        </TabItem>
        <TabItem key="Scripts">
          {#snippet titleSlot()}
            <span class="inline-flex items-center gap-1">
              <span>Scripts</span>
              {#if tab.preRequestScript.trim() || tab.postResponseScript.trim()}
                <span class="h-1.5 w-1.5 rounded-full bg-primary-500" aria-hidden="true"></span>
              {/if}
            </span>
          {/snippet}
        </TabItem>
        <TabItem key="Settings" title="Settings" />
        <TabItem key="Runner" title="Runner" />
      </Tabs>

      {#if requestPaneTab === "Body"}
        <div class="ml-auto flex items-center gap-1 px-2">
          <Label class="mr-1">Body type:</Label>
          <ButtonGroup>
            <Select
              bind:value={tab.bodyFormat}
              items={bodyFormatOptions}
              placeholder=""
              size="sm"
              class="w-20 shrink-0 border-none bg-transparent text-neutral-800 focus-within:outline-hidden  hover:bg-neutral-200 hover:text-neutral-800 dark:bg-transparent  dark:text-neutral-300 dark:hover:bg-gray-700 dark:hover:text-neutral-100"
              classes={{
                select:
                  "bg-transparent dark:bg-transparent border-none focus:outline-none dark:text-neutral-100 dark:hover:text-neutral-100"
              }}
              onchange={() => handleBodyFormatChange(tab.bodyFormat)}
            />
            {#if tab.bodyFormat !== "none"}
              <Button
                color="light"
                class="shrink-0 border-none bg-transparent text-neutral-800 inset-ring-primary-500 focus-within:inset-ring-1 focus-within:outline-hidden hover:bg-neutral-200 hover:text-neutral-800 focus:ring-0 focus:outline-hidden dark:border-none dark:bg-transparent dark:text-neutral-100 dark:hover:text-neutral-100"
                size="xs"
                title="Prettify / Format body"
                onclick={formatBody}
                disabled={tab.bodyFormat === "text"}
              >
                Beautify
              </Button>
            {/if}
          </ButtonGroup>
        </div>
      {/if}
    </div>

    <!-- Request tab content -->
    <div class="flex min-h-0 flex-1 flex-col overflow-hidden">
      {#if requestPaneTab === "Headers"}
        <RequestHeaders
          bind:headers={tab.headers}
          body={tab.body}
          bodyFormat={tab.bodyFormat}
          variableEntries={resolvedVariableEntries}
          onChange={onFieldChange}
        />
      {:else if requestPaneTab === "Body"}
        {#if tab.bodyFormat === "none"}
          <FeedbackEmptyState variant="info" title="This request does not have a body" compact />
        {:else}
          <RequestBody
            bind:requestBody={tab.body}
            bind:format={tab.bodyFormat}
            variableEntries={resolvedVariableEntries}
            onChange={onFieldChange}
          />
        {/if}
      {:else if requestPaneTab === "Auth"}
        {#key tab.id}
          <RequestAuth
            bind:auth={tab.auth}
            variableEntries={resolvedVariableEntries}
            onChange={onFieldChange}
          />
        {/key}
      {:else if requestPaneTab === "Scripts"}
        <RequestScripts
          bind:preRequestScript={tab.preRequestScript}
          bind:postResponseScript={tab.postResponseScript}
          variableEntries={resolvedVariableEntries}
          onPreChange={(val) => {
            tab.preRequestScript = val;
            onFieldChange();
          }}
          onPostChange={(val) => {
            tab.postResponseScript = val;
            onFieldChange();
          }}
        />
      {:else if requestPaneTab === "Settings"}
        <RequestSettings
          bind:requestSettings={tab.settings}
          {globalConfig}
          onChange={onFieldChange}
        />
      {:else if requestPaneTab === "Runner"}
        <RequestRunner
          method={tab.verb}
          url={tab.url}
          body={tab.body}
          collectionName={tab.collectionName || ""}
          bind:headers={tab.headers}
          bind:settings={tab.settings}
          preRequestScript={tab.preRequestScript}
          postResponseScript={tab.postResponseScript}
        />
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
          class="h-1 w-full shrink-0 cursor-row-resize rounded-none border-0 bg-neutral-200 p-0 shadow-none transition-colors hover:bg-primary-400 dark:bg-neutral-700 {isResizing
            ? 'bg-primary-500'
            : ''}"
          onmousedown={startResize}
          aria-label="Resize response panel"
        />
      {/if}

      <div
        class="flex shrink-0 cursor-pointer items-center border-b border-neutral-200 dark:border-neutral-700"
        role="button"
        tabindex="0"
        onclick={() => (responseCollapsed = !responseCollapsed)}
        onkeydown={(e) => {
          if (e.key === "Enter" || e.key === " ") {
            e.preventDefault();
            responseCollapsed = !responseCollapsed;
          }
        }}
        aria-expanded={!responseCollapsed}
        aria-label={responseCollapsed ? "Expand response" : "Collapse response"}
      >
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <div onclick={(e) => e.stopPropagation()}>
          <Tabs bind:selected={responseTab} tabStyle="underline" classes={{ content: "hidden" }}>
            <TabItem key="body" title="Body" />
            <TabItem key="headers" title="Headers" />
          </Tabs>
        </div>
        <div class="ml-auto flex items-center gap-2 px-2">
          {#if response}
            <!-- svelte-ignore a11y_click_events_have_key_events -->
            <!-- svelte-ignore a11y_no_static_element_interactions -->
            <div onclick={(e) => e.stopPropagation()} class="flex items-center gap-2">
              <Badge color={getStatusBadgeColor(response.status)}>{response.statusText}</Badge>
              <Badge color="gray">{response.time}ms</Badge>
            </div>
          {/if}
          <Button
            color="light"
            size="xs"
            class="border-0 shadow-none"
            onclick={(e: MouseEvent) => {
              e.stopPropagation();
              responseCollapsed = !responseCollapsed;
            }}
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
        {#if loading}
          <div class="flex h-full w-full items-center justify-center">
            <Spinner type="bars" color="primary" />
          </div>
        {:else if requestError}
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
                  readOnly
                  showCopyPaste
                />
              </div>
            {:else}
              <div class="flex h-full flex-col overflow-hidden">
                <div
                  class="shrink-0 border-b border-neutral-100 bg-neutral-50 px-3 py-1 dark:border-neutral-800 dark:bg-neutral-900/50"
                >
                  <div class="flex gap-4 text-xs font-semibold text-neutral-500 uppercase">
                    <button
                      class="pb-1 {responseHeaderView === 'received'
                        ? 'border-b-2 border-primary-500 text-primary-600'
                        : ''}"
                      onclick={() => (responseHeaderView = "received")}
                    >
                      Received
                    </button>
                    <button
                      class="pb-1 {responseHeaderView === 'sent'
                        ? 'border-b-2 border-primary-500 text-primary-600'
                        : ''}"
                      onclick={() => (responseHeaderView = "sent")}
                    >
                      Sent
                    </button>
                  </div>
                </div>
                <div class="min-h-0 flex-1 overflow-y-auto">
                  {#if responseHeaderView === "received"}
                    {#each Object.entries(response.headers) as [key, value] (key)}
                      <div
                        class="flex items-start gap-2 border-b border-neutral-100 px-3 py-1.5 text-sm dark:border-neutral-800"
                      >
                        <span class="shrink-0 font-medium text-neutral-700 dark:text-neutral-300"
                          >{key}:</span
                        >
                        <span class="min-w-0 break-all text-neutral-500 dark:text-neutral-400"
                          >{value}</span
                        >
                      </div>
                    {/each}
                  {:else}
                    {#each Object.entries(response.requestHeaders || {}) as [key, value] (key)}
                      <div
                        class="flex items-start gap-2 border-b border-neutral-100 px-3 py-1.5 text-sm dark:border-neutral-800"
                      >
                        <span class="shrink-0 font-medium text-neutral-700 dark:text-neutral-300"
                          >{key}:</span
                        >
                        <span class="min-w-0 break-all text-neutral-500 dark:text-neutral-400"
                          >{value}</span
                        >
                      </div>
                    {/each}
                    {#if !response.requestHeaders}
                      <div class="p-4 text-center text-sm text-neutral-500">
                        Request headers not available for this request.
                      </div>
                    {/if}
                  {/if}
                </div>
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

{#if showCurlModal}
  <Modal title="Export as cURL" bind:open={showCurlModal} size="xl">
    <div class="h-80 overflow-hidden rounded border border-neutral-200 dark:border-neutral-700">
      <CodeMirrorEditor value={curlPreview} language="text" readOnly showCopyPaste />
    </div>
    {#snippet footer()}
      <div class="flex w-full gap-2">
        <Button
          color="primary"
          onclick={async () => {
            try {
              await SaveCurlFile(curlPreview, "request.sh");
            } catch (err) {
              notifications.error("Failed to save file", String(err));
            }
          }}
        >
          Download .sh
        </Button>
      </div>
    {/snippet}
  </Modal>
{/if}

<SaveRequestModal
  bind:show={showSaveDialog.open}
  modalId={showSaveDialog.id}
  bind:requestName
  onSave={handleSaveRequest}
  onCancel={() => (showSaveDialog.open = false)}
/>
