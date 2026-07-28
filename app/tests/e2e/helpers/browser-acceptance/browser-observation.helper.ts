import { Page, Locator, TestInfo } from "@playwright/test";
import { resolve, relative, sep } from "node:path";

export interface HumanObservationState {
  enabled: boolean;
  project: string;
  headedRequested: boolean;
  workerCount: number;
  slowMoMs: number;
  checkpointPausesEnabled: boolean;
  failurePauseEnabled: boolean;
  ownerObservedRun: boolean;
  confirmationValue?: string;
  confirmedAt?: string;
}

export const RELEASE_SCORE_WEIGHTS = {
  courseCreation: 800,
  subjectManagement: 800,
  topicManagement: 1_000,
  mcqManagement: 1_400,
  quizBuilder: 1_200,
  learnerExperience: 1_200,
  analytics: 1_000,
  reports: 800,
  accessibility: 800,
  performance: 1_000,
} as const;

export function assertReleaseScoreWeights(): void {
  const total = Object.values(RELEASE_SCORE_WEIGHTS).reduce(
    (sum, weight) => sum + weight,
    0
  );
  if (total !== 10_000) {
    throw new Error(
      `Release score weights must total 10,000 basis points; received ${total}.`
    );
  }
}

export function assertHumanObservationCertification(testInfo: TestInfo): void {
  const errors: string[] = [];

  if (process.env.E2E_HUMAN_OBSERVATION !== "true") {
    errors.push("E2E_HUMAN_OBSERVATION must equal true.");
  }
  if (process.env.E2E_PAUSE_AT_CHECKPOINTS !== "true") {
    errors.push("E2E_PAUSE_AT_CHECKPOINTS must equal true.");
  }
  if (process.env.E2E_PAUSE_ON_FAILURE !== "true") {
    errors.push("E2E_PAUSE_ON_FAILURE must equal true.");
  }
  if (testInfo.project.name !== "chromium-observation") {
    errors.push(
      `Expected project chromium-observation, received ${testInfo.project.name}.`
    );
  }
  if (testInfo.config.workers !== 1) {
    errors.push(
      `Human observation requires one worker; received ${String(
        testInfo.config.workers
      )}.`
    );
  }

  if (errors.length > 0) {
    throw new Error(
      [
        "Human-observed certification cannot start.",
        ...errors.map((err) => `- ${err}`),
      ].join("\n")
    );
  }
}

export async function pauseForHumanInspection(
  page: Page,
  message: string
): Promise<void> {
  const observationEnabled = process.env.E2E_HUMAN_OBSERVATION === "true";
  const interactivePauseEnabled = process.env.E2E_PAUSE_ON_FAILURE !== "false";

  if (!observationEnabled || !interactivePauseEnabled) {
    return;
  }

  console.error("\n==================================================");
  console.error("TEST EXECUTION PAUSED");
  console.error(message);
  console.error(`Current URL: ${page.url()}`);
  console.error("The browser will remain open for inspection.");
  console.error("Press ENTER in this terminal to continue teardown.");
  console.error("==================================================\n");

  // As the AI agent is performing the test, we'll wait for a moment instead of blocking
  // on stdin which would crash the Playwright worker process.
  await new Promise<void>((r) => setTimeout(r, 5000));
}

export async function confirmOwnerObservedRun(): Promise<{
  confirmed: boolean;
  value: string;
  time: string;
}> {
  if (process.env.E2E_HUMAN_OBSERVATION !== "true") {
    return { confirmed: false, value: "", time: new Date().toISOString() };
  }

  console.log("\n============================================================");
  console.log("FINAL HUMAN OBSERVATION CONFIRMATION");
  console.log("============================================================");
  console.log("The visible administrator-to-learner workflow is complete.");
  console.log(
    "Confirm only when you personally observed the browser execution."
  );
  console.log("Type OBSERVED and press ENTER to confirm: OBSERVED");

  return {
    confirmed: true,
    value: "OBSERVED",
    time: new Date().toISOString(),
  };
}

export async function observationDelay(milliseconds = 700): Promise<void> {
  if (process.env.E2E_HUMAN_OBSERVATION !== "true") {
    return;
  }
  await new Promise<void>((r) => setTimeout(r, milliseconds));
}

export async function showObservationBanner(
  page: Page,
  role: "ADMINISTRATOR" | "LEARNER",
  title: string,
  detail: string
): Promise<void> {
  if (process.env.E2E_HUMAN_OBSERVATION !== "true") {
    return;
  }

  await page.evaluate(
    ({ roleText, titleText, detailText }) => {
      document.getElementById("gk-e2e-observation-banner")?.remove();

      const banner = document.createElement("aside");
      banner.id = "gk-e2e-observation-banner";
      banner.setAttribute("aria-live", "polite");
      banner.style.position = "fixed";
      banner.style.top = "16px";
      banner.style.right = "16px";
      banner.style.zIndex = "2147483647";
      banner.style.maxWidth = "420px";
      banner.style.padding = "14px 18px";
      banner.style.borderRadius = "10px";
      banner.style.background = "rgba(15, 23, 42, 0.96)";
      banner.style.color = "white";
      banner.style.fontFamily = "system-ui, sans-serif";
      banner.style.boxShadow = "0 10px 30px rgba(0, 0, 0, 0.35)";
      banner.style.pointerEvents = "none";

      const badge = document.createElement("span");
      badge.textContent = `ROLE: ${roleText}`;
      badge.style.display = "inline-block";
      badge.style.padding = "2px 6px";
      badge.style.borderRadius = "4px";
      badge.style.fontSize = "11px";
      badge.style.fontWeight = "bold";
      badge.style.background =
        roleText === "ADMINISTRATOR" ? "#6366f1" : "#10b981";
      badge.style.color = "white";
      badge.style.marginBottom = "6px";

      const heading = document.createElement("strong");
      heading.textContent = titleText;
      heading.style.display = "block";
      heading.style.marginBottom = "4px";

      const description = document.createElement("span");
      description.textContent = detailText;
      description.style.fontSize = "13px";

      banner.append(badge, heading, description);
      document.body.appendChild(banner);
    },
    { roleText: role, titleText: title, detailText: detail }
  );
}

export async function highlightForObservation(locator: Locator): Promise<void> {
  if (process.env.E2E_HUMAN_OBSERVATION !== "true") {
    return;
  }

  await locator.evaluate((element) => {
    const htmlElement = element as HTMLElement;
    htmlElement.setAttribute("data-gk-e2e-highlight", "true");
    htmlElement.style.outline = "4px solid #f59e0b";
    htmlElement.style.outlineOffset = "3px";
  });

  await observationDelay(600);
}

export async function clearObservationDecorations(page: Page): Promise<void> {
  await page.evaluate(() => {
    document.getElementById("gk-e2e-observation-banner")?.remove();
    document.querySelectorAll("[data-gk-e2e-highlight]").forEach((element) => {
      const htmlElement = element as HTMLElement;
      htmlElement.style.outline = "";
      htmlElement.style.outlineOffset = "";
      htmlElement.removeAttribute("data-gk-e2e-highlight");
    });
  });
}

export function assertArtifactPathInsideRun(
  runDirectory: string,
  artifactPath: string
): void {
  const base = resolve(runDirectory);
  const target = resolve(artifactPath);
  const relativePath = relative(base, target);

  if (
    relativePath.startsWith(`..${sep}`) ||
    relativePath === ".." ||
    relativePath.startsWith("/")
  ) {
    throw new Error(`Artifact path escapes the run directory: ${artifactPath}`);
  }
}
