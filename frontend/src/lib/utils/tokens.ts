/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: GPL-3.0-only
 */

export interface TextSegment {
  text: string;
  isToken: boolean;
  tokenKey?: string;
}

export function splitTextSegments(value: string): TextSegment[] {
  if (!value) return [];

  const segments: TextSegment[] = [];
  const tokenRegex = /(\{\{([^{}\r\n]+?)\}\})/g;
  let cursor = 0;

  for (const match of value.matchAll(tokenRegex)) {
    const index = match.index ?? 0;
    const fullToken = match[0];
    const tokenKey = match[2].trim();

    if (index > cursor) {
      segments.push({ text: value.slice(cursor, index), isToken: false });
    }

    segments.push({ text: fullToken, isToken: true, tokenKey });
    cursor = index + fullToken.length;
  }

  if (cursor < value.length) {
    segments.push({ text: value.slice(cursor), isToken: false });
  }

  return segments;
}
