# GK Circle AI Context

## Project context

* **Product**: GK Circle
* **Codebase**: Nuxt application in `app/`, Go API in `api/`
* **Local stack**: Docker Compose with PostgreSQL, Redis, Kratos, MinIO, and Mailpit

## Allowed technology

* **Backend**: Go 1.25, Fiber v2, goqu
* **Frontend**: Nuxt 3, Vue 3, Pinia, Tailwind
* **Database**: PostgreSQL 15 with additive sql-migrate migrations
* **Forbidden without ADR**: NestJS, Next.js, Prisma, or a second authentication/test engine

## Active project state

* **Module**: Course System (`docs/development/modules/course-system/`)
* **Ledger Version**: 1
* **Schema Version**: 1
* **Active Phase**: COURSE-P2 — Learning Items and Information Blocks
* **In-progress Task**: None
* **Next Task**: None (Phase 2 complete)
* **Status**: IN_PROGRESS

`COURSE-P1-T01` through `COURSE-P1-T10`, `COURSE-P2-T01` through
`COURSE-P2-T17`, `COURSE-P2-T20`, `COURSE-P2-T21`, `COURSE-P2-T22`, `COURSE-P2-T23`, `COURSE-P2-T24`, `COURSE-P2-T25`, and `COURSE-P2-T26` are verified. COURSE-P2-T12 enrollment gate is VERIFIED (D-006). ADR-023 is the hierarchy
authority. Module clarifications **D-004** / **D-005** / **D-006** live in
`modules/course-system/DECISIONS.md` and
`modules/course-system/architecture/current.md`. Implementation roadmap
(DOC-CS-T02-R1 freeze): `modules/course-system/ROADMAP.md` + `phases/`.
Agent protocol: `modules/course-system/README.md`. Freeze certificate:
`modules/course-system/DOCUMENTATION_FREEZE.md` (DOC-CS-T04).
Parallel Phase 2 work remains authorized by D-003; `COURSE-P1-T11` available.
Ledger next: `COURSE-P2-T18`. COURSE-P2-T17 is VERIFIED with focused and full
Playwright coverage of the real deep-node create/publish/learner workflow,
responsive evidence, verified cleanup, and no production source modification.

**Mandatory agent rule:** Before implementing any Course hierarchy task, read
`architecture/current.md`, D-004, D-005, ADR-023, the relevant phase, and all
dependency tasks. Prefer canonical sources over copied architecture text.

## Architecture Freeze Marker

```text
Course System Architecture Freeze:
DOC-CS-T01

Roadmap Freeze:
DOC-CS-T02-R1

Documentation Normalization:
DOC-CS-T03

Documentation Certification:
DOC-CS-T04

Status:
Documentation Freeze Approved

Completed implementation:
COURSE-P2-T09 (VERIFIED)
COURSE-P2-T12 (VERIFIED)
COURSE-P2-T15 (VERIFIED)
COURSE-P2-T14 (VERIFIED)
COURSE-P2-T16 (VERIFIED)
COURSE-P2-T17 (VERIFIED)

Blocked (not complete):
—

Next ledger candidate:
COURSE-P2-T18

Documentation governance is complete. COURSE-P2-T12 and COURSE-P2-T15 are
VERIFIED after D-006 enrollment persistence. COURSE-P2-T14, COURSE-P2-T16,
and COURSE-P2-T17 are VERIFIED. Local course-admin seed:
`docs/development/local-course-admin.md`.
```

## Constraints

1. Additive migrations only; never edit an applied migration.
2. Do not access production or the NUC without explicit approval.
3. Task status promotion is manual and requires canonical evidence.
4. Keep business rules and authorization in the Go API.
5. Reuse the existing quiz, question, session, scoring, and report systems.
6. Treat ADR-023 as the hierarchy authority; Phase 1 UI (T11) and Phase 2
   admin item composition UI (`COURSE-P2-T09`) remain separately gated.
7. Architecture is frozen. AI must not invent: product `MAX_DEPTH` /
   `MAX_*_DEPTH`, authoritative `child_node_ids[]`, persisted depth as
   authority, fixed Subject→Chapter→Lecture schema levels, or a parallel
   hierarchy engine. Preserve `parent_id` authority. Keep the three graphs
   distinct (structure ≠ learning sequence ≠ prerequisite DAG).
   CourseNode.`status` ≠ LearningItem.`publish_state`.
8. Future Course System documentation changes require ADR and/or governance
   approval, canonical-document updates, and an append-only CHANGELOG entry.
