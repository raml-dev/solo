<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import SoloLightning from "$src/lib/components/common/SoloLightning.svelte";
  import SoloSvg from "$src/lib/components/common/SoloSvg.svelte";
  import { appInfoState } from "$src/lib/stores/appInfo.svelte";
  import { BrowserOpenURL } from "$wails/runtime/runtime";
  import BookSolid from "flowbite-svelte-icons/BookSolid.svelte";
  import GithubSolid from "flowbite-svelte-icons/GithubSolid.svelte";
  import Alert from "flowbite-svelte/Alert.svelte";
  import Button from "flowbite-svelte/Button.svelte";

  const info = $derived(appInfoState.info);
  const docsLink = $derived((info?.docsLink || "").trim());
  const ghLink = $derived((info?.ghLink || "").trim());
  const orgLink = $derived((info?.orgLink || "").trim());

  function handleButtonClick(link: string) {
    BrowserOpenURL(link);
  }
</script>

<div class="flex flex-col items-center gap-4 text-center">
  <div>
    <h2 class="text-base font-semibold text-neutral-900 dark:text-neutral-100">About</h2>
  </div>

  {#if appInfoState.loading}
    <p class="text-sm text-neutral-500 dark:text-neutral-400">Loading application info...</p>
  {:else if appInfoState.error}
    <Alert color="red">
      Failed to load app info: {appInfoState.error}
    </Alert>
  {:else if info}
    <div class="flex w-full max-w-xl flex-col items-center gap-3 p-4">
      <div class="flex flex-col items-center gap-1">
        <div class="flex items-center gap-4">
          <h1 aria-label="Solo" class="flex flex-row items-baseline gap-4 select-none">
            <SoloSvg
              aria-hidden="true"
              size={52}
              class="pointer-events-none cursor-default text-primary-700 select-none dark:text-primary-500"
            /><SoloLightning
              aria-hidden="true"
              size={22}
              class="pointer-events-none cursor-default text-primary-700 select-none dark:text-primary-500"
            />
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
