import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import { resolve } from "node:path";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [svelte()],
  resolve: {
    alias: {
      $wails: resolve(__dirname, "wailsjs"),
      $src: resolve(__dirname, "src")
    }
  }
});
