import path from "node:path"
import tailwindcss from "@tailwindcss/vite"
import react from "@vitejs/plugin-react"
import { defineConfig } from "vitest/config"

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": path.resolve(import.meta.dirname, "./src"),
    },
  },
  build: {
    rolldownOptions: {
      output: {
        // Stable, readable vendor chunks: React loads with the Add screen; charts only with Month/Charts.
        codeSplitting: {
          groups: [
            { name: "react", test: /node_modules[\\/](react|react-dom|scheduler|react-router|clsx)[\\/]/, priority: 20 },
            { name: "charts", test: /node_modules[\\/](recharts|d3-[^\\/]+|victory-vendor|internmap|decimal\.js-light|@reduxjs|redux|immer|reselect|react-redux)[\\/]/, priority: 10 },
          ],
        },
      },
    },
  },
  server: {
    proxy: {
      "/api": "http://localhost:7447",
      "/healthz": "http://localhost:7447",
    },
  },
  test: {
    globals: true,
    environment: "node",
    include: ["src/**/*.test.ts"],
  },
})
