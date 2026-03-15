<script lang="ts">
  import type { configuration as conf } from "$wails/go/models";

  interface Props {
    requestSettings: conf.RequestSettingsOverride;
    globalConfig: conf.Configuration;
    onChange?: () => void;
  }

  let { requestSettings = $bindable(), globalConfig, onChange }: Props = $props();

  function handleChange() {
    onChange?.();
  }
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
      bind:value={requestSettings.timeoutSeconds}
      oninput={handleChange}
    />
  </div>
  <div class="form-group">
    <label for="user-agent">User Agent</label>
    <input
      id="user-agent"
      type="text"
      placeholder={`Global: ${globalConfig.request.defaultUserAgent || "Yapla/1.0"}`}
      bind:value={requestSettings.defaultUserAgent}
      oninput={handleChange}
    />
  </div>
  <div class="form-group-row">
    <label class="checkbox-group">
      <input
        type="checkbox"
        bind:checked={requestSettings.followRedirects}
        onchange={handleChange}
      />
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
        bind:value={requestSettings.maxRedirects}
        oninput={handleChange}
        disabled={!requestSettings.followRedirects}
      />
    </div>
  </div>
  <div class="form-group">
    <label class="checkbox-group">
      <input type="checkbox" bind:checked={requestSettings.validateSSL} onchange={handleChange} />
      Validate SSL Certificates
    </label>
  </div>
  <div class="form-group">
    <label for="proxy">Proxy URL</label>
    <input
      id="proxy"
      type="text"
      placeholder={`Global: ${globalConfig.request.proxyUrl || "Use system settings"}`}
      bind:value={requestSettings.proxyUrl}
      oninput={handleChange}
    />
  </div>
</div>
