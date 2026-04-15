/**
 * Copyright 2026-present raml-dev
 * SPDX-License-Identifier: AGPL-3.0-only
 */

import { svelte } from "@sveltejs/vite-plugin-svelte";
import tailwindcss from "@tailwindcss/vite";
import { resolve } from "node:path";
import { defineConfig, normalizePath } from "vite";

const srcDir = `${normalizePath(resolve(__dirname, "src"))}/`;

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    tailwindcss(),
    svelte({
      dynamicCompileOptions: ({ filename, compileOptions }) => {
        const normalized = normalizePath(filename);

        if (normalized.startsWith(srcDir) && compileOptions.runes !== true) {
          return { runes: true };
        }
      }
    })
  ],
  server: {
    hmr: {
      host: "localhost",
      protocol: "ws",
      port: 5173
    }
  },
  resolve: {
    alias: {
      $wails: resolve(__dirname, "wailsjs"),
      $src: resolve(__dirname, "src")
    }
  }
});
