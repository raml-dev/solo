import { writable, derived } from "svelte/store";
import {
  CreateEnvironment,
  LoadEnvironments,
  LoadEnvironment,
  UpdateEnvironment,
  DeleteEnvironment,
  GetSelectedEnvironment,
  SetSelectedEnvironment
} from "../../../wailsjs/go/main/App";
import { environment } from "../../../wailsjs/go/models";
import { notifications } from "./notificationStore";

// Re-export types for convenience
export type Environment = environment.Environment;
export type EnvironmentValue = environment.ValueType;

// Store state
interface EnvironmentState {
  environments: Environment[];
  selectedEnvironmentName: string | null;
  loading: boolean;
}

const initialState: EnvironmentState = {
  environments: [],
  selectedEnvironmentName: null,
  loading: false,
};

// Create the main store
function createEnvironmentStore() {
  const { subscribe, update } = writable<EnvironmentState>(initialState);

  return {
    subscribe,

    async loadEnvironments() {
      update((state) => ({ ...state, loading: true }));
      try {
        const [environmentNames, persistedSelection] = await Promise.all([
          LoadEnvironments(),
          GetSelectedEnvironment()
        ]);

        if (!environmentNames || environmentNames.length === 0) {
          update((state) => ({ ...state, environments: [], loading: false }));
          return;
        }

        const environments: Environment[] = [];
        for (const name of environmentNames) {
          const cleanName = name.replace(".json", "");
          try {
            const env = await LoadEnvironment(cleanName);
            if (env) environments.push(env);
          } catch (err) {
            notifications.error(`Failed to load environment "${cleanName}"`, String(err));
          }
        }

        update((state) => ({
          ...state,
          environments,
          selectedEnvironmentName: persistedSelection || state.selectedEnvironmentName,
          loading: false
        }));
      } catch (err) {
        notifications.error("Failed to load environments", String(err), true);
        update((state) => ({ ...state, loading: false }));
      }
    },

    async createEnvironment(name: string) {
      update((state) => ({ ...state, loading: true }));
      try {
        await CreateEnvironment(name);
        await this.loadEnvironments();
      } catch (err) {
        notifications.error("Failed to create environment", String(err));
        update((state) => ({ ...state, loading: false }));
        throw err;
      }
    },

    async deleteEnvironment(name: string) {
      update((state) => ({ ...state, loading: true }));
      try {
        await DeleteEnvironment(name);
        update((state) => {
          const newState = {
            ...state,
            environments: state.environments.filter((e) => e.name !== name),
            loading: false
          };
          if (state.selectedEnvironmentName === name) newState.selectedEnvironmentName = null;
          return newState;
        });
      } catch (err) {
        notifications.error("Failed to delete environment", String(err));
        update((state) => ({ ...state, loading: false }));
        throw err;
      }
    },

    async updateEnvironment(env: Environment) {
      try {
        const envInstance = new environment.Environment(env);
        await UpdateEnvironment(envInstance);
        update((state) => ({
          ...state,
          environments: state.environments.map((e) => (e.name === env.name ? env : e))
        }));
      } catch (err) {
        notifications.error("Failed to save environment", String(err));
        throw err;
      }
    },

    selectEnvironment(name: string) {
      if (name === "") throw new Error("Environment name must not be empty string");
      update((state) => ({ ...state, selectedEnvironmentName: name }));
      SetSelectedEnvironment(name).catch((err) => {
        notifications.error("Failed to persist selected environment", String(err));
      });
    },
  };
}

export const environmentStore = createEnvironmentStore();

// Derived store for the selected environment
export const selectedEnvironment = derived(
  environmentStore,
  ($store) => $store.environments.find((e) => e.name === $store.selectedEnvironmentName) || null
);
