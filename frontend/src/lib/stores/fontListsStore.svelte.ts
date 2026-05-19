/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: AGPL-3.0-only
 */

import { splitSystemFonts, type FontLists } from "$src/lib/fonts";
import { notifications } from "$src/lib/stores/notificationStore";
import { ListSystemFonts } from "$wails/go/main/App";
import type { fonts } from "$wails/go/models";

export interface FontListsStoreState {
  lists: FontLists;
  loading: boolean;
  ready: boolean;
  error: string | null;
}

export const fontListsStoreState: FontListsStoreState = $state({
  lists: { allDefault: [], allMono: [] },
  loading: false,
  ready: false,
  error: null
});

async function loadFontLists(refresh: boolean) {
  const systemFonts = (await ListSystemFonts(refresh)) || ([] as fonts.SystemFont[]);
  return splitSystemFonts(systemFonts);
}

async function load(refresh: boolean) {
  if (fontListsStoreState.loading) return;

  fontListsStoreState.loading = true;
  fontListsStoreState.error = null;

  try {
    fontListsStoreState.lists = await loadFontLists(refresh);
    fontListsStoreState.ready = true;
  } catch (error) {
    fontListsStoreState.error = String(error);
    notifications.error("Could not load system fonts", String(error));
  } finally {
    fontListsStoreState.loading = false;
  }
}

export const fontListsStore = {
  async init() {
    if (fontListsStoreState.ready || fontListsStoreState.loading) return;
    await load(false);
  },

  async refresh() {
    await load(true);
  }
};
