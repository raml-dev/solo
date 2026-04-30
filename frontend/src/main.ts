/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: AGPL-3.0-only
 */

import "$src/app.css";
import App from "$src/App.svelte";
import "$src/assets/styles/codemirror-theme.css";
import initLogger from "$src/logger";
import { mount } from "svelte";

initLogger();

let target = document.getElementById("app");

if (!target) {
  target = document.createElement("div");
  target.id = "app";
  document.body.appendChild(target);
}
const app = mount(App, { target });

export default app;
