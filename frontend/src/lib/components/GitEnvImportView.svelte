<script lang="ts">
  import Alert from "flowbite-svelte/Alert.svelte";
  import Button from "flowbite-svelte/Button.svelte";
  import Helper from "flowbite-svelte/Helper.svelte";
  import Input from "flowbite-svelte/Input.svelte";
  import Label from "flowbite-svelte/Label.svelte";
  import Select from "flowbite-svelte/Select.svelte";
  import { environmentStore } from "$src/lib/stores/environmentStore";
  import { notifications } from "$src/lib/stores/notificationStore";
  import {
    GetGitRemoteBranches,
    IdentifyGitProvider,
    SetupGitEnvironment
  } from "$wails/go/main/App";

  interface GitEnvImportActionState {
    loading: boolean;
    disabled: boolean;
    submit: () => void;
  }

  interface Props {
    onImported?: () => void;
    onActionStateChange?: (state: GitEnvImportActionState) => void;
  }

  let { onImported, onActionStateChange }: Props = $props();

  let gitUrl = $state("");
  let remotePath = $state("");
  let loading = $state(false);
  let fetchingBranches = $state(false);
  let detectedProvider = "git";
  let errorMessage = $state("");
  let branches: string[] = $state([]);
  let selectedBranch = $state("");
  const branchOptions = $derived(branches.map((branch) => ({ value: branch, name: branch })));

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

  $effect(() => {
    onActionStateChange?.({
      loading,
      disabled: !gitUrl || !remotePath || loading || fetchingBranches,
      submit: handleImport
    });
  });
</script>

<div class="space-y-4">
  {#if errorMessage}
    <Alert color="red">{errorMessage}</Alert>
  {/if}

  <div class="space-y-2">
    <Label for="git-url">Repository URL</Label>
    <div class="flex items-center gap-2">
      <Input
        id="git-url"
        type="text"
        bind:value={gitUrl}
        onblur={handleUrlChange}
        placeholder="https://github.com/user/repo.git"
        disabled={loading || fetchingBranches}
        class="flex-1"
      />
      <Button
        color="light"
        size="sm"
        loading={fetchingBranches}
        disabled={!gitUrl || loading || fetchingBranches}
        onclick={fetchBranches}
      >
        Fetch Branches
      </Button>
    </div>
  </div>

  {#if branches.length > 0}
    <div class="space-y-2">
      <Label for="git-branch">Branch</Label>
      <Select
        id="git-branch"
        bind:value={selectedBranch}
        items={branchOptions}
        size="sm"
        disabled={loading}
      />
    </div>
  {/if}

  <div class="space-y-2">
    <Label for="remote-path">Remote File Path</Label>
    <Input
      id="remote-path"
      type="text"
      bind:value={remotePath}
      placeholder="e.g. environments/dev.json or env.bru"
      disabled={loading}
    />
    <Helper>The path to the environment file within the repository.</Helper>
  </div>

  <Alert color="blue">
    solo will automatically detect the environment name from the file content.
  </Alert>
</div>
