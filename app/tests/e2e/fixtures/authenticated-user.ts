import {
  expect,
  type Browser,
  type BrowserContext,
  type Page,
} from "@playwright/test";

export interface TestAccount {
  email: string;
  password: string;
  firstName: string;
  lastName: string;
}

export interface AuthenticatedUser {
  context: BrowserContext;
  page: Page;
}

const authenticatedPages = new WeakMap<BrowserContext, Page>();

export function pageForContext(context: BrowserContext): Page {
  const page = authenticatedPages.get(context);
  if (!page) throw new Error("No authenticated page is registered for context");
  return page;
}

function requiredEnvironment(name: string): string {
  const value = process.env[name]?.trim();
  if (!value)
    throw new Error(`${name} is required for authenticated E2E tests`);
  return value;
}

export function authenticatedTestEnvironment() {
  return {
    webBaseUrl: requiredEnvironment("PLAYWRIGHT_BASE_URL"),
    apiBaseUrl: requiredEnvironment("PLAYWRIGHT_API_BASE_URL"),
    password: requiredEnvironment("E2E_TEST_PASSWORD"),
    creatorEmail: requiredEnvironment("E2E_CREATOR_EMAIL"),
    studentEmail: requiredEnvironment("E2E_STUDENT_EMAIL"),
    otherStudentEmail: requiredEnvironment("E2E_OTHER_STUDENT_EMAIL"),
    otherCreatorEmail: requiredEnvironment("E2E_OTHER_CREATOR_EMAIL"),
  };
}

export async function login(page: Page, account: TestAccount): Promise<void> {
  await page.goto("/account/login", { waitUntil: "domcontentloaded" });
  await expect(page.getByText("Loading…")).toHaveCount(0, { timeout: 20_000 });
  const identifier = page.locator('input[name="identifier"]');
  if (await identifier.isEditable()) {
    await identifier.fill(account.email);
  }
  await page.locator('input[name="password"]').fill(account.password);
  await Promise.all([
    page.waitForURL((url) => !url.pathname.includes("/account/login"), {
      timeout: 45_000,
    }),
    page.locator('button[type="submit"]').click(),
  ]);
}

export async function registerAuthenticatedUser(
  browser: Browser,
  webBaseUrl: string,
  account: TestAccount
): Promise<AuthenticatedUser> {
  const context = await browser.newContext({ baseURL: webBaseUrl });
  const page = await context.newPage();

  await page.goto("/account/register", { waitUntil: "domcontentloaded" });
  await expect(page.getByText("Loading…")).toHaveCount(0, { timeout: 20_000 });
  await page.locator('input[name="traits.name.first"]').fill(account.firstName);
  await page.locator('input[name="traits.name.last"]').fill(account.lastName);
  await page.locator('input[name="traits.email"]').fill(account.email);
  await page.locator('input[name="password"]').fill(account.password);
  const registered = await Promise.all([
    page.waitForURL((url) => !url.pathname.includes("/account/register"), {
      timeout: 20_000,
    }),
    page.locator('button[type="submit"]').click(),
  ])
    .then(() => true)
    .catch(() => false);
  if (!registered) {
    await login(page, account);
  } else {
    await page
      .waitForURL((url) => url.origin === new URL(webBaseUrl).origin, {
        timeout: 45_000,
      })
      .catch(() => undefined);
  }

  await page.goto("/admin", { waitUntil: "domcontentloaded" });
  if (page.url().includes("/account/login")) await login(page, account);
  await page.goto("/admin", { waitUntil: "domcontentloaded" });
  await expect(page.getByText("Profile Settings")).toBeVisible({
    timeout: 30_000,
  });

  authenticatedPages.set(context, page);
  return { context, page };
}
