# Phase 7 — Templates and Builder

* **Status**: NOT_STARTED
* **Weight**: 10%
* **Phase owner**: Unassigned
* **Started at**: —
* **Verified at**: —

## Objective

Deliver recursive course templates with unlimited template depth, template kind
metadata (non-normative examples), template instantiation to live Courses, an
integrated studio **reusing COURSE-P1-T09/T10 and COURSE-P2-T06 APIs** (no
parallel live hierarchy), and Cross-Course copy — backend and UI — with
production-grade builder UX.

## Architectural outcome

Architecture freeze (all phases):

```text
Course → CourseNode(parent_id authoritative; children derived) → LearningItem*
```

* Unlimited logical depth (**D-004**). No product `MAX_*_DEPTH`.
* No authoritative `child_node_ids[]`. Children are derived projections only.
* Persisted `node_type` remains `SECTION`|`SUBJECT`|`TOPIC`; semantic labels
  and template kinds are **non-normative**.
* CourseNode lifecycle field is **`status`**; LearningItem delivery field is
  **`publish_state`** — names remain distinct.
* Three graphs remain distinct: structure ≠ sequence (Phase 5) ≠ prerequisites (Phase 5).
* Foreign keys use **`ON DELETE RESTRICT`**, not CASCADE, unless a future ADR
  explicitly changes a specific edge.

Templates are recursive snapshots or blueprints that instantiate into the same
Course → CourseNode → LearningItem model. The integrated studio **reuses**
COURSE-P1-T09/T10 tree APIs and COURSE-P2-T06 LearningItem APIs — it does not
maintain a parallel live hierarchy engine.

## Current verified baseline

No Phase 7 tasks are verified. Phase 7 may not start persistence work until
**COURSE-P7-T01** (template persistence decision) is ACCEPTED and recorded in
`DECISIONS.md` and/or an ADR.

## In scope

* Design gate: template persistence decision (T01).
* Additive migrations for course template persistence (after T01).
* Recursive template model with unlimited template depth.
* Template kind metadata (Foundation, Crash Course, Mock Series, etc. as examples).
* Template instantiation producing live Course + CourseNode + LearningItem trees.
* Integrated studio builder reusing COURSE-P1-T09/T10 and COURSE-P2-T06 APIs.
* Cross-Course copy — dedicated backend service and UI tasks.
* Builder drag-drop reorder, keyboard move, bulk publish/schedule, autosave, preview.
* Large tree performance (virtualization, lazy loading) in builder UI.
* Admin template and builder APIs; backend, frontend, Playwright, ledger evidence.

## Out of scope

* New resource upload pipelines beyond Phase 3 integration in templates.
* New assessment scoring engines beyond Phase 4 bindings in templates.
* Prerequisite DAG productization beyond Phase 5 reuse in instantiated courses.
* Progress aggregation logic beyond Phase 6 hooks on instantiated courses.
* Production hardening soak (Phase 8).
* Parallel live hierarchy engine separate from COURSE-P1-T09/T10 and COURSE-P2-T06 APIs.
* Product depth caps or authoritative child-ID arrays.

## Dependencies

* COURSE-P1-T09 — Admin tree APIs.
* COURSE-P1-T10 — Admin hierarchy mutation APIs.
* COURSE-P1-T13 — Recursive tree-builder UI (patterns).
* COURSE-P2-T06 — Admin learning item API endpoints.
* COURSE-P2-T07 — LearningItem `publish_state`.
* COURSE-P2-T09 — Admin item composition UI (patterns).
* COURSE-P3-T11 — Admin resource API endpoints.
* COURSE-P4-T10 — Admin assessment API endpoints.
* COURSE-P5-T11 — Admin chaining & unlock APIs (optional in instantiated courses).
* COURSE-P6-T08 — Dashboard progress APIs (hooks on instantiated courses).
* ADR-023; D-004; D-005.

## Phase boundaries

Phase 7 owns **templates, instantiation, integrated studio, and Cross-Course copy**.

| Owned here | Owned elsewhere |
|---|---|
| Template persistence & kinds | Live tree mutation semantics (Phase 1) |
| Instantiation to live Course | LearningItem metadata rules (Phase 2) |
| Integrated studio (reuses COURSE-P1-T09/T10, COURSE-P2-T06) | Resource/assessment engines (Phases 3–4) |
| Cross-Course copy backend + UI | Unlock/progress logic (Phases 5–6) |
| Builder UX polish (drag, keyboard, bulk) | Production auth model (Phase 8 T01) |

Phase 7 must **not** create a parallel live hierarchy that bypasses COURSE-P1-T09/T10 or COURSE-P2-T06 APIs.

## Execution sequence

1. **Design gate** — COURSE-P7-T01 must be ACCEPTED before any migration.
2. Persistence — template migration → recursive template model.
3. Instantiation — template → live Course pipeline.
4. Studio — integrated builder reusing P1/P2 APIs.
5. Copy — Cross-Course copy backend → Cross-Course copy UI.
6. UX — drag-drop, keyboard, bulk publish/schedule, autosave, preview, performance.
7. HTTP — admin template & builder APIs.
8. Verification — backend suite → frontend tests → Playwright → docs → full phase verify.

Typical layering per task: Design → Migration → Repository → Service → DTO →
Controller → API tests → Frontend → Playwright → Docs.

## Completion formula

Phase completion equals:

$$\text{Phase Completion} = \frac{\text{Verified Task Points}}{\text{Total Task Points}} \times 100$$

Only `VERIFIED` tasks count. Declared `Total points` must equal the arithmetic
sum of checklist points.

## Task checklist

Checklist columns: Evidence is column 7 for ledger tooling; Dependencies is last.

| ID | Task | Points | Status | Difficulty | Est. Time | Evidence | Dependencies |
|---|---|---:|---|---|---|---|---|
| COURSE-P7-T01 | Template persistence decision | 3 | NOT_STARTED | S | 2h | — | COURSE-P1-T03, COURSE-P2-T02 |
| COURSE-P7-T02 | Course template schema & migration | 5 | NOT_STARTED | S | 2h | — | COURSE-P7-T01 |
| COURSE-P7-T03 | Recursive template model & template kinds | 8 | NOT_STARTED | M | 4h | — | COURSE-P7-T02 |
| COURSE-P7-T04 | Template instantiation to live Course | 9 | NOT_STARTED | L | 6h | — | COURSE-P7-T03 |
| COURSE-P7-T05 | Integrated studio builder (reuse P1/P2 APIs) | 9 | NOT_STARTED | L | 6h | — | COURSE-P1-T09, COURSE-P1-T10, COURSE-P2-T06 |
| COURSE-P7-T06 | Cross-Course copy backend | 7 | NOT_STARTED | M | 4h | — | COURSE-P1-T10, COURSE-P7-T04 |
| COURSE-P7-T07 | Cross-Course copy UI | 6 | NOT_STARTED | M | 4h | — | COURSE-P7-T06 |
| COURSE-P7-T08 | Builder drag-drop, keyboard move & accessibility | 8 | NOT_STARTED | M | 4h | — | COURSE-P7-T05 |
| COURSE-P7-T09 | Bulk publish, schedule, autosave & preview | 8 | NOT_STARTED | M | 4h | — | COURSE-P7-T05, COURSE-P2-T07 |
| COURSE-P7-T10 | Large tree performance in builder | 7 | NOT_STARTED | M | 4h | — | COURSE-P7-T05 |
| COURSE-P7-T11 | Admin template & builder APIs | 9 | NOT_STARTED | M | 4h | — | COURSE-P7-T04, COURSE-P7-T06, COURSE-P7-T09 |
| COURSE-P7-T12 | Backend unit & integration tests | 9 | NOT_STARTED | M | 4h | — | COURSE-P7-T11 |
| COURSE-P7-T13 | Frontend unit tests | 4 | NOT_STARTED | S | 2h | — | COURSE-P7-T05, COURSE-P7-T07 |
| COURSE-P7-T14 | Playwright E2E verification | 6 | NOT_STARTED | M | 4h | — | COURSE-P7-T12, COURSE-P7-T13 |
| COURSE-P7-T15 | Phase 7 documentation & ledger sync | 3 | NOT_STARTED | S | 2h | — | COURSE-P7-T14 |
| COURSE-P7-T16 | Full canonical Phase 7 verification | 5 | NOT_STARTED | S | 2h | — | COURSE-P7-T12, COURSE-P7-T13, COURSE-P7-T14, COURSE-P7-T15 |

Total points: 106

## Task-specific acceptance criteria

### COURSE-P7-T01 — Template persistence decision
<!-- TASK:COURSE-P7-T01:ACCEPTANCE:START -->
- [ ] Decision record compares template storage strategies (separate template tables vs snapshot JSON vs hybrid).
- [ ] Chosen model instantiates into live Course/CourseNode/LearningItem without parallel hierarchy engine.
- [ ] Decision preserves `parent_id` authority, `ON DELETE RESTRICT`, and unlimited depth (D-004).
- [ ] Decision recorded in `DECISIONS.md` and/or ADR before T02 migration.
- [ ] Evidence under `docs/features/course-system/evidence/course-p7-t01/`.
- [ ] No migration, HTTP, or UI work is performed.
<!-- TASK:COURSE-P7-T01:ACCEPTANCE:END -->

### COURSE-P7-T02 — Course template schema & migration
<!-- TASK:COURSE-P7-T02:ACCEPTANCE:START -->
- [ ] Additive up/down migration implements T01 decision.
- [ ] Template nodes use `parent_id`-authoritative recursion; no `child_node_ids[]` or depth column.
- [ ] FK rules use `ON DELETE RESTRICT`; isolated PostgreSQL apply/inspect/rollback/re-apply succeeds.
- [ ] Evidence under `docs/features/course-system/evidence/course-p7-t02/`.
- [ ] No instantiation, HTTP, or UI work in this task.
<!-- TASK:COURSE-P7-T02:ACCEPTANCE:END -->

### COURSE-P7-T03 — Recursive template model & template kinds
<!-- TASK:COURSE-P7-T03:ACCEPTANCE:START -->
- [ ] Template model supports unlimited nesting depth (D-004).
- [ ] Template kinds (Foundation, Crash Course, Mock Series, etc.) are metadata labels, not schema levels.
- [ ] Repository CRUD follows goqu, UUID, and semantic-error conventions.
- [ ] Focused tests pass; evidence under `docs/features/course-system/evidence/course-p7-t03/`.
<!-- TASK:COURSE-P7-T03:ACCEPTANCE:END -->

### COURSE-P7-T04 — Template instantiation to live Course
<!-- TASK:COURSE-P7-T04:ACCEPTANCE:START -->
- [ ] Instantiation creates live Course + CourseNode + LearningItem tree via COURSE-P1 / COURSE-P2 persistence models.
- [ ] Instantiated tree preserves parent_id authority and sibling positions.
- [ ] Resource and assessment references copy or relink per documented policy.
- [ ] Transactional instantiation; partial failure rolls back.
- [ ] Evidence under `docs/features/course-system/evidence/course-p7-t04/`.
- [ ] No parallel hierarchy engine is introduced.
<!-- TASK:COURSE-P7-T04:ACCEPTANCE:END -->

### COURSE-P7-T05 — Integrated studio builder (reuse P1/P2 APIs)
<!-- TASK:COURSE-P7-T05:ACCEPTANCE:START -->
- [ ] Studio calls COURSE-P1-T09/T10 for tree mutations and COURSE-P2-T06 for LearningItem CRUD.
- [ ] No duplicate client-side hierarchy authority or parallel mutation endpoints.
- [ ] Builder works on live Courses and template editing surfaces per T01 model.
- [ ] Frontend lint/tests for changed files pass.
- [ ] Evidence under `docs/features/course-system/evidence/course-p7-t05/`.
<!-- TASK:COURSE-P7-T05:ACCEPTANCE:END -->

### COURSE-P7-T06 — Cross-Course copy backend
<!-- TASK:COURSE-P7-T06:ACCEPTANCE:START -->
- [ ] Cross-Course copy is a validated workflow (D-005 invariant 14); not a structural move.
- [ ] Copy preserves parent_id semantics in the target Course with new UUIDs.
- [ ] Unauthorized cross-Course copy is rejected.
- [ ] Focused tests pass; evidence under `docs/features/course-system/evidence/course-p7-t06/`.
- [ ] No CASCADE deletes introduced.
<!-- TASK:COURSE-P7-T06:ACCEPTANCE:END -->

### COURSE-P7-T07 — Cross-Course copy UI
<!-- TASK:COURSE-P7-T07:ACCEPTANCE:START -->
- [ ] Admin UI triggers Cross-Course copy via COURSE-P7-T06 backend APIs.
- [ ] Copy progress and error states are surfaced to the admin.
- [ ] UI does not perform client-only copy without server confirmation.
- [ ] Frontend lint/tests pass; evidence under `docs/features/course-system/evidence/course-p7-t07/`.
<!-- TASK:COURSE-P7-T07:ACCEPTANCE:END -->

### COURSE-P7-T08 — Builder drag-drop, keyboard move & accessibility
<!-- TASK:COURSE-P7-T08:ACCEPTANCE:START -->
- [ ] Drag-drop reorder/move reconciles to server via Phase 1 mutation APIs.
- [ ] Keyboard move/reorder path exists for non-drag users (mobile/accessibility).
- [ ] DnD does not persist `child_node_ids[]` or local-only hierarchy.
- [ ] Evidence under `docs/features/course-system/evidence/course-p7-t08/`.
<!-- TASK:COURSE-P7-T08:ACCEPTANCE:END -->

### COURSE-P7-T09 — Bulk publish, schedule, autosave & preview
<!-- TASK:COURSE-P7-T09:ACCEPTANCE:START -->
- [ ] Bulk publish/schedule operates on LearningItem `publish_state` and CourseNode `status` distinctly.
- [ ] Autosave debounces mutations to server APIs; conflict handling documented.
- [ ] Preview mode renders learner-safe projections without mutating live publish state.
- [ ] Evidence under `docs/features/course-system/evidence/course-p7-t09/`.
<!-- TASK:COURSE-P7-T09:ACCEPTANCE:END -->

### COURSE-P7-T10 — Large tree performance in builder
<!-- TASK:COURSE-P7-T10:ACCEPTANCE:START -->
- [ ] Builder uses lazy/virtualized rendering for large trees (patterns from COURSE-P1-T14 where applicable).
- [ ] Performance measurements documented with thresholds; no product MAX_DEPTH introduced.
- [ ] Collapsed branches do not fetch children eagerly.
- [ ] Evidence under `docs/features/course-system/evidence/course-p7-t10/`.
<!-- TASK:COURSE-P7-T10:ACCEPTANCE:END -->

### COURSE-P7-T11 — Admin template & builder APIs
<!-- TASK:COURSE-P7-T11:ACCEPTANCE:START -->
- [ ] Admin APIs cover template CRUD, instantiation, and studio operations.
- [ ] Controllers are transport-only; hierarchy rules remain COURSE-P1 / COURSE-P2 service-owned.
- [ ] Controller tests cover auth, validation, and happy paths.
- [ ] Evidence under `docs/features/course-system/evidence/course-p7-t11/`.
<!-- TASK:COURSE-P7-T11:ACCEPTANCE:END -->

### COURSE-P7-T12 — Backend unit & integration tests
<!-- TASK:COURSE-P7-T12:ACCEPTANCE:START -->
- [ ] `go test` covers template CRUD, instantiation, cross-Course copy, and HTTP mapping.
- [ ] Deep template nesting and instantiation regression included.
- [ ] Regression on Phase 1–2 packages passes.
- [ ] Evidence under `docs/features/course-system/evidence/course-p7-t12/`.
<!-- TASK:COURSE-P7-T12:ACCEPTANCE:END -->

### COURSE-P7-T13 — Frontend unit tests
<!-- TASK:COURSE-P7-T13:ACCEPTANCE:START -->
- [ ] Vitest coverage exists for studio, copy UI, and builder helpers.
- [ ] `npm run lint` and `npm test -- --run` pass for changed app scope.
- [ ] Evidence under `docs/features/course-system/evidence/course-p7-t13/`.
<!-- TASK:COURSE-P7-T13:ACCEPTANCE:END -->

### COURSE-P7-T14 — Playwright E2E verification
<!-- TASK:COURSE-P7-T14:ACCEPTANCE:START -->
- [ ] Playwright covers template create → instantiate → edit in studio → cross-Course copy flow.
- [ ] At least one deep-nested template instantiation is exercised.
- [ ] Screenshots/logs stored; `npx playwright test` exit code recorded.
- [ ] Evidence under `docs/features/course-system/evidence/course-p7-t14/`.
<!-- TASK:COURSE-P7-T14:ACCEPTANCE:END -->

### COURSE-P7-T15 — Phase 7 documentation & ledger sync
<!-- TASK:COURSE-P7-T15:ACCEPTANCE:START -->
- [ ] Phase docs state studio reuses P1/P2 APIs; no parallel live hierarchy.
- [ ] Cross-Course copy documented as separate workflow per D-005.
- [ ] `npm run course-system:status:sync` then `npm run course-system:status:check` succeed; second sync is a no-op.
- [ ] Evidence under `docs/features/course-system/evidence/course-p7-t15/`.
<!-- TASK:COURSE-P7-T15:ACCEPTANCE:END -->

### COURSE-P7-T16 — Full canonical Phase 7 verification
<!-- TASK:COURSE-P7-T16:ACCEPTANCE:START -->
- [ ] All Phase 7 verification commands pass with logged exit codes and timestamps.
- [ ] Phase acceptance criteria checklist is evaluated; failures recorded explicitly.
- [ ] Evidence under `docs/features/course-system/evidence/course-p7-t16/`.
<!-- TASK:COURSE-P7-T16:ACCEPTANCE:END -->

## Phase acceptance criteria

- [ ] COURSE-P7-T01 decision is ACCEPTED before persistence work.
- [ ] Template persistence exists with additive migrations and unlimited depth.
- [ ] Template instantiation produces valid live Course/CourseNode/LearningItem trees.
- [ ] Integrated studio reuses COURSE-P1-T09/T10 and COURSE-P2-T06 APIs; no parallel live hierarchy engine.
- [ ] Cross-Course copy backend and UI functional.
- [ ] Builder drag-drop, keyboard move, bulk publish/schedule, autosave, preview work.
- [ ] Large tree performance acceptable with documented measurements.
- [ ] FK rules use `ON DELETE RESTRICT` unless an ADR documents otherwise.
- [ ] Backend, frontend, Playwright, and ledger sync/check pass.

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

**Full phase verify** (COURSE-P7-T16):

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
`docs/features/course-system/evidence/course-p7-tXX/`:

* README summarizing scope and non-goals.
* Command logs with exit codes and verification classification.
* Hash / sync artifacts when ledger tooling is run.
* Migration prove-out notes when a migration is in scope.
* Confirmation that studio reuses P1/P2 APIs (no parallel hierarchy).
* Explicit statement of Database Migration: NONE when applicable.

## Security requirements

* Template and instantiation operations require admin authorization.
* Cross-Course copy requires explicit authorization; target Course scope validated.
* Preview mode must not expose draft content to unauthorized learners.
* Autosave endpoints require same auth as underlying mutation APIs.

## Performance requirements

* Large template trees use lazy/virtualized rendering; document node-count thresholds.
* Instantiation of large templates completes within documented bounds or streams progress.
* Autosave debounce prevents mutation storms; document rate limits.
* No product MAX_*_DEPTH introduced for performance.

## Accessibility / mobile

* Keyboard move/reorder path required alongside drag-drop (COURSE-P7-T08).
* Studio controls have accessible names and focus order.
* Mobile builder supports non-drag move/reorder per Phase 1 patterns.
* Bulk operations confirm destructive actions accessibly.

## Risks

| Risk | Mitigation |
|---|---|
| Parallel live hierarchy engine | T05 mandates P1/P2 API reuse |
| Cross-Course copy treated as move | D-005 invariant 14; separate workflow |
| Template kinds become schema levels | Kinds are metadata only |
| Client-side hierarchy authority in studio | All mutations reconcile to server |
| Instantiation partial failure corrupts live Course | Transactional instantiation in T04 |

## Known limitations

* Live-recording ingest in templates is out of scope.
* Full production caching defers to Phase 8.
* Course admin authorization model replacement defers to Phase 8 T01 gate.
* Template marketplace/sharing beyond Cross-Course copy is out of scope.

## Exit criteria

Phase 7 may be marked complete only when:

1. All checklist tasks are `VERIFIED` with canonical evidence.
2. Declared total points (**106**) match the ledger sum and status tooling.
3. Full phase verification commands pass for changed surfaces.
4. Studio confirmed reusing P1/P2 APIs; Cross-Course copy backend+UI verified.
5. COURSE-P7-T01 decision ACCEPTED in `DECISIONS.md` and/or ADR.
6. HANDOFF records the next safe action (typically Phase 8 start).
