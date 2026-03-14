import { ClearSessionVars, GetSessionVars, RemoveSessionVar } from "$wails/go/main/App";
import { EventsOn } from "$wails/runtime/runtime";
import { writable } from "svelte/store";

function createSessionVarsStore() {
  const { subscribe, set } = writable<Record<string, string>>({});

  return {
    subscribe,

    /** Called once at app startup to load existing vars and listen for updates */
    init() {
      // Load current vars from backend
      GetSessionVars()
        .then((vars) => set(vars ?? {}))
        .catch(() => {});

      // Listen for real-time updates emitted by env.set() in Lua
      EventsOn("session_vars_updated", (vars: Record<string, string>) => {
        set(vars ?? {});
      });
    },

    /** Refresh from backend (e.g. after a request completes) */
    async refresh() {
      try {
        const vars = await GetSessionVars();
        set(vars ?? {});
      } catch {
        /* silent */
      }
    },

    /** Remove one session var by key */
    async remove(key: string) {
      try {
        await RemoveSessionVar(key);
        const vars = await GetSessionVars();
        set(vars ?? {});
      } catch {
        /* silent */
      }
    },

    /** Clear all session vars */
    async clear() {
      try {
        await ClearSessionVars();
        set({});
      } catch {
        /* silent */
      }
    }
  };
}

export const sessionVarsStore = createSessionVarsStore();
