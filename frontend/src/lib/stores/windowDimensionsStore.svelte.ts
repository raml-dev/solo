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
  } catch (e) {
    console.log(e);
  } finally {
    console.log($state.snapshot(windowState));
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

export async function saveWindowState(): Promise<void> {
  await trackWindowState();
  localStorage.setItem(WINDOW_STATE_KEY, JSON.stringify(windowState));
}

export function initWindowDimensions() {
  const clampedWidth = Math.min(windowState.width, screen.width);
  const clampedHeight = Math.min(windowState.height, screen.height);

  console.log(
    JSON.stringify({
      screenw: screen.width,
      screenh: screen.height
    })
  );
  WindowSetSize(clampedWidth, clampedHeight);

  // only restore position in normal state — let the OS place the
  // window when maximised or fullscreen
  if (windowState.fullscreen) WindowFullscreen();
  else if (windowState.maximised) WindowMaximise();

  setTimeout(() => {
    WindowShow();
  }, 150);

  registerWindowStateTracking();
  return () => {
    window.removeEventListener("resize", trackWindowState);
  };
}
