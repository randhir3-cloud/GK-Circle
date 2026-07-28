# COURSE-P1-T02 Canonical Evidence

This directory is the authoritative evidence store for
`COURSE-P1-T02 - Course model and additive migration`.

## Outcome

- Course root model: implemented in `api/models/course.go`.
- Persistence: existing `goqu` model pattern, with prepared statements.
- Migration: additive sql-migrate up/down pair for `courses`.
- Focused and full backend verification: passed.
- Isolated local migration apply/rollback/reapply: passed.
- Frontend lint: passed from the repository-supported `app/` package.
- Production and Intel NUC access: not performed.

## Evidence index

- [Migration verification](migration.md)
- [Human-readable verification](verification.md)
- [Machine-readable verification](verification.json)
- [Sanitized command record](commands/course-p1-t02.log)
- [Implementation hashes](hashes/implementation.sha256)

The repository-wide historical migration rollback remains affected by an older
invalid down migration. This does not originate in T02; the isolated T02 pair
successfully applies, rolls back, and reapplies through the repository's
sql-migrate runner.
