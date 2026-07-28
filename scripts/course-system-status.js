"use strict";

const fs = require("fs");
const path = require("path");
const { execFileSync } = require("child_process");

const ROOT_DIR = path.resolve(__dirname, "..");
const DEV_DIR = path.join(ROOT_DIR, "docs", "development");
const MODULE_DIR = path.join(DEV_DIR, "modules", "course-system");
const PHASES_DIR = path.join(MODULE_DIR, "phases");
const STATUS_JSON_PATH = path.join(MODULE_DIR, "status.json");
const MANIFEST_JSON_PATH = path.join(MODULE_DIR, "manifest.json");
const CURRENT_STATUS_PATH = path.join(MODULE_DIR, "CURRENT_STATUS.md");
const MASTER_PLAN_PATH = path.join(MODULE_DIR, "MASTER_PLAN.md");
const RISKS_PATH = path.join(DEV_DIR, "RISKS.md");
const PROJECT_INDEX_PATH = path.join(DEV_DIR, "PROJECT_INDEX.md");
const AI_CONTEXT_PATH = path.join(DEV_DIR, "AI_CONTEXT.md");
const CANONICAL_EVIDENCE_ROOT =
  "docs/features/course-system/evidence/ledger-initialization";
const GENERATED_START = "<!-- COURSE_STATUS:GENERATED:START -->";
const GENERATED_END = "<!-- COURSE_STATUS:GENERATED:END -->";
const VALID_STATUSES = new Set([
  "NOT_STARTED",
  "IN_PROGRESS",
  "BLOCKED",
  "VERIFIED",
]);

const args = process.argv.slice(2);
if (args.includes("--help") || args.includes("-h")) {
  console.log("GK Circle Course System status utility");
  console.log("Usage:");
  console.log("  node scripts/course-system-status.js --check");
  console.log("  node scripts/course-system-status.js --sync");
  process.exit(0);
}
const checkMode = args.length === 1 && args[0] === "--check";
const syncMode = args.length === 1 && args[0] === "--sync";
if (!checkMode && !syncMode) {
  fail("Specify exactly one mode: --check or --sync.");
}

function fail(message) {
  console.error(`Course System ledger error: ${message}`);
  process.exit(1);
}

function readText(filePath) {
  return fs.readFileSync(filePath, "utf8");
}

function readJson(filePath) {
  return JSON.parse(readText(filePath));
}

function normalizeNewlines(value) {
  return value.replace(/\r\n/g, "\n");
}

function stableJson(value) {
  return `${JSON.stringify(value, null, 2)}\n`;
}

function roundDisplay(value) {
  return Number(value.toFixed(2));
}

function valuesEqual(left, right) {
  return Math.abs(Number(left) - Number(right)) < 1e-10;
}

function getGitValue(gitArgs, fallback) {
  try {
    return execFileSync("git", gitArgs, {
      cwd: ROOT_DIR,
      encoding: "utf8",
      stdio: ["ignore", "pipe", "ignore"],
    }).trim();
  } catch {
    return fallback;
  }
}

function parsePhaseWeights() {
  const weights = new Map();
  for (const line of readText(MASTER_PLAN_PATH).split(/\r?\n/)) {
    const match = line.match(
      /^\|\s*(\d+)\s*\|\s*[^|]+\|\s*(\d+(?:\.\d+)?)%\s*\|/,
    );
    if (match) {
      weights.set(Number(match[1]), Number(match[2]) / 100);
    }
  }
  if (weights.size === 0) {
    throw new Error("MASTER_PLAN.md contains no parseable phase weights.");
  }
  const total = [...weights.values()].reduce((sum, weight) => sum + weight, 0);
  if (!valuesEqual(total, 1)) {
    throw new Error(`Phase weights total ${total * 100}%, expected 100%.`);
  }
  return weights;
}

function parseAcceptanceBlocks(content) {
  const blocks = new Map();
  const pattern =
    /<!-- TASK:([A-Z0-9-]+):ACCEPTANCE:START -->([\s\S]*?)<!-- TASK:\1:ACCEPTANCE:END -->/g;
  for (const match of content.matchAll(pattern)) {
    const criteria = [];
    for (const line of match[2].split(/\r?\n/)) {
      const criterion = line.match(/^\s*-\s*\[([ xX])\]\s+(.+?)\s*$/);
      if (criterion) {
        criteria.push({
          checked: criterion[1].toLowerCase() === "x",
          text: criterion[2],
        });
      }
    }
    blocks.set(match[1], criteria);
  }
  return blocks;
}

function parseLedger() {
  const phaseFiles = fs
    .readdirSync(PHASES_DIR)
    .filter((name) => /^phase-\d+-.+\.md$/.test(name))
    .sort();
  const phases = [];
  const tasks = [];

  for (const fileName of phaseFiles) {
    const phaseNumber = Number(fileName.match(/^phase-0?(\d+)/)[1]);
    const phaseId = `COURSE-P${phaseNumber}`;
    const filePath = path.join(PHASES_DIR, fileName);
    const content = readText(filePath);
    const statusMatch = content.match(/^\*\s+\*\*Status\*\*:\s+([A-Z_]+)/m);
    if (!statusMatch || !VALID_STATUSES.has(statusMatch[1])) {
      throw new Error(`${fileName} has no valid phase status.`);
    }
    const acceptanceBlocks = parseAcceptanceBlocks(content);
    let totalPoints = 0;
    let verifiedPoints = 0;
    const phaseTasks = [];

    for (const line of content.split(/\r?\n/)) {
      const columns = line.split("|").map((part) => part.trim());
      if (columns.length < 9 || !/^COURSE-P\d+-T\d+[A-Z]*$/.test(columns[1])) {
        continue;
      }
      const points = Number(columns[3]);
      const status = columns[4];
      if (!Number.isInteger(points) || points <= 0) {
        throw new Error(`${columns[1]} has invalid points: ${columns[3]}.`);
      }
      if (!VALID_STATUSES.has(status)) {
        throw new Error(`${columns[1]} has invalid status: ${status}.`);
      }
      const task = {
        id: columns[1],
        name: columns[2],
        points,
        status,
        evidence: columns[7],
        phaseId,
        phaseNumber,
        filePath,
        acceptance: acceptanceBlocks.get(columns[1]) || [],
      };
      phaseTasks.push(task);
      tasks.push(task);
      totalPoints += points;
      if (status === "VERIFIED") {
        verifiedPoints += points;
      }
    }

    const declaredTotalMatch = content.match(/^Total points:\s*(\d+)\s*$/m);
    if (!declaredTotalMatch) {
      throw new Error(`${fileName} has no declared total points.`);
    }
    const declaredTotal = Number(declaredTotalMatch[1]);
    if (declaredTotal !== totalPoints) {
      throw new Error(
        `${fileName} declares ${declaredTotal} points but task rows total ${totalPoints}.`,
      );
    }
    phases.push({
      id: phaseId,
      number: phaseNumber,
      status: statusMatch[1],
      fileName,
      tasks: phaseTasks,
      totalPoints,
      verifiedPoints,
      completionPercent:
        totalPoints === 0 ? 0 : (verifiedPoints / totalPoints) * 100,
    });
  }
  return { phases, tasks };
}

function evidencePathExists(task) {
  const candidates = [];
  for (const match of task.evidence.matchAll(/`([^`]+\.(?:md|json|log|txt))`/g)) {
    candidates.push(match[1]);
  }
  if (candidates.length === 0) {
    for (const match of task.evidence.matchAll(
      /(?:^|[\s;])([A-Za-z0-9_./\\-]+\.(?:md|json|log|txt))/g,
    )) {
      candidates.push(match[1]);
    }
  }
  return candidates.some((candidate) => {
    const normalized = candidate.replaceAll("/", path.sep);
    return [
      path.resolve(ROOT_DIR, normalized),
      path.resolve(MODULE_DIR, normalized),
      path.resolve(path.dirname(task.filePath), normalized),
    ].some((candidatePath) => fs.existsSync(candidatePath));
  });
}

function validateTasks(tasks) {
  const inProgress = tasks.filter((task) => task.status === "IN_PROGRESS");
  const parallelAllowed = readText(CURRENT_STATUS_PATH).includes("parallel=true");
  if (inProgress.length > 1 && !parallelAllowed) {
    throw new Error(
      `Multiple tasks are IN_PROGRESS without parallel=true: ${inProgress
        .map((task) => task.id)
        .join(", ")}.`,
    );
  }
  const risks = fs.existsSync(RISKS_PATH) ? readText(RISKS_PATH) : "";
  for (const task of tasks) {
    if (task.status !== "VERIFIED") {
      continue;
    }
    const evidence = task.evidence.replace(/[`*_\s]/g, "");
    if (!evidence || ["-", "—", "none"].includes(evidence.toLowerCase())) {
      throw new Error(`${task.id} is VERIFIED without evidence.`);
    }
    if (!evidencePathExists(task)) {
      throw new Error(`${task.id} evidence does not reference an existing file.`);
    }
    if (task.acceptance.length === 0) {
      throw new Error(`${task.id} is VERIFIED without explicit acceptance criteria.`);
    }
    const unchecked = task.acceptance.filter((criterion) => !criterion.checked);
    if (unchecked.length > 0) {
      throw new Error(
        `${task.id} is VERIFIED with unchecked acceptance criteria: ${unchecked
          .map((criterion) => criterion.text)
          .join("; ")}.`,
      );
    }
    if (
      risks.includes(`blocker:${task.id}`) ||
      risks.includes(`unresolved:${task.id}`)
    ) {
      throw new Error(`${task.id} has an unresolved blocker in RISKS.md.`);
    }
  }
  return inProgress;
}

function deriveState(ledger, weights, storedStatus) {
  const inProgress = validateTasks(ledger.tasks);
  const courseRootVerified =
    ledger.tasks.find((task) => task.id === "COURSE-P1-T02")?.status ===
    "VERIFIED";
  const hierarchyDecisionVerified =
    ledger.tasks.find((task) => task.id === "COURSE-P1-T02B")?.status ===
    "VERIFIED";
  const courseNodeVerified =
    ledger.tasks.find((task) => task.id === "COURSE-P1-T03")?.status ===
    "VERIFIED";
  // Prefer the phase that owns an in-progress task so Phase 2+ can be active
  // while an earlier unfinished phase remains open under parallel work.
  let activePhase = null;
  if (inProgress.length > 0) {
    activePhase = ledger.phases.find(
      (phase) => phase.id === inProgress[0].phaseId,
    );
  }
  if (!activePhase) {
    const inProgressPhases = ledger.phases.filter(
      (phase) => phase.status === "IN_PROGRESS",
    );
    if (inProgressPhases.length > 0) {
      // Prefer the latest open phase so parallel later-phase work remains the
      // active surface after its tasks complete while earlier phases stay open.
      activePhase = inProgressPhases[inProgressPhases.length - 1];
    } else {
      activePhase =
        ledger.phases.find((phase) => phase.status !== "VERIFIED") ||
        ledger.phases[ledger.phases.length - 1];
    }
  }
  const nextTask =
    activePhase.tasks.find((task) => task.status === "NOT_STARTED") || null;
  const phase2Started = ledger.phases.some(
    (phase) =>
      phase.id === "COURSE-P2" &&
      (phase.status === "IN_PROGRESS" ||
        phase.status === "VERIFIED" ||
        phase.verifiedPoints > 0),
  );

  let overallCompletionPercent = 0;
  for (const phase of ledger.phases) {
    const weight = weights.get(phase.number);
    if (weight === undefined) {
      throw new Error(`No phase weight is defined for ${phase.id}.`);
    }
    overallCompletionPercent += phase.completionPercent * weight;
  }

  const semantic = {
    ledgerVersion: 1,
    schemaVersion: 1,
    project: "Course System",
    branch: getGitValue(["branch", "--show-current"], "unknown"),
    commit: getGitValue(["rev-parse", "HEAD"], "UNCOMMITTED"),
    activePhaseId: activePhase.id,
    phaseStatus: activePhase.status,
    inProgressTaskId: inProgress[0]?.id || null,
    nextTaskId: nextTask?.id || null,
    phaseProgress: {
      verifiedPoints: activePhase.verifiedPoints,
      totalPoints: activePhase.totalPoints,
      completionPercent: activePhase.completionPercent,
    },
    overallCompletionPercent,
    display: {
      phaseCompletionPercent: roundDisplay(activePhase.completionPercent),
      overallCompletionPercent: roundDisplay(overallCompletionPercent),
    },
    blocked: activePhase.status === "BLOCKED",
    criticalBlockers: [],
    verifiedTaskCount: ledger.tasks.filter((task) => task.status === "VERIFIED")
      .length,
    totalTaskCount: ledger.tasks.length,
    nextAction:
      nextTask?.id === "COURSE-P1-T02"
        ? "Begin Course model and migration after explicit approval"
        : nextTask?.id === "COURSE-P1-T03"
          ? hierarchyDecisionVerified
            ? "Begin CourseNode model and migration after explicit approval"
            : "Verify and accept ADR-023 before CourseNode implementation"
        : nextTask?.id === "COURSE-P1-T04"
          ? "Begin tree repository queries after separate explicit approval"
        : nextTask?.id === "COURSE-P1-T05"
          ? "Begin transactional branch move after separate explicit approval"
        : nextTask?.id === "COURSE-P1-T06"
          ? "Begin transactional sibling reorder after separate explicit approval"
        : nextTask
          ? `Begin ${nextTask.name}`
          : "No unstarted task in active phase",
    courseModelImplemented: courseRootVerified,
    courseMigrationCreated: courseRootVerified,
    courseNodeImplemented: courseNodeVerified,
    courseNodeMigrationCreated: courseNodeVerified,
    databaseMigrationRun: courseRootVerified,
    phase2Started,
    productionTouched: false,
    breakingChanges: false,
    canonicalEvidenceRoot: CANONICAL_EVIDENCE_ROOT,
  };
  const previousSemantic = storedStatus
    ? Object.fromEntries(
        Object.entries(storedStatus).filter(([key]) => key !== "updatedAt"),
      )
    : null;
  const semanticChanged =
    JSON.stringify(previousSemantic) !== JSON.stringify(semantic);
  return {
    ...semantic,
    updatedAt:
      !semanticChanged && storedStatus?.updatedAt
        ? storedStatus.updatedAt
        : new Date().toISOString(),
  };
}

function progressBar(percent) {
  const filled = Math.round((percent / 100) * 20);
  return `${"█".repeat(filled)}${"░".repeat(20 - filled)}`;
}

function generatedBlock(state) {
  const phaseDisplay = state.display.phaseCompletionPercent;
  const overallDisplay = state.display.overallCompletionPercent;
  const activePhaseNumber = Number(String(state.activePhaseId).replace(/^COURSE-P/, ""));
  return [
    GENERATED_START,
    "## Progress",
    "",
    `- Phase ${activePhaseNumber} points: ${state.phaseProgress.verifiedPoints} / ${state.phaseProgress.totalPoints}`,
    `- Active phase completion: ${phaseDisplay.toFixed(2)}%`,
    `- Overall project completion: ${overallDisplay.toFixed(2)}%`,
    `- In-progress task: ${state.inProgressTaskId === null ? "None" : state.inProgressTaskId}`,
    `- Next task: ${state.nextTaskId === null ? "None" : state.nextTaskId}`,
    "",
    `Overall: \`${progressBar(state.overallCompletionPercent)}\` ${overallDisplay.toFixed(2)}%`,
    `Phase ${activePhaseNumber}: \`${progressBar(state.phaseProgress.completionPercent)}\` ${phaseDisplay.toFixed(2)}%`,
    GENERATED_END,
  ].join("\n");
}

function replaceGeneratedBlock(content, replacement) {
  const normalized = normalizeNewlines(content);
  const start = normalized.indexOf(GENERATED_START);
  const end = normalized.indexOf(GENERATED_END);
  if (start < 0 || end < start) {
    throw new Error("CURRENT_STATUS.md has invalid generated-region markers.");
  }
  const suffixStart = end + GENERATED_END.length;
  return `${normalized.slice(0, start)}${replacement}${normalized.slice(suffixStart)}`;
}

function expectedManifest(state, currentManifest) {
  const activePhaseNumber = Number(String(state.activePhaseId).replace(/^COURSE-P/, ""));
  return {
    ...currentManifest,
    phase: activePhaseNumber,
    completion: state.overallCompletionPercent,
    status: state.phaseStatus,
    inProgressTaskId: state.inProgressTaskId,
    nextTaskId: state.nextTaskId,
  };
}

function escapeRegex(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

function assertStatusSurface(filePath, patterns) {
  const content = readText(filePath);
  for (const [description, pattern] of patterns) {
    if (!pattern.test(content)) {
      throw new Error(`${path.relative(ROOT_DIR, filePath)} has inconsistent ${description}.`);
    }
  }
}

function validateHumanStatus(state) {
  const inProgressText = state.inProgressTaskId || "None";
  const nextTaskText = state.nextTaskId || "None";
  const activePhaseId = state.activePhaseId;
  const activePhaseNumber = Number(String(activePhaseId).replace(/^COURSE-P/, ""));
  if (!Number.isInteger(activePhaseNumber) || activePhaseNumber <= 0) {
    throw new Error(`Invalid activePhaseId: ${activePhaseId}`);
  }
  assertStatusSurface(CURRENT_STATUS_PATH, [
    [
      "active phase",
      new RegExp(
        `\\*\\s+\\*\\*Active Phase\\*\\*:\\s+${escapeRegex(activePhaseId)}\\b`,
      ),
    ],
    ["phase status", /\*\s+\*\*Phase status\*\*:\s+IN_PROGRESS\b/],
    [
      "in-progress task",
      new RegExp(
        `\\*\\s+\\*\\*In-progress task\\*\\*:\\s+${escapeRegex(inProgressText)}(?:\\s|$)`,
      ),
    ],
    [
      "next task",
      new RegExp(
        `\\*\\s+\\*\\*Next task\\*\\*:\\s+${escapeRegex(nextTaskText)}(?:\\s|$)`,
      ),
    ],
  ]);
  assertStatusSurface(PROJECT_INDEX_PATH, [
    [
      "active phase",
      new RegExp(
        `\\*\\s+\\*\\*Active Phase\\*\\*:\\s+${escapeRegex(activePhaseId)}\\b`,
      ),
    ],
    [
      "in-progress task",
      new RegExp(
        `\\*\\s+\\*\\*In-progress Task\\*\\*:\\s+${escapeRegex(inProgressText)}(?:\\s|$)`,
      ),
    ],
    [
      "next task",
      new RegExp(
        `\\*\\s+\\*\\*Next Task\\*\\*:\\s+${escapeRegex(nextTaskText)}(?:\\s|$)`,
      ),
    ],
  ]);
  assertStatusSurface(AI_CONTEXT_PATH, [
    [
      "active phase",
      new RegExp(
        `\\*\\s+\\*\\*Active Phase\\*\\*:\\s+${escapeRegex(activePhaseId)}\\b`,
      ),
    ],
    [
      "in-progress task",
      new RegExp(
        `\\*\\s+\\*\\*In-progress Task\\*\\*:\\s+${escapeRegex(inProgressText)}(?:\\s|$)`,
      ),
    ],
    [
      "next task",
      new RegExp(
        `\\*\\s+\\*\\*Next Task\\*\\*:\\s+${escapeRegex(nextTaskText)}(?:\\s|$)`,
      ),
    ],
  ]);
  assertStatusSurface(MASTER_PLAN_PATH, [
    [
      `Phase ${activePhaseNumber} completion`,
      new RegExp(
        `^\\|\\s*${activePhaseNumber}\\s*\\|[^\\n]*\\|\\s*${state.display.phaseCompletionPercent.toFixed(2)}%\\s*\\|\\s*$`,
        "m",
      ),
    ],
  ]);
  assertStatusSurface(PROJECT_INDEX_PATH, [
    [
      "overall module completion",
      new RegExp(
        `\\*\\s+\\*\\*Overall Module Completion\\*\\*:\\s+${state.display.overallCompletionPercent.toFixed(2)}%`,
      ),
    ],
  ]);
}

function validateCanonicalPointers() {
  const required = [
    [
      path.join(MODULE_DIR, "reports", "repository-assessment.md"),
      `${CANONICAL_EVIDENCE_ROOT}/repository-assessment.md`,
    ],
    [
      path.join(MODULE_DIR, "verification", "latest.md"),
      `${CANONICAL_EVIDENCE_ROOT}/baseline-verification.md`,
    ],
    [
      path.join(MODULE_DIR, "verification", "latest.json"),
      CANONICAL_EVIDENCE_ROOT,
    ],
  ];
  for (const [filePath, canonicalReference] of required) {
    if (!readText(filePath).includes(canonicalReference)) {
      throw new Error(
        `${path.relative(ROOT_DIR, filePath)} does not declare canonical evidence ${canonicalReference}.`,
      );
    }
  }
}

function validateConsistency(state, manifest, currentStatus) {
  const stored = readJson(STATUS_JSON_PATH);
  if (stableJson(stored) !== stableJson(state)) {
    throw new Error("status.json does not match computed ledger state; run --sync.");
  }
  const expectedManifestContent = stableJson(expectedManifest(state, manifest));
  if (stableJson(manifest) !== expectedManifestContent) {
    throw new Error("manifest.json does not match computed ledger state; run --sync.");
  }
  const expectedCurrentStatus = replaceGeneratedBlock(
    currentStatus,
    generatedBlock(state),
  );
  if (normalizeNewlines(currentStatus) !== expectedCurrentStatus) {
    throw new Error(
      "CURRENT_STATUS.md generated region does not match computed ledger state; run --sync.",
    );
  }
  validateHumanStatus(state);
  validateCanonicalPointers();
}

function writeIfChanged(filePath, content) {
  const existing = fs.existsSync(filePath) ? normalizeNewlines(readText(filePath)) : "";
  if (existing === content) {
    return false;
  }
  fs.writeFileSync(filePath, content, "utf8");
  return true;
}

function main() {
  try {
    const ledger = parseLedger();
    const weights = parsePhaseWeights();
    const storedStatus = fs.existsSync(STATUS_JSON_PATH)
      ? readJson(STATUS_JSON_PATH)
      : null;
    const manifest = readJson(MANIFEST_JSON_PATH);
    const currentStatus = readText(CURRENT_STATUS_PATH);
    const state = deriveState(ledger, weights, storedStatus);
    if (checkMode) {
      validateConsistency(state, manifest, currentStatus);
      console.log("Check successful: Course System ledger is consistent and unchanged.");
      return;
    }
    const nextManifest = expectedManifest(state, manifest);
    const nextCurrentStatus = replaceGeneratedBlock(
      currentStatus,
      generatedBlock(state),
    );
    const changed = [
      writeIfChanged(STATUS_JSON_PATH, stableJson(state)),
      writeIfChanged(MANIFEST_JSON_PATH, stableJson(nextManifest)),
      writeIfChanged(CURRENT_STATUS_PATH, nextCurrentStatus),
    ].filter(Boolean).length;
    console.log(
      changed === 0
        ? "Sync successful: no changes required."
        : `Sync successful: updated ${changed} generated file(s).`,
    );
  } catch (error) {
    fail(error.message);
  }
}

main();
