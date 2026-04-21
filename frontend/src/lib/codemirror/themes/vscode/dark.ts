/**
 * This file is excerpted from @fsegurai/codemirror-theme-vscode-dark@6.2.6
 * Copyright 2026 fsegurai
 *
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: AGPL-3.0-only AND MIT
 */

import { HighlightStyle } from "@codemirror/language";
import { tags } from "@lezer/highlight";

/**
 * Enhanced VSCode Dark theme color definitions
 * --------------------------------------------
 * Colors organized by function with visual color blocks
 */
// Base colors
const base03 = "#838383", // Comments, invisibles
  base05 = "#d4d4d4", // Default foreground
  // Accent colors
  base08 = "#569cd6", // Keywords, storage
  base09 = "#c586c0", // Control keywords, operators
  base0A = "#9cdcfe", // Variables, parameters
  base0B = "#4ec9b0", // Classes, types
  base0C = "#dcdcaa", // Functions, attributes
  base0D = "#b5cea8", // Numbers, constants
  base0E = "#ce9178", // Strings
  base0F = "#f44747", // Errors, invalid
  base10 = "#d7ba7d", // Modified elements
  base11 = "#6a9955"; // Comments

// UI specific colors
const invalid = base0F, // Line highlight with transparency
  linkColor = "#3794ff", // Link color
  visitedLinkColor = "#c586c0"; // Visited link color

export const vsCodeDarkHighlightStyle = /*@__PURE__*/ HighlightStyle.define([
  // Keywords and control flow
  { tag: tags.keyword, color: base08, fontWeight: "bold" },
  { tag: tags.controlKeyword, color: base09, fontWeight: "bold" },
  { tag: tags.moduleKeyword, color: base08, fontWeight: "bold" },
  // Names and variables
  { tag: [tags.name, tags.deleted, tags.character, tags.macroName], color: base05 },
  { tag: [tags.variableName], color: base0A },
  { tag: [tags.propertyName], color: base0A, fontStyle: "normal" },
  // Classes and types
  { tag: [tags.typeName], color: base0B },
  { tag: [tags.className], color: base0B, fontStyle: "normal" },
  { tag: [tags.namespace], color: base05, fontStyle: "normal" },
  // Operators and punctuation
  { tag: [tags.operator, tags.operatorKeyword], color: base05 },
  { tag: [tags.bracket], color: base05 },
  { tag: [tags.brace], color: base05 },
  { tag: [tags.punctuation], color: base05 },
  // Functions and parameters
  { tag: [/*@__PURE__*/ tags.function(tags.variableName)], color: base0C },
  { tag: [tags.labelName], color: base0C, fontStyle: "normal" },
  {
    tag: [/*@__PURE__*/ tags.definition(/*@__PURE__*/ tags.function(tags.variableName))],
    color: base0C
  },
  { tag: [/*@__PURE__*/ tags.definition(tags.variableName)], color: base0A },
  // Constants and literals
  { tag: tags.number, color: base0D },
  { tag: tags.changed, color: base10 },
  { tag: tags.annotation, color: base10, fontStyle: "italic" },
  { tag: tags.modifier, color: base08, fontStyle: "normal" },
  { tag: tags.self, color: base08 },
  {
    tag: [
      tags.color,
      /*@__PURE__*/ tags.constant(tags.name),
      /*@__PURE__*/ tags.standard(tags.name)
    ],
    color: base0A
  },
  { tag: [tags.atom, tags.bool, /*@__PURE__*/ tags.special(tags.variableName)], color: base08 },
  // Strings and regex
  { tag: [tags.processingInstruction, tags.inserted], color: base0E },
  { tag: [/*@__PURE__*/ tags.special(tags.string), tags.regexp], color: "#d16969" },
  { tag: tags.string, color: base0E },
  // Punctuation and structure
  { tag: /*@__PURE__*/ tags.definition(tags.typeName), color: base0B, fontWeight: "bold" },
  { tag: [/*@__PURE__*/ tags.definition(tags.name), tags.separator], color: base05 },
  // Comments and documentation
  { tag: tags.meta, color: base03 },
  { tag: tags.comment, fontStyle: "italic", color: base11 },
  { tag: tags.docComment, fontStyle: "italic", color: base11 },
  // HTML/XML elements
  { tag: [tags.tagName], color: base08 },
  { tag: [tags.attributeName], color: base0A },
  // Markdown and text formatting
  { tag: [tags.heading], fontWeight: "bold", color: base08 },
  { tag: tags.heading1, color: base08, fontWeight: "bold" },
  { tag: tags.heading2, color: base08 },
  { tag: tags.heading3, color: base08 },
  { tag: tags.heading4, color: base08 },
  { tag: tags.heading5, color: base08 },
  { tag: tags.heading6, color: base08 },
  { tag: [tags.strong], fontWeight: "bold", color: base08 },
  { tag: [tags.emphasis], fontStyle: "italic", color: base0B },
  // Links and URLs
  {
    tag: [tags.link],
    color: visitedLinkColor,
    textDecoration: "underline",
    textUnderlinePosition: "under"
  },
  {
    tag: [tags.url],
    color: linkColor,
    textDecoration: "underline",
    textUnderlineOffset: "2px"
  },
  // Special states
  {
    tag: [tags.invalid],
    color: base05,
    textDecoration: "underline wavy",
    borderBottom: `1px wavy ${invalid}`
  },
  { tag: [tags.strikethrough], color: invalid, textDecoration: "line-through" },
  // Enhanced syntax highlighting
  { tag: /*@__PURE__*/ tags.constant(tags.name), color: base0A },
  { tag: tags.deleted, color: invalid },
  { tag: tags.squareBracket, color: base05 },
  { tag: tags.angleBracket, color: base05 },
  // Additional specific styles
  { tag: tags.monospace, color: base05 },
  { tag: [tags.contentSeparator], color: base05 },
  { tag: tags.quote, color: base11 }
]);
