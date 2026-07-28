import { chromium } from "@playwright/test";
import fs from "fs";
import path from "path";
import { execSync } from "child_process";

const base = process.env.PLAYWRIGHT_BASE_URL || "http://localhost:3000";
const api = process.env.PLAYWRIGHT_API_BASE_URL || "http://localhost:3010";
const email = process.env.E2E_CREATOR_EMAIL || "course.admin@example.com";
const password = process.env.E2E_TEST_PASSWORD || "Password123!";
const out = path.join("test-results", "exam-p7-t01-result-smoke");
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
  // Step 1: Login
  await page.goto(`${base}/account/login`, { waitUntil: "networkidle" });
  await page.waitForSelector('input[name="identifier"]');
  await page.locator('input[name="identifier"]').fill(email);
  await page.locator('input[name="password"]').fill(password);
  await Promise.all([
    page.waitForURL((u) => !u.pathname.includes("/account/login"), {
      timeout: 45000,
    }),
    page.locator('button[type="submit"]').first().click(),
  ]);
  push(`LOGIN_OK url=${page.url()}`);

  // Step 2: Create Quiz with allow_answer_review = true
  const title = `Result Smoke Quiz ${Date.now()}`;
  await page.goto(`${base}/admin/quiz/list-quiz`);
  await page.getByText("Create Quiz", { exact: true }).first().click();
  await page.locator('form input[type="text"][required]').first().fill(title);
  await page
    .locator("form textarea")
    .first()
    .fill("Disposable quiz for EXAM-P7-T01 result smoke.");
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

  // Set SELF_PACED + PUBLISHED + allow_answer_review = true
  const pgCmd = `docker compose exec -T db psql -U gk_circle -d gk_circle -c "UPDATE quizzes SET assessment_mode='SELF_PACED', status='PUBLISHED', max_attempts=5, duration_seconds=1800, allow_answer_review=true WHERE id='${quizId}';"`;
  const pgOut = execSync(pgCmd, { cwd: path.join(process.cwd(), ".."), encoding: "utf8" });
  push(`SET_ALLOW_REVIEW_TRUE=${pgOut.trim()}`);

  // Create Collection
  const collRes = await page.request.post(
    `${api}/api/v1/quizzes/${quizId}/collections`,
    {
      data: {
        title: "Result Smoke Collection",
        kind: "STATIC",
        position: 0,
      },
    }
  );
  const collJson = await collRes.json();
  const collectionId = collJson?.data?.id;

  // Create Question with Explanation (resource)
  const qRes = await page.request.post(
    `${api}/api/v1/quizzes/${quizId}/questions`,
    {
      data: {
        question: "Which city is known as the Silicon Valley of India?",
        type: 1,
        options: { 1: "Bengaluru", 2: "Hyderabad", 3: "Pune", 4: "Chennai" },
        answers: [1],
        options_media: "text",
        question_media: "text",
        resource: "Bengaluru is the IT hub of India.",
        points: 5,
        duration_in_seconds: 60,
      },
    }
  );
  const qJson = await qRes.json();
  const questionId = qJson?.data;

  // Add question to collection & build snapshot
  await page.request.put(
    `${api}/api/v1/quizzes/${quizId}/collections/${collectionId}/members`,
    { data: { question_ids: [questionId] } }
  );

  const snapRes = await page.request.post(
    `${api}/api/v1/quizzes/${quizId}/test-snapshots`,
    { data: { collection_ids: [collectionId] } }
  );
  const snapJson = await snapRes.json();
  const snapshotId = snapJson?.data?.id;
  push(`SNAPSHOT_ID=${snapshotId}`);

  // Case 1: Review Enabled Attempt
  await page.goto(`${base}/attempt/quizzes/${quizId}?snapshot_id=${snapshotId}`);
  await page.waitForTimeout(2000);
  const startBtn = page.getByRole("button", { name: /start|resume/i }).first();
  if (await startBtn.count()) {
    await startBtn.click();
    await page.waitForTimeout(2000);
  }

  // Answer question (select option 1: Bengaluru)
  const option = page.locator('input[type="radio"]').first();
  if (await option.count()) {
    await option.check({ force: true });
    await page.waitForTimeout(1000);
  }

  // Submit attempt
  await page.locator(".attempt-player__nav-button--submit").click();
  await page.waitForTimeout(1000);
  await page.locator(".submit-dialog__button--confirm").click();
  await page.waitForTimeout(3000);

  // Click "View Assessment Results" button on Submitted Screen
  const viewResultsBtn = page.locator(".submitted-screen__button--primary");
  push(`VIEW_RESULTS_BTN_VISIBLE=${(await viewResultsBtn.count()) > 0}`);
  await viewResultsBtn.click();
  await page.waitForTimeout(3000);

  push(`RESULT_PAGE_URL=${page.url()}`);
  await page.screenshot({
    path: path.join(out, "01-result-page-review-enabled.png"),
    fullPage: true,
  });

  // Verify Result Page Components for Review Enabled
  const scoreCard = page.locator(".score-card");
  const summaryGrid = page.locator(".attempt-summary");
  const questionReview = page.locator(".question-review");

  push(`SCORECARD_VISIBLE=${(await scoreCard.count()) > 0}`);
  push(`SUMMARY_VISIBLE=${(await summaryGrid.count()) > 0}`);
  push(`QUESTION_REVIEW_VISIBLE=${(await questionReview.count()) > 0}`);

  const reviewText = await questionReview.innerText();
  push(`QUESTION_REVIEW_HAS_STEM=${reviewText.includes("Silicon Valley")}`);
  push(`QUESTION_REVIEW_HAS_EXPLANATION=${reviewText.includes("IT hub of India")}`);

  // Test Refresh Safety on Result Page
  await page.reload({ waitUntil: "domcontentloaded" });
  await page.waitForTimeout(2000);
  push(`RELOAD_SCORECARD_VISIBLE=${(await page.locator(".score-card").count()) > 0}`);

  // Case 2: Review Disabled Test
  // Update quiz to allow_answer_review = false
  const pgCmdFalse = `docker compose exec -T db psql -U gk_circle -d gk_circle -c "UPDATE quizzes SET allow_answer_review=false WHERE id='${quizId}';"`;
  execSync(pgCmdFalse, { cwd: path.join(process.cwd(), ".."), encoding: "utf8" });

  await page.reload({ waitUntil: "domcontentloaded" });
  await page.waitForTimeout(2000);

  await page.screenshot({
    path: path.join(out, "02-result-page-review-disabled.png"),
    fullPage: true,
  });

  const reviewDisabledNotice = page.locator(".result-page__review-disabled");
  push(`REVIEW_DISABLED_NOTICE_VISIBLE=${(await reviewDisabledNotice.count()) > 0}`);
  push(`QUESTION_REVIEW_ABSENT=${(await page.locator(".question-review").count()) === 0}`);

  push(`PAGE_ERRORS=${pageErrors.length}`);
  fs.writeFileSync(path.join(out, "smoke.log"), log.join("\n"));
  push("EXAM_P7_T01_RESULT_SMOKE_DONE");
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
