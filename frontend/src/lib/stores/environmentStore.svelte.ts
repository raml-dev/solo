/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: AGPL-3.0-only
 */

import { notifications } from "$src/lib/stores/notificationStore";
import { DEFAULT_ENV_NAME } from "$src/lib/utils/constants";
import {
  CreateEnvironment,
  DeleteEnvironment,
  GetSelectedEnvironment,
  LoadEnvironment,
  LoadEnvironments,
  RenameEnvironment,
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
  renameEnvironmentName: string | null;
}

const initialState: EnvironmentState = {
  environments: [],
  selectedEnvironmentName: null,
  loading: false,
  renameEnvironmentName: null
};

export const environmentStoreState = $state<EnvironmentState>({ ...initialState });

function hasEnvironmentName(name: string): boolean {
  return environmentStoreState.environments.some(
    (environment) => environment.name.toLowerCase() === name.toLowerCase()
  );
}

function getDuplicateEnvironmentName(sourceName: string): string {
  let copyIndex = 1;
  let candidateName = `${sourceName} (copy)`;

  while (hasEnvironmentName(candidateName)) {
    copyIndex += 1;
    candidateName = `${sourceName} (copy ${copyIndex})`;
  }

  return candidateName;
}

function cloneEnvironmentValues(values: Record<string, environment.ValueType> | undefined) {
  return Object.fromEntries(
    Object.entries(values || {}).map(([key, value]) => [
      key,
      new environment.ValueType({
        value: value?.value || "",
        type: value?.type || "text"
      })
    ])
  );
}

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

  startRenameEnvironment(name: string) {
    environmentStoreState.renameEnvironmentName = name;
  },

  consumeRenameEnvironment() {
    environmentStoreState.renameEnvironmentName = null;
  },

  async renameEnvironment(currentName: string, nextName: string) {
    const trimmedName = nextName.trim();
    if (!trimmedName) {
      throw new Error("Environment name cannot be empty");
    }

    if (currentName === trimmedName) {
      return;
    }

    environmentStoreState.loading = true;
    try {
      await RenameEnvironment(currentName, trimmedName);
      await this.loadEnvironments();
      return environmentStoreState.environments.find((e) => e.name === trimmedName);
    } catch (err) {
      notifications.error("Failed to rename environment", String(err));
      throw err;
    } finally {
      environmentStoreState.loading = false;
      environmentStoreState.renameEnvironmentName = null;
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

  async duplicateEnvironment(name: string) {
    const sourceEnvironment = environmentStoreState.environments.find(
      (environment) => environment.name === name
    );
    if (!sourceEnvironment) {
      throw new Error(`Environment "${name}" not found`);
    }

    const duplicateName = getDuplicateEnvironmentName(sourceEnvironment.name);

    try {
      await this.createEnvironment(duplicateName);

      const createdEnvironment = environmentStoreState.environments.find(
        (environment) => environment.name === duplicateName
      );
      if (!createdEnvironment) {
        throw new Error(`Failed to create duplicate environment "${duplicateName}"`);
      }

      await this.updateEnvironment(
        new environment.Environment({
          ...createdEnvironment,
          values: cloneEnvironmentValues(sourceEnvironment.values)
        })
      );

      await this.loadEnvironments();
      this.selectEnvironment(duplicateName);
      return duplicateName;
    } catch (err) {
      notifications.error("Failed to duplicate environment", String(err));
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
