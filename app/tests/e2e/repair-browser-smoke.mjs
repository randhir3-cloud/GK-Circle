import { chromium } from "@playwright/test";
import fs from "fs";
import path from "path";

const base = process.env.PLAYWRIGHT_BASE_URL || "http://localhost:3000";
const email = process.env.E2E_CREATOR_EMAIL;
const password = process.env.E2E_TEST_PASSWORD;
const outDir = path.join("test-results", "repair-smoke");
fs.mkdirSync(outDir, { recursive: true });

const log = [];
const push = (message) => {
  log.push(message);
  console.log(message);
};

const browser = await chromium.launch({ headless: true });
const page = await browser.newPage();
const consoleErrors = [];
page.on("pageerror", (error) => consoleErrors.push(String(error)));
page.on("console", (msg) => {
  if (msg.type() === "error") consoleErrors.push(msg.text());
});

try {
  if (!email || !password) {
    throw new Error("E2E_CREATOR_EMAIL and E2E_TEST_PASSWORD are required");
  }

  await page.goto(`${base}/`, { waitUntil: "networkidle" });
  push(
    `HOME_OK=${await page.getByText("GK Circle").first().isVisible()}`
  );

  await page.goto(`${base}/account/login`, { waitUntil: "domcontentloaded" });
  await page.locator('input[name="identifier"]').fill(email);
  await page.locator('input[name="password"]').fill(password);
  await Promise.all([
    page.waitForURL((url) => !url.pathname.includes("/account/login"), {
      timeout: 45_000,
    }),
    page.locator('button[type="submit"]').click(),
  ]);
  push("LOGIN_OK=true");

  await page.goto(`${base}/admin/quiz/list-quiz`, {
    waitUntil: "domcontentloaded",
  });
  await page.waitForTimeout(2000);
  push(`QUIZ_LIST_URL=${page.url()}`);
  await page.screenshot({
    path: path.join(outDir, "quiz-list.png"),
    fullPage: true,
  });

  for (const route of [
    "/admin/reports",
    "/admin/played_quiz",
    "/attempt/quizzes/smoke-quiz-id?snapshot_id=smoke-snap",
  ]) {
    await page.goto(`${base}${route}`, { waitUntil: "domcontentloaded" });
    await page.waitForTimeout(1500);
    const bodyLen = (await page.locator("body").innerText()).trim().length;
    push(`ROUTE ${route} url=${page.url()} bodyLen=${bodyLen}`);
    const shot = route.replace(/[/?=&]/g, "_").replace(/^_/, "");
    await page.screenshot({
      path: path.join(outDir, `${shot || "route"}.png`),
      fullPage: true,
    });
  }

  push(`CONSOLE_ERRORS=${consoleErrors.length}`);
  if (consoleErrors.length) {
    push(`CONSOLE_SAMPLE=${consoleErrors.slice(0, 8).join(" | ")}`);
  }
  fs.writeFileSync(path.join(outDir, "smoke.log"), log.join("\n"));
  push("SMOKE_DONE");
} catch (error) {
  push(`SMOKE_FAIL=${error.message}`);
  await page
    .screenshot({ path: path.join(outDir, "smoke-fail.png"), fullPage: true })
    .catch(() => undefined);
  fs.writeFileSync(path.join(outDir, "smoke.log"), log.join("\n"));
  process.exitCode = 1;
} finally {
  await browser.close();
}
