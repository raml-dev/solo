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

// Re-export types for convenience
export type Environment = environment.Environment;
export type EnvironmentValue = environment.ValueType;

// Store state
interface EnvironmentState {
  environments: Environment[];
  selectedEnvironmentName: string | null;
  loading: boolean;
  error: string | null;
}

const initialState: EnvironmentState = {
  environments: [],
  selectedEnvironmentName: null,
  loading: false,
  error: null
};

// Create the main store
function createEnvironmentStore() {
  const { subscribe, update } = writable<EnvironmentState>(initialState);

  return {
    subscribe,

    // Load all environments
    async loadEnvironments() {
      update((state) => ({ ...state, loading: true, error: null }));
      try {
        const [environmentNames, persistedSelection] = await Promise.all([
          LoadEnvironments(),
          GetSelectedEnvironment()
        ]);

        if (!environmentNames || environmentNames.length === 0) {
          update((state) => ({ ...state, environments: [], loading: false }));
          return;
        }

        // Load full content for each environment
        const environments: Environment[] = [];
        for (const name of environmentNames) {
          // Remove .json extension if present
          const cleanName = name.replace(".json", "");
          try {
            const env = await LoadEnvironment(cleanName);
            if (env) {
              environments.push(env);
            }
          } catch (err) {
            console.error(`Error loading environment ${cleanName}:`, err);
          }
        }

        update((state) => ({
          ...state,
          environments,
          selectedEnvironmentName: persistedSelection || state.selectedEnvironmentName,
          loading: false
        }));
      } catch (err) {
        update((state) => ({
          ...state,
          error: err.message || "Failed to load environments",
          loading: false
        }));
      }
    },

    // Create a new environment
    async createEnvironment(name: string) {
      update((state) => ({ ...state, loading: true, error: null }));
      try {
        await CreateEnvironment(name);
        // Reload all environments to get the updated list
        await this.loadEnvironments();
      } catch (err) {
        update((state) => ({
          ...state,
          error: err.message || "Failed to create environment",
          loading: false
        }));
        throw err;
      }
    },

    // Delete an environment
    async deleteEnvironment(name: string) {
      update((state) => ({ ...state, loading: true, error: null }));
      try {
        await DeleteEnvironment(name);
        update((state) => {
          const newState = {
            ...state,
            environments: state.environments.filter((e) => e.name !== name),
            loading: false
          };
          // If we deleted the selected environment, clear the selection
          if (state.selectedEnvironmentName === name) {
            newState.selectedEnvironmentName = null;
          }
          return newState;
        });
      } catch (err) {
        update((state) => ({
          ...state,
          error: err.message || "Failed to delete environment",
          loading: false
        }));
        throw err;
      }
    },

    // Update an environment
    async updateEnvironment(env: Environment) {
      update((state) => ({ ...state, loading: true, error: null }));
      try {
        // Create a new instance with the correct class methods
        const envInstance = new environment.Environment(env);
        await UpdateEnvironment(envInstance);
        update((state) => ({
          ...state,
          environments: state.environments.map((e) => (e.name === env.name ? env : e)),
          loading: false
        }));
      } catch (err) {
        update((state) => ({
          ...state,
          error: err.message || "Failed to update environment",
          loading: false
        }));
        throw err;
      }
    },

    // Select an environment and persist the selection
    selectEnvironment(name: string | null) {
      update((state) => ({ ...state, selectedEnvironmentName: name }));
      SetSelectedEnvironment(name ?? "").catch((err) => {
        console.error("Failed to persist selected environment:", err);
      });
    },

    // Clear error
    clearError() {
      update((state) => ({ ...state, error: null }));
    }
  };
}

export const environmentStore = createEnvironmentStore();

// Derived store for the selected environment
export const selectedEnvironment = derived(
  environmentStore,
  ($store) => $store.environments.find((e) => e.name === $store.selectedEnvironmentName) || null
);
