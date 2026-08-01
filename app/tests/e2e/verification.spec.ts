import { execFile } from "node:child_process";
import { randomUUID } from "node:crypto";
import path from "node:path";
import { promisify } from "node:util";
import { expect, test } from "@playwright/test";

const execFileAsync = promisify(execFile);

test.use({ trace: "off", video: "off", screenshot: "off" });

test.describe("Email verification and courier integration", () => {
  const baseUrl = process.env.PLAYWRIGHT_BASE_URL || "http://localhost:5000";
  const mailpitUrl = process.env.E2E_MAILPIT_URL || "http://localhost:8025";
  const kratosUrl =
    process.env.E2E_KRATOS_PUBLIC_URL || "http://localhost:4433";
  let runId = "";
  let testEmail = "";
  let testPassword = "";
  let kratosIdentityId = "";
  let expectedNodeOrder: Array<string> = [];

  test.beforeEach(({ page }) => {
    runId = `e2e${randomUUID().replaceAll("-", "").slice(0, 16)}`;
    testEmail = `gkc.verification+${runId}@example.test`;
    testPassword = `Gkc!${randomUUID()}aA9`;
    kratosIdentityId = "";
    expectedNodeOrder = [];

    // Preserve the pre-existing finite timeout hardening.
    page.setDefaultTimeout(30_000);
    page.setDefaultNavigationTimeout(30_000);

    page.on("response", async (response) => {
      const url = response.url();
      if (
        url.includes("/self-service/verification/flows") &&
        response.status() === 200
      ) {
        try {
          const flow = await response.json();
          expectedNodeOrder = (flow?.ui?.nodes || []).map(
            (node: {
              type?: string;
              group?: string;
              attributes?: { name?: string; type?: string };
            }) =>
              `${node.group || ""}:${node.type || ""}:${
                node.attributes?.type || ""
              }:${node.attributes?.name || ""}`
          );
        } catch {
          // A later successful flow response can still provide the ordering evidence.
        }
      }
      if (url.includes("/sessions/whoami") && response.status() === 200) {
        try {
          const session = await response.json();
          if (session?.identity?.id) kratosIdentityId = session.identity.id;
        } catch {
          // Cleanup remains guarded and will not run without a verified identity ID.
        }
      }
    });
  });

  test.afterEach(async () => {
    if (!kratosIdentityId) return;
    const repositoryRoot = path.resolve(process.cwd(), "..");
    await execFileAsync(
      "docker",
      [
        "compose",
        "-f",
        "docker-compose.yaml",
        "exec",
        "-T",
        "api",
        "./gk-circle",
        "qa",
        "cleanup-verification",
        "--identity-id",
        kratosIdentityId,
        "--run-id",
        runId,
        "--confirm",
      ],
      { cwd: repositoryRoot, windowsHide: true }
    );
  });

  test("registers, sends a real courier message, verifies, refreshes the session, and unlocks a protected route", async ({
    page,
    request,
  }) => {
    const testStartTime = Date.now();

    // Preserve the pre-existing domcontentloaded navigation hardening.
    await page.goto(`${baseUrl}/account/register`, {
      waitUntil: "domcontentloaded",
    });
    await page.fill("#firstname", "E2E");
    await page.fill("#lastname", "Verification");
    await page.fill("#email", testEmail);
    await page.fill("#password", testPassword);
    await page.click("button[type='submit']");

    await page.waitForURL(
      (url) =>
        url.origin === new URL(baseUrl).origin &&
        ["/", "/account/login", "/verification"].some(
          (pathname) => url.pathname === pathname
        ),
      { timeout: 30_000 }
    );

    const session = await page.evaluate(async (publicUrl) => {
      const response = await fetch(`${publicUrl}/sessions/whoami`, {
        credentials: "include",
        headers: { Accept: "application/json" },
      });
      return response.ok ? response.json() : null;
    }, kratosUrl);
    kratosIdentityId = session?.identity?.id || kratosIdentityId;
    expect(kratosIdentityId).not.toBe("");

    await page.goto(`${baseUrl}/instructor/reports`, {
      waitUntil: "domcontentloaded",
    });
    await page.waitForURL(/\/verification/);
    expect(new URL(page.url()).searchParams.get("return_to")).toBe(
      "/instructor/reports"
    );

    if (!page.url().includes("flow=")) {
      await page
        .getByRole("button", { name: "Send Verification Code" })
        .click();
      await page.waitForURL(/flow=/);
    }

    const sendCode = page.locator('button[name="method"][value="code"]');
    await expect(sendCode).toBeVisible();

    const renderedNodeOrder = await page
      .locator("[data-kratos-node-index]")
      .evaluateAll((elements) =>
        elements.map(
          (element) =>
            `${element.getAttribute("data-kratos-node-group") || ""}:` +
            `${element.getAttribute("data-kratos-node-type") || ""}:` +
            `${element.getAttribute("data-kratos-node-input-type") || ""}:` +
            `${element.getAttribute("data-kratos-node-name") || ""}`
        )
      );
    expect(renderedNodeOrder).toEqual(expectedNodeOrder);

    await sendCode.click();

    interface MailpitMessage {
      ID: string;
      To: Array<{ Address: string }>;
      Created: string;
    }

    let verificationMail: MailpitMessage | undefined;
    for (let attempt = 0; attempt < 20 && !verificationMail; attempt += 1) {
      const response = await request.get(`${mailpitUrl}/api/v1/messages`);
      if (response.ok()) {
        const data = (await response.json()) as { messages?: MailpitMessage[] };
        verificationMail = data.messages?.find(
          (message) =>
            Date.parse(message.Created) >= testStartTime &&
            message.To?.some((recipient) => recipient.Address === testEmail)
        );
      }
      if (!verificationMail) await page.waitForTimeout(500);
    }
    expect(verificationMail).toBeDefined();

    const detail = await request.get(
      `${mailpitUrl}/api/v1/message/${verificationMail?.ID}`
    );
    expect(detail.ok()).toBe(true);
    const body = (await detail.json()) as { HTML?: string; Text?: string };
    const match = (body.HTML || body.Text || "").match(/\b\d{6}\b/);
    expect(match).not.toBeNull();

    const codeInput = page.locator('input[name="code"]');
    await expect(codeInput).toBeVisible();
    await codeInput.fill(match?.[0] || "");
    await page.locator('button[name="method"][value="code"]').click();

    await page.waitForURL(/\/instructor\/reports/);
    expect(page.url()).toContain("/instructor/reports");
  });
});
