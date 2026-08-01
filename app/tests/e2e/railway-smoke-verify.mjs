import { chromium } from "@playwright/test";
import fs from "fs";
import path from "path";

const base = process.env.BASE_URL || process.env.PLAYWRIGHT_BASE_URL || "https://gateway-production-ed03.up.railway.app";
const outDir = path.join("test-results", "railway-smoke-verify");
fs.mkdirSync(outDir, { recursive: true });

let passed = 0;
let failed = 0;

const assertCondition = (name, cond, msg) => {
  if (cond) {
    passed++;
    console.log(`[PASS] ${name}`);
  } else {
    failed++;
    console.log(`[FAIL] ${name}: ${msg}`);
  }
};

const browser = await chromium.launch({ headless: true });
const page = await browser.newPage();

const consoleErrors = [];
page.on("pageerror", (error) => {
  consoleErrors.push(String(error));
});
page.on("console", (msg) => {
  if (msg.type() === "error") {
    const text = msg.text();
    if (!text.includes("status of 401")) {
      consoleErrors.push(text);
    }
  }
});

const failedRequests = [];
page.on("requestfailed", (request) => {
  failedRequests.push(`${request.url()}: ${request.failure()?.errorText || "unknown error"}`);
});

try {
  // 1. Homepage loads
  console.log(`Loading homepage: ${base}/`);
  const homeRes = await page.goto(`${base}/`, { waitUntil: "load" });
  assertCondition("Homepage loads with 200", homeRes.status() === 200, `got status ${homeRes.status()}`);
  
  // Check static assets
  const entries = await page.evaluate(() => {
    return performance.getEntriesByType("resource").map(r => r.name);
  });
  const hasCSS = entries.some(name => name.includes(".css")) || await page.evaluate(() => document.querySelectorAll("style").length > 0);
  const hasJS = entries.some(name => name.includes(".js"));

  assertCondition("Static CSS loaded", hasCSS, "no CSS files or style blocks detected");
  assertCondition("Static JS loaded", hasJS, "no JS files detected in resources");

  // 2. Gateway /healthz
  console.log("Checking gateway health...");
  const gwHealth = await page.request.get(`${base}/healthz`);
  assertCondition("Gateway health status 200", gwHealth.status() === 200, `got status ${gwHealth.status()}`);

  // 3. API health
  console.log("Checking API health...");
  const apiHealth = await page.request.get(`${base}/api/healthz`);
  assertCondition("API health status 200", apiHealth.status() === 200, `got status ${apiHealth.status()}`);

  // 4. Public quizzes
  console.log("Checking public quizzes endpoint...");
  const quizzes = await page.request.get(`${base}/api/v1/quizzes/public`);
  assertCondition("Public quizzes status 2xx", quizzes.status() >= 200 && quizzes.status() < 300, `got status ${quizzes.status()}`);

  // 5. Kratos readiness
  console.log("Checking Kratos readiness...");
  const kratos = await page.request.get(`${base}/kratos/health/ready`);
  assertCondition("Kratos readiness status 200", kratos.status() === 200, `got status ${kratos.status()}`);

  // 6. Registration page
  console.log("Checking registration route...");
  const regRes = await page.goto(`${base}/account/register`, { waitUntil: "load" });
  assertCondition("Registration route loads (2xx/3xx)", regRes.status() >= 200 && regRes.status() < 400, `got status ${regRes.status()}`);

  // 7. Login page
  console.log("Checking login route...");
  const loginRes = await page.goto(`${base}/account/login`, { waitUntil: "load" });
  assertCondition("Login route loads (2xx/3xx)", loginRes.status() >= 200 && loginRes.status() < 400, `got status ${loginRes.status()}`);

  // 8. Console errors check
  assertCondition("No critical console errors", consoleErrors.length === 0, `errors: ${consoleErrors.join(" | ")}`);

  // 9. Failed network requests check
  assertCondition("No failed network requests", failedRequests.length === 0, `failed: ${failedRequests.join(" | ")}`);

  await page.screenshot({ path: path.join(outDir, "smoke-success.png"), fullPage: true });

} catch (err) {
  failed++;
  console.log(`[FAIL] Verification execution error: ${err.message}`);
} finally {
  await browser.close();
  
  console.log(`\nFinal Results: Passed: ${passed}, Failed: ${failed}`);
  if (failed > 0) {
    process.exit(1);
  } else {
    process.exit(0);
  }
}
