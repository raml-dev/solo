<script lang="ts">
  import { onMount } from "svelte";
  import { currentTheme, changeTheme } from "../stores/themeStore";
  import { GetPredefinedThemes, GetCustomThemes } from "../../../wailsjs/go/main/App";
  import type { Theme } from "../stores/themeStore";
  import Button from "./base/Button.svelte";
  import Modal from "./base/Modal.svelte";

  let predefinedThemes: Theme[] = [];
  let customThemes: Theme[] = [];
  let showCustomizer = false;

  onMount(async () => {
    await loadThemes();
  });

  async function loadThemes() {
    try {
      predefinedThemes = await GetPredefinedThemes();
      customThemes = await GetCustomThemes();
    } catch (error) {
      console.error("Failed to load themes:", error);
    }
  }

  async function handleThemeChange(themeName: string) {
    try {
      await changeTheme(themeName);
    } catch (error) {
      console.error("Failed to change theme:", error);
    }
  }

  function getThemePreview(theme: Theme): string {
    return theme.colors["bg-primary"] || "#ffffff";
  }

  function openCustomizer() {
    showCustomizer = true;
  }
</script>

<div class="theme-selector">
  <h3 class="text-base font-semibold">Theme</h3>

  <!-- Predefined Themes -->
  <div class="theme-section">
    <h4 class="text-sm font-medium text-muted">Predefined Themes</h4>
    <div class="theme-grid">
      {#each predefinedThemes as theme}
        <button
          class="theme-card"
          class:active={$currentTheme?.name === theme.name}
          on:click={() => handleThemeChange(theme.name)}
        >
          <div class="theme-preview" style="background: {getThemePreview(theme)}" />
          <div class="theme-colors">
            <span class="color-dot" style="background: {theme.colors.primary}" />
            <span class="color-dot" style="background: {theme.colors.success}" />
            <span class="color-dot" style="background: {theme.colors.warning}" />
            <span class="color-dot" style="background: {theme.colors.danger}" />
          </div>
          <span class="theme-name">{theme.name}</span>
        </button>
      {/each}
    </div>
  </div>

  <!-- Custom Themes -->
  {#if customThemes.length > 0}
    <div class="theme-section">
      <h4 class="text-sm font-medium text-muted">Custom Themes</h4>
      <div class="theme-grid">
        {#each customThemes as theme}
          <button
            class="theme-card"
            class:active={$currentTheme?.name === theme.name}
            on:click={() => handleThemeChange(theme.name)}
          >
            <div class="theme-preview" style="background: {getThemePreview(theme)}" />
            <div class="theme-colors">
              <span class="color-dot" style="background: {theme.colors.primary}" />
              <span class="color-dot" style="background: {theme.colors.success}" />
              <span class="color-dot" style="background: {theme.colors.warning}" />
              <span class="color-dot" style="background: {theme.colors.danger}" />
            </div>
            <span class="theme-name">{theme.name}</span>
          </button>
        {/each}
      </div>
    </div>
  {/if}

  <!-- Create Custom Theme Button -->
  <Button variant="secondary" on:click={openCustomizer}>Create Custom Theme</Button>

  {#if showCustomizer}
    <Modal toggleFn={() => (showCustomizer = false)}>
      <p class="text-muted">Theme customizer will go here</p>
    </Modal>
  {/if}
</div>

<style>
  .theme-selector {
    padding: var(--space-lg);
  }

  .theme-section {
    margin: var(--space-lg) 0;
  }

  .theme-section h4 {
    margin-bottom: var(--space-md);
  }

  .theme-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
    gap: var(--space-md);
    margin-bottom: var(--space-lg);
  }

  .theme-card {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: var(--space-sm);
    padding: var(--space-md);
    border: 2px solid var(--border);
    border-radius: var(--radius-lg);
    background: var(--bg-secondary);
    cursor: pointer;
    transition: all var(--transition-base);
    font-family: var(--font-sans);
  }

  .theme-card:hover {
    border-color: var(--border-dark);
    transform: translateY(-2px);
  }

  .theme-card.active {
    border-color: var(--primary);
    background: var(--bg-tertiary);
  }

  .theme-preview {
    width: 100%;
    height: 60px;
    border-radius: var(--radius-md);
    border: 1px solid var(--border);
  }

  .theme-colors {
    display: flex;
    gap: var(--space-xs);
  }

  .color-dot {
    width: 12px;
    height: 12px;
    border-radius: 50%;
    border: 1px solid rgba(0, 0, 0, 0.1);
  }

  .theme-name {
    font-size: var(--font-size-xs);
    font-weight: var(--font-weight-medium);
    text-transform: capitalize;
  }
</style>
