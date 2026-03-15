<script lang="ts">
  import Badge from "flowbite-svelte/Badge.svelte";
  import Button from "flowbite-svelte/Button.svelte";
  import Card from "flowbite-svelte/Card.svelte";
  import CodeMirrorEditor from "$src/lib/components/RequestBuilder/CodeMirrorEditor.svelte";
  import { sessionVarsStore } from "$src/lib/stores/sessionVarsStore";

  interface Props {
    preRequestScript?: string;
    postResponseScript?: string;
    onPreChange?: (val: string) => void;
    onPostChange?: (val: string) => void;
  }

  let {
    preRequestScript = $bindable(""),
    postResponseScript = $bindable(""),
    onPreChange = () => {},
    onPostChange = () => {}
  }: Props = $props();

  type ScriptSection = "pre" | "post";
  let activeSection: ScriptSection = $state("pre");

  let sessionEntries: [string, string][] = $derived(
    Object.entries($sessionVarsStore) as [string, string][]
  );

  const LUA_HINT = `-- Available globals:
-- request.method, request.url, request.headers, request.body  (pre only, mutable)
-- response.status, response.headers, response.body, response.time  (post only)
-- env.get("key")  /  env.set("key", "value")  /  env.key = "value"
-- env.log("message")`;
</script>

<div class="scripts-panel">
  <nav class="scripts-nav">
    <button
      class="script-nav-item"
      class:active={activeSection === "pre"}
      onclick={() => (activeSection = "pre")}
    >
      <span class="script-nav-label">Pre-request</span>
      {#if preRequestScript.trim()}
        <span class="script-dot" title="Script active"></span>
      {/if}
    </button>

    <button
      class="script-nav-item"
      class:active={activeSection === "post"}
      onclick={() => (activeSection = "post")}
    >
      <span class="script-nav-label">Post-response</span>
      {#if postResponseScript.trim()}
        <span class="script-dot" title="Script active"></span>
      {/if}
    </button>

    <Card class="session-vars-section p-3">
      <div class="session-vars-header">
        <span class="session-vars-title">Session Vars</span>
        {#if sessionEntries.length > 0}
          <Button
            color="light"
            size="xs"
            class="session-vars-clear"
            onclick={() => sessionVarsStore.clear()}
            title="Clear all session variables"
          >
            Clear
          </Button>
        {/if}
      </div>
      {#if sessionEntries.length === 0}
        <p class="session-vars-empty">
          No session vars yet.<br />Use <code>env.set()</code> in a script.
        </p>
      {:else}
        <ul class="session-vars-list">
          {#each sessionEntries as [key, value] (key)}
            <li class="session-var-item">
              <span class="session-var-key">{key}</span>
              <span class="session-var-value">{value}</span>
            </li>
          {/each}
        </ul>
      {/if}
    </Card>
  </nav>

  <Card class="scripts-editor p-0">
    <div class="editor-header">
      <span class="editor-title">
        {activeSection === "pre" ? "Pre-request Script" : "Post-response Script"}
      </span>
      <Badge color="blue" class="editor-lang">Lua</Badge>
    </div>

    <div class="editor-hint">
      <pre>{LUA_HINT}</pre>
    </div>

    {#if activeSection === "pre"}
      <div class="editor-wrap">
        <CodeMirrorEditor value={preRequestScript} language="lua" onChange={onPreChange} />
      </div>
    {:else}
      <div class="editor-wrap">
        <CodeMirrorEditor value={postResponseScript} language="lua" onChange={onPostChange} />
      </div>
    {/if}
  </Card>
</div>
