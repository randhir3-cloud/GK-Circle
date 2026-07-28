# Phase 3 — Resources and Native Content

* **Status**: NOT_STARTED
* **Weight**: 15%
* **Phase owner**: Unassigned
* **Started at**: —
* **Verified at**: —

## Objective

Deliver native and uploaded learning resources — video, PDF, audio, image, rich
text, Markdown, external links, downloadables, streaming, signed URLs, multilingual
explanations, question booklets, video solutions, and lecture recordings — attached
to LearningItems under CourseNodes at any logical depth, with upload and replacement
lifecycles, admin and learner APIs, and integrated viewers.

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
* Media bytes live in MinIO; metadata and access rules live in PostgreSQL /
  LearningItem metadata — not as fixed Course columns.
* Resources attach to LearningItems on CourseNodes at any nesting level without
  altering ADR-023 hierarchy contract (**D-005**).

## Current verified baseline

No Phase 3 tasks are verified. Phase 3 may not start migrations or persistence
work until **COURSE-P3-T01** (resource persistence decision) is accepted and recorded
in `DECISIONS.md` and/or an ADR.

## In scope

* Design gate: first-class Resource table vs typed LearningItem metadata (T01).
* Additive migrations for resource / attachment metadata (after T01).
* MinIO upload lifecycle, object keys, replacement lifecycle, and signed URL delivery.
* Native content types: rich text, Markdown, PDF, document, image, audio, video,
  external link, downloadable attachments.
* Streaming policy and time-limited signed URLs for media delivery.
* Multilingual resource metadata (English, Hindi, extensible locale fields).
* Specialized attachments: question booklet, video solution, lecture recording.
* Server-side URL safety and download-policy enforcement.
* Admin CRUD and authenticated learner delivery APIs.
* Viewer UI for supported resource types (admin preview + learner consumption).
* Backend, frontend, Playwright, ledger, and migration evidence.

## Out of scope

* Quiz engine scoring / sessions (Phase 4).
* Prerequisite DAG and learning-sequence graphs (Phase 5).
* Progress aggregation and completion denominators (Phase 6).
* Course templates and full visual builder (Phase 7).
* Production cutover, analytics productization, and hardening soak (Phase 8).
* Introducing `MAX_*_DEPTH`, persisted `depth` enforcement, or `child_node_ids[]`.
* Renaming CourseNode `status` to `publish_state` (or the reverse).

## Dependencies

* COURSE-P2-T02 — LearningItem repository CRUD.
* COURSE-P2-T03 — Information block metadata envelope.
* COURSE-P2-T07 — LearningItem `publish_state`.
* COURSE-P2-T08 — Learner LearningItem APIs.
* COURSE-P1-T03 — CourseNode persistence (attachment scope).
* ADR-023; module decisions **D-004** / **D-005**.
* Existing MinIO Compose service and object-storage patterns in the API.

## Phase boundaries

Phase 3 owns **resource persistence, upload, delivery, and viewers only**.

| Owned here | Owned elsewhere |
|---|---|
| MinIO upload/replace/signed delivery | Quiz scoring (Phase 4) |
| Resource metadata and URL safety | Sequence / unlock graphs (Phase 5) |
| Native content block extensions | Progress denominators (Phase 6) |
| Admin/learner resource APIs + UI | Templates / studio (Phase 7) |
| | Course admin auth model (Phase 8 T01 gate) |

## Execution sequence

1. **Design gate** — COURSE-P3-T01 must be ACCEPTED before any migration.
2. Persistence — migration → model → repository tests.
3. Storage — MinIO lifecycle → signed URL policy.
4. Content types — native blocks → media → links → multilingual/specialized.
5. HTTP — admin APIs → learner delivery APIs.
6. UI — resource viewer (admin preview + learner).
7. Verification — backend suite → frontend tests → Playwright → docs → full phase verify.

Typical layering per task: Design → Migration → Repository → Service → DTO →
Controller → API tests → Frontend → Playwright → Docs.

## Completion formula

Phase completion equals:

$$\text{Phase Completion} = \frac{\text{Verified Task Points}}{\text{Total Task Points}} \times 100$$

Only tasks marked `VERIFIED` contribute to the numerator. Declared
`Total points` must equal the arithmetic sum of checklist points (ledger denominator).

## Task checklist

Checklist columns: Evidence is column 7 for ledger tooling; Dependencies is last.

| ID | Task | Points | Status | Difficulty | Est. Time | Evidence | Dependencies |
|---|---|---:|---|---|---|---|---|
| COURSE-P3-T01 | Resource persistence decision (first-class Resource vs LearningItem metadata) | 3 | NOT_STARTED | S | 2h | — | COURSE-P2-T03 |
| COURSE-P3-T02 | Resource attachment schema & migration | 5 | NOT_STARTED | S | 2h | — | COURSE-P3-T01 |
| COURSE-P3-T03 | Resource Go model & repository | 7 | NOT_STARTED | M | 4h | — | COURSE-P3-T02 |
| COURSE-P3-T04 | MinIO upload lifecycle & object keys | 8 | NOT_STARTED | M | 4h | — | COURSE-P3-T03 |
| COURSE-P3-T05 | Signed URL delivery & streaming policy | 7 | NOT_STARTED | M | 4h | — | COURSE-P3-T04 |
| COURSE-P3-T06 | Rich text & Markdown native content blocks | 6 | NOT_STARTED | M | 4h | — | COURSE-P3-T03 |
| COURSE-P3-T07 | PDF, document, image, audio & video handling | 8 | NOT_STARTED | M | 4h | — | COURSE-P3-T04, COURSE-P3-T05 |
| COURSE-P3-T08 | External link URL safety & downloadables | 5 | NOT_STARTED | S | 2h | — | COURSE-P3-T03 |
| COURSE-P3-T09 | Multilingual metadata & specialized attachments | 7 | NOT_STARTED | M | 4h | — | COURSE-P3-T06, COURSE-P3-T07 |
| COURSE-P3-T10 | Resource replacement lifecycle | 6 | NOT_STARTED | M | 4h | — | COURSE-P3-T04 |
| COURSE-P3-T11 | Admin resource API endpoints | 8 | NOT_STARTED | M | 4h | — | COURSE-P3-T06, COURSE-P3-T07, COURSE-P3-T08, COURSE-P3-T09, COURSE-P3-T10 |
| COURSE-P3-T12 | Learner resource delivery APIs | 7 | NOT_STARTED | M | 4h | — | COURSE-P3-T11, COURSE-P2-T08 |
| COURSE-P3-T13 | Resource viewer UI (admin + learner) | 8 | NOT_STARTED | L | 6h | — | COURSE-P3-T12 |
| COURSE-P3-T14 | Backend unit & integration tests | 8 | NOT_STARTED | M | 4h | — | COURSE-P3-T12 |
| COURSE-P3-T15 | Frontend unit tests | 4 | NOT_STARTED | S | 2h | — | COURSE-P3-T13 |
| COURSE-P3-T16 | Playwright E2E verification | 6 | NOT_STARTED | M | 4h | — | COURSE-P3-T14, COURSE-P3-T15 |
| COURSE-P3-T17 | Phase 3 documentation & ledger sync | 3 | NOT_STARTED | S | 2h | — | COURSE-P3-T16 |
| COURSE-P3-T18 | Full canonical Phase 3 verification | 5 | NOT_STARTED | S | 2h | — | COURSE-P3-T14, COURSE-P3-T15, COURSE-P3-T16, COURSE-P3-T17 |

Total points: 111

## Task-specific acceptance criteria

### COURSE-P3-T01 — Resource persistence decision (first-class Resource vs LearningItem metadata)
<!-- TASK:COURSE-P3-T01:ACCEPTANCE:START -->
- [ ] Decision record compares Option A (first-class Resource table) vs Option B (typed LearningItem metadata) with repository evidence.
- [ ] Chosen option preserves `parent_id` authority, `ON DELETE RESTRICT`, and no `child_node_ids[]`.
- [ ] Decision is recorded in `DECISIONS.md` and/or an ADR before T02 begins.
- [ ] Evidence under `docs/features/course-system/evidence/course-p3-t01/`.
- [ ] No migration, HTTP, MinIO wiring, or UI work is performed.
<!-- TASK:COURSE-P3-T01:ACCEPTANCE:END -->

### COURSE-P3-T02 — Resource attachment schema & migration
<!-- TASK:COURSE-P3-T02:ACCEPTANCE:START -->
- [ ] Additive sql-migrate up/down pair implements the T01 decision only.
- [ ] Resources attach to LearningItems without introducing `child_node_ids[]` or a depth column.
- [ ] FK rules use `ON DELETE RESTRICT`; Course scope is preserved.
- [ ] Isolated PostgreSQL apply, inspect, rollback, and re-apply succeed.
- [ ] Evidence under `docs/features/course-system/evidence/course-p3-t02/`.
- [ ] No HTTP controllers, frontend, quiz engine, or Phase 4+ work is performed.
<!-- TASK:COURSE-P3-T02:ACCEPTANCE:END -->

### COURSE-P3-T03 — Resource Go model & repository
<!-- TASK:COURSE-P3-T03:ACCEPTANCE:START -->
- [ ] Repository Create/Get/List/Update/Delete follow goqu, UUID, and semantic-error conventions.
- [ ] List ordering is deterministic; empty valid scopes return non-nil empty slices.
- [ ] CourseNode `status` and LearningItem `publish_state` remain distinct in model code.
- [ ] Focused repository tests plus Course/CourseNode/LearningItem regression pass.
- [ ] Evidence under `docs/features/course-system/evidence/course-p3-t03/`; Database Migration: NONE unless T02 revision approved.
- [ ] No MinIO wiring, HTTP, or UI work is performed.
<!-- TASK:COURSE-P3-T03:ACCEPTANCE:END -->

### COURSE-P3-T04 — MinIO upload lifecycle & object keys
<!-- TASK:COURSE-P3-T04:ACCEPTANCE:START -->
- [ ] Uploads use existing MinIO/S3-compatible service; object keys are Course-scoped and non-guessable.
- [ ] Upload lifecycle covers initiate, complete, and abort with server-side MIME and size validation.
- [ ] Secrets remain outside Git; example config uses placeholders only.
- [ ] Focused storage integration tests pass with logged commands.
- [ ] Evidence under `docs/features/course-system/evidence/course-p3-t04/`.
- [ ] No product `MAX_*_DEPTH` or hierarchy mutation changes are introduced.
<!-- TASK:COURSE-P3-T04:ACCEPTANCE:END -->

### COURSE-P3-T05 — Signed URL delivery & streaming policy
<!-- TASK:COURSE-P3-T05:ACCEPTANCE:START -->
- [ ] Learner/admin delivery uses time-limited signed URLs; permanent public URLs are not the default.
- [ ] Streaming policy (inline vs download, range requests where applicable) is server-authoritative.
- [ ] Draft LearningItems (`publish_state=DRAFT`) do not receive learner signed URLs.
- [ ] Focused tests pass; evidence under `docs/features/course-system/evidence/course-p3-t05/`.
- [ ] No CDN productization or Phase 8 analytics is required.
<!-- TASK:COURSE-P3-T05:ACCEPTANCE:END -->

### COURSE-P3-T06 — Rich text & Markdown native content blocks
<!-- TASK:COURSE-P3-T06:ACCEPTANCE:START -->
- [ ] Rich text and Markdown block types extend the Phase 2 Information Block envelope without breaking version validation.
- [ ] Invalid payloads are rejected before SQL; canonical remashal preserves valid `data` objects.
- [ ] Content attaches under LearningItems on CourseNodes at any depth (D-004); no depth rejection.
- [ ] Focused metadata/repository tests pass; evidence under `docs/features/course-system/evidence/course-p3-t06/`.
- [ ] No PDF/video viewers, quiz sessions, or frontend work is performed.
<!-- TASK:COURSE-P3-T06:ACCEPTANCE:END -->

### COURSE-P3-T07 — PDF, document, image, audio & video handling
<!-- TASK:COURSE-P3-T07:ACCEPTANCE:START -->
- [ ] PDF, document, image, audio, and video types have typed metadata and MIME allowlists.
- [ ] Signed delivery works for each media class without leaking draft LearningItems.
- [ ] Deep-tree placement succeeds without depth rejection or MAX_DEPTH constant.
- [ ] Focused tests pass; evidence under `docs/features/course-system/evidence/course-p3-t07/`.
- [ ] No assessment engine or progress aggregation work is performed.
<!-- TASK:COURSE-P3-T07:ACCEPTANCE:END -->

### COURSE-P3-T08 — External link URL safety & downloadables
<!-- TASK:COURSE-P3-T08:ACCEPTANCE:START -->
- [ ] External URLs are validated server-side (scheme allowlist, host checks; reject javascript/data).
- [ ] Downloadable attachment policy is enforced in the API layer, not only UI.
- [ ] Malformed, empty, and unsafe URLs map to semantic errors before persistence.
- [ ] Focused tests pass; evidence under `docs/features/course-system/evidence/course-p3-t08/`.
- [ ] No frontend-only URL gating is treated as sufficient.
<!-- TASK:COURSE-P3-T08:ACCEPTANCE:END -->

### COURSE-P3-T09 — Multilingual metadata & specialized attachments
<!-- TASK:COURSE-P3-T09:ACCEPTANCE:START -->
- [ ] English and Hindi explanation fields (or locale-keyed metadata) are validated and stored.
- [ ] Question booklet, video solution, and lecture recording types are modelled with typed metadata.
- [ ] Attachments bind to LearningItems at any CourseNode depth; assessment launch stays Phase 4.
- [ ] Focused tests pass; evidence under `docs/features/course-system/evidence/course-p3-t09/`.
- [ ] No machine-translation pipeline is required.
<!-- TASK:COURSE-P3-T09:ACCEPTANCE:END -->

### COURSE-P3-T10 — Resource replacement lifecycle
<!-- TASK:COURSE-P3-T10:ACCEPTANCE:START -->
- [ ] Admin can replace an existing resource object while preserving LearningItem identity and position.
- [ ] Old MinIO objects are orphaned or tombstoned per documented policy.
- [ ] Replacement is transactional where metadata and object key must stay consistent.
- [ ] Focused tests pass; evidence under `docs/features/course-system/evidence/course-p3-t10/`.
- [ ] No progress recalculation (Phase 6) is required.
<!-- TASK:COURSE-P3-T10:ACCEPTANCE:END -->

### COURSE-P3-T11 — Admin resource API endpoints
<!-- TASK:COURSE-P3-T11:ACCEPTANCE:START -->
- [ ] Admin resource routes are registered under authenticated admin course paths.
- [ ] Controllers are transport-only; validation and storage rules remain repository/service-owned.
- [ ] Responses never treat computed `children[]` as writable structural input.
- [ ] Controller tests cover auth, validation, and happy paths.
- [ ] Evidence under `docs/features/course-system/evidence/course-p3-t11/`.
<!-- TASK:COURSE-P3-T11:ACCEPTANCE:END -->

### COURSE-P3-T12 — Learner resource delivery APIs
<!-- TASK:COURSE-P3-T12:ACCEPTANCE:START -->
- [ ] Authenticated learner delivery endpoints return only published LearningItems (`publish_state=PUBLISHED`).
- [ ] Draft discovery is prevented (404 / not-found semantics consistent with COURSE-P2-T08).
- [ ] Signed URLs and link payloads are generated server-side.
- [ ] Controller + model tests pass; evidence under `docs/features/course-system/evidence/course-p3-t12/`.
<!-- TASK:COURSE-P3-T12:ACCEPTANCE:END -->

### COURSE-P3-T13 — Resource viewer UI (admin + learner)
<!-- TASK:COURSE-P3-T13:ACCEPTANCE:START -->
- [ ] Admin can attach and preview supported resource types in builder surfaces.
- [ ] Learner UI renders rich text, Markdown, documents, media, external links, and multilingual explanations.
- [ ] UI does not invent hierarchy depth limits or persist `child_node_ids[]`.
- [ ] Frontend lint/tests for changed files pass.
- [ ] Evidence under `docs/features/course-system/evidence/course-p3-t13/`.
<!-- TASK:COURSE-P3-T13:ACCEPTANCE:END -->

### COURSE-P3-T14 — Backend unit & integration tests
<!-- TASK:COURSE-P3-T14:ACCEPTANCE:START -->
- [ ] `go test` coverage includes resource repository, URL safety, upload/replace lifecycle, and HTTP mapping.
- [ ] Regression on Course, CourseNode, and LearningItem packages passes.
- [ ] Multi-level `parent_id` chains do not break resource attachment.
- [ ] Evidence under `docs/features/course-system/evidence/course-p3-t14/`.
- [ ] No production or NUC deployment is performed.
<!-- TASK:COURSE-P3-T14:ACCEPTANCE:END -->

### COURSE-P3-T15 — Frontend unit tests
<!-- TASK:COURSE-P3-T15:ACCEPTANCE:START -->
- [ ] Vitest coverage exists for resource viewer/helpers introduced in this phase.
- [ ] `npm run lint` and `npm test -- --run` pass for changed app scope.
- [ ] Evidence under `docs/features/course-system/evidence/course-p3-t15/`.
- [ ] No dummy success stubs replace real assertions.
<!-- TASK:COURSE-P3-T15:ACCEPTANCE:END -->

### COURSE-P3-T16 — Playwright E2E verification
<!-- TASK:COURSE-P3-T16:ACCEPTANCE:START -->
- [ ] Playwright covers admin upload/attach + learner view for document, media, and external link paths.
- [ ] Flow exercises a nested CourseNode (not only root) to prove D-004 any-depth placement.
- [ ] Screenshots/logs stored; `npx playwright test` exit code recorded.
- [ ] Evidence under `docs/features/course-system/evidence/course-p3-t16/`.
<!-- TASK:COURSE-P3-T16:ACCEPTANCE:END -->

### COURSE-P3-T17 — Phase 3 documentation & ledger sync
<!-- TASK:COURSE-P3-T17:ACCEPTANCE:START -->
- [ ] Phase docs, HANDOFF, CHANGELOG, and architecture boundary map mention Phase 3 ownership.
- [ ] Architecture freeze invariants are restated (parent_id authoritative; no child_node_ids[]; no MAX_DEPTH; status vs publish_state).
- [ ] `npm run course-system:status:sync` then `npm run course-system:status:check` succeed; second sync is a no-op.
- [ ] Evidence under `docs/features/course-system/evidence/course-p3-t17/`.
- [ ] No unverified tasks are marked VERIFIED.
<!-- TASK:COURSE-P3-T17:ACCEPTANCE:END -->

### COURSE-P3-T18 — Full canonical Phase 3 verification
<!-- TASK:COURSE-P3-T18:ACCEPTANCE:START -->
- [ ] All Phase 3 verification commands pass with logged exit codes and timestamps.
- [ ] Phase acceptance criteria checklist is evaluated; failures are recorded explicitly.
- [ ] Evidence under `docs/features/course-system/evidence/course-p3-t18/`.
- [ ] No tasks are marked VERIFIED without canonical evidence.
<!-- TASK:COURSE-P3-T18:ACCEPTANCE:END -->

## Phase acceptance criteria

- [ ] COURSE-P3-T01 decision is ACCEPTED and recorded before persistence work.
- [ ] Resource/attachment persistence exists with additive up/down migrations.
- [ ] MinIO upload, replacement lifecycle, and signed delivery work for supported media classes.
- [ ] Rich text, Markdown, document, image, audio, video, external link, and downloadable types are supported.
- [ ] Multilingual explanations, question booklets, video solutions, and lecture recordings are modelled.
- [ ] URL safety and download rules are enforced server-side.
- [ ] Admin and learner APIs respect LearningItem `publish_state` and CourseNode `status`.
- [ ] Resources attach under CourseNodes at any logical depth (D-004); hierarchy remains parent_id-authoritative (D-005 / ADR-023).
- [ ] FK rules use `ON DELETE RESTRICT` unless an ADR documents otherwise.
- [ ] Backend, frontend, Playwright, Compose verify profile, and ledger sync/check pass.
- [ ] Documentation matches implementation.

## Verification commands

**Task-focused** (run for the task's layer):

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

**Full phase verify** (COURSE-P3-T18):

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
`docs/features/course-system/evidence/course-p3-tXX/`:

* README summarizing scope and non-goals.
* Command logs with exit codes and verification classification (task / integration / full phase).
* Hash / sync artifacts when ledger tooling is run.
* Migration prove-out notes when a migration is in scope.
* Explicit statement of Database Migration: NONE when applicable.
* Confirmation that hierarchy invariants were not violated.

## Security requirements

* Signed URLs are time-limited; draft items never receive learner delivery URLs.
* External URL validation rejects unsafe schemes; download policy is server-enforced.
* Admin resource mutations require authenticated admin authorization (current allowlist until Phase 8).
* MinIO credentials and bucket policies remain outside Git.
* Resource metadata must not leak cross-Course object keys or signed URLs.
* All authorization checks remain server-side; Nuxt client is not source of truth.

## Performance requirements

* Upload size limits are operational constraints, not product depth caps.
* Signed URL generation must not block the request thread beyond documented timeouts.
* List endpoints return deterministic ordering without loading entire Course trees.
* Large media delivery uses streaming/range where applicable; document payload thresholds in evidence.
* Deep-tree attachment (≥25 levels) must succeed without depth rejection (D-004).

## Accessibility / mobile

* Learner viewers expose accessible labels for media controls and document links.
* Admin upload flows are keyboard-operable where the platform pattern supports it.
* Mobile layouts render readable resource content without horizontal overflow on standard viewports.
* Full WCAG audit defers to Phase 8; Phase 3 must not introduce known blocking a11y regressions.

## Risks

| Risk | Mitigation |
|---|---|
| Treating MinIO object trees as the Course hierarchy | CourseNode `parent_id` remains sole structural edge |
| Client-only URL or download gating bypassed by direct API calls | Enforce policy in API layer |
| Confusing CourseNode `status` with LearningItem `publish_state` | Keep names distinct in APIs and docs |
| Accidental introduction of depth caps "for performance" | Operational limits ≠ product MAX_*_DEPTH (D-004) |
| Starting migrations before T01 decision | Design gate blocks T02+ until ACCEPTED |
| CASCADE delete on resource FKs | Mandate `ON DELETE RESTRICT` in T01/T02 |

## Known limitations

* Full CDN / adaptive streaming productization is not required.
* Enrollment commerce may remain deferred; learner auth may follow Phase 2 pattern until authorized.
* Machine translation for multilingual explanations is out of scope.
* Operational payload/rate/page limits are not product `MAX_*_DEPTH`.
* Course admin authorization model replacement defers to Phase 8 T01 gate.

## Exit criteria

Phase 3 may be marked complete only when:

1. All checklist tasks are `VERIFIED` with canonical evidence.
2. Declared total points (**111**) match the ledger sum and status tooling.
3. Full phase verification commands pass for changed surfaces.
4. Architecture freeze invariants hold: no `child_node_ids[]` authority, no `MAX_*_DEPTH`, status/publish_state naming correct, `ON DELETE RESTRICT` preserved.
5. COURSE-P3-T01 decision is ACCEPTED in `DECISIONS.md` and/or ADR.
6. HANDOFF records the next safe action (typically Phase 4 start or an approved parallel exception).
