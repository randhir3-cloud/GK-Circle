# Phase 2 — Learning Items and Information Blocks

* **Status**: IN_PROGRESS
* **Weight**: 15%
* **Phase owner**: Unassigned
* **Started at**: 2026-07-26T08:30:00+05:30
* **Verified at**: —

## Objective

Compose ordered LearningItems and information blocks on every CourseNode at any
depth — CRUD, metadata, publication, node-local reorder/move, enrollment gate for
item delivery, runtime visibility, node-local previous/next chain, item composition
UI, and learner item shell — without owning course-wide sequence or unlock graphs.

## Architectural outcome

* LearningItems attach to any CourseNode; nesting depth is irrelevant to attachment (D-004).
* Hierarchy remains `parent_id`-authoritative (D-005); items never store child-node arrays.
* CourseNode `status` ≠ LearningItem `publish_state`.
* Phase 2 previous/next is **node-local only**; Phase 5 owns course-wide sequence and Continue Learning.

## Current verified baseline

Verified through COURSE-P2-T01 … COURSE-P2-T08, COURSE-P2-T10, COURSE-P2-T11, COURSE-P2-T13, COURSE-P2-T20, COURSE-P2-T21, and COURSE-P2-T24:

* `learning_items` persistence with CourseNode-scoped positions.
* Information-block metadata, placeholders, visibility **write-time validation**.
* Admin LearningItem CRUD HTTP.
* `publish_state` (`DRAFT`|`PUBLISHED`).
* Authenticated learner published-only GET list/get (Kratos-only; enrollment deferred).
* Admin LearningItem sibling reorder within a node (canonical `0..n-1`).
* Admin LearningItem move across nodes within the same Course (subset move + dual-node compaction).
* Repository-owned learner runtime visibility projection (HIDDEN/INSTRUCTOR omitted; PREMIUM omitted without premium source).

**Not proven:** enrollment gate,
node-local chain APIs, composition UI, learner renderer shell.

## In scope

* Item move (same Course), enrollment gate for learner items, runtime visibility.
* Node-local previous/next and learning-chain projection.
* Admin item composition UI; learner item rendering shell.
* Tests, Playwright, documentation, phase verification.

## Out of scope

* Course-wide sequence, unlock-aware prev/next, Continue Learning (Phase 5).
* Resource upload pipelines (Phase 3); assessment engines (Phase 4).
* Integrated studio (Phase 7); product max-depth / `child_node_ids[]`.

## Dependencies

* Phase 1 CourseNode persistence and admin node APIs (`COURSE-P1-T03` through `COURSE-P1-T10` verified).
* D-003 parallel-phase authorization; D-004; D-005; ADR-023.

## Phase boundaries

| Owns | Does not own |
|---|---|
| LearningItem composition, publish_state, node-local chain, item enrollment gate, runtime visibility, item UIs | Structural tree APIs, course-wide sequence/prereqs, resources, assessments, progress, templates |

## Execution sequence

Verified baseline → reorder/move repository+API → enrollment → runtime visibility →
node-local chain → composition UI → learner shell → tests → Playwright → docs → full verify.

## Completion calculation

$$\text{Phase Completion} = \frac{\text{Verified Task Points}}{\text{Total Task Points}} \times 100$$

## Parallel-phase note

D-003 authorizes Phase 2 while Phase 1 UI remains open. `COURSE-P1-T11` remains available.

## Task checklist

| ID | Task | Points | Status | Difficulty | Est. Time | Evidence | Dependencies |
|---|---|---:|---|---|---|---|---|
| COURSE-P2-T01 | LearningItem database schema & migration | 5 | VERIFIED | S | 2h | `docs/features/course-system/evidence/course-p2-t01/README.md` | COURSE-P1-T03 |
| COURSE-P2-T02 | LearningItem Go model & CRUD queries | 7 | VERIFIED | M | 4h | `docs/features/course-system/evidence/course-p2-t02/README.md` | COURSE-P2-T01 |
| COURSE-P2-T03 | Information block content metadata structure | 6 | VERIFIED | M | 4h | `docs/features/course-system/evidence/course-p2-t03/README.md` | COURSE-P2-T02 |
| COURSE-P2-T04 | Creator placeholders metadata structure | 6 | VERIFIED | M | 4h | `docs/features/course-system/evidence/course-p2-t04/README.md` | COURSE-P2-T03 |
| COURSE-P2-T05 | Visibility filter server-side validation | 6 | VERIFIED | M | 4h | `docs/features/course-system/evidence/course-p2-t05/README.md` | COURSE-P2-T03 |
| COURSE-P2-T06 | Admin learning item API endpoints | 8 | VERIFIED | M | 4h | `docs/features/course-system/evidence/course-p2-t06/README.md` | COURSE-P2-T02, COURSE-P2-T03, COURSE-P2-T04, COURSE-P2-T05 |
| COURSE-P2-T07 | LearningItem publish state | 8 | VERIFIED | M | 4h | `docs/features/course-system/evidence/course-p2-t07/README.md` | COURSE-P2-T06 |
| COURSE-P2-T08 | Learner LearningItem APIs | 8 | VERIFIED | M | 4h | `docs/features/course-system/evidence/course-p2-t08/README.md` | COURSE-P2-T07 |
| COURSE-P2-T10 | Reorder LearningItems within a node | 5 | VERIFIED | M | 4h | `docs/features/course-system/evidence/course-p2-t10/README.md` | COURSE-P2-T02, COURSE-P2-T06 |
| COURSE-P2-T11 | Move LearningItem across nodes (same Course) | 5 | VERIFIED | M | 4h | `docs/features/course-system/evidence/course-p2-t11/README.md` | COURSE-P2-T10 |
| COURSE-P2-T20 | Deep-node LearningItem ownership tests | 5 | VERIFIED | M | 4h | `docs/features/course-system/evidence/course-p2-t20/README.md` | COURSE-P2-T02, COURSE-P2-T08 |
| COURSE-P2-T21 | Previous/Next item resolution (repository, node-local) | 5 | VERIFIED | M | 4h | `docs/features/course-system/evidence/course-p2-t21/README.md` | COURSE-P2-T02 |
| COURSE-P2-T24 | Draft filtering regression pack | 4 | VERIFIED | S | 2h | `docs/features/course-system/evidence/course-p2-t24/README.md` | COURSE-P2-T07, COURSE-P2-T08 |
| COURSE-P2-T12 | Enrollment gate for learner item APIs | 6 | VERIFIED | M | 4h | `docs/features/course-system/evidence/course-p2-t12/README.md` | COURSE-P2-T08 |
| COURSE-P2-T13 | Runtime visibility evaluation | 6 | VERIFIED | M | 4h | `docs/features/course-system/evidence/course-p2-t13/README.md` | COURSE-P2-T05, COURSE-P2-T08 |
| COURSE-P2-T22 | Previous/Next learner API (node-local) | 4 | VERIFIED | S | 2h | `docs/features/course-system/evidence/course-p2-t22/README.md` | COURSE-P2-T21, COURSE-P2-T08 |
| COURSE-P2-T23 | Node-local learning-chain projection | 5 | VERIFIED | M | 4h | `docs/features/course-system/evidence/course-p2-t23/README.md` | COURSE-P2-T21 |
| COURSE-P2-T25 | Publish-filter controller contract tests | 4 | VERIFIED | S | 2h | `docs/features/course-system/evidence/course-p2-t25/README.md` | COURSE-P2-T08 |
| COURSE-P2-T09 | Admin item composition UI | 8 | VERIFIED | L | 6h | `docs/features/course-system/evidence/course-p2-t09/README.md` | COURSE-P2-T06 |
| COURSE-P2-T14 | Learner item rendering shell UI | 8 | VERIFIED | L | 6h | `docs/features/course-system/evidence/course-p2-t14/README.md` | COURSE-P2-T08, COURSE-P2-T13 |
| COURSE-P2-T15 | Backend LearningItem test suite | 8 | VERIFIED | M | 4h | `docs/features/course-system/evidence/course-p2-t15/README.md` | COURSE-P2-T10, COURSE-P2-T11, COURSE-P2-T12, COURSE-P2-T13, COURSE-P2-T20, COURSE-P2-T21, COURSE-P2-T24 |
| COURSE-P2-T16 | Frontend LearningItem tests | 4 | VERIFIED | S | 2h | `docs/features/course-system/evidence/course-p2-t16/README.md` | COURSE-P2-T09, COURSE-P2-T14 |
| COURSE-P2-T17 | Playwright LearningItem E2E | 6 | VERIFIED | M | 4h | `docs/features/course-system/evidence/course-p2-t17/README.md` | COURSE-P2-T09, COURSE-P2-T14 |
| COURSE-P2-T26 | LearningItem API examples documentation | 3 | VERIFIED | S | 2h | `docs/features/course-system/evidence/course-p2-t26/README.md` | COURSE-P2-T06, COURSE-P2-T08 |
| COURSE-P2-T18 | Phase 2 documentation update | 3 | VERIFIED | S | 2h | `docs/features/course-system/evidence/course-p2-t18/README.md` | COURSE-P2-T26 |
| COURSE-P2-T19 | Full canonical Phase 2 verification | 5 | VERIFIED | S | 2h | `docs/features/course-system/evidence/course-p2-t19/README.md` | COURSE-P2-T15, COURSE-P2-T16, COURSE-P2-T17, COURSE-P2-T18, COURSE-P2-T22, COURSE-P2-T23, COURSE-P2-T25 |

Total points: 148

## Task-specific acceptance criteria

### COURSE-P2-T01 — LearningItem database schema & migration
<!-- TASK:COURSE-P2-T01:ACCEPTANCE:START -->
- [x] An additive sql-migrate up/down pair creates and removes only the `learning_items` table and its indexes/constraints.
- [x] Composite FK `(course_id, course_node_id)` references `course_nodes (course_id, id)` with `ON DELETE RESTRICT`.
- [x] Unique `(course_node_id, position)`, non-negative position, nonblank title, and typed `item_type` checks are enforced.
- [x] Isolated local PostgreSQL apply, inspect, rollback, and re-apply succeed without changing Course/CourseNode data.
- [x] Canonical evidence is complete under `docs/features/course-system/evidence/course-p2-t01/`; HTTP APIs: No.
- [x] No controllers, routes, DTOs, frontend, or Phase 1 UI work is performed.
<!-- TASK:COURSE-P2-T01:ACCEPTANCE:END -->

### COURSE-P2-T02 — LearningItem Go model & CRUD queries
<!-- TASK:COURSE-P2-T02:ACCEPTANCE:START -->
- [x] CourseNode-scoped Create/Get/List/Update/Delete repository methods follow goqu, UUID, and semantic-error conventions.
- [x] Create is append-only (`MAX(position)+1` or `0`) inside one transaction; unique conflicts map to `ErrLearningItemConflict`.
- [x] List returns `ORDER BY position ASC, id ASC` and a non-nil empty slice for an existing empty node; missing node is distinct.
- [x] Update is presence-aware for title/type/description/metadata only; rejects empty updates and metadata null.
- [x] Focused repository tests plus Course/CourseNode regression pass; evidence under `docs/features/course-system/evidence/course-p2-t02/`; HTTP APIs: No.
- [x] No controllers, routes, DTOs, reorder APIs, frontend, or Phase 1 UI work is performed.
<!-- TASK:COURSE-P2-T02:ACCEPTANCE:END -->

### COURSE-P2-T03 — Information block content metadata structure
<!-- TASK:COURSE-P2-T03:ACCEPTANCE:START -->
- [x] Structural Information Block metadata envelope is defined (`version`, `blocks[]` with `id`/`type`/`data`) and stored in the existing LearningItem JSONB column.
- [x] Validation enforces `version >= 1`, array `blocks` (empty allowed), unique non-blank block IDs, typed block enums, and object-only `data` (arrays/scalars/null/missing rejected).
- [x] Canonical remashal preserves original valid `data` object content and does not coerce non-empty objects to `{}`.
- [x] Create/Update LearningItem validate metadata before persistence; no HTTP, rendering, migration, or per-type content-field contracts.
- [x] Focused metadata/repository tests plus Course/CourseNode regression pass; evidence under `docs/features/course-system/evidence/course-p2-t03/`.
- [x] No controllers, routes, DTOs, frontend, placeholders (P2-T04), or Phase 1 UI work is performed.
<!-- TASK:COURSE-P2-T03:ACCEPTANCE:END -->

### COURSE-P2-T04 — Creator placeholders metadata structure
<!-- TASK:COURSE-P2-T04:ACCEPTANCE:START -->
- [x] Placeholder syntax `{{identifier}}` is validated inside LearningItem block `data` string leaves (including nested objects/arrays).
- [x] Identifier rules enforce first letter, ASCII letters/digits/underscore, max length 64; empty/malformed delimiters map to semantic placeholder errors.
- [x] Zero, repeated, and multiple placeholders are allowed; meaning/resolution are not validated.
- [x] Original `data` bytes and placeholder text are preserved; Create/Update reject invalid placeholders before SQL.
- [x] Focused placeholder/repository tests plus regression pass; evidence under `docs/features/course-system/evidence/course-p2-t04/`.
- [x] No HTTP, rendering, AI replacement, migration, frontend, or Phase 1 UI work is performed.
<!-- TASK:COURSE-P2-T04:ACCEPTANCE:END -->

### COURSE-P2-T05 — Visibility filter server-side validation
<!-- TASK:COURSE-P2-T05:ACCEPTANCE:START -->
- [x] Optional block `visibility` metadata is validated with typed modes ALL/AUTHENTICATED/PREMIUM/INSTRUCTOR/HIDDEN.
- [x] Omitted visibility defaults to `{"mode":"ALL"}` in the persisted canonical document; null/non-object/missing mode/unknown mode are rejected.
- [x] Visibility validation runs after data/placeholder checks; no runtime evaluation, permissions, HTTP, or rendering.
- [x] Create/Update reject invalid visibility before SQL; focused tests plus regression pass.
- [x] Canonical evidence under `docs/features/course-system/evidence/course-p2-t05/`; Database Migration: NONE.
- [x] No controllers, routes, DTOs, frontend, or Phase 1 UI work is performed.
<!-- TASK:COURSE-P2-T05:ACCEPTANCE:END -->

### COURSE-P2-T06 — Admin learning item API endpoints
<!-- TASK:COURSE-P2-T06:ACCEPTANCE:START -->
- [x] Admin LearningItem CRUD is exposed under `/api/v1/admin/courses/:course_id/nodes/:node_id/learning-items`.
- [x] Controllers are transport-only; metadata validation remains repository-owned (P2-T03–T05).
- [x] Authentication reuses Kratos + quiz-admin allowlist; owner is never client-supplied.
- [x] Controller tests cover create/read/list/update/delete, auth, and metadata error mapping.
- [x] Canonical evidence under `docs/features/course-system/evidence/course-p2-t06/`; Database Migration: NONE.
- [x] No frontend, rendering, placeholder resolution, or runtime visibility evaluation.
<!-- TASK:COURSE-P2-T06:ACCEPTANCE:END -->

### COURSE-P2-T07 — LearningItem publish state
<!-- TASK:COURSE-P2-T07:ACCEPTANCE:START -->
- [x] Additive migration adds `learning_items.publish_state` TEXT NOT NULL DEFAULT `DRAFT` with constraint `learning_items_publish_state_check` (`DRAFT`|`PUBLISHED`).
- [x] Repository defaults create to DRAFT; update may set publish_state; invalid/empty/lowercase rejected via `ErrLearningItemPublishStateInvalid`.
- [x] Admin create/update DTOs use `OptionalString`; pointer constructed only after exact-enum proof; omitted vs null remain distinct.
- [x] Responses include `publish_state`; no new `/publish` routes; no learner delivery.
- [x] Isolated PostgreSQL apply/inspect/rollback/reapply; controller/repository tests; evidence under `docs/features/course-system/evidence/course-p2-t07/`.
- [x] Learner LearningItem API endpoints remain explicitly deferred (displaced from T07, not cancelled).
<!-- TASK:COURSE-P2-T07:ACCEPTANCE:END -->

### COURSE-P2-T08 — Learner LearningItem APIs
<!-- TASK:COURSE-P2-T08:ACCEPTANCE:START -->
- [x] Authenticated read-only learner GET list/get under `/api/v1/learner/courses/:course_id/nodes/:node_id/learning-items`.
- [x] Auth is `KratosAuthenticated` only (no quiz-admin allowlist; no anonymous access); enrollment deferred.
- [x] Repository-owned publish filter via `GetPublishedLearningItemByID` / `ListPublishedLearningItemsByNode`; controllers never filter `publish_state`.
- [x] Private unexported helpers share SQL (`getLearningItem` / `listLearningItems` with `publishedOnly`); admin CRUD unchanged.
- [x] Missing or draft items map to `ErrLearningItemNotFound` (404; no draft discovery).
- [x] Controller + model tests; evidence under `docs/features/course-system/evidence/course-p2-t08/`; Database Migration: NONE.
- [x] Admin item-chain builder UI remains explicitly deferred (displaced from T08, not cancelled).
<!-- TASK:COURSE-P2-T08:ACCEPTANCE:END -->

### COURSE-P2-T09 — Admin item composition UI
<!-- TASK:COURSE-P2-T09:ACCEPTANCE:START -->
- [x] Admin UI lists, creates, edits, and deletes LearningItems on the selected CourseNode at any depth via admin APIs.
- [x] Item order display matches server `position`; no client-side hierarchy authority via `child_node_ids[]`.
- [x] Does not require unfinished Phase 1 tree editor; may embed in a simple course/node picker shell.
- [x] Evidence under `docs/features/course-system/evidence/course-p2-t09/`; Database Migration: NONE.
<!-- TASK:COURSE-P2-T09:ACCEPTANCE:END -->

### COURSE-P2-T10 — Reorder LearningItems within a node
<!-- TASK:COURSE-P2-T10:ACCEPTANCE:START -->
- [x] Transactional sibling reorder API validates ordered ID sets and writes canonical `0..n-1` positions.
- [x] Empty and already-canonical cases succeed without partial writes; focused tests pass.
- [x] Evidence under `docs/features/course-system/evidence/course-p2-t10/`; Database Migration: NONE.
- [x] Bidirectional exact-set validation: every existing sibling exactly once; no missing IDs; no extra IDs; no duplicates (`ErrLearningItemReorderMismatch` / duplicate).
- [x] Concurrent reorder requests on the same CourseNode serialize correctly; no duplicate positions can be committed; losing/conflicting transaction returns deterministic conflict/retry (`ErrLearningItemReorderConflict` → HTTP 409).
- [x] Idempotent: submitting the same canonical order repeatedly is a noop without unnecessary row rewrites.
- [x] Frozen response payload includes `course_node_id`, `learning_item_count`, `positions_updated`, and `noop`.
<!-- TASK:COURSE-P2-T10:ACCEPTANCE:END -->

### COURSE-P2-T11 — Move LearningItem across nodes (same Course)
<!-- TASK:COURSE-P2-T11:ACCEPTANCE:START -->
- [x] Move between CourseNodes in the same Course updates `course_node_id` and destination position transactionally.
- [x] Cross-Course / foreign-node IDs in `ordered_item_ids` map to `ErrLearningItemMoveMismatch` (HTTP 400) without existence leak; missing source/destination nodes map to not-found.
- [x] Evidence under `docs/features/course-system/evidence/course-p2-t11/` (including `transaction-timeline.md` and `move-algorithm.md`); Database Migration: NONE.
- [x] Lock order: course → source+dest nodes by ascending UUID → siblings in the same node-UUID order; six-step staging (temps → ownership → final source → final dest → verify → commit).
- [x] Empty `ordered_item_ids` is a successful noop after course/node validation with real locked sibling counts (not forced zeros).
- [x] `source == destination` → `ErrLearningItemMoveSameNode` (HTTP 400); unique/serialization conflicts → `ErrLearningItemMoveConflict` (HTTP 409).
- [x] Frozen response payload includes `source_node_id`, `destination_node_id`, `items_moved`, `source_item_count`, `destination_item_count`, and `noop`.
<!-- TASK:COURSE-P2-T11:ACCEPTANCE:END -->

### COURSE-P2-T12 — Enrollment gate for learner item APIs
<!-- TASK:COURSE-P2-T12:ACCEPTANCE:START -->
- [x] Learner LearningItem GET requires Course enrollment (or documented equivalent) server-side.
- [x] Unenrolled authenticated users cannot discover draft or published item payloads beyond documented denial.
- [x] Evidence under `docs/features/course-system/evidence/course-p2-t12/`.
<!-- TASK:COURSE-P2-T12:ACCEPTANCE:END -->

### COURSE-P2-T13 — Runtime visibility evaluation
<!-- TASK:COURSE-P2-T13:ACCEPTANCE:START -->
- [x] Learner delivery evaluates block `visibility` modes server-side; HIDDEN omitted; unauthorized modes filtered.
- [x] Controllers do not reimplement publish_state filtering; evaluation is repository/service-owned.
- [x] Evidence under `docs/features/course-system/evidence/course-p2-t13/`.
<!-- TASK:COURSE-P2-T13:ACCEPTANCE:END -->

### COURSE-P2-T14 — Learner item rendering shell UI
<!-- TASK:COURSE-P2-T14:ACCEPTANCE:START -->
- [x] Learner UI renders ordered published items for a node at any depth via learner APIs.
- [x] Draft items never appear; empty node shows empty state; unauthorized access denied.
- [x] Evidence under `docs/features/course-system/evidence/course-p2-t14/`.
<!-- TASK:COURSE-P2-T14:ACCEPTANCE:END -->

### COURSE-P2-T15 — Backend LearningItem test suite
<!-- TASK:COURSE-P2-T15:ACCEPTANCE:START -->
- [x] From `api/`: `go test ./...` and docker `api-verify` cover reorder/move/enrollment/visibility/chain regressions.
- [x] Evidence under `docs/features/course-system/evidence/course-p2-t15/`.
<!-- TASK:COURSE-P2-T15:ACCEPTANCE:END -->

### COURSE-P2-T16 — Frontend LearningItem tests
<!-- TASK:COURSE-P2-T16:ACCEPTANCE:START -->
- [x] From `app/`: `npm run lint` and `npm test -- --run` cover composition/renderer suites.
- [x] Evidence under `docs/features/course-system/evidence/course-p2-t16/`.
<!-- TASK:COURSE-P2-T16:ACCEPTANCE:END -->

### COURSE-P2-T17 — Playwright LearningItem E2E
<!-- TASK:COURSE-P2-T17:ACCEPTANCE:START -->
- [x] `npx playwright test` covers admin create item on deep node → publish → learner view.
- [x] Evidence under `docs/features/course-system/evidence/course-p2-t17/`.
<!-- TASK:COURSE-P2-T17:ACCEPTANCE:END -->

### COURSE-P2-T18 — Phase 2 documentation update
<!-- TASK:COURSE-P2-T18:ACCEPTANCE:START -->
- [x] Docs keep `publish_state` vs CourseNode `status` distinct; node-local vs course-wide navigation clarified.
- [x] Ledger sync/check pass after documentation updates.
<!-- TASK:COURSE-P2-T18:ACCEPTANCE:END -->

### COURSE-P2-T19 — Full canonical Phase 2 verification
<!-- TASK:COURSE-P2-T19:ACCEPTANCE:START -->
- [x] Full backend/frontend/Playwright/ledger gates pass for Phase 2 scope.
- [x] No production/NUC access; evidence index complete.
<!-- TASK:COURSE-P2-T19:ACCEPTANCE:END -->

### COURSE-P2-T20 — Deep-node LearningItem ownership tests
<!-- TASK:COURSE-P2-T20:ACCEPTANCE:START -->
- [x] LearningItems attach and list correctly on CourseNodes at ≥10 nesting levels.
- [x] Cross-Course node/item access returns not-found; Course scope never leaks.
- [x] Evidence under `docs/features/course-system/evidence/course-p2-t20/`.
<!-- TASK:COURSE-P2-T20:ACCEPTANCE:END -->

### COURSE-P2-T21 — Previous/Next item resolution (repository, node-local)
<!-- TASK:COURSE-P2-T21:ACCEPTANCE:START -->
- [x] Repository resolves previous/next LearningItem by position within the same CourseNode only.
- [x] Ends of chain return documented empty/null semantics without inventing cross-node navigation.
- [x] Evidence under `docs/features/course-system/evidence/course-p2-t21/`; Database Migration: NONE.
<!-- TASK:COURSE-P2-T21:ACCEPTANCE:END -->

### COURSE-P2-T22 — Previous/Next learner API (node-local)
<!-- TASK:COURSE-P2-T22:ACCEPTANCE:START -->
- [x] Authenticated learner endpoints expose previous/next published items for a node-local chain.
- [x] Draft items are skipped; missing/draft current item maps to not-found.
- [x] Evidence under `docs/features/course-system/evidence/course-p2-t22/`.
<!-- TASK:COURSE-P2-T22:ACCEPTANCE:END -->

### COURSE-P2-T23 — Node-local learning-chain projection
<!-- TASK:COURSE-P2-T23:ACCEPTANCE:START -->
- [x] Projection returns ordered published item IDs/titles for a node without storing child_node_ids.
- [x] Distinct from structural tree edges and Phase 5 prerequisite DAG / course-wide sequence.
- [x] Evidence under `docs/features/course-system/evidence/course-p2-t23/`.
<!-- TASK:COURSE-P2-T23:ACCEPTANCE:END -->

### COURSE-P2-T24 — Draft filtering regression pack
<!-- TASK:COURSE-P2-T24:ACCEPTANCE:START -->
- [x] Learner list/get never returns `publish_state=DRAFT` payloads or discovery signals beyond documented 404.
- [x] Admin APIs continue to return drafts for authorized admins.
- [x] Evidence under `docs/features/course-system/evidence/course-p2-t24/`.
<!-- TASK:COURSE-P2-T24:ACCEPTANCE:END -->

### COURSE-P2-T25 — Publish-filter controller contract tests
<!-- TASK:COURSE-P2-T25:ACCEPTANCE:START -->
- [x] Controller tests prove publish filtering remains repository-owned.
- [x] Auth matrix covers unauthenticated 401 and authenticated learner success paths.
- [x] Evidence under `docs/features/course-system/evidence/course-p2-t25/`.
<!-- TASK:COURSE-P2-T25:ACCEPTANCE:END -->

### COURSE-P2-T26 — LearningItem API examples documentation
<!-- TASK:COURSE-P2-T26:ACCEPTANCE:START -->
- [x] Docs include snake_case request/response examples for admin CRUD and learner published reads.
- [x] Docs restate items-at-any-depth and status vs publish_state naming.
- [x] Evidence under `docs/features/course-system/evidence/course-p2-t26/`.
<!-- TASK:COURSE-P2-T26:ACCEPTANCE:END -->

## Phase acceptance criteria

- [x] Learning items attach to CourseNodes; ordered; admin and learner APIs exist.
- [x] Information blocks and placeholders modelled; visibility validated on write.
- [ ] Reorder/move, enrollment, runtime visibility, and node-local chain verified.
- [ ] Admin composition UI and learner shell ship.
- [ ] Backend and frontend verification pass.
- [ ] Documentation current with D-004/D-005 and Phase 5 boundary.

## Verification commands

```bash
# cwd: api/
go test ./...
# repo root
docker compose --profile verify run --rm api-verify
# cwd: app/
npm run lint
npm test -- --run
npx playwright test
# repo root
npm run course-system:status:sync
npm run course-system:status:check
```

## Evidence requirements

* VERIFIED evidence paths preserved under `docs/features/course-system/evidence/course-p2-tNN/`.
* Do not fabricate evidence; do not mark VERIFIED without proof.

## Security requirements

* Learner endpoints must not leak draft existence beyond documented not-found policy.
* Enrollment and entitlements are server-side (T12).
* Admin authorization remains server-side.

## Performance requirements

* Deep-node attachment tests (≥10 levels) without hierarchy-depth validation.
* Operational pagination allowed; not product MAX_DEPTH.

## Accessibility and mobile requirements

* Composition/renderer usable on narrow viewports; keyboard access for item reorder controls when present.

## Risks and mitigations

| Risk | Mitigation |
|---|---|
| Confusing status vs publish_state | Keep names distinct |
| Course-wide nav sneaking into P2 | Phase boundaries; P5 owns sequence |
| T09 blocked on P1 UI | T09 depends only on admin item APIs |

## Known limitations

* Enrollment deferred until T12; runtime visibility until T13.
* Node-local prev/next only until Phase 5.
* Admin quiz-admin allowlist remains temporary (Phase 8 auth gate).

## Exit criteria

* All Phase 2 tasks `VERIFIED` with evidence; denominator **148** points.
* Phase acceptance criteria checked; ledger sync/check pass.
* No production/NUC destructive operations.
