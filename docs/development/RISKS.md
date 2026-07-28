# GK Circle Risk Register

This document tracks active technical and product risks, status, mitigations, and owners.

## Active Risk Register

| Risk | Status | Impact / Probability | Mitigation | Owner |
|---|---|---|---|---|
| Course enrollment authority gap (historical T12) | Resolved | — / — | Closed by D-006 + COURSE-P2-T12 VERIFIED (`course_enrollments` + learner GET gate). | Course System (Phase 2) |
| Materialized path update performance on large subtrees | Monitoring | Medium / Low | Update descendants via a single recursive update query in a database transaction; index the `path` column. | Course System (Phase 1) |
| Concurrent reorders producing duplicate positions | Monitoring | Medium / Low | Enforce transactional locking on sibling checks during insertions and moves. | Course System (Phase 1) |
| Dependency cycles in unlocking logic | Monitoring | High / Low | Run a cycle-detection check (directed acyclic graph validation) before saving new prerequisite paths. | Course System (Phase 5) |
| CourseNode hierarchy conflicts with CourseSubject/CourseTopic persistence wording | Resolved | High / Medium | Accepted ADR-023 selects typed CourseNode persistence and retains CourseSubject/CourseTopic as domain/API projections; T03 remains separately gated. | Architecture |
| `course-rules.md` contains stale Prisma and internally conflicting marketplace terminology | Open | Medium / High | Reconcile the standard separately; follow AGENTS.md, CLAUDE.MD, and the actual Go/goqu/sql-migrate repository meanwhile. | Documentation governance |
| SSR API request timeout inside Docker | Resolved | High / High | Configured an Nginx reverse proxy listening internally on port 3010 to route requests directly within the Docker network. | CI Infrastructure (Phase 0) |
