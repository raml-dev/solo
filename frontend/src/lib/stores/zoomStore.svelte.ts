import { notifications } from "$src/lib/stores/notificationStore";

const STEP = 0.1;
const MIN = 0.5;
const MAX = 2.0;
const STORAGE_KEY = "ui-zoom";

const zoomState = $state({ level: parseFloat(localStorage.getItem(STORAGE_KEY) ?? "1") });
let notificationId = $state("");
let hideTimer: ReturnType<typeof setTimeout> | null = null;

function apply(notify: boolean = true) {
  document.documentElement.style.zoom = String(zoomState.level);
  localStorage.setItem(STORAGE_KEY, String(zoomState.level));
  if (notify) {
    if (notificationId) {
      notifications.updateNotificationById(
        notificationId,
        "Zoom:",
        Math.floor(zoomState.level * 100) + "%"
      );
    } else {
      notificationId = notifications.notify(
        "info",
        "Zoom:",
        Math.floor(zoomState.level * 100) + "%",
        true
      );
    }
    if (hideTimer) clearTimeout(hideTimer);
    hideTimer = setTimeout(() => {
      notifications.dismiss(notificationId);
      notificationId = "";
    }, 2000);
  }
}

function handleKeydown(e: KeyboardEvent) {
  const mod = e.metaKey || e.ctrlKey;
  if (!mod) return;
  if (e.key === "=" || e.key === "+") {
    e.preventDefault();
    zoomState.level = Math.min(MAX, parseFloat((zoomState.level + STEP).toFixed(2)));
    apply();
  } else if (e.key === "-") {
    e.preventDefault();
    zoomState.level = Math.max(MIN, parseFloat((zoomState.level - STEP).toFixed(2)));
    apply();
  } else if (e.key === "0") {
    e.preventDefault();
    zoomState.level = 1;
    apply();
  }
}

export function initZoom() {
  apply(false); // restore persisted zoom on startup
  window.addEventListener("keydown", handleKeydown);
  return () => window.removeEventListener("keydown", handleKeydown);
}

export { zoomState };
