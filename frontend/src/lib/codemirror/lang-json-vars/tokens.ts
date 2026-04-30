/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: AGPL-3.0-only
 */

import { ExternalTokenizer } from "@lezer/lr";
import { VariableRefToken } from "./parser.terms.js";

const OPEN_BRACE = 123;
const CLOSE_BRACE = 125;
const CARRIAGE_RETURN = 13;
const LINE_FEED = 10;

export const variableTokens = new ExternalTokenizer((input) => {
  if (input.next !== OPEN_BRACE || input.peek(1) !== OPEN_BRACE) return;

  input.advance();
  input.advance();

  let hasContent = false;

  for (;;) {
    const next = Number(input.next);

    if (next < 0 || next === CARRIAGE_RETURN || next === LINE_FEED) return;

    if (next === CLOSE_BRACE && input.peek(1) === CLOSE_BRACE) {
      if (!hasContent) return;

      input.advance();
      input.advance();
      input.acceptToken(VariableRefToken);
      return;
    }

    if (next === OPEN_BRACE || next === CLOSE_BRACE) return;

    hasContent = true;
    input.advance();
  }
});
