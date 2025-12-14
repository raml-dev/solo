<script lang="ts">
  import { onMount } from 'svelte';
  import MainLayout from './lib/components/MainLayout.svelte';
  import Sidebar from './lib/components/Sidebar.svelte';
  import HTTPRequestBuilder from './lib/components/HTTPRequestBuilder.svelte';
  import ThemeSelector from './lib/components/ThemeSelector.svelte';
  import { initTheme } from './lib/stores/themeStore';
  import Button from './lib/components/base/Button.svelte';

  // Sample data
  const collections = [
    { id: '1', name: 'Users API', method: 'GET' },
    { id: '2', name: 'Create User', method: 'POST' },
    { id: '3', name: 'Update User', method: 'PUT' },
    { id: '4', name: 'Delete User', method: 'DELETE' },
    { id: '5', name: 'Get Posts', method: 'GET' },
    { id: '6', name: 'Create Post', method: 'POST' },
  ];

  let selectedId: string | null = '1';
  let showThemeSelector = false;

  onMount(async () => {
    // Initialize theme on app start
    await initTheme();
  });

  function handleSelect(id: string) {
    selectedId = id;
    console.log('Selected:', id);
  }

  function toggleThemeSelector() {
    showThemeSelector = !showThemeSelector;
  }
</script>

<MainLayout title="yapla">
  <svelte:fragment slot="navbar-actions">
    <Button variant="secondary" on:click={toggleThemeSelector}>🎨 Theme</Button>
  </svelte:fragment>

  <Sidebar 
    items={collections} 
    {selectedId}
    onSelect={handleSelect}
  />
  <HTTPRequestBuilder />
</MainLayout>

{#if showThemeSelector}
  <div class="modal-overlay" on:click={toggleThemeSelector}>
    <div class="modal-panel" on:click|stopPropagation>
      <ThemeSelector />
      <div class="modal-footer">
        <Button variant="secondary" on:click={toggleThemeSelector}>Close</Button>
      </div>
    </div>
  </div>
{/if}

<style>
  :global(body) {
    margin: 0;
    padding: 0;
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