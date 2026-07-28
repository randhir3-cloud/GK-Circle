import { chromium } from "@playwright/test";
import fs from "fs";
import path from "path";
import { execSync } from "child_process";

const base = process.env.PLAYWRIGHT_BASE_URL || "http://localhost:3000";
const api = process.env.PLAYWRIGHT_API_BASE_URL || "http://localhost:3010";
const email = process.env.E2E_CREATOR_EMAIL || "course.admin@example.com";
const password = process.env.E2E_TEST_PASSWORD || "Password123!";
const out = path.join("test-results", "exam-p6-t03-expiry-smoke");
fs.mkdirSync(out, { recursive: true });
const log = [];
const push = (m) => {
  log.push(m);
  console.log(m);
};

const browser = await chromium.launch({ headless: true });
const page = await browser.newPage();
const pageErrors = [];
page.on("pageerror", (e) => pageErrors.push(String(e)));

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

  const title = `Expiry Smoke Quiz ${Date.now()}`;
  await page.goto(`${base}/admin/quiz/list-quiz`);
  await page.getByText("Create Quiz", { exact: true }).first().click();
  await page.locator('form input[type="text"][required]').first().fill(title);
  await page
    .locator("form textarea")
    .first()
    .fill("Disposable quiz for expiry smoke.");
  const createPromise = page.waitForResponse(
    (res) =>
      res.url().includes("/api/v1/quizzes") &&
      res.request().method() === "POST",
    { timeout: 30000 }
  );
  await page
    .locator("form")
    .filter({ hasText: "Create New Quiz" })
    .locator('button:has-text("Create Quiz")')
    .click();
  const createRes = await createPromise;
  push(`CREATE_QUIZ=${createRes.status()}`);
  await page.waitForURL(/\/admin\/quiz\/list-quiz\/[^/]+/, { timeout: 30000 });
  const quizId = page.url().split("/admin/quiz/list-quiz/")[1]?.split(/[?#]/)[0];
  push(`QUIZ_ID=${quizId}`);

  // Create question collection
  const collRes = await page.request.post(
    `${api}/api/v1/quizzes/${quizId}/collections`,
    {
      data: { title: "Expiry Smoke Collection", kind: "STATIC", position: 0 },
    }
  );
  const collJson = await collRes.json();
  const collectionId = collJson?.data?.id;

  // Create question
  const qRes = await page.request.post(
    `${api}/api/v1/quizzes/${quizId}/questions`,
    {
      data: {
        question: "What is 10 + 10?",
        type: 1,
        options: { 1: "15", 2: "20", 3: "25", 4: "30" },
        answers: [2],
        options_media: "text",
        question_media: "text",
        resource: "",
        points: 1,
        duration_in_seconds: 10,
      },
    }
  );
  const qJson = await qRes.json();
  const questionId = qJson?.data;

  // Member add
  await page.request.put(
    `${api}/api/v1/quizzes/${quizId}/collections/${collectionId}/members`,
    { data: { question_ids: [questionId] } }
  );

  // Snapshot
  const snapRes = await page.request.post(
    `${api}/api/v1/quizzes/${quizId}/test-snapshots`,
    { data: { collection_ids: [collectionId] } }
  );
  const snapJson = await snapRes.json();
  const snapshotId = snapJson?.data?.id;

  // Set quiz duration_seconds=6 to expire very quickly after start
  const pgCmd = `docker compose exec -T db psql -U gk_circle -d gk_circle -c "UPDATE quizzes SET assessment_mode='SELF_PACED', status='PUBLISHED', max_attempts=3, duration_seconds=6 WHERE id='${quizId}';"`;
  execSync(pgCmd, { cwd: path.join(process.cwd(), ".."), encoding: "utf8" });

  // Start attempt
  await page.goto(`${base}/attempt/quizzes/${quizId}?snapshot_id=${snapshotId}`);
  await page.waitForTimeout(2000);
  const startBtn = page.getByRole("button", { name: /start|resume/i }).first();
  if (await startBtn.count()) {
    await startBtn.click();
    await page.waitForTimeout(1000);
  }
  push(`PLAYER_URL=${page.url()}`);

  // Take screenshot during initial countdown
  await page.screenshot({
    path: path.join(out, "01-player-countdown.png"),
    fullPage: true,
  });

  // Wait 7 seconds for timer to reach 0 and trigger auto-submit
  push("WAITING_FOR_EXPIRY...");
  await page.waitForTimeout(7000);

  // Take screenshot of Auto-Submitted screen
  await page.screenshot({
    path: path.join(out, "02-auto-submitted-screen.png"),
    fullPage: true,
  });

  const submittedScreen = page.locator(".submitted-screen");
  push(`AUTO_SUBMITTED_SCREEN_VISIBLE=${(await submittedScreen.count()) > 0}`);
  const submittedText = await submittedScreen.innerText();
  push(`AUTO_SUBMITTED_TEXT=${submittedText.replace(/\n+/g, " | ").slice(0, 300)}`);

  // Query database to verify status is AUTO_SUBMITTED
  const dbCheckCmd = `docker compose exec -T db psql -U gk_circle -d gk_circle -c "SELECT status FROM assessment_attempts WHERE quiz_id='${quizId}';"`;
  const dbOut = execSync(dbCheckCmd, { cwd: path.join(process.cwd(), ".."), encoding: "utf8" });
  push(`DB_STATUS_CHECK=${dbOut.trim().replace(/\n+/g, " ")}`);

  push(`PAGE_ERRORS=${pageErrors.length}`);
  fs.writeFileSync(path.join(out, "smoke.log"), log.join("\n"));
  push("EXAM_P6_T03_EXPIRY_SMOKE_DONE");
} catch (e) {
  push(`FAIL=${e.message}`);
  await page
    .screenshot({ path: path.join(out, "fail.png"), fullPage: true })
    .catch(() => undefined);
  fs.writeFileSync(path.join(out, "smoke.log"), log.join("\n"));
  process.exitCode = 1;
} finally {
  await browser.close();
}
