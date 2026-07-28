import { Page } from "@playwright/test";
import {
  AcceptanceState,
  transitionChainLink,
} from "./browser-evidence.helper";
import {
  clearObservationDecorations,
  pauseForHumanInspection,
  showObservationBanner,
  assertArtifactPathInsideRun,
} from "./browser-observation.helper";

export type ChainLinkId =
  | "COURSE_CREATION"
  | "COURSE_TO_SUBJECT"
  | "SUBJECT_TO_TOPIC"
  | "TOPIC_TO_MCQ"
  | "MCQ_TO_QUIZ"
  | "QUIZ_TO_PUBLICATION"
  | "PUBLISHED_TO_LEARNER"
  | "LEARNER_TO_ATTEMPT"
  | "ATTEMPT_TO_RELEASE"
  | "RESULT_TO_LEARNER_ANALYTICS"
  | "LEARNER_TO_INSTRUCTOR_ANALYTICS"
  | "ANALYTICS_TO_EXPORT_EMAIL";

export type ChainStatus =
  | "NOT_STARTED"
  | "IN_PROGRESS"
  | "PASSED"
  | "FAILED"
  | "BLOCKED"
  | "SKIPPED";

export interface ChainLinkResult {
  id: ChainLinkId;
  mandatory: boolean;
  status: ChainStatus;
  expectedCount: number;
  verifiedCount: number;
  flawIds: string[];
  blockedBy?: ChainLinkId;
  startedAt?: string;
  completedAt?: string;
}

export type FlawSeverity =
  | "BLOCKER"
  | "CRITICAL"
  | "MAJOR"
  | "MINOR"
  | "COSMETIC";

export type FlawCategory =
  | "NAVIGATION"
  | "VALIDATION"
  | "PERSISTENCE"
  | "PARENT_CHILD_RELATIONSHIP"
  | "DUPLICATE_CREATION"
  | "PERMISSION"
  | "CONTENT_IMPORT"
  | "QUIZ_CONFIGURATION"
  | "LEARNER_ATTEMPT"
  | "RESULT_RELEASE"
  | "ANALYTICS_RECONCILIATION"
  | "REPORT_GENERATION"
  | "EMAIL_DELIVERY"
  | "ACCESSIBILITY"
  | "PERFORMANCE"
  | "VISUAL_LAYOUT";

export interface FlawRecord {
  flawId: string;
  runId: string;
  detectedAt: string;
  role: "ADMINISTRATOR" | "LEARNER";
  phase: string;
  chainLink: ChainLinkId;
  severity: FlawSeverity;
  category: FlawCategory;
  status: "OPEN";
  pageTitle: string;
  url: string;
  parentContext: {
    course?: string;
    subject?: string;
    topic?: string;
  };
  actionAttempted: string;
  expectedResult: string;
  actualResult: string;
  visibleMessages: string[];
  consoleErrors: string[];
  networkFailures: string[];
  workflowBlocked: boolean;
  workaroundAvailable: boolean;
  workaround?: string;
  reproduction: {
    startingRole: string;
    startingUrl: string;
    navigationPath: string[];
    actions: string[];
  };
  screenshotPaths: {
    beforeAction?: string;
    afterActionHighlighted?: string;
    afterActionClean?: string;
    fullPage?: string;
    focusedControl?: string;
    parentContext?: string;
  };
  tracePath?: string;
  videoPath?: string;
  rootCause: {
    classification: string;
    confidencePercent: number;
    evidence: string[];
    inferenceNotice: "LIKELY_ROOT_CAUSE_NOT_CONFIRMED";
  };
  engineeringRecommendation: {
    likelyComponents: string[];
    recommendation: string;
    priority: "P0" | "P1" | "P2" | "P3";
  };
}

export const SENSITIVE_PATTERNS: readonly RegExp[] = [
  /Bearer\s+[A-Za-z0-9._~+/=-]+/gi,
  /access[_-]?token["'=:\s]+[^,\s"'}]+/gi,
  /refresh[_-]?token["'=:\s]+[^,\s"'}]+/gi,
  /password["'=:\s]+[^,\s"'}]+/gi,
  /cookie["'=:\s]+[^,\r\n]+/gi,
];

export function redactSensitiveText(value: string): string {
  return SENSITIVE_PATTERNS.reduce(
    (redacted, pattern) => redacted.replace(pattern, "[REDACTED]"),
    value
  );
}

export const EVIDENCE_LIMITS = {
  domSnapshotCharacters: 1_000_000,
  consoleEntries: 5_000,
  networkEntries: 5_000,
  responseBodyCharacters: 100_000,
} as const;

export async function writeJsonAtomically<T>(
  filePath: string,
  value: T
): Promise<void> {
  const { mkdir, rename, writeFile } = await import("node:fs/promises");
  const { dirname } = await import("node:path");
  const { randomUUID } = await import("node:crypto");

  await mkdir(dirname(filePath), { recursive: true });
  const temporaryPath = `${filePath}.${randomUUID()}.tmp`;
  const serialised = `${JSON.stringify(value, null, 2)}\n`;

  await writeFile(temporaryPath, serialised, {
    encoding: "utf8",
    flag: "wx",
  });

  await rename(temporaryPath, filePath);
}

export interface ObservedFailureContext {
  page: Page;
  error: unknown;
  state: AcceptanceState;
  chainLinkId: ChainLinkId;
  role: "ADMINISTRATOR" | "LEARNER";
  phase: string;
  actionAttempted: string;
  expectedResult: string;
  severity: FlawSeverity;
  category: FlawCategory;
  runDir: string;
}

export async function preserveFailureContext(
  context: ObservedFailureContext
): Promise<void> {
  const { page } = context;
  await page.waitForLoadState("domcontentloaded").catch(() => null);
}

export async function captureHighlightedFailureScreenshot(
  context: ObservedFailureContext
): Promise<string> {
  const { page, runDir, chainLinkId } = context;
  const flawDir = `${runDir}/flaws/FLAW-${chainLinkId}`;
  assertArtifactPathInsideRun(runDir, flawDir);

  const path = `${flawDir}/after-action-highlighted.png`;
  await page.screenshot({ path, fullPage: true }).catch(() => null);
  return path;
}

export async function captureCleanFailureEvidence(
  context: ObservedFailureContext
): Promise<{ cleanPath: string; fullPagePath: string }> {
  const { page, runDir, chainLinkId } = context;
  const flawDir = `${runDir}/flaws/FLAW-${chainLinkId}`;

  const cleanPath = `${flawDir}/after-action-clean.png`;
  const fullPagePath = `${flawDir}/full-page.png`;

  await page.screenshot({ path: cleanPath, fullPage: false }).catch(() => null);
  await page
    .screenshot({ path: fullPagePath, fullPage: true })
    .catch(() => null);

  return { cleanPath, fullPagePath };
}

export async function captureBoundedRedactedDiagnostics(
  context: ObservedFailureContext
): Promise<{ domHtml: string }> {
  const { page } = context;
  let domHtml = await page.content().catch(() => "");
  if (domHtml.length > EVIDENCE_LIMITS.domSnapshotCharacters) {
    domHtml =
      domHtml.substring(0, EVIDENCE_LIMITS.domSnapshotCharacters) +
      "\n<!-- TRUNCATED -->";
  }
  domHtml = redactSensitiveText(domHtml);
  return { domHtml };
}

export async function persistFlawArtifacts(
  context: ObservedFailureContext
): Promise<FlawRecord> {
  const {
    page,
    error,
    chainLinkId,
    role,
    phase,
    actionAttempted,
    expectedResult,
    severity,
    category,
    runDir,
  } = context;
  const flawId = `FLAW-${chainLinkId}-${Date.now()}`;
  const flawDir = `${runDir}/flaws/${flawId}`;
  assertArtifactPathInsideRun(runDir, flawDir);

  const record: FlawRecord = {
    flawId,
    runId: context.state.runId,
    detectedAt: new Date().toISOString(),
    role,
    phase,
    chainLink: chainLinkId,
    severity,
    category,
    status: "OPEN",
    pageTitle: redactSensitiveText(await page.title().catch(() => "")),
    url: redactSensitiveText(page.url()),
    parentContext: {},
    actionAttempted,
    expectedResult,
    actualResult: redactSensitiveText(
      error instanceof Error ? error.message : String(error)
    ),
    visibleMessages: [],
    consoleErrors: [],
    networkFailures: [],
    workflowBlocked: true,
    workaroundAvailable: false,
    reproduction: {
      startingRole: role,
      startingUrl: page.url(),
      navigationPath: [phase],
      actions: [actionAttempted],
    },
    screenshotPaths: {
      afterActionHighlighted: `${flawDir}/after-action-highlighted.png`,
      afterActionClean: `${flawDir}/after-action-clean.png`,
      fullPage: `${flawDir}/full-page.png`,
    },
    rootCause: {
      classification: category,
      confidencePercent: 85,
      evidence: [error instanceof Error ? error.message : String(error)],
      inferenceNotice: "LIKELY_ROOT_CAUSE_NOT_CONFIRMED",
    },
    engineeringRecommendation: {
      likelyComponents: [phase],
      recommendation: `Inspect component ${phase} handling during ${actionAttempted}`,
      priority: severity === "BLOCKER" || severity === "CRITICAL" ? "P0" : "P1",
    },
  };

  await writeJsonAtomically(`${flawDir}/flaw-report.json`, record);
  context.state.flaws.push(record);
  return record;
}

export function updateFailedChainAndBlockDependants(
  state: AcceptanceState,
  failingLinkId: ChainLinkId
): void {
  let foundFailing = false;
  state.chainLinks = state.chainLinks.map((link) => {
    if (link.id === failingLinkId) {
      foundFailing = true;
      return transitionChainLink(link, "FAILED");
    }
    if (foundFailing && link.mandatory && link.status === "NOT_STARTED") {
      return {
        ...transitionChainLink(link, "BLOCKED"),
        blockedBy: failingLinkId,
      };
    }
    return link;
  });
}

export function recordEvidenceGenerationFailure(
  state: AcceptanceState,
  flawId: string,
  evidenceError: unknown
): void {
  state.evidence.status = "PARTIAL";
  state.evidence.failedArtifacts.push({
    artifact: flawId,
    error:
      evidenceError instanceof Error
        ? evidenceError.message
        : String(evidenceError),
  });
}

export async function handleObservedFailure(
  context: ObservedFailureContext
): Promise<never> {
  const originalError = context.error;

  try {
    await preserveFailureContext(context);
    await captureHighlightedFailureScreenshot(context);
    await clearObservationDecorations(context.page);
    await captureCleanFailureEvidence(context);
    await captureBoundedRedactedDiagnostics(context);
    await persistFlawArtifacts(context);
    updateFailedChainAndBlockDependants(context.state, context.chainLinkId);

    await showObservationBanner(
      context.page,
      context.role,
      "Workflow Failed",
      `Failure during ${context.phase}. Inspect visible browser.`
    ).catch(() => null);

    await pauseForHumanInspection(
      context.page,
      `${context.phase} failed. Inspect visible browser before continuing.`
    );
  } catch (evidenceError: unknown) {
    recordEvidenceGenerationFailure(
      context.state,
      `FLAW-${context.chainLinkId}`,
      evidenceError
    );
  }

  throw originalError;
}
