import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const packageJson = JSON.parse(readFileSync(path.resolve(__dirname, "package.json"), "utf-8"));

// 版本号优先级: VITE_APP_VERSION 环境变量 > package.json.version > "dev"
const appVersion = (process.env.VITE_APP_VERSION && process.env.VITE_APP_VERSION.trim()) ||
  packageJson.version ||
  "dev";

console.log(`[vite] App version: ${appVersion}`);

export default defineConfig({
  plugins: [vue()],
  define: {
    __APP_VERSION__: JSON.stringify(appVersion),
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "src"),
    },
  },
  server: {
    port: 5173,
    proxy: {
      "/api": {
        target: "http://localhost:8888",
        changeOrigin: true,
      },
      "/uploads": {
        target: "http://localhost:8888",
        changeOrigin: true,
      },
    },
  },
});
