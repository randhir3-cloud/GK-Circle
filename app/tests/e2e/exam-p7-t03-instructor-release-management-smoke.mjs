import { chromium } from "playwright";
import { execSync } from "child_process";

const WEB_BASE = process.env.WEB_BASE_URL || "http://localhost:3000";
const API_BASE = process.env.API_BASE_URL || "http://localhost:3010/api/v1";
const REPO_ROOT = "e:\\GK Circle v2";

function queryDB(sql) {
  const singleLineSql = sql.replace(/\s+/g, ' ').trim();
  const cmd = `docker compose exec -T db psql -U gk_circle -d gk_circle -t -A -c "${singleLineSql.replace(/"/g, '\\"')}"`;
  const out = execSync(cmd, { cwd: REPO_ROOT, encoding: "utf8" });
  const lines = out.trim().split('\n').map(l => l.trim()).filter(Boolean);
  return lines[0] || "";
}

async function runSmokeTests() {
  console.log("=== EXAM-P7-T03 Instructor Result Release Management Smoke Tests ===");

  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext();
  const page = await context.newPage();

  try {
    const timestamp = Date.now();
    const account = {
      email: `instructor.${timestamp}@example.test`,
      password: "Password123!",
      firstName: "Test",
      lastName: "Instructor"
    };

    console.log(`1. Registering instructor account: ${account.email}...`);
    await page.goto(`${WEB_BASE}/account/register`, { waitUntil: "networkidle" });
    await page.waitForSelector('input[name="traits.email"]', { timeout: 15000 });

    await page.locator('input[name="traits.name.first"]').fill(account.firstName);
    await page.locator('input[name="traits.name.last"]').fill(account.lastName);
    await page.locator('input[name="traits.email"]').fill(account.email);
    await page.locator('input[name="password"]').fill(account.password);
    
    await Promise.all([
      page.waitForNavigation({ waitUntil: "networkidle", timeout: 15000 }).catch(() => null),
      page.locator('button[type="submit"]').click()
    ]);

    await page.waitForTimeout(2000);

    const cookies = await context.cookies();
    console.log("   Cookies in context:", cookies.map(c => `${c.name}=${c.value.substring(0, 8)}... (${c.domain})`));

    // Fetch user ID from database with retry
    let userId = "";
    for (let i = 0; i < 5; i++) {
      userId = queryDB(`SELECT id FROM users WHERE email = '${account.email}' LIMIT 1`);
      if (userId) break;
      await new Promise(r => setTimeout(r, 1000));
    }
    if (!userId) {
      userId = queryDB("SELECT id FROM users ORDER BY created_at DESC LIMIT 1");
    }
    console.log(`   Instructor User ID: ${userId}`);

    // Helper for authenticated API calls via page.evaluate
    async function apiCall(method, path, body = null) {
      return await page.evaluate(async ({ method, path, body }) => {
        const opts = {
          method,
          headers: { "Content-Type": "application/json", Accept: "application/json" },
          credentials: "include"
        };
        if (body) opts.body = JSON.stringify(body);
        const res = await fetch(path, opts);
        const text = await res.text();
        let json = null;
        try { json = JSON.parse(text); } catch (e) {}
        return { status: res.status, data: json?.data || json, raw: text };
      }, { method, path: `http://localhost:3010${path}`, body });
    }

    // 2. Create quiz directly in DB owned by this user
    const quizTitle = `EXAM-P7-T03 Release Admin Quiz ${timestamp}`;
    const quizId = queryDB(
      `INSERT INTO quizzes (
        id, title, description, creator_id, is_public, assessment_mode, status, duration_seconds, max_attempts,
        negative_marks_per_question, allow_answer_review, result_release_policy, results_released,
        show_score, show_pass_fail, show_correctness, show_explanations, created_at, updated_at
      ) VALUES (
        gen_random_uuid(), '${quizTitle}', 'Release Admin Test Quiz', '${userId}', true, 'SELF_PACED', 'PUBLISHED', 1800, 1,
        0, true, 'IMMEDIATE', true, true, true, true, true, NOW(), NOW()
      ) RETURNING id`
    );
    console.log(`[Target Quiz Created]: ${quizId}`);

    // Case 1: Owner updates settings (SCHEDULED policy without date -> 400)
    console.log("\n[Case 3]: Testing SCHEDULED policy without date validation...");
    const badSchedResp = await apiCall("PATCH", `/api/v1/quizzes/${quizId}/results/settings`, {
      result_release_policy: "SCHEDULED",
      results_scheduled_at: null
    });
    console.log(` -> Status: ${badSchedResp.status}`);
    if (badSchedResp.status !== 400) {
      throw new Error(`Expected 400 when SCHEDULED policy missing date, got ${badSchedResp.status} body=${badSchedResp.raw}`);
    }
    console.log(" ✓ Case 3 Passed: Rejected invalid SCHEDULED date.");

    // Case 2: Owner updates settings to SCHEDULED with valid future date
    console.log("\n[Case 1]: Owner updates settings to SCHEDULED with future date...");
    const futureDate = new Date(Date.now() + 86400000).toISOString();
    const updateResp = await apiCall("PATCH", `/api/v1/quizzes/${quizId}/results/settings`, {
      result_release_policy: "SCHEDULED",
      results_scheduled_at: futureDate,
      show_score: true,
      show_pass_fail: true,
      allow_answer_review: true,
      show_correctness: true,
      show_explanations: true
    });
    console.log(` -> Status: ${updateResp.status}`);
    if (updateResp.data.result_release_policy !== "SCHEDULED" || updateResp.data.is_currently_released !== false) {
      throw new Error(`Settings update failed or released status incorrect: ${JSON.stringify(updateResp.data)}`);
    }
    console.log(" ✓ Case 1 Passed: Settings saved successfully.");

    // Case 3: Owner performs Manual Release Override
    console.log("\n[Case 2]: Owner performs manual release override on SCHEDULED quiz...");
    const releaseResp = await apiCall("POST", `/api/v1/quizzes/${quizId}/results/release`);
    console.log(` -> Status: ${releaseResp.status}`);
    if (!releaseResp.data.results_released || !releaseResp.data.is_currently_released) {
      throw new Error(`Manual release override failed: ${JSON.stringify(releaseResp.data)}`);
    }
    console.log(" ✓ Case 2 Passed: Manual release override successful.");

    // Case 4: Repeated Manual Release is Idempotent
    console.log("\n[Case 8]: Repeated manual release call (Idempotency check)...");
    const repeatReleaseResp = await apiCall("POST", `/api/v1/quizzes/${quizId}/results/release`);
    console.log(` -> Status: ${repeatReleaseResp.status}`);
    if (repeatReleaseResp.status !== 200) {
      throw new Error(`Repeated manual release failed with status ${repeatReleaseResp.status}`);
    }
    console.log(" ✓ Case 8 Passed: Idempotent manual release succeeded.");

    // Case 5: Transition to IMMEDIATE policy clears scheduled timestamp
    console.log("\n[Case 4]: Transition to IMMEDIATE policy clears scheduled date...");
    const immediateResp = await apiCall("PATCH", `/api/v1/quizzes/${quizId}/results/settings`, {
      result_release_policy: "IMMEDIATE"
    });
    if (immediateResp.data.results_scheduled_at !== null || immediateResp.data.is_currently_released !== true) {
      throw new Error(`IMMEDIATE transition did not clear scheduled date: ${JSON.stringify(immediateResp.data)}`);
    }
    console.log(" ✓ Case 4 Passed: IMMEDIATE policy cleared schedule and set results_released=true.");

    // Case 6: Manual release rejected on IMMEDIATE policy
    console.log("\n[Case 4b]: Manual release rejected on IMMEDIATE policy...");
    const badReleaseResp = await apiCall("POST", `/api/v1/quizzes/${quizId}/results/release`);
    console.log(` -> Status: ${badReleaseResp.status}`);
    if (badReleaseResp.status !== 400) {
      throw new Error(`Expected 400 when manually releasing IMMEDIATE policy, got ${badReleaseResp.status}`);
    }
    console.log(" ✓ Case 4b Passed: Manual release rejected on IMMEDIATE policy.");

    // Case 7: Learner access protection
    console.log("\n[Case 5]: Learner access protection on release admin endpoints...");
    const unauthContext = await browser.newContext();
    const unauthPage = await unauthContext.newPage();
    
    const unauthResp = await unauthPage.evaluate(async ({ url }) => {
      const res = await fetch(url, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ result_release_policy: "MANUAL" })
      });
      return res.status;
    }, { url: `http://localhost:3010/api/v1/quizzes/${quizId}/results/settings` });

    console.log(` -> Unauthenticated update status: ${unauthResp}`);
    if (unauthResp !== 401 && unauthResp !== 403) {
      throw new Error(`Expected 401/403 for unauthorized learner update, got ${unauthResp}`);
    }
    console.log(" ✓ Case 5 Passed: Unauthenticated learner blocked with HTTP 401/403.");
    await unauthContext.close();

    // Case 8: Fetch release status endpoint
    console.log("\n[Case 6]: Fetching release status metadata...");
    const statusResp = await apiCall("GET", `/api/v1/quizzes/${quizId}/results/status`);
    console.log(` -> Status: ${statusResp.status}`);
    if (!statusResp.data.quiz_id || statusResp.data.total_submitted_attempts === undefined) {
      throw new Error(`Release status response invalid: ${JSON.stringify(statusResp.data)}`);
    }
    console.log(" ✓ Case 6 Passed: Release status endpoint returned valid metadata.");

    // Case 9: Page refresh state persistence
    console.log("\n[Case 9]: Verifying page refresh state persistence...");
    await page.goto(`${WEB_BASE}/quiz/${quizId}/results/manage`, { waitUntil: "domcontentloaded" });
    await page.waitForSelector("h1", { timeout: 10000 });
    const pageText = await page.textContent("body");
    if (!pageText.includes("Result Release Administration")) {
      throw new Error("Result release management page failed to load rendered state.");
    }
    console.log(" ✓ Case 9 Passed: Management UI rendered persisted state successfully.");

    console.log("\n========================================================");
    console.log("🎉 ALL EXAM-P7-T03 SMOKE TEST SCENARIOS PASSED 100% 🎉");
    console.log("========================================================");

  } catch (err) {
    console.error("\n❌ SMOKE TEST FAILED:", err);
    process.exitCode = 1;
  } finally {
    await browser.close();
  }
}

runSmokeTests();
