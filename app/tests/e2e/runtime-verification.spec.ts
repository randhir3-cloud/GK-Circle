import {
  test,
  expect,
  type Browser,
  type BrowserContext,
  type Page,
  type Request,
  type Response,
} from "@playwright/test";
import * as fs from "fs";
import { login } from "./fixtures/authenticated-user.js";
import * as path from "path";
import { fileURLToPath } from "url";

interface RequestFailure {
  url: string;
  failureText: string;
}

interface ResponseFailure {
  url: string;
  status: number;
  statusText: string;
}

interface Diagnostics {
  consoleErrors: string[];
  pageErrors: string[];
  failedRequests: RequestFailure[];
  badResponses: ResponseFailure[];
}

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const screenshotDir = path.join(
  __dirname,
  "../../test-results/runtime-verification/screenshots"
);

if (!fs.existsSync(screenshotDir)) {
  fs.mkdirSync(screenshotDir, { recursive: true });
}

const HARMLESS_CONSOLE: RegExp[] = [
  /Download the Vue Devtools/i,
  /\[HMR\]/i,
  /favicon\.ico/i,
];

function isHarmlessConsole(message: string): boolean {
  return HARMLESS_CONSOLE.some((pattern) => pattern.test(message));
}

function isAssetUrl(url: string): boolean {
  return /\.(js|css|png|jpe?g|gif|svg|webp|woff2?|ttf|ico)(\?|$)/i.test(url);
}

function isApiOrKratosUrl(url: string): boolean {
  return (
    url.includes("localhost:3010") ||
    url.includes("localhost:4433") ||
    url.includes("/api/") ||
    url.includes("/self-service/")
  );
}

function createDiagnostics(): Diagnostics {
  return {
    consoleErrors: [],
    pageErrors: [],
    failedRequests: [],
    badResponses: [],
  };
}

function attachDiagnostics(page: Page, diagnostics: Diagnostics): void {
  page.on("console", (msg) => {
    if (msg.type() === "error") {
      const text = msg.text();
      if (!isHarmlessConsole(text)) {
        diagnostics.consoleErrors.push(text);
      }
    }
  });

  page.on("pageerror", (err) => {
    diagnostics.pageErrors.push(err.message);
  });

  page.on("requestfailed", (req: Request) => {
    const sanitizedUrl = req
      .url()
      .replace(/(token|password|code|verification)=[^&]+/gi, "$1=[REDACTED]");
    diagnostics.failedRequests.push({
      url: sanitizedUrl,
      failureText: req.failure()?.errorText || "Unknown error",
    });
  });

  page.on("response", (res: Response) => {
    if (res.status() >= 400) {
      const sanitizedUrl = res
        .url()
        .replace(/(token|password|code|verification)=[^&]+/gi, "$1=[REDACTED]");
      diagnostics.badResponses.push({
        url: sanitizedUrl,
        status: res.status(),
        statusText: res.statusText(),
      });
    }
  });
}

function resetDiagnostics(diagnostics: Diagnostics): void {
  diagnostics.consoleErrors = [];
  diagnostics.pageErrors = [];
  diagnostics.failedRequests = [];
  diagnostics.badResponses = [];
}

function assertNoCriticalFailures(
  diagnostics: Diagnostics,
  context: string
): void {
  const connectionFailures = diagnostics.failedRequests.filter(
    (f) =>
      f.failureText.includes("ERR_CONNECTION_REFUSED") ||
      f.failureText.includes("ERR_NAME_NOT_RESOLVED") ||
      /Failed to fetch/i.test(f.failureText)
  );
  expect(
    connectionFailures,
    `${context}: connection failures ${JSON.stringify(connectionFailures)}`
  ).toHaveLength(0);

  expect(
    diagnostics.pageErrors,
    `${context}: page errors ${diagnostics.pageErrors.join(" | ")}`
  ).toHaveLength(0);

  const hydrationErrors = diagnostics.consoleErrors.filter((e) =>
    /hydrat|mismatch/i.test(e)
  );
  expect(
    hydrationErrors,
    `${context}: hydration errors ${hydrationErrors.join(" | ")}`
  ).toHaveLength(0);

  const corsErrors = diagnostics.consoleErrors.filter((e) =>
    /CORS|Cross-Origin/i.test(e)
  );
  expect(
    corsErrors,
    `${context}: CORS errors ${corsErrors.join(" | ")}`
  ).toHaveLength(0);

  const failedAssets = diagnostics.failedRequests.filter(
    (f) =>
      isAssetUrl(f.url) &&
      // Navigations abort in-flight chunk fetches; that is not an asset outage.
      !f.failureText.includes("ERR_ABORTED")
  );
  expect(
    failedAssets,
    `${context}: failed assets ${JSON.stringify(failedAssets)}`
  ).toHaveLength(0);

  const apiServerErrors = diagnostics.badResponses.filter(
    (r) => isApiOrKratosUrl(r.url) && r.status >= 500
  );
  expect(
    apiServerErrors,
    `${context}: API/Kratos 5xx ${JSON.stringify(apiServerErrors)}`
  ).toHaveLength(0);
}

function assertNoRawTechnicalLeak(pageContent: string, context: string): void {
  expect(pageContent, `${context}: leaked raw [GET] error`).not.toMatch(
    /\[GET\]\s+http/i
  );
  expect(pageContent, `${context}: leaked internal kratos host`).not.toContain(
    "http://kratos:"
  );
  expect(pageContent, `${context}: leaked internal api host`).not.toContain(
    "http://api:"
  );
}

async function waitForAuthFormReady(page: Page): Promise<void> {
  await expect(page.getByText("Loading…")).toHaveCount(0, { timeout: 20_000 });
  await expect(page.getByText("Service Offline")).toHaveCount(0);
}

async function captureShot(page: Page, name: string): Promise<void> {
  if (process.env.PLAYWRIGHT_SANITIZED_EVIDENCE === "true") return;
  await page.screenshot({
    path: path.join(screenshotDir, name),
    fullPage: true,
  });
}

test.describe.configure({ mode: "serial" });

test.describe("Unauthenticated page checks", () => {
  const diagnostics = createDiagnostics();

  test.beforeEach(({ page }) => {
    resetDiagnostics(diagnostics);
    attachDiagnostics(page, diagnostics);
  });

  test("A. Homepage loads with GK Circle brand", async ({ page }) => {
    const response = await page.goto("/", { waitUntil: "networkidle" });
    expect(response?.ok()).toBeTruthy();

    await expect(page.getByText("GK Circle").first()).toBeVisible();
    const bodyText = await page.locator("body").innerText();
    expect(bodyText.trim().length).toBeGreaterThan(20);

    assertNoCriticalFailures(diagnostics, "homepage");
    await captureShot(page, "homepage.png");
  });

  test("B. Login page renders form when Kratos is healthy", async ({
    page,
  }) => {
    const flowPromise = page.waitForResponse(
      (res) =>
        res.url().includes("/self-service/login/browser") &&
        res.request().method() === "GET",
      { timeout: 20_000 }
    );

    await page.goto("/account/login", { waitUntil: "domcontentloaded" });
    const flowResponse = await flowPromise;
    expect(flowResponse.status(), "Kratos login flow status").toBeLessThan(400);

    await waitForAuthFormReady(page);
    await expect(page.locator('input[name="identifier"]')).toBeVisible();
    await expect(page.locator('input[name="password"]')).toBeVisible();
    await expect(
      page.locator('button[type="submit"], button:has-text("Sign In")').first()
    ).toBeVisible();

    const content = await page.locator("body").innerText();
    assertNoRawTechnicalLeak(content, "login");
    expect(content).not.toContain("localhost:4433");

    assertNoCriticalFailures(diagnostics, "login");
    await captureShot(page, "login.png");
  });

  test("C. Registration page renders form", async ({ page }) => {
    const flowPromise = page.waitForResponse(
      (res) =>
        res.url().includes("/self-service/registration/browser") &&
        res.request().method() === "GET",
      { timeout: 20_000 }
    );

    await page.goto("/account/register", { waitUntil: "domcontentloaded" });
    const flowResponse = await flowPromise;
    expect(
      flowResponse.status(),
      "Kratos registration flow status"
    ).toBeLessThan(400);

    await waitForAuthFormReady(page);
    await expect(page.locator('input[name="traits.name.first"]')).toBeVisible();
    await expect(page.locator('input[name="traits.name.last"]')).toBeVisible();
    await expect(page.locator('input[name="traits.email"]')).toBeVisible();
    await expect(page.locator('input[name="password"]')).toBeVisible();
    await expect(
      page
        .locator(
          'button[type="submit"], button:has-text("Sign Up"), button:has-text("Create Account")'
        )
        .first()
    ).toBeVisible();

    const content = await page.locator("body").innerText();
    assertNoRawTechnicalLeak(content, "registration");

    assertNoCriticalFailures(diagnostics, "registration");
    await captureShot(page, "registration.png");
  });

  test("D. Unauthenticated protected routes redirect to login", async ({
    page,
  }) => {
    for (const route of ["/admin"]) {
      await page.goto(route, { waitUntil: "domcontentloaded" });
      await page.waitForURL(/\/account\/login/, { timeout: 20_000 });

      const content = await page.locator("body").innerText();
      expect(content.trim().length, `${route} blank content`).toBeGreaterThan(
        10
      );
      assertNoRawTechnicalLeak(content, route);
    }

    assertNoCriticalFailures(diagnostics, "protected-routes");
  });
});

test.describe("Authenticated end-to-end flow", () => {
  test.describe.configure({ mode: "serial" });

  const diagnostics = createDiagnostics();
  const testStartTime = Date.now();

  const baseURL = process.env.PLAYWRIGHT_BASE_URL;
  const mailpitURL = process.env.MAILPIT_API_URL;
  const password = process.env.PLAYWRIGHT_TEST_PASSWORD;
  const runID = process.env.E2E_RUN_ID;

  if (!baseURL || !mailpitURL || !password || !runID) {
    throw new Error(
      "Sanitized configuration error: missing required E2E variables (PLAYWRIGHT_BASE_URL, MAILPIT_API_URL, PLAYWRIGHT_TEST_PASSWORD, E2E_RUN_ID)"
    );
  }

  if (!/^[a-zA-Z0-9_-]+$/.test(runID)) {
    throw new Error(
      "Sanitized configuration error: invalid E2E_RUN_ID format."
    );
  }

  const email = `gkcircle.e2e.${runID}@example.test`;
  const quizTitle = `E2E Quiz ${runID}`;

  let browserRef: Browser;
  let context: BrowserContext;
  let page: Page;
  let emailVerificationRequired = false;

  test.beforeAll(async ({ browser }) => {
    browserRef = browser;
    context = await browserRef.newContext({
      baseURL: baseURL,
    });
    page = await context.newPage();
    attachDiagnostics(page, diagnostics);
  });

  test.afterAll(async () => {
    await context.close();
  });

  test.beforeEach(() => {
    resetDiagnostics(diagnostics);
  });

  test("E. Registration creates authenticated session", async () => {
    test.setTimeout(120_000);

    await page.goto("/account/register", { waitUntil: "domcontentloaded" });
    await waitForAuthFormReady(page);

    await page.fill('input[name="traits.name.first"]', "Runtime");
    await page.fill('input[name="traits.name.last"]', "Verifier");
    await page.fill('input[name="traits.email"]', email);
    await page.fill('input[name="password"]', password);

    await Promise.all([
      page.waitForURL(
        (url) =>
          !url.pathname.includes("/account/register") ||
          url.hostname.includes("3010"),
        { timeout: 45_000 }
      ),
      page.locator('button[type="submit"]').click(),
    ]);

    // Allow API auth callback redirect chain to settle on the Nuxt origin
    await page
      .waitForURL(/localhost:3000/, { timeout: 45_000 })
      .catch(() => undefined);

    if (page.url().includes("/verification")) {
      emailVerificationRequired = true;
    }

    // Poll Mailpit until the verification email arrives or timeout is reached
    let verificationMail;
    for (let attempt = 0; attempt < 10; attempt++) {
      const mailpitRes = await fetch(`${mailpitURL}/api/v1/messages`).catch(
        () => null
      );
      if (mailpitRes && mailpitRes.ok) {
        const mailData = (await mailpitRes.json()) as {
          messages: Array<{
            ID: string;
            Subject: string;
            To: Array<{ Address: string }>;
            Created: string;
          }>;
        };
        verificationMail = mailData.messages.find((m) => {
          const createdTime = Date.parse(m.Created);
          return (
            m.To?.some((to) => to.Address === email) &&
            createdTime >= testStartTime
          );
        });
        if (verificationMail) break;
      }
      await page.waitForTimeout(1000);
    }

    if (verificationMail) {
      emailVerificationRequired = true;
      const detailRes = await fetch(
        `${mailpitURL}/api/v1/message/${verificationMail.ID}`
      );
      if (detailRes.ok) {
        const detailData = (await detailRes.json()) as {
          HTML: string;
          Text: string;
        };
        const bodyText = detailData.HTML || detailData.Text;
        const linkMatch =
          bodyText.match(/href="([^"]*verification[^"]*)"/i) ||
          bodyText.match(/(https?:\/\/[^\s"'<>]+verification[^\s"'<>]*)/i);
        if (linkMatch) {
          const verificationLink = linkMatch[1].replace(/&amp;/g, "&");
          await page.goto(verificationLink, { waitUntil: "domcontentloaded" });
          await page
            .waitForURL(
              (url) =>
                url.pathname.includes("/account/login") ||
                url.pathname.includes("/admin"),
              { timeout: 15_000 }
            )
            .catch(() => null);
        }
      }
    }

    // Prefer authenticated navigation; only use login if session is missing.
    await page.goto("/admin", { waitUntil: "domcontentloaded" });
    if (page.url().includes("/account/login")) {
      await login(page, {
        email,
        password,
        firstName: "Runtime",
        lastName: "Verifier",
      });
      await page.goto("/admin", { waitUntil: "domcontentloaded" });
    }

    await expect(page.getByText("Profile Settings")).toBeVisible({
      timeout: 30_000,
    });

    const cookies = await context.cookies();
    const sessionCookie = cookies.find(
      (c) =>
        c.name.toLowerCase().includes("ory_kratos_session") ||
        c.name.toLowerCase().includes("session")
    );
    expect(sessionCookie, "session cookie exists").toBeTruthy();

    assertNoCriticalFailures(diagnostics, "registration-auth");
    void emailVerificationRequired;
  });

  test("F. Session refresh and profile content", async () => {
    test.setTimeout(60_000);

    await page.goto("/admin", { waitUntil: "domcontentloaded" });
    await expect(page.getByText("Pending...")).toHaveCount(0, {
      timeout: 20_000,
    });
    await expect(page.getByText("Profile Settings")).toBeVisible({
      timeout: 20_000,
    });
    await expect(page.locator("#first-name")).toHaveValue("Runtime", {
      timeout: 20_000,
    });
    await expect(page.locator("#last-name")).toHaveValue("Verifier");

    await page.reload({ waitUntil: "domcontentloaded" });
    await expect(page.getByText("Pending...")).toHaveCount(0, {
      timeout: 20_000,
    });
    await expect(page.locator("#first-name")).toHaveValue("Runtime");

    const content = await page.locator("body").innerText();
    assertNoRawTechnicalLeak(content, "profile");
    expect(content).not.toMatch(/Failed to fetch/i);

    assertNoCriticalFailures(diagnostics, "profile");
    await captureShot(page, "profile.png");
  });

  test("G. Quiz list loads from Go API", async () => {
    test.setTimeout(60_000);

    await page.goto("/admin/quiz/list-quiz", { waitUntil: "domcontentloaded" });

    await expect(page.getByText("Quiz List")).toBeVisible();
    await expect(page.getByText("Loading...")).toHaveCount(0, {
      timeout: 20_000,
    });
    await expect(page.getByText("Create Quiz").first()).toBeVisible();

    const filterButton = page
      .locator('button[aria-haspopup="listbox"]')
      .first();
    await expect(filterButton).toBeVisible();
    await filterButton.click();
    await page.keyboard.press("Escape").catch(() => undefined);

    const search = page
      .locator('input[type="search"][placeholder*="Search"]')
      .first();
    await expect(search).toBeVisible();
    await search.fill("nonexistent-e2e-query-xyz");
    await search.fill("");

    const content = await page.locator("body").innerText();
    assertNoRawTechnicalLeak(content, "quiz-list");
    expect(content).not.toMatch(/Failed to fetch/i);

    assertNoCriticalFailures(diagnostics, "quiz-list");
    await captureShot(page, "quiz_list.png");
  });

  test("H. Create quiz, add question, persist and search", async () => {
    test.setTimeout(120_000);

    await page.goto("/admin/quiz/list-quiz", { waitUntil: "domcontentloaded" });
    await expect(page.getByText("Quiz List")).toBeVisible();

    await page.getByText("Create Quiz", { exact: true }).first().click();
    await expect(page.getByText("Create New Quiz")).toBeVisible();

    await page
      .locator('form input[type="text"][required]')
      .first()
      .fill(quizTitle);
    await page
      .locator("form textarea")
      .first()
      .fill("E2E disposable quiz for runtime verification.");

    const createPromise = page.waitForResponse(
      (res) =>
        res.url().includes("/api/v1/quizzes") &&
        res.request().method() === "POST",
      { timeout: 30_000 }
    );

    await page
      .locator("form")
      .filter({ hasText: "Create New Quiz" })
      .locator('button:has-text("Create Quiz")')
      .click();
    const createRes = await createPromise;
    expect(
      createRes.status(),
      `create quiz ${createRes.status()}`
    ).toBeLessThan(400);

    await page.waitForURL(/\/admin\/quiz\/list-quiz\/[^/]+/, {
      timeout: 30_000,
    });
    await expect(page.getByText(quizTitle).first()).toBeVisible({
      timeout: 20_000,
    });

    await page.getByText("Add Question", { exact: true }).first().click();
    await expect(page.getByPlaceholder("Enter question...")).toBeVisible();
    await page.getByPlaceholder("Enter question...").fill("What is 2 + 2?");
    await page.getByPlaceholder("Option A").fill("3");
    await page.getByPlaceholder("Option B").fill("4");
    await page.getByLabel("Correct answer option 2").check();

    const questionPromise = page.waitForResponse(
      (res) =>
        res.url().includes("/questions") && res.request().method() === "POST",
      { timeout: 30_000 }
    );
    await page
      .locator("form")
      .filter({ has: page.getByPlaceholder("Enter question...") })
      .getByRole("button", { name: /^Add Question$/ })
      .click();
    const questionRes = await questionPromise;
    expect(
      questionRes.status(),
      `add question ${questionRes.status()}`
    ).toBeLessThan(400);
    await expect(page.getByText("What is 2 + 2?")).toBeVisible({
      timeout: 20_000,
    });

    await page.goto("/admin/quiz/list-quiz", { waitUntil: "domcontentloaded" });
    await expect(page.getByText(quizTitle).first()).toBeVisible({
      timeout: 20_000,
    });
    await captureShot(page, "created_quiz_in_list.png");

    await page.reload({ waitUntil: "domcontentloaded" });
    await expect(page.getByText(quizTitle).first()).toBeVisible({
      timeout: 20_000,
    });

    const search = page
      .locator('input[type="search"][placeholder*="Search"]')
      .first();
    await search.fill(quizTitle);
    await expect(page.getByText(quizTitle).first()).toBeVisible();

    assertNoCriticalFailures(diagnostics, "create-quiz");
  });

  test("I. Reports page loads without fetch errors", async () => {
    test.setTimeout(60_000);

    await page.goto("/admin/reports", { waitUntil: "domcontentloaded" });

    await expect(page.getByText("Loading...")).toHaveCount(0, {
      timeout: 20_000,
    });
    await expect(page.getByText(/Reports/i).first()).toBeVisible();

    const search = page.locator('input[type="search"]').first();
    await expect(search).toBeVisible();
    await search.fill("e2e");

    const dateControl = page.locator('input[type="datetime-local"]').first();
    await expect(dateControl).toBeVisible();
    await dateControl.fill("2026-01-01T00:00");

    const content = await page.locator("body").innerText();
    assertNoRawTechnicalLeak(content, "reports");
    expect(content).not.toMatch(/Failed to fetch/i);

    assertNoCriticalFailures(diagnostics, "reports");
    await captureShot(page, "reports.png");
  });

  test("J. Logout invalidates session for protected routes", async () => {
    test.setTimeout(60_000);

    await page.goto("/admin", { waitUntil: "domcontentloaded" });
    await expect(page.getByText("Profile Settings")).toBeVisible({
      timeout: 20_000,
    });

    const profileMenu = page
      .locator('button[aria-label="Profile menu"]')
      .first();
    await profileMenu.click();
    await page.getByText("Log Out", { exact: true }).click();

    // App navigates to home after a successful Kratos logout.
    await page.waitForURL(
      (url) => url.pathname === "/" || url.pathname.includes("/account/login"),
      { timeout: 30_000 }
    );

    const cookies = await context.cookies();
    const sessionCookie = cookies.find((c) =>
      c.name.toLowerCase().includes("ory_kratos_session")
    );
    expect(
      !sessionCookie || sessionCookie.value === "",
      "session cookie cleared"
    ).toBeTruthy();

    await page.goto("/admin", { waitUntil: "domcontentloaded" });
    await page.waitForURL(/\/account\/login/, { timeout: 20_000 });

    assertNoCriticalFailures(diagnostics, "logout");
  });
});
