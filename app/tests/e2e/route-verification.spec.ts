import { test, expect, type Page, type Request } from "@playwright/test";

// Helpers to capture and audit traffic for illegal hostnames/protocols
function setupRouteAudit(page: Page) {
  const violations: string[] = [];

  page.on("request", (req: Request) => {
    const urlStr = req.url();
    try {
      const u = new URL(urlStr);
      const host = u.hostname;

      // We block/reject loopback, internal, or Railway public domains from browser requests
      if (host === "localhost" || host === "127.0.0.1" || host === "::1") {
        // Exception: allow local testing if BASE_URL is set to localhost
        const baseUrl =
          process.env.PLAYWRIGHT_BASE_URL || "http://localhost:3000";
        if (!baseUrl.includes("localhost") && !baseUrl.includes("127.0.0.1")) {
          violations.push(`Illegal loopback request: ${urlStr}`);
        }
      }
      if (host.endsWith(".up.railway.app")) {
        violations.push(`Illegal public Railway domain request: ${urlStr}`);
      }
      if (host.endsWith(".railway.internal")) {
        violations.push(
          `Illegal internal Railway domain request in browser: ${urlStr}`
        );
      }
    } catch (e) {
      // Ignored malformed URLs
    }
  });

  page.on("console", (msg) => {
    if (msg.type() === "error") {
      const text = msg.text();
      if (!text.includes("favicon.ico") && !text.includes("Vue Devtools")) {
        violations.push(`Console error: ${text}`);
      }
    }
  });

  return {
    assertNoViolations() {
      expect(violations).toEqual([]);
    },
  };
}

test.describe("Production Route & Redirect Audit", () => {
  const baseUrl = process.env.PLAYWRIGHT_BASE_URL || "http://localhost:3000";

  test("Homepage loads with no localhost or railway.internal leaks", async ({
    page,
  }) => {
    const audit = setupRouteAudit(page);
    await page.goto(baseUrl);
    await expect(page).toHaveTitle(/GK Circle/i);
    audit.assertNoViolations();
  });

  test("Registration page and failure cases render in-place", async ({
    page,
  }) => {
    const audit = setupRouteAudit(page);
    await page.goto(`${baseUrl}/account/register`);
    await expect(page).toHaveURL(/account\/register/);

    // Trigger validation failure with invalid fields (e.g. short password)
    await page.fill("#email", "e2e-invalid-user@gkcircle.com");
    await page.fill("#password", "123");
    await page.fill("#confirmPassword", "123");
    await page.click("button[type='submit']");

    // Should remain on register page and display validation errors
    await expect(page).toHaveURL(/account\/register/);
    audit.assertNoViolations();
  });

  test("Kratos error details display allowlisted fields only", async ({
    page,
  }) => {
    const audit = setupRouteAudit(page);

    // Navigate to error page with a stub Kratos error ID
    await page.goto(`${baseUrl}/error?id=stub:500`);

    // Verify it doesn't show raw JSON or secrets, but maps allowlisted info
    const content = await page.textContent("body");
    expect(content).not.toContain("dsn");
    expect(content).not.toContain("password");
    expect(content).not.toContain("database");

    audit.assertNoViolations();
  });

  test("Settings flow redirects unauthenticated users to login", async ({
    page,
  }) => {
    const audit = setupRouteAudit(page);
    // Settings requires an active session; unauthenticated access must redirect to login
    await page.goto(`${baseUrl}/account/change-password`);
    await page.waitForURL(/account\/login/);
    await expect(page).toHaveURL(/account\/login/);
    audit.assertNoViolations();
  });

  test("Recovery and Verification pages load properly", async ({ page }) => {
    const audit = setupRouteAudit(page);
    await page.goto(`${baseUrl}/recovery`);
    await expect(page).toHaveURL(/recovery/);

    await page.goto(`${baseUrl}/verification`);
    await expect(page).toHaveURL(/verification/);
    audit.assertNoViolations();
  });
});
