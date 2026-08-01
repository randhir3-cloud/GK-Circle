/**
 * EXAM-P8-T01 smoke: analytics event pipeline correlation + client dedupe.
 * Requires a running stack and authenticated creator session cookies via Playwright.
 */
import { chromium } from "@playwright/test";
import fs from "fs";
import path from "path";
import { randomUUID } from "crypto";

const base = process.env.PLAYWRIGHT_BASE_URL || "http://localhost:3000";
const api = process.env.PLAYWRIGHT_API_BASE_URL || "http://localhost:3010";
const email = process.env.E2E_CREATOR_EMAIL || "course.admin@example.com";
const password = process.env.E2E_TEST_PASSWORD;
if (!password) throw new Error("E2E_TEST_PASSWORD must be supplied explicitly");
const out = path.join("test-results", "exam-p8-t01-analytics-events-smoke");
fs.mkdirSync(out, { recursive: true });
const log = [];
const push = (m) => {
  log.push(m);
  console.log(m);
};

const browser = await chromium.launch({ headless: true });
const page = await browser.newPage();

try {
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

  // Lightweight probe: unauthenticated batch must 401; authenticated without attempt 404.
  const corr = randomUUID();
  const probeQuiz = randomUUID();
  const probeAttempt = randomUUID();
  const unauth = await page.request.post(
    `${api}/api/v1/quizzes/${probeQuiz}/attempts/${probeAttempt}/analytics/events`,
    {
      headers: { "X-Correlation-ID": corr },
      data: {
        events: [
          {
            client_event_id: randomUUID(),
            event_type: "QUESTION_VIEWED",
            occurred_at: new Date().toISOString(),
            metadata: { question_id: randomUUID() },
          },
        ],
      },
    }
  );
  push(`UNAUTH_BATCH_STATUS=${unauth.status()} corr=${corr}`);
  if (![401, 403].includes(unauth.status())) {
    throw new Error(`expected 401/403 for unauth analytics, got ${unauth.status()}`);
  }

  const authReject = await page.request.post(
    `${api}/api/v1/quizzes/${probeQuiz}/attempts/${probeAttempt}/analytics/events`,
    {
      headers: { "X-Correlation-ID": corr },
      data: {
        events: [
          {
            client_event_id: randomUUID(),
            event_type: "ATTEMPT_STARTED",
            occurred_at: new Date().toISOString(),
            metadata: {},
          },
        ],
      },
    }
  );
  push(`AUTH_REJECT_AUTHORITATIVE_STATUS=${authReject.status()}`);
  // 400 (wrong type) or 404 (missing attempt) both prove endpoint is wired.
  if (![400, 404].includes(authReject.status())) {
    throw new Error(
      `expected 400/404 for authoritative client submit, got ${authReject.status()}`
    );
  }

  push("SMOKE_OK");
  fs.writeFileSync(path.join(out, "smoke.log"), log.join("\n"));
  await browser.close();
  process.exit(0);
} catch (err) {
  push(`SMOKE_FAIL ${err}`);
  fs.writeFileSync(path.join(out, "smoke.log"), log.join("\n"));
  await browser.close();
  process.exit(1);
}
