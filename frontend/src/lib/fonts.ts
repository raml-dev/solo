/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: AGPL-3.0-only
 */

import type { fonts } from "$wails/go/models";

export interface FontLists {
  allDefault: string[];
  allMono: string[];
}

function sortFamilies(families: Set<string>) {
  return [...families].sort((a, b) => a.localeCompare(b));
}

export function splitSystemFonts(systemFonts: fonts.SystemFont[]): FontLists {
  const allDefault = new Set<string>();
  const allMono = new Set<string>();

  for (const font of systemFonts) {
    const family = font.family?.trim();
    if (!family) continue;

    const allTarget = font.isMonospace ? allMono : allDefault;

    allTarget.add(family);
  }

  return {
    // use all available fonts for the default font selector
    allDefault: sortFamilies(new Set([...allDefault, ...allMono])),
    allMono: sortFamilies(allMono)
  };
}

export function sanitizeFontFamily(family: string | undefined) {
  return (family ?? "").replaceAll('"', "").replaceAll("'", "").trim();
}
