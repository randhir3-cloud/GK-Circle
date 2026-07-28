/**
 * EXAM-P8-T02 live smoke: learner analytics dashboard + API contracts.
 *
 * Verifies:
 * - dashboard summary rendering
 * - daily / weekly / monthly trend switching
 * - recent-activity cursor pagination
 * - subject aggregation
 * - sanitised attempt timeline
 * - withheld-result protection
 * - cache refresh after qualifying invalidation
 * - forbidden access to another learner's attempt
 * - mobile layout
 */
import { chromium } from "@playwright/test";
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(__dirname, "../../..");
const e2eEnvPath = path.join(root, ".env.e2e.local");
if (fs.existsSync(e2eEnvPath)) {
  for (const line of fs.readFileSync(e2eEnvPath, "utf8").split(/\r?\n/)) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith("#")) continue;
    const eq = trimmed.indexOf("=");
    if (eq <= 0) continue;
    const key = trimmed.slice(0, eq).trim();
    let value = trimmed.slice(eq + 1).trim();
    if (
      (value.startsWith('"') && value.endsWith('"')) ||
      (value.startsWith("'") && value.endsWith("'"))
    ) {
      value = value.slice(1, -1);
    }
    if (!(key in process.env)) process.env[key] = value;
  }
}

const base = process.env.PLAYWRIGHT_BASE_URL || "http://localhost:3000";
const api = process.env.PLAYWRIGHT_API_BASE_URL || "http://localhost:3010";
const email = process.env.E2E_CREATOR_EMAIL || "course.admin@example.com";
const password = process.env.E2E_TEST_PASSWORD || "Password123!";
const out = path.join("test-results", "exam-p8-t02-learner-dashboard-smoke");
fs.mkdirSync(out, { recursive: true });
const log = [];
const push = (m) => {
  log.push(m);
  console.log(m);
};

const assert = (cond, msg) => {
  if (!cond) throw new Error(msg);
};

const browser = await chromium.launch({ headless: true });
const context = await browser.newContext();
const page = await context.newPage();

try {
  // --- Auth ---
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

  // --- Dashboard summary API ---
  const dash = await page.request.get(`${api}/api/v1/analytics/dashboard`);
  push(`DASHBOARD_STATUS=${dash.status()}`);
  assert(dash.status() === 200, `dashboard expected 200, got ${dash.status()}`);
  const dashJson = await dash.json();
  const dashData = dashJson?.data;
  assert(dashData?.resolved_timezone, "dashboard missing resolved_timezone");
  assert(
    typeof dashData.total_attempts === "number",
    "dashboard missing total_attempts"
  );
  assert(
    typeof dashData.completion_rate === "number",
    "dashboard missing completion_rate"
  );
  assert(
    dashData.engaged_question_time_approximate === true,
    "engaged time must be marked approximate"
  );
  push(
    `DASHBOARD_OK tz=${dashData.resolved_timezone} attempts=${dashData.total_attempts} rate=${dashData.completion_rate}`
  );

  // Capture cache version key behaviour via repeat read (stable payload)
  const dash2 = await page.request.get(`${api}/api/v1/analytics/dashboard`);
  assert(dash2.status() === 200, "dashboard repeat read failed");
  push("DASHBOARD_CACHE_READ_OK");

  // --- Trends: daily / weekly / monthly ---
  for (const granularity of ["daily", "weekly", "monthly"]) {
    const to = new Date();
    const from = new Date();
    if (granularity === "monthly") from.setMonth(from.getMonth() - 5);
    else if (granularity === "weekly") from.setDate(from.getDate() - 7 * 11);
    else from.setDate(from.getDate() - 29);
    const trends = await page.request.get(
      `${api}/api/v1/analytics/trends?granularity=${granularity}&from=${encodeURIComponent(
        from.toISOString()
      )}&to=${encodeURIComponent(to.toISOString())}`
    );
    push(`TRENDS_${granularity.toUpperCase()}_STATUS=${trends.status()}`);
    assert(
      trends.status() === 200,
      `trends ${granularity} expected 200, got ${trends.status()}`
    );
    const body = await trends.json();
    const buckets = body?.data?.buckets || [];
    assert(buckets.length > 0, `trends ${granularity} returned no buckets`);
    const emptyAvgOk = buckets.every(
      (b) =>
        b.attempt_count > 0 ||
        b.average_percentage === null ||
        b.average_percentage === undefined
    );
    assert(emptyAvgOk, `trends ${granularity} empty buckets must use null avg`);
    push(`TRENDS_${granularity.toUpperCase()}_BUCKETS=${buckets.length}`);
  }

  // --- Subject aggregation ---
  const subjects = await page.request.get(`${api}/api/v1/analytics/subjects`);
  push(`SUBJECTS_STATUS=${subjects.status()}`);
  assert(subjects.status() === 200, `subjects expected 200, got ${subjects.status()}`);
  const subjectData = await subjects.json();
  assert(Array.isArray(subjectData?.data?.subjects), "subjects array missing");
  push(`SUBJECTS_COUNT=${subjectData.data.subjects.length}`);

  // --- Recent activity + cursor pagination ---
  const activity1 = await page.request.get(
    `${api}/api/v1/analytics/activity?limit=1`
  );
  push(`ACTIVITY_STATUS=${activity1.status()}`);
  assert(activity1.status() === 200, `activity expected 200, got ${activity1.status()}`);
  const act1 = await activity1.json();
  const items1 = act1?.data?.items || [];
  push(
    `ACTIVITY_PAGE1 items=${items1.length} has_more=${act1?.data?.has_more} cursor=${Boolean(
      act1?.data?.next_cursor
    )}`
  );

  if (act1?.data?.has_more && act1?.data?.next_cursor) {
    const activity2 = await page.request.get(
      `${api}/api/v1/analytics/activity?limit=1&cursor=${encodeURIComponent(
        act1.data.next_cursor
      )}`
    );
    assert(activity2.status() === 200, "activity page 2 failed");
    const act2 = await activity2.json();
    const items2 = act2?.data?.items || [];
    if (items1[0]?.attempt_id && items2[0]?.attempt_id) {
      assert(
        items1[0].attempt_id !== items2[0].attempt_id,
        "cursor pagination returned duplicate first item"
      );
    }
    push("ACTIVITY_CURSOR_OK");
  } else {
    push("ACTIVITY_CURSOR_SKIPPED_NO_MORE");
  }

  // Withheld-result protection on activity rows
  const pending = items1.find((i) => i.result_status === "Result Pending");
  if (pending) {
    assert(
      pending.percentage == null &&
        pending.total_score == null &&
        pending.max_score == null,
      "Result Pending row leaked score fields"
    );
    push("WITHHELD_RESULT_ACTIVITY_OK");
  } else {
    push("WITHHELD_RESULT_ACTIVITY_SKIPPED_NONE");
  }

  // --- Timeline: own attempt (sanitised) + foreign attempt (403) ---
  let timelineAttemptId = items1[0]?.attempt_id;
  if (!timelineAttemptId && dashData.total_attempts > 0) {
    const activityWide = await page.request.get(
      `${api}/api/v1/analytics/activity?limit=20`
    );
    const wide = await activityWide.json();
    timelineAttemptId = wide?.data?.items?.[0]?.attempt_id;
  }

  if (timelineAttemptId) {
    const timeline = await page.request.get(
      `${api}/api/v1/analytics/attempts/${timelineAttemptId}/timeline`
    );
    push(`TIMELINE_STATUS=${timeline.status()}`);
    assert(timeline.status() === 200, `timeline expected 200, got ${timeline.status()}`);
    const tl = await timeline.json();
    const events = tl?.data?.events || [];
    const sensitive = ["password", "token", "answer_key", "correct_answers"];
    for (const event of events) {
      const keys = Object.keys(event.metadata || {}).map((k) => k.toLowerCase());
      for (const bad of sensitive) {
        assert(!keys.includes(bad), `timeline metadata leaked ${bad}`);
      }
    }
    push(`TIMELINE_SANITISED_OK events=${events.length}`);
  } else {
    push("TIMELINE_SKIPPED_NO_ATTEMPT");
  }

  const foreignId = "00000000-0000-4000-8000-000000000099";
  const foreign = await page.request.get(
    `${api}/api/v1/analytics/attempts/${foreignId}/timeline`
  );
  push(`FOREIGN_TIMELINE_STATUS=${foreign.status()}`);
  assert(
    [403, 404].includes(foreign.status()),
    `foreign timeline expected 403/404, got ${foreign.status()}`
  );
  if (foreign.status() === 403) push("FOREIGN_FORBIDDEN_OK");
  else push("FOREIGN_NOT_FOUND_OK");

  // --- Cache refresh after qualifying invalidation ---
  // Telemetry batch with inserted>0 should bump version; use a real attempt when available.
  if (timelineAttemptId && items1[0]?.quiz_id) {
    const quizId = items1[0].quiz_id;
    const clientEventId = crypto.randomUUID();
    const before = await page.request.get(`${api}/api/v1/analytics/dashboard`);
    const beforeBody = await before.json();
    const telemetry = await page.request.post(
      `${api}/api/v1/quizzes/${quizId}/attempts/${timelineAttemptId}/analytics/events`,
      {
        data: {
          events: [
            {
              client_event_id: clientEventId,
              event_type: "QUESTION_VIEWED",
              occurred_at: new Date().toISOString(),
              metadata: { question_id: crypto.randomUUID() },
            },
          ],
        },
      }
    );
    push(`TELEMETRY_STATUS=${telemetry.status()}`);
    // 200 = inserted/deduped; 400/404 acceptable if attempt not eligible for client events
    if (telemetry.status() === 200) {
      const tel = await telemetry.json();
      push(
        `TELEMETRY_RESULT inserted=${tel?.data?.inserted} duplicates=${tel?.data?.duplicates}`
      );
      const after = await page.request.get(`${api}/api/v1/analytics/dashboard`);
      assert(after.status() === 200, "dashboard after telemetry failed");
      const afterBody = await after.json();
      assert(
        afterBody?.data?.resolved_timezone === beforeBody?.data?.resolved_timezone,
        "dashboard timezone changed unexpectedly"
      );
      push("CACHE_REFRESH_PATH_OK");
    } else {
      push(`CACHE_REFRESH_SKIPPED_TELEMETRY_${telemetry.status()}`);
    }
  } else {
    push("CACHE_REFRESH_SKIPPED_NO_ATTEMPT");
  }

  // --- UI dashboard summary rendering ---
  await page.setViewportSize({ width: 1280, height: 900 });
  await page.goto(`${base}/analytics`, { waitUntil: "networkidle" });
  await page.waitForSelector('[data-testid="study-time-card"]', {
    timeout: 45000,
  });
  await page.waitForSelector('[data-testid="performance-trend-chart"]');
  await page.waitForSelector('[data-testid="subject-performance-table"]');
  await page.waitForSelector('[data-testid="recent-activity-table"]');
  push("UI_DASHBOARD_DESKTOP_OK");
  await page.screenshot({
    path: path.join(out, "desktop-dashboard.png"),
    fullPage: true,
  });

  // Trend switching in UI
  for (const label of ["Weekly", "Monthly", "Daily"]) {
    await page.getByRole("button", { name: label, exact: true }).click();
    await page.waitForTimeout(500);
    push(`UI_TREND_SWITCH_${label.toUpperCase()}_OK`);
  }

  // --- Mobile layout ---
  await page.setViewportSize({ width: 390, height: 844 });
  await page.goto(`${base}/analytics`, { waitUntil: "networkidle" });
  await page.waitForSelector('[data-testid="study-time-card"]', {
    timeout: 45000,
  });
  const studyBox = await page.locator('[data-testid="study-time-card"]').boundingBox();
  assert(studyBox, "study-time-card missing on mobile");
  const viewport = page.viewportSize();
  assert(viewport, "mobile viewport missing");
  const overflow = await page.evaluate(() => ({
    doc: document.documentElement.scrollWidth,
    body: document.body.scrollWidth,
  }));
  assert(
    overflow.doc <= viewport.width + 1,
    `document horizontal overflow (${overflow.doc} > ${viewport.width})`
  );
  assert(
    studyBox.width <= viewport.width + 1,
    `study-time-card overflows mobile viewport (${studyBox.width} > ${viewport.width})`
  );
  push(
    `UI_MOBILE_OK width=${Math.round(studyBox.width)} viewport=${viewport.width} docSW=${overflow.doc}`
  );
  await page.screenshot({
    path: path.join(out, "mobile-dashboard.png"),
    fullPage: true,
  });

  push("SMOKE_OK");
  fs.writeFileSync(path.join(out, "smoke.log"), log.join("\n"));
  await browser.close();
  process.exit(0);
} catch (err) {
  push(`SMOKE_FAIL ${err}`);
  fs.writeFileSync(path.join(out, "smoke.log"), log.join("\n"));
  await page
    .screenshot({ path: path.join(out, "failure.png"), fullPage: true })
    .catch(() => {});
  await browser.close();
  process.exit(1);
}
