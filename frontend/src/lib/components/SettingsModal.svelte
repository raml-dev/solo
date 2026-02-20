<script lang="ts">
  import { onMount } from "svelte";
  import { configStore } from "../stores/configurationStore";
  import Button from "./base/Button.svelte";
  import { configuration } from "../../../wailsjs/go/models";

  export let onClose: () => void;

  let localConfig: configuration.Configuration;
  let loading = true;
  let error = "";

  onMount(async () => {
    try {
      await configStore.init();
      // Clone to avoid mutating store directly until save
      // Need manual deep clone or JSON roundtrip for simplicity as generated models might not have clone
      const current = $configStore;
      localConfig = JSON.parse(JSON.stringify(current));
    } catch (e) {
      error = "Failed to load settings";
    } finally {
      loading = false;
    }
  });

  async function handleSave() {
    try {
        // Ensure numbers are numbers (input bind:value sometimes treats as string)
        localConfig.request.timeoutSeconds = Number(localConfig.request.timeoutSeconds);
        localConfig.request.maxRedirects = Number(localConfig.request.maxRedirects);

        await configStore.save(localConfig);
        onClose();
    } catch (e) {
        error = "Failed to save settings: " + e;
    }
  }
</script>

<div class="modal-overlay" on:click={onClose}>
  <div class="modal-panel" on:click|stopPropagation>
    <div class="modal-header">
      <h2>Global Settings</h2>
      <Button variant="ghost" on:click={onClose}>✕</Button>
    </div>

    <div class="modal-body">
      {#if loading}
        <p>Loading...</p>
      {:else if error}
        <p class="error">{error}</p>
      {:else}
        <div class="section">
          <h3>HTTP Request Defaults</h3>
          
          <div class="form-group">
            <label for="timeout">Timeout (seconds)</label>
            <input 
                id="timeout" 
                type="number" 
                min="0"
                bind:value={localConfig.request.timeoutSeconds} 
            />
            <span class="help-text">Max duration for requests (0 = no timeout)</span>
          </div>

          <div class="form-group">
            <label class="checkbox-label">
                <input type="checkbox" bind:checked={localConfig.request.validateSSL} />
                Validate SSL Certificates
            </label>
            <span class="help-text">Disable for self-signed certificates (INSECURE)</span>
          </div>

          <div class="form-group">
            <label class="checkbox-label">
                <input type="checkbox" bind:checked={localConfig.request.followRedirects} />
                Follow Redirects
            </label>
          </div>

          <div class="form-group">
            <label for="maxRedirects">Max Redirects</label>
            <input 
                id="maxRedirects" 
                type="number" 
                min="1"
                disabled={!localConfig.request.followRedirects}
                bind:value={localConfig.request.maxRedirects} 
            />
          </div>

          <div class="form-group">
            <label for="userAgent">Default User Agent</label>
            <input 
                id="userAgent" 
                type="text" 
                bind:value={localConfig.request.defaultUserAgent} 
            />
          </div>

           <div class="form-group">
            <label for="proxy">Proxy URL</label>
            <input 
                id="proxy" 
                type="text" 
                placeholder="http://user:pass@host:port"
                bind:value={localConfig.request.proxyUrl} 
            />
             <span class="help-text">Leave empty to disable proxy</span>
          </div>
        </div>

        <div class="section">
            <h3>General</h3>
             <div class="form-group">
                <label for="theme">Theme</label>
                 <!-- Simple text for now, should ideally be a select matching available themes -->
                 <input id="theme" type="text" bind:value={localConfig.general.theme} />
            </div>
             <div class="form-group">
                <label class="checkbox-label">
                    <input type="checkbox" bind:checked={localConfig.general.checkForUpdates} />
                    Check for Updates automatically
                </label>
            </div>
        </div>
      {/if}
    </div>

    <div class="modal-footer">
      <Button variant="secondary" on:click={onClose}>Cancel</Button>
      <Button variant="primary" on:click={handleSave}>Save Changes</Button>
    </div>
  </div>
</div>

<style>
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
    width: 600px;
    max-width: 90%;
    max-height: 85vh;
    display: flex;
    flex-direction: column;
    box-shadow: var(--shadow-lg);
  }

  .modal-header {
    padding: var(--space-lg);
    border-bottom: 1px solid var(--border);
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .modal-header h2 {
    margin: 0;
    font-size: var(--font-size-lg);
  }

  .modal-body {
    padding: var(--space-lg);
    overflow-y: auto;
    flex: 1;
  }

  .modal-footer {
    padding: var(--space-lg);
    border-top: 1px solid var(--border);
    display: flex;
    justify-content: flex-end;
    gap: var(--space-sm);
  }

  .section {
    margin-bottom: var(--space-xl);
  }

  .section h3 {
    margin-top: 0;
    margin-bottom: var(--space-md);
    font-size: var(--font-size-md);
    color: var(--text-secondary);
    border-bottom: 1px solid var(--border-light);
    padding-bottom: var(--space-xs);
  }

  .form-group {
    margin-bottom: var(--space-md);
    display: flex;
    flex-direction: column;
    gap: var(--space-xs);
  }

  label {
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-medium);
  }

  input[type="text"],
  input[type="number"] {
    padding: var(--space-sm);
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    background: var(--bg-secondary);
    color: var(--text);
  }

  .checkbox-label {
    flex-direction: row;
    align-items: center;
    gap: var(--space-sm);
    cursor: pointer;
  }

  .help-text {
    font-size: var(--font-size-xs);
    color: var(--text-secondary);
  }

  .error {
    color: var(--danger);
  }
</style>
