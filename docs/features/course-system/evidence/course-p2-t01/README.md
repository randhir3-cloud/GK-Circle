# COURSE-P2-T01 Canonical Evidence

Task: `COURSE-P2-T01 — LearningItem database schema & migration`

Status: VERIFIED

- Ledger Version: 1
- Schema Version: 1
- Verified on: 2026-07-26
- Branch: `chore/ci-verification`
- Commit at verification: `eeac599f05eaf936c7f61db4a3deeac3c9063f59`
- Parallel-phase decision: `docs/development/modules/course-system/DECISIONS.md` D-003
- Governing hierarchy decision: `docs/architecture/ADR/ADR-023-canonical-course-hierarchy-model.md`
- LearningItem repository: Yes (paired with COURSE-P2-T02)
- HTTP APIs: No
- Course hierarchy modified: No
- Frontend implemented: No
- Database Migration: ONE (`20260726083000_create_learning_items_table`)
- Production touched: No
- NUC touched: No
- Breaking Changes: NO

This directory is the authoritative T01 evidence store:

- `technical-verification.md` / `technical-verification.json`: command index
- `migration-verification.md`: isolated PostgreSQL apply/inspect/rollback/reapply
- `changed-files.md`: authorized change inventory
- `commands/course-p2-t01.log`: sanitized command record
- `hashes/`: ledger sync/check integrity records

The temporary database was deleted after verification. Existing unrelated
tracked and untracked changes were preserved.
