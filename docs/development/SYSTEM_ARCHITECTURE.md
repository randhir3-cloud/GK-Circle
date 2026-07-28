# GK Circle System Architecture

This document outlines the system architecture of the GK Circle Educational Operating System.

## Architecture Overview

| Area | Technology | Location |
|---|---|---|
| Web application | Nuxt 3, Vue 3, TypeScript/JavaScript, Pinia | `app/` |
| API | Go 1.25, Fiber v2 | `api/` |
| Database | PostgreSQL 15, SQL migrations | `api/database/migrations/` |
| Authentication | Ory Kratos | `api/pkg/kratos/` |
| Cache & Coordination | Redis-compatible service | Compose service `redis` |
| Live interaction | WebSockets | Go API and Nuxt client |
| Object storage | MinIO | Compose service `minio` |
| Local email | Mailpit | Compose service `mailpit` |
| Runtime | Docker Compose | repository root |

## Module Boundaries

GK Circle is composed of decoupled functional modules under `docs/development/modules/`.
Each module maintains its own:
* `MASTER_PLAN.md`
* `CURRENT_STATUS.md`
* `HANDOFF.md`
* `phases/` checklist

The active module is the **Universal Chained Course System**.

## Course hierarchy

ADR-023 defines the canonical Course hierarchy. Course is the aggregate,
ownership, and authorization root. Typed CourseNode rows are the single
hierarchy persistence backbone; CourseSubject and CourseTopic remain domain
vocabulary and API projections.

`CourseNode.parent_id` is the authoritative structural edge. Nested
`children[]` in API responses are computed projections (module decision
**D-005**). Unlimited logical nesting is a product invariant; there is no
product max-depth constant and no persisted `depth` column today (**D-004**).
Identifier-based UUID materialized `path` supports subtree operations; slug
paths are non-normative illustrations only.

### Implemented today (repository evidence; not aspirational)

* Course root persistence (`courses` migration + Go model).
* Recursive CourseNode persistence (`course_nodes` migration + Go model).
* Course-scoped parent relationships (`parent_id`, same-Course FK).
* Root and child ordering via `position`.
* Hierarchy reads (roots, direct children, full tree with computed `children[]`).
* Transactional branch movement, sibling reordering, and subtree deletion.
* Admin Course APIs under `/api/v1/admin/courses`.
* Admin CourseNode APIs under `/api/v1/admin/courses/:course_id/nodes`
  (create/read, move, reorder, delete).
* LearningItem admin CRUD and authenticated learner published reads (Phase 2).

### Explicit gaps (do not claim as existing)

* Dedicated subtree read API.
* Ancestor / breadcrumb APIs.
* Publish-safe learner outline / tree API.
* Lazy/paginated large-branch builder projections.
* Course admin UI / recursive tree builder.
* Cross-Course copy subtree.
* Enrollment enforcement and runtime visibility evaluation.

Canonical decision:
`docs/architecture/ADR/ADR-023-canonical-course-hierarchy-model.md`.

Module architecture (relationship contract, invariants, API labels):
`docs/development/modules/course-system/architecture/current.md`.

Module decisions:
`docs/development/modules/course-system/DECISIONS.md` (**D-001** through **D-005**).
