/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: AGPL-3.0-only
 */

import { notifications } from "$src/lib/stores/notificationStore";
import type { ThemeMode } from "$src/lib/theme/themeModel";
import {
  applyTheme,
  applyThemeMode as applyThemeModeRuntime,
  ensureSystemThemeModeListener,
  setActiveThemeMode
} from "$src/lib/theme/themeRuntime";
import { debounce } from "$src/lib/utils/debounce";
import {
  GetAllThemes,
  GetConfiguration,
  SetActiveTheme,
  UpdateConfiguration
} from "$wails/go/main/App";
import { configuration, theme } from "$wails/go/models";

const SAVE_DEBOUNCE_MS = 800;
const SAVED_STATUS_TIMEOUT_MS = 1500;

function createEmptyConfig(): configuration.Configuration {
  const general = { ...new configuration.GeneralSettings() };
  const request = { ...new configuration.RequestSettings() };

  return {
    general,
    request,
    customThemes: []
  } as unknown as configuration.Configuration;
}

function cloneConfig(configToClone: configuration.Configuration): configuration.Configuration {
  return JSON.parse(JSON.stringify(configToClone)) as configuration.Configuration;
}

function normalizeConfig(raw?: configuration.Configuration | null): configuration.Configuration {
  const base = createEmptyConfig();
  if (!raw) return base;

  const source = cloneConfig(raw);
  return {
    general: {
      ...base.general,
      ...(source.general || {})
    },
    request: {
      ...base.request,
      ...(source.request || {})
    },
    customThemes: (source.customThemes || []).map((t) => new theme.Theme(t))
  } as unknown as configuration.Configuration;
}

function getPersistenceSignature(cfg: configuration.Configuration): string {
  return JSON.stringify({
    general: {
      activeTheme: cfg.general?.activeTheme ?? "",
      themeMode: cfg.general?.themeMode ?? "system",
      checkForUpdates: cfg.general?.checkForUpdates ?? false,
      debugMode: cfg.general?.debugMode ?? false,
      dayTheme: cfg.general?.dayTheme ?? "",
      nightTheme: cfg.general?.nightTheme ?? "",
      selectedEnvironment: cfg.general?.selectedEnvironment ?? ""
    },
    request: {
      timeoutSeconds: cfg.request?.timeoutSeconds ?? 30,
      followRedirects: cfg.request?.followRedirects ?? true,
      maxRedirects: cfg.request?.maxRedirects ?? 10,
      validateSSL: cfg.request?.validateSSL ?? true,
      defaultUserAgent: cfg.request?.defaultUserAgent ?? "",
      proxyUrl: cfg.request?.proxyUrl ?? ""
    },
    customThemes: cfg.customThemes || []
  });
}

export interface ConfigurationStoreState {
  config: configuration.Configuration;
  allThemes: theme.Theme[];
  initialized: boolean;
  saveStatus: "idle" | "saving" | "saved" | "error";
}

export const configurationStoreState: ConfigurationStoreState = $state({
  config: createEmptyConfig(),
  allThemes: [],
  initialized: false,
  saveStatus: "idle"
});

let isHydratingConfig = false;
let lastPersistedSignature: string | null = null;

export function getConfigSnapshot(): configuration.Configuration {
  return cloneConfig(configurationStoreState.config);
}

export function getActiveTheme(): theme.Theme | null {
  const { config, allThemes } = configurationStoreState;
  if (!config?.general?.activeTheme || allThemes.length === 0) {
    return null;
  }
  return allThemes.find((t) => t.id === config.general.activeTheme) || null;
}

function applyActiveTheme() {
  const currentTheme = getActiveTheme();
  if (currentTheme) {
    applyTheme(currentTheme);
  }
}

async function persistConfig() {
  if (!configurationStoreState.initialized || isHydratingConfig) return;

  const snapshot = getConfigSnapshot();
  const signature = getPersistenceSignature(snapshot);
  if (signature === lastPersistedSignature) return;

  configurationStoreState.saveStatus = "saving";
  try {
    await UpdateConfiguration(snapshot);
    lastPersistedSignature = signature;
    configurationStoreState.saveStatus = "saved";
    setTimeout(() => {
      if (configurationStoreState.saveStatus === "saved") {
        configurationStoreState.saveStatus = "idle";
      }
    }, SAVED_STATUS_TIMEOUT_MS);
  } catch (err) {
    configurationStoreState.saveStatus = "error";
    notifications.error("Could not save settings", String(err));
  }
}

const debouncedPersistConfig = debounce(() => {
  void persistConfig();
}, SAVE_DEBOUNCE_MS);

export function saveConfig() {
  debouncedPersistConfig();
}

export const configurationStore = {
  async init() {
    isHydratingConfig = true;
    try {
      const [cfg, themes] = await Promise.all([GetConfiguration(), GetAllThemes()]);
      const normalizedConfig = normalizeConfig(cfg);
      const normalizedThemes = (themes || []).map((t) => new theme.Theme(t));

      configurationStoreState.config = normalizedConfig;
      configurationStoreState.allThemes = normalizedThemes;
      configurationStoreState.initialized = true;
      configurationStoreState.saveStatus = "idle";
      lastPersistedSignature = getPersistenceSignature(normalizedConfig);

      const mode = normalizedConfig.general?.themeMode as ThemeMode;
      if (mode) {
        setActiveThemeMode(mode);
        applyThemeModeRuntime(mode);
      }
      applyActiveTheme();
    } catch (error) {
      notifications.error("Failed to initialize configuration", String(error), true);
    } finally {
      isHydratingConfig = false;
    }
  },

  async changeTheme(themeId: string) {
    try {
      configurationStoreState.config.general.activeTheme = themeId;
      await SetActiveTheme(themeId);
      lastPersistedSignature = getPersistenceSignature(configurationStoreState.config);
      applyActiveTheme();
    } catch (error) {
      notifications.error("Failed to change theme", String(error));
      throw error;
    }
  },

  applyThemeMode(mode: ThemeMode) {
    setActiveThemeMode(mode);
    applyThemeModeRuntime(mode);
  }
};

ensureSystemThemeModeListener();
