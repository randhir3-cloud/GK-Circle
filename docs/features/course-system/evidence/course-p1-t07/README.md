# COURSE-P1-T07 Canonical Evidence

Task: `COURSE-P1-T07 — Transactional CourseNode subtree deletion`

Status: VERIFIED

- Ledger Version: 1
- Schema Version: 1
- Governing decision: `docs/architecture/ADR/ADR-023-canonical-course-hierarchy-model.md`
- Subtree deletion implemented: Yes
- Archive implemented: No
- Restore implemented: No
- Database Migration: NONE
- Production touched: No
- NUC touched: No
- Breaking Changes: NO

This directory is the authoritative evidence store for T07 transactional subtree
deletion verification.

- `technical-verification.md` and `.json` record command outcomes.
- `delete-tests.md` records delete semantics and regression coverage.
- `postgresql-verification.md` records isolated database verification.
- `changed-files.md` records scoped changes.
- `commands/` contains sanitized command results.
- `hashes/` records final ledger integrity hashes.

The temporary PostgreSQL verification database was deleted after evidence was
captured. Archive, restore, APIs, authorization, and UI remain outside T07 scope.
