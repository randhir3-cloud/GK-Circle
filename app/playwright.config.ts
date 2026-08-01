import { defineConfig, devices } from "@playwright/test";

const productionAudit = process.env.PRODUCTION_AUDIT === "true";
const headless = process.env.HEADLESS === "true" || Boolean(process.env.CI);

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
    baseURL: process.env.PLAYWRIGHT_BASE_URL || "http://localhost:5000",
    headless,
    viewport: {
      width: 1920,
      height: 1080,
    },
    launchOptions: {
      slowMo: Number(process.env.E2E_SLOW_MO_MS ?? "300"),
    },
    video: productionAudit ? "off" : "on",
    screenshot: productionAudit ? "off" : "only-on-failure",
    trace: productionAudit ? "off" : "retain-on-failure",
  },
  projects: [
    {
      name: "chromium",
      testIgnore: /exam-full-browser-operator-learner-readiness\.spec\.ts$/,
      use: {
        browserName: "chromium",
        headless,
      },
    },
    {
      name: "chromium-observation",
      testMatch: /exam-full-browser-operator-learner-readiness\.spec\.ts$/,
      use: {
        browserName: "chromium",
        headless,
        launchOptions: {
          slowMo: Number(process.env.E2E_SLOW_MO_MS ?? "300"),
        },
      },
    },
  ],
});
