# Phase 4 — Assessments (any-depth bindings)

* **Status**: NOT_STARTED
* **Weight**: 15%
* **Phase owner**: Unassigned
* **Started at**: —
* **Verified at**: —

## Objective

Connect practice sets, topic/chapter/sectional tests, full mocks, PYQ sets, and
live competitions into the course tree by **reusing and extending** the existing
quiz engine — never inventing a parallel assessment runtime. Assessment modes bind
to LearningItems / quiz references at **any CourseNode depth**, with instructions,
attempt limits, resume, autosave, negative marking, leaderboards, analytics, and
supplementary resources (video solutions, question booklets, explanation PDFs).

## Architectural outcome

Architecture freeze (all phases):

```text
Course → CourseNode(parent_id authoritative; children derived) → LearningItem*
```

* Unlimited logical depth (**D-004**). No product `MAX_*_DEPTH`.
* No authoritative `child_node_ids[]`. Children are derived projections only.
* Persisted `node_type` remains `SECTION`|`SUBJECT`|`TOPIC`; semantic labels
  (Module, Chapter, Lecture) are **non-normative**.
* CourseNode lifecycle field is **`status`**; LearningItem delivery field is
  **`publish_state`** — names remain distinct.
* Three graphs remain distinct: structure ≠ sequence (Phase 5) ≠ prerequisites (Phase 5).
* Foreign keys use **`ON DELETE RESTRICT`**, not CASCADE, unless a future ADR
  explicitly changes a specific edge.

Assessment links are LearningItem / CourseNode attachments that launch the
existing quiz, session, answer, scoring, and results infrastructure:

* **Reuse** jovVix/GK Circle quiz engine entities; do not create a second scorer.
* `QUIZ_REFERENCE` item type exists; do not assume a new assessment table without T01.
* Practice, Topic, Chapter, Sectional, Mock, PYQ, and Live are **modes** expressed
  as LearningItem/quiz bindings — not fixed schema levels or hierarchy depths.
* Prerequisite unlocking and learning-sequence navigation stay Phase 5 graphs;
  Phase 4 must not collapse them into the structural tree.

## Current verified baseline

No Phase 4 tasks are verified. Phase 4 may not start migrations or binding work
until **COURSE-P4-T01** (quiz/course binding integration assessment) is accepted
and recorded in `DECISIONS.md` and/or an ADR.

## In scope

* Design gate: quiz/course binding integration assessment (T01).
* Additive schema for assessment links / quiz bindings (after T01).
* Practice, topic, chapter, sectional, full mock, PYQ, and live competition modes.
* Instructions metadata, attempt limits, resume, and autosave integration.
* Negative marking and leaderboard surfacing via existing engine capabilities.
* Assessment analytics and results display in course context.
* Video solution, question booklet, and explanation PDF links (Phase 3 resources).
* Admin attach/configure APIs and learner launch/results APIs.
* UI for outline assessment entry points, attempt flow, and results display.
* Any-depth placement verification and regression against the quiz engine.

## Out of scope

* Native media upload and MinIO resource productization (Phase 3).
* Prerequisite DAG edges and prev/next sequence engines (Phase 5).
* Recursive progress denominators (Phase 6).
* Template generation of whole courses (Phase 7).
* Production hardening soak and analytics productization (Phase 8).
* NestJS/Next/Prisma or a second authentication/test engine.
* Authoritative `child_node_ids[]` or product depth caps.

## Dependencies

* COURSE-P1-T03 — CourseNode persistence.
* COURSE-P2-T02 — LearningItem repository.
* COURSE-P2-T07 — LearningItem `publish_state`.
* COURSE-P2-T08 — Learner LearningItem APIs.
* COURSE-P3-T12 — Learner resource delivery (supplementary PDFs/media).
* Existing quiz, question, session, answer, scoring, and reporting packages in `api/`.
* ADR-023; D-004; D-005.

## Phase boundaries

Phase 4 owns **assessment bindings and quiz-engine integration only**.

| Owned here | Owned elsewhere |
|---|---|
| Quiz/course binding metadata | Resource upload/delivery (Phase 3) |
| Assessment mode configuration | Sequence / unlock graphs (Phase 5) |
| Launch, attempt, resume, results via existing engine | Progress denominators (Phase 6) |
| Admin/learner assessment APIs + UI | Templates / studio (Phase 7) |
| | Course admin auth model (Phase 8 T01 gate) |

Phase 4 must **not** introduce a second scorer, session runtime, or fixed
Subject→Chapter→Topic schema levels.

## Execution sequence

1. **Design gate** — COURSE-P4-T01 must be ACCEPTED before any migration.
2. Persistence — assessment link migration → model → repository tests.
3. Mode bindings — practice/topic/chapter/sectional → mock/PYQ/live.
4. Session integration — instructions, attempts, resume, autosave.
5. Scoring integration — negative marking, leaderboards, analytics.
6. Supplementary links — Phase 3 resource references.
7. HTTP — admin APIs → learner launch/results APIs.
8. UI — outline entry + attempt flow + results.
9. Verification — backend suite → frontend tests → Playwright → docs → full phase verify.

Typical layering per task: Design → Migration → Repository → Service → DTO →
Controller → API tests → Frontend → Playwright → Docs.

## Completion formula

Phase completion equals:

$$\text{Phase Completion} = \frac{\text{Verified Task Points}}{\text{Total Task Points}} \times 100$$

Only tasks marked `VERIFIED` contribute. Declared `Total points` is the ledger
denominator and must equal the arithmetic sum of checklist points.

## Task checklist

Checklist columns: Evidence is column 7 for ledger tooling; Dependencies is last.

| ID | Task | Points | Status | Difficulty | Est. Time | Evidence | Dependencies |
|---|---|---:|---|---|---|---|---|
| COURSE-P4-T01 | Quiz/course binding integration assessment (reuse quiz engine; QUIZ_REFERENCE) | 3 | NOT_STARTED | S | 2h | — | COURSE-P2-T02 |
| COURSE-P4-T02 | Assessment link schema & migration | 5 | NOT_STARTED | S | 2h | — | COURSE-P4-T01 |
| COURSE-P4-T03 | Assessment link Go model (quiz engine reuse) | 8 | NOT_STARTED | M | 4h | — | COURSE-P4-T02 |
| COURSE-P4-T04 | Practice, topic, chapter & sectional bindings | 7 | NOT_STARTED | M | 4h | — | COURSE-P4-T03 |
| COURSE-P4-T05 | Full mock, PYQ & live competition bindings | 8 | NOT_STARTED | M | 4h | — | COURSE-P4-T04 |
| COURSE-P4-T06 | Instructions, attempts, resume & autosave | 8 | NOT_STARTED | L | 6h | — | COURSE-P4-T05 |
| COURSE-P4-T07 | Negative marking & leaderboard integration | 6 | NOT_STARTED | M | 4h | — | COURSE-P4-T06 |
| COURSE-P4-T08 | Assessment analytics & results surfacing | 6 | NOT_STARTED | M | 4h | — | COURSE-P4-T07 |
| COURSE-P4-T09 | Supplementary resource links (video solution, booklet, explanation PDF) | 5 | NOT_STARTED | S | 2h | — | COURSE-P3-T12, COURSE-P4-T05 |
| COURSE-P4-T10 | Admin assessment API endpoints | 8 | NOT_STARTED | M | 4h | — | COURSE-P4-T06, COURSE-P4-T07, COURSE-P4-T08, COURSE-P4-T09 |
| COURSE-P4-T11 | Learner launch & results APIs | 8 | NOT_STARTED | M | 4h | — | COURSE-P4-T10, COURSE-P2-T08 |
| COURSE-P4-T12 | Assessment UI (outline + attempt flow) | 8 | NOT_STARTED | L | 6h | — | COURSE-P4-T11 |
| COURSE-P4-T13 | Backend tests (engine reuse + any-depth) | 8 | NOT_STARTED | M | 4h | — | COURSE-P4-T11 |
| COURSE-P4-T14 | Frontend unit tests | 4 | NOT_STARTED | S | 2h | — | COURSE-P4-T12 |
| COURSE-P4-T15 | Playwright E2E verification | 6 | NOT_STARTED | M | 4h | — | COURSE-P4-T13, COURSE-P4-T14 |
| COURSE-P4-T16 | Phase 4 documentation & ledger sync | 3 | NOT_STARTED | S | 2h | — | COURSE-P4-T15 |
| COURSE-P4-T17 | Full canonical Phase 4 verification | 5 | NOT_STARTED | S | 2h | — | COURSE-P4-T13, COURSE-P4-T14, COURSE-P4-T15, COURSE-P4-T16 |

Total points: 106

## Task-specific acceptance criteria

### COURSE-P4-T01 — Quiz/course binding integration assessment (reuse quiz engine; QUIZ_REFERENCE)
<!-- TASK:COURSE-P4-T01:ACCEPTANCE:START -->
- [ ] Assessment maps existing quiz engine entities (quiz, session, answer, scoring) to Course/LearningItem bindings with repository evidence.
- [ ] Documents whether `QUIZ_REFERENCE` LearningItem metadata suffices vs additive link table; no parallel scorer.
- [ ] Decision recorded in `DECISIONS.md` and/or ADR before T02 begins.
- [ ] Evidence under `docs/features/course-system/evidence/course-p4-t01/`.
- [ ] No migration, HTTP, scoring duplication, or Phase 5+ work is performed.
<!-- TASK:COURSE-P4-T01:ACCEPTANCE:END -->

### COURSE-P4-T02 — Assessment link schema & migration
<!-- TASK:COURSE-P4-T02:ACCEPTANCE:START -->
- [ ] Additive up/down migration implements the T01 decision; FKs reference existing engine tables.
- [ ] No parallel quiz/session tables; schema does not introduce `child_node_ids[]` or depth columns.
- [ ] FK rules use `ON DELETE RESTRICT`; isolated PostgreSQL apply/inspect/rollback/re-apply succeeds.
- [ ] Evidence under `docs/features/course-system/evidence/course-p4-t02/`.
- [ ] No HTTP controllers or Phase 5+ work is performed.
<!-- TASK:COURSE-P4-T02:ACCEPTANCE:END -->

### COURSE-P4-T03 — Assessment link Go model (quiz engine reuse)
<!-- TASK:COURSE-P4-T03:ACCEPTANCE:START -->
- [ ] Repository methods bind LearningItems to quiz/practice-set IDs using existing engine models.
- [ ] Create/Get/List/Update/Delete follow goqu, UUID, and semantic-error conventions.
- [ ] CourseNode `status` and LearningItem `publish_state` remain distinct in all code paths.
- [ ] Focused repository tests plus quiz-engine regression pass.
- [ ] Evidence under `docs/features/course-system/evidence/course-p4-t03/`.
- [ ] No second scoring engine or session runtime is created.
<!-- TASK:COURSE-P4-T03:ACCEPTANCE:END -->

### COURSE-P4-T04 — Practice, topic, chapter & sectional bindings
<!-- TASK:COURSE-P4-T04:ACCEPTANCE:START -->
- [ ] Practice, topic, chapter, and sectional modes bind to LearningItems at any CourseNode depth (D-004).
- [ ] Mode metadata is stored in LearningItem/assessment-link metadata, not as fixed schema levels.
- [ ] Invalid quiz references map to semantic errors before persistence.
- [ ] Focused tests pass; evidence under `docs/features/course-system/evidence/course-p4-t04/`.
- [ ] No prerequisite/unlock logic (Phase 5) is implemented here.
<!-- TASK:COURSE-P4-T04:ACCEPTANCE:END -->

### COURSE-P4-T05 — Full mock, PYQ & live competition bindings
<!-- TASK:COURSE-P4-T05:ACCEPTANCE:START -->
- [ ] Full mock, PYQ, and live competition modes bind with taxonomy metadata (commission, year, stage).
- [ ] PYQ metadata is configurable data, not hardcoded to one state commission.
- [ ] Live bindings reuse existing WebSocket/live quiz infrastructure.
- [ ] Deep-tree placement succeeds without depth rejection.
- [ ] Focused tests pass; evidence under `docs/features/course-system/evidence/course-p4-t05/`.
<!-- TASK:COURSE-P4-T05:ACCEPTANCE:END -->

### COURSE-P4-T06 — Instructions, attempts, resume & autosave
<!-- TASK:COURSE-P4-T06:ACCEPTANCE:START -->
- [ ] Assessment instructions metadata is validated and surfaced on launch.
- [ ] Attempt limits, resume, and autosave integrate with existing session/answer infrastructure.
- [ ] Server-side rules govern attempt eligibility; client cannot bypass limits.
- [ ] Focused tests pass; evidence under `docs/features/course-system/evidence/course-p4-t06/`.
- [ ] No progress aggregation denominators (Phase 6) are required.
<!-- TASK:COURSE-P4-T06:ACCEPTANCE:END -->

### COURSE-P4-T07 — Negative marking & leaderboard integration
<!-- TASK:COURSE-P4-T07:ACCEPTANCE:START -->
- [ ] Negative marking configuration flows through existing scoring engine rules.
- [ ] Leaderboard data surfaces in course context without duplicating ranking logic.
- [ ] Scoring remains server-authoritative; Nuxt client is not source of truth.
- [ ] Focused tests pass; evidence under `docs/features/course-system/evidence/course-p4-t07/`.
- [ ] No custom scoring tables parallel to the quiz engine.
<!-- TASK:COURSE-P4-T07:ACCEPTANCE:END -->

### COURSE-P4-T08 — Assessment analytics & results surfacing
<!-- TASK:COURSE-P4-T08:ACCEPTANCE:START -->
- [ ] Course-context analytics (score, accuracy, time) reuse existing reporting infrastructure.
- [ ] Results are scoped to enrolled/authenticated learners per documented policy.
- [ ] Analytics do not require structural hierarchy changes.
- [ ] Focused tests pass; evidence under `docs/features/course-system/evidence/course-p4-t08/`.
- [ ] No Phase 8 production analytics productization is required.
<!-- TASK:COURSE-P4-T08:ACCEPTANCE:END -->

### COURSE-P4-T09 — Supplementary resource links (video solution, booklet, explanation PDF)
<!-- TASK:COURSE-P4-T09:ACCEPTANCE:START -->
- [ ] Assessment LearningItems may reference Phase 3 resources: video solution, question booklet, explanation PDF.
- [ ] Links are metadata references, not structural hierarchy edges.
- [ ] Learner delivery respects `publish_state` for both assessment and linked resources.
- [ ] Focused tests pass; evidence under `docs/features/course-system/evidence/course-p4-t09/`.
- [ ] No new upload pipeline is introduced beyond Phase 3.
<!-- TASK:COURSE-P4-T09:ACCEPTANCE:END -->

### COURSE-P4-T10 — Admin assessment API endpoints
<!-- TASK:COURSE-P4-T10:ACCEPTANCE:START -->
- [ ] Admin routes attach/configure assessments under authenticated admin course paths.
- [ ] Controllers are transport-only; quiz engine integration remains repository/service-owned.
- [ ] Responses never treat computed `children[]` as writable structural input.
- [ ] Controller tests cover auth, validation, and mode-specific happy paths.
- [ ] Evidence under `docs/features/course-system/evidence/course-p4-t10/`.
<!-- TASK:COURSE-P4-T10:ACCEPTANCE:END -->

### COURSE-P4-T11 — Learner launch & results APIs
<!-- TASK:COURSE-P4-T11:ACCEPTANCE:START -->
- [ ] Authenticated learner endpoints launch assessments only for published LearningItems.
- [ ] Draft discovery is prevented (404 / not-found semantics consistent with COURSE-P2-T08).
- [ ] Results and resume state are server-authoritative.
- [ ] Controller + integration tests pass; evidence under `docs/features/course-system/evidence/course-p4-t11/`.
- [ ] Prerequisite unlock gating (Phase 5) may remain a documented deferral.
<!-- TASK:COURSE-P4-T11:ACCEPTANCE:END -->

### COURSE-P4-T12 — Assessment UI (outline + attempt flow)
<!-- TASK:COURSE-P4-T12:ACCEPTANCE:START -->
- [ ] Admin can attach assessments to nodes at any depth; learner sees outline entry points.
- [ ] Attempt flow covers instructions, answering, autosave/resume, and results display.
- [ ] UI does not invent hierarchy depth limits or persist `child_node_ids[]`.
- [ ] Frontend lint/tests for changed files pass.
- [ ] Evidence under `docs/features/course-system/evidence/course-p4-t12/`.
<!-- TASK:COURSE-P4-T12:ACCEPTANCE:END -->

### COURSE-P4-T13 — Backend tests (engine reuse + any-depth)
<!-- TASK:COURSE-P4-T13:ACCEPTANCE:START -->
- [ ] `go test` covers assessment bindings, attempt/resume, scoring integration, and any-depth placement.
- [ ] Regression on quiz engine, Course, CourseNode, and LearningItem packages passes.
- [ ] Deep nesting does not break assessment launch or results retrieval.
- [ ] Evidence under `docs/features/course-system/evidence/course-p4-t13/`.
- [ ] No production or NUC deployment is performed.
<!-- TASK:COURSE-P4-T13:ACCEPTANCE:END -->

### COURSE-P4-T14 — Frontend unit tests
<!-- TASK:COURSE-P4-T14:ACCEPTANCE:START -->
- [ ] Vitest coverage exists for assessment UI components and helpers.
- [ ] `npm run lint` and `npm test -- --run` pass for changed app scope.
- [ ] Evidence under `docs/features/course-system/evidence/course-p4-t14/`.
- [ ] No dummy success stubs replace real assertions.
<!-- TASK:COURSE-P4-T14:ACCEPTANCE:END -->

### COURSE-P4-T15 — Playwright E2E verification
<!-- TASK:COURSE-P4-T15:ACCEPTANCE:START -->
- [ ] Playwright covers admin attach assessment → learner launch → submit → view results on a nested node.
- [ ] At least one mode (practice or mock) is exercised end-to-end.
- [ ] Screenshots/logs stored; `npx playwright test` exit code recorded.
- [ ] Evidence under `docs/features/course-system/evidence/course-p4-t15/`.
<!-- TASK:COURSE-P4-T15:ACCEPTANCE:END -->

### COURSE-P4-T16 — Phase 4 documentation & ledger sync
<!-- TASK:COURSE-P4-T16:ACCEPTANCE:START -->
- [ ] Phase docs state assessment modes bind at any depth; quiz engine reuse is documented.
- [ ] Architecture freeze invariants are restated; no "leaf node" product language.
- [ ] `npm run course-system:status:sync` then `npm run course-system:status:check` succeed; second sync is a no-op.
- [ ] Evidence under `docs/features/course-system/evidence/course-p4-t16/`.
- [ ] No unverified tasks are marked VERIFIED.
<!-- TASK:COURSE-P4-T16:ACCEPTANCE:END -->

### COURSE-P4-T17 — Full canonical Phase 4 verification
<!-- TASK:COURSE-P4-T17:ACCEPTANCE:START -->
- [ ] All Phase 4 verification commands pass with logged exit codes and timestamps.
- [ ] Phase acceptance criteria checklist is evaluated; failures are recorded explicitly.
- [ ] Evidence under `docs/features/course-system/evidence/course-p4-t17/`.
- [ ] Quiz engine reuse confirmed; no parallel scorer introduced.
<!-- TASK:COURSE-P4-T17:ACCEPTANCE:END -->

## Phase acceptance criteria

- [ ] COURSE-P4-T01 integration assessment is ACCEPTED before persistence work.
- [ ] Assessment-link persistence exists with additive up/down migrations.
- [ ] Practice, topic, chapter, sectional, mock, PYQ, and live modes bind via LearningItem/quiz references.
- [ ] Instructions, attempts, resume, autosave, negative marking, and leaderboards work via engine reuse.
- [ ] Analytics and results surface in course context.
- [ ] Supplementary Phase 3 resource links integrate correctly.
- [ ] Assessments attach at any logical depth (D-004); hierarchy remains parent_id-authoritative (D-005).
- [ ] Admin and learner APIs respect `publish_state` and CourseNode `status`.
- [ ] FK rules use `ON DELETE RESTRICT` unless an ADR documents otherwise.
- [ ] Backend, frontend, Playwright, and ledger sync/check pass.
- [ ] Documentation matches implementation.

## Verification commands

**Task-focused**:

```text
# API repository/model (cwd: api/)
go test ./models/... -count=1 -run <TaskPattern>

# API controller (cwd: api/)
go test ./controllers/api/v1/... -count=1 -run <TaskPattern>
```

**Phase integration**:

```text
# API (cwd: api/)
go test ./...

# Compose verify profile (repo root)
docker compose --profile verify run --rm api-verify

# Frontend (cwd: app/)
npm run lint
npm test -- --run
```

**Full phase verify** (COURSE-P4-T17):

```text
# API (cwd: api/)
go test ./...

# Compose verify profile (repo root)
docker compose --profile verify run --rm api-verify

# Frontend (cwd: app/)
npm run lint
npm test -- --run

# E2E (cwd: app/) — CLI note only
npx playwright test

# Ledger (repo root)
npm run course-system:status:sync
npm run course-system:status:check
```

Record exact timestamps, durations, exit codes, and environments in task evidence.

## Evidence requirements

Each VERIFIED task must provide under
`docs/features/course-system/evidence/course-p4-tXX/`:

* README summarizing scope and non-goals.
* Command logs with exit codes and verification classification.
* Hash / sync artifacts when ledger tooling is run.
* Migration prove-out notes when a migration is in scope.
* Explicit statement of Database Migration: NONE when applicable.
* Confirmation that quiz engine reuse was preserved (no parallel scorer).

## Security requirements

* Attempt limits, resume tokens, and scoring are server-authoritative.
* Draft assessments are not discoverable via learner APIs.
* Admin attach/configure requires authenticated admin authorization.
* Live competition scheduling metadata is validated server-side.
* Client-side answer submission cannot bypass session validation.
* All authorization checks remain server-side.

## Performance requirements

* Assessment launch must not load entire Course trees; scope queries to node/item context.
* Autosave frequency must not overwhelm session write paths; document operational limits.
* Deep-tree bindings (≥25 levels) must succeed without depth rejection.
* Leaderboard queries reuse existing engine indexes; no N+1 full-tree scans.

## Accessibility / mobile

* Attempt flow supports keyboard navigation for question selection where platform patterns allow.
* Instructions and timer displays are readable on mobile viewports.
* Results display uses semantic headings and accessible score summaries.
* Full WCAG audit defers to Phase 8.

## Risks

| Risk | Mitigation |
|---|---|
| Parallel assessment engine created | Mandate quiz engine reuse in T01 and every task boundary |
| Assessment modes hardcoded as schema levels | Modes are LearningItem metadata at any depth |
| Collapsing unlock graphs into structural tree | Phase 5 owns prerequisite DAG separately |
| Client-side scoring or attempt bypass | Server-authoritative session and scoring |
| Assuming new table when QUIZ_REFERENCE suffices | T01 gate documents binding strategy |
| CASCADE delete on assessment FKs | Mandate `ON DELETE RESTRICT` |

## Known limitations

* Prerequisite unlock gating on assessment launch defers to Phase 5.
* Enrollment commerce may remain deferred per Phase 2 pattern.
* Full production analytics productization defers to Phase 8.
* Assessment mode product labels may evolve in data without schema changes.
* Course admin authorization model replacement defers to Phase 8 T01 gate.

## Exit criteria

Phase 4 may be marked complete only when:

1. All checklist tasks are `VERIFIED` with canonical evidence.
2. Declared total points (**106**) match the ledger sum and status tooling.
3. Full phase verification commands pass for changed surfaces.
4. Architecture freeze invariants hold; quiz engine reuse confirmed.
5. COURSE-P4-T01 assessment is ACCEPTED in `DECISIONS.md` and/or ADR.
6. HANDOFF records the next safe action (typically Phase 5 start or approved parallel exception).
