<script lang="ts">
  import { envAutocomplete } from "../../../lib/actions/envAutocomplete";
  import { selectedEnvironment } from "../../../lib/stores/environmentStore";

  export let requestBody: string;

  $: environmentEntries = Object.entries($selectedEnvironment?.values ?? {}).map(([key, val]) => ({
    key,
    value: String(val?.value ?? "")
  }));
</script>

<div class="body-editor">
  <textarea
    class="input code-input"
    rows="12"
    placeholder="Request body (JSON, XML, etc.)"
    bind:value={requestBody}
    use:envAutocomplete={{ entries: environmentEntries, insertMode: "token" }}
  ></textarea>
</div>

<style>
  .body-editor {
    padding: var(--space-md);
  }

  .code-input {
    width: 100%;
    font-family: var(--font-mono);
    font-size: var(--font-size-sm);
    line-height: var(--line-height-relaxed);
    resize: vertical;
  }
</style>
