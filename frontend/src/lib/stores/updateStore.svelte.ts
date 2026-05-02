/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: AGPL-3.0-only
 */

import { configurationStoreState } from "$src/lib/stores/configurationStore.svelte";
import { notifications } from "$src/lib/stores/notificationStore";
import { GetAppInfo, GetUpdatesFromRepo } from "$wails/go/main/App";
import { appinfo } from "$wails/go/models";
import { BrowserOpenURL } from "$wails/runtime/runtime";

interface UpdateStoreState {
  initialized: boolean;
  visible: boolean;
  loading: boolean;
  ignoredVersion: string;
  currentVersion: string;
  appInfo: appinfo.AppInfo | null;
  updateInfo: appinfo.GitHubResponse | null;
  selectedRelease: appinfo.GitHubRelease | null;
  downloadedPath: string;
  downloadCompleteOpen: boolean;
}

const initialState: UpdateStoreState = {
  initialized: false,
  visible: false,
  loading: false,
  ignoredVersion: "",
  currentVersion: "",
  appInfo: null,
  updateInfo: null,
  selectedRelease: null,
  downloadedPath: "",
  downloadCompleteOpen: false
};

export const updateStoreState = $state<UpdateStoreState>({ ...initialState });

function getVersion(release: appinfo.GitHubRelease | null | undefined): string {
  return (release?.tag_name || release?.name || "").trim();
}

function getReleasePageURL(): string {
  const tagName = getVersion(updateStoreState.selectedRelease);
  const ghLink = (updateStoreState.appInfo?.ghLink || "").replace(/\/+$/, "");

  if (!tagName || !ghLink) {
    return "";
  }

  return `${ghLink}/releases/tag/${tagName}`;
}

function applyAvailableUpdate(payload: unknown) {
  const safePayload = payload && typeof payload === "object" ? payload : {};
  const releaseInfo = appinfo.GitHubResponse.createFrom(safePayload);
  if (!releaseInfo?.Release) {
    updateStoreState.updateInfo = null;
    updateStoreState.selectedRelease = null;
    updateStoreState.visible = false;
    return;
  }

  const version = getVersion(releaseInfo.Release);
  if (!version || updateStoreState.ignoredVersion === version) {
    updateStoreState.updateInfo = releaseInfo;
    updateStoreState.selectedRelease = releaseInfo.Release;
    updateStoreState.visible = false;
    return;
  }

  updateStoreState.updateInfo = releaseInfo;
  updateStoreState.selectedRelease = releaseInfo.Release;
  updateStoreState.visible = true;
}

function resetUpdateState() {
  updateStoreState.visible = false;
  updateStoreState.loading = false;
  updateStoreState.updateInfo = null;
  updateStoreState.selectedRelease = null;
}

export const updateStore = {
  async init() {
    if (updateStoreState.initialized) {
      return;
    }

    updateStoreState.initialized = true;

    try {
      const info = await GetAppInfo();
      updateStoreState.appInfo = new appinfo.AppInfo(info);
      updateStoreState.currentVersion = updateStoreState.appInfo.productVersion ?? "";

      if (configurationStoreState.config.general.checkForUpdates) {
        await this.refresh();
      }
    } catch (err) {
      updateStoreState.currentVersion = "";
      notifications.error("Failed to initialize updates", String(err));
    }
  },

  async refresh() {
    try {
      const latestUpdate = await GetUpdatesFromRepo();
      applyAvailableUpdate(latestUpdate);
    } catch (err) {
      notifications.error("Failed to check for updates", String(err));
    }
  },

  ignoreCurrentRelease() {
    if (!updateStoreState.selectedRelease) {
      return;
    }

    updateStoreState.ignoredVersion = getVersion(updateStoreState.selectedRelease);
    updateStoreState.visible = false;
  },

  openReleasePage() {
    const releaseURL = getReleasePageURL();
    if (!releaseURL) {
      return;
    }

    BrowserOpenURL(releaseURL);
  },

  async syncWithConfiguration() {
    if (!configurationStoreState.config.general.checkForUpdates) {
      resetUpdateState();
      return;
    }

    await this.refresh();
  }
};
