import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./tests/e2e",
  timeout: 8 * 60 * 60 * 1000, // 8 hours for human observation & inspection pauses
  expect: {
    timeout: 15_000,
  },
  fullyParallel: false,
  workers: 1,
  retries: 0,
  reporter: [["line"]],
  use: {
    ...devices["Desktop Chrome"],
    baseURL: process.env.PLAYWRIGHT_BASE_URL || "http://localhost:3000",
    headless: process.env.HEADLESS === "true",
    viewport: {
      width: 1920,
      height: 1080,
    },
    launchOptions: {
      slowMo: Number(process.env.E2E_SLOW_MO_MS ?? "300"),
    },
    video: "on",
    screenshot: "only-on-failure",
    trace: "retain-on-failure",
  },
  projects: [
    {
      name: "chromium-observation",
      use: {
        browserName: "chromium",
        headless: process.env.HEADLESS === "true",
        launchOptions: {
          slowMo: Number(process.env.E2E_SLOW_MO_MS ?? "300"),
        },
      },
    },
  ],
});
