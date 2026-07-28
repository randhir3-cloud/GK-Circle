# COURSE-P1-T10 Canonical Evidence

Task: `COURSE-P1-T10 — Admin CourseNode Hierarchy Mutation HTTP APIs`

Status: VERIFIED

- Ledger Version: 1
- Schema Version: 1
- Governing decision: `docs/architecture/ADR/ADR-023-canonical-course-hierarchy-model.md`
- Hierarchy mutation APIs implemented: Yes
- Course APIs modified: No
- CourseNode read APIs modified: No
- Archive implemented: No
- Frontend implemented: No
- Database Migration: NONE
- Production touched: No
- NUC touched: No
- Breaking Changes: NO

## Known limitation

T10 reuses the existing quiz-admin allowlist as the repository's current
general administrative gate. T10 does not introduce a new role or permission
model. Repository MoveNode / ReorderChildren / DeleteSubtree remain the
authoritative hierarchy mutation implementations.

This directory is the authoritative evidence store for T10 admin CourseNode
hierarchy mutation HTTP API verification.
