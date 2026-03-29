/**
 * Removes elements matching the given predicate in the array _in place_.
 *
 * Similar in effect to the native `Array.filter` function, but in place.
 *
 * @param array the array to filter
 * @param predicate the predicate to match elements to be filtered
 */
export function filterInPlace<T>(array: Array<T>, predicate: (element: T) => boolean) {
  for (let i = array.length - 1; i >= 0; i--) {
    if (predicate(array[i])) array.splice(i, 1);
  }
}
