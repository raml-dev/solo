/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: GPL-3.0-only
 */

import { derived, writable } from "svelte/store";

const stack = writable<string[]>([]);

function open(id: string) {
  stack.update((current) => (current.includes(id) ? current : [...current, id]));
}

function close(id: string) {
  stack.update((current) => current.filter((item) => item !== id));
}

export const topModalId = derived(stack, ($stack) => $stack[$stack.length - 1] ?? null);
export const hasOpenModals = derived(stack, ($stack) => $stack.length > 0);

export const modalStack = {
  open,
  close
};
