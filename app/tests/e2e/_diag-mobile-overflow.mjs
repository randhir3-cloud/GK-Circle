import { chromium } from "@playwright/test";
import fs from "fs";
import path from "path";
import { fileURLToPath } from "url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../..");
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
const email = process.env.E2E_CREATOR_EMAIL || "course.admin@example.com";
const password = process.env.E2E_TEST_PASSWORD;
if (!password) throw new Error("E2E_TEST_PASSWORD must be supplied explicitly");

const browser = await chromium.launch({ headless: true });
const page = await browser.newPage({ viewport: { width: 390, height: 844 } });
await page.goto(`${base}/account/login`, { waitUntil: "networkidle" });
await page.locator('input[name="identifier"]').fill(email);
await page.locator('input[name="password"]').fill(password);
await Promise.all([
  page.waitForURL((u) => !u.pathname.includes("/account/login"), {
    timeout: 45000,
  }),
  page.locator('button[type="submit"]').first().click(),
]);
await page.goto(`${base}/analytics`, { waitUntil: "networkidle" });
await page.waitForSelector('[data-testid="study-time-card"]');

const info = await page.evaluate(() => {
  const el = document.querySelector('[data-testid="study-time-card"]');
  const rect = el.getBoundingClientRect();
  const chain = [];
  let n = el;
  while (n && n !== document.documentElement) {
    const s = getComputedStyle(n);
    const r = n.getBoundingClientRect();
    chain.push({
      tag: n.tagName,
      id: n.id || "",
      cls: String(n.className || "").slice(0, 140),
      w: Math.round(r.width * 100) / 100,
      left: Math.round(r.left * 100) / 100,
      right: Math.round(r.right * 100) / 100,
      mw: s.minWidth,
      box: s.boxSizing,
      overflow: s.overflowX,
      pad: `${s.paddingLeft}/${s.paddingRight}`,
      margin: `${s.marginLeft}/${s.marginRight}`,
    });
    n = n.parentElement;
  }
  const offenders = [...document.querySelectorAll("body *")]
    .map((node) => {
      const r = node.getBoundingClientRect();
      return {
        tag: node.tagName,
        cls: String(node.className || "").slice(0, 80),
        w: r.width,
        right: r.right,
        left: r.left,
      };
    })
    .filter((x) => x.right > window.innerWidth + 1 || x.w > window.innerWidth + 1)
    .slice(0, 20);
  return {
    rect: { w: rect.width, left: rect.left, right: rect.right },
    sw: document.documentElement.scrollWidth,
    iw: window.innerWidth,
    offenders,
    chain,
  };
});
console.log(JSON.stringify(info, null, 2));
await browser.close();
