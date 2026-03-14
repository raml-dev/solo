import { defineConfig, normalizePath } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import { resolve } from "node:path";

const srcDir = `${normalizePath(resolve(__dirname, "src"))}/`;

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    svelte({
      dynamicCompileOptions: ({ filename, compileOptions }) => {
        const normalized = normalizePath(filename);
        if (normalized.startsWith(srcDir) && compileOptions.runes !== true) {
          return { runes: true };
        }
      }
    })
  ],
  resolve: {
    alias: {
      $wails: resolve(__dirname, "wailsjs"),
      $src: resolve(__dirname, "src")
    }
  }
});
