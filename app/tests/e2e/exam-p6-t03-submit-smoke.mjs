import { chromium } from "@playwright/test";
import fs from "fs";
import path from "path";
import { execSync } from "child_process";

const base = process.env.PLAYWRIGHT_BASE_URL || "http://localhost:3000";
const api = process.env.PLAYWRIGHT_API_BASE_URL || "http://localhost:3010";
const email = process.env.E2E_CREATOR_EMAIL || "course.admin@example.com";
const password = process.env.E2E_TEST_PASSWORD;
if (!password) throw new Error("E2E_TEST_PASSWORD must be supplied explicitly");
const out = path.join("test-results", "exam-p6-t03-submit-smoke");
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

  const title = `Submit Smoke Quiz ${Date.now()}`;
  await page.goto(`${base}/admin/quiz/list-quiz`);
  await page.getByText("Create Quiz", { exact: true }).first().click();
  await page.locator('form input[type="text"][required]').first().fill(title);
  await page
    .locator("form textarea")
    .first()
    .fill("Disposable quiz for submit smoke.");
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
  if (!quizId) throw new Error("quiz id missing from URL after create");

  // Patch quiz to SELF_PACED + PUBLISHED + duration_seconds=600.
  const pgCmd = `docker compose exec -T db psql -U gk_circle -d gk_circle -c "UPDATE quizzes SET assessment_mode='SELF_PACED', status='PUBLISHED', max_attempts=3, duration_seconds=600 WHERE id='${quizId}';"`;
  const pgOut = execSync(pgCmd, { cwd: path.join(process.cwd(), ".."), encoding: "utf8" });
  push(`SET_SELF_PACED=${pgOut.trim()}`);

  // Create question collection
  const collRes = await page.request.post(
    `${api}/api/v1/quizzes/${quizId}/collections`,
    {
      data: {
        title: "Submit Smoke Collection",
        kind: "STATIC",
        position: 0,
      },
    }
  );
  const collJson = await collRes.json();
  const collectionId = collJson?.data?.id;
  push(`COLLECTION_ID=${collectionId}`);

  // Create question
  const qRes = await page.request.post(
    `${api}/api/v1/quizzes/${quizId}/questions`,
    {
      data: {
        question: "What is the capital of Bihar?",
        type: 1,
        options: { 1: "Patna", 2: "Gaya", 3: "Muzaffarpur", 4: "Bhagalpur" },
        answers: [1],
        options_media: "text",
        question_media: "text",
        resource: "",
        points: 1,
        duration_in_seconds: 60,
      },
    }
  );
  const qJson = await qRes.json();
  const questionId = qJson?.data;
  push(`QUESTION_ID=${questionId}`);

  // Add question to collection
  await page.request.put(
    `${api}/api/v1/quizzes/${quizId}/collections/${collectionId}/members`,
    { data: { question_ids: [questionId] } }
  );

  // Create snapshot
  const snapRes = await page.request.post(
    `${api}/api/v1/quizzes/${quizId}/test-snapshots`,
    { data: { collection_ids: [collectionId] } }
  );
  const snapJson = await snapRes.json();
  const snapshotId = snapJson?.data?.id;
  push(`SNAPSHOT_ID=${snapshotId}`);

  // Go to attempt instructions
  await page.goto(`${base}/attempt/quizzes/${quizId}?snapshot_id=${snapshotId}`);
  await page.waitForTimeout(2000);
  const startBtn = page.getByRole("button", { name: /start|resume/i }).first();
  if (await startBtn.count()) {
    await startBtn.click();
    await page.waitForTimeout(2000);
  }
  push(`PLAYER_URL=${page.url()}`);

  // Verify timer component is rendered
  const timer = page.locator(".attempt-timer");
  push(`TIMER_VISIBLE=${(await timer.count()) > 0}`);

  // Answer question
  const option = page.locator('input[type="radio"]').first();
  if (await option.count()) {
    await option.check({ force: true });
    await page.waitForTimeout(1500);
    push("SELECTED_OPTION=true");
  }

  // Take screenshot of player with timer
  await page.screenshot({
    path: path.join(out, "01-player-with-timer.png"),
    fullPage: true,
  });

  // Click Submit Attempt button
  const submitBtn = page.locator(".attempt-player__nav-button--submit");
  push(`SUBMIT_BUTTON_VISIBLE=${(await submitBtn.count()) > 0}`);
  await submitBtn.click();
  await page.waitForTimeout(1000);

  // Take screenshot of Submit Dialog
  await page.screenshot({
    path: path.join(out, "02-submit-dialog.png"),
    fullPage: true,
  });

  // Click Confirm Submission inside modal
  const confirmBtn = page.locator(".submit-dialog__button--confirm");
  await confirmBtn.click();
  await page.waitForTimeout(3000);

  // Take screenshot of Submitted Screen
  await page.screenshot({
    path: path.join(out, "03-submitted-screen.png"),
    fullPage: true,
  });

  const submittedScreen = page.locator(".submitted-screen");
  push(`SUBMITTED_SCREEN_VISIBLE=${(await submittedScreen.count()) > 0}`);
  const submittedText = await submittedScreen.innerText();
  push(`SUBMITTED_TEXT=${submittedText.replace(/\n+/g, " | ").slice(0, 300)}`);

  // Reload page and verify it remains on the submitted terminal screen
  await page.reload({ waitUntil: "domcontentloaded" });
  await page.waitForTimeout(2000);
  push(`AFTER_RELOAD_SUBMITTED=${(await page.locator(".submitted-screen").count()) > 0}`);

  push(`PAGE_ERRORS=${pageErrors.length}`);
  fs.writeFileSync(path.join(out, "smoke.log"), log.join("\n"));
  push("EXAM_P6_T03_SUBMIT_SMOKE_DONE");
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
