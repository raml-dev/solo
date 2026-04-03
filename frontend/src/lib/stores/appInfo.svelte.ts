/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: GPL-3.0-only
 */

import { GetAppInfo } from "$wails/go/main/App";
import type { appinfo } from "$wails/go/models";

export interface AppInfoState {
  initialized: boolean;
  loading: boolean;
  error: string | null;
  info: appinfo.AppInfo | null;
}

export const appInfoState: AppInfoState = $state({
  initialized: false,
  loading: false,
  error: null,
  info: null
});

let initStarted = false;

async function initializeAppInfo() {
  if (initStarted || appInfoState.loading || appInfoState.initialized) return;

  initStarted = true;
  appInfoState.loading = true;
  appInfoState.error = null;

  try {
    const info = await GetAppInfo();
    appInfoState.info = info;
  } catch (err) {
    appInfoState.error = String(err);
  } finally {
    appInfoState.loading = false;
    appInfoState.initialized = true;
  }
}

void initializeAppInfo();
