<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { SaveCustomTheme } from '../../../wailsjs/go/main/App';
  import { changeTheme } from '../stores/themeStore';
  import type { Theme } from '../stores/themeStore';
  import Button from './base/Button.svelte';

  export let baseTheme: Theme | null = null;

  const dispatch = createEventDispatcher();

  let themeName = '';
  let colors = {
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
  };

  $: if (baseTheme) {
    themeName = `${baseTheme.name}-custom`;
    colors = { ...baseTheme.colors };
  }

  const colorLabels = {
    'primary': 'Primary Color',
    'primary-dark': 'Primary Dark',
    'success': 'Success',
    'warning': 'Warning',
    'danger': 'Danger',
    'info': 'Info',
    'bg-primary': 'Background Primary',
    'bg-secondary': 'Background Secondary',
    'bg-tertiary': 'Background Tertiary',
    'border': 'Border',
    'border-dark': 'Border Dark',
    'text': 'Text',
    'text-muted': 'Text Muted',
    'text-light': 'Text Light',
  };

  async function saveTheme() {
    if (!themeName.trim()) {
      alert('Please enter a theme name');
      return;
    }

    const theme: Theme = {
      name: themeName,
      colors: colors,
    };

    try {
      await SaveCustomTheme(theme);
      await changeTheme(themeName);
      dispatch('saved', theme);
      dispatch('close');
    } catch (error) {
      console.error('Failed to save theme:', error);
      alert('Failed to save theme');
    }
  }

  function cancel() {
    dispatch('close');
  }

  function previewColor() {
    // Apply colors temporarily to preview
    const root = document.documentElement;
    Object.entries(colors).forEach(([key, value]) => {
      root.style.setProperty(`--${key}`, value);
    });
  }

  $: colors && previewColor();
</script>

<div class="customizer">
  <div class="customizer-header">
    <h3 class="text-lg font-semibold">Customize Theme</h3>
    <Button variant="primary" on:click={cancel}>×</Button>
  </div>

  <div class="customizer-content">
    <!-- Theme Name -->
    <div class="form-group">
      <label for="theme-name" class="label">Theme Name</label>
      <input
        id="theme-name"
        type="text"
        class="input"
        bind:value={themeName}
        placeholder="My Custom Theme"
      />
    </div>

    <!-- Color Groups -->
    <div class="color-sections">
      <!-- Primary Colors -->
      <div class="color-section">
        <h4 class="section-title">Primary Colors</h4>
        <div class="color-grid">
          {#each ['primary', 'primary-dark', 'success', 'warning', 'danger', 'info'] as colorKey}
            <div class="color-input-group">
              <label for={colorKey} class="color-label">{colorLabels[colorKey]}</label>
              <div class="color-input-wrapper">
                <input
                  id={colorKey}
                  type="color"
                  bind:value={colors[colorKey]}
                  class="color-picker"
                />
                <input
                  type="text"
                  bind:value={colors[colorKey]}
                  class="input input-sm"
                  placeholder="#000000"
                />
              </div>
            </div>
          {/each}
        </div>
      </div>

      <!-- Background Colors -->
      <div class="color-section">
        <h4 class="section-title">Background Colors</h4>
        <div class="color-grid">
          {#each ['bg-primary', 'bg-secondary', 'bg-tertiary'] as colorKey}
            <div class="color-input-group">
              <label for={colorKey} class="color-label">{colorLabels[colorKey]}</label>
              <div class="color-input-wrapper">
                <input
                  id={colorKey}
                  type="color"
                  bind:value={colors[colorKey]}
                  class="color-picker"
                />
                <input
                  type="text"
                  bind:value={colors[colorKey]}
                  class="input input-sm"
                  placeholder="#000000"
                />
              </div>
            </div>
          {/each}
        </div>
      </div>

      <!-- Border Colors -->
      <div class="color-section">
        <h4 class="section-title">Border Colors</h4>
        <div class="color-grid">
          {#each ['border', 'border-dark'] as colorKey}
            <div class="color-input-group">
              <label for={colorKey} class="color-label">{colorLabels[colorKey]}</label>
              <div class="color-input-wrapper">
                <input
                  id={colorKey}
                  type="color"
                  bind:value={colors[colorKey]}
                  class="color-picker"
                />
                <input
                  type="text"
                  bind:value={colors[colorKey]}
                  class="input input-sm"
                  placeholder="#000000"
                />
              </div>
            </div>
          {/each}
        </div>
      </div>

      <!-- Text Colors -->
      <div class="color-section">
        <h4 class="section-title">Text Colors</h4>
        <div class="color-grid">
          {#each ['text', 'text-muted', 'text-light'] as colorKey}
            <div class="color-input-group">
              <label for={colorKey} class="color-label">{colorLabels[colorKey]}</label>
              <div class="color-input-wrapper">
                <input
                  id={colorKey}
                  type="color"
                  bind:value={colors[colorKey]}
                  class="color-picker"
                />
                <input
                  type="text"
                  bind:value={colors[colorKey]}
                  class="input input-sm"
                  placeholder="#000000"
                />
              </div>
            </div>
          {/each}
        </div>
      </div>
    </div>
  </div>

  <div class="customizer-footer">
    <Button variant="secondary" on:click={cancel}>Cancel</Button>
    <Button variant="primary" on:click={saveTheme}>Save Theme</Button>
  </div>
</div>

<style>
  .customizer {
    display: flex;
    flex-direction: column;
    height: 100%;
    max-height: 80vh;
  }

  .customizer-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: var(--space-lg);
    border-bottom: 1px solid var(--border);
  }

  .btn-close {
    width: 32px;
    height: 32px;
    border: none;
    background: var(--bg-tertiary);
    border-radius: var(--radius-md);
    font-size: var(--font-size-2xl);
    line-height: 1;
    cursor: pointer;
    color: var(--text-muted);
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all var(--transition-fast);
    font-family: var(--font-sans);
  }

  .btn-close:hover {
    background: var(--border);
    color: var(--text);
  }

  .customizer-content {
    flex: 1;
    overflow-y: auto;
    padding: var(--space-lg);
  }

  .form-group {
    margin-bottom: var(--space-xl);
  }

  .label {
    display: block;
    margin-bottom: var(--space-sm);
    font-weight: var(--font-weight-medium);
    font-size: var(--font-size-sm);
    color: var(--text);
  }

  .color-sections {
    display: flex;
    flex-direction: column;
    gap: var(--space-xl);
  }

  .color-section {
    display: flex;
    flex-direction: column;
    gap: var(--space-md);
  }

  .section-title {
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-semibold);
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .color-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
    gap: var(--space-md);
  }

  .color-input-group {
    display: flex;
    flex-direction: column;
    gap: var(--space-xs);
  }

  .color-label {
    font-size: var(--font-size-xs);
    font-weight: var(--font-weight-medium);
    color: var(--text);
  }

  .color-input-wrapper {
    display: flex;
    gap: var(--space-sm);
    align-items: center;
  }

  .color-picker {
    width: 50px;
    height: 38px;
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    cursor: pointer;
    background: none;
  }

  .customizer-footer {
    display: flex;
    justify-content: flex-end;
    gap: var(--space-sm);
    padding: var(--space-lg);
    border-top: 1px solid var(--border);
    background: var(--bg-secondary);
  }
</style>