<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import CodeMirrorEditor from "$src/lib/components/RequestBuilder/CodeMirrorEditor.svelte";
  import type { InputFormat } from "$src/lib/components/RequestBuilder/types";
  import { environmentStoreState } from "$src/lib/stores/environmentStore.svelte";

  interface Props {
    requestBody: string;
    format: InputFormat;
    onChange?: () => void;
  }

  let { requestBody = $bindable(), format = $bindable(), onChange }: Props = $props();

  let selectedEnvironment = $derived(
    environmentStoreState.environments.find(
      (e) => e.name === environmentStoreState.selectedEnvironmentName
    ) || null
  );

  let environmentEntries = $derived(
    Object.entries(selectedEnvironment?.values ?? {}).map(([key, val]) => ({
      key,
      value: String(val?.value ?? "")
    }))
  );
</script>

<div class="min-h-0 w-full flex-1 overflow-hidden">
  <CodeMirrorEditor
    bind:value={requestBody}
    bind:format
    showCopyPaste
    {environmentEntries}
    onChange={(value) => {
      requestBody = value;
      onChange?.();
    }}
  />
</div>
