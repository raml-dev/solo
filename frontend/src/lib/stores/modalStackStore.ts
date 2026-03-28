/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: GPL-3.0-only
 */

import { derived, writable, type Readable, type Writable } from "svelte/store";

const stack = writable<string[]>([]);

function open(id: string) {
  stack.update((current) => (current.includes(id) ? current : [...current, id]));
}

function close(id: string) {
  stack.update((current) => current.filter((item) => item !== id));
}

function isOpen(id: string): Readable<boolean> {
  return derived(stack, ($stack) => $stack.includes(id));
}

function binding(id: string): Writable<boolean> {
  const openStore = isOpen(id);

  const set = (value: boolean) => {
    if (value) {
      open(id);
    } else {
      close(id);
    }
  };

  return {
    subscribe: openStore.subscribe,
    set,
    update(updater) {
      let current = false;
      const unsubscribe = openStore.subscribe((value) => {
        current = value;
      });
      unsubscribe();
      set(updater(current));
    }
  };
}

export const topModalId = derived(stack, ($stack) => $stack[$stack.length - 1] ?? null);
export const hasOpenModals = derived(stack, ($stack) => $stack.length > 0);

export const modalStack = {
  open,
  close,
  isOpen,
  binding
};
