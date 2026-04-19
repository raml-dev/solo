<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import CodeMirrorEditor from "$src/lib/components/RequestBuilder/CodeMirrorEditor.svelte";
  import type { InputFormat } from "$src/lib/components/RequestBuilder/types";
  import type { ResolvedVariableEntry } from "$src/lib/utils/variableResolution";

  interface Props {
    requestBody: string;
    format: InputFormat;
    variableEntries?: ResolvedVariableEntry[];
    onChange?: () => void;
  }

  let {
    requestBody = $bindable(),
    format = $bindable(),
    variableEntries = [],
    onChange
  }: Props = $props();
</script>

<div class="min-h-0 w-full flex-1 overflow-hidden">
  <CodeMirrorEditor
    bind:value={requestBody}
    bind:format
    showCopyPaste
    {variableEntries}
    onChange={(value) => {
      requestBody = value;
      onChange?.();
    }}
  />
</div>
