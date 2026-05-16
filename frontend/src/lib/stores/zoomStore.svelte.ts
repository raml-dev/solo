/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: AGPL-3.0-only
 */

import { setConfigZoomLevel } from "$src/lib/stores/configurationStore.svelte";
import { notifications } from "$src/lib/stores/notificationStore";
import {
  DEFAULT_ZOOM_LEVEL,
  MAX_ZOOM_LEVEL,
  MIN_ZOOM_LEVEL,
  ZOOM_STEP
} from "$src/lib/utils/constants";
import { SetZoomLevel } from "$wails/go/main/App";

export const zoomState = $state({
  level: DEFAULT_ZOOM_LEVEL,
  initialized: false
});

let notificationId = $state("");
let hideTimer: ReturnType<typeof setTimeout> | null = null;

function clampZoomLevel(level: number | null | undefined): number {
  const numericLevel = Number(level);
  if (!Number.isFinite(numericLevel)) {
    return DEFAULT_ZOOM_LEVEL;
  }

  return Math.min(MAX_ZOOM_LEVEL, Math.max(MIN_ZOOM_LEVEL, Number(numericLevel.toFixed(2))));
}

function applyZoomLevel(level: number, notify: boolean = false) {
  const nextLevel = clampZoomLevel(level);
  zoomState.level = nextLevel;
  document.documentElement.style.zoom = String(nextLevel);

  if (!notify) {
    return;
  }

  const message = `${getZoomPercent(nextLevel)}%`;
  if (notificationId) {
    notifications.updateNotificationById(notificationId, "Zoom:", message);
  } else {
    notificationId = notifications.notify("info", "Zoom:", message, true);
  }

  if (hideTimer) {
    clearTimeout(hideTimer);
  }

  hideTimer = setTimeout(() => {
    notifications.dismiss(notificationId);
    notificationId = "";
  }, 2000);
}

async function persistAndSyncZoom(nextLevel: number, previousLevel: number) {
  try {
    await SetZoomLevel(nextLevel);
    setConfigZoomLevel(nextLevel);
  } catch (error) {
    applyZoomLevel(previousLevel, false);
    setConfigZoomLevel(previousLevel);
    notifications.error("Failed to set zoom level", String(error));
    throw error;
  }
}

function handleKeydown(event: KeyboardEvent) {
  const hasModifier = event.metaKey || event.ctrlKey;
  if (!hasModifier) {
    return;
  }

  if (event.key === "=" || event.key === "+") {
    event.preventDefault();
    void increaseZoom(true);
    return;
  }

  if (event.key === "-") {
    event.preventDefault();
    void decreaseZoom(true);
    return;
  }

  if (event.key === "0") {
    event.preventDefault();
    void resetZoom(true);
  }
}

export function getZoomPercent(level: number): number {
  return Math.round(clampZoomLevel(level) * 100);
}

export function initZoom(level: number = DEFAULT_ZOOM_LEVEL) {
  applyZoomLevel(level, false);
  setConfigZoomLevel(level);
  zoomState.initialized = true;
}

export async function setZoomLevel(level: number, notify: boolean = false) {
  const previousLevel = zoomState.level;
  const nextLevel = clampZoomLevel(level);

  if (zoomState.initialized && nextLevel === previousLevel) {
    return;
  }

  applyZoomLevel(nextLevel, notify);

  if (!zoomState.initialized) {
    zoomState.initialized = true;
    return;
  }

  await persistAndSyncZoom(nextLevel, previousLevel);
}

export async function increaseZoom(notify: boolean = false) {
  await setZoomLevel(zoomState.level + ZOOM_STEP, notify);
}

export async function decreaseZoom(notify: boolean = false) {
  await setZoomLevel(zoomState.level - ZOOM_STEP, notify);
}

export async function resetZoom(notify: boolean = false) {
  await setZoomLevel(DEFAULT_ZOOM_LEVEL, notify);
}

export function registerZoomShortcuts() {
  window.addEventListener("keydown", handleKeydown);
  return () => window.removeEventListener("keydown", handleKeydown);
}
