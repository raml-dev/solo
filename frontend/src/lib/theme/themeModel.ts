/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: AGPL-3.0-only
 */

import type { HttpMethod } from "$src/lib/utils/http";

export type ThemeMode = "light" | "dark" | "system";

export type ThemeSeeds = {
  primary: string;
  success: string;
  warning: string;
  danger: string;
  neutral: string;
  surface?: string;
};

export type ThemePresetId = "ocean" | "ember" | "forest" | "violet" | "nord" | "pastel";

export type PresetThemeConfig = {
  type: "preset";
  presetId: ThemePresetId;
  mode: ThemeMode;
};

export type CustomThemeConfig = {
  type: "custom";
  mode: ThemeMode;
  seeds: ThemeSeeds;
};

export type ThemeConfig = PresetThemeConfig | CustomThemeConfig;

export type ThemePresetDefinition = {
  id: ThemePresetId;
  name: string;
  seeds: ThemeSeeds;
};

export const THEME_PRESETS: ThemePresetDefinition[] = [
  {
    id: "ocean",
    name: "Ocean",
    seeds: {
      primary: "#0ea5e9",
      success: "#10b981",
      warning: "#f59e0b",
      danger: "#ef4444",
      neutral: "#52525b"
    }
  },
  {
    id: "ember",
    name: "Ember",
    seeds: {
      primary: "#f97316",
      success: "#22c55e",
      warning: "#f59e0b",
      danger: "#ef4444",
      neutral: "#52525b"
    }
  },
  {
    id: "forest",
    name: "Forest",
    seeds: {
      primary: "#22c55e",
      success: "#16a34a",
      warning: "#eab308",
      danger: "#dc2626",
      neutral: "#52525b"
    }
  },
  {
    id: "violet",
    name: "Violet",
    seeds: {
      primary: "#8b5cf6",
      success: "#10b981",
      warning: "#f59e0b",
      danger: "#ef4444",
      neutral: "#52525b"
    }
  },
  {
    id: "nord",
    name: "Nord",
    seeds: {
      primary: "#5e81ac",
      success: "#a3be8c",
      warning: "#ebcb8b",
      danger: "#bf616a",
      neutral: "#4c566a",
      surface: "#2e3440"
    }
  },
  {
    id: "pastel",
    name: "Pastel",
    seeds: {
      primary: "#7dd3fc",
      success: "#86efac",
      warning: "#fde68a",
      danger: "#fda4af",
      neutral: "#94a3b8",
      surface: "#e2e8f0"
    }
  }
];

export type MethodSemanticFamily = "primary" | "success" | "warning" | "danger" | "neutral";

export const HTTP_METHOD_COLOR_MAP: Record<HttpMethod, MethodSemanticFamily> = {
  GET: "success",
  POST: "primary",
  PUT: "warning",
  PATCH: "warning",
  DELETE: "danger",
  HEAD: "neutral",
  OPTIONS: "neutral"
};
