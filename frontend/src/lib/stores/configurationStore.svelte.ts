/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: GPL-3.0-only
 */

import { notifications } from "$src/lib/stores/notificationStore";
import { THEME_PRESETS, type ThemeMode, type ThemeSeeds } from "$src/lib/theme/themeModel";
import {
  GetAllThemes,
  GetConfiguration,
  SetActiveTheme,
  UpdateConfiguration
} from "$wails/go/main/App";
import { configuration, theme } from "$wails/go/models";

function createEmptyConfig() {
  const cfg = new configuration.Configuration();
  cfg.general = new configuration.GeneralSettings();
  cfg.request = new configuration.RequestSettings();
  cfg.customThemes = [] as theme.Theme[];
  return cfg;
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
  // Mode is not read from the theme — it is applied separately via general.themeMode.
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

// --- $state-based store ---

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

// --- Helper getters ---

export function getConfigSnapshot(): configuration.Configuration {
  return JSON.parse(JSON.stringify(configurationStoreState.config));
}

export function getActiveTheme(): theme.Theme | null {
  const { config, allThemes } = configurationStoreState;
  if (!config?.general?.activeTheme || allThemes.length === 0) {
    return null;
  }
  return allThemes.find((t) => t.id === config.general.activeTheme) || null;
}

// --- Persistence helpers ---

import { debounce } from "$src/lib/utils/debounce";

const SAVE_DEBOUNCE_MS = 800;
let saveTimer: ReturnType<typeof setTimeout> | null = null;

async function persistConfig() {
  const cfg = configurationStoreState.config;
  if (!cfg) return;

  configurationStoreState.saveStatus = "saving";
  try {
    await UpdateConfiguration(cfg);
    configurationStoreState.saveStatus = "saved";
    setTimeout(() => (configurationStoreState.saveStatus = "idle"), 600);
  } catch (err) {
    configurationStoreState.saveStatus = "error";
    notifications.error("Could not save settings", String(err));
  }
}

const debouncedPersistConfig = debounce(persistConfig, SAVE_DEBOUNCE_MS);

export async function saveConfig(immediate = false) {
  if (saveTimer) {
    clearTimeout(saveTimer);
    saveTimer = null;
  }
  if (immediate) {
    await persistConfig();
    return;
  }
  debouncedPersistConfig();
}

export function scheduleConfigSave(isTextInput = true) {
  if (isTextInput) {
    debouncedPersistConfig();
  } else {
    void saveConfig(true);
  }
}

// --- Imperative API ---

function applyActiveTheme() {
  const theme = getActiveTheme();
  if (theme) {
    applyThemeToDom(theme);
  }
}

export const configurationStore = {
  async init() {
    try {
      const [cfg, themes] = await Promise.all([GetConfiguration(), GetAllThemes()]);
      const normalized = new configuration.Configuration(cfg);
      if (!normalized.general) normalized.general = new configuration.GeneralSettings();
      if (!normalized.request) normalized.request = new configuration.RequestSettings();
      if (!normalized.customThemes) normalized.customThemes = [] as theme.Theme[];

      const normalizedThemes = (themes || []).map((t) => new theme.Theme(t));

      // Direct state mutation
      configurationStoreState.config = normalized;
      configurationStoreState.allThemes = normalizedThemes;
      configurationStoreState.initialized = true;

      // Apply theme immediately after load
      applyActiveTheme();

      // Apply mode
      const mode = normalized.general?.themeMode;
      if (mode) {
        activeThemeMode = mode;
        applyThemeMode(mode);
      }
    } catch (error) {
      notifications.error("Failed to initialize configuration", String(error), true);
    }
  },

  async save(newConfig: configuration.Configuration) {
    try {
      await UpdateConfiguration(newConfig);
      configurationStoreState.config = newConfig;
    } catch (error) {
      notifications.error("Failed to save configuration", String(error));
      throw error;
    }
  },

  async changeTheme(themeId: string) {
    try {
      if (!configurationStoreState.config.general) {
        configurationStoreState.config.general = new configuration.GeneralSettings();
      }
      configurationStoreState.config.general.activeTheme = themeId;

      await SetActiveTheme(themeId);

      // Apply theme to DOM
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

// --- Initialize system theme listener ---
ensureSystemThemeSyncListener();
