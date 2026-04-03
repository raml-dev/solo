<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: GPL-3.0-only
-->

<script lang="ts">
  import { appInfoState } from "$src/lib/stores/appInfo.svelte";
  import Alert from "flowbite-svelte/Alert.svelte";
  import Button from "flowbite-svelte/Button.svelte";
  import GithubSolid from "flowbite-svelte-icons/GithubSolid.svelte";
  import BookSolid from "flowbite-svelte-icons/BookSolid.svelte";
  import { BrowserOpenURL } from "$wails/runtime/runtime";

  const info = $derived(appInfoState.info);
  const docsLink = $derived((info?.docsLink || "").trim());
  const ghLink = $derived((info?.ghLink || "").trim());
  const orgLink = $derived((info?.orgLink || "").trim());

  const appNameAscii = `     █
▄▄▄  ▄▄▄  █  ▄▄▄
▀▄▄  █   █ █ █   █
▄▄▄▀ ▀▄▄▄▀ █ ▀▄▄▄▀`;

  function handleButtonClick(link: string) {
    BrowserOpenURL(link);
  }
</script>

<div class="flex flex-col items-center gap-4 text-center">
  <div>
    <h2 class="text-base font-semibold text-neutral-900 dark:text-neutral-100">About</h2>
  </div>

  {#if appInfoState.loading}
    <p class="text-sm text-neutral-500 dark:text-neutral-400">Loading application info…</p>
  {:else if appInfoState.error}
    <Alert color="red">
      Failed to load app info: {appInfoState.error}
    </Alert>
  {:else if info}
    <div class="flex w-full max-w-xl flex-col items-center gap-3 p-4">
      <div class="flex flex-col items-center gap-1">
        <div class="flex items-center gap-4">
          <h1 aria-label="solo" class="select-none">
            <pre
              aria-hidden="true"
              class="pointer-events-none cursor-default font-mono text-[0.35rem]/[1.3] text-primary-700 select-none dark:text-primary-500">{appNameAscii}</pre>
          </h1>
        </div>
        <p class="text-xs font-medium tracking-wide text-neutral-500 dark:text-neutral-400">
          version
          {#if ghLink}
            <button
              class="cursor-pointer underline"
              title={`${ghLink}/releases/tag/${info.productVersion}`}
              onclick={() => handleButtonClick(`${ghLink}/releases/tag/${info.productVersion}`)}
              >{info.productVersion || "-"}</button
            >
          {:else}
            {info.productVersion || "-"}
          {/if}
        </p>
      </div>

      <div class="flex flex-col items-center gap-2">
        <p class="text-xs font-medium tracking-wide text-neutral-500 dark:text-neutral-400">
          Licensed under the {info.license || "-"}
        </p>
      </div>
      <div class="flex flex-col items-center gap-2">
        <p
          class="text-xs font-medium tracking-wide text-neutral-500 uppercase dark:text-neutral-400"
        >
          {#if ghLink}
            <Button onclick={() => handleButtonClick(ghLink)} target="_blank">
              Source code <GithubSolid class="ms-2 h-5 w-5" />
            </Button>
          {/if}
          {#if docsLink}
            <Button onclick={() => handleButtonClick(docsLink)} target="_blank">
              Docs <BookSolid class="ms-2 h-5 w-5" />
            </Button>
          {/if}
        </p>
      </div>
    </div>
    <div class="flex flex-col items-center gap-1">
      <p class="text-sm text-neutral-900 dark:text-neutral-100">
        Built with ❤️ by <button
          class="cursor-pointer underline"
          title={orgLink}
          onclick={() => handleButtonClick(orgLink)}>{info.companyName || "raml-dev"}</button
        >
      </p>
    </div>
  {:else}
    <p class="text-sm text-neutral-500 dark:text-neutral-400">Application info not available.</p>
  {/if}
</div>
