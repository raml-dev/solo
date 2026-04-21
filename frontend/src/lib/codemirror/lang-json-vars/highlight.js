/**
 * Original work Copyright (C) 2021 by Marijn Haverbeke and others
 * SPDX-License-Identifier: MIT
 *
 * Adapted for Solo to add VariableRef highlighting while preserving the
 * upstream JSON highlight mapping.
 */

import { styleTags, tags as t } from "@lezer/highlight";

export const jsonHighlighting = styleTags({
  String: t.string,
  Number: t.number,
  "True False": t.bool,
  PropertyName: t.propertyName,
  Null: t.null,
  VariableRef: t.special(t.variableName),
  ", :": t.separator,
  "[ ]": t.squareBracket,
  "{ }": t.brace
});
