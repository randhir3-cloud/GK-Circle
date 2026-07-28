# COURSE-P1-T04 Canonical Evidence

Task: `COURSE-P1-T04 — CourseNode hierarchy repository queries`

Status: VERIFIED

- Ledger Version: 1
- Schema Version: 1
- Governing decision: `docs/architecture/ADR/ADR-023-canonical-course-hierarchy-model.md`
- Database Migration: NONE
- Production touched: No
- NUC touched: No

This directory is the authoritative evidence store for T04.

- `technical-verification.md` and `.json` record checks and environments.
- `repository-tests.md` records hierarchy-query coverage.
- `changed-files.md` records the scoped implementation inventory.
- `commands/` contains sanitized command results.
- `hashes/` records the final ledger control hashes.

The temporary local PostgreSQL database used to validate the recursive CTE was
deleted after verification. No repository migration was created or applied.
