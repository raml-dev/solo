<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import CodeMirrorEditor from "./CodeMirrorEditor.svelte";
  import type { InputFormat } from "./types";
  import { selectedEnvironment } from "../../../lib/stores/environmentStore";

  export let requestBody: string;
  export let format: InputFormat;

  const dispatch = createEventDispatcher();

  $: environmentEntries = Object.entries($selectedEnvironment?.values ?? {}).map(([key, val]) => ({
    key,
    value: String(val?.value ?? "")
  }));
</script>

<div class="body-editor-wrapper">
  <CodeMirrorEditor
    bind:value={requestBody}
    bind:format={format}
    {environmentEntries}
    on:change={(e) => { requestBody = e.detail; dispatch("change"); }}
  />
</div>

<style>
  .body-editor-wrapper {
    display: flex;
    flex-direction: column;
    height: 100%;
    padding: var(--space-md);
    flex: 1;
  }
</style>
