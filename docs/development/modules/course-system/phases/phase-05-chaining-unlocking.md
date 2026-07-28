# Phase 5 — Navigation, Unlocking, and Learning Sequence

* **Status**: NOT_STARTED
* **Weight**: 10%
* **Phase owner**: Unassigned
* **Started at**: —
* **Verified at**: —

## Objective

Implement course-wide learning sequence, unlock-aware Previous/Next and Continue
Learning, prerequisite DAG evaluation, role/enrollment/manual unlock — while
keeping the **three related graphs** distinct: structural tree, learning sequence,
and prerequisite DAG. Consume Phase 1 breadcrumb/ancestor APIs; do not redefine them.

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
* Foreign keys use **`ON DELETE RESTRICT`**, not CASCADE, unless a future ADR
  explicitly changes a specific edge.

| Graph | Edge meaning | Must not be collapsed into |
|---|---|---|
| Structural CourseNode tree | `parent_id` hierarchy | Sequence order or unlock edges |
| Learning sequence | Deterministic course-wide order / prev-next / continue-learning | Structural parentage |
| Prerequisite dependency DAG | Unlocking edges; DAG cycle checks | Structural tree cycle rules alone |

Normative constraints:

* `parent_id` is the sole authoritative structural relationship; children are derived.
* Sequence and unlock evaluation must work at any depth.
* Structural move/cycle rules remain ADR-023; unlock DAG cycles are a separate checker.
* Phase 5 **consumes** COURSE-P1-T16/T28/T29 breadcrumb/ancestor APIs — does not redefine them.

## Current verified baseline

No Phase 5 tasks are verified. Phase 5 may not start persistence work until
**COURSE-P5-T01** (sequence representation decision) and **COURSE-P5-T02**
(prerequisite persistence decision) are both ACCEPTED and recorded in
`DECISIONS.md` and/or ADRs.

## In scope

* Design gates: sequence representation (T01) and prerequisite persistence (T02).
* Prerequisite / dependency persistence (additive migrations, after T02).
* Course-wide learning-sequence resolution (preorder, prev/next, continue-learning).
* DAG cycle prevention for unlock edges.
* Server-side unlock evaluation on learner access paths.
* Role access, enrollment access, and manual unlock admin operations.
* Admin APIs to manage dependency edges; learner APIs for lock state and navigation.
* Outline lock/unlock indicators and navigation UI.
* Tests proving the three graphs remain distinct.

## Out of scope

* Resource upload productization (Phase 3) except consuming published content.
* New quiz scoring engines (Phase 4 owns assessment launch reuse).
* Recursive completion denominators and progress % (Phase 6).
* Course templates and integrated studio (Phase 7).
* Production analytics soak (Phase 8).
* Storing sequence or unlock edges as structural `parent_id` replacements.
* Redefining ancestor/breadcrumb repository queries (Phase 1 owns those).
* Structural tree editor or LearningItem metadata authoring (Phases 1–2).
* Product depth caps or authoritative child-ID arrays.

## Dependencies

* COURSE-P1-T04 — Tree repository queries.
* COURSE-P1-T16 — Ancestor and breadcrumb repository queries.
* COURSE-P1-T28 — Ancestor HTTP API.
* COURSE-P1-T29 — Breadcrumb HTTP API.
* COURSE-P2-T02 — LearningItem repository.
* COURSE-P2-T08 — Learner LearningItem APIs.
* COURSE-P2-T21 — Previous/Next item resolution (node-local; Phase 5 extends course-wide).
* COURSE-P3-T12 — Learner resource delivery APIs (unlock/nav consumes published resource delivery).
* COURSE-P4-T11 — Learner assessment launch (unlock may gate assessments).
* ADR-023; D-004; D-005.

## Phase boundaries

Phase 5 owns **sequence graph, prerequisite DAG, unlock evaluation, and navigation UX**.

| Owned here | Owned elsewhere |
|---|---|
| Course-wide sequence & Continue Learning | Structural tree mutations (Phase 1) |
| Prerequisite DAG persistence & cycle checks | Resource delivery (Phase 3) |
| Unlock rules (role, enrollment, manual) | Assessment scoring (Phase 4) |
| Lock-state learner APIs + navigation UI | Progress denominators (Phase 6) |
| | Ancestor/breadcrumb query definitions (Phase 1 — consume only) |
| | Templates / studio (Phase 7) |

## Execution sequence

1. **Design gates** — COURSE-P5-T01 and COURSE-P5-T02 must be ACCEPTED before migrations.
2. Persistence — prerequisite migration → sequence model.
3. Graph integrity — DAG cycle prevention → three-graph separation tests.
4. Navigation — consume P1 breadcrumbs/ancestors → course-wide prev/next/continue.
5. Unlock — server-side evaluation → role/enrollment/manual unlock.
6. HTTP — admin chaining/unlock APIs → learner lock-state APIs.
7. UI — navigation and lock/unlock indicators.
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
| COURSE-P5-T01 | Sequence representation decision | 3 | NOT_STARTED | S | 2h | — | COURSE-P2-T21 |
| COURSE-P5-T02 | Prerequisite persistence decision | 3 | NOT_STARTED | S | 2h | — | COURSE-P1-T04 |
| COURSE-P5-T03 | Prerequisite dependency schema & migration | 5 | NOT_STARTED | S | 2h | — | COURSE-P5-T02 |
| COURSE-P5-T04 | Learning sequence model (course-wide preorder / prev-next) | 7 | NOT_STARTED | M | 4h | — | COURSE-P5-T01 |
| COURSE-P5-T05 | Prerequisite DAG cycle prevention | 8 | NOT_STARTED | L | 6h | — | COURSE-P5-T03 |
| COURSE-P5-T06 | Three-graph separation tests & documentation | 6 | NOT_STARTED | M | 4h | — | COURSE-P5-T04, COURSE-P5-T05 |
| COURSE-P5-T07 | Breadcrumb & ancestor navigation integration (consume P1 APIs) | 5 | NOT_STARTED | M | 4h | — | COURSE-P1-T28, COURSE-P1-T29 |
| COURSE-P5-T08 | Unlock rules & server-side evaluation | 8 | NOT_STARTED | M | 4h | — | COURSE-P5-T05 |
| COURSE-P5-T09 | Role, enrollment & manual unlock operations | 7 | NOT_STARTED | M | 4h | — | COURSE-P5-T08 |
| COURSE-P5-T10 | Continue Learning & unlock-aware Previous/Next | 7 | NOT_STARTED | M | 4h | — | COURSE-P5-T04, COURSE-P5-T08 |
| COURSE-P5-T11 | Admin chaining & unlock APIs | 7 | NOT_STARTED | M | 4h | — | COURSE-P5-T09 |
| COURSE-P5-T12 | Learner lock-state & navigation APIs | 7 | NOT_STARTED | M | 4h | — | COURSE-P5-T11, COURSE-P2-T08 |
| COURSE-P5-T13 | Navigation & lock/unlock UI | 8 | NOT_STARTED | L | 6h | — | COURSE-P5-T12 |
| COURSE-P5-T14 | Backend unit & integration tests | 8 | NOT_STARTED | M | 4h | — | COURSE-P5-T12 |
| COURSE-P5-T15 | Frontend unit tests | 4 | NOT_STARTED | S | 2h | — | COURSE-P5-T13 |
| COURSE-P5-T16 | Playwright E2E verification | 6 | NOT_STARTED | M | 4h | — | COURSE-P5-T14, COURSE-P5-T15 |
| COURSE-P5-T17 | Phase 5 documentation & ledger sync | 3 | NOT_STARTED | S | 2h | — | COURSE-P5-T16 |
| COURSE-P5-T18 | Full canonical Phase 5 verification | 5 | NOT_STARTED | S | 2h | — | COURSE-P5-T14, COURSE-P5-T15, COURSE-P5-T16, COURSE-P5-T17 |

Total points: 107

## Task-specific acceptance criteria

### COURSE-P5-T01 — Sequence representation decision
<!-- TASK:COURSE-P5-T01:ACCEPTANCE:START -->
- [ ] Decision record defines how course-wide learning sequence is stored/computed (explicit edges vs derived preorder).
- [ ] Chosen representation is distinct from `parent_id` structural edges and prerequisite DAG edges.
- [ ] Decision recorded in `DECISIONS.md` and/or ADR before T04 implementation.
- [ ] Evidence under `docs/features/course-system/evidence/course-p5-t01/`.
- [ ] No migration, HTTP, or UI work is performed.
<!-- TASK:COURSE-P5-T01:ACCEPTANCE:END -->

### COURSE-P5-T02 — Prerequisite persistence decision
<!-- TASK:COURSE-P5-T02:ACCEPTANCE:START -->
- [ ] Decision record defines prerequisite/unlock edge persistence separately from sequence representation (T01).
- [ ] Chosen model preserves Course scope, `ON DELETE RESTRICT`, and no `child_node_ids[]`.
- [ ] Decision recorded in `DECISIONS.md` and/or ADR before T03 migration.
- [ ] Evidence under `docs/features/course-system/evidence/course-p5-t02/`.
- [ ] No migration, HTTP, or UI work is performed.
<!-- TASK:COURSE-P5-T02:ACCEPTANCE:END -->

### COURSE-P5-T03 — Prerequisite dependency schema & migration
<!-- TASK:COURSE-P5-T03:ACCEPTANCE:START -->
- [ ] Additive up/down migration implements T02 decision; edges distinct from `course_nodes.parent_id`.
- [ ] Edges are Course-scoped; cross-Course dependencies rejected by constraint or repository rules.
- [ ] FK rules use `ON DELETE RESTRICT`; isolated PostgreSQL apply/inspect/rollback/re-apply succeeds.
- [ ] Evidence under `docs/features/course-system/evidence/course-p5-t03/`.
- [ ] No sequence computation, HTTP, or UI work in this task.
<!-- TASK:COURSE-P5-T03:ACCEPTANCE:END -->

### COURSE-P5-T04 — Learning sequence model (course-wide preorder / prev-next)
<!-- TASK:COURSE-P5-T04:ACCEPTANCE:START -->
- [ ] Course-wide sequence resolution implements T01 decision at any depth (D-004).
- [ ] Prev/next resolution is distinct from structural parent/child and prerequisite edges.
- [ ] Deterministic ordering documented; no MAX_DEPTH rejection.
- [ ] Focused repository tests pass; evidence under `docs/features/course-system/evidence/course-p5-t04/`.
- [ ] Does not redefine COURSE-P2-T21 node-local chain; extends to course-wide scope.
<!-- TASK:COURSE-P5-T04:ACCEPTANCE:END -->

### COURSE-P5-T05 — Prerequisite DAG cycle prevention
<!-- TASK:COURSE-P5-T05:ACCEPTANCE:START -->
- [ ] Creating or updating prerequisite edges detects cycles in the unlock DAG.
- [ ] Cycle detection is separate from structural tree cycle rules (ADR-023 move checker).
- [ ] Cross-Course and self-loop edges are rejected with semantic errors.
- [ ] Focused tests pass; evidence under `docs/features/course-system/evidence/course-p5-t05/`.
- [ ] No HTTP or UI work in this task.
<!-- TASK:COURSE-P5-T05:ACCEPTANCE:END -->

### COURSE-P5-T06 — Three-graph separation tests & documentation
<!-- TASK:COURSE-P5-T06:ACCEPTANCE:START -->
- [ ] Automated tests prove structural, sequence, and prerequisite graphs can diverge without corruption.
- [ ] Documentation states the three graphs explicitly with non-collapse examples.
- [ ] Tests cover at least one scenario where sequence order ≠ structural sibling order.
- [ ] Evidence under `docs/features/course-system/evidence/course-p5-t06/`.
- [ ] No new persistence objects are introduced beyond T03/T04 scope.
<!-- TASK:COURSE-P5-T06:ACCEPTANCE:END -->

### COURSE-P5-T07 — Breadcrumb & ancestor navigation integration (consume P1 APIs)
<!-- TASK:COURSE-P5-T07:ACCEPTANCE:START -->
- [ ] Navigation UI/API layer consumes COURSE-P1-T28/T29; does not duplicate ancestor query logic.
- [ ] Breadcrumbs work at any depth without depth caps.
- [ ] Cross-Course access returns 404; no path leakage.
- [ ] Evidence under `docs/features/course-system/evidence/course-p5-t07/`.
- [ ] No changes to Phase 1 ancestor repository contracts unless a bugfix ADR is approved.
<!-- TASK:COURSE-P5-T07:ACCEPTANCE:END -->

### COURSE-P5-T08 — Unlock rules & server-side evaluation
<!-- TASK:COURSE-P5-T08:ACCEPTANCE:START -->
- [ ] Lock state is computed server-side from prerequisite DAG + published content rules.
- [ ] Client cannot bypass lock by calling deep URLs directly.
- [ ] Draft/unpublished targets do not appear as unlocked learner destinations.
- [ ] Focused tests pass; evidence under `docs/features/course-system/evidence/course-p5-t08/`.
- [ ] No progress percentage logic (Phase 6) is required.
<!-- TASK:COURSE-P5-T08:ACCEPTANCE:END -->

### COURSE-P5-T09 — Role, enrollment & manual unlock operations
<!-- TASK:COURSE-P5-T09:ACCEPTANCE:START -->
- [ ] Role-based, enrollment-based, and manual admin unlock paths are server-validated.
- [ ] Manual unlock is auditable (actor, target, timestamp) even if full audit schema is Phase 8.
- [ ] Unauthorized manual unlock attempts are rejected without side effects.
- [ ] Focused tests pass; evidence under `docs/features/course-system/evidence/course-p5-t09/`.
- [ ] Enrollment commerce productization may remain deferred with documented policy.
<!-- TASK:COURSE-P5-T09:ACCEPTANCE:END -->

### COURSE-P5-T10 — Continue Learning & unlock-aware Previous/Next
<!-- TASK:COURSE-P5-T10:ACCEPTANCE:START -->
- [ ] Continue Learning resolves the next unlocked published LearningItem in course-wide sequence.
- [ ] Previous/Next skip locked/unpublished items per server rules.
- [ ] Resolution works at any depth; no MAX_DEPTH constant.
- [ ] Focused tests pass; evidence under `docs/features/course-system/evidence/course-p5-t10/`.
- [ ] Progress-weighted Continue Learning enhancements defer to Phase 6 integration.
<!-- TASK:COURSE-P5-T10:ACCEPTANCE:END -->

### COURSE-P5-T11 — Admin chaining & unlock APIs
<!-- TASK:COURSE-P5-T11:ACCEPTANCE:START -->
- [ ] Admin APIs manage prerequisite edges and manual unlock without mutating `parent_id`.
- [ ] Controllers are transport-only; graph rules remain service/repository-owned.
- [ ] Cycle attempts return semantic errors; no partial edge persistence.
- [ ] Controller tests pass; evidence under `docs/features/course-system/evidence/course-p5-t11/`.
<!-- TASK:COURSE-P5-T11:ACCEPTANCE:END -->

### COURSE-P5-T12 — Learner lock-state & navigation APIs
<!-- TASK:COURSE-P5-T12:ACCEPTANCE:START -->
- [ ] Learner APIs expose lock state, prev/next, and continue-learning projections.
- [ ] Responses respect `publish_state` and enrollment/role rules.
- [ ] Assessment launch endpoints (COURSE-P4-T11) can consume lock state when integrated.
- [ ] Controller tests pass; evidence under `docs/features/course-system/evidence/course-p5-t12/`.
<!-- TASK:COURSE-P5-T12:ACCEPTANCE:END -->

### COURSE-P5-T13 — Navigation & lock/unlock UI
<!-- TASK:COURSE-P5-T13:ACCEPTANCE:START -->
- [ ] Outline shows lock/unlock indicators driven by server lock-state APIs.
- [ ] Continue Learning and prev/next controls reflect unlock-aware resolution.
- [ ] UI does not persist structural authority or invent depth limits.
- [ ] Frontend lint/tests pass; evidence under `docs/features/course-system/evidence/course-p5-t13/`.
<!-- TASK:COURSE-P5-T13:ACCEPTANCE:END -->

### COURSE-P5-T14 — Backend unit & integration tests
<!-- TASK:COURSE-P5-T14:ACCEPTANCE:START -->
- [ ] `go test` covers sequence resolution, DAG cycle prevention, unlock evaluation, and HTTP mapping.
- [ ] Three-graph separation scenarios included in integration suite.
- [ ] Regression on Phase 1 ancestor APIs and Phase 2 LearningItem packages passes.
- [ ] Evidence under `docs/features/course-system/evidence/course-p5-t14/`.
<!-- TASK:COURSE-P5-T14:ACCEPTANCE:END -->

### COURSE-P5-T15 — Frontend unit tests
<!-- TASK:COURSE-P5-T15:ACCEPTANCE:START -->
- [ ] Vitest coverage exists for navigation and lock-state UI helpers.
- [ ] `npm run lint` and `npm test -- --run` pass for changed app scope.
- [ ] Evidence under `docs/features/course-system/evidence/course-p5-t15/`.
<!-- TASK:COURSE-P5-T15:ACCEPTANCE:END -->

### COURSE-P5-T16 — Playwright E2E verification
<!-- TASK:COURSE-P5-T16:ACCEPTANCE:START -->
- [ ] Playwright covers locked item blocked → prerequisite satisfied → unlock → continue learning flow.
- [ ] Breadcrumb navigation exercised on a nested node.
- [ ] Screenshots/logs stored; `npx playwright test` exit code recorded.
- [ ] Evidence under `docs/features/course-system/evidence/course-p5-t16/`.
<!-- TASK:COURSE-P5-T16:ACCEPTANCE:END -->

### COURSE-P5-T17 — Phase 5 documentation & ledger sync
<!-- TASK:COURSE-P5-T17:ACCEPTANCE:START -->
- [ ] Phase docs document three-graph separation and P1 breadcrumb consumption.
- [ ] Architecture freeze invariants restated; no "leaf node" product language.
- [ ] `npm run course-system:status:sync` then `npm run course-system:status:check` succeed; second sync is a no-op.
- [ ] Evidence under `docs/features/course-system/evidence/course-p5-t17/`.
<!-- TASK:COURSE-P5-T17:ACCEPTANCE:END -->

### COURSE-P5-T18 — Full canonical Phase 5 verification
<!-- TASK:COURSE-P5-T18:ACCEPTANCE:START -->
- [ ] All Phase 5 verification commands pass with logged exit codes and timestamps.
- [ ] Phase acceptance criteria checklist is evaluated; failures recorded explicitly.
- [ ] Evidence under `docs/features/course-system/evidence/course-p5-t18/`.
- [ ] Three-graph separation confirmed in integration evidence.
<!-- TASK:COURSE-P5-T18:ACCEPTANCE:END -->

## Phase acceptance criteria

- [ ] COURSE-P5-T01 and COURSE-P5-T02 decisions are ACCEPTED before persistence work.
- [ ] Prerequisite persistence exists distinct from `parent_id` structural edges.
- [ ] Course-wide sequence, prev/next, and Continue Learning work at any depth.
- [ ] Prerequisite DAG cycle prevention is separate from structural cycle rules.
- [ ] Three graphs remain distinct under test and documentation.
- [ ] Breadcrumb/ancestor navigation consumes Phase 1 APIs without redefinition.
- [ ] Role, enrollment, and manual unlock are server-enforced.
- [ ] Lock-state learner APIs and navigation UI are functional.
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

**Full phase verify** (COURSE-P5-T18):

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
`docs/features/course-system/evidence/course-p5-tXX/`:

* README summarizing scope and non-goals.
* Command logs with exit codes and verification classification.
* Hash / sync artifacts when ledger tooling is run.
* Migration prove-out notes when a migration is in scope.
* Explicit three-graph non-collapse confirmation where applicable.
* Explicit statement of Database Migration: NONE when applicable.

## Security requirements

* Lock state and unlock evaluation are server-authoritative.
* Manual unlock requires admin authorization; unauthorized attempts are rejected.
* Deep-link access to locked content returns consistent not-found/forbidden semantics.
* Prerequisite edge mutations require admin authorization.
* Enrollment and role checks remain server-side.

## Performance requirements

* Course-wide sequence resolution must not require loading entire unpublished subtrees eagerly.
* DAG cycle detection must complete within documented bounds for realistic edge counts.
* Lock-state queries should be batchable for outline views; document N+1 avoidance.
* Deep-tree navigation (≥25 levels) must succeed without depth rejection.

## Accessibility / mobile

* Lock/unlock indicators expose accessible state (not color-only).
* Continue Learning and prev/next controls are keyboard-operable.
* Breadcrumb navigation is screen-reader friendly with meaningful link text.
* Mobile outline navigation remains usable without horizontal overflow.

## Risks

| Risk | Mitigation |
|---|---|
| Collapsing sequence into structural tree | T01/T06 gates and separation tests |
| Collapsing prerequisites into parent_id | T02/T03 distinct persistence |
| Redefining Phase 1 ancestor APIs | T07 consumes P1-T28/T29 only |
| Client-side unlock bypass | Server lock evaluation on every learner path |
| Confusing structural cycles with DAG cycles | Separate checkers documented and tested |

## Known limitations

* Progress-weighted Continue Learning fully integrates in Phase 6.
* Enrollment commerce may remain deferred per Phase 2 pattern.
* Full audit log schema defers to Phase 8.
* Operational edge-count limits are not product MAX_*_DEPTH.

## Exit criteria

Phase 5 may be marked complete only when:

1. All checklist tasks are `VERIFIED` with canonical evidence.
2. Declared total points (**107**) match the ledger sum and status tooling.
3. Full phase verification commands pass for changed surfaces.
4. Three graphs confirmed distinct; Phase 1 breadcrumbs consumed not redefined.
5. COURSE-P5-T01 and COURSE-P5-T02 decisions ACCEPTED in `DECISIONS.md` and/or ADRs.
6. HANDOFF records the next safe action (typically Phase 6 start or approved parallel exception).
