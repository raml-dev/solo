/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: AGPL-3.0-only
 */

import {
  WindowFullscreen,
  WindowGetSize,
  WindowIsFullscreen,
  WindowIsMaximised,
  WindowMaximise,
  WindowSetSize,
  WindowShow
} from "$wails/runtime/runtime";
interface WindowState {
  width: number;
  height: number;
  maximised: boolean;
  fullscreen: boolean;
}

const WINDOW_STATE_KEY = "windowState";
const DEFAULT_WIDTH = 1024;
const DEFAULT_HEIGHT = 768;

/**
 *  windowState holds the window status:
 * - dimensions (when not maximized/fullscreen, used to restore to those when unmaximizing/toggling fullscreen)
 * - whether it's fullscreen
 * - whether it's maximized
 */
const windowState = $state(loadWindowState());

function registerWindowStateTracking(): void {
  window.addEventListener("resize", trackWindowState);
}

async function trackWindowState(): Promise<void> {
  try {
    const [maximised, fullscreen, { w, h }] = await Promise.all([
      WindowIsMaximised(),
      WindowIsFullscreen(),
      WindowGetSize()
    ]);
    windowState.maximised = maximised;
    windowState.fullscreen = fullscreen;
    if (!maximised && !fullscreen) {
      windowState.width = w;
      windowState.height = h;
    }
    saveWindowState();
  } catch (e) {
    console.log(e);
  }
}

function loadWindowState(): WindowState {
  try {
    const saved = localStorage.getItem(WINDOW_STATE_KEY);
    if (saved) return JSON.parse(saved) as WindowState;
  } catch {
    // ignore malformed data
  }
  return {
    width: DEFAULT_WIDTH,
    height: DEFAULT_HEIGHT,
    maximised: false,
    fullscreen: false
  };
}

export function saveWindowState() {
  localStorage.setItem(WINDOW_STATE_KEY, JSON.stringify(windowState));
}

export function initWindowDimensions() {
  const clampedWidth = Math.min(windowState.width, screen.width);
  const clampedHeight = Math.min(windowState.height, screen.height);

  WindowSetSize(clampedWidth, clampedHeight);

  if (windowState.fullscreen) WindowFullscreen();
  else if (windowState.maximised) WindowMaximise();

  WindowShow();

  registerWindowStateTracking();
  return () => {
    window.removeEventListener("resize", trackWindowState);
  };
}
