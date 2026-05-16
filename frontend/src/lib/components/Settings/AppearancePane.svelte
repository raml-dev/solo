<!--
 Copyright 2026-present raml-dev
 SPDX-License-Identifier: AGPL-3.0-only
-->

<script lang="ts">
  import type { FontLists } from "$src/lib/fonts";
  import type { ThemeMode } from "$src/lib/theme/themeModel";
  import { DEFAULT_ZOOM_LEVEL, ZOOM_LEVEL_OPTIONS } from "$src/lib/utils/constants";
  import type { theme } from "$wails/go/models";
  import Button from "flowbite-svelte/Button.svelte";
  import Helper from "flowbite-svelte/Helper.svelte";
  import Label from "flowbite-svelte/Label.svelte";
  import FontFamilySelect from "$src/lib/components/Settings/FontFamilySelect.svelte";
  import ThemeModeSelect from "$src/lib/components/Settings/ThemeModeSelect.svelte";
  import ColorSchemeSelect from "$src/lib/components/Settings/ColorSchemeSelect.svelte";
  import ZoomLevelSelect from "$src/lib/components/Settings/ZoomLevelSelect.svelte";
  import { getZoomPercent } from "$src/lib/stores/zoomStore.svelte";

  interface Props {
    themeMode: ThemeMode;
    activeTheme: string;
    defaultFontFamily: string;
    monoFontFamily: string;
    themes: theme.Theme[];
    fontLists: FontLists;
    fontsLoading: boolean;
    fontsReady: boolean;
    fontsError: string | null;
    zoomLevel: number;
    onThemeModeChange: (mode: ThemeMode) => void;
    onThemeSelect: (themeId: string) => void;
    onSansFontChange: (fontFamily: string) => void;
    onMonoFontChange: (fontFamily: string) => void;
    onZoomLevelChange: (level: number) => void | Promise<void>;
    onReset: () => void | Promise<void>;
  }

  let {
    themeMode,
    activeTheme,
    defaultFontFamily,
    monoFontFamily,
    themes,
    fontLists,
    fontsLoading,
    fontsReady,
    fontsError,
    zoomLevel,
    onThemeModeChange,
    onThemeSelect,
    onSansFontChange,
    onMonoFontChange,
    onZoomLevelChange,
    onReset
  }: Props = $props();
</script>

<div class="flex flex-col gap-4">
  <div>
    <h2 class="text-base font-semibold text-neutral-900 dark:text-neutral-100">Appearance</h2>
    <p class="mt-1 text-sm text-neutral-500 dark:text-neutral-400">
      Personalize your experience to match your style.
    </p>
  </div>

  <div class="flex flex-col gap-4 rounded-lg border border-neutral-200 p-4 dark:border-neutral-700">
    <div>
      <h3 class="text-sm font-semibold text-neutral-700 dark:text-neutral-300">Themes</h3>
      <p class="mt-1 text-sm text-neutral-500 dark:text-neutral-400">
        Pick how Solo looks and which color scheme it uses.
      </p>
    </div>

    <div class="grid gap-3 md:grid-cols-2">
      <div class="flex min-w-0 flex-col gap-1">
        <Label for="theme-mode-select">Display mode</Label>
        <ThemeModeSelect id="theme-mode-select" value={themeMode} onchange={onThemeModeChange} />
      </div>

      <div class="flex min-w-0 flex-col gap-1">
        <Label for="color-scheme-select">Color scheme</Label>
        <ColorSchemeSelect
          id="color-scheme-select"
          value={activeTheme}
          {themes}
          onchange={onThemeSelect}
        />
      </div>
    </div>
  </div>

  <div class="flex flex-col gap-4 rounded-lg border border-neutral-200 p-4 dark:border-neutral-700">
    <div>
      <h3 class="text-sm font-semibold text-neutral-700 dark:text-neutral-300">Interface</h3>
      <p class="mt-1 text-sm text-neutral-500 dark:text-neutral-400">
        Adjust the interface zoom and choose custom sans and monospace fonts.
      </p>
    </div>

    <div class="grid grid-cols-2 gap-3">
      <div class="flex flex-col gap-1">
        <Label for="sans-font-family">Interface font</Label>
        <FontFamilySelect
          id="sans-font-family"
          value={defaultFontFamily}
          families={fontLists.allDefault}
          placeholder="Default interface font"
          searchPlaceholder="Search interface fonts"
          previewKind="sans"
          disabled={!fontsReady}
          onchange={(fontFamily) => onSansFontChange(fontFamily)}
        />
      </div>

      <div class="flex flex-col gap-1">
        <Label for="mono-font-family">Monospace font</Label>
        <FontFamilySelect
          id="mono-font-family"
          value={monoFontFamily}
          families={fontLists.allMono}
          placeholder="Default monospace font"
          searchPlaceholder="Search monospace fonts"
          previewKind="mono"
          disabled={!fontsReady}
          onchange={(fontFamily) => onMonoFontChange(fontFamily)}
        />
      </div>
    </div>

    <div class="flex max-w-xs flex-col gap-1">
      <Label for="zoom-level-select">Zoom level</Label>
      <ZoomLevelSelect
        id="zoom-level-select"
        value={zoomLevel}
        options={ZOOM_LEVEL_OPTIONS}
        onchange={onZoomLevelChange}
      />
    </div>

    {#if fontsLoading}
      <Helper color="gray">Loading system fonts...</Helper>
    {:else if fontsError}
      <Helper color="red">Could not load system fonts.</Helper>
    {/if}

    <div class="flex justify-end">
      <Button color="light" onclick={onReset}>
        Reset fonts and zoom ({getZoomPercent(DEFAULT_ZOOM_LEVEL)}%)
      </Button>
    </div>
  </div>
</div>
