<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import ToastContainer from "$src/lib/components/base/ToastContainer.svelte";
  import { modalStack, topModalId } from "$src/lib/stores/modalStackStore.svelte";
  import { notifications } from "$src/lib/stores/notificationStore";
  import { DownloadAssets, GetAppInfo, GetUpdatesFromRepo } from "$wails/go/main/App";
  import { appinfo } from "$wails/go/models";
  import { EventsOff, EventsOn } from "$wails/runtime";
  import { BrowserOpenURL } from "$wails/runtime/runtime";
  import DOMPurify from "dompurify";
  import Button from "flowbite-svelte/Button.svelte";
  import Modal from "flowbite-svelte/Modal.svelte";
  import { marked } from "marked";
  import { onDestroy, onMount } from "svelte";

  let visible = $state(false);
  let loading = $state(false);
  let ignoredVersion = $state("");
  let currentVersion = $state("");
  let updateInfo: appinfo.GitHubResponse | null = $state(null);
  let selectedRelease: appinfo.GitHubRelease | null = $state(null);
  let releaseNotesHtml = $state("");
  let downloadedPath = $state("");

  const releaseNotesModal = modalStack.createModal("update-release-notes");
  const downloadCompleteModal = modalStack.createModal("update-downloaded");

  function escapeHtml(input: string): string {
    return input
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;")
      .replaceAll("'", "&#39;");
  }

  function getReleaseNotes(release: appinfo.GitHubRelease): string {
    const body = (release as appinfo.GitHubRelease & { body?: string; string?: string }).body;
    const legacy = (release as appinfo.GitHubRelease & { body?: string; string?: string }).string;
    const raw = (body || legacy || "").trim();
    if (!raw) return "";

    // Normalize visual newline marker often seen in copied/debug payloads.
    const normalizedArrows = raw.replaceAll("↵", "\n");

    // Some payloads may arrive with escaped newlines instead of real line breaks.
    const hasRealNewlines = normalizedArrows.includes("\n");
    const maybeEscaped = normalizedArrows.includes("\\n");
    if (!hasRealNewlines && maybeEscaped) {
      return normalizedArrows
        .replaceAll("\\r\\n", "\n")
        .replaceAll("\\n", "\n")
        .replaceAll("\\r", "");
    }

    return normalizedArrows.replaceAll("\r\n", "\n");
  }

  function getVersion(release: appinfo.GitHubRelease): string {
    return (release.tag_name || release.name || "").trim();
  }

  function handleUpdateAvailable(payload: unknown) {
    const release = appinfo.GitHubResponse.createFrom(payload);
    if (!release || !release.Release) return;

    const version = getVersion(release.Release);
    if (!version) return;
    if (ignoredVersion === version) return;

    updateInfo = release;
    selectedRelease = release.Release;
    visible = true;
  }

  function handleDownloaded(payload?: unknown) {
    loading = false;
    visible = false;
    downloadedPath =
      typeof payload === "object" &&
      payload !== null &&
      "path" in payload &&
      typeof payload.path === "string"
        ? payload.path
        : "";
    downloadCompleteModal.open = true;
  }

  function handleIgnore() {
    if (!selectedRelease) return;
    ignoredVersion = getVersion(selectedRelease);
    visible = false;
  }

  async function handleUpdate() {
    if (!updateInfo || loading) return;

    loading = true;
    try {
      const result = await DownloadAssets(updateInfo, currentVersion);
      if (!result) {
        loading = false;
      }
    } catch (err) {
      loading = false;
      notifications.error("Failed to download update", String(err));
    }
  }

  function openReleaseNotesModal() {
    const markdown = selectedRelease ? getReleaseNotes(selectedRelease) : "";

    if (!markdown) {
      releaseNotesHtml = "<p>No release notes available.</p>";
      releaseNotesModal.open = true;
      return;
    }

    try {
      const rendered = marked.parse(markdown, { gfm: true, breaks: true });
      const html = typeof rendered === "string" ? rendered : String(rendered);
      releaseNotesHtml = DOMPurify.sanitize(html);
    } catch (err) {
      console.warn("Release notes markdown parsing failed", err);
      releaseNotesHtml = `<pre>${escapeHtml(markdown)}</pre>`;
    }

    releaseNotesModal.open = true;
  }

  function openReleasePage() {
    const releaseURL = (selectedRelease?.html_url || "").trim();
    if (!releaseURL) return;
    BrowserOpenURL(releaseURL);
  }

  onMount(() => {
    EventsOn("updates:available", handleUpdateAvailable);
    EventsOn("updates:downloaded", handleDownloaded);

    (async () => {
      try {
        const [appData, latestUpdate] = await Promise.all([GetAppInfo(), GetUpdatesFromRepo()]);
        currentVersion = appData?.productVersion ?? "";
        handleUpdateAvailable(latestUpdate);
      } catch {
        currentVersion = "";
      }
    })();
  });

  onDestroy(() => {
    EventsOff("updates:available");
    EventsOff("updates:downloaded");
    modalStack.destroyModal(releaseNotesModal.id);
    modalStack.destroyModal(downloadCompleteModal.id);
  });
</script>

{#if visible && selectedRelease}
  <div
    class="sticky top-0 z-40 border-b border-blue-200 bg-primary-500/40 px-3 py-2 dark:border-primary-700 dark:bg-primary-900/40"
  >
    <div class="flex flex-wrap items-center gap-3">
      <div class="min-w-0 flex-1">
        <p class="text-sm font-semibold text-blue-900 dark:text-blue-100">Update available</p>
        <p class="text-sm text-blue-900/90 dark:text-blue-100/90">
          Solo version {getVersion(selectedRelease)} is available.
          <button
            type="button"
            class="ml-1 underline underline-offset-2 hover:no-underline"
            onclick={openReleaseNotesModal}
          >
            Show release notes
          </button>
        </p>
      </div>
      <div class="flex items-center gap-2">
        <Button size="xs" color="light" onclick={handleIgnore} disabled={loading}>Ignore</Button>
        {#if selectedRelease.html_url}
          <Button size="xs" color="light" onclick={openReleasePage} disabled={loading}>
            Release page
          </Button>
        {/if}
        <Button size="xs" color="primary" onclick={handleUpdate} disabled={loading} {loading}
          >Download latest version</Button
        >
      </div>
    </div>
  </div>
{/if}

{#if releaseNotesModal.open}
  <Modal title="Release notes" bind:open={releaseNotesModal.open} size="lg">
    {#if $topModalId === releaseNotesModal.id}
      <ToastContainer />
    {/if}
    <article class="release-notes max-h-[65vh] overflow-y-auto" aria-label="Release notes content">
      <!-- eslint-disable-next-line svelte/no-at-html-tags -->
      {@html releaseNotesHtml}
    </article>
  </Modal>
{/if}

{#if downloadCompleteModal.open}
  <Modal title="Update downloaded" bind:open={downloadCompleteModal.open} size="md">
    {#if $topModalId === downloadCompleteModal.id}
      <ToastContainer />
    {/if}
    <p class="text-sm text-neutral-700 dark:text-neutral-200">
      Downloaded the latest release package.
      {#if downloadedPath}
        It was saved to <span class="font-medium">{downloadedPath}</span>.
      {/if}
      Install or replace the app manually using your preferred update path.
    </p>
  </Modal>
{/if}

<style>
  .release-notes :global(h1) {
    margin: 0 0 0.75rem;
    font-size: 1.25rem;
    font-weight: 700;
    line-height: 1.4;
  }

  .release-notes :global(h2) {
    margin: 1rem 0 0.5rem;
    font-size: 1.05rem;
    font-weight: 650;
    line-height: 1.4;
  }

  .release-notes :global(h3) {
    margin: 0.875rem 0 0.5rem;
    font-size: 0.95rem;
    font-weight: 650;
  }

  .release-notes :global(p) {
    margin: 0.5rem 0;
    line-height: 1.55;
  }

  .release-notes :global(ul),
  .release-notes :global(ol) {
    margin: 0.5rem 0 0.75rem 1.25rem;
  }

  .release-notes :global(li) {
    margin: 0.25rem 0;
  }

  .release-notes :global(code) {
    border-radius: 0.25rem;
    background: rgba(115, 115, 115, 0.12);
    padding: 0.05rem 0.3rem;
    font-size: 0.85em;
  }

  .release-notes :global(pre) {
    overflow-x: auto;
    border-radius: 0.5rem;
    background: rgba(115, 115, 115, 0.12);
    padding: 0.65rem 0.8rem;
  }
</style>
