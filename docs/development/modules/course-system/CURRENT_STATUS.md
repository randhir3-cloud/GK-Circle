# Current Implementation Status

Last updated: 2026-07-27T04:05:57+05:30
Updated by: Codex
Branch: chore/ci-verification
Commit: eeac599f05eaf936c7f61db4a3deeac3c9063f59
Working tree: DIRTY

## Current position

* **Active Module**: Course System (v0.1.0)
* **Active Phase**: COURSE-P2 — Learning Items and Information Blocks
* **Phase status**: IN_PROGRESS
* **In-progress task**: None
* **Next task**: None
* **Next milestone**: Phase 2 complete — begin Phase 3
* **Blocked**: No
* **Task-level blocker**: None
* **parallel=true**: Phase 1 and Phase 2 remain concurrently open under D-003; `COURSE-P1-T11` remains available.
* **Docs governance**: Documentation Freeze Approved (`DOCUMENTATION_FREEZE.md`); COURSE-P2-T09, COURSE-P2-T12, COURSE-P2-T14, COURSE-P2-T15, COURSE-P2-T16, and COURSE-P2-T17 VERIFIED

<!-- COURSE_STATUS:GENERATED:START -->
## Progress

- Phase 2 points: 148 / 148
- Active phase completion: 100.00%
- Overall project completion: 19.84%
- In-progress task: None
- Next task: None

Overall: `████░░░░░░░░░░░░░░░░` 19.84%
Phase 2: `████████████████████` 100.00%
<!-- COURSE_STATUS:GENERATED:END -->

## Current task

### COURSE-P2-T17 — Playwright LearningItem E2E

* **Status**: VERIFIED
* **Evidence**: `docs/features/course-system/evidence/course-p2-t17/README.md`
* **Implemented**: Existing Playwright Chromium project covers real deep-node admin DRAFT creation, learner exclusion, publication, API-ordered learner list/detail, responsive views, signed-out denial, and verified cleanup.
* **Verification**: Focused 1/1 and full 11/11 Playwright pass without retries; lint/build/Go/Compose pass; PostgreSQL confirms a four-level chain and zero temporary T17 rows. The inherited 35-failure Vitest baseline is unrelated; all 63 focused Course/LearningItem tests pass.
* **Production source modified**: NO
* **Database Migration**: NONE
* **Breaking Changes**: NO

### COURSE-P2-T16 — Frontend LearningItem tests

* **Status**: VERIFIED
* **Evidence**: `docs/features/course-system/evidence/course-p2-t16/README.md`
* **Implemented**: Direct reusable composition-component and admin transport tests; 9 focused files / 63 tests pass.
* **Verification**: Lint, build, Go, Compose, authenticated deep-node CRUD/learner smoke, PostgreSQL cleanup, screenshots, and production-source audit passed. The unchanged 35-failure legacy frontend baseline remains unrelated; full passing count increased to 105.
* **Production source modified**: NO
* **Database Migration**: NONE
* **Breaking Changes**: NO

### COURSE-P2-T15 — Backend LearningItem test suite

* **Status**: VERIFIED
* **Evidence**: `docs/features/course-system/evidence/course-p2-t15/README.md`
* **Scope**: Backend verification; enrollment regressions passed after T12/D-006.
* **Database Migration**: NONE
* **Breaking Changes**: NO

### COURSE-P2-T26 — LearningItem API examples documentation

* **Status**: VERIFIED
* **Session**: Codex `/root`, started 2026-07-26T22:19:08+05:30
* **Evidence**: `docs/features/course-system/evidence/course-p2-t26/README.md`
* **Implemented**: Canonical snake_case admin CRUD and learner published-read API examples, source mappings, state/depth boundaries, and JSend error examples.
* **Verification**: 22 JSON examples parsed; focused and full local Go tests passed; Docker vet/tests passed; production-code baseline unchanged by T26.
* **Dependencies**: COURSE-P2-T06 VERIFIED; COURSE-P2-T08 VERIFIED.
* **Database Migration**: NONE
* **Breaking Changes**: NO

### COURSE-P2-T14 — Learner item rendering shell UI

* **Status**: VERIFIED
* **Evidence**: `docs/features/course-system/evidence/course-p2-t14/README.md`
* **Implemented**: Authenticated learner list/detail routes, transport-only learner composable, immutable ordered block renderer, safe content URL utility, API-owned previous/next navigation, responsive presentation components, T12 enrollment handling, and 42 focused tests.
* **Verification**: Authenticated Kratos enrollment/deep-node list/detail smoke, persisted PostgreSQL/API/UI correlation, draft exclusion, empty and signed-out states, desktop/mobile screenshots, fixture cleanup, lint/build/Compose, and evidence integrity passed. The unchanged 35-failure legacy frontend baseline is unrelated; all 42 T14 tests pass.
* **Database Migration**: NONE
* **Breaking Changes**: NO

### COURSE-P2-T09 — Admin item composition UI

* **Status**: VERIFIED
* **Evidence**: `docs/features/course-system/evidence/course-p2-t09/README.md`
* **Implemented**: Responsive Course/CourseNode lazy picker, node-local LearningItem CRUD, authorized sidebar entry, transport-only API composable, and focused Vitest coverage.
* **Verification**: Allowlisted admin runtime CRUD + desktop/mobile screenshots passed via local seed (`npm run seed:local-course-admin`). Focused Vitest 12/12 PASS. Full suite remains inherited baseline (35 failed / 94 passed; T09 Course tests pass).
* **Database Migration**: NONE
* **Breaking Changes**: NO

### COURSE-P2-T12 — Enrollment gate for learner item APIs

* **Status**: VERIFIED
* **Evidence**: `docs/features/course-system/evidence/course-p2-t12/README.md`
* **Complete**: Server-owned enrollment persistence and learner API gate verified under D-006.

## Recently completed tasks

* **COURSE-P2-T17** — Playwright LearningItem E2E (VERIFIED; deep-node create → publish → learner view, desktop/mobile, cleanup verified).
* **COURSE-P2-T16** — Frontend LearningItem tests (VERIFIED; 63 focused tests, runtime/persistence evidence, no production source changes).
* **COURSE-P2-T09** — Admin item composition UI (VERIFIED; allowlisted runtime + local course-admin seed).
* **COURSE-P2-T26** — LearningItem API examples documentation (VERIFIED).
* **COURSE-P2-T14** — Learner item rendering shell UI (VERIFIED; authenticated depth-4 runtime and screenshots).
* **COURSE-P2-T25** — Publish-filter controller contract tests (VERIFIED).
* **COURSE-P2-T23** — Node-local learning-chain projection (VERIFIED).
* **COURSE-P2-T22** — Previous/Next learner API (node-local) (VERIFIED).
* **COURSE-P2-T13** — Runtime visibility evaluation (VERIFIED).
* **COURSE-P2-T24** — Draft filtering regression pack (VERIFIED).
* **COURSE-P2-T21** — Previous/Next item resolution (repository, node-local) (VERIFIED).
* **COURSE-P2-T20** — Deep-node LearningItem ownership tests (VERIFIED).
* **COURSE-P2-T11** — Move LearningItem across nodes (same Course) (VERIFIED).
* **COURSE-P2-T10** — Reorder LearningItems within a node (VERIFIED).
* **COURSE-P2-T01** through **COURSE-P2-T08** remain verified.
* **COURSE-P1-T01** through **COURSE-P1-T10** remain verified.

## Next five tasks

1. **COURSE-P2-T18** — Phase 2 documentation update
2. **COURSE-P1-T11** — Admin course list UI (available under D-003)
3. **COURSE-P2-T19** — Full canonical Phase 2 verification
4. **COURSE-P3-T01** — later phase, dependency-gated
5. **COURSE-P1-T12** — dependency-gated Phase 1 work

Blocked independently: none.

## Verification state

| Check | Status | Last run | Result |
|---|---|---|---|
| T17 focused/full Playwright | PASSED | 2026-07-27 | 1/1 focused; 11/11 full; no retries |
| T17 focused frontend tests | PASSED | 2026-07-27 | 9 files, 63 tests |
| T17 full frontend tests | INHERITED_BASELINE | 2026-07-27 | 35 unrelated failures / 105 passed; no LearningItem failures |
| T17 lint / Nuxt build | PASSED | 2026-07-27 | exit 0 |
| T17 Compose web build/config | PASSED | 2026-07-27 | verified `web` image; config valid |
| T17 runtime/persistence/cleanup | PASSED | 2026-07-27 | depth-4 create/publish/learner view; desktop/mobile; final temp rows 0 |
| T16 focused frontend tests | PASSED | 2026-07-27 | 9 files, 63 tests |
| T16 full frontend tests | INHERITED_BASELINE | 2026-07-27 | 35 unrelated failures / 105 passed; all T16 tests passed |
| T16 lint / Nuxt build | PASSED | 2026-07-27 | exit 0 |
| T16 Compose web build/config | PASSED | 2026-07-27 | verified `web` image; config valid |
| T16 authenticated runtime/persistence | PASSED | 2026-07-27 | depth-4 CRUD/read-back/cleanup; learner draft exclusion; desktop/mobile |
| T14 focused frontend tests | PASSED | 2026-07-27 | 4 files, 42 tests |
| T14 full frontend tests | INHERITED_BASELINE | 2026-07-27 | 35 unrelated failures / 96 passed; all T14 tests passed |
| T14 lint / Nuxt build | PASSED | 2026-07-27 | exit 0 |
| T14 Compose web build/config | PASSED | 2026-07-27 | final `web` image built; config valid |
| T14 authenticated learner smoke/screenshots | PASSED | 2026-07-27 | depth-4 enrollment/list/detail; desktop + mobile evidence; cleanup verified |
| T09 focused frontend tests | PASSED | 2026-07-27 | 3 files, 12 tests |
| T09 full frontend tests | INHERITED_BASELINE | 2026-07-27 | 35 failed / 94 passed; T09 Course tests passed |
| T09 lint / Nuxt build | PASSED | 2026-07-26 | exit 0 |
| T09 Compose web build/config | PASSED | 2026-07-26 | `web` image built; config valid |
| T09 authenticated CRUD smoke/screenshots | PASSED | 2026-07-27 | Allowlisted local seed; desktop + mobile evidence |
| Publish-filter controller contract | PASSED | 2026-07-26 | COURSE-P2-T25 evidence |
| LearningItem chain projection | PASSED | 2026-07-26 | COURSE-P2-T23 evidence |
| Runtime visibility projection | PASSED | 2026-07-26 | COURSE-P2-T13 evidence |
| Draft filtering regression | PASSED | 2026-07-26 | COURSE-P2-T24 evidence |
| LearningItem adjacent prev/next | PASSED | 2026-07-26 | COURSE-P2-T21 evidence |
| LearningItem deep-node ownership | PASSED | 2026-07-26 | COURSE-P2-T20 evidence |
| Docker api-verify vet/test | PASSED | 2026-07-26 | exit 0 |
| Ledger integrity | PASSED | 2026-07-27 | Double sync + check (next=COURSE-P2-T16; T09/T12/T14 VERIFIED) |
| Frontend | NOT_APPLICABLE | 2026-07-26 | No frontend in T13 |

## Safety status

* Production deployment performed: No
* Production migration performed: No
* Production data modified: No
* Intel NUC accessed: No
* Breaking changes introduced: No (T09); earlier T22 wrapped prev/next payload remains YES at API level

## Known limitation

`COURSE-P2-T09`, `COURSE-P2-T12`, `COURSE-P2-T14`, `COURSE-P2-T15`, `COURSE-P2-T16`, and `COURSE-P2-T17` are VERIFIED. The inherited full frontend suite remains red with 35 unrelated legacy failures, while all 63 focused Course/LearningItem tests and all 11 Playwright tests pass. Learner prev/next HTTP (T22) and runtime visibility projection (T13) are verified; Phase 1 structural UI remains NOT_STARTED. Premium authorization stays false until a real entitlement source exists.

## Next safe action

COURSE-P2-T18 is dependency-complete after VERIFIED T26. Do not begin it without a separate authorized implementation prompt.
