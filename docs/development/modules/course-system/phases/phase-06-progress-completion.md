# Phase 6 — Recursive Progress and Completion

* **Status**: NOT_STARTED
* **Weight**: 15%
* **Phase owner**: Unassigned
* **Started at**: —
* **Verified at**: —

## Objective

Implement LearningItem completion, CourseNode recursive aggregation with explicit
completion denominators, dashboard and Continue Learning progress integration,
revision tracking, watch/read/assessment progress signals, deep ancestor
recalculation, and consistency after move/deletion/publication changes — without
using "leaf node" as product architecture.

## Architectural outcome

Architecture freeze (all phases):

```text
Course → CourseNode(parent_id authoritative; children derived) → LearningItem*
```

* Unlimited logical depth (**D-004**). No product `MAX_*_DEPTH`.
* No authoritative `child_node_ids[]`. Children are derived projections only.
* Persisted `node_type` remains `SECTION`|`SUBJECT`|`TOPIC`; semantic labels
  are **non-normative**.
* CourseNode lifecycle field is **`status`**; LearningItem delivery field is
  **`publish_state`** — names remain distinct.
* Three graphs remain distinct: structure ≠ sequence (Phase 5) ≠ prerequisites (Phase 5).
* Foreign keys use **`ON DELETE RESTRICT`**, not CASCADE, unless a future ADR
  explicitly changes a specific edge.

Progress is computed from LearningItem completion signals and recursively
aggregated up the structural tree via `parent_id`. Progress does **not** redefine
hierarchy, sequence, or prerequisite graphs. Phase 6 owns **denominator design**
(how many items count toward completion at each node).

## Current verified baseline

No Phase 6 tasks are verified. Phase 6 may not start migrations or aggregation
work until **COURSE-P6-T01** (progress event/aggregate design) is ACCEPTED and
recorded in `DECISIONS.md` and/or an ADR.

## In scope

* Design gate: progress event vs aggregate model (T01).
* Additive migrations for progress events and aggregated completion state (after T01).
* LearningItem completion tracking (watch, read, assessment attempt completion).
* CourseNode recursive aggregation with explicit denominators.
* Deep ancestor recalculation after child progress changes.
* Progress consistency after node move, deletion, and publish_state/status changes.
* Dashboard progress APIs and Continue Learning progress integration.
* Revision and review progress surfacing.
* Admin progress inspection APIs; learner progress and completion UI.

## Out of scope

* Resource upload pipelines (Phase 3).
* New assessment scoring engines (Phase 4).
* Prerequisite DAG definition (Phase 5); Phase 6 consumes unlock/completion inputs.
* Course templates and integrated studio (Phase 7).
* Production hardening, caching soak, and analytics productization (Phase 8).
* Product depth caps or authoritative child-ID arrays.
* "Leaf node" product language — use **LearningItem completion** or **terminal content**.

## Dependencies

* COURSE-P1-T05 — Transactional branch move.
* COURSE-P1-T07 — Transactional subtree deletion.
* COURSE-P2-T02 — LearningItem repository.
* COURSE-P2-T07 — LearningItem `publish_state`.
* COURSE-P2-T08 — Learner LearningItem APIs.
* COURSE-P3-T12 — Learner resource delivery (watch/read signals).
* COURSE-P4-T11 — Learner assessment launch (assessment completion signals).
* COURSE-P5-T10 — Continue Learning resolution (integration point).
* ADR-023; D-004; D-005.

## Phase boundaries

Phase 6 owns **completion signals, denominators, recursive aggregation, and progress UI**.

| Owned here | Owned elsewhere |
|---|---|
| Progress events & aggregates | Resource delivery (Phase 3) |
| Completion denominators | Assessment scoring (Phase 4) |
| Recursive aggregation up parent_id | Sequence/unlock graphs (Phase 5) |
| Dashboard & Continue Learning progress hooks | Templates (Phase 7) |
| Consistency after move/delete/publish | Production caching soak (Phase 8) |

## Execution sequence

1. **Design gate** — COURSE-P6-T01 must be ACCEPTED before any migration.
2. Persistence — progress event/aggregate migration → completion model.
3. Signals — watch/read/assessment progress tracking.
4. Aggregation — recursive CourseNode totals with denominators → ancestor recalculation.
5. Consistency — move/deletion/publication change handlers.
6. HTTP — dashboard APIs → admin inspection → learner progress APIs.
7. Integration — Continue Learning progress (consumes COURSE-P5-T10).
8. UI — learner progress and completion surfaces.
9. Verification — backend suite → frontend tests → Playwright → docs → full phase verify.

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
| COURSE-P6-T01 | Progress event/aggregate design decision | 3 | NOT_STARTED | S | 2h | — | COURSE-P2-T02 |
| COURSE-P6-T02 | Progress event schema & migration | 5 | NOT_STARTED | S | 2h | — | COURSE-P6-T01 |
| COURSE-P6-T03 | LearningItem completion model | 7 | NOT_STARTED | M | 4h | — | COURSE-P6-T02 |
| COURSE-P6-T04 | Watch, read & assessment progress tracking | 7 | NOT_STARTED | M | 4h | — | COURSE-P6-T03, COURSE-P3-T12, COURSE-P4-T11 |
| COURSE-P6-T05 | CourseNode recursive aggregation & denominators | 8 | NOT_STARTED | L | 6h | — | COURSE-P6-T03 |
| COURSE-P6-T06 | Deep ancestor recalculation | 7 | NOT_STARTED | M | 4h | — | COURSE-P6-T05 |
| COURSE-P6-T07 | Progress consistency after move, deletion & publication | 7 | NOT_STARTED | M | 4h | — | COURSE-P6-T06, COURSE-P1-T05, COURSE-P1-T07, COURSE-P2-T07 |
| COURSE-P6-T08 | Dashboard progress APIs | 7 | NOT_STARTED | M | 4h | — | COURSE-P6-T05 |
| COURSE-P6-T09 | Continue Learning progress integration | 6 | NOT_STARTED | M | 4h | — | COURSE-P6-T08, COURSE-P5-T10 |
| COURSE-P6-T10 | Revision & review progress surfacing | 6 | NOT_STARTED | M | 4h | — | COURSE-P6-T04 |
| COURSE-P6-T11 | Admin progress inspection APIs | 6 | NOT_STARTED | M | 4h | — | COURSE-P6-T08 |
| COURSE-P6-T12 | Learner progress & completion UI | 8 | NOT_STARTED | L | 6h | — | COURSE-P6-T08 |
| COURSE-P6-T13 | Backend unit & integration tests | 8 | NOT_STARTED | M | 4h | — | COURSE-P6-T11 |
| COURSE-P6-T14 | Frontend unit tests | 4 | NOT_STARTED | S | 2h | — | COURSE-P6-T12 |
| COURSE-P6-T15 | Playwright E2E verification | 6 | NOT_STARTED | M | 4h | — | COURSE-P6-T13, COURSE-P6-T14 |
| COURSE-P6-T16 | Phase 6 documentation & ledger sync | 3 | NOT_STARTED | S | 2h | — | COURSE-P6-T15 |
| COURSE-P6-T17 | Full canonical Phase 6 verification | 5 | NOT_STARTED | S | 2h | — | COURSE-P6-T13, COURSE-P6-T14, COURSE-P6-T15, COURSE-P6-T16 |

Total points: 103

## Task-specific acceptance criteria

### COURSE-P6-T01 — Progress event/aggregate design decision
<!-- TASK:COURSE-P6-T01:ACCEPTANCE:START -->
- [ ] Decision record defines event-sourced vs aggregate-table (or hybrid) progress model.
- [ ] Denominator rules documented: which LearningItems count toward node/course completion.
- [ ] Model preserves `parent_id` authority, `ON DELETE RESTRICT`, and no "leaf node" product language.
- [ ] Decision recorded in `DECISIONS.md` and/or ADR before T02 migration.
- [ ] Evidence under `docs/features/course-system/evidence/course-p6-t01/`.
- [ ] No migration, HTTP, or UI work is performed.
<!-- TASK:COURSE-P6-T01:ACCEPTANCE:END -->

### COURSE-P6-T02 — Progress event schema & migration
<!-- TASK:COURSE-P6-T02:ACCEPTANCE:START -->
- [ ] Additive up/down migration implements T01 decision; scoped to Course and learner.
- [ ] Schema references LearningItems and CourseNodes without `child_node_ids[]` or depth columns.
- [ ] FK rules use `ON DELETE RESTRICT`; isolated PostgreSQL apply/inspect/rollback/re-apply succeeds.
- [ ] Evidence under `docs/features/course-system/evidence/course-p6-t02/`.
- [ ] No HTTP, aggregation logic, or UI work in this task.
<!-- TASK:COURSE-P6-T02:ACCEPTANCE:END -->

### COURSE-P6-T03 — LearningItem completion model
<!-- TASK:COURSE-P6-T03:ACCEPTANCE:START -->
- [ ] LearningItem completion is tracked per learner with idempotent write semantics.
- [ ] Completion criteria vary by item type but share a common repository interface.
- [ ] Draft/unpublished items are excluded from learner-facing completion totals per denominator rules.
- [ ] Focused repository tests pass; evidence under `docs/features/course-system/evidence/course-p6-t03/`.
- [ ] No recursive aggregation logic in this task.
<!-- TASK:COURSE-P6-T03:ACCEPTANCE:END -->

### COURSE-P6-T04 — Watch, read & assessment progress tracking
<!-- TASK:COURSE-P6-T04:ACCEPTANCE:START -->
- [ ] Watch/read signals integrate with Phase 3 resource consumption where applicable.
- [ ] Assessment completion signals integrate with Phase 4 session completion without duplicate scoring.
- [ ] Partial progress (e.g., video watch percentage) follows T01 denominator rules.
- [ ] Focused tests pass; evidence under `docs/features/course-system/evidence/course-p6-t04/`.
- [ ] No UI work in this task.
<!-- TASK:COURSE-P6-T04:ACCEPTANCE:END -->

### COURSE-P6-T05 — CourseNode recursive aggregation & denominators
<!-- TASK:COURSE-P6-T05:ACCEPTANCE:START -->
- [ ] Aggregation rolls up LearningItem completion via `parent_id` chain to any ancestor depth (D-004).
- [ ] Denominator (total countable items) and numerator (completed items) are explicit and documented.
- [ ] Empty nodes (zero countable items) have defined aggregation behavior (not division-by-zero).
- [ ] Focused tests include multi-level nesting; evidence under `docs/features/course-system/evidence/course-p6-t05/`.
- [ ] No "leaf node" terminology in code comments or API docs.
<!-- TASK:COURSE-P6-T05:ACCEPTANCE:END -->

### COURSE-P6-T06 — Deep ancestor recalculation
<!-- TASK:COURSE-P6-T06:ACCEPTANCE:START -->
- [ ] Completing one LearningItem triggers ancestor aggregate updates up the full chain.
- [ ] Recalculation uses iterative/CTE traversal; no unsafe stack recursion on deep trees.
- [ ] ≥25-level chains recalculate correctly in tests.
- [ ] Evidence under `docs/features/course-system/evidence/course-p6-t06/`.
- [ ] No HTTP or UI work in this task.
<!-- TASK:COURSE-P6-T06:ACCEPTANCE:END -->

### COURSE-P6-T07 — Progress consistency after move, deletion & publication
<!-- TASK:COURSE-P6-T07:ACCEPTANCE:START -->
- [ ] Branch move (COURSE-P1-T05) recalculates affected ancestor aggregates.
- [ ] Subtree deletion (COURSE-P1-T07) removes or archives progress per T01 policy without orphan aggregates.
- [ ] `publish_state` and CourseNode `status` changes update denominators consistently.
- [ ] Focused tests pass; evidence under `docs/features/course-system/evidence/course-p6-t07/`.
- [ ] No CASCADE deletes introduced.
<!-- TASK:COURSE-P6-T07:ACCEPTANCE:END -->

### COURSE-P6-T08 — Dashboard progress APIs
<!-- TASK:COURSE-P6-T08:ACCEPTANCE:START -->
- [ ] Learner dashboard APIs return course-level and node-level progress with denominators.
- [ ] Progress percentages are server-computed; client cannot inflate completion.
- [ ] Cross-Course access returns 404.
- [ ] Controller tests pass; evidence under `docs/features/course-system/evidence/course-p6-t08/`.
<!-- TASK:COURSE-P6-T08:ACCEPTANCE:END -->

### COURSE-P6-T09 — Continue Learning progress integration
<!-- TASK:COURSE-P6-T09:ACCEPTANCE:START -->
- [ ] Continue Learning (COURSE-P5-T10) incorporates incomplete-item progress where applicable.
- [ ] Integration does not redefine sequence or unlock graphs.
- [ ] Focused integration tests pass; evidence under `docs/features/course-system/evidence/course-p6-t09/`.
- [ ] No new navigation graph persistence is introduced.
<!-- TASK:COURSE-P6-T09:ACCEPTANCE:END -->

### COURSE-P6-T10 — Revision & review progress surfacing
<!-- TASK:COURSE-P6-T10:ACCEPTANCE:START -->
- [ ] Revision/review states surface for completed items per documented policy.
- [ ] Re-completion or review does not corrupt denominator totals.
- [ ] Focused tests pass; evidence under `docs/features/course-system/evidence/course-p6-t10/`.
- [ ] No spaced-repetition productization is required.
<!-- TASK:COURSE-P6-T10:ACCEPTANCE:END -->

### COURSE-P6-T11 — Admin progress inspection APIs
<!-- TASK:COURSE-P6-T11:ACCEPTANCE:START -->
- [ ] Admin APIs expose per-learner and aggregate progress for inspection.
- [ ] Admin authorization required; learner privacy rules documented.
- [ ] Controller tests pass; evidence under `docs/features/course-system/evidence/course-p6-t11/`.
<!-- TASK:COURSE-P6-T11:ACCEPTANCE:END -->

### COURSE-P6-T12 — Learner progress & completion UI
<!-- TASK:COURSE-P6-T12:ACCEPTANCE:START -->
- [ ] Learner UI shows node and course progress bars driven by server APIs.
- [ ] Denominator semantics are reflected in UI labels (e.g., "12 of 15 items").
- [ ] UI works at nested depths without depth-cap messaging.
- [ ] Frontend lint/tests pass; evidence under `docs/features/course-system/evidence/course-p6-t12/`.
<!-- TASK:COURSE-P6-T12:ACCEPTANCE:END -->

### COURSE-P6-T13 — Backend unit & integration tests
<!-- TASK:COURSE-P6-T13:ACCEPTANCE:START -->
- [ ] `go test` covers completion, aggregation, recalculation, consistency, and HTTP mapping.
- [ ] Deep-tree and move/delete/publication scenarios included.
- [ ] Regression on Phase 1–5 packages passes.
- [ ] Evidence under `docs/features/course-system/evidence/course-p6-t13/`.
<!-- TASK:COURSE-P6-T13:ACCEPTANCE:END -->

### COURSE-P6-T14 — Frontend unit tests
<!-- TASK:COURSE-P6-T14:ACCEPTANCE:START -->
- [ ] Vitest coverage exists for progress UI components and helpers.
- [ ] `npm run lint` and `npm test -- --run` pass for changed app scope.
- [ ] Evidence under `docs/features/course-system/evidence/course-p6-t14/`.
<!-- TASK:COURSE-P6-T14:ACCEPTANCE:END -->

### COURSE-P6-T15 — Playwright E2E verification
<!-- TASK:COURSE-P6-T15:ACCEPTANCE:START -->
- [ ] Playwright covers complete item → ancestor progress update → dashboard display on nested node.
- [ ] Move item between nodes and verify progress recalculation.
- [ ] Screenshots/logs stored; `npx playwright test` exit code recorded.
- [ ] Evidence under `docs/features/course-system/evidence/course-p6-t15/`.
<!-- TASK:COURSE-P6-T15:ACCEPTANCE:END -->

### COURSE-P6-T16 — Phase 6 documentation & ledger sync
<!-- TASK:COURSE-P6-T16:ACCEPTANCE:START -->
- [ ] Phase docs document denominator rules and recursive aggregation without "leaf node" language.
- [ ] Architecture freeze invariants restated.
- [ ] `npm run course-system:status:sync` then `npm run course-system:status:check` succeed; second sync is a no-op.
- [ ] Evidence under `docs/features/course-system/evidence/course-p6-t16/`.
<!-- TASK:COURSE-P6-T16:ACCEPTANCE:END -->

### COURSE-P6-T17 — Full canonical Phase 6 verification
<!-- TASK:COURSE-P6-T17:ACCEPTANCE:START -->
- [ ] All Phase 6 verification commands pass with logged exit codes and timestamps.
- [ ] Phase acceptance criteria checklist is evaluated; failures recorded explicitly.
- [ ] Evidence under `docs/features/course-system/evidence/course-p6-t17/`.
<!-- TASK:COURSE-P6-T17:ACCEPTANCE:END -->

## Phase acceptance criteria

- [ ] COURSE-P6-T01 design is ACCEPTED before persistence work.
- [ ] Progress events/aggregates exist with additive migrations and explicit denominators.
- [ ] LearningItem completion tracks watch, read, and assessment signals.
- [ ] CourseNode recursive aggregation works at any depth (D-004).
- [ ] Deep ancestor recalculation and move/delete/publication consistency verified.
- [ ] Dashboard and Continue Learning progress integration functional.
- [ ] Admin inspection and learner progress UI functional.
- [ ] No "leaf node" product language in docs or APIs.
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

**Full phase verify** (COURSE-P6-T17):

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
`docs/features/course-system/evidence/course-p6-tXX/`:

* README summarizing scope, denominator rules, and non-goals.
* Command logs with exit codes and verification classification.
* Hash / sync artifacts when ledger tooling is run.
* Migration prove-out notes when a migration is in scope.
* Explicit statement of Database Migration: NONE when applicable.
* Confirmation that hierarchy invariants were not violated.

## Security requirements

* Progress writes are scoped to authenticated learners; cross-learner writes rejected.
* Admin inspection requires admin authorization.
* Client cannot mark items complete without server validation of signals.
* Progress APIs do not leak unpublished content existence beyond documented policy.

## Performance requirements

* Ancestor recalculation on deep trees (≥25 levels) completes within documented bounds.
* Dashboard queries must not scan entire Course trees naively; document query strategy.
* Batch progress reads for outline views should avoid N+1 per-node queries.
* Operational write-rate limits are not product MAX_*_DEPTH.

## Accessibility / mobile

* Progress bars expose accessible labels with numerator/denominator text.
* Dashboard progress readable on mobile without horizontal scroll.
* Completion UI does not rely on color alone for status.

## Risks

| Risk | Mitigation |
|---|---|
| Undefined denominators produce misleading % | T01 gate mandates explicit rules |
| "Leaf node" language reintroduced | Ban in acceptance criteria and review |
| Progress orphans after move/delete | T07 consistency pack |
| Client-side completion inflation | Server-authoritative completion writes |
| Unsafe recursion on deep trees | CTE/iterative ancestor walks |

## Known limitations

* Spaced-repetition and advanced revision algorithms are out of scope.
* Production progress caching defers to Phase 8.
* Enrollment commerce may remain deferred per Phase 2 pattern.
* Real-time progress WebSocket push is not required.

## Exit criteria

Phase 6 may be marked complete only when:

1. All checklist tasks are `VERIFIED` with canonical evidence.
2. Declared total points (**103**) match the ledger sum and status tooling.
3. Full phase verification commands pass for changed surfaces.
4. Denominator rules documented; no "leaf node" product language.
5. COURSE-P6-T01 design ACCEPTED in `DECISIONS.md` and/or ADR.
6. HANDOFF records the next safe action (typically Phase 7 start or approved parallel exception).
