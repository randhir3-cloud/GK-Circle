import { chromium } from "@playwright/test";
import fs from "fs";
import path from "path";

const base = process.env.PLAYWRIGHT_BASE_URL;
const api = process.env.PLAYWRIGHT_API_BASE_URL;
const email = process.env.E2E_CREATOR_EMAIL;
const password = process.env.E2E_TEST_PASSWORD;
const out = path.join("test-results", "repair-smoke-attempt");
fs.mkdirSync(out, { recursive: true });
const log = [];
const push = (m) => {
  log.push(m);
  console.log(m);
};

const browser = await chromium.launch({ headless: true });
const page = await browser.newPage();
const errs = [];
page.on("pageerror", (e) => errs.push(String(e)));

try {
  await page.goto(`${base}/account/login`);
  await page.locator('input[name="identifier"]').fill(email);
  await page.locator('input[name="password"]').fill(password);
  await Promise.all([
    page.waitForURL((u) => !u.pathname.includes("/account/login"), {
      timeout: 45000,
    }),
    page.locator('button[type="submit"]').click(),
  ]);
  push("LOGIN_OK");

  const quizzes = await page.request.get(`${api}/api/v1/quizzes`);
  push(`GET /api/v1/quizzes => ${quizzes.status()}`);

  await page.goto(`${base}/admin/quiz/list-quiz`);
  await page.waitForTimeout(2000);
  const links = page.locator('a[href*="/admin/quiz/list-quiz/"]');
  const count = await links.count();
  push(`QUIZ_LINKS=${count}`);

  await page.goto(
    `${base}/attempt/quizzes/00000000-0000-0000-0000-000000000001?snapshot_id=00000000-0000-0000-0000-000000000002`
  );
  await page.waitForTimeout(2000);
  const body = (await page.locator("body").innerText())
    .trim()
    .replace(/\s+/g, " ");
  push(`ATTEMPT_BODY=${body.slice(0, 240)}`);
  await page.screenshot({
    path: path.join(out, "attempt.png"),
    fullPage: true,
  });

  await page.goto(`${base}/admin/reports`);
  await page.waitForTimeout(1000);
  push(
    `REPORTS_OK=${await page.getByText("Reports").first().isVisible()}`
  );

  push(`PAGE_ERRORS=${errs.length}`);
  fs.writeFileSync(path.join(out, "smoke.log"), log.join("\n"));
  push("DONE");
} catch (e) {
  push(`FAIL=${e.message}`);
  fs.writeFileSync(path.join(out, "smoke.log"), log.join("\n"));
  process.exitCode = 1;
} finally {
  await browser.close();
}
