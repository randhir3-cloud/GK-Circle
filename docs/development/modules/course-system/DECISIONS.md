# Architecture and Product Decisions

## D-001 — Use identifier-based materialised paths

* **Date**: 2026-07-25
* **Status**: ACCEPTED
* **Phase**: 1

### Context
Node slugs may change, while subtree lookup and branch movement require a stable path.

### Decision
Store an ordered logical sequence of immutable node UUIDs instead of slugs.
The exact encoded delimiter is an implementation detail. ADR-023 is the
canonical source for the complete path invariants and hierarchy decision.

### Rationale & Alternatives Rejected
* **Rejected Alternative: Slug-based path**: Slugs change when titles are updated, which would require recursively rewriting paths of all descendants on any rename.
* **Rejected Alternative: Nested Sets**: While nested sets are extremely fast for reads, they are notoriously slow and complex to update when branches are moved or siblings are reordered, as they require rewriting left/right boundaries of most nodes in the tree. Materialized paths strike the perfect balance between fast subtree reads and simple branch updates.

### Consequences
* Renaming a node does not rewrite descendants.
* Moving a branch still rewrites the moved subtree.
* Public URLs must not be derived directly from the internal path.

## D-002 — Canonical typed CourseNode hierarchy

* **Date**: 2026-07-25
* **Status**: ACCEPTED
* **Phase**: 1
* **Canonical ADR**: `docs/architecture/ADR/ADR-023-canonical-course-hierarchy-model.md`

### Decision

Use Course as the aggregate root and typed CourseNode rows as the single
hierarchy persistence model. Keep CourseSubject and CourseTopic as domain
vocabulary and API projections.

### Governance

This module summary is not a competing decision record. ADR-023 and its
canonical T02B evidence are authoritative.

## D-003 — Parallel Phase 2 LearningItem persistence before Phase 1 UI completion

* **Date**: 2026-07-26
* **Status**: ACCEPTED
* **Phase**: 1 + 2

### Context

Phase 1 Course/CourseNode admin HTTP APIs (T08–T10) are verified. Remaining
Phase 1 work is UI (T11–T13+) and later hardening. LearningItem persistence is
the start of Phase 2 and does not depend on admin UI surfaces.

MASTER_PLAN allows a later phase to begin before the preceding phase is
`VERIFIED` only when an explicit architectural decision records why parallel
work is safe.

### Decision

Authorize parallel Phase 2 LearningItem schema and repository work
(`COURSE-P2-T01`, `COURSE-P2-T02`) while Phase 1 remains `IN_PROGRESS` with
`COURSE-P1-T11` and later UI tasks still `NOT_STARTED`.

### Rationale

* Phase 1 remaining work is UI-only.
* LearningItem persistence is additive and independent of the remaining Phase 1 UI tasks.
* Parallel Phase 2 work does not modify Course or CourseNode hierarchy
  semantics, repositories, or verified HTTP behaviour.

### Consequences

* `parallel=true` is required while multiple tasks are `IN_PROGRESS`.
* Phase 1 UI (`COURSE-P1-T11`) remains available under this decision.
* Status tooling must derive `activePhaseId` from the in-progress task’s phase.

## D-004 — Unlimited Logical Course Tree Depth

* **Date**: 2026-07-26
* **Status**: ACCEPTED
* **Phase**: 1+
* **Clarifies**: ADR-023 (no depth column; recursive adjacency + path)
* **Canonical detail**: `architecture/current.md`

### Context

Course builders and learner navigation require arbitrary nesting (file-explorer
style). Operational protections (payload size, batch size, rate limits, page
size) must not be mistaken for a fixed curriculum schema. Examples such as
`Subject → Module → Chapter → Lecture` are **non-normative** illustrations only.

### Decision

Unlimited logical CourseNode nesting is a **product invariant**.

* CourseNode hierarchy is recursive.
* Every CourseNode may contain zero, one, or unlimited child CourseNodes.
* Every CourseNode may also contain zero or more ordered LearningItems.
* The database schema does not change with hierarchy depth.
* There is no product-level hierarchy depth limit.
* Constants such as `MAX_COURSE_DEPTH`, `MAX_NODE_DEPTH`, or
  `MAX_HIERARCHY_LEVELS` are forbidden as product rules.
* Operational limits for payload size, rate limiting, page size, or abuse
  protection do not redefine the hierarchy as fixed-depth.
* Deep-tree operations must avoid unsafe application recursion or stack
  overflow (prefer CTE / iterative traversal).
* Existing CourseNode path and hierarchy rules remain governed by ADR-023.
* ADR-023 / T03 forbid a persisted `depth` column today. Depth is computable
  from the hierarchy or materialized path. A denormalized `depth` column MAY
  be added later only via ADR-approved additive migration, and MUST NEVER
  enforce a maximum level.
* Structural tree edges, learning-sequence order, and prerequisite DAG edges
  are three related but distinct graphs (see `architecture/current.md`).

### Consequences

* Deep-tree tests and UI lazy-loading remain roadmap work; historical VERIFIED
  tasks are not rewritten to claim those suites.
* Performance/abuse limits must be documented as operational, not product depth.

## D-005 — CourseNode Parent-Child Relationship Contract

* **Date**: 2026-07-26
* **Status**: ACCEPTED
* **Phase**: 1+
* **Canonical detail**: `architecture/current.md` (Course Node Relationship Contract)
* **Aligns with**: ADR-023 adjacency list + identifier-based materialized path

### Context

Tree APIs and builders need nested `children[]`, but persisting both
`parent_id` and an authoritative `child_node_ids[]` duplicates edges and drifts
under move, delete, reorder, and concurrency.

### Decision

* **`parent_id` is the source of truth** for structural relationships.
* Root nodes have `parent_id = NULL`; nested nodes have
  `parent_id =` direct parent CourseNode ID.
* **`children[]` is a response projection.** `child_count` and `has_children`
  are computed values. They MUST NOT be stored as an authoritative child-ID
  array, children JSONB column, or equivalent duplicated relationship
  persistence.
* Children are derived using
  `child.course_id = parent.course_id AND child.parent_id = parent.id`,
  ordered by `position ASC, id ASC`.
* The same CourseNode can contain both child CourseNodes and LearningItems.
* Structural tree edges, learning sequence edges, and prerequisite dependency
  edges are separate graph concepts.
* CourseNode lifecycle field remains **`status`** (`DRAFT`|`PUBLISHED`|`ARCHIVED`).
  LearningItem uses **`publish_state`** — names stay distinct.
* Persisted `node_type` remains `SECTION`|`SUBJECT`|`TOPIC` until an
  ADR-approved enum expansion; MODULE/CHAPTER/LECTURE and similar labels are
  vocabulary/examples only.
* Course FK on `course_nodes.course_id` is `ON DELETE RESTRICT`
  (migration `20260725134707`).

### Required invariants (normative; verbatim)

1. Every CourseNode belongs to exactly one Course.
2. Every root has parent_id = NULL.
3. Every non-root has exactly one parent_id.
4. Parent must belong to the same Course.
5. A node cannot parent itself.
6. A node cannot move under one of its descendants.
7. Children are derived from parent_id references and are not stored as child arrays.
8. A node may have zero, one, or unlimited children.
9. Sibling order is determined by position; child positions are parent-scoped and root positions are Course-scoped.
10. Moving a node moves its complete descendant subtree.
11. Node IDs remain unchanged during movement.
12. Parent, paths, sibling positions, and any future denormalized hierarchy metadata are updated transactionally.
13. Cross-Course movement is rejected.
14. Cross-Course copying is a separate validated workflow.
15. Full-tree and subtree APIs may expose computed children[].
16. Large trees may use lazy or paginated child APIs without changing the relationship contract.
17. No product MAX_*_DEPTH rule is permitted.

### Consequences

* Normative docs forbid introducing `child_node_ids[]` as source of truth.
* Do not document Course delete as CASCADE unless a future migration changes it.

## D-006 — Course enrollment persistence for learner LearningItem delivery

* **Date**: 2026-07-27
* **Status**: ACCEPTED
* **Phase**: 2
* **Authorizes**: `COURSE-P2-T12` (unblocks prior `blocker:COURSE-P2-T12`)
* **Prompt authority**: `COURSE-P2-T12-UNBLOCK-01`

### Context

Learner LearningItem GET APIs were authenticated via Kratos only. T12 requires
Course enrollment (or a documented equivalent) server-side. No existing
persisted Course-access relation, identity claim, or approved equivalent
existed. The prior disposition froze Migration as NONE pending governance.

Authenticated-only access is **not** a documented enrollment equivalent.

### Decision

Authorize an additive `course_enrollments` persistence model and server-side
enrollment gate for learner LearningItem GET list and get endpoints.

1. Persist enrollment as `(course_id, user_id)` with uniqueness.
2. Learner LearningItem GET list/get require an active enrollment row for the
   authenticated user and Course.
3. Unenrolled authenticated callers receive a documented denial **without**
   LearningItem payloads (HTTP 404; `course enrollment required`).
4. Learners may self-enroll via authenticated learner enrollment endpoints.
5. Kratos authentication remains required; quiz-admin allowlist is not used for
   learner delivery.

### Rationale

* Category A repository blocker (missing schema/API) within project scope.
* UNBLOCK-01 authorizes governance path 3 (schema change) rather than leaving
  T12 permanently blocked.
* Self-enrollment provides a testable, production-usable access source without
  inventing external secrets.

### Consequences

* Additive migration is required for T12 (Migration: YES for T12).
* Prior `blocker:COURSE-P2-T12` is resolved by this decision.
* T15 enrollment regressions become implementable.
* Breaking change for learner LearningItem GET clients that assumed
  authentication alone was sufficient.
