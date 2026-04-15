/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: AGPL-3.0-only
 */

import type { InputFormat } from "$src/lib/components/RequestBuilder/types";
import { collectionStore, collectionStoreState } from "$src/lib/stores/collectionStore.svelte";
import { notifications } from "$src/lib/stores/notificationStore";
import { filterInPlace } from "$src/lib/utils/helpers";
import type { configuration as conf } from "$wails/go/models";
import { collection } from "$wails/go/models";
import { SvelteDate } from "svelte/reactivity";

export interface TabHeader {
  id: string;
  key: string;
  value: string;
  enabled: boolean;
  autoInjectedContentType?: boolean;
  injectedContentTypeValue?: string;
}

export interface TabResponse {
  status: number;
  statusText: string;
  time: number;
  headers: Record<string, string>;
  requestHeaders?: Record<string, string>;
  body: string;
}

export interface TabState {
  /** Unique tab id */
  id: string;
  /** null for unsaved requests */
  requestId: string | null;
  collectionName: string | null;
  label: string;
  verb: string;
  /** Unsaved changes indicator */
  isDirty: boolean;
  /** Preview mode: can be replaced by another request if not fixed */
  isPreview: boolean;
  // --- form state preserved across tab switches ---
  url: string;
  body: string;
  bodyFormat: InputFormat;
  headers: TabHeader[];
  auth: collection.AuthConfiguration;
  settings: conf.RequestSettingsOverride;
  preRequestScript: string;
  postResponseScript: string;
  // --- response state preserved across tab switches ---
  response: TabResponse | null;
  requestError: string | null;
}

interface TabStoreState {
  tabs: TabState[];
  activeTabIndex: number; // -1 when no active tab
}

const EMPTY_TAB_LABEL = "New Request";
const MAX_OPEN_TABS = 15;
const STORAGE_KEY = "tabs";

export const tabStoreState: TabStoreState = $state({
  tabs: (JSON.parse(localStorage.getItem(STORAGE_KEY) || "{}").tabs || []) as TabState[],
  activeTabIndex:
    (JSON.parse(localStorage.getItem(STORAGE_KEY) || "{}").activeTabIndex as number) || -1
});

/** Get mutable reference to active tab */
export function getActiveTab(): TabState {
  return tabStoreState.tabs[tabStoreState.activeTabIndex];
}

function getTabIndexById(tabId: string): number {
  return tabStoreState.tabs.findIndex((t) => t.id === tabId);
}

function normalizeRequestSettings(
  settings?: conf.RequestSettingsOverride | null
): conf.RequestSettingsOverride {
  if (!settings) return {};
  return {
    timeoutSeconds: settings.timeoutSeconds,
    followRedirects: settings.followRedirects,
    maxRedirects: settings.maxRedirects,
    validateSSL: settings.validateSSL,
    defaultUserAgent: settings.defaultUserAgent,
    proxyUrl: settings.proxyUrl
  };
}

function normalizeAuthConfiguration(
  auth?: collection.AuthConfiguration | null
): collection.AuthConfiguration {
  return {
    enabled: auth?.enabled ?? false,
    tokenUrl: auth?.tokenUrl ?? "",
    template: { ...(auth?.template ?? {}) },
    tokenPath: auth?.tokenPath ?? "access_token"
  };
}

export function makeEmptyTab(): TabState {
  const newTab: TabState = {
    id: crypto.randomUUID(),
    requestId: null,
    collectionName: null,
    label: EMPTY_TAB_LABEL,
    verb: "GET",
    isDirty: false,
    isPreview: true,
    url: "",
    body: "",
    bodyFormat: "json",
    headers: [],
    auth: normalizeAuthConfiguration(),
    settings: normalizeRequestSettings(),
    preRequestScript: "",
    postResponseScript: "",
    response: null,
    requestError: null
  };
  tabStoreState.tabs.push(newTab);
  tabStoreState.activeTabIndex = tabStoreState.tabs.length - 1;
  storeTabsInLocalStorage();
  return newTab;
}

export function openTab(
  requestId: string,
  collectionName: string,
  meta: {
    label: string;
    verb: string;
    url: string;
    body: string;
    bodyFormat: InputFormat;
    headers: TabHeader[];
    auth?: collection.AuthConfiguration;
    settings: conf.RequestSettingsOverride;
    preRequestScript?: string;
    postResponseScript?: string;
  }
) {
  // 1. If already open, just activate it
  const existingIndex = tabStoreState.tabs.findIndex((t) => t.requestId === requestId);
  if (existingIndex !== -1) {
    tabStoreState.activeTabIndex = existingIndex;
    storeTabsInLocalStorage();
    return;
  }

  // 2. Look for a replaceable tab (isPreview mode)
  const activeTab = getActiveTab();

  let tabIndexToReplace: number | undefined;
  if (activeTab && activeTab.isPreview) {
    tabIndexToReplace = tabStoreState.activeTabIndex;
  } else {
    tabIndexToReplace = tabStoreState.tabs.findIndex((t) => t.isPreview);
  }

  const defaultAuth = normalizeAuthConfiguration();

  if (tabIndexToReplace !== undefined && tabIndexToReplace >= 0) {
    const tab = tabStoreState.tabs[tabIndexToReplace];
    tab.requestId = requestId;
    tab.collectionName = collectionName;
    tab.label = meta.label;
    tab.verb = meta.verb;
    tab.isDirty = false;
    tab.isPreview = true;
    tab.url = meta.url;
    tab.body = meta.body;
    tab.bodyFormat = meta.bodyFormat;
    tab.headers = meta.headers;
    tab.auth = normalizeAuthConfiguration(meta.auth ?? defaultAuth);
    tab.settings = normalizeRequestSettings(meta.settings);
    tab.preRequestScript = meta.preRequestScript ?? "";
    tab.postResponseScript = meta.postResponseScript ?? "";
    tab.response = null;
    tab.requestError = null;
    tabStoreState.activeTabIndex = tabIndexToReplace;
    storeTabsInLocalStorage();
    return;
  }

  // 3. If no replaceable tab, create a new one (respecting limit)
  if (tabStoreState.tabs.length >= MAX_OPEN_TABS) {
    notifications.warning(`Maximum ${MAX_OPEN_TABS} tabs open. Close a tab to open another.`);
    return;
  }

  const newTab: TabState = {
    id: crypto.randomUUID(),
    requestId,
    collectionName,
    label: meta.label,
    verb: meta.verb,
    isDirty: false,
    isPreview: true,
    url: meta.url,
    body: meta.body,
    bodyFormat: meta.bodyFormat,
    headers: meta.headers,
    auth: normalizeAuthConfiguration(meta.auth ?? defaultAuth),
    settings: normalizeRequestSettings(meta.settings),
    preRequestScript: meta.preRequestScript ?? "",
    postResponseScript: meta.postResponseScript ?? "",
    response: null,
    requestError: null
  };
  tabStoreState.tabs.push(newTab);
  tabStoreState.activeTabIndex = tabStoreState.tabs.length - 1;
  storeTabsInLocalStorage();
}

export function closeTab(tabId: string) {
  const idx = getTabIndexById(tabId);
  if (idx === -1) return;
  tabStoreState.tabs.splice(idx, 1);
  if (tabStoreState.activeTabIndex === idx) {
    if (tabStoreState.tabs.length === 0) {
      tabStoreState.activeTabIndex = -1;
    } else {
      tabStoreState.activeTabIndex = Math.max(0, idx - 1);
    }
  } else if (tabStoreState.activeTabIndex > idx) {
    tabStoreState.activeTabIndex--;
  }
  storeTabsInLocalStorage();
}

export function setActiveTab(tabId: string) {
  const idx = getTabIndexById(tabId);
  tabStoreState.activeTabIndex = idx;
}

export function updateTabFormState(
  tabId: string,
  partial: Partial<
    Pick<
      TabState,
      | "url"
      | "body"
      | "bodyFormat"
      | "headers"
      | "auth"
      | "settings"
      | "verb"
      | "label"
      | "isDirty"
      | "preRequestScript"
      | "postResponseScript"
    >
  >
) {
  const tab = tabStoreState.tabs.find((t) => t.id === tabId);
  if (tab) {
    Object.assign(tab, partial, {
      isDirty: partial.isDirty ?? true,
      isPreview: false
    });
  }
  storeTabsInLocalStorage();
}

/** Called after a successful save: bind a previously-unsaved tab to a real requestId */
export function bindTabToRequest(
  tabId: string,
  requestId: string,
  collectionName: string,
  label: string
) {
  const tab = tabStoreState.tabs.find((t) => t.id === tabId);
  if (tab) {
    tab.requestId = requestId;
    tab.collectionName = collectionName;
    tab.label = label;
    tab.isDirty = false;
    tab.isPreview = false;
  }
  storeTabsInLocalStorage();
}

export async function saveTab(tabId: string) {
  const tab = tabStoreState.tabs.find((t) => t.id === tabId);
  if (!tab || !tab.requestId || !tab.collectionName) return;

  const collections = collectionStoreState.collections;
  let originalRequest: collection.Request | null = null;
  for (const coll of collections) {
    const found = coll.requests.find((r) => r.id === tab.requestId);
    if (found) {
      originalRequest = found;
      break;
    }
  }

  if (!originalRequest) return;

  const headersObj = tab.headers
    .filter((h) => h.enabled && h.key)
    .reduce((acc, { key, value }) => ({ ...acc, [key]: value }), {} as Record<string, string>);

  try {
    await collectionStore.updateRequest(
      tab.collectionName,
      collection.Request.createFrom({
        ...originalRequest,
        name: tab.label,
        url: tab.url,
        verb: tab.verb,
        body: tab.body,
        headers: headersObj,
        auth: tab.auth,
        settings: tab.settings,
        preRequestScript: tab.preRequestScript,
        postResponseScript: tab.postResponseScript,
        lastUpdateTimestamp: new SvelteDate().toISOString()
      })
    );
    tab.isDirty = false;
    tab.isPreview = false;
  } catch (error) {
    notifications.error("Failed to save request");
    throw error;
  }
  storeTabsInLocalStorage();
}

export function updateTabResponse(
  tabId: string,
  response: TabResponse | null,
  requestError: string | null
) {
  const tab = tabStoreState.tabs.find((t) => t.id === tabId);
  if (tab) {
    tab.response = response;
    tab.requestError = requestError;
    tab.isPreview = false;
  }
  storeTabsInLocalStorage();
}

export function renameTabsByRequestId(requestId: string, label: string) {
  const tab = tabStoreState.tabs.find((t) => t.requestId === requestId);
  if (tab) {
    tab.label = label;
  }
  storeTabsInLocalStorage();
}

/** Remove all tabs referencing a deleted request */
export function removeTabsForRequest(requestId: string) {
  const wasActiveTab = tabStoreState.tabs[tabStoreState.activeTabIndex]?.requestId === requestId;
  filterInPlace(tabStoreState.tabs, (t: TabState) => t.requestId !== requestId);
  if (wasActiveTab) {
    tabStoreState.activeTabIndex =
      tabStoreState.tabs.length > 0 ? tabStoreState.tabs.length - 1 : -1;
  }
  storeTabsInLocalStorage();
}

export function storeTabsInLocalStorage() {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(tabStoreState));
}

export const tabStore = {
  makeEmptyTab,
  openTab,
  closeTab,
  setActiveTab,
  updateTabFormState,
  bindTabToRequest,
  saveTab,
  updateTabResponse,
  renameTabsByRequestId,
  removeTabsForRequest,
  storeTabsInLocalStorage
};
