# Course System Development Changelog

## [0.1.0] — Phase 1/2 — Course Tree + Learning Items (Unreleased)

### Added

- Persistent ledger, canonical evidence, Course root persistence, ADR-023, CourseNode persistence, and hierarchy reads.
- Transactional CourseNode branch moves, sibling reordering, and subtree deletion.
- Admin Course HTTP APIs under `/api/v1/admin/courses`.
- Admin CourseNode create/read HTTP APIs under `/api/v1/admin/courses/:course_id/nodes`.
- Admin CourseNode hierarchy mutation HTTP APIs: PATCH move, POST reorder, DELETE node.
- LearningItem additive migration and CourseNode-scoped append-only repository CRUD (COURSE-P2-T01/T02).
- Structural Information Block metadata envelope on LearningItem JSONB (COURSE-P2-T03).
- Creator placeholder syntax validation inside block data strings (COURSE-P2-T04).
- Optional block visibility metadata with typed modes (COURSE-P2-T05).
- Admin LearningItem HTTP CRUD nested under CourseNode (COURSE-P2-T06).
- LearningItem `publish_state` (`DRAFT`|`PUBLISHED`) migration, repository, and admin HTTP (COURSE-P2-T07).
- Authenticated learner LearningItem GET list/get under `/api/v1/learner` with repository-owned publish filtering (COURSE-P2-T08).
- Admin LearningItem sibling reorder under `/api/v1/admin/courses/:course_id/nodes/:node_id/learning-items/reorder` (COURSE-P2-T10).
- Admin LearningItem cross-node move under `/api/v1/admin/courses/:course_id/nodes/:node_id/learning-items/move` (COURSE-P2-T11).
- Deep-node LearningItem ownership sqlmock pack (≥10 nesting levels, root-vs-deep parity, cross-Course no-leak) (COURSE-P2-T20).
- Node-local LearningItem previous/next repository resolution via `GetAdjacentLearningItems` (COURSE-P2-T21).
- Draft filtering regression pack proving learner hide-draft / admin keep-draft contracts (COURSE-P2-T24).
- Publish-filter controller contract tests proving repository-owned filtering, Kratos auth boundaries, transport-only ordering/serialization, and stable error mapping (COURSE-P2-T25).
- Repository-owned learner runtime visibility projection on published getters (COURSE-P2-T13).
- Expose node-local previous/next published item navigation via learner HTTP GET endpoints (COURSE-P2-T22).
- Node-local published LearningItem projection (id, title) for a single CourseNode (COURSE-P2-T23).
- Canonical snake_case LearningItem API examples for admin CRUD and authenticated learner published reads (COURSE-P2-T26).
- Course enrollment persistence (`course_enrollments`) with learner LearningItem GET enrollment gate, self-enroll endpoints, and learner enroll CTA (COURSE-P2-T12 VERIFIED; D-006).
- Local development Course-admin seed helper (`npm run seed:local-course-admin`) and docs (`docs/development/local-course-admin.md`) for automated allowlisted verification.
- **COURSE-P2-T14 VERIFIED**: authenticated learner list/detail rendering at depth 4, server-order and draft-exclusion correlation, enrollment/empty/denial flows, responsive desktop/mobile evidence, and signed-out layout hardening.
- **COURSE-P2-T16 VERIFIED**: direct frontend composition-component and admin-transport coverage, 63 focused LearningItem tests, authenticated desktop/mobile CRUD and learner smoke, persistence cleanup, and production-source audit (`NO`).
- **COURSE-P2-T17 VERIFIED**: existing Playwright Chromium suite covers deep-node admin DRAFT create, learner exclusion, publish, API-ordered learner list/detail, responsive desktop/mobile views, signed-out denial, and verified zero-residue cleanup; production-source audit (`NO`).
- Canonical T05–T10 and P2-T01–T08 / P2-T10 / P2-T11 / P2-T13 / P2-T20 / P2-T21 / P2-T22 / P2-T23 / P2-T24 evidence under `docs/features/course-system/evidence/`.

### Changed

- **COURSE-P2-T12**: D-006 authorized additive `course_enrollments` migration; learner LearningItem GET list/get require enrollment; unenrolled denial is HTTP 404 `course enrollment required` without payloads; learner self-enroll GET/POST/DELETE; learner UI enroll CTA; tests updated. Database Migration: YES. Breaking Changes: YES.
- **COURSE-P2-T09**: added `/admin/courses/learning-items`, reusable selectors/table/dialogs/empty state, and a transport-only admin Course/LearningItem composable. Items preserve server order and metadata is omitted from scalar create/update payloads. Focused tests, lint, build, Compose config, and `web` image build pass. Allowlisted runtime CRUD + desktop/mobile screenshots verified via local course-admin seed; task **VERIFIED**. Full frontend suite remains inherited baseline outside T09. Database Migration: NONE. Breaking Changes: NO.

- Authorized parallel Phase 2 LearningItem persistence via D-003 while Phase 1 UI remains open.
- Retitled COURSE-P2-T07 from Learner API endpoints to LearningItem publish state.
- Retitled COURSE-P2-T08 from Admin item-chain builder UI to Learner LearningItem APIs; admin builder is now `COURSE-P2-T09`.
- Documentation governance: D-004/D-005 Relationship Contract; agent protocol in module README.
- DOC-CS-T01: canonical architecture and relationship contract.
- DOC-CS-T02: initial eight-phase expansion draft (incomplete for freeze readiness).
- **DOC-CS-T02-R1**: complete phase rewrite with ownership freeze, design gates, explicit dependencies, arithmetic repair, and status regeneration. Roadmap freeze-ready; DOC-CS-T03/T04 still pending.
- Phase 1 denominator 186 points (60 verified → 32.26%); Phase 2 denominator 148 points (54 verified → 36.49%); overall completion 10.31%.
- Next coding task selected as `COURSE-P2-T10` (verified-only deps; reorder API gap confirmed).
- **DOC-CS-T03**: documentation normalization; cross-reference cleanup; terminology normalization; Canonical Reference Policy applied (module README contract collapsed to links); Architecture Freeze Marker recorded in `AI_CONTEXT.md` and `HANDOFF.md`. No architecture changes. No production code changes. Remaining docs task: DOC-CS-T04.
- **DOC-CS-T04**: final documentation audit and freeze certification. Issued `DOCUMENTATION_FREEZE.md` (Documentation Freeze Approved). Objective defect fixes only: learning-sequence ownership label in `architecture/current.md` §19; added `COURSE-P3-T12` to Phase 5 dependencies. No architecture changes. No roadmap redesign. No production code changes. Implementation baseline remains `COURSE-P2-T10`.
- **COURSE-P2-T10**: transactional admin LearningItem sibling reorder with bidirectional exact-set validation, `ORDER BY position, id` locks, idempotent noop, frozen response payload, concurrency conflict mapping, and evidence under `course-p2-t10/` (including `transaction-timeline.md`). Database Migration: NONE.
- **COURSE-P2-T11**: transactional admin LearningItem cross-node move with UUID-ordered dual-node locks, six-step staging, subset validation without existence leak, empty noop with real counts, frozen response payload, concurrency conflict mapping, and evidence under `course-p2-t11/` (including `transaction-timeline.md` and `move-algorithm.md`). Database Migration: NONE.
- **COURSE-P2-T20**: deep-node LearningItem ownership tests — attach/list/get at depths 1..10, root-vs-deep parity, cross-Course not-found without scope leak, published reads on depth-10 leaf; evidence under `course-p2-t20/` (including `ownership-tests.md`). Database Migration: NONE. No new APIs.
- **COURSE-P2-T21**: node-local previous/next repository resolution with deterministic `(position, id)` sibling order, nil chain-end semantics, scoped ownership validation, no publish filtering; evidence under `course-p2-t21/` (including `previous-next-semantics.md` and `query-contract.md`). Database Migration: NONE. No HTTP (T22).
- **COURSE-P2-T24**: draft filtering regression pack — learner list/get never expose DRAFT (404 equivalence for draft/missing get); admin list/get still return drafts; discovery-leak assertions; evidence under `course-p2-t24/`. Database Migration: NONE. No production code changes required.
- **COURSE-P2-T12 BLOCKED disposition** (docs/ledger only): recorded enrollment gate as BLOCKED because no Course enrollment persistence model, helper, relation, or documented equivalent exists; Migration frozen NONE. Blocker evidence under `course-p2-t12/` (`README.md`, `blocker.md`). No production code, no API, no migration, no frontend, no VERIFIED claim; T12 acceptance remains unchecked. Risk token `blocker:COURSE-P2-T12`. Breaking Changes: NO. Database Migration: NONE.
- **COURSE-P2-T13**: repository-owned learner runtime visibility projection (`ProjectLearningItemForLearner`) after SQL publish filtering; HIDDEN/INSTRUCTOR omitted; AUTHENTICATED kept; PREMIUM omitted under production defaults (`Authenticated=true`, `PremiumAuthorized=false`); deep-copy metadata; atomic list projection; admin unfiltered; controllers transport-only; evidence under `course-p2-t13/`. Database Migration: NONE. Frontend: NO. Breaking Changes: NO. Ledger next becomes `COURSE-P2-T22`.
- **COURSE-P2-T22**: node-local previous/next learner API with database-level published filtering, skipping draft items at query boundaries (no Go loops), deterministic ordering, all-or-nothing controller flow returning wrapped detail response contract; evidence under `course-p2-t22/`. Database Migration: NONE. Frontend: NO. Breaking Changes: YES (response payload wrapped).
- **COURSE-P2-T23**: node-local learning-chain projection returning ordered published item IDs/titles for a single CourseNode, ordered deterministically by position and ID, returning empty non-nil slice on empty nodes, transport-neutral DTO structure, strict SQL boundaries; evidence under `course-p2-t23/`. Database Migration: NONE. Frontend: NO. Breaking Changes: NO.
- **COURSE-P2-T25**: learner controller contract tests covering unauthenticated 401/no-query behavior, authenticated list/detail repository delegation, adversarial draft forwarding, repository-order preservation, wrapped previous/next serialization, projected metadata forwarding, and 404/500 mapping; evidence under `course-p2-t25/`. Production code: NONE. Database Migration: NONE. Frontend: NO. Breaking Changes: NO.
- **COURSE-P2-T26**: canonical `docs/api/course-learning-items-v1.md` documents source-verified admin CRUD, learner published reads, JSend envelopes, metadata enums, node-local navigation, any-depth attachment, and `CourseNode.status` versus `LearningItem.publish_state`; evidence under `course-p2-t26/`. Production code: NONE. Database Migration: NONE. Breaking Changes: NO.



### Verification

- Docker Go 1.23 vet, full tests, race tests, and build passed.
- Isolated PostgreSQL prove-out for `learning_items` and `publish_state` apply/rollback/reapply.
- LearningItem repository, metadata, placeholder, visibility, admin HTTP, publish_state, learner HTTP, reorder, move, deep-node ownership, adjacent previous/next, draft-filtering regression, and runtime visibility projection tests passed.
- Course/CourseNode regression remained green.

### Known limitations

- Admin Course APIs (T06–T10) reuse the existing quiz-admin allowlist as the repository's current general administrative gate. They do not introduce a new role or permission model.
- Learner LearningItem APIs require Kratos authentication and Course enrollment (`course_enrollments`; D-006 / T12).
- Admin composition UI (T09) is VERIFIED. Learner rendering UI (T14) remains IN_PROGRESS pending learner runtime screenshots. Placeholder resolution and authoring remain unavailable. Runtime visibility write-time validation (T05) and learner runtime projection (T13) are verified; no Course premium entitlement source exists yet (`PremiumAuthorized=false`).

### Deferred

- Coding: `COURSE-P2-T17` Playwright LearningItem E2E, `COURSE-P1-T11` structural UI, and later phases.
- Docs governance: programme complete (DOC-CS-T01 through DOC-CS-T04); future changes follow `DOCUMENTATION_FREEZE.md`.
