/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: AGPL-3.0-only
 */

import { LogDebug, LogError, LogInfo, LogWarning } from "$wails/runtime/runtime";

const _console = {
  debug: console.debug.bind(console),
  log: console.log.bind(console),
  warn: console.warn.bind(console),
  error: console.error.bind(console)
};

function serialize(arg: unknown) {
  if (arg === null) return "null";
  if (arg === undefined) return "undefined";
  if (typeof arg !== "object" && typeof arg !== "function") return String(arg);
  if (arg instanceof Error) return arg.stack ?? `${arg.name}: ${arg.message}`;
  try {
    return JSON.stringify(arg);
  } catch {
    return Object.prototype.toString.call(arg);
  }
}

function serializeArgs(args: unknown[]) {
  return args.map(serialize).join(" ");
}

function forward(
  goFn: (msg: string) => void,
  nativeFn: (...data: unknown[]) => void,
  args: unknown[]
) {
  nativeFn(...args);
  goFn(serializeArgs(args));
}

export default function initLogger() {
  console.debug = (...args) => forward(LogDebug, _console.debug, args);
  console.log = (...args) => forward(LogInfo, _console.log, args);
  console.warn = (...args) => forward(LogWarning, _console.warn, args);
  console.error = (...args) => forward(LogError, _console.error, args);
}
