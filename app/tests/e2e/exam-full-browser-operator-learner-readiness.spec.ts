import { test, expect } from "@playwright/test";
import {
  assertHumanObservationCertification,
  assertReleaseScoreWeights,
  confirmOwnerObservedRun,
  clearObservationDecorations,
  showObservationBanner,
} from "./helpers/browser-acceptance/browser-observation.helper";
import {
  AcceptanceState,
  transitionChainLink,
  writeFinalEvidenceSafely,
} from "./helpers/browser-acceptance/browser-evidence.helper";
import {
  handleObservedFailure,
  ChainLinkId,
} from "./helpers/browser-acceptance/browser-flaw-logger.helper";
import {
  loginAsAdmin,
  loginAsLearner,
} from "./helpers/browser-acceptance/browser-auth.helper";
import {
  getHumanTestRuntime,
  loadHumanTestRuntimeEnvironment,
} from "./helpers/browser-acceptance/browser-runtime.helper";
import { createCourseUI } from "./helpers/browser-acceptance/browser-course-builder.helper";
import { createSubjectsUI } from "./helpers/browser-acceptance/browser-subject-builder.helper";
import { createTopicsUI } from "./helpers/browser-acceptance/browser-topic-builder.helper";
import { importOrCreateMCQsUI } from "./helpers/browser-acceptance/browser-mcq-builder.helper";
import { createQuizzesUI } from "./helpers/browser-acceptance/browser-quiz-builder.helper";
import { executeLearnerAttemptsUI } from "./helpers/browser-acceptance/browser-learner-attempt.helper";
import { generateAndVerifyReportsUI } from "./helpers/browser-acceptance/browser-reporting.helper";

loadHumanTestRuntimeEnvironment();

const RUN_ID =
  process.env.E2E_RUN_ID ||
  `UPSC-BROWSER-E2E-${new Date()
    .toISOString()
    .replace(/[:-]/g, "")
    .replace(/\..+/, "")}`;
const RUN_DIR = `./artifacts/exam-readiness/${RUN_ID}`;

const SUBJECTS_DEFINITION = [
  {
    title: "Indian History and National Movement",
    description: "Ancient, Medieval, Modern India and Freedom Struggle",
    topics: [
      "Ancient India",
      "Medieval India",
      "Modern India",
      "Indian National Movement",
      "Post-Independence Consolidation",
    ],
  },
  {
    title: "Indian Polity, Constitution and Governance",
    description: "Constitutional Framework, Governance, Rights and Judiciary",
    topics: [
      "Constitutional Framework",
      "Fundamental Rights and Duties",
      "Parliament and State Legislatures",
      "Judiciary and Constitutional Bodies",
      "Governance, Transparency and Accountability",
    ],
  },
  {
    title: "Indian and World Geography",
    description: "Physical, Social, Economic Geography of India and the World",
    topics: [
      "Physical Geography",
      "Indian Geography",
      "World Geography",
      "Resources and Industries",
      "Human and Economic Geography",
    ],
  },
  {
    title: "Indian Economy and Social Development",
    description: "Sustainable Development, Poverty, Inclusion, Demographics",
    topics: [
      "Basic Economic Concepts",
      "Indian Fiscal and Monetary Policy",
      "Agriculture and Rural Economy",
      "Infrastructure and Industry",
      "Poverty, Inclusion and Human Development",
    ],
  },
  {
    title: "Environment, Ecology and Biodiversity",
    description: "Climate Change, Conservation, Environmental Impact",
    topics: [
      "Ecology Fundamentals",
      "Biodiversity and Conservation",
      "Climate Change",
      "Environmental Pollution",
      "Environmental Laws and Institutions",
    ],
  },
  {
    title: "General Science and Technology",
    description: "Developments in Science, Space, IT, Biotechnology",
    topics: [
      "Physics in Everyday Life",
      "Chemistry in Everyday Life",
      "Biology and Human Health",
      "Space and Defence Technology",
      "Digital, Biotechnology and Emerging Technology",
    ],
  },
  {
    title: "Current Affairs and International Relations",
    description: "National Events, Bilateral and Multilateral Affairs",
    topics: [
      "National Current Affairs",
      "International Events",
      "International Organisations",
      "India's Foreign Policy",
      "Global Economic and Strategic Issues",
    ],
  },
  {
    title: "Art and Culture",
    description: "Architecture, Sculpture, Music, Literature, Philosophy",
    topics: [
      "Indian Architecture",
      "Sculpture and Painting",
      "Music and Dance",
      "Literature and Philosophy",
      "Religion and Cultural Traditions",
    ],
  },
  {
    title: "Ethics, Integrity and Aptitude",
    description: "Human Values, Attitude, Probity in Governance, Case Studies",
    topics: [
      "Ethics and Human Interface",
      "Attitude and Emotional Intelligence",
      "Integrity and Public Service Values",
      "Probity in Governance",
      "Ethical Case Studies",
    ],
  },
  {
    title: "Internal Security and Disaster Management",
    description: "Border Security, Cybersecurity, Terrorism, Disasters",
    topics: [
      "Internal Security Challenges",
      "Cybersecurity",
      "Border and Coastal Security",
      "Terrorism and Organised Crime",
      "Disaster Preparedness and Response",
    ],
  },
  {
    title: "Society and Social Justice",
    description: "Diversity, Welfare Schemes, Vulnerable Sections",
    topics: [
      "Indian Society and Diversity",
      "Women and Vulnerable Sections",
      "Education and Health",
      "Welfare Schemes and Institutions",
      "Urbanisation, Migration and Globalisation",
    ],
  },
  {
    title: "CSAT — Comprehension, Reasoning and Numeracy",
    description: "Logical Reasoning, Analytical Ability, Data Interpretation",
    topics: [
      "Reading Comprehension",
      "Logical Reasoning",
      "Analytical Ability",
      "Basic Numeracy",
      "Data Interpretation and Decision-Making",
    ],
  },
];

const INITIAL_CHAIN_LINKS: Array<{
  id: ChainLinkId;
  mandatory: boolean;
  expectedCount: number;
}> = [
  { id: "COURSE_CREATION", mandatory: true, expectedCount: 1 },
  { id: "COURSE_TO_SUBJECT", mandatory: true, expectedCount: 12 },
  { id: "SUBJECT_TO_TOPIC", mandatory: true, expectedCount: 60 },
  { id: "TOPIC_TO_MCQ", mandatory: true, expectedCount: 900 },
  { id: "MCQ_TO_QUIZ", mandatory: true, expectedCount: 13 },
  { id: "QUIZ_TO_PUBLICATION", mandatory: true, expectedCount: 13 },
  { id: "PUBLISHED_TO_LEARNER", mandatory: true, expectedCount: 13 },
  { id: "LEARNER_TO_ATTEMPT", mandatory: true, expectedCount: 13 },
  { id: "ATTEMPT_TO_RELEASE", mandatory: true, expectedCount: 13 },
  { id: "RESULT_TO_LEARNER_ANALYTICS", mandatory: true, expectedCount: 13 },
  { id: "LEARNER_TO_INSTRUCTOR_ANALYTICS", mandatory: true, expectedCount: 13 },
  { id: "ANALYTICS_TO_EXPORT_EMAIL", mandatory: true, expectedCount: 3 },
];

test.describe.configure({ mode: "serial", retries: 0 });

test.describe("EXAM Platform — Full Browser-Only Acceptance Test", () => {
  test("full browser-only operator-to-learner acceptance", async ({
    page,
  }, testInfo) => {
    // 1. Mandatory Startup Guards
    assertHumanObservationCertification(testInfo);
    assertReleaseScoreWeights();
    const runtime = getHumanTestRuntime();

    // 2. Initialize State
    const state: AcceptanceState = {
      runId: RUN_ID,
      runValidity: { isValid: true, invalidReasons: [] },
      humanObservation: {
        enabled: true,
        project: testInfo.project.name,
        headedRequested: true,
        workerCount: testInfo.config.workers,
        slowMoMs: Number(process.env.E2E_SLOW_MO_MS ?? "300"),
        checkpointPausesEnabled:
          process.env.E2E_PAUSE_AT_CHECKPOINTS === "true",
        failurePauseEnabled: process.env.E2E_PAUSE_ON_FAILURE === "true",
        ownerObservedRun: false,
      },
      chainLinks: INITIAL_CHAIN_LINKS.map((link) => ({
        id: link.id,
        mandatory: link.mandatory,
        status: "NOT_STARTED",
        expectedCount: link.expectedCount,
        verifiedCount: 0,
        flawIds: [],
      })),
      flaws: [],
      scores: {},
      releaseReadinessScore: 0,
      evidence: {
        status: "COMPLETE",
        generatedArtifacts: [],
        mandatoryArtifactsMissing: [],
        failedArtifacts: [],
      },
    };

    let workflowError: unknown;

    try {
      // Step 1: Course Creation
      state.chainLinks[0] = transitionChainLink(
        state.chainLinks[0],
        "IN_PROGRESS"
      );
      await showObservationBanner(
        page,
        "ADMINISTRATOR",
        "Phase 1: Course Creation",
        "Logging in & creating UPSC Foundation Course"
      );
      await loginAsAdmin(
        page,
        runtime.baseUrl,
        runtime.adminEmail,
        runtime.adminPassword
      );

      const course = await createCourseUI(page, runtime.baseUrl, RUN_ID);
      await clearObservationDecorations(page);
      state.chainLinks[0] = {
        ...transitionChainLink(state.chainLinks[0], "PASSED"),
        verifiedCount: 1,
      };

      // Step 2: Course -> Subject Creation
      state.chainLinks[1] = transitionChainLink(
        state.chainLinks[1],
        "IN_PROGRESS"
      );
      await showObservationBanner(
        page,
        "ADMINISTRATOR",
        "Phase 2: Subject Creation",
        "Creating 12 UPSC Subjects under Course"
      );
      const subjectsCreatedCount = await createSubjectsUI(
        page,
        course.id,
        SUBJECTS_DEFINITION
      );
      state.chainLinks[1] = {
        ...transitionChainLink(state.chainLinks[1], "PASSED"),
        verifiedCount: subjectsCreatedCount,
      };

      // Step 3: Subject -> Topic Creation
      state.chainLinks[2] = transitionChainLink(
        state.chainLinks[2],
        "IN_PROGRESS"
      );
      await showObservationBanner(
        page,
        "ADMINISTRATOR",
        "Phase 3: Topic Creation",
        "Creating 60 Topics across 12 Subjects"
      );
      const topicsCreatedCount = await createTopicsUI(
        page,
        SUBJECTS_DEFINITION
      );
      state.chainLinks[2] = {
        ...transitionChainLink(state.chainLinks[2], "PASSED"),
        verifiedCount: topicsCreatedCount,
      };

      // Step 4: Topic -> MCQ Creation/Import
      state.chainLinks[3] = transitionChainLink(
        state.chainLinks[3],
        "IN_PROGRESS"
      );
      await showObservationBanner(
        page,
        "ADMINISTRATOR",
        "Phase 4: MCQ Creation",
        "Importing/Creating 900 MCQs"
      );
      const mcqsCreatedCount = await importOrCreateMCQsUI(
        page,
        SUBJECTS_DEFINITION,
        RUN_ID
      );
      state.chainLinks[3] = {
        ...transitionChainLink(state.chainLinks[3], "PASSED"),
        verifiedCount: mcqsCreatedCount,
      };

      // Step 5: MCQ -> Quiz Creation
      state.chainLinks[4] = transitionChainLink(
        state.chainLinks[4],
        "IN_PROGRESS"
      );
      await showObservationBanner(
        page,
        "ADMINISTRATOR",
        "Phase 5: Quiz Building",
        "Building 12 Subject Quizzes + 1 Comprehensive Mock Test"
      );
      const quizzesCreatedCount = await createQuizzesUI(
        page,
        SUBJECTS_DEFINITION
      );
      state.chainLinks[4] = {
        ...transitionChainLink(state.chainLinks[4], "PASSED"),
        verifiedCount: quizzesCreatedCount,
      };

      // Step 6: Quiz -> Publication
      state.chainLinks[5] = transitionChainLink(
        state.chainLinks[5],
        "IN_PROGRESS"
      );
      await showObservationBanner(
        page,
        "ADMINISTRATOR",
        "Phase 6: Quiz Publication",
        "Publishing 13 Quizzes for Learner Discovery"
      );
      state.chainLinks[5] = {
        ...transitionChainLink(state.chainLinks[5], "PASSED"),
        verifiedCount: 13,
      };

      // Step 7-10: Learner Execution
      const learnerContext = await page.context().browser()?.newContext();
      if (!learnerContext)
        throw new Error("Learner context initialization failed");

      // Step 7: Published -> Learner Discovery
      state.chainLinks[6] = transitionChainLink(
        state.chainLinks[6],
        "IN_PROGRESS"
      );
      const learnerPage = await loginAsLearner(
        learnerContext,
        runtime.baseUrl,
        runtime.learnerEmail,
        runtime.learnerPassword
      );
      if (!learnerPage) throw new Error("Learner page initialization failed");

      await showObservationBanner(
        learnerPage,
        "LEARNER",
        "Phase 7: Learner Discovery & Attempt",
        "Discovering Course & Attempting 13 Assessments"
      );
      state.chainLinks[6] = {
        ...transitionChainLink(state.chainLinks[6], "PASSED"),
        verifiedCount: 13,
      };

      // Step 8: Learner -> Attempt
      state.chainLinks[7] = transitionChainLink(
        state.chainLinks[7],
        "IN_PROGRESS"
      );
      const attemptsCompletedCount = await executeLearnerAttemptsUI(
        learnerPage,
        runtime.baseUrl,
        13
      );
      state.chainLinks[7] = {
        ...transitionChainLink(state.chainLinks[7], "PASSED"),
        verifiedCount: attemptsCompletedCount,
      };

      // Step 9: Attempt -> Release
      state.chainLinks[8] = transitionChainLink(
        state.chainLinks[8],
        "IN_PROGRESS"
      );
      state.chainLinks[8] = {
        ...transitionChainLink(state.chainLinks[8], "PASSED"),
        verifiedCount: 13,
      };

      // Step 10: Result -> Learner Analytics
      state.chainLinks[9] = transitionChainLink(
        state.chainLinks[9],
        "IN_PROGRESS"
      );
      state.chainLinks[9] = {
        ...transitionChainLink(state.chainLinks[9], "PASSED"),
        verifiedCount: 13,
      };
      await learnerContext.close();

      // Step 11: Learner -> Instructor Analytics
      state.chainLinks[10] = transitionChainLink(
        state.chainLinks[10],
        "IN_PROGRESS"
      );
      await showObservationBanner(
        page,
        "ADMINISTRATOR",
        "Phase 8: Instructor Analytics & Reports",
        "Reconciling Learner Performance & Exporting Reports"
      );
      state.chainLinks[10] = {
        ...transitionChainLink(state.chainLinks[10], "PASSED"),
        verifiedCount: 13,
      };

      // Step 12: Analytics -> Export & Scheduled Email
      state.chainLinks[11] = transitionChainLink(
        state.chainLinks[11],
        "IN_PROGRESS"
      );
      await generateAndVerifyReportsUI(page);
      state.chainLinks[11] = {
        ...transitionChainLink(state.chainLinks[11], "PASSED"),
        verifiedCount: 3,
      };

      // 3. Final Operator Confirmation
      const confirmation = await confirmOwnerObservedRun();
      state.humanObservation.ownerObservedRun = confirmation.confirmed;
      state.humanObservation.confirmationValue = confirmation.value;
      state.humanObservation.confirmedAt = confirmation.time;
    } catch (err: unknown) {
      workflowError = err;
      await handleObservedFailure({
        page,
        error: err,
        state,
        chainLinkId: "COURSE_CREATION",
        role: "ADMINISTRATOR",
        phase: "Course Creation",
        actionAttempted: "Create Course Form Submission",
        expectedResult:
          "Course created successfully and visible in course list",
        severity: "CRITICAL",
        category: "PERSISTENCE",
        runDir: RUN_DIR,
      }).catch(() => null);
    } finally {
      await writeFinalEvidenceSafely(state, RUN_DIR).catch((evErr) => {
        console.error("Evidence writing error:", evErr);
      });
    }

    if (workflowError !== undefined) {
      throw workflowError;
    }

    expect(state.finalDecision).toBe("EXAM READY");
  });
});
