// @ts-check
import react from "@astrojs/react";
import cloudflare from "@astrojs/cloudflare";
import { defineConfig } from "astro/config";
import tailwindcss from "@tailwindcss/vite";

import { execFile } from "node:child_process";
import { promisify } from "node:util";
import path from "node:path";

const execFileAsync = promisify(execFile);

/** @returns {import("vite").Plugin} */
const generateDesignTokensPlugin = () => ({
  name: "generate-design-tokens",

  configureServer(server) {
    const sourcePath = path.resolve("src/types/store.ts");
    const scriptPath = path.resolve("scripts/generate-design-tokens.mjs");
    let generating = false;

    server.watcher.add(sourcePath);

    server.watcher.on("change", async (file) => {
      if (path.resolve(file) !== sourcePath || generating) return;

      generating = true;

      try {
        await execFileAsync(process.execPath, [scriptPath]);
        server.config.logger.info(
          "[generate-design-tokens] Tokens actualizados.",
        );
      } catch (error) {
        server.config.logger.error(
          `[generate-design-tokens] Error: ${error instanceof Error ? error.message : String(error)}`,
        );
      } finally {
        generating = false;
      }
    });
  },
});

// https://astro.build/config
export default defineConfig({
  output: "server",
  integrations: [react()],
  adapter: cloudflare({
    configPath: "./wrangler.jsonc",
  }),
  vite: {
    plugins: [generateDesignTokensPlugin(), tailwindcss()],
    server: {
      proxy: {
        "/api": {
          target: "http://localhost:8787",
          changeOrigin: true,
        },
      },
    },
  },
});
