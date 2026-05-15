/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: AGPL-3.0-only
 */

import { MAX_REQUESTS_BEFORE_TRIM } from "$src/lib/utils/constants";
import type { collection } from "$wails/go/models";

/**
 * Filters the array _in place_ to only retain elements matching the given predicate.
 *
 * Similar in effect to the native `Array.filter` function, but in place.
 *
 * @param array the array to filter
 * @param predicate the predicate to match elements to be filtered
 */
export function filterInPlace<T>(array: Array<T>, predicate: (element: T) => boolean) {
  for (let i = array.length - 1; i >= 0; i--) {
    if (!predicate(array[i])) array.splice(i, 1);
  }
}

/**
 * Truncates a string to a set number of characters if it is longer, addind an ellipsis if that's the case.
 *
 * @param str the string to truncate
 * @param max max characters to be shown
 * @returns the string truncated to the max number of characters specified
 */
export function truncateString(url: string, max = 60): string {
  return url.length > max ? url.slice(0, max) + "..." : url;
}

/**
 * Formats a timestamp string using JS objects
 * @param ts the timestamp to be formatted
 * @returns a string representation of the parsed Date object from the timestamp
 */
export function formatTime(ts: string): string {
  const d = new Date(ts);
  return d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", second: "2-digit" });
}

export function clampNumberToMax(numberToClamp: number) {
  if (numberToClamp > MAX_REQUESTS_BEFORE_TRIM) return `${MAX_REQUESTS_BEFORE_TRIM}+`;
  return String(numberToClamp);
}

export function getTotalRequestCount(
  collectionOrFolder: collection.Collection | collection.Folder
): number {
  return (
    collectionOrFolder.requests.length ||
    0 +
      collectionOrFolder.folders.reduce(
        (count: number, folder: collection.Folder) =>
          count + (folder.requests?.length || 0) + getTotalRequestCount(folder),
        0
      )
  );
}
