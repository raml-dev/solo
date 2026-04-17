/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: AGPL-3.0-only
 */

import { notifications } from "$src/lib/stores/notificationStore";
import {
  CreateEnvironment,
  DeleteEnvironment,
  GetSelectedEnvironment,
  LoadEnvironment,
  LoadEnvironments,
  SetSelectedEnvironment,
  UpdateEnvironment
} from "$wails/go/main/App";
import { environment } from "$wails/go/models";

// Re-export types for convenience
export type Environment = environment.Environment;
export type EnvironmentValue = environment.ValueType;

interface EnvironmentState {
  environments: Environment[];
  selectedEnvironmentName: string | null;
  loading: boolean;
}

const initialState: EnvironmentState = {
  environments: [],
  selectedEnvironmentName: null,
  loading: false
};

export const DEFAULT_ENV_NAME = "default";

export const environmentStoreState = $state<EnvironmentState>({ ...initialState });

export const environmentStore = {
  async loadEnvironments() {
    environmentStoreState.loading = true;
    try {
      const [environmentNames, persistedSelection] = await Promise.all([
        LoadEnvironments(),
        GetSelectedEnvironment()
      ]);

      if (!environmentNames || environmentNames.length === 0) {
        environmentStoreState.environments = [];
        environmentStoreState.loading = false;
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

      environmentStoreState.environments = environments;
      environmentStoreState.selectedEnvironmentName =
        persistedSelection || environmentStoreState.selectedEnvironmentName;
      environmentStoreState.loading = false;
    } catch (err) {
      notifications.error("Failed to load environments", String(err), true);
      environmentStoreState.loading = false;
    }
  },

  async createEnvironment(name: string) {
    environmentStoreState.loading = true;
    try {
      await CreateEnvironment(name);
      await this.loadEnvironments();
    } catch (err) {
      notifications.error("Failed to create environment", String(err));
      environmentStoreState.loading = false;
      throw err;
    }
  },

  async deleteEnvironment(name: string) {
    environmentStoreState.loading = true;
    try {
      await DeleteEnvironment(name);
      environmentStoreState.environments = environmentStoreState.environments.filter(
        (e) => e.name !== name
      );
      if (environmentStoreState.selectedEnvironmentName === name) {
        this.selectEnvironment(DEFAULT_ENV_NAME);
      }
      environmentStoreState.loading = false;
    } catch (err) {
      notifications.error("Failed to delete environment", String(err));
      environmentStoreState.loading = false;
      throw err;
    }
  },

  async updateEnvironment(env: Environment) {
    try {
      const envInstance = new environment.Environment(env);
      await UpdateEnvironment(envInstance);
      environmentStoreState.environments = environmentStoreState.environments.map((e) =>
        e.name === env.name ? env : e
      );
    } catch (err) {
      notifications.error("Failed to save environment", String(err));
      throw err;
    }
  },

  selectEnvironment(name: string) {
    if (name === "") throw new Error("Environment name must not be empty string");
    environmentStoreState.selectedEnvironmentName = name;
    SetSelectedEnvironment(name).catch((err) => {
      notifications.error("Failed to persist selected environment", String(err));
    });
  }
};
