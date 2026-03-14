<script lang="ts">
  import CodeMirrorEditor from "./CodeMirrorEditor.svelte";
  import type { InputFormat } from "./types";
  import { selectedEnvironment } from "../../../lib/stores/environmentStore";

  interface Props {
    requestBody: string;
    format: InputFormat;
    onChange?: () => void;
  }

  let { requestBody = $bindable(), format = $bindable(), onChange }: Props = $props();

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
    onChange={(value) => {
      requestBody = value;
      onChange?.();
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
