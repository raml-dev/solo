/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: AGPL-3.0-only
 */

// wails events
export const CONTEXT_MENU_OPEN_EVENT = "solo:context-menu-open";

export const DEFAULT_ENV_NAME = "default";
export const EMPTY_TAB_LABEL = "New Request";

// window management
export const WINDOW_DEFAULT_WIDTH = 1024;
export const WINDOW_DEFAULT_HEIGHT = 768;
export const WINDOW_STATE_STORAGE_KEY = "windowState";
export const TAB_STORAGE_KEY = "tabs";

export const ENV_AUTOCOMPLETE_DEFAULT_TRIGGER = "{{";
export const ENV_AUTOCOMPLETE_DEFAULT_MAX_ITEMS = 8;
export const ENV_AUTOCOMPLETE_DEFAULT_INSERT_MODE = "value";

export const DEFAULT_ZOOM_LEVEL = 1;
export const MIN_ZOOM_LEVEL = 0.5;
export const MAX_ZOOM_LEVEL = 2.0;
export const ZOOM_STEP = 0.1;
export const ZOOM_LEVEL_OPTIONS = Array.from(
  { length: Math.round((MAX_ZOOM_LEVEL - MIN_ZOOM_LEVEL) / ZOOM_STEP) + 1 },
  (_, index) => Number((MIN_ZOOM_LEVEL + index * ZOOM_STEP).toFixed(2))
);

export const SAVE_DEBOUNCE_MS = 800;
export const SAVED_STATUS_TIMEOUT_MS = 1500;

export const ENV_TOKEN_REGEX = /\{\{([^{}\r\n]+?)\}\}/g;

export const XML_INDENT = "  ";

export const APP_HISTORY_PANE_MIN_HEIGHT = 120;
export const APP_HISTORY_PANE_MAX_HEIGHT = 700;
export const HISTORY_MAX_ENTRIES = 500;

export const MAX_OPEN_TABS = 15;
export const MAX_REQUESTS_BEFORE_TRIM = 99;
export const REQUEST_RUNNER_MAX_VISIBLE_RESULTS = 50;

export const COLLECTION_OUTLINE_BUTTON_CLASSES =
  "text-neutral-800/70 hover:text-neutral-800 dark:text-neutral-100/70 dark:hover:text-neutral-100";

export const FONT_FAMILY_SELECT_LIST_ROW_HEIGHT = 36;
export const FONT_FAMILY_SELECT_LIST_OVERSCAN = 5;
export const FONT_FAMILY_SELECT_PANEL_GAP = 8;
export const FONT_FAMILY_SELECT_PREFERRED_PANEL_HEIGHT = 360;
