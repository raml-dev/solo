<script lang="ts">
  import Card from "flowbite-svelte/Card.svelte";
  import CodeMirrorEditor from "$src/lib/components/RequestBuilder/CodeMirrorEditor.svelte";
  import type { InputFormat } from "$src/lib/components/RequestBuilder/types";
  import { selectedEnvironment } from "$src/lib/stores/environmentStore";

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

<Card class="body-editor-wrapper p-0">
  <CodeMirrorEditor
    bind:value={requestBody}
    bind:format
    {environmentEntries}
    onChange={(value) => {
      requestBody = value;
      onChange?.();
    }}
  />
</Card>
