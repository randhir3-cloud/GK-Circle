# COURSE-P1-T08 Canonical Evidence

Task: `COURSE-P1-T08 — Admin Course HTTP APIs`

Status: VERIFIED

- Ledger Version: 1
- Schema Version: 1
- Governing decision: `docs/architecture/ADR/ADR-023-canonical-course-hierarchy-model.md`
- Admin Course APIs implemented: Yes
- CourseNode APIs implemented: No
- Hierarchy mutation APIs implemented: No
- Archive/restore implemented: No
- Frontend implemented: No
- Database Migration: NONE
- Production touched: No
- NUC touched: No
- Breaking Changes: NO

## Known limitation

T08 reuses the existing quiz-admin allowlist as the repository's current
general administrative gate. T08 does not introduce a new role or permission
model.

This directory is the authoritative evidence store for T08 admin Course HTTP API
verification.
