import { defineConfig, devices } from "@playwright/test"

// Runs against an already running app (make dev-api / docker). Override with BASE_URL.
export default defineConfig({
  testDir: ".",
  testMatch: "*.spec.ts",
  timeout: 30_000,
  expect: { timeout: 5_000 },
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  // Tests share one live database and assert on month totals: never run them concurrently.
  workers: 1,
  reporter: process.env.CI ? "line" : "list",
  use: {
    baseURL: process.env.BASE_URL ?? "http://localhost:7447",
    trace: "retain-on-failure",
  },
  projects: [{ name: "phone", use: { ...devices["Pixel 7"] } }],
})
