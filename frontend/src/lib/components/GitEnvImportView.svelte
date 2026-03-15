<script lang="ts">
  import Button from "flowbite-svelte/Button.svelte";
  import { environmentStore } from "$src/lib/stores/environmentStore";
  import { notifications } from "$src/lib/stores/notificationStore";
  import {
    GetGitRemoteBranches,
    IdentifyGitProvider,
    SetupGitEnvironment
  } from "$wails/go/main/App";

  interface Props {
    onImported?: () => void;
  }

  let { onImported }: Props = $props();

  let gitUrl = $state("");
  let remotePath = $state("");
  let loading = $state(false);
  let fetchingBranches = $state(false);
  let detectedProvider = "git";
  let errorMessage = $state("");
  let branches: string[] = $state([]);
  let selectedBranch = $state("");

  async function handleUrlChange() {
    if (!gitUrl) return;
    try {
      detectedProvider = await IdentifyGitProvider(gitUrl);
    } catch (err) {
      console.error("Failed to identify provider", err);
    }
  }

  async function fetchBranches() {
    if (!gitUrl) return;
    fetchingBranches = true;
    errorMessage = "";
    try {
      const b = await GetGitRemoteBranches(gitUrl);
      branches = b || [];
      if (branches.length > 0) {
        if (branches.includes("main")) selectedBranch = "main";
        else if (branches.includes("master")) selectedBranch = "master";
        else selectedBranch = branches[0];
      }
    } catch (err) {
      console.error("Failed to fetch branches", err);
      errorMessage = "Failed to fetch branches. Check URL and credentials.";
    } finally {
      fetchingBranches = false;
    }
  }

  async function handleImport() {
    if (!gitUrl || !remotePath) {
      errorMessage = "URL and Remote Path are required";
      return;
    }

    loading = true;
    errorMessage = "";
    try {
      await handleUrlChange();

      let finalUrl = gitUrl;
      if (selectedBranch) {
        finalUrl = `${gitUrl}#${selectedBranch}`;
      }

      console.log("[GitEnvImport] Setting up environment from URL:", finalUrl, "path:", remotePath);
      await SetupGitEnvironment(finalUrl, remotePath, "", detectedProvider);
      await environmentStore.loadEnvironments();

      notifications.success("Git environment imported successfully");
      onImported?.();
    } catch (err) {
      console.error("[GitEnvImport] Import failed:", err);
      errorMessage = String(err);
      notifications.error("Failed to setup Git environment", errorMessage);
    } finally {
      loading = false;
    }
  }
</script>

<div class="git-import-container">
  {#if errorMessage}
    <div class="error-box">
      <div class="error-content">
        <svg
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
        >
          <circle cx="12" cy="12" r="10"></circle>
          <line x1="12" y1="16" x2="12" y2="12"></line>
          <line x1="12" y1="8" x2="12.01" y2="8"></line>
        </svg>
        <span class="error-text">{errorMessage}</span>
      </div>
      <button class="dismiss-btn" onclick={() => (errorMessage = "")}>✕</button>
    </div>
  {/if}

  <div class="form-group">
    <label for="git-url">Repository URL</label>
    <div class="input-with-button">
      <input
        id="git-url"
        type="text"
        bind:value={gitUrl}
        onblur={handleUrlChange}
        placeholder="https://github.com/user/repo.git"
        class="form-input"
        disabled={loading || fetchingBranches}
      />
      <Button
        color="light"
        size="sm"
        disabled={!gitUrl || loading || fetchingBranches}
        onclick={fetchBranches}
      >
        {fetchingBranches ? "..." : "Fetch Branches"}
      </Button>
    </div>
  </div>

  {#if branches.length > 0}
    <div class="form-group">
      <label for="git-branch">Branch</label>
      <select id="git-branch" bind:value={selectedBranch} class="form-select" disabled={loading}>
        {#each branches as branch (branch)}
          <option value={branch}>{branch}</option>
        {/each}
      </select>
    </div>
  {/if}

  <div class="form-group">
    <label for="remote-path">Remote File Path</label>
    <input
      id="remote-path"
      type="text"
      bind:value={remotePath}
      placeholder="e.g. environments/dev.json or env.bru"
      class="form-input"
      disabled={loading}
    />
    <p class="field-hint">The path to the environment file within the repository.</p>
  </div>

  <div class="info-box">
    <svg
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="2"
    >
      <circle cx="12" cy="12" r="10"></circle>
      <line x1="12" y1="16" x2="12" y2="12"></line>
      <line x1="12" y1="8" x2="12.01" y2="8"></line>
    </svg>
    <span>Yapla will automatically detect the environment name from the file content.</span>
  </div>

  <div class="actions">
    <Button
      color="primary"
      disabled={!gitUrl || !remotePath || loading || fetchingBranches}
      onclick={handleImport}
    >
      {loading ? "Importing..." : "Import from Git"}
    </Button>
  </div>
</div>
