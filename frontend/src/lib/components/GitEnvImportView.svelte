<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import GitImportCommon from "$src/lib/components/imports/GitImportCommon.svelte";
  import { environmentStore } from "$src/lib/stores/environmentStore.svelte";
  import { SetupGitEnvironment } from "$wails/go/main/App";

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
</script>

<GitImportCommon
  remotePathLabel="Remote File Path"
  remotePathPlaceholder="e.g. environments/dev.json or env.bru"
  remotePathHelper="The path to the environment file within the repository."
  successMessage="Git environment imported successfully"
  errorTitle="Failed to setup Git environment"
  infoMessage="solo will automatically detect the environment name from the file content."
  setupFromGit={(urlWithBranch, remotePath, provider) =>
    SetupGitEnvironment(urlWithBranch, remotePath, "", provider)}
  reloadAfterImport={environmentStore.loadEnvironments}
  {onImported}
  {onActionStateChange}
/>
