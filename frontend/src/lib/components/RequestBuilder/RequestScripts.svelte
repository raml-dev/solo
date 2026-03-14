<script lang="ts">
  import { run } from "svelte/legacy";

  import CodeMirrorEditor from "$src/lib/components/RequestBuilder/CodeMirrorEditor.svelte";
  import { sessionVarsStore } from "$src/lib/stores/sessionVarsStore";

  interface Props {
    preRequestScript?: string;
    postResponseScript?: string;
    // Callbacks per notificare il parent dei cambiamenti
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

  let sessionEntries: [string, string][] = $state([]);
  run(() => {
    sessionEntries = Object.entries($sessionVarsStore);
  });

  const LUA_HINT = `-- Available globals:
-- request.method, request.url, request.headers, request.body  (pre only, mutable)
-- response.status, response.headers, response.body, response.time  (post only)
-- env.get("key")  /  env.set("key", "value")  /  env.key = "value"
-- env.log("message")`;
</script>

<div class="scripts-panel">
  <!-- Sidebar -->
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

    <!-- Session vars section -->
    <div class="session-vars-section">
      <div class="session-vars-header">
        <span class="session-vars-title">Session Vars</span>
        {#if sessionEntries.length > 0}
          <button
            class="session-vars-clear"
            onclick={() => sessionVarsStore.clear()}
            title="Clear all session variables">Clear</button
          >
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
    </div>
  </nav>

  <!-- Editor area -->
  <div class="scripts-editor">
    <div class="editor-header">
      <span class="editor-title">
        {activeSection === "pre" ? "Pre-request Script" : "Post-response Script"}
      </span>
      <span class="editor-lang">Lua</span>
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
  </div>
</div>

<style>
  .scripts-panel {
    display: flex;
    height: 100%;
    overflow: hidden;
  }

  /* ---- Sidebar ---- */
  .scripts-nav {
    width: 180px;
    flex-shrink: 0;
    border-right: 1px solid var(--border);
    display: flex;
    flex-direction: column;
    padding: var(--space-sm) 0;
    background: var(--bg-secondary);
    overflow-y: auto;
  }

  .script-nav-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-xs);
    padding: var(--space-sm) var(--space-md);
    background: none;
    border: none;
    cursor: pointer;
    color: var(--text-muted);
    font-size: var(--font-size-sm);
    text-align: left;
    transition:
      background 0.15s,
      color 0.15s;
    font-family: inherit;
    width: 100%;
  }
  .script-nav-item:hover {
    background: var(--bg-tertiary);
    color: var(--text);
  }
  .script-nav-item.active {
    background: var(--bg-tertiary);
    color: var(--text);
    font-weight: var(--font-weight-semibold);
  }

  .script-nav-label {
    flex: 1;
  }

  .script-dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--success);
    flex-shrink: 0;
  }

  /* ---- Session vars ---- */
  .session-vars-section {
    margin-top: auto;
    border-top: 1px solid var(--border);
    padding: var(--space-sm) var(--space-md);
    display: flex;
    flex-direction: column;
    gap: var(--space-xs);
  }

  .session-vars-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .session-vars-title {
    font-size: 0.65rem;
    font-weight: var(--font-weight-semibold);
    letter-spacing: 0.07em;
    text-transform: uppercase;
    color: var(--text-muted);
  }

  .session-vars-clear {
    background: none;
    border: none;
    cursor: pointer;
    font-size: 0.65rem;
    color: var(--danger);
    padding: 0;
    font-family: inherit;
  }
  .session-vars-clear:hover {
    text-decoration: underline;
  }

  .session-vars-empty {
    font-size: 0.7rem;
    color: var(--text-muted);
    line-height: 1.4;
    margin: 0;
  }
  .session-vars-empty code {
    font-family: var(--font-mono);
    background: var(--bg-tertiary);
    padding: 0 3px;
    border-radius: 2px;
  }

  .session-vars-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 3px;
    max-height: 180px;
    overflow-y: auto;
  }

  .session-var-item {
    display: flex;
    flex-direction: column;
    gap: 1px;
    padding: 3px var(--space-xs);
    background: var(--bg-tertiary);
    border-radius: var(--radius-sm);
  }

  .session-var-key {
    font-size: 0.65rem;
    font-family: var(--font-mono);
    color: var(--info);
    font-weight: var(--font-weight-semibold);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .session-var-value {
    font-size: 0.65rem;
    font-family: var(--font-mono);
    color: var(--text-muted);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  /* ---- Editor area ---- */
  .scripts-editor {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    background: var(--bg-primary);
  }

  .editor-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-sm) var(--space-md);
    border-bottom: 1px solid var(--border);
    background: var(--bg-secondary);
    flex-shrink: 0;
  }

  .editor-title {
    font-size: var(--font-size-sm);
    font-weight: var(--font-weight-semibold);
    color: var(--text);
  }

  .editor-lang {
    font-size: 0.65rem;
    font-weight: var(--font-weight-semibold);
    letter-spacing: 0.07em;
    text-transform: uppercase;
    color: var(--text-muted);
    background: var(--bg-tertiary);
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    padding: 1px 6px;
  }

  .editor-hint {
    background: color-mix(in srgb, var(--info) 6%, var(--bg-secondary));
    border-bottom: 1px solid color-mix(in srgb, var(--info) 20%, transparent);
    flex-shrink: 0;
  }

  .editor-hint pre {
    margin: 0;
    padding: var(--space-sm) var(--space-md);
    font-family: var(--font-mono);
    font-size: 0.72rem;
    color: var(--text-muted);
    line-height: 1.5;
    white-space: pre;
    overflow-x: auto;
  }

  .editor-wrap {
    flex: 1;
    overflow: hidden;
    display: flex;
    flex-direction: column;
  }

  .editor-wrap :global(.cm-editor) {
    height: 100%;
  }
</style>
