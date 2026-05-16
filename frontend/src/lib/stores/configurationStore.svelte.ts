/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: AGPL-3.0-only
 */

import { sanitizeFontFamily } from "$src/lib/fonts";
import { notifications } from "$src/lib/stores/notificationStore";
import type { ThemeMode } from "$src/lib/theme/themeModel";
import {
  applyTheme,
  applyThemeMode as applyThemeModeRuntime,
  applyTypography,
  ensureSystemThemeModeListener,
  setActiveThemeMode
} from "$src/lib/theme/themeRuntime";
import {
  DEFAULT_ZOOM_LEVEL,
  MAX_ZOOM_LEVEL,
  MIN_ZOOM_LEVEL,
  SAVE_DEBOUNCE_MS,
  SAVED_STATUS_TIMEOUT_MS
} from "$src/lib/utils/constants";
import { debounce } from "$src/lib/utils/debounce";
import {
  GetAllThemes,
  GetConfiguration,
  SetActiveTheme,
  SetDefaultFontFamily,
  SetMonoFontFamily,
  UpdateConfiguration
} from "$wails/go/main/App";
import { configuration, theme } from "$wails/go/models";

type TypographySettingsSnapshot = {
  defaultFontFamily: string;
  monoFontFamily: string;
};

function normalizeZoomLevel(level: number | null | undefined): number {
  if (!Number.isFinite(level)) {
    return DEFAULT_ZOOM_LEVEL;
  }

  return Math.min(MAX_ZOOM_LEVEL, Math.max(MIN_ZOOM_LEVEL, Number(level)));
}

function createEmptyConfig(): configuration.Configuration {
  const general = { ...new configuration.GeneralSettings() };
  general.defaultFontFamily ??= "";
  general.monoFontFamily ??= "";
  general.zoomLevel = normalizeZoomLevel(general.zoomLevel);

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
      ...(source.general || {}),
      defaultFontFamily: sanitizeFontFamily(source.general?.defaultFontFamily),
      monoFontFamily: sanitizeFontFamily(source.general?.monoFontFamily),
      zoomLevel: normalizeZoomLevel(source.general?.zoomLevel)
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
      includePrereleaseUpdates: cfg.general?.includePrereleaseUpdates ?? false,
      debugMode: cfg.general?.debugMode ?? false,
      dayTheme: cfg.general?.dayTheme ?? "",
      nightTheme: cfg.general?.nightTheme ?? "",
      selectedEnvironment: cfg.general?.selectedEnvironment ?? "",
      zoomLevel: normalizeZoomLevel(cfg.general?.zoomLevel)
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

function getTypographySnapshot(
  cfg: configuration.Configuration = configurationStoreState.config
): TypographySettingsSnapshot {
  return {
    defaultFontFamily: sanitizeFontFamily(cfg.general?.defaultFontFamily),
    monoFontFamily: sanitizeFontFamily(cfg.general?.monoFontFamily)
  };
}

function getTypographySignature(snapshot: TypographySettingsSnapshot): string {
  return JSON.stringify(snapshot);
}

export interface ConfigurationStoreState {
  config: configuration.Configuration;
  allThemes: theme.Theme[];
  appliedThemeMode: ThemeMode;
  initialized: boolean;
  saveStatus: "idle" | "saving" | "saved" | "error";
}

const prefersDark = window.matchMedia("(prefers-color-scheme: dark)");

let systemIsDark = $state(prefersDark.matches);

prefersDark.addEventListener("change", (e) => {
  systemIsDark = e.matches;
});

export const configurationStoreState: ConfigurationStoreState = $state({
  config: createEmptyConfig(),
  allThemes: [],
  appliedThemeMode: "dark",
  initialized: false,
  saveStatus: "idle"
});

let isHydratingConfig = false;
let lastPersistedSignature: string | null = null;
let lastPersistedDefaultFontFamily: string | null = null;
let lastPersistedMonoFontFamily: string | null = null;
let typographyPersistenceChain: Promise<void> = Promise.resolve();

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

function getConfigSnapshotForPersistence(): configuration.Configuration {
  const snapshot = getConfigSnapshot();
  snapshot.general.defaultFontFamily = lastPersistedDefaultFontFamily ?? "";
  snapshot.general.monoFontFamily = lastPersistedMonoFontFamily ?? "";
  return snapshot;
}

async function persistConfig() {
  if (!configurationStoreState.initialized || isHydratingConfig) return;

  const signature = getPersistenceSignature(configurationStoreState.config);
  if (signature === lastPersistedSignature) return;

  configurationStoreState.saveStatus = "saving";
  try {
    await UpdateConfiguration(getConfigSnapshotForPersistence());
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

function enqueueTypographyPersistence(task: () => Promise<void>) {
  const run = typographyPersistenceChain.then(task, task);
  typographyPersistenceChain = run.catch(() => undefined);
  return run;
}

function applyTypographySnapshot(snapshot: TypographySettingsSnapshot) {
  configurationStoreState.config.general.defaultFontFamily = snapshot.defaultFontFamily;
  configurationStoreState.config.general.monoFontFamily = snapshot.monoFontFamily;
  applyTypography(configurationStoreState.config.general);
}

async function persistDefaultFontFamily(fontFamily: string, previousFontFamily: string) {
  if (!configurationStoreState.initialized || isHydratingConfig) return;
  if (fontFamily === lastPersistedDefaultFontFamily) return;

  await enqueueTypographyPersistence(async () => {
    if (fontFamily === lastPersistedDefaultFontFamily) return;

    try {
      await SetDefaultFontFamily(fontFamily);
      lastPersistedDefaultFontFamily = fontFamily;
    } catch (error) {
      if (
        sanitizeFontFamily(configurationStoreState.config.general.defaultFontFamily) === fontFamily
      ) {
        applyTypographySnapshot({
          ...getTypographySnapshot(),
          defaultFontFamily: previousFontFamily
        });
      }
      notifications.error("Failed to set default font family", String(error));
      throw error;
    }
  });
}

async function persistMonoFontFamily(fontFamily: string, previousFontFamily: string) {
  if (!configurationStoreState.initialized || isHydratingConfig) return;
  if (fontFamily === lastPersistedMonoFontFamily) return;

  await enqueueTypographyPersistence(async () => {
    if (fontFamily === lastPersistedMonoFontFamily) return;

    try {
      await SetMonoFontFamily(fontFamily);
      lastPersistedMonoFontFamily = fontFamily;
    } catch (error) {
      if (
        sanitizeFontFamily(configurationStoreState.config.general.monoFontFamily) === fontFamily
      ) {
        applyTypographySnapshot({
          ...getTypographySnapshot(),
          monoFontFamily: previousFontFamily
        });
      }
      notifications.error("Failed to set monospace font family", String(error));
      throw error;
    }
  });
}

async function persistTypographyReset(
  nextSnapshot: TypographySettingsSnapshot,
  previousSnapshot: TypographySettingsSnapshot
) {
  if (!configurationStoreState.initialized || isHydratingConfig) return;

  const nextSignature = getTypographySignature(nextSnapshot);
  const previousSignature = getTypographySignature(previousSnapshot);
  if (nextSignature === previousSignature) return;

  await enqueueTypographyPersistence(async () => {
    try {
      await Promise.all([
        SetDefaultFontFamily(nextSnapshot.defaultFontFamily),
        SetMonoFontFamily(nextSnapshot.monoFontFamily)
      ]);
      lastPersistedDefaultFontFamily = nextSnapshot.defaultFontFamily;
      lastPersistedMonoFontFamily = nextSnapshot.monoFontFamily;
    } catch (error) {
      if (getTypographySignature(getTypographySnapshot()) === nextSignature) {
        applyTypographySnapshot(previousSnapshot);
      }
      notifications.error("Failed to reset typography", String(error));
      throw error;
    }
  });
}

function resolveAppliedMode(themeMode: ThemeMode) {
  if (themeMode === "system") return systemIsDark ? "dark" : "light";
  return themeMode;
}

export function saveConfig() {
  debouncedPersistConfig();
}

export function setConfigZoomLevel(level: number) {
  configurationStoreState.config.general.zoomLevel = normalizeZoomLevel(level);
}

export const configurationStore = {
  async init() {
    isHydratingConfig = true;
    try {
      const [cfg, themes] = await Promise.all([GetConfiguration(), GetAllThemes()]);
      const normalizedConfig = normalizeConfig(cfg);
      const normalizedThemes = (themes || []).map((t) => new theme.Theme(t));
      const typographySnapshot = getTypographySnapshot(normalizedConfig);

      configurationStoreState.config = normalizedConfig;
      configurationStoreState.allThemes = normalizedThemes;
      configurationStoreState.initialized = true;
      configurationStoreState.appliedThemeMode = resolveAppliedMode(
        (normalizedConfig.general.themeMode as ThemeMode) ?? "system"
      );
      configurationStoreState.saveStatus = "idle";
      lastPersistedSignature = getPersistenceSignature(normalizedConfig);
      lastPersistedDefaultFontFamily = typographySnapshot.defaultFontFamily;
      lastPersistedMonoFontFamily = typographySnapshot.monoFontFamily;

      const mode = normalizedConfig.general?.themeMode as ThemeMode;
      if (mode) {
        setActiveThemeMode(mode);
        applyThemeModeRuntime(mode);
      }
      applyActiveTheme();
      applyTypography(normalizedConfig.general);
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
    configurationStoreState.config.general.themeMode = mode;
    configurationStoreState.appliedThemeMode = resolveAppliedMode(mode);
  },

  async changeDefaultFontFamily(family: string) {
    const previousSnapshot = getTypographySnapshot();
    const nextSnapshot = {
      ...previousSnapshot,
      defaultFontFamily: sanitizeFontFamily(family)
    };

    applyTypographySnapshot(nextSnapshot);
    await persistDefaultFontFamily(
      nextSnapshot.defaultFontFamily,
      previousSnapshot.defaultFontFamily
    );
  },

  async changeMonoFontFamily(family: string) {
    const previousSnapshot = getTypographySnapshot();
    const nextSnapshot = {
      ...previousSnapshot,
      monoFontFamily: sanitizeFontFamily(family)
    };

    applyTypographySnapshot(nextSnapshot);
    await persistMonoFontFamily(nextSnapshot.monoFontFamily, previousSnapshot.monoFontFamily);
  },

  async resetTypography() {
    const previousSnapshot = getTypographySnapshot();
    const nextSnapshot: TypographySettingsSnapshot = {
      defaultFontFamily: "",
      monoFontFamily: ""
    };

    applyTypographySnapshot(nextSnapshot);
    await persistTypographyReset(nextSnapshot, previousSnapshot);
  }
};

ensureSystemThemeModeListener();
