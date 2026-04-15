/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: AGPL-3.0-only
 */

import { writable } from "svelte/store";

interface TokenTooltipState {
  visible: boolean;
  tokenKey: string;
  x: number;
  y: number;
}

export const tokenTooltipStore = writable<TokenTooltipState>({
  visible: false,
  tokenKey: "",
  x: 0,
  y: 0
});

let hideTimeout: ReturnType<typeof setTimeout> | null = null;
export let mouseOverTooltip = false;

export function showTokenTooltip(tokenKey: string, x: number, y: number) {
  if (hideTimeout) {
    clearTimeout(hideTimeout);
    hideTimeout = null;
  }
  tokenTooltipStore.set({ visible: true, tokenKey, x, y });
}

export function hideTokenTooltipDelay() {
  if (mouseOverTooltip) return;
  if (hideTimeout) clearTimeout(hideTimeout);
  hideTimeout = setTimeout(() => {
    tokenTooltipStore.update((s) => ({ ...s, visible: false }));
    hideTimeout = null;
  }, 80);
}

export function cancelHideTokenTooltip() {
  mouseOverTooltip = true;
  if (hideTimeout) {
    clearTimeout(hideTimeout);
    hideTimeout = null;
  }
}

export function tooltipMouseLeave() {
  mouseOverTooltip = false;
  hideTokenTooltipDelay();
}

export function forceHideTokenTooltip() {
  mouseOverTooltip = false;
  if (hideTimeout) clearTimeout(hideTimeout);
  tokenTooltipStore.update((s) => ({ ...s, visible: false }));
}
