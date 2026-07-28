# COURSE-P1-T09 Canonical Evidence

Task: `COURSE-P1-T09 — Admin CourseNode HTTP APIs`

Status: VERIFIED

- Ledger Version: 1
- Schema Version: 1
- Governing decision: `docs/architecture/ADR/ADR-023-canonical-course-hierarchy-model.md`
- CourseNode Admin APIs implemented: Yes
- Hierarchy mutation APIs implemented: No
- Archive/restore implemented: No
- Frontend implemented: No
- Database Migration: NONE
- Production touched: No
- NUC touched: No
- Breaking Changes: NO

## Known limitation

T09 reuses the existing quiz-admin allowlist as the repository's current
general administrative gate. T09 does not introduce a new role or permission
model. CourseNode has no owner field; create bodies may include unknown
`owner_id`/`course_id` keys which are ignored. Course scope comes only from
the route.

This directory is the authoritative evidence store for T09 admin CourseNode
create/read HTTP API verification.
