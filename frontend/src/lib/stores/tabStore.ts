import { writable, derived } from "svelte/store";
import type { configuration as conf } from "../../../wailsjs/go/models";
import { notifications } from "./notificationStore";

export interface TabHeader {
  id: string;
  key: string;
  value: string;
  enabled: boolean;
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
  // --- form state preserved across tab switches ---
  url: string;
  body: string;
  bodyFormat: string;
  headers: TabHeader[];
  settings: conf.RequestSettingsOverride;
  preRequestScript: string;
  postResponseScript: string;
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
    url: "",
    body: "",
    bodyFormat: "json",
    headers: [],
    settings: {},
    preRequestScript: "",
    postResponseScript: ""
  };
}

function createTabStore() {
  const { subscribe, update } = writable<TabStoreState>({
    tabs: [],
    activeTabId: null
  });

  return {
    subscribe,

    /** Open a tab for an existing saved request. If already open, just activate it. */
    openTab(
      requestId: string,
      collectionName: string,
      meta: { label: string; verb: string; url: string; body: string; bodyFormat: string; headers: TabHeader[]; settings: conf.RequestSettingsOverride; preRequestScript?: string; postResponseScript?: string }
    ) {
      update((state) => {
        const existing = state.tabs.find((t) => t.requestId === requestId);
        if (existing) {
          return { ...state, activeTabId: existing.id };
        }
        if (state.tabs.length >= MAX_OPEN_TABS) {
          notifications.warning(`Maximum ${MAX_OPEN_TABS} tabs open. Close a tab to open another.`);
          return state;
        }
        const tab: TabState = {
          id: crypto.randomUUID(),
          requestId,
          collectionName,
          label: meta.label,
          verb: meta.verb,
          isDirty: false,
          url: meta.url,
          body: meta.body,
          bodyFormat: meta.bodyFormat,
          headers: meta.headers,
          settings: meta.settings,
          preRequestScript: meta.preRequestScript ?? "",
          postResponseScript: meta.postResponseScript ?? ""
        };
        return { tabs: [...state.tabs, tab], activeTabId: tab.id };
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
      partial: Partial<Pick<TabState, "url" | "body" | "bodyFormat" | "headers" | "settings" | "verb" | "label" | "isDirty" | "preRequestScript" | "postResponseScript">>
    ) {
      update((state) => ({
        ...state,
        tabs: state.tabs.map((t) => (t.id === tabId ? { ...t, ...partial } : t))
      }));
    },

    /** Called after a successful save: bind a previously-unsaved tab to a real requestId */
    bindTabToRequest(tabId: string, requestId: string, collectionName: string, label: string) {
      update((state) => ({
        ...state,
        tabs: state.tabs.map((t) =>
          t.id === tabId ? { ...t, requestId, collectionName, label, isDirty: false } : t
        )
      }));
    },

    markDirty(tabId: string, isDirty: boolean) {
      update((state) => ({
        ...state,
        tabs: state.tabs.map((t) => (t.id === tabId ? { ...t, isDirty } : t))
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

export const activeTab = derived(tabStore, ($s) =>
  $s.tabs.find((t) => t.id === $s.activeTabId) ?? null
);
