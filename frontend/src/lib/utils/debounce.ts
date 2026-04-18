/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: AGPL-3.0-only
 */

export type DebouncedFunction<T extends (...args: unknown[]) => unknown> = ((
  ...args: Parameters<T>
) => void) & {
  cancel: () => void;
};

export function debounce<T extends (...args: unknown[]) => unknown>(
  fn: T,
  ms: number
): DebouncedFunction<T> {
  let timer: ReturnType<typeof setTimeout> | null = null;

  const debounced = ((...args: Parameters<T>) => {
    if (timer !== null) {
      clearTimeout(timer);
    }
    timer = setTimeout(() => {
      timer = null;
      fn(...args);
    }, ms);
  }) as DebouncedFunction<T>;

  debounced.cancel = () => {
    if (timer === null) {
      return;
    }

    clearTimeout(timer);
    timer = null;
  };

  return debounced;
}
