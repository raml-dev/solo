<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import type { configuration as conf } from "../../../../wailsjs/go/models";

  export let requestSettings: conf.RequestSettingsOverride;
  export let globalConfig: conf.Configuration;

  const dispatch = createEventDispatcher();
  function onChange() { dispatch("change"); }
</script>

<div class="settings-form-container">
  <div class="form-group">
    <label for="timeout">Timeout (seconds)</label>
    <input
      id="timeout"
      type="number"
      min="0"
      step="1"
      placeholder={`Global: ${globalConfig.request.timeoutSeconds || "30"}`}
      bind:value={requestSettings.timeoutSeconds} on:input={onChange}
    />
  </div>
  <div class="form-group">
    <label for="user-agent">User Agent</label>
    <input
      id="user-agent"
      type="text"
      placeholder={`Global: ${globalConfig.request.defaultUserAgent || "Yapla/1.0"}`}
      bind:value={requestSettings.defaultUserAgent} on:input={onChange}
    />
  </div>
  <div class="form-group-row">
    <label class="checkbox-group">
      <input type="checkbox" bind:checked={requestSettings.followRedirects} on:change={onChange} />
      Follow Redirects
    </label>
    <div class="form-group">
      <label for="max-redirects">Max Redirects</label>
      <input
        id="max-redirects"
        type="number"
        min="0"
        step="1"
        placeholder={`Global: ${globalConfig.request.maxRedirects || "10"}`}
        bind:value={requestSettings.maxRedirects} on:input={onChange}
        disabled={!requestSettings.followRedirects}
      />
    </div>
  </div>
  <div class="form-group">
    <label class="checkbox-group">
      <input type="checkbox" bind:checked={requestSettings.validateSSL} on:change={onChange} />
      Validate SSL Certificates
    </label>
  </div>
  <div class="form-group">
    <label for="proxy">Proxy URL</label>
    <input
      id="proxy"
      type="text"
      placeholder={`Global: ${globalConfig.request.proxyUrl || "Use system settings"}`}
      bind:value={requestSettings.proxyUrl} on:input={onChange}
    />
  </div>
</div>

<style>
  .settings-form-container {
    padding: var(--space-lg);
    display: flex;
    flex-direction: column;
    gap: var(--space-lg);
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: var(--space-xs);
  }

  .form-group-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-lg);
    align-items: flex-end;
  }

  .checkbox-group {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
    cursor: pointer;
    user-select: none;
  }

  label {
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-medium);
    color: var(--text-muted);
  }

  input[type="text"],
  input[type="number"] {
    padding: var(--space-sm);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    background: var(--bg-secondary);
    color: var(--text);
    font-size: var(--font-size-sm);
  }

  input:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
</style>
