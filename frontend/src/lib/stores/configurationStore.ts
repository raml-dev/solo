import { writable, derived } from "svelte/store";
import {
  GetConfiguration,
  UpdateConfiguration,
  GetAllThemes,
  SetActiveTheme
} from "../../../wailsjs/go/main/App";
import { configuration } from "../../../wailsjs/go/models";
import type { theme } from "../../../wailsjs/go/models";

// This is the single source of truth for all app configuration.
function createEmptyConfig() {
  const cfg = new configuration.Configuration();
  cfg.general = new configuration.GeneralSettings();
  cfg.request = new configuration.RequestSettings();
  cfg.customThemes = [] as theme.Theme[];
  return cfg;
}

function createConfigurationStore() {
  const config = writable<configuration.Configuration>(createEmptyConfig());
  const allThemes = writable<theme.Theme[]>([]);

  // Derived store to get the full object for the active theme
  const activeTheme = derived([config, allThemes], ([$config, $allThemes]) => {
    if (!$config?.general?.activeTheme || $allThemes.length === 0) {
      return null;
    }
    return $allThemes.find((t) => t.name === $config.general.activeTheme) || null;
  });

  // Apply theme to CSS variables whenever the active theme object changes
  activeTheme.subscribe((theme) => {
    if (theme) {
      applyThemeToDom(theme);
    }
  });

  function applyThemeToDom(theme: theme.Theme) {
    const root = document.documentElement;
    Object.entries(theme.colors).forEach(([key, value]) => {
      root.style.setProperty(`--${key}`, value as string);
    });
  }

  return {
    config: {
      subscribe: config.subscribe
    },
    allThemes: {
      subscribe: allThemes.subscribe
    },
    activeTheme: {
      subscribe: activeTheme.subscribe
    },

    init: async () => {
      try {
        const [cfg, themes] = await Promise.all([GetConfiguration(), GetAllThemes()]);
        const normalized = new configuration.Configuration(cfg);
        if (!normalized.general) normalized.general = new configuration.GeneralSettings();
        if (!normalized.request) normalized.request = new configuration.RequestSettings();
        if (!normalized.customThemes) normalized.customThemes = [] as theme.Theme[];

        allThemes.set(themes);
        config.set(normalized);
      } catch (error) {
        console.error("Failed to initialize configuration and themes:", error);
      }
    },

    save: async (newConfig: configuration.Configuration) => {
      try {
        await UpdateConfiguration(newConfig);
        config.set(newConfig);
      } catch (error) {
        console.error("Failed to save configuration:", error);
        throw error;
      }
    },

    changeTheme: async (themeName: string) => {
      try {
        config.update((c) => {
          if (!c.general) c.general = new configuration.GeneralSettings();
          c.general.activeTheme = themeName;
          return c;
        });
        await SetActiveTheme(themeName);
      } catch (error) {
        console.error("Failed to change theme:", error);
        throw error;
      }
    }
  };
}

export const configurationStore = createConfigurationStore();
