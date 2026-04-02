/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: GPL-3.0-only
 */

import { THEME_PRESETS, type ThemeMode, type ThemeSeeds } from "$src/lib/theme/themeModel";
import type { theme } from "$wails/go/models";

let activeThemeMode: ThemeMode;
let mediaQueryCleanup: (() => void) | null = null;

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

export function resolveThemeSeeds(t: theme.Theme): ThemeSeeds {
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

export function applyTheme(themeDef: theme.Theme) {
  const root = document.documentElement;
  const seeds = resolveThemeSeeds(themeDef);

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

export function applyThemeMode(mode: ThemeMode) {
  const root = document.documentElement;

  if (mode === "dark") {
    root.classList.add("dark");
    return;
  }

  if (mode === "light") {
    root.classList.remove("dark");
    return;
  }

  if (window.matchMedia("(prefers-color-scheme: dark)").matches) {
    root.classList.add("dark");
  } else {
    root.classList.remove("dark");
  }
}

export function setActiveThemeMode(mode: ThemeMode) {
  activeThemeMode = mode;
}

export function getCurrentThemeMode() {
  return activeThemeMode;
}

export function ensureSystemThemeModeListener(options?: { onSystemChange?: () => void }) {
  if (mediaQueryCleanup || typeof window === "undefined") return;

  const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
  const listener = () => {
    if (activeThemeMode === "system") {
      applyThemeMode("system");
      options?.onSystemChange?.();
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
