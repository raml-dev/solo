<script lang="ts">
  import { onMount } from "svelte";
  import MainLayout from "./lib/components/MainLayout.svelte";
  import CollectionList from "./lib/components/CollectionList.svelte";
  import EnvironmentManager from "./lib/components/EnvironmentManager.svelte";
  import HTTPRequestBuilder from "./lib/components/HTTPRequestBuilder.svelte";
  import ThemeSelector from "./lib/components/ThemeSelector.svelte";
  import { initTheme } from "./lib/stores/themeStore";
  import {
    collectionStore,
    selectedRequest,
  } from "./lib/stores/collectionStore";
  import { environmentStore } from "./lib/stores/environmentStore";
  import Button from "./lib/components/base/Button.svelte";

  let showThemeSelector = false;
  let showEnvironmentManager = false;
  let activeView: "builder" | "editor" = "builder";

  onMount(async () => {
    // Initialize theme on app start
    await initTheme();

    // Load collections and environments
    await collectionStore.loadCollections();
    await environmentStore.loadEnvironments();
  });

  function toggleThemeSelector() {
    showThemeSelector = !showThemeSelector;
  }

  function toggleEnvironmentManager() {
    showEnvironmentManager = !showEnvironmentManager;
  }
</script>

<MainLayout title="yapla">
  <svelte:fragment slot="navbar-actions">
    <div class="nav-actions">
      <div class="view-switcher">
        <Button variant="primary" size="small">Request Builder</Button>
      </div>
      <Button variant="secondary" on:click={toggleEnvironmentManager}
        >🌍 Environments</Button
      >
      <Button variant="secondary" on:click={toggleThemeSelector}
        >🎨 Theme</Button
      >
    </div>
  </svelte:fragment>

  <CollectionList />

  {#if activeView === "builder"}
    <HTTPRequestBuilder />
  {/if}
</MainLayout>

{#if showThemeSelector}
  <div class="modal-overlay" on:click={toggleThemeSelector}>
    <div class="modal-panel" on:click|stopPropagation>
      <ThemeSelector />
      <div class="modal-footer">
        <Button variant="secondary" on:click={toggleThemeSelector}>Close</Button
        >
      </div>
    </div>
  </div>
{/if}

{#if showEnvironmentManager}
  <EnvironmentManager onClose={toggleEnvironmentManager} />
{/if}

<style>
  :global(body) {
    margin: 0;
    padding: 0;
  }

  .nav-actions {
    display: flex;
    gap: var(--spacing-md);
    align-items: center;
  }

  .view-switcher {
    display: flex;
    gap: var(--spacing-xs);
  }

  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: var(--z-modal);
  }

  .modal-panel {
    background: var(--bg-primary);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    max-width: 800px;
    width: 90%;
    max-height: 90vh;
    overflow-y: auto;
    box-shadow: var(--shadow-lg);
  }

  .modal-footer {
    padding: var(--space-lg);
    border-top: 1px solid var(--border);
    display: flex;
    justify-content: flex-end;
  }
</style>
