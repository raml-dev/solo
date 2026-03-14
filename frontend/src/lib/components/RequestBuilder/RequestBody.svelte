<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import CodeMirrorEditor from "./CodeMirrorEditor.svelte";
  import type { InputFormat } from "./types";
  import { selectedEnvironment } from "../../../lib/stores/environmentStore";

  interface Props {
    requestBody: string;
    format: InputFormat;
  }

  let { requestBody = $bindable(), format = $bindable() }: Props = $props();

  const dispatch = createEventDispatcher();

  let environmentEntries = $derived(
    Object.entries($selectedEnvironment?.values ?? {}).map(([key, val]) => ({
      key,
      value: String(val?.value ?? "")
    }))
  );
</script>

<div class="body-editor-wrapper">
  <CodeMirrorEditor
    bind:value={requestBody}
    bind:format
    {environmentEntries}
    on:change={(e) => {
      requestBody = e.detail;
      dispatch("change");
    }}
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
