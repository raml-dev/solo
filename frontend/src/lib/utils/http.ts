import { HTTP_METHOD_COLOR_MAP, type HttpMethod } from "$src/lib/theme/themeModel";

/**
 * Status Badge Colors (Flowbite native colors)
 */
export type StatusBadgeColor = "green" | "blue" | "yellow" | "red";

/**
 * Returns the Flowbite badge color based on the HTTP status code.
 * @param status The HTTP status code.
 * @returns The color for the Badge component.
 */
export function getStatusBadgeColor(status: number): StatusBadgeColor {
  if (status >= 200 && status < 300) return "green";
  if (status >= 300 && status < 400) return "blue";
  if (status >= 400 && status < 500) return "yellow";
  return "red";
}

/**
 * Method Badge Styling (Tailwind-based, matching CollectionList)
 */

/**
 * Returns the exact Tailwind class string for HTTP method badges.
 * This ensures consistency between CollectionList, RequestTabBar, and other views.
 * @param method The HTTP method (GET, POST, etc.)
 * @returns CSS class string
 */
export function getMethodBadgeClass(method: string): string {
  const base = "inline-flex min-w-14 justify-center rounded px-1.5 py-0.5 text-[10px] font-semibold uppercase";
  const m = (method || "GET").toUpperCase();
  let colorClass = "bg-neutral-100 text-neutral-700 dark:bg-neutral-800 dark:text-neutral-300";

  if (m === "GET")
    colorClass = "bg-success-100 text-success-700 dark:bg-success-900/50 dark:text-success-300";
  else if (m === "POST")
    colorClass = "bg-primary-100 text-primary-700 dark:bg-primary-900/50 dark:text-primary-300";
  else if (m === "PUT" || m === "PATCH")
    colorClass = "bg-warning-100 text-warning-700 dark:bg-warning-900/50 dark:text-warning-300";
  else if (m === "DELETE")
    colorClass = "bg-danger-100 text-danger-700 dark:bg-danger-900/50 dark:text-danger-300";

  return `${base} ${colorClass}`;
}

/**
 * Legacy Method Badge Colors (Flowbite native colors)
 * Kept for backward compatibility if needed, but getMethodBadgeClass is preferred for consistency.
 */
export type MethodBadgeColor = "green" | "blue" | "yellow" | "red" | "dark" | "none" | "purple" | "indigo" | "pink";

export function getMethodBadgeColor(verb: string): MethodBadgeColor {
  const family = HTTP_METHOD_COLOR_MAP[(verb || "GET").toUpperCase() as HttpMethod] || "neutral";
  
  switch (family) {
    case "success": return "green";
    case "primary": return "blue";
    case "warning": return "yellow";
    case "danger":  return "red";
    default:        return "dark";
  }
}
