/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: GPL-3.0-only
 */

import { notifications } from "$src/lib/stores/notificationStore";
import { THEME_PRESETS, type ThemeMode, type ThemeSeeds } from "$src/lib/theme/themeModel";
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

function hexToRgb(hex: string): [number, number, number] {
  const sanitized = hex.replace("#", "").trim();
  const full =
    sanitized.length === 3
      ? sanitized
          .split("")
          .map((c) => c + c)
          .join("")
      : sanitized;
  const int = Number.parseInt(full, 16);
  return [(int >> 16) & 255, (int >> 8) & 255, int & 255];
}

function rgbToHex([r, g, b]: [number, number, number]): string {
  return `#${[r, g, b]
    .map((v) =>
      Math.max(0, Math.min(255, Math.round(v)))
        .toString(16)
        .padStart(2, "0")
    )
    .join("")}`;
}

function mix(a: string, b: string, ratio: number): string {
  const [ar, ag, ab] = hexToRgb(a);
  const [br, bg, bb] = hexToRgb(b);
  const r = ar + (br - ar) * ratio;
  const g = ag + (bg - ag) * ratio;
  const b2 = ab + (bb - ab) * ratio;
  return rgbToHex([r, g, b2]);
}

function generateScale(seed: string) {
  const white = "#ffffff";
  const black = "#000000";

  return {
    50: mix(seed, white, 0.94),
    100: mix(seed, white, 0.86),
    200: mix(seed, white, 0.72),
    300: mix(seed, white, 0.54),
    400: mix(seed, white, 0.32),
    500: seed,
    600: mix(seed, black, 0.14),
    700: mix(seed, black, 0.28),
    800: mix(seed, black, 0.42),
    900: mix(seed, black, 0.56)
  } as const;
}

function resolveSeeds(t: theme.Theme): ThemeSeeds {
  if (t?.config?.type === "preset") {
    const preset = THEME_PRESETS.find((p) => p.id === t.config.presetId);
    if (preset) return preset.seeds;
  }

  if (t?.config?.seeds) {
    return {
      primary: t.config.seeds.primary,
      success: t.config.seeds.success,
      warning: t.config.seeds.warning,
      danger: t.config.seeds.danger,
      neutral: t.config.seeds.neutral,
      surface: t.config.seeds.surface
    };
  }

  return THEME_PRESETS[0].seeds;
}

function applyThemeMode(mode: ThemeMode | string | undefined) {
  const root = document.documentElement;

  if (mode === "dark") {
    root.classList.add("dark");
    return;
  }

  if (mode === "light") {
    root.classList.remove("dark");
    return;
  }

  // system/default
  if (window.matchMedia("(prefers-color-scheme: dark)").matches) {
    root.classList.add("dark");
  } else {
    root.classList.remove("dark");
  }
}

// Global display mode — driven by general.themeMode, not per-theme config.
let activeThemeMode: ThemeMode | string | undefined;

function applyThemeToDom(t: theme.Theme) {
  const root = document.documentElement;
  const seeds = resolveSeeds(t);

  const families = {
    primary: generateScale(seeds.primary),
    success: generateScale(seeds.success),
    warning: generateScale(seeds.warning),
    danger: generateScale(seeds.danger),
    neutral: generateScale(seeds.neutral)
  };

  (Object.keys(families) as (keyof typeof families)[]).forEach((family) => {
    const scale = families[family] as Record<string, string>;
    Object.entries(scale).forEach(([step, value]) => {
      root.style.setProperty(`--color-${family}-${step}`, value);
    });
  });

  if (seeds.surface) {
    root.style.setProperty("--color-surface", seeds.surface);
  } else {
    root.style.removeProperty("--color-surface");
  }
}

let mediaQueryCleanup: (() => void) | null = null;

function ensureSystemThemeSyncListener() {
  if (mediaQueryCleanup || typeof window === "undefined") return;

  const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
  const listener = () => {
    if (activeThemeMode === "system") {
      applyThemeMode("system");
    }
  };

  mediaQuery.addEventListener("change", listener);
  mediaQueryCleanup = () => mediaQuery.removeEventListener("change", listener);

  if (import.meta.hot) {
    import.meta.hot.dispose(() => {
      mediaQueryCleanup?.();
      mediaQueryCleanup = null;
    });
  }
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
    applyThemeToDom(currentTheme);
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

export function saveConfig() {debounce(() => {
  void persistConfig();
}, SAVE_DEBOUNCE_MS);
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

      const mode = normalizedConfig.general?.themeMode;
      if (mode) {
        activeThemeMode = mode;
        applyThemeMode(mode);
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

  applyThemeMode(mode: string) {
    activeThemeMode = mode;
    applyThemeMode(mode);
  }
};

ensureSystemThemeSyncListener();
