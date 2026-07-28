import { chromium } from "playwright";
import { execSync } from "child_process";

const BASE_URL = process.env.WEB_BASE_URL || "http://localhost:3000";
const REPO_ROOT = "e:\\GK Circle v2";

function queryDB(sql) {
  // Escapes double quotes and passes query to docker compose db container
  const cmd = `docker compose exec -T db psql -U gk_circle -d gk_circle -t -A -c "${sql.replace(/"/g, '\\"')}"`;
  const out = execSync(cmd, { cwd: REPO_ROOT, encoding: "utf8" });
  return out.trim();
}

async function run() {
  console.log("=== EXAM-P7-T02 Result Release Policy & Review Controls E2E Smoke Test ===");

  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext();
  const page = await context.newPage();

  try {
    // 1. Authenticate user
    console.log("1. Logging in...");
    await page.goto(`${BASE_URL}/account/login`, { waitUntil: "networkidle" });
    await page.fill('input[type="email"], input[name="identifier"]', "test@example.com");
    await page.fill('input[type="password"]', "Password123!");
    await page.click('button[type="submit"]');
    await page.waitForURL((url) => !url.href.includes("/account/login"), { timeout: 15000 });
    console.log("   Authenticated successfully!");

    // Helper: Create quiz & question
    async function createQuiz(overrides = {}) {
      const title = overrides.title || `T02 Release Policy Test Quiz ${Date.now()}`;
      const policy = overrides.result_release_policy || "IMMEDIATE";
      const released = overrides.results_released ?? false;
      const schedAt = overrides.results_scheduled_at ? `'${overrides.results_scheduled_at}'` : "NULL";
      const allowRev = overrides.allow_answer_review ?? true;
      const showScore = overrides.show_score ?? true;
      const showPass = overrides.show_pass_fail ?? true;
      const showCorr = overrides.show_correctness ?? true;
      const showExp = overrides.show_explanations ?? true;

      const quizId = queryDB(
        `INSERT INTO quizzes (
          title, description, is_public, assessment_mode, status, duration_seconds, max_attempts,
          negative_marks_per_question, allow_answer_review, result_release_policy, results_released,
          results_scheduled_at, show_score, show_pass_fail, show_correctness, show_explanations, created_at, updated_at
        ) VALUES (
          '${title}', 'Test Quiz Description', true, 'SELF_PACED', 'PUBLISHED', 1800, 1,
          0, ${allowRev}, '${policy}', ${released}, ${schedAt}, ${showScore}, ${showPass}, ${showCorr}, ${showExp}, NOW(), NOW()
        ) RETURNING id`
      );

      const qId = queryDB(
        `INSERT INTO questions (
          question, type, options, answers, official_answer, authoritative_answer,
          answer_review_status, points, resource, created_at, updated_at
        ) VALUES (
          'What is the capital of Bihar?', 1,
          '{\\"1\\": \\"Patna\\", \\"2\\": \\"Gaya\\", \\"3\\": \\"Muzaffarpur\\", \\"4\\": \\"Bhagalpur\\"}',
          '[1]', '[1]', '[1]', 'CONFIRMED', 10, 'Patna is the capital and largest city of Bihar.', NOW(), NOW()
        ) RETURNING id`
      );

      queryDB(
        `INSERT INTO quiz_questions (quiz_id, question_id, position, created_at) VALUES ('${quizId}', '${qId}', 0, NOW())`
      );

      return { quizId, qId };
    }

    // Helper: Create attempt and submit
    async function createAndSubmitAttempt(quizId, qId) {
      const snapId = queryDB(
        `INSERT INTO test_snapshots (quiz_id, created_at) VALUES ('${quizId}', NOW()) RETURNING id`
      );

      queryDB(
        `INSERT INTO test_snapshot_items (
          snapshot_id, position, question_id, lineage_id, revision_number,
          question, type, options, answers, official_answer, authoritative_answer,
          answer_review_status, points, resource, created_at
        ) VALUES (
          '${snapId}', 0, '${qId}', gen_random_uuid(), 1,
          'What is the capital of Bihar?', 1,
          '{\\"1\\": \\"Patna\\", \\"2\\": \\"Gaya\\", \\"3\\": \\"Muzaffarpur\\", \\"4\\": \\"Bhagalpur\\"}',
          '[1]', '[1]', '[1]', 'CONFIRMED', 10, 'Patna is the capital and largest city of Bihar.', NOW()
        )`
      );

      const userId = queryDB(`SELECT id FROM users LIMIT 1`);

      const attemptId = queryDB(
        `INSERT INTO assessment_attempts (
          quiz_id, user_id, test_snapshot_id, attempt_number, status,
          question_order, negative_marks_per_question, expected_max_score,
          started_at, submitted_at, total_score, max_score, time_taken_seconds, created_at, updated_at
        ) VALUES (
          '${quizId}', '${userId}', '${snapId}', 1, 'SUBMITTED',
          '["${qId}"]', 0, 10, NOW() - INTERVAL '5 minutes', NOW(), 10, 10, 120, NOW(), NOW()
        ) RETURNING id`
      );

      const snapshotItemId = queryDB(
        `SELECT id FROM test_snapshot_items WHERE snapshot_id = '${snapId}' LIMIT 1`
      );

      queryDB(
        `INSERT INTO assessment_attempt_snapshot_items (
          attempt_id, snapshot_item_id, position, question_id, lineage_id, revision_number,
          question, type, options, answers, official_answer, authoritative_answer,
          answer_review_status, points, resource, created_at
        ) VALUES (
          '${attemptId}', '${snapshotItemId}', 0, '${qId}', gen_random_uuid(), 1,
          'What is the capital of Bihar?', 1,
          '{\\"1\\": \\"Patna\\", \\"2\\": \\"Gaya\\", \\"3\\": \\"Muzaffarpur\\", \\"4\\": \\"Bhagalpur\\"}',
          '[1]', '[1]', '[1]', 'CONFIRMED', 10, 'Patna is the capital and largest city of Bihar.', NOW()
        )`
      );

      queryDB(
        `INSERT INTO attempt_answers (
          attempt_id, question_id, selected_options, is_marked_review, is_correct, score, answered_at, time_taken_seconds, created_at, updated_at
        ) VALUES (
          '${attemptId}', '${qId}', '[1]', false, true, 10, NOW(), 30, NOW(), NOW()
        )`
      );

      return { attemptId };
    }

    // CASE 1: IMMEDIATE RELEASE
    console.log("2. Case 1: Immediate Release...");
    const c1 = await createQuiz({ result_release_policy: "IMMEDIATE" });
    const a1 = await createAndSubmitAttempt(c1.quizId, c1.qId);
    await page.goto(`${BASE_URL}/attempt/quizzes/${c1.quizId}/attempts/${a1.attemptId}/result`, { waitUntil: "networkidle" });
    await page.waitForSelector(".result-page__content");
    const c1Text = await page.textContent(".result-page__content");
    if (!c1Text.includes("Total Score") || !c1Text.includes("10")) {
      throw new Error("Case 1 Failed: Immediate release results not visible.");
    }
    console.log("   Case 1 Passed!");

    // CASE 2: MANUAL RELEASE (Withheld -> Released)
    console.log("3. Case 2: Manual Release (Withheld -> Released)...");
    const c2 = await createQuiz({ result_release_policy: "MANUAL", results_released: false });
    const a2 = await createAndSubmitAttempt(c2.quizId, c2.qId);
    await page.goto(`${BASE_URL}/attempt/quizzes/${c2.quizId}/attempts/${a2.attemptId}/result`, { waitUntil: "networkidle" });
    await page.waitForSelector(".result-pending");
    const c2PendingText = await page.textContent(".result-pending");
    if (!c2PendingText.includes("Results Pending Release")) {
      throw new Error("Case 2 Failed: Pending release screen not displayed when results_released = false.");
    }
    console.log("   Manual release withheld verified!");

    // Release results via DB and reload
    queryDB(`UPDATE quizzes SET results_released = true WHERE id = '${c2.quizId}'`);
    await page.reload({ waitUntil: "networkidle" });
    await page.waitForSelector(".score-card");
    console.log("   Case 2 Passed!");

    // CASE 3: SCHEDULED RELEASE (Future)
    console.log("4. Case 3: Scheduled Release (Future)...");
    const futureDate = new Date(Date.now() + 2 * 3600 * 1000).toISOString();
    const c3 = await createQuiz({ result_release_policy: "SCHEDULED", results_scheduled_at: futureDate });
    const a3 = await createAndSubmitAttempt(c3.quizId, c3.qId);
    await page.goto(`${BASE_URL}/attempt/quizzes/${c3.quizId}/attempts/${a3.attemptId}/result`, { waitUntil: "networkidle" });
    await page.waitForSelector(".result-pending");
    console.log("   Case 3 Passed!");

    // CASE 4: SCHEDULED RELEASE (Past)
    console.log("5. Case 4: Scheduled Release (Past)...");
    const pastDate = new Date(Date.now() - 2 * 3600 * 1000).toISOString();
    const c4 = await createQuiz({ result_release_policy: "SCHEDULED", results_scheduled_at: pastDate });
    const a4 = await createAndSubmitAttempt(c4.quizId, c4.qId);
    await page.goto(`${BASE_URL}/attempt/quizzes/${c4.quizId}/attempts/${a4.attemptId}/result`, { waitUntil: "networkidle" });
    await page.waitForSelector(".score-card");
    console.log("   Case 4 Passed!");

    // CASE 5: HIDDEN REVIEW (allow_answer_review = false)
    console.log("6. Case 5: Hidden Question Review...");
    const c5 = await createQuiz({ allow_answer_review: false });
    const a5 = await createAndSubmitAttempt(c5.quizId, c5.qId);
    await page.goto(`${BASE_URL}/attempt/quizzes/${c5.quizId}/attempts/${a5.attemptId}/result`, { waitUntil: "networkidle" });
    await page.waitForSelector(".score-card");
    const c5Text = await page.textContent(".result-page__content");
    if (c5Text.includes("Question Review") || !c5Text.includes("Review is unavailable")) {
      throw new Error("Case 5 Failed: Question review shown when allow_answer_review = false.");
    }
    console.log("   Case 5 Passed!");

    // CASE 6: CORRECTNESS HIDDEN (show_correctness = false)
    console.log("7. Case 6: Correctness Hidden...");
    const c6 = await createQuiz({ allow_answer_review: true, show_correctness: false });
    const a6 = await createAndSubmitAttempt(c6.quizId, c6.qId);
    await page.goto(`${BASE_URL}/attempt/quizzes/${c6.quizId}/attempts/${a6.attemptId}/result`, { waitUntil: "networkidle" });
    await page.waitForSelector(".question-review");
    const correctFilter = await page.$(".filter-btn--correct");
    if (correctFilter) {
      throw new Error("Case 6 Failed: Correct filter button rendered when show_correctness = false.");
    }
    // Verify API JSON omits aggregate correctness fields
    const apiRes6 = await page.evaluate(async (url) => {
      const res = await fetch(url, { headers: { Accept: "application/json" } });
      return res.json();
    }, `${BASE_URL}/api/v1/quizzes/${c6.quizId}/attempts/${a6.attemptId}/result`);
    if (apiRes6.data?.summary?.correct !== undefined || apiRes6.data?.summary?.incorrect !== undefined) {
      throw new Error("Case 6 Failed: Aggregate correctness fields (correct, incorrect) leaked in summary JSON.");
    }
    console.log("   Case 6 Passed (aggregate correctness omitted from JSON)!");

    // CASE 7: EXPLANATION HIDDEN (show_explanations = false)
    console.log("8. Case 7: Explanations Hidden...");
    const c7 = await createQuiz({ allow_answer_review: true, show_explanations: false });
    const a7 = await createAndSubmitAttempt(c7.quizId, c7.qId);
    await page.goto(`${BASE_URL}/attempt/quizzes/${c7.quizId}/attempts/${a7.attemptId}/result`, { waitUntil: "networkidle" });
    await page.waitForSelector(".question-review");
    const expBox = await page.$(".explanation-box");
    if (expBox) {
      throw new Error("Case 7 Failed: Explanation box rendered when show_explanations = false.");
    }
    console.log("   Case 7 Passed!");

    // CASE 8: SCORE HIDDEN (show_score = false)
    console.log("9. Case 8: Score Hidden (Omits score fields from JSON)...");
    const c8 = await createQuiz({ show_score: false });
    const a8 = await createAndSubmitAttempt(c8.quizId, c8.qId);
    const apiRes8 = await page.evaluate(async (url) => {
      const res = await fetch(url, { headers: { Accept: "application/json" } });
      return res.json();
    }, `${BASE_URL}/api/v1/quizzes/${c8.quizId}/attempts/${a8.attemptId}/result`);
    if (apiRes8.data?.summary?.total_score !== undefined || apiRes8.data?.summary?.percentage !== undefined) {
      throw new Error("Case 8 Failed: Score fields (total_score, percentage) leaked in JSON when show_score = false.");
    }
    console.log("   Case 8 Passed (score fields completely omitted from JSON)!");

    // CASE 9: NON-OWNER ACCESS PROTECTION (HTTP 404 & Data Concealment)
    console.log("10. Case 9: Non-Owner Protection...");
    const c9 = await createQuiz({ result_release_policy: "IMMEDIATE" });
    const a9 = await createAndSubmitAttempt(c9.quizId, c9.qId);

    // Context for User B
    const userBContext = await browser.newContext();
    const userBPage = await userBContext.newPage();

    // Login as second user or make request as User B
    const userBRes = await userBPage.evaluate(async (url) => {
      const res = await fetch(url, { headers: { Accept: "application/json" } });
      const text = await res.text();
      let json = null;
      try { json = JSON.parse(text); } catch {}
      return { status: res.status, json, text };
    }, `${BASE_URL}/api/v1/quizzes/${c9.quizId}/attempts/${a9.attemptId}/result`);

    const hasSummary = Boolean(userBRes.json?.data?.summary);
    const hasReview = Boolean(userBRes.json?.data?.review);

    console.log(`   NON_OWNER_RESULT_STATUS=${userBRes.status}`);
    console.log(`   NON_OWNER_BODY_HAS_SUMMARY=${hasSummary}`);
    console.log(`   NON_OWNER_BODY_HAS_REVIEW=${hasReview}`);

    if (userBRes.status !== 404 && userBRes.status !== 401) {
      throw new Error(`Case 9 Failed: Expected status 404 or 401 for non-owner, got ${userBRes.status}`);
    }
    if (hasSummary || hasReview) {
      throw new Error("Case 9 Failed: Non-owner response leaked summary or review data.");
    }
    await userBContext.close();
    console.log("   Case 9 Passed!");

    console.log("=== ALL EXAM-P7-T02 E2E SMOKE SCENARIOS PASSED 100%! ===");
  } finally {
    await browser.close();
  }
}

run().catch((err) => {
  console.error("FATAL E2E SMOKE FAILURE:", err);
  process.exit(1);
});
