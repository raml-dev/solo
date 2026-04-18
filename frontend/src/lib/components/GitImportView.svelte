<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import GitImportCommon from "$src/lib/components/imports/GitImportCommon.svelte";
  import { collectionStore } from "$src/lib/stores/collectionStore.svelte";
  import { SetupGitCollection } from "$wails/go/main/App";

  interface GitImportActionState {
    loading: boolean;
    disabled: boolean;
    submit: () => void;
  }

  interface Props {
    onImported?: () => void;
    onActionStateChange?: (state: GitImportActionState) => void;
  }

  let { onImported, onActionStateChange }: Props = $props();
</script>

<GitImportCommon
  remotePathLabel="Remote File/Folder Path"
  remotePathPlaceholder="e.g. my_collection.json, bruno_folder, or . for repo root"
  remotePathHelper="The path to the collection within the repository. Use . for a Bruno collection at the repository root."
  successMessage="Git collection imported successfully"
  errorTitle="Failed to setup Git collection"
  infoMessage="solo will automatically detect the collection name from the file content."
  setupFromGit={(urlWithBranch, remotePath, provider) =>
    SetupGitCollection(urlWithBranch, remotePath, "", provider)}
  reloadAfterImport={collectionStore.loadCollections}
  {onImported}
  {onActionStateChange}
/>
