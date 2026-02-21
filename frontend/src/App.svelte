<script lang="ts">
  import { onMount } from "svelte";
  import MainLayout from "./lib/components/MainLayout.svelte";
  import CollectionList from "./lib/components/CollectionList.svelte";
  import HTTPRequestBuilder from "./lib/components/HTTPRequestBuilder.svelte";
  import { configurationStore } from "./lib/stores/configurationStore";
  import { collectionStore } from "./lib/stores/collectionStore";
  import { environmentStore } from "./lib/stores/environmentStore";

  let activeView: "builder" | "editor" = "builder";

  onMount(async () => {
    // Initialize all app configuration, including theme
    await configurationStore.init();

    // Load collections and environments
    await collectionStore.loadCollections();
    await environmentStore.loadEnvironments();
  });
</script>

<MainLayout title="yapla">
  <svelte:fragment slot="navbar-actions">
    <div class="nav-actions">
      <div class="view-switcher">
        <!-- <Button variant="primary" size="small">Request Builder</Button> -->
      </div>
    </div>
  </svelte:fragment>

  <CollectionList />

  {#if activeView === "builder"}
    <HTTPRequestBuilder />
  {/if}
</MainLayout>

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
