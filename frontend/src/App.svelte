<script lang="ts">
  import { onMount } from "svelte";
  import MainLayout from "./lib/components/MainLayout.svelte";
  import CollectionList from "./lib/components/CollectionList.svelte";
  import EnvironmentManager from "./lib/components/EnvironmentManager.svelte";
  import HTTPRequestBuilder from "./lib/components/HTTPRequestBuilder.svelte";
  import ThemeSelector from "./lib/components/ThemeSelector.svelte";
  import { initTheme } from "./lib/stores/themeStore";
  import { collectionStore } from "./lib/stores/collectionStore";
  import { environmentStore } from "./lib/stores/environmentStore";
  import Button from "./lib/components/base/Button.svelte";
  import Modal from "./lib/components/base/Modal.svelte";

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
        <!-- <Button variant="primary" size="small">Request Builder</Button> -->
      </div>
      <Button variant="secondary" click={toggleEnvironmentManager}>🌍 Environments</Button>
      <Button variant="secondary" click={toggleThemeSelector}>🎨 Theme</Button>
    </div>
  </svelte:fragment>

  <CollectionList />

  {#if activeView === "builder"}
    <HTTPRequestBuilder />
  {/if}
</MainLayout>

{#if showThemeSelector}
  <Modal toggleFn={toggleThemeSelector}>
    <ThemeSelector />
  </Modal>
{/if}

{#if showEnvironmentManager}
  <Modal toggleFn={toggleEnvironmentManager}>
    <EnvironmentManager />
  </Modal>
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
</style>
