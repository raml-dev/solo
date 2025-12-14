// themeStore.ts
// Place in frontend/src/lib/stores/themeStore.ts
import { writable, derived } from 'svelte/store';
import { GetActiveTheme, GetThemeByName, SetActiveTheme } from '../../../wailsjs/go/main/App';

type ThemeColors = 
  | "primary"
  | "primary-dark"
  | "success"
  | "warning"
  | "danger"
  | "info"
  | "bg-primary"
  | "bg-secondary"
  | "bg-tertiary"
  | "border"
  | "border-dark"
  | "text"
  | "text-muted"
  | "text-light"
export interface Theme {
  name: string;
  colors: Record<ThemeColors, string>;
}

// Create writable store for current theme
export const currentTheme = writable<Theme | null>(null);

// Derived store for checking if theme is loaded
export const themeLoaded = derived(currentTheme, $theme => $theme !== null);

// Initialize theme from backend
export async function initTheme() {
  try {
    const activeThemeName = await GetActiveTheme();
    const theme = await GetThemeByName(activeThemeName);
    currentTheme.set(theme);
    applyTheme(theme);
  } catch (error) {
    console.error('Failed to load theme:', error);
    // Fallback to default theme
    currentTheme.set(getDefaultTheme());
  }
}

// Apply theme to CSS variables
export function applyTheme(theme: Theme) {
  const root = document.documentElement;
  
  Object.entries(theme.colors).forEach(([key, value]) => {
    root.style.setProperty(`--${key}`, value);
  });
}

// Change active theme
export async function changeTheme(themeName: string) {
  try {
    await SetActiveTheme(themeName);
    const theme = await GetThemeByName(themeName);
    currentTheme.set(theme);
    applyTheme(theme);
  } catch (error) {
    console.error('Failed to change theme:', error);
    throw error;
  }
}

// Default theme for fallback
function getDefaultTheme(): Theme {
  return {
    name: 'default-light',
    colors: {
      'primary': '#2563eb',
      'primary-dark': '#1e40af',
      'success': '#10b981',
      'warning': '#f59e0b',
      'danger': '#ef4444',
      'info': '#06b6d4',
      'bg-primary': '#ffffff',
      'bg-secondary': '#f9fafb',
      'bg-tertiary': '#f3f4f6',
      'border': '#e5e7eb',
      'border-dark': '#d1d5db',
      'text': '#111827',
      'text-muted': '#6b7280',
      'text-light': '#9ca3af',
    },
  };
}