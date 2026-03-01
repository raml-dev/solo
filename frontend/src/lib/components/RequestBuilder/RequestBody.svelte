<script lang="ts">
  export let requestBody: string;
</script>

<div class="body-editor">
  <textarea
    class="input code-input"
    rows="12"
    placeholder="Request body (JSON, XML, etc.)"
    bind:value={requestBody}
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
