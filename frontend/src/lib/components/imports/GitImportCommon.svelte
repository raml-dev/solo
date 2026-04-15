<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import { notifications } from "$src/lib/stores/notificationStore";
  import { GetGitRemoteBranches, IdentifyGitProvider } from "$wails/go/main/App";
  import Alert from "flowbite-svelte/Alert.svelte";
  import Button from "flowbite-svelte/Button.svelte";
  import Helper from "flowbite-svelte/Helper.svelte";
  import Input from "flowbite-svelte/Input.svelte";
  import Label from "flowbite-svelte/Label.svelte";
  import Select from "flowbite-svelte/Select.svelte";

  interface GitImportActionState {
    loading: boolean;
    disabled: boolean;
    submit: () => void;
  }

  interface Props {
    remotePathLabel: string;
    remotePathPlaceholder: string;
    remotePathHelper: string;
    successMessage: string;
    errorTitle: string;
    infoMessage: string;
    setupFromGit: (urlWithBranch: string, remotePath: string, provider: string) => Promise<void>;
    reloadAfterImport: () => Promise<void>;
    onImported?: () => void;
    onActionStateChange?: (state: GitImportActionState) => void;
  }

  let {
    remotePathLabel,
    remotePathPlaceholder,
    remotePathHelper,
    successMessage,
    errorTitle,
    infoMessage,
    setupFromGit,
    reloadAfterImport,
    onImported,
    onActionStateChange
  }: Props = $props();

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
      const fetchedBranches = await GetGitRemoteBranches(gitUrl);
      branches = fetchedBranches || [];

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

      await setupFromGit(finalUrl, remotePath, detectedProvider);
      await reloadAfterImport();

      notifications.success(successMessage);
      onImported?.();
    } catch (err) {
      errorMessage = String(err);
      notifications.error(errorTitle, errorMessage);
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
    <Label for="remote-path">{remotePathLabel}</Label>
    <Input
      id="remote-path"
      type="text"
      bind:value={remotePath}
      placeholder={remotePathPlaceholder}
      disabled={loading}
    />
    <Helper>{remotePathHelper}</Helper>
  </div>

  <Alert color="blue">{infoMessage}</Alert>
</div>
