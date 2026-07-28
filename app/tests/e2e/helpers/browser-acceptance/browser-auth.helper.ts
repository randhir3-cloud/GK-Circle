import { Page, BrowserContext } from "@playwright/test";

export async function loginAsAdmin(
  page: Page,
  baseUrl: string,
  email: string,
  pass: string
): Promise<void> {
  await page.goto(`${baseUrl}/account/login`, { waitUntil: "networkidle" });
  await page.fill(
    'input[type="email"], input[name="identifier"], input[name="email"]',
    email
  );
  await page.fill('input[type="password"]', pass);
  await Promise.all([
    page.waitForNavigation({ waitUntil: "networkidle" }).catch(() => null),
    page.click('button[type="submit"]'),
  ]);
}

export async function loginAsLearner(
  context: BrowserContext,
  baseUrl: string,
  email: string,
  pass: string
): Promise<Page> {
  const page = await context.newPage();
  await page.goto(`${baseUrl}/account/login`, { waitUntil: "networkidle" });
  await page.fill(
    'input[type="email"], input[name="identifier"], input[name="email"]',
    email
  );
  await page.fill('input[type="password"]', pass);
  await Promise.all([
    page.waitForNavigation({ waitUntil: "networkidle" }).catch(() => null),
    page.click('button[type="submit"]'),
  ]);
  return page;
}

export async function logout(page: Page, baseUrl: string): Promise<void> {
  await page
    .goto(`${baseUrl}/logout`, { waitUntil: "networkidle" })
    .catch(() => null);
  await page.context().clearCookies();
}
