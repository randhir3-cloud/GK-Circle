import { chromium, expect } from '@playwright/test';
import fs from 'fs';
import path from 'path';

const base = process.env.PLAYWRIGHT_BASE_URL;
const api = process.env.PLAYWRIGHT_API_BASE_URL || 'http://localhost:3010';
const email = process.env.E2E_CREATOR_EMAIL;
const password = process.env.E2E_TEST_PASSWORD;
const outDir = path.join('test-results', 'repair-smoke-full');
fs.mkdirSync(outDir, { recursive: true });
const log = [];
const push = (m) => { log.push(m); console.log(m); };

const browser = await chromium.launch({ headless: true });
const context = await browser.newContext({ baseURL: base });
const page = await context.newPage();
const consoleErrors = [];
page.on('pageerror', (e) => consoleErrors.push(String(e)));
page.on('console', (msg) => { if (msg.type() === 'error' && !/favicon|Download the Vue/.test(msg.text())) consoleErrors.push(msg.text()); });

try {
  await page.goto('/', { waitUntil: 'networkidle' });
  push('HOME=' + await page.getByText('GK Circle').first().isVisible());

  await page.goto('/account/login');
  await page.locator('input[name="identifier"]').fill(email);
  await page.locator('input[name="password"]').fill(password);
  await Promise.all([
    page.waitForURL((u) => !u.pathname.includes('/account/login'), { timeout: 45000 }),
    page.locator('button[type="submit"]').click(),
  ]);
  push('LOGIN_OK');

  await page.goto('/admin/quiz/list-quiz');
  await page.waitForTimeout(1500);
  push('QUIZ_LIST=' + page.url());
  await page.screenshot({ path: path.join(outDir, '01-quiz-list.png'), fullPage: true });

  // Live surfaces that exist in current UI
  for (const route of ['/admin/reports']) {
    await page.goto(route);
    await page.waitForTimeout(1000);
    push('ROUTE ' + route + ' body=' + (await page.locator('body').innerText()).trim().slice(0,80).replace(/\s+/g,' '));
  }

  // Attempt instructions shell (fresh build)
  await page.goto('/attempt/quizzes/demo?snapshot_id=demo');
  await page.waitForTimeout(1500);
  const attemptText = (await page.locator('body').innerText()).trim();
  push('ATTEMPT_SHELL_VISIBLE=' + (attemptText.length > 20));
  push('ATTEMPT_SNIPPET=' + attemptText.slice(0, 160).replace(/\s+/g, ' '));
  await page.screenshot({ path: path.join(outDir, '02-attempt-shell.png'), fullPage: true });

  // Probe API healthz through browser cookies
  const apiHealth = await page.request.get(api + '/api/healthz/');
  push('API_HEALTH=' + apiHealth.status());

  push('CONSOLE_ERRORS=' + consoleErrors.length);
  if (consoleErrors.length) push('CONSOLE=' + consoleErrors.slice(0,6).join(' || '));
  fs.writeFileSync(path.join(outDir, 'smoke.log'), log.join('\n'));
  push('SMOKE_FULL_DONE');
} catch (e) {
  push('FAIL=' + e.message);
  await page.screenshot({ path: path.join(outDir, 'fail.png'), fullPage: true }).catch(()=>{});
  fs.writeFileSync(path.join(outDir, 'smoke.log'), log.join('\n'));
  process.exitCode = 1;
} finally {
  await browser.close();
}
