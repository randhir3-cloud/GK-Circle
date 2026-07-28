# Phase 1 — Unlimited Recursive Course Tree Foundation

* **Status**: IN_PROGRESS
* **Weight**: 15%
* **Phase owner**: Unassigned
* **Started at**: 2026-07-25T10:05:00+05:30
* **Verified at**: —

## Objective

Build the Course aggregate and unlimited-depth recursive CourseNode structure
(`parent_id` authoritative; children derived) with transactional structural
mutations, structural read APIs (tree/children/subtree/ancestors/breadcrumbs),
publish-safe learner outline, minimum structural tree editor, and deep/large-tree
verification — without LearningItems, resources, assessments, sequence, or progress.

## Architectural outcome

```text
Course
└── CourseNode
    ├── parent_id → CourseNode | NULL   # AUTHORITATIVE
    ├── children[] → computed projection
    └── (LearningItems owned by Phase 2)
```

* D-004 / D-005 and ADR-023 govern hierarchy.
* No product `MAX_*_DEPTH`; no authoritative `child_node_ids[]`.
* CourseNode lifecycle field is `status`; UUID materialised path; no depth column.
* Phase 1 builder is the **minimum structural editor**; Phase 7 owns the integrated studio.

## Current verified baseline

Verified through COURSE-P1-T01 … COURSE-P1-T10 (including T01A, T02B):

* Course root persistence and admin Course APIs.
* CourseNode persistence with Course-scoped `parent_id`, path, position, `status`.
* Hierarchy reads: roots, direct children, full tree with computed `children[]`.
* Transactional move, sibling reorder, subtree deletion.
* Admin create/read and mutation HTTP under `/api/v1/admin/courses/:course_id/nodes`.

**Not proven by VERIFIED tasks (follow-ups only):** dedicated subtree GET; ancestor/breadcrumb HTTP;
learner structural outline; minimum structural UI; 25/50-level stress; lazy pagination;
tree integrity pack beyond existing mutation tests.

## In scope

* Structural Course/CourseNode APIs and repository gaps above.
* Minimum structural tree editor (list/create courses; add root/child; rename where API exists;
  expand/collapse; move; reorder; delete; lazy children; accessible non-drag move).
* Deep-tree and large-tree structural correctness/performance measurement.
* Structural authorization matrices for admin mutations.

## Out of scope

* LearningItems, resources, assessments, course-wide sequence, prerequisites, progress.
* Templates and integrated authoring studio (Phase 7).
* Product max-depth rules; authoritative child arrays; depth column without ADR.

## Dependencies

* ADR-023; module decisions D-001, D-002, D-004, D-005.
* Architecture: `../architecture/current.md`.

## Phase boundaries

| Owns | Does not own |
|---|---|
| Structure, path, structural mutations, structural reads, learner outline, minimum structural editor | Item composition, resources, assessments, course-wide nav/unlock, progress, templates/studio |

Breadcrumbs/ancestors are defined here. Phase 5 may consume them; it must not redefine those APIs.

## Execution sequence

1. Preserve verified baseline (done).
2. Structural read gaps (ancestors, breadcrumbs, subtree, lazy children, learner outline).
3. Integrity / depth / ordering / auth regression packs.
4. Minimum structural editor UI + a11y/mobile.
5. Frontend / Playwright / docs / full phase verification.

Layering for new work: Repository → repository tests → DTO/Controller → API tests → Frontend → Playwright → Docs.

## Completion calculation

$$\text{Phase Completion} = \frac{\text{Verified Task Points}}{\text{Total Task Points}} \times 100$$

Only `VERIFIED` tasks contribute.

## Task checklist

| ID | Task | Points | Status | Difficulty | Est. Time | Evidence | Dependencies |
|---|---|---:|---|---|---|---|---|
| COURSE-P1-T01 | Repository architecture assessment | 2 | VERIFIED | S | 2 hours | `docs/features/course-system/evidence/ledger-initialization/repository-assessment.md` | — |
| COURSE-P1-T01A | Ledger integrity verification | 1 | VERIFIED | XS | 1 hour | `docs/features/course-system/evidence/ledger-initialization/isolated-clean-copy.md`; `docs/features/course-system/evidence/ledger-initialization/checker-integrity.md` | COURSE-P1-T01 |
| COURSE-P1-T02 | Course model and migration | 5 | VERIFIED | S | 2 hours | `docs/features/course-system/evidence/course-p1-t02/verification.md`; `docs/features/course-system/evidence/course-p1-t02/migration.md` | COURSE-P1-T01 |
| COURSE-P1-T02B | Course hierarchy architecture decision | 3 | VERIFIED | S | 2 hours | `docs/features/course-system/evidence/course-p1-t02b/README.md` | COURSE-P1-T02 |
| COURSE-P1-T03 | CourseNode model and migration | 7 | VERIFIED | M | 4 hours | `docs/features/course-system/evidence/course-p1-t03/README.md` | COURSE-P1-T02B |
| COURSE-P1-T04 | Tree repository queries | 5 | VERIFIED | M | 4 hours | `docs/features/course-system/evidence/course-p1-t04/README.md` | COURSE-P1-T03 |
| COURSE-P1-T05 | Transactional branch move | 8 | VERIFIED | XL | 6-8 hours | `docs/features/course-system/evidence/course-p1-t05/README.md` | COURSE-P1-T04 |
| COURSE-P1-T06 | Transactional sibling reorder | 5 | VERIFIED | M | 4 hours | `docs/features/course-system/evidence/course-p1-t06/README.md` | COURSE-P1-T05 |
| COURSE-P1-T07 | Transactional subtree deletion | 5 | VERIFIED | M | 4 hours | `docs/features/course-system/evidence/course-p1-t07/README.md` | COURSE-P1-T06 |
| COURSE-P1-T08 | Admin course APIs | 7 | VERIFIED | M | 4 hours | `docs/features/course-system/evidence/course-p1-t08/README.md` | COURSE-P1-T02 |
| COURSE-P1-T09 | Admin tree APIs | 7 | VERIFIED | M | 4 hours | `docs/features/course-system/evidence/course-p1-t09/README.md` | COURSE-P1-T04, COURSE-P1-T08 |
| COURSE-P1-T10 | Admin hierarchy mutation APIs | 5 | VERIFIED | M | 4 hours | `docs/features/course-system/evidence/course-p1-t10/README.md` | COURSE-P1-T05, COURSE-P1-T06, COURSE-P1-T07, COURSE-P1-T09 |
| COURSE-P1-T11 | Admin course list UI | 4 | NOT_STARTED | S | 2 hours | — | COURSE-P1-T08 |
| COURSE-P1-T12 | Course creation UI | 4 | NOT_STARTED | S | 2 hours | — | COURSE-P1-T11 |
| COURSE-P1-T13 | Minimum structural tree editor UI | 8 | NOT_STARTED | L | 6 hours | — | COURSE-P1-T09, COURSE-P1-T10, COURSE-P1-T12 |
| COURSE-P1-T14 | Lazy-load children in structural editor | 5 | NOT_STARTED | M | 4 hours | — | COURSE-P1-T13, COURSE-P1-T34 |
| COURSE-P1-T15 | Accessible non-drag structural move/reorder | 6 | NOT_STARTED | M | 4 hours | — | COURSE-P1-T13 |
| COURSE-P1-T16 | Ancestor repository queries | 5 | NOT_STARTED | M | 4 hours | — | COURSE-P1-T04 |
| COURSE-P1-T17 | Dedicated subtree read repository + API | 5 | NOT_STARTED | M | 4 hours | — | COURSE-P1-T04, COURSE-P1-T09 |
| COURSE-P1-T18 | Publish-safe learner structural outline API | 6 | NOT_STARTED | M | 4 hours | — | COURSE-P1-T04, COURSE-P1-T16 |
| COURSE-P1-T19 | 25-level hierarchy contract tests | 6 | NOT_STARTED | M | 4 hours | — | COURSE-P1-T05, COURSE-P1-T09 |
| COURSE-P1-T20 | Move rejection matrix (self/descendant/cross-Course) | 6 | NOT_STARTED | M | 4 hours | — | COURSE-P1-T05, COURSE-P1-T10 |
| COURSE-P1-T21 | Empty-tree and full-tree retrieval tests | 4 | NOT_STARTED | S | 2 hours | — | COURSE-P1-T04, COURSE-P1-T09 |
| COURSE-P1-T22 | Unauthorized structural mutation matrix | 4 | NOT_STARTED | S | 2 hours | — | COURSE-P1-T08, COURSE-P1-T09, COURSE-P1-T10 |
| COURSE-P1-T23 | Backend deep-tree integration suite | 6 | NOT_STARTED | M | 4 hours | — | COURSE-P1-T19, COURSE-P1-T20, COURSE-P1-T21, COURSE-P1-T22 |
| COURSE-P1-T24 | Frontend structural-editor unit tests | 4 | NOT_STARTED | S | 2 hours | — | COURSE-P1-T13, COURSE-P1-T14, COURSE-P1-T15 |
| COURSE-P1-T25 | Playwright structural editor smoke | 6 | NOT_STARTED | M | 4 hours | — | COURSE-P1-T13, COURSE-P1-T18 |
| COURSE-P1-T26 | Phase 1 documentation alignment | 3 | NOT_STARTED | S | 2 hours | — | COURSE-P1-T31 |
| COURSE-P1-T27 | Full canonical Phase 1 verification | 5 | NOT_STARTED | S | 2 hours | — | COURSE-P1-T23, COURSE-P1-T24, COURSE-P1-T25, COURSE-P1-T26, COURSE-P1-T32, COURSE-P1-T33, COURSE-P1-T35, COURSE-P1-T36 |
| COURSE-P1-T28 | Ancestor HTTP API | 4 | NOT_STARTED | S | 2 hours | — | COURSE-P1-T16, COURSE-P1-T09 |
| COURSE-P1-T29 | Breadcrumb HTTP API | 3 | NOT_STARTED | S | 2 hours | — | COURSE-P1-T28 |
| COURSE-P1-T30 | 50-level hierarchy stress tests | 5 | NOT_STARTED | M | 4 hours | — | COURSE-P1-T19 |
| COURSE-P1-T31 | Relationship Contract API examples docs | 4 | NOT_STARTED | S | 2 hours | — | COURSE-P1-T10 |
| COURSE-P1-T32 | Tree integrity invariant verification pack | 5 | NOT_STARTED | M | 4 hours | — | COURSE-P1-T03, COURSE-P1-T07, COURSE-P1-T10 |
| COURSE-P1-T33 | Root and child ordering contract tests | 4 | NOT_STARTED | S | 2 hours | — | COURSE-P1-T06, COURSE-P1-T10 |
| COURSE-P1-T34 | Lazy/paginated children API | 5 | NOT_STARTED | M | 4 hours | — | COURSE-P1-T09 |
| COURSE-P1-T35 | Builder expand/collapse UX polish | 4 | NOT_STARTED | S | 2 hours | — | COURSE-P1-T13 |
| COURSE-P1-T36 | Large-tree performance measurement pack | 5 | NOT_STARTED | M | 4 hours | — | COURSE-P1-T21, COURSE-P1-T34 |

Total points: 186

## Task-specific acceptance criteria

### COURSE-P1-T01 — Repository architecture assessment
<!-- TASK:COURSE-P1-T01:ACCEPTANCE:START -->
- [x] Backend, frontend, database, authentication, testing, Compose, and CI conclusions cite direct repository evidence.
- [x] The canonical assessment distinguishes detected architecture from Course System implementation.
- [x] Baseline verification records exact commands, timestamps, durations, exit codes, and environments.
- [x] No Course model, migration, API, or frontend implementation is introduced.
<!-- TASK:COURSE-P1-T01:ACCEPTANCE:END -->

### COURSE-P1-T01A — Ledger integrity verification
<!-- TASK:COURSE-P1-T01A:ACCEPTANCE:START -->
- [x] A second sync is a byte-identical no-op.
- [x] Check passes immediately after sync and changes no files.
- [x] Human-authored text outside generated markers remains unchanged.
- [x] CURRENT_STATUS.md, status.json, and manifest.json remain synchronized.
- [x] An isolated clean HEAD copy initializes and validates with only the authorized overlay.
<!-- TASK:COURSE-P1-T01A:ACCEPTANCE:END -->

### COURSE-P1-T02 — Course model and additive migration
<!-- TASK:COURSE-P1-T02:ACCEPTANCE:START -->
- [x] An additive sql-migrate up/down pair creates and removes only the Course root table and its index.
- [x] The Course model and goqu-backed repository methods follow existing model conventions.
- [x] Local migration apply, rollback, and re-apply complete successfully.
- [x] Focused Course model tests and repository-supported backend verification pass.
- [x] No CourseNode, hierarchy, API, frontend, Phase 2, production, or NUC work is performed.
- [x] Canonical evidence is complete under `docs/features/course-system/evidence/course-p1-t02/`.
<!-- TASK:COURSE-P1-T02:ACCEPTANCE:END -->

### COURSE-P1-T02B — Course hierarchy architecture decision
<!-- TASK:COURSE-P1-T02B:ACCEPTANCE:START -->
- [x] ADR-023 defines the canonical Course aggregate, typed CourseNode persistence, hierarchy roots, logical-path invariants, lifecycle, curriculum-profile boundary, and T03 contract.
- [x] ADR-023 distinguishes normative requirements from implementation guidance and documents compatibility, governance, lifecycle, version, glossary, and supersession rules.
- [x] Course standards and Phase 1 documentation reference ADR-023 without creating a competing authoritative hierarchy.
- [x] Direct repository evidence confirms the decision is additive and does not duplicate an existing CourseNode, CourseSubject, or CourseTopic persistence model or migration.
- [x] Ledger sync/check and hash evidence prove the 108-point denominator, read-only check behavior, and byte-identical no-op sync.
- [x] No CourseNode code, migration, API, UI, later-phase implementation, production access, or NUC access is performed.
- [x] Canonical evidence is complete under `docs/features/course-system/evidence/course-p1-t02b/`.
<!-- TASK:COURSE-P1-T02B:ACCEPTANCE:END -->

### COURSE-P1-T03 — CourseNode model and additive migration
<!-- TASK:COURSE-P1-T03:ACCEPTANCE:START -->
- [x] An additive sql-migrate up/down pair creates and removes only the CourseNode persistence objects defined by ADR-023.
- [x] The CourseNode model provides transactional creation and Course-scoped lookup using repository Go, goqu, UUID, lifecycle, and error conventions.
- [x] Database constraints enforce same-Course parenting, valid values, Course-scoped logical-path uniqueness, and separate top-level/child sibling positions.
- [x] Focused CourseNode tests, existing Course tests, backend vet/tests/race/build, and formatting verification pass.
- [x] Isolated local PostgreSQL apply, schema inspection, constraint checks, rollback, and reapply pass without changing existing Course data.
- [x] Canonical evidence is complete under `docs/features/course-system/evidence/course-p1-t03/`.
- [x] No T04 work, tree queries, moves, reordering, semantic-profile validation, API, UI, Phase 2, production, or NUC work is performed.
<!-- TASK:COURSE-P1-T03:ACCEPTANCE:END -->

### COURSE-P1-T04 — CourseNode hierarchy repository queries
<!-- TASK:COURSE-P1-T04:ACCEPTANCE:START -->
- [x] Read-only Course-scoped root, child, and nested hierarchy repository operations follow ADR-023 and expose no stored paths.
- [x] Reads distinguish an unknown Course from an existing Course with an empty hierarchy and return non-nil empty slices for valid empty results.
- [x] The recursive hierarchy query uses explicit numeric-safe preorder ordering and detects incomplete reachable graphs defensively.
- [x] Focused hierarchy tests and backend regression verification pass.
- [x] Canonical evidence is complete under `docs/features/course-system/evidence/course-p1-t04/`; Database Migration: NONE.
- [x] No move, reorder, deletion, path mutation, cycle prevention, API, UI, authorization, LearningItem, Phase 2, production, or NUC work is performed.
<!-- TASK:COURSE-P1-T04:ACCEPTANCE:END -->

### COURSE-P1-T05 — Transactional CourseNode branch move
<!-- TASK:COURSE-P1-T05:ACCEPTANCE:START -->
- [x] MoveNode validates Course scope, exact destination position, parent validity, and move-specific cycle boundaries in one transaction.
- [x] Moves rewrite root and descendant paths using private boundary-safe codec matching and preserve IDs, Course membership, and lifecycle status.
- [x] The path rewrite count exactly equals the locked subtree size, including the moved root; mismatch rolls back with subtree conflict.
- [x] Focused mutation tests plus T03/T04 regression and backend verification pass.
- [x] Canonical evidence is complete under `docs/features/course-system/evidence/course-p1-t05/`; Database Migration: NONE.
- [x] No T06 reorder, deletion, API, UI, authorization, LearningItem, Phase 2, production, or NUC work is performed.
<!-- TASK:COURSE-P1-T05:ACCEPTANCE:END -->

### COURSE-P1-T06 — Transactional sibling reorder
<!-- TASK:COURSE-P1-T06:ACCEPTANCE:START -->
- [x] `ReorderChildren` validates nil/duplicate IDs before database access and rejects mismatched sibling ID sets.
- [x] Reorder locks the Course row, optionally the parent, and all siblings (`id ASC`) inside one transaction.
- [x] Empty sibling scopes succeed only with an empty ordered list; already-canonical orders commit without UPDATEs.
- [x] Non-canonical reorders use int64-checked two-phase temporary staging, then canonical `0..n-1` positions with strict update-set verification.
- [x] Focused reorder tests plus T03/T04/T05 regression and backend verification pass.
- [x] Canonical evidence is complete under `docs/features/course-system/evidence/course-p1-t06/`; Database Migration: NONE.
- [x] No deletion, API, UI, authorization, LearningItem, Phase 2, production, or NUC work is performed.
<!-- TASK:COURSE-P1-T06:ACCEPTANCE:END -->

### COURSE-P1-T07 — Transactional subtree deletion
<!-- TASK:COURSE-P1-T07:ACCEPTANCE:START -->
- [x] `DeleteSubtree` validates nil Course/node IDs without database access.
- [x] Deletion locks the Course row, target node, and boundary-safe subtree (`path ASC`) in one transaction.
- [x] The target node and every descendant are deleted; unrelated branches remain untouched.
- [x] Returned deleted IDs are verified exactly; mismatches return `ErrCourseNodeDeleteConflict` and roll back.
- [x] No automatic sibling reordering, path rewriting, archive, or soft-delete behaviour is introduced.
- [x] Focused delete tests plus T03–T06 regression and backend verification pass.
- [x] Canonical evidence is complete under `docs/features/course-system/evidence/course-p1-t07/`; Database Migration: NONE.
- [x] No API, UI, authorization, LearningItem, Phase 2, production, or NUC work is performed.
<!-- TASK:COURSE-P1-T07:ACCEPTANCE:END -->

### COURSE-P1-T08 — Admin course APIs
<!-- TASK:COURSE-P1-T08:ACCEPTANCE:START -->
- [x] Admin Course routes are registered under `/api/v1/admin/courses` with Kratos authentication and allowlist authorisation.
- [x] Create, list, get, and patch Course endpoints reuse CourseModel persistence without CourseNode hierarchy routes.
- [x] PATCH uses presence-aware omit/null/empty semantics and rejects empty patches without read-before-update comparisons.
- [x] Owner is derived only from authenticated context; status/archive/publish mutation and DELETE Course remain deferred.
- [x] Focused HTTP/auth tests plus Course and T03–T07 CourseNode regressions pass.
- [x] Canonical evidence is complete under `docs/features/course-system/evidence/course-p1-t08/`; Database Migration: NONE.
- [x] Temporary reuse of the quiz-admin allowlist is documented as a known limitation without introducing a new role model.
- [x] No CourseNode APIs, hierarchy mutation APIs, frontend, production, or NUC work is performed.
<!-- TASK:COURSE-P1-T08:ACCEPTANCE:END -->

### COURSE-P1-T09 — Admin CourseNode create/read APIs
<!-- TASK:COURSE-P1-T09:ACCEPTANCE:START -->
- [x] Admin CourseNode create/read routes are registered under `/api/v1/admin/courses/:course_id/nodes` with Kratos authentication and allowlist authorisation.
- [x] Create requires explicit position presence; parent_id omit/null creates a root; empty/whitespace/malformed parent_id is rejected.
- [x] Node reads are always Course-scoped (`course_id` + `node_id`); cross-Course node access returns 404.
- [x] Missing Course maps to 404 via repository `ErrCourseNodeCourseNotFound`; empty existing trees return non-nil empty roots.
- [x] Hierarchy mutation HTTP routes (PATCH/DELETE/move/reorder/delete-subtree) are absent (404 or 405).
- [x] Focused HTTP/auth tests plus T03–T08 regressions pass; canonical evidence is complete under `docs/features/course-system/evidence/course-p1-t09/`; Database Migration: NONE.
- [x] Temporary reuse of the quiz-admin allowlist is documented as a known limitation without introducing a new role model.
- [x] No hierarchy mutation APIs, LearningItems, frontend, production, or NUC work is performed.
<!-- TASK:COURSE-P1-T09:ACCEPTANCE:END -->

### COURSE-P1-T10 — Admin CourseNode hierarchy mutation APIs
<!-- TASK:COURSE-P1-T10:ACCEPTANCE:START -->
- [x] Admin mutation routes are registered: PATCH `.../nodes/:node_id/move`, POST `.../nodes/reorder`, DELETE `.../nodes/:node_id`, with Kratos authentication and allowlist authorisation.
- [x] Move requires presence-aware position; new_parent_id omit/null moves to root; empty/whitespace/malformed parent is rejected before repository call.
- [x] Reorder structurally validates ordered_node_ids (missing/null/empty/duplicates/empty/whitespace/malformed) before ReorderChildren.
- [x] Mutations return T08/T09 JSONSuccess 200 envelopes; errors use errors.Is mapping (400/404/409/500).
- [x] Still-absent routes remain absent: POST `.../move`, DELETE `.../subtree`, PATCH `.../:node_id` (non-move).
- [x] Focused HTTP tests include post-mutation GET tree/children checks plus T05–T09 regressions; evidence under `docs/features/course-system/evidence/course-p1-t10/`; Database Migration: NONE.
- [x] Temporary reuse of the quiz-admin allowlist is documented as a known limitation without introducing a new role model.
- [x] No repository mutation logic changes, archive/restore, LearningItems, frontend, permission redesign, production, or NUC work is performed.
<!-- TASK:COURSE-P1-T10:ACCEPTANCE:END -->

### COURSE-P1-T11 — Admin course list UI
<!-- TASK:COURSE-P1-T11:ACCEPTANCE:START -->
- [ ] Authenticated admin course list loads from `GET /api/v1/admin/courses` with empty and non-empty states.
- [ ] Unauthorized users are denied; no client-side authority bypass.
- [ ] Evidence under `docs/features/course-system/evidence/course-p1-t11/`; Database Migration: NONE.
<!-- TASK:COURSE-P1-T11:ACCEPTANCE:END -->

### COURSE-P1-T12 — Course creation UI
<!-- TASK:COURSE-P1-T12:ACCEPTANCE:START -->
- [ ] Admin can create a Course via UI calling `POST /api/v1/admin/courses`.
- [ ] Validation errors surface; owner is never client-supplied.
- [ ] Evidence under `docs/features/course-system/evidence/course-p1-t12/`.
<!-- TASK:COURSE-P1-T12:ACCEPTANCE:END -->

### COURSE-P1-T13 — Minimum structural tree editor UI
<!-- TASK:COURSE-P1-T13:ACCEPTANCE:START -->
- [ ] Editor can add root and child nodes via `parent_id` APIs at arbitrary depth without client depth rejection.
- [ ] Expand/collapse works; children load from API projections (not authoritative local `child_node_ids[]`).
- [ ] Move/reorder/delete call verified mutation APIs; destructive actions confirm.
- [ ] Evidence under `docs/features/course-system/evidence/course-p1-t13/`.
<!-- TASK:COURSE-P1-T13:ACCEPTANCE:END -->

### COURSE-P1-T14 — Lazy-load children in structural editor
<!-- TASK:COURSE-P1-T14:ACCEPTANCE:START -->
- [ ] Collapsed branches do not eagerly fetch all descendants.
- [ ] Expanding a node loads direct children via children/lazy API; expand state survives sibling refresh.
- [ ] Evidence under `docs/features/course-system/evidence/course-p1-t14/`.
<!-- TASK:COURSE-P1-T14:ACCEPTANCE:END -->

### COURSE-P1-T15 — Accessible non-drag structural move/reorder
<!-- TASK:COURSE-P1-T15:ACCEPTANCE:START -->
- [ ] Keyboard users can move a selected node before/after a sibling via server APIs.
- [ ] ARIA live-region (or equivalent) confirms successful move/reorder.
- [ ] Usable at ~360px width for structural operations.
- [ ] Evidence under `docs/features/course-system/evidence/course-p1-t15/`.
<!-- TASK:COURSE-P1-T15:ACCEPTANCE:END -->

### COURSE-P1-T16 — Ancestor repository queries
<!-- TASK:COURSE-P1-T16:ACCEPTANCE:START -->
- [ ] Course-scoped ancestor query returns ordered root→current chain without exposing path on normal DTOs.
- [ ] Missing node is distinct from empty ancestors; iterative/CTE preferred over unsafe deep recursion.
- [ ] Evidence under `docs/features/course-system/evidence/course-p1-t16/`; Database Migration: NONE.
<!-- TASK:COURSE-P1-T16:ACCEPTANCE:END -->

### COURSE-P1-T17 — Dedicated subtree read repository + API
<!-- TASK:COURSE-P1-T17:ACCEPTANCE:START -->
- [ ] Admin GET returns target node plus all descendants as computed nested `children[]`.
- [ ] Cross-Course access returns 404; empty subtree returns non-null empty children.
- [ ] Evidence under `docs/features/course-system/evidence/course-p1-t17/`.
<!-- TASK:COURSE-P1-T17:ACCEPTANCE:END -->

### COURSE-P1-T18 — Publish-safe learner structural outline API
<!-- TASK:COURSE-P1-T18:ACCEPTANCE:START -->
- [ ] Authenticated learner outline excludes draft CourseNodes (`status != PUBLISHED`) per documented omission/not-found policy.
- [ ] Outline works for unlimited depth; enrollment may remain deferred if documented.
- [ ] Evidence under `docs/features/course-system/evidence/course-p1-t18/`.
<!-- TASK:COURSE-P1-T18:ACCEPTANCE:END -->

### COURSE-P1-T19 — 25-level hierarchy contract tests
<!-- TASK:COURSE-P1-T19:ACCEPTANCE:START -->
- [ ] Creating and retrieving a ≥25-level CourseNode chain succeeds without schema changes, application depth validation, truncated results, or stack overflow.
- [ ] No product `MAX_*_DEPTH` constant is introduced.
- [ ] Evidence under `docs/features/course-system/evidence/course-p1-t19/`.
<!-- TASK:COURSE-P1-T19:ACCEPTANCE:END -->

### COURSE-P1-T20 — Move rejection matrix (self/descendant/cross-Course)
<!-- TASK:COURSE-P1-T20:ACCEPTANCE:START -->
- [ ] Self-parent, descendant-parent, and cross-Course moves are rejected.
- [ ] Rejected moves leave `parent_id`, position, and descendant paths unchanged (no partial writes).
- [ ] Evidence under `docs/features/course-system/evidence/course-p1-t20/`.
<!-- TASK:COURSE-P1-T20:ACCEPTANCE:END -->

### COURSE-P1-T21 — Empty-tree and full-tree retrieval tests
<!-- TASK:COURSE-P1-T21:ACCEPTANCE:START -->
- [ ] Empty Course hierarchy returns successful non-null empty collections.
- [ ] Full-tree retrieval returns each reachable node once in deterministic preorder.
- [ ] Evidence under `docs/features/course-system/evidence/course-p1-t21/`.
<!-- TASK:COURSE-P1-T21:ACCEPTANCE:END -->

### COURSE-P1-T22 — Unauthorized structural mutation matrix
<!-- TASK:COURSE-P1-T22:ACCEPTANCE:START -->
- [ ] Unauthenticated callers receive 401; non-admin authenticated callers receive 403 for admin mutations.
- [ ] Failed auth leaves no node/path/position/descendant row changes.
- [ ] Evidence under `docs/features/course-system/evidence/course-p1-t22/`.
<!-- TASK:COURSE-P1-T22:ACCEPTANCE:END -->

### COURSE-P1-T23 — Backend deep-tree integration suite
<!-- TASK:COURSE-P1-T23:ACCEPTANCE:START -->
- [ ] From `api/`: `go test ./...` covers T19–T22 regressions plus T03–T10; docker `api-verify` recorded in evidence.
- [ ] Evidence under `docs/features/course-system/evidence/course-p1-t23/`.
<!-- TASK:COURSE-P1-T23:ACCEPTANCE:END -->

### COURSE-P1-T24 — Frontend structural-editor unit tests
<!-- TASK:COURSE-P1-T24:ACCEPTANCE:START -->
- [ ] From `app/`: `npm run lint` and `npm test -- --run` cover structural editor suites for T13–T15.
- [ ] Evidence under `docs/features/course-system/evidence/course-p1-t24/`.
<!-- TASK:COURSE-P1-T24:ACCEPTANCE:END -->

### COURSE-P1-T25 — Playwright structural editor smoke
<!-- TASK:COURSE-P1-T25:ACCEPTANCE:START -->
- [ ] `npx playwright test` covers create course → nest nodes → deep-link/outline smoke.
- [ ] Evidence under `docs/features/course-system/evidence/course-p1-t25/`.
<!-- TASK:COURSE-P1-T25:ACCEPTANCE:END -->

### COURSE-P1-T26 — Phase 1 documentation alignment
<!-- TASK:COURSE-P1-T26:ACCEPTANCE:START -->
- [ ] Phase docs, architecture/current.md, and D-004/D-005 remain consistent; examples labeled non-normative.
- [ ] Ledger sync/check pass after documentation updates.
<!-- TASK:COURSE-P1-T26:ACCEPTANCE:END -->

### COURSE-P1-T27 — Full canonical Phase 1 verification
<!-- TASK:COURSE-P1-T27:ACCEPTANCE:START -->
- [ ] Full backend/frontend/Playwright/ledger gates pass for Phase 1 scope.
- [ ] Canonical evidence index complete; no production/NUC access.
<!-- TASK:COURSE-P1-T27:ACCEPTANCE:END -->

### COURSE-P1-T28 — Ancestor HTTP API
<!-- TASK:COURSE-P1-T28:ACCEPTANCE:START -->
- [ ] Admin GET returns ordered root→current ancestors for a Course-scoped node without exposing path on normal DTOs.
- [ ] Missing node returns 404; roots return non-null empty ancestor list.
- [ ] Evidence under `docs/features/course-system/evidence/course-p1-t28/`; Database Migration: NONE.
<!-- TASK:COURSE-P1-T28:ACCEPTANCE:END -->

### COURSE-P1-T29 — Breadcrumb HTTP API
<!-- TASK:COURSE-P1-T29:ACCEPTANCE:START -->
- [ ] Breadcrumb projection reuses ancestor data with stable titles/ids for builder and learner UI.
- [ ] Cross-Course access returns 404; no path leakage.
- [ ] Evidence under `docs/features/course-system/evidence/course-p1-t29/`.
<!-- TASK:COURSE-P1-T29:ACCEPTANCE:END -->

### COURSE-P1-T30 — 50-level hierarchy stress tests
<!-- TASK:COURSE-P1-T30:ACCEPTANCE:START -->
- [ ] Creating and retrieving a ≥50-level CourseNode chain succeeds without schema changes, application depth validation, truncated results, or stack overflow.
- [ ] No product `MAX_*_DEPTH` constant is introduced.
- [ ] Evidence under `docs/features/course-system/evidence/course-p1-t30/`.
<!-- TASK:COURSE-P1-T30:ACCEPTANCE:END -->

### COURSE-P1-T31 — Relationship Contract API examples docs
<!-- TASK:COURSE-P1-T31:ACCEPTANCE:START -->
- [ ] Docs include snake_case request/response examples for create root/child, children, tree, move, reorder, delete.
- [ ] Examples state `parent_id` authority; Subject→Module→Chapter→Lecture labeled non-normative.
- [ ] Evidence under `docs/features/course-system/evidence/course-p1-t31/`.
<!-- TASK:COURSE-P1-T31:ACCEPTANCE:END -->

### COURSE-P1-T32 — Tree integrity invariant verification pack
<!-- TASK:COURSE-P1-T32:ACCEPTANCE:START -->
- [ ] Automated checks prove D-005 invariants after create/move/reorder/delete sequences.
- [ ] No orphan `parent_id`, dual-parenting, or cross-Course edges remain after mutations.
- [ ] Evidence under `docs/features/course-system/evidence/course-p1-t32/`.
<!-- TASK:COURSE-P1-T32:ACCEPTANCE:END -->

### COURSE-P1-T33 — Root and child ordering contract tests
<!-- TASK:COURSE-P1-T33:ACCEPTANCE:START -->
- [ ] Root positions are Course-scoped unique; child positions are parent-scoped unique.
- [ ] Reorder/move preserve canonical `0..n-1` sibling positions for affected scopes.
- [ ] Evidence under `docs/features/course-system/evidence/course-p1-t33/`.
<!-- TASK:COURSE-P1-T33:ACCEPTANCE:END -->

### COURSE-P1-T34 — Lazy/paginated children API
<!-- TASK:COURSE-P1-T34:ACCEPTANCE:START -->
- [ ] Direct-children listing supports pagination/limit without changing the parent_id contract.
- [ ] Responses may include computed `has_children` / `child_count`; never authoritative `child_node_ids[]`.
- [ ] Evidence under `docs/features/course-system/evidence/course-p1-t34/`.
<!-- TASK:COURSE-P1-T34:ACCEPTANCE:END -->

### COURSE-P1-T35 — Builder expand/collapse UX polish
<!-- TASK:COURSE-P1-T35:ACCEPTANCE:START -->
- [ ] Expand/collapse state is stable across lazy child loads.
- [ ] Empty-child nodes show empty state; expand does not invent local hierarchy authority.
- [ ] Evidence under `docs/features/course-system/evidence/course-p1-t35/`.
<!-- TASK:COURSE-P1-T35:ACCEPTANCE:END -->

### COURSE-P1-T36 — Large-tree performance measurement pack
<!-- TASK:COURSE-P1-T36:ACCEPTANCE:START -->
- [ ] Evidence records measurement of full-tree vs lazy-children retrieval for a large synthetic Course.
- [ ] Operational limits documented as non-depth; no product MAX_DEPTH introduced.
- [ ] Evidence under `docs/features/course-system/evidence/course-p1-t36/`.
<!-- TASK:COURSE-P1-T36:ACCEPTANCE:END -->

## Phase acceptance criteria

- [x] Course CRUD admin APIs verified.
- [x] CourseNode persistence, hierarchy reads, move, reorder, delete APIs verified.
- [ ] Subtree/ancestor/breadcrumb and learner outline APIs verified.
- [ ] Unlimited-depth contract tests (≥25 and ≥50) pass with no product depth cap.
- [ ] Minimum structural editor ships (desktop + mobile + keyboard).
- [ ] Unauthorized mutations change no rows.
- [ ] Backend, frontend, Playwright, and ledger verification pass.
- [ ] Documentation matches D-004/D-005 and ADR-023.

## Verification commands

Task-focused: relevant `go test` packages / Vitest suites only.
Phase integration / full verify:

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

* VERIFIED tasks retain existing evidence paths under `docs/features/course-system/evidence/`.
* New tasks use `course-p1-tNN/` when verified. Do not fabricate evidence.

## Security requirements

* Admin mutations require Kratos + current admin gate; failed auth changes no rows.
* Cross-Course ID probing returns 404 without leaking sibling structure.
* Learner outline must not leak draft node existence beyond documented policy.

## Performance requirements

* Prefer CTE/iterative traversal; operational page/payload limits are not product depth.
* Large-tree measurements recorded in T36.

## Accessibility and mobile requirements

* Structural editor keyboard path (T15); usable near 360px width; ARIA confirmation for moves.

## Risks and mitigations

| Risk | Mitigation |
|---|---|
| UI invents hierarchy authority | Server APIs remain SoT |
| Accidental MAX_DEPTH | D-004; reject in review |
| Studio scope creep | Phase boundaries; Phase 7 owns studio |

## Known limitations

* Admin APIs reuse quiz-admin allowlist (Phase 8 auth model gate).
* Cross-Course copy is Phase 7, not Phase 1.
* VERIFIED tasks do not claim 25/50-level suites, ancestor APIs, learner outline, or builder UI.

## Exit criteria

* All Phase 1 tasks `VERIFIED` with evidence; denominator **186** points.
* Phase acceptance criteria checked; ledger sync/check pass.
* No production/NUC destructive operations.
