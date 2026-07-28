import {
  ChainStatus,
  ChainLinkResult,
  FlawRecord,
  writeJsonAtomically,
  redactSensitiveText,
} from "./browser-flaw-logger.helper";
import {
  HumanObservationState,
  assertArtifactPathInsideRun,
} from "./browser-observation.helper";

export type FinalDecision =
  | "EXAM READY"
  | "EXAM NOT READY"
  | "TEST RUN INVALID";

export interface RunValidity {
  isValid: boolean;
  invalidReasons: string[];
}

export interface EvidenceState {
  status: "COMPLETE" | "PARTIAL" | "FAILED";
  generatedArtifacts: string[];
  mandatoryArtifactsMissing: string[];
  failedArtifacts: Array<{ artifact: string; error: string }>;
}

export interface AcceptanceState {
  runId: string;
  runValidity: RunValidity;
  humanObservation: HumanObservationState;
  chainLinks: ChainLinkResult[];
  flaws: FlawRecord[];
  scores: Record<string, number>;
  releaseReadinessScore: number;
  evidence: EvidenceState;
  finalDecision?: FinalDecision;
}

export function transitionChainLink(
  link: ChainLinkResult,
  nextStatus: ChainStatus
): ChainLinkResult {
  const allowedTransitions: Readonly<
    Record<ChainStatus, readonly ChainStatus[]>
  > = {
    NOT_STARTED: ["IN_PROGRESS", "BLOCKED", "SKIPPED"],
    IN_PROGRESS: ["PASSED", "FAILED"],
    PASSED: [],
    FAILED: [],
    BLOCKED: [],
    SKIPPED: [],
  };

  if (!allowedTransitions[link.status].includes(nextStatus)) {
    throw new Error(
      `Invalid chain transition for ${link.id}: ${link.status} -> ${nextStatus}`
    );
  }

  return {
    ...link,
    status: nextStatus,
  };
}

export function determineFinalDecision(state: AcceptanceState): FinalDecision {
  if (!state.runValidity.isValid) {
    return "TEST RUN INVALID";
  }

  if (
    state.evidence.status === "FAILED" ||
    state.evidence.mandatoryArtifactsMissing.length > 0
  ) {
    return "TEST RUN INVALID";
  }

  const mandatoryLinkIncomplete = state.chainLinks.some(
    (link) => link.mandatory && link.status !== "PASSED"
  );

  const blockingFlawExists = state.flaws.some(
    (flaw) => flaw.severity === "BLOCKER" || flaw.severity === "CRITICAL"
  );

  if (mandatoryLinkIncomplete || blockingFlawExists) {
    return "EXAM NOT READY";
  }

  return "EXAM READY";
}

export function calculateReadinessScore(state: AcceptanceState): number {
  let earnedBasisPoints = 0;
  const passedLinks = state.chainLinks.filter(
    (l) => l.status === "PASSED"
  ).length;
  const totalLinks = state.chainLinks.length;

  if (totalLinks > 0) {
    earnedBasisPoints = Math.round((passedLinks / totalLinks) * 10000);
  }
  return earnedBasisPoints;
}

export async function writeFinalEvidenceSafely(
  state: AcceptanceState,
  runDir: string
): Promise<void> {
  const targetPath = `${runDir}/acceptance-results.json`;
  assertArtifactPathInsideRun(runDir, targetPath);

  const finalScore = calculateReadinessScore(state);
  state.releaseReadinessScore = finalScore;

  if (!state.humanObservation.ownerObservedRun) {
    state.runValidity.isValid = false;
    state.runValidity.invalidReasons.push(
      "Final operator observation confirmation was not supplied."
    );
  }

  state.finalDecision = determineFinalDecision(state);

  const sanitizedState = {
    ...state,
    runValidity: {
      ...state.runValidity,
      invalidReasons: state.runValidity.invalidReasons.map(redactSensitiveText),
    },
  };

  await writeJsonAtomically(targetPath, sanitizedState);
}
