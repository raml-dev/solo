/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: GPL-3.0-only
 */

import { HTTP_METHOD_COLOR_MAP } from "$src/lib/theme/themeModel";

export type HttpMethod = "GET" | "POST" | "PUT" | "PATCH" | "DELETE" | "HEAD" | "OPTIONS";

/**
 * Status Badge Colors (Flowbite native colors)
 */
export type StatusBadgeColor = "green" | "blue" | "yellow" | "red";

type HTTP_STATUS =
  | 100
  | 101
  | 102
  | 103
  | 200
  | 201
  | 202
  | 203
  | 204
  | 205
  | 206
  | 207
  | 208
  | 226
  | 300
  | 301
  | 302
  | 303
  | 304
  | 305
  | 307
  | 308
  | 400
  | 401
  | 402
  | 403
  | 404
  | 405
  | 406
  | 407
  | 408
  | 409
  | 410
  | 411
  | 412
  | 413
  | 414
  | 415
  | 416
  | 417
  | 418
  | 421
  | 422
  | 423
  | 424
  | 425
  | 426
  | 428
  | 429
  | 431
  | 451
  | 500
  | 501
  | 502
  | 503
  | 504
  | 505
  | 506
  | 507
  | 508
  | 510
  | 511;

export const HTTP_STATUS_STRINGS: Record<HTTP_STATUS, string> = {
  100: "Continue",
  101: "Switching Protocols",
  102: "Processing",
  103: "Early Hints",
  200: "OK",
  201: "Created",
  202: "Accepted",
  203: "Non-Authoritative Information",
  204: "No Content",
  205: "Reset Content",
  206: "Partial Content",
  207: "Multi-Status",
  208: "Already Reported",
  226: "IM Used",
  300: "Multiple Choices",
  301: "Moved Permanently",
  302: "Found",
  303: "See Other",
  304: "Not Modified",
  305: "Use Proxy",
  307: "Temporary Redirect",
  308: "Permanent Redirect",
  400: "Bad Request",
  401: "Unauthorized",
  402: "Payment Required",
  403: "Forbidden",
  404: "Not Found",
  405: "Method Not Allowed",
  406: "Not Acceptable",
  407: "Proxy Authentication Required",
  408: "Request Timeout",
  409: "Conflict",
  410: "Gone",
  411: "Length Required",
  412: "Precondition Failed",
  413: "Content Too Large",
  414: "URI Too Long",
  415: "Unsupported Media Type",
  416: "Range Not Satisfiable",
  417: "Expectation Failed",
  418: "I'm a Teapot",
  421: "Misdirected Request",
  422: "Unprocessable Content",
  423: "Locked",
  424: "Failed Dependency",
  425: "Too Early",
  426: "Upgrade Required",
  428: "Precondition Required",
  429: "Too Many Requests",
  431: "Request Header Fields Too Large",
  451: "Unavailable For Legal Reasons",
  500: "Internal Server Error",
  501: "Not Implemented",
  502: "Bad Gateway",
  503: "Service Unavailable",
  504: "Gateway Timeout",
  505: "HTTP Version Not Supported",
  506: "Variant Also Negotiates",
  507: "Insufficient Storage",
  508: "Loop Detected",
  510: "Not Extended",
  511: "Network Authentication Required"
}

export function getHttpStatusString(code: number): string {
  if (code in HTTP_STATUS_STRINGS) {
    return `${code} ${HTTP_STATUS_STRINGS[code as HTTP_STATUS]}`;
  }
  // fallback to return only status code if not mapped
  return `${code}`;
}

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
  const base =
    "inline-flex min-w-14 justify-center rounded px-1.5 py-0.5 text-[10px] font-semibold uppercase";
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
export type MethodBadgeColor =
  | "green"
  | "blue"
  | "yellow"
  | "red"
  | "dark"
  | "none"
  | "purple"
  | "indigo"
  | "pink";

export function getMethodBadgeColor(verb: string): MethodBadgeColor {
  const family = HTTP_METHOD_COLOR_MAP[(verb || "GET").toUpperCase() as HttpMethod] || "neutral";

  switch (family) {
    case "success":
      return "green";
    case "primary":
      return "blue";
    case "warning":
      return "yellow";
    case "danger":
      return "red";
    default:
      return "dark";
  }
}

export interface RequestHeaderRow {
  key: string;
  value: string;
  enabled: boolean;
}

interface BuildResolvedRequestPayloadOptions {
  body: string;
  headers: RequestHeaderRow[];
  resolveTokens: (value: string) => string;
}

export function buildResolvedRequestPayload({
  body,
  headers,
  resolveTokens
}: BuildResolvedRequestPayloadOptions): {
  body: string;
  headers: Record<string, string>;
} {
  const resolvedBody = resolveTokens(body);
  const resolvedHeaders = headers
    .filter((header) => header.enabled)
    .reduce(
      (acc, { key, value }) => ({
        ...acc,
        [resolveTokens(key)]: resolveTokens(value)
      }),
      {} as Record<string, string>
    );

  return {
    body: resolvedBody,
    headers: resolvedHeaders
  };
}
