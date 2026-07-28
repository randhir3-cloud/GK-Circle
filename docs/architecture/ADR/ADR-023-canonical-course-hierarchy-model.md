# ADR-023: Canonical Course Hierarchy Model

- ADR Version: 1
- Schema Version: 1
- Status: Accepted
- Proposed: 2026-07-25
- Accepted: 2026-07-25T12:52:50.0300196+05:30
- Owners: Architecture and Course System
- Decision task: `COURSE-P1-T02B`

## Normative language

The key words **MUST**, **MUST NOT**, **SHOULD**, **SHOULD NOT**, and **MAY**
are to be interpreted as described in RFC 2119.

ADR Version tracks revisions to this architectural decision document. Schema
Version tracks the logical hierarchy model and is independent of migration
filenames, timestamps, or individual SQL revisions. Neither value is a database
column.

## Executive summary

Course remains the aggregate, ownership, and authorization root. Curriculum is
stored as one generic, typed `CourseNode` hierarchy using an adjacency list plus
an identifier-based materialized path. A Course has one or more ordered
top-level nodes; it does not have a synthetic hierarchy-root row.

`CourseSubject` and `CourseTopic` remain domain vocabulary and API projections.
Their persistence type is `CourseNode`, with initial node types `SECTION`,
`SUBJECT`, and `TOPIC`.

## Out of scope

This ADR does not authorize CourseNode code or migrations, APIs, tree
operations, recursive-query optimization, caching, event sourcing, audit
history, search indexing, builder UI, LearningItem changes, or permissions
beyond existing Course ownership. Those remain separately approved tasks.

## Context

The Phase 1 plan selected a flexible CourseNode tree and identifier-based
materialized paths. `docs/standards/course-rules.md` instead described
CourseSubject and CourseTopic as a fixed physical hierarchy. Implementing either
interpretation without resolving that conflict would create competing
persistence models.

GK Circle must support academic curricula as well as broader structures that do
not naturally map every grouping to a Subject. The decision must preserve the
academic default without baking that semantic profile into the database.

## Decision

### Aggregate and hierarchy roots

- `Course` MUST remain the aggregate, ownership, and authorization root.
- A Course hierarchy MUST contain one or more ordered top-level `CourseNode`
  rows whose `parent_id` is null.
- `parent_id` MUST be null if and only if a node is top-level.
- The model MUST NOT introduce a synthetic root-node type.

This separates the DDD aggregate root from persistence hierarchy roots.

### Persistence and vocabulary

- Persistence type: `CourseNode`.
- Domain vocabulary: `CourseSubject` and `CourseTopic`.
- API projection: Subject and Topic resources.
- Storage: typed `CourseNode` rows.
- Separate CourseSubject and CourseTopic persistence models MUST NOT be
  introduced.

Initial node types are:

- `SECTION`: a purely structural grouping node. It carries no curriculum
  semantics by itself and is reusable across curriculum profiles.
- `SUBJECT`: an academic subject node.
- `TOPIC`: an academic topic node.

Unknown node types MUST be rejected until a new ADR and additive migration
introduce them. Existing node types MUST remain backward compatible unless a
superseding ADR defines a migration strategy.

### Logical and encoded paths

The hierarchy MUST use an adjacency list plus an identifier-based materialized
path.

A logical path is an ordered sequence of immutable node UUIDs:

1. A top-level node's path contains exactly its own UUID.
2. A child's path begins with its parent's complete path and ends with the
   child's UUID.
3. Every UUID appears at most once in a path.
4. The final UUID equals the current node's ID.
5. Every referenced node belongs to the same Course.
6. The logical path is unique within a Course. Identical encoded values MAY
   exist in different Courses.
7. `course_id` scopes the hierarchy and MUST NOT be included in the logical
   path.

The exact string encoding and delimiter are implementation details. A codec
change is not breaking when the logical contract and persistence interfaces
remain unchanged.

Paths are backend-owned:

- Create, update, and move DTOs MUST NOT contain a path.
- API validation MUST reject client attempts to supply a path.
- Persistence models MUST NOT expose a path through standard JSON.
- A dedicated read-only diagnostic DTO MAY expose path data to internal
  administrative tooling.
- The persistence layer MUST construct and rebuild stored paths.
- Higher layers request moves by identifiers and positions and MUST NOT
  manipulate encoded paths.

Node IDs MUST use the repository's canonical UUID generation strategy. T03 must
recheck that policy before implementation; the current repository strategy is
`uuid.NewUUID()`.

### Position and atomicity

Sibling ordering MUST use integer positions. Parent validation, sibling-position
allocation, and path construction MUST occur atomically in one database
transaction.

Implementation guidance:

- The persistence layer SHOULD lock the parent while creating or moving a
  child.
- Top-level allocation SHOULD lock an appropriate Course-level row.
- Code SHOULD NOT calculate `MAX(position)` outside the transaction.
- Unique indexes SHOULD provide the final concurrent-integrity guard.
- The path codec SHOULD remain encapsulated in persistence code.

### Lifecycle and effective visibility

Initial CourseNode lifecycle states are `DRAFT`, `PUBLISHED`, and `ARCHIVED`.
`ARCHIVED` is a lifecycle state, not an access-control decision. Effective
learner visibility is derived later from Course visibility, ancestor lifecycle,
and authorization rules.

Changing status MUST NOT alter parent-child relationships, sibling positions,
Course membership, or stored paths.

### Controlled node-type changes

A node's type is immutable through ordinary updates. A dedicated
administrative/domain operation may change it only after revalidating the
affected subtree against the active curriculum profile.

Type changes MUST be rejected if they invalidate descendant semantics or
published-content policies. The initial conservative policy requires the
affected subtree to be draft-only and free of published-content, progress, or
audit dependencies. The dedicated operation and full validation belong to T07,
not T03.

### Curriculum profiles

T03 MUST NOT add a `curriculum_profile` column, profile table, or persistent
profile metadata. Until profile persistence exists, all repositories, services,
and tests MUST assume the default academic profile; callers MUST NOT override
it.

The academic profile requires a TOPIC to resolve to a nearest SUBJECT ancestor.
SECTION nodes may occur between them. Future profiles may relax or replace
these semantic rules without changing the hierarchy backbone. Persistent
profile selection requires demonstrated business need, a separate ADR, and
additive implementation.

### Move and copy semantics

- Move preserves node IDs, Course membership, lifecycle status, and existing
  LearningItem references.
- Copy always creates new node IDs and regenerates all descendant paths.
- Copy MUST NOT implicitly copy LearningItem references, learner progress,
  enrollment, analytics, or audit data.

### Query strategy

Recursive CTEs, prefix scans, caching, and other hierarchy query strategies are
implementation choices provided they preserve the normative hierarchy
contracts. This ADR does not select or authorize a query optimization.

## T03 implementation contract

The future additive `course_nodes` schema is limited to:

- UUID primary key;
- Course UUID;
- nullable parent UUID;
- node type;
- title;
- integer sibling position;
- encoded path;
- lifecycle status;
- created and updated timestamps;
- Course foreign key;
- same-Course parent integrity;
- logical-path uniqueness per Course;
- separate top-level and child sibling-position uniqueness; and
- path-prefix lookup support.

T03 MUST NOT add slugs, descriptions, depth, visibility, profile persistence,
metadata blobs, soft deletion, audit-history tables, tree APIs, move/copy
services, or frontend work.

T03 repository behavior is limited to node creation and Course-scoped lookup by
ID. Tree reads, moves, reordering, controlled type changes, semantic validation,
APIs, and authorization integration remain later tasks.

## Compatibility

- Existing Courses remain valid.
- Existing Course IDs, ownership, lifecycle, and APIs are unaffected.
- CourseNode is additive and requires no Course backfill or persistence change.

## Consequences and risks

The hybrid model avoids competing Subject and Topic tables while retaining
familiar domain/API language. It supports non-academic structural grouping
without weakening the default academic semantics.

Materialized paths require transactional subtree rewrites for moves. Concurrent
position allocation needs locking plus database uniqueness. These risks are
tracked for later implementation tasks and do not authorize implementation in
T02B.

## Non-goals

- No recursive-query optimization.
- No caching.
- No event sourcing.
- No audit history.
- No search indexing.
- No builder UI.
- No permissions beyond Course ownership.

## Glossary

- **Aggregate Root**: the Course boundary through which hierarchy ownership and
  authorization resolve.
- **CourseNode**: the single typed persistence record used for Course hierarchy
  entries.
- **Curriculum Profile**: a set of semantic placement rules applied to the
  generic hierarchy.
- **Academic Profile**: the default profile requiring TOPIC nodes to resolve to
  a nearest SUBJECT ancestor.
- **Logical Path**: the ordered sequence of immutable node UUIDs from a
  hierarchy root to a node.
- **Encoded Path**: the private stored representation of a logical path.
- **Diagnostic DTO**: a dedicated read-only internal projection that may expose
  path information without exposing it from the persistence model.

This glossary is authoritative for Course System terminology unless a
project-wide glossary or superseding ADR explicitly replaces a definition.

## Governance

ADR identifiers follow `docs/architecture/ADR/README.md`: repository-wide
sequential numbers are permanent and never reused. ADR lifecycle states are
Proposed, Accepted, Superseded, and optionally Deprecated.

After acceptance, future incompatible hierarchy decisions MUST use a new ADR
that explicitly supersedes ADR-023. Accepted ADR-023 Version 1 is not edited in
place except for logged nonsemantic corrections.

## Evidence

Canonical technical-verification and acceptance evidence:
`docs/features/course-system/evidence/course-p1-t02b/`.
