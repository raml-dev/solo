import { writable, derived, get } from "svelte/store";
import type { configuration as conf } from "../../../wailsjs/go/models";
import { collection } from "../../../wailsjs/go/models";
import { notifications } from "./notificationStore";
import { collectionStore } from "./collectionStore";

export interface TabHeader {
  id: string;
  key: string;
  value: string;
  enabled: boolean;
}

export interface TabResponse {
  status: number;
  statusText: string;
  time: number;
  headers: Record<string, string>;
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
  bodyFormat: string;
  headers: TabHeader[];
  settings: conf.RequestSettingsOverride;
  preRequestScript: string;
  postResponseScript: string;
  // --- response state preserved across tab switches ---
  response: TabResponse | null;
  requestError: string | null;
}

interface TabStoreState {
  tabs: TabState[];
  activeTabId: string | null;
}

const EMPTY_TAB_LABEL = "New Request";
const MAX_OPEN_TABS = 15;

function makeEmptyTab(): TabState {
  return {
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
    settings: {},
    preRequestScript: "",
    postResponseScript: "",
    response: null,
    requestError: null
  };
}

function createTabStore() {
  const { subscribe, update } = writable<TabStoreState>({
    tabs: [],
    activeTabId: null
  });

  return {
    subscribe,

    /** Open a tab for an existing saved request. If already open, just activate it.
     *  Implements "preview tab" logic: replaces an existing tab if it's in preview mode.
     */
    openTab(
      requestId: string,
      collectionName: string,
      meta: {
        label: string;
        verb: string;
        url: string;
        body: string;
        bodyFormat: string;
        headers: TabHeader[];
        settings: conf.RequestSettingsOverride;
        preRequestScript?: string;
        postResponseScript?: string;
      }
    ) {
      update((state) => {
        // 1. If already open, just activate it
        const existing = state.tabs.find((t) => t.requestId === requestId);
        if (existing) {
          return { ...state, activeTabId: existing.id };
        }

        // 2. Look for a replaceable tab (isPreview mode)
        // We prioritize the active tab if it's replaceable, otherwise the first replaceable one.
        const activeTab = state.tabs.find((t) => t.id === state.activeTabId);

        let tabToReplace: TabState | undefined;
        if (activeTab && activeTab.isPreview) {
          tabToReplace = activeTab;
        } else {
          tabToReplace = state.tabs.find((t) => t.isPreview);
        }

        if (tabToReplace) {
          return {
            ...state,
            activeTabId: tabToReplace.id,
            tabs: state.tabs.map((t) =>
              t.id === tabToReplace?.id
                ? {
                    ...t,
                    requestId,
                    collectionName,
                    label: meta.label,
                    verb: meta.verb,
                    isDirty: false,
                    isPreview: true, // Remains in preview until interacted with
                    url: meta.url,
                    body: meta.body,
                    bodyFormat: meta.bodyFormat,
                    headers: meta.headers,
                    settings: meta.settings,
                    preRequestScript: meta.preRequestScript ?? "",
                    postResponseScript: meta.postResponseScript ?? "",
                    response: null,
                    requestError: null
                  }
                : t
            )
          };
        }

        // 3. If no replaceable tab, create a new one (respecting limit)
        if (state.tabs.length >= MAX_OPEN_TABS) {
          notifications.warning(`Maximum ${MAX_OPEN_TABS} tabs open. Close a tab to open another.`);
          return state;
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
          settings: meta.settings,
          preRequestScript: meta.preRequestScript ?? "",
          postResponseScript: meta.postResponseScript ?? "",
          response: null,
          requestError: null
        };
        return { tabs: [...state.tabs, newTab], activeTabId: newTab.id };
      });
    },

    /** Open a blank unsaved tab (from + button) */
    newEmptyTab() {
      const tab = makeEmptyTab();
      update((state) => {
        if (state.tabs.length >= MAX_OPEN_TABS) {
          notifications.warning(`Maximum ${MAX_OPEN_TABS} tabs open. Close a tab to open another.`);
          return state;
        }
        return {
          tabs: [...state.tabs, tab],
          activeTabId: tab.id
        };
      });
    },

    /** Close a tab. If it was active, activate the nearest one. */
    closeTab(tabId: string) {
      update((state) => {
        const idx = state.tabs.findIndex((t) => t.id === tabId);
        if (idx === -1) return state;
        const newTabs = state.tabs.filter((t) => t.id !== tabId);
        let newActiveId = state.activeTabId;
        if (state.activeTabId === tabId) {
          if (newTabs.length === 0) {
            newActiveId = null;
          } else {
            // activate the tab to the left, or the first one
            newActiveId = newTabs[Math.max(0, idx - 1)].id;
          }
        }
        return { tabs: newTabs, activeTabId: newActiveId };
      });
    },

    setActiveTab(tabId: string) {
      update((state) => ({ ...state, activeTabId: tabId }));
    },

    /** Update persisted form state for the active tab (called by the builder on every change) */
    updateTabFormState(
      tabId: string,
      partial: Partial<
        Pick<
          TabState,
          | "url"
          | "body"
          | "bodyFormat"
          | "headers"
          | "settings"
          | "verb"
          | "label"
          | "isDirty"
          | "preRequestScript"
          | "postResponseScript"
        >
      >
    ) {
      update((state) => ({
        ...state,
        tabs: state.tabs.map((t) =>
          t.id === tabId
            ? {
                ...t,
                ...partial,
                isDirty: partial.isDirty ?? true,
                isPreview: false // Interaction fixes the tab
              }
            : t
        )
      }));
    },

    /** Called after a successful save: bind a previously-unsaved tab to a real requestId */
    bindTabToRequest(tabId: string, requestId: string, collectionName: string, label: string) {
      update((state) => ({
        ...state,
        tabs: state.tabs.map((t) =>
          t.id === tabId
            ? { ...t, requestId, collectionName, label, isDirty: false, isPreview: false }
            : t
        )
      }));
    },

    async saveTab(tabId: string) {
      const state = get(this);
      const tab = state.tabs.find((t) => t.id === tabId);
      if (!tab || !tab.requestId || !tab.collectionName) return;

      const collections = get(collectionStore).collections;
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
            url: tab.url,
            verb: tab.verb,
            body: tab.body,
            headers: headersObj,
            settings: tab.settings,
            preRequestScript: tab.preRequestScript,
            postResponseScript: tab.postResponseScript,
            lastUpdateTimestamp: new Date().toISOString()
          })
        );
        this.markDirty(tabId, false);
        this.fixTab(tabId);
      } catch (error) {
        notifications.error("Failed to save request");
        throw error;
      }
    },

    updateTabResponse(tabId: string, response: TabResponse | null, requestError: string | null) {
      update((state) => ({
        ...state,
        tabs: state.tabs.map((t) =>
          t.id === tabId ? { ...t, response, requestError, isPreview: false } : t
        )
      }));
    },

    markDirty(tabId: string, isDirty: boolean) {
      update((state) => ({
        ...state,
        tabs: state.tabs.map((t) =>
          t.id === tabId ? { ...t, isDirty, isPreview: isDirty ? false : t.isPreview } : t
        )
      }));
    },

    fixTab(tabId: string) {
      update((state) => ({
        ...state,
        tabs: state.tabs.map((t) => (t.id === tabId ? { ...t, isPreview: false } : t))
      }));
    },

    /** Remove all tabs referencing a deleted request */
    removeTabsForRequest(requestId: string) {
      update((state) => {
        const newTabs = state.tabs.filter((t) => t.requestId !== requestId);
        let newActiveId = state.activeTabId;
        if (!newTabs.find((t) => t.id === state.activeTabId)) {
          newActiveId = newTabs.length > 0 ? newTabs[newTabs.length - 1].id : null;
        }
        return { tabs: newTabs, activeTabId: newActiveId };
      });
    }
  };
}

export const tabStore = createTabStore();

export const activeTab = derived(
  tabStore,
  ($s) => $s.tabs.find((t) => t.id === $s.activeTabId) ?? null
);
