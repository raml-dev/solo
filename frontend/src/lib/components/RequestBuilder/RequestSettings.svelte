<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: GPL-3.0-only
-->

<script lang="ts">
  import type { configuration as conf } from "$wails/go/models";
  import Helper from "flowbite-svelte/Helper.svelte";
  import Input from "flowbite-svelte/Input.svelte";
  import Label from "flowbite-svelte/Label.svelte";
  import Toggle from "flowbite-svelte/Toggle.svelte";

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

<div class="flex-1 space-y-4 overflow-y-auto p-3">
  <div class="space-y-2">
    <Label for="timeout">Timeout (seconds)</Label>
    <Input
      id="timeout"
      type="number"
      min="0"
      step="1"
      size="sm"
      placeholder={`Global: ${globalConfig.request.timeoutSeconds || "30"}`}
      bind:value={requestSettings.timeoutSeconds}
      oninput={handleChange}
    />
  </div>

  <div class="space-y-2">
    <Label for="user-agent">User Agent</Label>
    <Input
      id="user-agent"
      type="text"
      size="sm"
      placeholder={`Global: ${globalConfig.request.defaultUserAgent || "Solo/1.0"}`}
      bind:value={requestSettings.defaultUserAgent}
      oninput={handleChange}
    />
  </div>

  <div class="grid grid-cols-1 gap-3 md:grid-cols-[auto_1fr] md:items-end">
    <div class="md:pb-1">
      <Toggle bind:checked={requestSettings.followRedirects} size="small" onchange={handleChange}
        >Follow Redirects</Toggle
      >
    </div>
    <div class="space-y-2">
      <Label for="max-redirects">Max Redirects</Label>
      <Input
        id="max-redirects"
        type="number"
        min="0"
        step="1"
        size="sm"
        placeholder={`Global: ${globalConfig.request.maxRedirects || "10"}`}
        bind:value={requestSettings.maxRedirects}
        oninput={handleChange}
        disabled={!requestSettings.followRedirects}
      />
    </div>
  </div>

  <div>
    <Toggle bind:checked={requestSettings.validateSSL} size="small" onchange={handleChange}
      >Validate SSL Certificates</Toggle
    >
  </div>

  <div class="space-y-2">
    <Label for="proxy">Proxy URL</Label>
    <Input
      id="proxy"
      type="text"
      size="sm"
      placeholder={`Global: ${globalConfig.request.proxyUrl || "Use system settings"}`}
      bind:value={requestSettings.proxyUrl}
      oninput={handleChange}
    />
    <Helper>Leave empty to use global/system proxy behavior.</Helper>
  </div>
</div>
