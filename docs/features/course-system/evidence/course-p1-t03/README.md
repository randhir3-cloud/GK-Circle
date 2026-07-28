# COURSE-P1-T03 Canonical Evidence

Task: `COURSE-P1-T03 — CourseNode model and additive migration`

Status: VERIFIED

- Ledger Version: 1
- Schema Version: 1
- Verified on: 2026-07-25
- Branch: `chore/ci-verification`
- Commit at verification: `eeac599f05eaf936c7f61db4a3deeac3c9063f59`
- Governing decision: `docs/architecture/ADR/ADR-023-canonical-course-hierarchy-model.md`
- Production touched: No
- NUC touched: No
- Breaking Changes: NO
- Database migration status: CREATED; apply/rollback/reapply verified only in a temporary local PostgreSQL database; NOT RUN in production

This directory is the authoritative T03 evidence store:

- `technical-verification.md` and `technical-verification.json`: command and outcome index.
- `migration-verification.md`: isolated PostgreSQL schema and constraint proof.
- `model-test-results.md`: focused and repository-wide Go verification.
- `changed-files.md`: authorized change inventory and scope exclusions.
- `commands/course-p1-t03.log`: sanitized command record.
- `hashes/`: final ledger read-only and idempotency records.

The temporary database and migration overlay were deleted after verification.
Existing unrelated tracked and untracked changes were preserved.
