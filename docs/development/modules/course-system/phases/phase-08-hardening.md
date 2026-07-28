# Phase 8 — Production Hardening

* **Status**: NOT_STARTED
* **Weight**: 5%
* **Phase owner**: Unassigned
* **Started at**: —
* **Verified at**: —

## Objective

Harden the Course System for production: course admin authorization model (replacing
quiz-admin allowlist), audit logs, permission validation matrix, caching, large tree
performance, stress and concurrency tests, accessibility, security review, Sentry
and distributed tracing, analytics instrumentation, backup and recovery procedures,
and a production deployment checklist.

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

Phase 8 adds operational resilience without altering the hierarchy contract.
Performance optimizations (caching, indexing, batching) are operational concerns,
not product depth caps.

## Current verified baseline

No Phase 8 tasks are verified. Phase 8 may not start authorization or audit
implementation until **COURSE-P8-T01** (course admin authorization model decision)
is ACCEPTED and recorded in `DECISIONS.md` and/or an ADR.

## In scope

* Design gate: course admin authorization model (T01) — replace quiz-admin allowlist.
* Audit log schema and emission for admin mutations and sensitive learner actions.
* Permission validation matrix covering all Course System admin/learner endpoints.
* Caching layer for tree reads, lock state, and aggregated progress.
* Large tree performance optimization and benchmarking.
* Concurrency and stress tests for hierarchy mutations and progress updates.
* Accessibility audit and remediation for builder and learner surfaces.
* Security review and hardening (authorization, input validation, rate limits).
* Sentry error reporting and distributed tracing integration.
* Analytics instrumentation for course engagement events.
* Backup and recovery procedures for course data.
* Production deployment checklist and runbook.

## Out of scope

* New feature development (resources, assessments, unlock rules, progress logic).
* NUC production cutover without explicit approval and rollback plan.
* Destructive database or volume operations.
* Product depth caps introduced as schema rules.
* Replacement of Go/Nuxt/Kratos architecture.

## Dependencies

* COURSE-P1-T08 — Admin course APIs.
* COURSE-P1-T10 — Admin hierarchy mutation APIs.
* COURSE-P2-T06 — Admin learning item API endpoints.
* COURSE-P3-T11 — Admin resource API endpoints.
* COURSE-P4-T10 — Admin assessment API endpoints.
* COURSE-P5-T11 — Admin chaining & unlock APIs.
* COURSE-P5-T12 — Learner lock-state & navigation APIs.
* COURSE-P6-T08 — Dashboard progress APIs.
* COURSE-P7-T11 — Admin template & builder APIs.
* ADR-023; D-004; D-005; AGENTS.md NUC deployment safety rules.
* Existing Compose, Sentry, and observability infrastructure in the repository.

## Phase boundaries

Phase 8 owns **operational hardening only** — no new product features.

| Owned here | Owned elsewhere |
|---|---|
| Course admin auth model (T01 gate) | Feature logic (Phases 1–7) |
| Audit logs, permission matrix | Hierarchy semantics (Phase 1 / ADR-023) |
| Caching, performance, stress tests | Resource/assessment engines (Phases 3–4) |
| Security review, Sentry, tracing | Unlock/progress algorithms (Phases 5–6) |
| Backup/recovery, deployment checklist | Template/studio features (Phase 7) |

## Execution sequence

1. **Design gate** — COURSE-P8-T01 must be ACCEPTED before auth/audit rollout.
2. Authorization — implement course admin model per T01 decision.
3. Observability — audit logs → Sentry/tracing → analytics instrumentation.
4. Security — permission matrix → security review & hardening.
5. Performance — caching → large tree optimization → concurrency/stress tests.
6. Accessibility — audit and remediation.
7. Operations — backup/recovery → production deployment checklist.
8. Verification — hardening test suites → Playwright smoke → docs → full phase verify.

Typical layering per task: Design → Migration (if needed) → Service → Tests →
Docs → Integration verify.

## Completion formula

Phase completion equals:

$$\text{Phase Completion} = \frac{\text{Verified Task Points}}{\text{Total Task Points}} \times 100$$

Only `VERIFIED` tasks count. Declared `Total points` must equal the arithmetic
sum of checklist points.

## Task checklist

Checklist columns: Evidence is column 7 for ledger tooling; Dependencies is last.

| ID | Task | Points | Status | Difficulty | Est. Time | Evidence | Dependencies |
|---|---|---:|---|---|---|---|---|
| COURSE-P8-T01 | Course admin authorization model decision (replace quiz-admin allowlist) | 3 | NOT_STARTED | S | 2h | — | COURSE-P1-T08 |
| COURSE-P8-T02 | Audit log schema & emission | 5 | NOT_STARTED | S | 2h | — | COURSE-P8-T01 |
| COURSE-P8-T03 | Permission validation matrix | 7 | NOT_STARTED | M | 4h | — | COURSE-P8-T01 |
| COURSE-P8-T04 | Caching layer (tree, lock, progress) | 6 | NOT_STARTED | M | 4h | — | COURSE-P1-T04, COURSE-P5-T12, COURSE-P6-T08 |
| COURSE-P8-T05 | Large tree performance optimization | 6 | NOT_STARTED | M | 4h | — | COURSE-P1-T36, COURSE-P7-T10 |
| COURSE-P8-T06 | Concurrency & stress tests | 7 | NOT_STARTED | L | 6h | — | COURSE-P8-T05 |
| COURSE-P8-T07 | Accessibility audit & fixes | 6 | NOT_STARTED | M | 4h | — | COURSE-P1-T15, COURSE-P7-T08 |
| COURSE-P8-T08 | Security review & hardening | 7 | NOT_STARTED | M | 4h | — | COURSE-P8-T03 |
| COURSE-P8-T09 | Sentry & distributed tracing | 5 | NOT_STARTED | M | 4h | — | COURSE-P8-T02 |
| COURSE-P8-T10 | Analytics instrumentation | 4 | NOT_STARTED | S | 2h | — | COURSE-P6-T08 |
| COURSE-P8-T11 | Backup & recovery procedures | 5 | NOT_STARTED | M | 4h | — | COURSE-P1-T03 |
| COURSE-P8-T12 | Production deployment checklist | 5 | NOT_STARTED | S | 2h | — | COURSE-P8-T11 |
| COURSE-P8-T13 | Backend hardening test suite | 7 | NOT_STARTED | M | 4h | — | COURSE-P8-T02, COURSE-P8-T03, COURSE-P8-T04, COURSE-P8-T06, COURSE-P8-T08 |
| COURSE-P8-T14 | Frontend hardening tests | 4 | NOT_STARTED | S | 2h | — | COURSE-P8-T07 |
| COURSE-P8-T15 | Playwright smoke on production-like stack | 5 | NOT_STARTED | M | 4h | — | COURSE-P8-T13, COURSE-P8-T14 |
| COURSE-P8-T16 | Phase 8 documentation & ledger sync | 3 | NOT_STARTED | S | 2h | — | COURSE-P8-T15 |
| COURSE-P8-T17 | Full canonical Phase 8 verification | 5 | NOT_STARTED | S | 2h | — | COURSE-P8-T13, COURSE-P8-T14, COURSE-P8-T15, COURSE-P8-T16 |

Total points: 90

## Task-specific acceptance criteria

### COURSE-P8-T01 — Course admin authorization model decision (replace quiz-admin allowlist)
<!-- TASK:COURSE-P8-T01:ACCEPTANCE:START -->
- [ ] Decision record defines course admin roles/permissions replacing quiz-admin allowlist pattern.
- [ ] Model covers all Phase 1–7 admin endpoints with server-side enforcement plan.
- [ ] Decision recorded in `DECISIONS.md` and/or ADR before T02/T03 implementation.
- [ ] Evidence under `docs/features/course-system/evidence/course-p8-t01/`.
- [ ] No code changes to authorization middleware in this task unless decision requires spike behind flag.
<!-- TASK:COURSE-P8-T01:ACCEPTANCE:END -->

### COURSE-P8-T02 — Audit log schema & emission
<!-- TASK:COURSE-P8-T02:ACCEPTANCE:START -->
- [ ] Additive migration creates audit log persistence for Course System admin mutations.
- [ ] Audit entries capture actor, action, target Course/node/item, timestamp, and before/after snapshot where applicable.
- [ ] Audit emission does not alter hierarchy schema or introduce `child_node_ids[]`.
- [ ] FK rules use `ON DELETE RESTRICT`; focused tests pass.
- [ ] Evidence under `docs/features/course-system/evidence/course-p8-t02/`.
- [ ] Secrets are not written to audit payloads.
<!-- TASK:COURSE-P8-T02:ACCEPTANCE:END -->

### COURSE-P8-T03 — Permission validation matrix
<!-- TASK:COURSE-P8-T03:ACCEPTANCE:START -->
- [ ] Documented matrix maps every admin and learner Course System endpoint to required auth per T01 model.
- [ ] Automated tests verify unauthorized access is denied for each endpoint class.
- [ ] Matrix confirms server-side enforcement; client checks are not sufficient.
- [ ] Evidence under `docs/features/course-system/evidence/course-p8-t03/`.
- [ ] Quiz-admin allowlist replaced or bridged per T01 decision.
<!-- TASK:COURSE-P8-T03:ACCEPTANCE:END -->

### COURSE-P8-T04 — Caching layer (tree, lock, progress)
<!-- TASK:COURSE-P8-T04:ACCEPTANCE:START -->
- [ ] Caching strategy documented for tree reads, lock state, and progress aggregates.
- [ ] Cache invalidation hooks exist for hierarchy mutations (COURSE-P1-T10) and progress updates (COURSE-P6-T08).
- [ ] Stale cache cannot serve cross-Course data or draft content to learners.
- [ ] Focused tests pass; evidence under `docs/features/course-system/evidence/course-p8-t04/`.
- [ ] Caching is operational; no product MAX_DEPTH introduced.
<!-- TASK:COURSE-P8-T04:ACCEPTANCE:END -->

### COURSE-P8-T05 — Large tree performance optimization
<!-- TASK:COURSE-P8-T05:ACCEPTANCE:START -->
- [ ] Benchmarks compare before/after for large synthetic Courses (thresholds in evidence).
- [ ] Optimizations use lazy/paginated patterns; do not introduce depth columns or caps.
- [ ] Query plans or measurement notes included for hot paths.
- [ ] Evidence under `docs/features/course-system/evidence/course-p8-t05/`.
<!-- TASK:COURSE-P8-T05:ACCEPTANCE:END -->

### COURSE-P8-T06 — Concurrency & stress tests
<!-- TASK:COURSE-P8-T06:ACCEPTANCE:START -->
- [ ] Stress tests cover concurrent hierarchy mutations and progress updates without invariant violation.
- [ ] No orphan `parent_id`, dual-parenting, or cross-Course edges after concurrent operations.
- [ ] Test commands and exit codes logged.
- [ ] Evidence under `docs/features/course-system/evidence/course-p8-t06/`.
- [ ] No production or NUC deployment is performed.
<!-- TASK:COURSE-P8-T06:ACCEPTANCE:END -->

### COURSE-P8-T07 — Accessibility audit & fixes
<!-- TASK:COURSE-P8-T07:ACCEPTANCE:START -->
- [ ] Accessibility audit covers builder and learner Course System surfaces.
- [ ] Critical WCAG blockers identified in audit are remediated or explicitly deferred with ticket references.
- [ ] Keyboard navigation paths verified for tree builder and navigation UI.
- [ ] Evidence under `docs/features/course-system/evidence/course-p8-t07/`.
<!-- TASK:COURSE-P8-T07:ACCEPTANCE:END -->

### COURSE-P8-T08 — Security review & hardening
<!-- TASK:COURSE-P8-T08:ACCEPTANCE:START -->
- [ ] Security review covers authorization, input validation, rate limits, and signed URL policies.
- [ ] Findings triaged; critical items fixed or explicitly accepted with rationale.
- [ ] No secrets in Git; example configs use placeholders.
- [ ] Evidence under `docs/features/course-system/evidence/course-p8-t08/`.
<!-- TASK:COURSE-P8-T08:ACCEPTANCE:END -->

### COURSE-P8-T09 — Sentry & distributed tracing
<!-- TASK:COURSE-P8-T09:ACCEPTANCE:START -->
- [ ] Course System handlers emit Sentry errors for unhandled failures per existing project patterns.
- [ ] Distributed tracing spans cover critical admin mutations and learner navigation paths.
- [ ] PII/secrets redaction confirmed in trace/error payloads.
- [ ] Evidence under `docs/features/course-system/evidence/course-p8-t09/`.
<!-- TASK:COURSE-P8-T09:ACCEPTANCE:END -->

### COURSE-P8-T10 — Analytics instrumentation
<!-- TASK:COURSE-P8-T10:ACCEPTANCE:START -->
- [ ] Course engagement events instrumented (completion, assessment launch, continue learning) per documented schema.
- [ ] Analytics do not duplicate server-authoritative scoring or progress as client truth.
- [ ] Evidence under `docs/features/course-system/evidence/course-p8-t10/`.
- [ ] Full analytics productization beyond instrumentation is not required.
<!-- TASK:COURSE-P8-T10:ACCEPTANCE:END -->

### COURSE-P8-T11 — Backup & recovery procedures
<!-- TASK:COURSE-P8-T11:ACCEPTANCE:START -->
- [ ] Documented backup procedure for Course System tables and MinIO objects referenced by resources.
- [ ] Recovery drill steps recorded with timestamps (local/staging; not production NUC without approval).
- [ ] Rollback route documented per AGENTS.md NUC safety rules.
- [ ] Evidence under `docs/features/course-system/evidence/course-p8-t11/`.
- [ ] No destructive volume operations performed.
<!-- TASK:COURSE-P8-T11:ACCEPTANCE:END -->

### COURSE-P8-T12 — Production deployment checklist
<!-- TASK:COURSE-P8-T12:ACCEPTANCE:START -->
- [ ] Checklist covers migrate up, API health, web build, Kratos auth, Redis, WebSocket, HTTPS verification.
- [ ] Explicit rollback steps and "do not cutover traffic" guardrails included.
- [ ] Aligns with AGENTS.md NUC deployment safety rules.
- [ ] Evidence under `docs/features/course-system/evidence/course-p8-t12/`.
- [ ] No production traffic cutover performed as part of this task.
<!-- TASK:COURSE-P8-T12:ACCEPTANCE:END -->

### COURSE-P8-T13 — Backend hardening test suite
<!-- TASK:COURSE-P8-T13:ACCEPTANCE:START -->
- [ ] `go test ./...` passes including audit, permission, cache, and stress-related tests.
- [ ] `docker compose --profile verify run --rm api-verify` passes.
- [ ] Evidence under `docs/features/course-system/evidence/course-p8-t13/`.
<!-- TASK:COURSE-P8-T13:ACCEPTANCE:END -->

### COURSE-P8-T14 — Frontend hardening tests
<!-- TASK:COURSE-P8-T14:ACCEPTANCE:START -->
- [ ] `npm run lint` and `npm test -- --run` pass for app workspace.
- [ ] Accessibility-related unit tests added where applicable.
- [ ] Evidence under `docs/features/course-system/evidence/course-p8-t14/`.
<!-- TASK:COURSE-P8-T14:ACCEPTANCE:END -->

### COURSE-P8-T15 — Playwright smoke on production-like stack
<!-- TASK:COURSE-P8-T15:ACCEPTANCE:START -->
- [ ] Playwright smoke covers admin course edit, learner navigation, and one assessment/resource path.
- [ ] Run against production-like Compose stack; exit code recorded.
- [ ] Evidence under `docs/features/course-system/evidence/course-p8-t15/`.
- [ ] CLI note: `npx playwright test` from `app/`.
<!-- TASK:COURSE-P8-T15:ACCEPTANCE:END -->

### COURSE-P8-T16 — Phase 8 documentation & ledger sync
<!-- TASK:COURSE-P8-T16:ACCEPTANCE:START -->
- [ ] HANDOFF, CHANGELOG, and deployment docs updated for Phase 8 hardening.
- [ ] Course admin auth model (T01) referenced in architecture docs.
- [ ] `npm run course-system:status:sync` then `npm run course-system:status:check` succeed; second sync is a no-op.
- [ ] Evidence under `docs/features/course-system/evidence/course-p8-t16/`.
<!-- TASK:COURSE-P8-T16:ACCEPTANCE:END -->

### COURSE-P8-T17 — Full canonical Phase 8 verification
<!-- TASK:COURSE-P8-T17:ACCEPTANCE:START -->
- [ ] All Phase 8 verification commands pass with logged exit codes and timestamps.
- [ ] Phase acceptance criteria checklist is evaluated; failures recorded explicitly.
- [ ] Evidence under `docs/features/course-system/evidence/course-p8-t17/`.
- [ ] Confirms all eight phases can report VERIFIED status in MASTER_PLAN.
<!-- TASK:COURSE-P8-T17:ACCEPTANCE:END -->

## Phase acceptance criteria

- [ ] COURSE-P8-T01 authorization model is ACCEPTED and implemented.
- [ ] Audit logs emit for admin mutations; permission matrix tests pass.
- [ ] Caching layer operational with correct invalidation.
- [ ] Large tree performance benchmarks meet documented thresholds.
- [ ] Concurrency/stress tests pass without hierarchy invariant violations.
- [ ] Accessibility audit critical items remediated or explicitly deferred.
- [ ] Security review findings triaged; critical items addressed.
- [ ] Sentry/tracing and analytics instrumentation functional.
- [ ] Backup/recovery and deployment checklist documented.
- [ ] Backend, frontend, Playwright smoke, and ledger sync/check pass.
- [ ] No production cutover performed without approval.

## Verification commands

**Task-focused**:

```text
# API targeted tests (cwd: api/)
go test ./... -count=1 -run <TaskPattern>

# Permission matrix / stress (cwd: api/)
go test ./... -count=1 -run <HardeningPattern>
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

**Full phase verify** (COURSE-P8-T17):

```text
# API (cwd: api/)
go test ./...

# Compose verify profile (repo root)
docker compose --profile verify run --rm api-verify

# Frontend (cwd: app/)
npm run lint
npm test -- --run

# E2E smoke (cwd: app/) — CLI note only
npx playwright test

# Ledger (repo root)
npm run course-system:status:sync
npm run course-system:status:check
```

Record exact timestamps, durations, exit codes, and environments in task evidence.

## Evidence requirements

Each VERIFIED task must provide under
`docs/features/course-system/evidence/course-p8-tXX/`:

* README summarizing scope and non-goals.
* Command logs with exit codes and verification classification.
* Hash / sync artifacts when ledger tooling is run.
* Migration prove-out notes when a migration is in scope.
* Security/analytics review notes when applicable.
* Explicit statement of Database Migration: NONE when applicable.

## Security requirements

* Course admin authorization model (T01) replaces quiz-admin allowlist for all admin endpoints.
* Permission matrix tests must cover every admin and learner Course System route class.
* Rate limits documented for mutation-heavy and signed-URL endpoints.
* Audit logs must not contain secrets, tokens, or full signed URLs.
* Input validation hardening covers all user-controlled hierarchy and metadata fields.
* NUC deployment requires explicit approval; no destructive operations in evidence gathering.

## Performance requirements

* Large tree read latency documented before/after optimization (COURSE-P8-T05).
* Cache hit/miss invalidation must not serve stale cross-Course or draft data.
* Concurrency tests validate hierarchy mutations under parallel load.
* Operational limits (payload, rate, page size) documented as non-depth constraints.

## Accessibility / mobile

* WCAG audit covers builder drag-drop alternatives and learner navigation (COURSE-P8-T07).
* Critical keyboard traps remediated in tree builder and assessment flows.
* Mobile layouts verified for learner navigation and progress surfaces.
* Color-only lock/progress indicators remediated or documented with accessible alternatives.

## Risks

| Risk | Mitigation |
|---|---|
| Shipping with quiz-admin allowlist indefinitely | T01 gate blocks T02/T03 until ACCEPTED |
| Cache serving stale lock/progress state | Invalidation hooks on mutation paths |
| Production cutover without rollback | Checklist + AGENTS.md guardrails |
| Stress tests exposing race in move/reorder | T06 before declaring phase complete |
| Audit log PII leakage | Redaction rules in T02/T08 |

## Known limitations

* NUC production cutover requires explicit human approval outside this phase.
* Full CDN/adaptive streaming productization remains out of scope.
* Some WCAG items may defer with explicit ticket references if non-blocking.
* Analytics instrumentation ≠ full BI productization.

## Exit criteria

Phase 8 may be marked complete only when:

1. All checklist tasks are `VERIFIED` with canonical evidence.
2. Declared total points (**90**) match the ledger sum and status tooling.
3. Full phase verification commands pass for changed surfaces.
4. Course admin authorization model (T01) ACCEPTED and implemented.
5. Backup/recovery and deployment checklist documented; no unapproved production cutover.
6. All eight Course System phases report VERIFIED in MASTER_PLAN; HANDOFF records module completion.
