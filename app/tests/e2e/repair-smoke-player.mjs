import { chromium } from "@playwright/test";
import fs from "fs";
import path from "path";

const base = process.env.PLAYWRIGHT_BASE_URL || "http://localhost:3000";
const api = process.env.PLAYWRIGHT_API_BASE_URL || "http://localhost:3010";
const email = process.env.E2E_CREATOR_EMAIL;
const password = process.env.E2E_TEST_PASSWORD;
const out = path.join("test-results", "repair-smoke-player");
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

  const title = `Repair Smoke Quiz ${Date.now()}`;
  await page.goto(`${base}/admin/quiz/list-quiz`);
  await page.getByText("Create Quiz", { exact: true }).first().click();
  await page.locator('form input[type="text"][required]').first().fill(title);
  await page
    .locator("form textarea")
    .first()
    .fill("Disposable quiz for repair smoke.");
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
  const created = await createRes.json();
  push(`CREATE_BODY=${JSON.stringify(created).slice(0, 300)}`);
  await page.waitForURL(/\/admin\/quiz\/list-quiz\/[^/]+/, { timeout: 30000 });
  const quizId = page.url().split("/admin/quiz/list-quiz/")[1]?.split(/[?#]/)[0];
  push(`QUIZ_ID=${quizId}`);
  if (!quizId) throw new Error("quiz id missing from URL after create");

  // Patch quiz to SELF_PACED + PUBLISHED via direct DB update.
  // The UI creates quizzes as LIVE/DRAFT; there is no self-paced toggle yet.
  const { execSync } = await import("child_process");
  const pgCmd = `docker compose exec -T db psql -U gk_circle -d gk_circle -c "UPDATE quizzes SET assessment_mode='SELF_PACED', status='PUBLISHED', max_attempts=3 WHERE id='${quizId}';"`;
  try {
    const pgOut = execSync(pgCmd, { cwd: process.cwd() + "/../", encoding: "utf8" });
    push(`SET_SELF_PACED=${pgOut.trim()}`);
  } catch (pgErr) {
    push(`SET_SELF_PACED_ERR=${pgErr.message.slice(0, 200)}`);
    throw new Error("Failed to set quiz self-paced mode");
  }

  // Prefer API question create to avoid brittle editor chrome selectors.
  const questionRes = await page.request.post(
    `${api}/api/v1/quizzes/${quizId}/questions`,
    {
      data: {
        question: "What is 2 + 2?",
        type: 1,
        options: { 1: "3", 2: "4", 3: "5", 4: "6" },
        answers: [2],
        options_media: "text",
        question_media: "text",
        resource: "",
        points: 1,
        duration_in_seconds: 30,
      },
    }
  );
  push(
    `ADD_QUESTION=${questionRes.status()} ${(await questionRes.text()).slice(0, 200)}`
  );
  if (questionRes.status() >= 400) {
    throw new Error("question create failed");
  }

  // Snapshot + attempt via API using authenticated browser context.
  // Step 1: Create a STATIC question collection for this quiz.
  const collRes = await page.request.post(
    `${api}/api/v1/quizzes/${quizId}/collections`,
    {
      data: {
        title: "Repair Smoke Collection",
        kind: "STATIC",
        position: 0,
      },
    }
  );
  const collText = await collRes.text();
  push(`CREATE_COLLECTION=${collRes.status()} ${collText.slice(0, 240)}`);
  let collectionId = null;
  try {
    const collJson = JSON.parse(collText);
    collectionId = collJson?.data?.id || null;
  } catch {
    collectionId = null;
  }
  push(`COLLECTION_ID=${collectionId}`);
  if (!collectionId) throw new Error("Unable to create or parse collection id");

  // Step 2: Extract question ID from add-question response and add it to the collection.
  let questionId = null;
  try {
    const qRes = await page.request.post(
      `${api}/api/v1/quizzes/${quizId}/questions`,
      {
        data: {
          question: "What is 3 + 3?",
          type: 1,
          options: { 1: "5", 2: "6", 3: "7", 4: "8" },
          answers: [2],
          options_media: "text",
          question_media: "text",
          resource: "",
          points: 1,
          duration_in_seconds: 30,
        },
      }
    );
    const qText = await qRes.text();
    push(`ADD_QUESTION2=${qRes.status()} ${qText.slice(0, 200)}`);
    const qJson = JSON.parse(qText);
    questionId = qJson?.data || null;
  } catch {}
  push(`QUESTION_ID=${questionId}`);
  if (!questionId) throw new Error("Unable to create second question or parse ID");

  // Step 3: Add question to collection.
  const memberRes = await page.request.put(
    `${api}/api/v1/quizzes/${quizId}/collections/${collectionId}/members`,
    { data: { question_ids: [questionId] } }
  );
  push(`ADD_MEMBER=${memberRes.status()} ${(await memberRes.text()).slice(0, 200)}`);
  if (memberRes.status() >= 400) throw new Error("member add failed");

  // Step 4: Create snapshot with the collection.
  const snapRes = await page.request.post(
    `${api}/api/v1/quizzes/${quizId}/test-snapshots`,
    { data: { collection_ids: [collectionId] } }
  );
  const snapText = await snapRes.text();
  push(`CREATE_SNAPSHOT=${snapRes.status()} ${snapText.slice(0, 240)}`);
  let snapshotId = null;
  try {
    const snapJson = JSON.parse(snapText);
    snapshotId =
      snapJson?.data?.id ||
      snapJson?.data?.snapshot_id ||
      snapJson?.id ||
      null;
  } catch {
    snapshotId = null;
  }
  push(`SNAPSHOT_ID=${snapshotId}`);

  if (!snapshotId) {
    throw new Error("Unable to create or parse snapshot id");
  }

  await page.goto(
    `${base}/attempt/quizzes/${quizId}?snapshot_id=${snapshotId}`
  );
  await page.waitForTimeout(2000);
  await page.screenshot({
    path: path.join(out, "instructions.png"),
    fullPage: true,
  });
  const start = page.getByRole("button", { name: /start|resume/i }).first();
  if (await start.count()) {
    await start.click();
    await page.waitForTimeout(2500);
  }
  push(`PLAYER_URL=${page.url()}`);
  await page.screenshot({
    path: path.join(out, "player.png"),
    fullPage: true,
  });

  const option = page.locator('input[type="radio"]').first();
  if (await option.count()) {
    await option.check({ force: true });
    await page.waitForTimeout(1500);
    push("SELECTED_OPTION=true");
    const answered = page.getByText(/answered|saved|saving/i);
    push(`SAVE_UI=${(await answered.count()) > 0}`);
    await page.reload({ waitUntil: "domcontentloaded" });
    await page.waitForTimeout(2000);
    const restored = page.locator('input[type="radio"]:checked');
    push(`RESTORED=${(await restored.count()) > 0}`);
  } else {
    push("SELECTED_OPTION=false");
    push(`PLAYER_BODY=${(await page.locator("body").innerText()).slice(0, 200)}`);
  }

  push(`PAGE_ERRORS=${pageErrors.length}`);
  fs.writeFileSync(path.join(out, "smoke.log"), log.join("\n"));
  push("SMOKE_PLAYER_DONE");
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
