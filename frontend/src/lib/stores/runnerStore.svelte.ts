/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: AGPL-3.0-only
 */

import { CancelRunner, RunParallel } from "$wails/go/main/App";
import type { configuration as conf } from "$wails/go/models";
import { main, runner } from "$wails/go/models";
import { EventsOff, EventsOn } from "$wails/runtime";

export interface RunnerStoreState {
  running: boolean;
  progress: number;
  stats: runner.RunnerStats | null;
  lastResults: runner.RunnerResult[];
}

export interface RunnerStartOptions {
  method: string;
  url: string;
  body: string;
  headers: Record<string, string>;
  collectionName: string;
  settings: conf.RequestSettingsOverride;
  preRequestScript: string;
  postResponseScript: string;
  concurrency: number;
  iterations: number;
  stopOnError: boolean;
}

const MAX_VISIBLE_RESULTS = 50;

export const runnerStoreState: RunnerStoreState = $state({
  running: false,
  progress: 0,
  stats: null,
  lastResults: []
});

export const runnerStore = {
  async startRun(opts: RunnerStartOptions): Promise<void> {
    if (runnerStoreState.running) return;

    runnerStoreState.running = true;
    runnerStoreState.progress = 0;
    runnerStoreState.stats = null;
    runnerStoreState.lastResults = [];

    const requestOptions = new main.RequestOptions({
      body: opts.body,
      headers: opts.headers,
      method: opts.method,
      url: opts.url,
      collectionName: opts.collectionName,
      settings: opts.settings,
      preRequestScript: opts.preRequestScript || "",
      postResponseScript: opts.postResponseScript || ""
    });

    try {
      EventsOn("runner:result", (res: runner.RunnerResult) => {
        runnerStoreState.lastResults = [res, ...runnerStoreState.lastResults].slice(
          0,
          MAX_VISIBLE_RESULTS
        );
        runnerStoreState.progress = Math.min(
          100,
          (runnerStoreState.lastResults.length / opts.iterations) * 100
        );
      });

      runnerStoreState.stats = await RunParallel(
        requestOptions,
        opts.concurrency,
        opts.iterations,
        opts.stopOnError
      );
    } catch (err) {
      console.error("Runner failed", err);
    } finally {
      runnerStoreState.running = false;
      EventsOff("runner:result");
    }
  },

  async stopRun(): Promise<void> {
    try {
      await CancelRunner();
    } catch (err) {
      console.error("Failed to stop runner", err);
    }
  }
};
