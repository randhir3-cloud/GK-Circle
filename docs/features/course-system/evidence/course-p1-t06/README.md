# COURSE-P1-T06 Canonical Evidence

Task: `COURSE-P1-T06 — Transactional CourseNode sibling reorder`

Status: VERIFIED

- Ledger Version: 1
- Schema Version: 1
- Governing decision: `docs/architecture/ADR/ADR-023-canonical-course-hierarchy-model.md`
- Database Migration: NONE
- Production touched: No
- NUC touched: No

This directory is the authoritative evidence store for T06 transactional sibling
reorder verification.

- `technical-verification.md` and `.json` record command outcomes.
- `mutation-tests.md` records reorder semantics and regression coverage.
- `changed-files.md` records scoped changes.
- `commands/` contains sanitized command results.
- `hashes/` records final ledger integrity hashes.

The temporary PostgreSQL verification database was deleted after evidence was
captured. Deletion, APIs, authorization, and UI remain outside T06 scope.
