# COURSE-P1-T03 Migration Verification

## Environment

- Local PostgreSQL Compose service only.
- Temporary database: `course_p1_t03_verify_20260725`.
- Temporary overlay: `.tmp-course-p1-t03-migrations`.
- Existing schema was established from the repository's pre-T03 migration set,
  with only the authorized T02/T03 migration overlay.

## Procedure and results

1. The temporary database was created at
   `2026-07-25T08:22:14.0216010Z` (1,550 ms, exit 0).
2. The pre-T03 Course schema was applied, two owners and two Courses were
   seeded, and the Course-row checksum was recorded as
   `b715243cc0f5d3a63a5accd21c9177ae`.
3. T03 was applied through the repository migration runner at
   `2026-07-25T08:35:34.6389941Z` (2,562 ms, exit 0).
4. Inspection confirmed the exact ten columns, defaults, ten named
   constraints, two partial sibling-position indexes, and the
   `(course_id, path text_pattern_ops)` prefix index.
5. Valid top-level and child rows were inserted. Database checks rejected:
   cross-Course parents, self-parenting, unsupported node types/statuses,
   blank titles, negative positions, duplicate top-level positions, duplicate
   child positions, duplicate Course-scoped paths, and unsupported statuses.
6. Constraint execution completed at
   `2026-07-25T08:36:20.2622761Z` (1,272 ms, exit 0).
7. Only the T03 down SQL was applied at
   `2026-07-25T08:36:54.4567090Z` (643 ms, exit 0). `course_nodes` and its
   dependent objects were absent afterward; Courses remained.
8. T03 was reapplied through the repository runner at
   `2026-07-25T08:40:42.5039887Z` (3,124 ms, exit 0).
9. The Course checksum remained
   `b715243cc0f5d3a63a5accd21c9177ae` after every stage.
10. The exact temporary database and overlay were removed after evidence was
    preserved.

The repository runner records split `.up.sql` and `.down.sql` files
individually. Its normal rollback command can therefore continue into older
down migrations, so the isolated proof applied only the T03 down SQL and
removed only its temporary migration metadata before reapply. This avoids
touching unrelated schema while still proving the repository-authored down
migration.

The uniqueness definition permits identical encoded strings in different
Courses. An actual duplicate complete logical path across Courses is not
applicable to valid generated paths because the final segment is a globally
unique node primary key; Course scoping is established by the
`UNIQUE(course_id, path)` definition.

Database migration status: created and verified locally; not run in production.
