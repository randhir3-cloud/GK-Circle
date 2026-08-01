import { test, expect } from "@playwright/test";

test.describe("Production Email Verification & Courier Audit", () => {
  const baseUrl = process.env.PLAYWRIGHT_BASE_URL || "https://gkcircle.com";
  // We expose Mailpit via the gateway proxy at /mailpit/
  const mailpitUrl = `${baseUrl}/mailpit`;
  const jwtSecret =
    process.env.JWT_SECRET ||
    "UiOe_yhM6RwzEp1ciiNlvX6ierlOfOS1KmaHTnYrgChBHivTQPeKAco19uN1kj7t";

  let runId = "";
  let testEmail = "";
  let kratosIdentityId = "";

  test.beforeEach(() => {
    runId = `e2e${Date.now()}`;
    testEmail = `gkcircle.e2e.run.${runId}@example.test`;
  });

  test.afterEach(async ({ request }) => {
    if (kratosIdentityId) {
      console.log(
        `Cleaning up test user: ${testEmail} (Identity: ${kratosIdentityId})`
      );
      const cleanupUrl = `${baseUrl}/api/v1/kratos/e2e-cleanup?kratos_id=${kratosIdentityId}&email=${testEmail}&run_id=${runId}`;
      const response = await request.delete(cleanupUrl, {
        headers: {
          "X-E2E-Secret": jwtSecret,
        },
      });
      console.log(`Cleanup response status: ${response.status()}`);
      try {
        const json = await response.json();
        console.log(`Cleanup response body:`, json);
      } catch (e) {
        console.log(`Cleanup raw response:`, await response.text());
      }
    }
  });

  test("Should complete registration, trigger verification email, extract Mailpit code, and verify successfully", async ({
    page,
  }) => {
    // Intercept Kratos session to capture identity ID securely
    page.on("response", async (res) => {
      if (res.url().includes("/sessions/whoami") && res.status() === 200) {
        try {
          const json = await res.json();
          if (json.identity?.id) {
            kratosIdentityId = json.identity.id;
          }
        } catch (e) {
          // Ignored
        }
      }
    });

    const testStartTime = Date.now();

    // 1. Register a new user
    await page.goto(`${baseUrl}/account/register`);
    await page.fill("#firstname", "E2E");
    await page.fill("#lastname", "Verification");
    await page.fill("#email", testEmail);
    await page.fill("#password", "SuperSecretPassword123!");
    await page.click("button[type='submit']");

    // 2. Wait for redirect or check that user is logged in
    await page.waitForURL(new RegExp(baseUrl));

    // 3. Skip verification first and try to access a protected route (e.g. /instructor/reports)
    await page.goto(`${baseUrl}/instructor/reports`);

    // 4. Verify we are redirected to /verification
    await page.waitForURL(/\/verification/);
    expect(page.url()).toContain("/verification");

    // 5. The page should render the "Send Verification Code" submit button dynamically
    const sendButton = page.locator(
      "button:has-text('Send Verification Code')"
    );
    await expect(sendButton).toBeVisible();
    await sendButton.click();

    interface MailpitMessage {
      ID: string;
      Subject: string;
      To: Array<{ Address: string }>;
      Created: string;
    }

    // 7. Poll Mailpit API to retrieve the enqueued verification email
    let verificationMail: MailpitMessage | null = null;
    for (let attempt = 0; attempt < 15; attempt++) {
      const mailpitRes = await fetch(`${mailpitUrl}/api/v1/messages`).catch(
        () => null
      );
      if (mailpitRes && mailpitRes.ok) {
        const mailData = (await mailpitRes.json()) as {
          messages: MailpitMessage[];
        };
        verificationMail =
          mailData.messages.find((m) => {
            const createdTime = Date.parse(m.Created);
            return (
              m.To?.some((to) => to.Address === testEmail) &&
              createdTime >= testStartTime
            );
          }) || null;
        if (verificationMail) break;
      }
      await page.waitForTimeout(1000);
    }

    expect(verificationMail).not.toBeNull();
    const targetMail = verificationMail;
    if (!targetMail) {
      throw new Error("Verification mail not found");
    }

    // 8. Extract the 6-digit code from the email body
    const detailRes = await fetch(
      `${mailpitUrl}/api/v1/message/${targetMail.ID}`
    );
    expect(detailRes.ok).toBe(true);

    const detailData = (await detailRes.json()) as {
      HTML: string;
      Text: string;
    };
    const bodyText = detailData.HTML || detailData.Text;
    const codeMatch = bodyText.match(/\b\d{6}\b/);
    expect(codeMatch).not.toBeNull();
    const verificationCode = codeMatch ? codeMatch[0] : "";
    console.log(`Extracted verification code: ${verificationCode}`);

    // 9. Input the code and submit verification
    await codeInput.fill(verificationCode);

    const verifyButton = page.locator("button[type='submit']");
    await verifyButton.click();

    // 10. Verification should complete successfully and redirect/refresh
    await page.waitForURL(/\/account\/login/);
    expect(page.url()).toContain("/account/login");
  });
});
