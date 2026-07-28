# COURSE-P2-T01 Migration Verification

## Environment

- Local PostgreSQL Compose service `gk-circle-db` only.
- Temporary database: `course_p2_t01_verify_20260726`.
- Precondition schema: minimal `users` stub + repository `courses` and
  `course_nodes` up migrations (required FK chain).
- Production and the primary `gk_circle` database were not migrated.

## Procedure and results

1. Created temporary database `course_p2_t01_verify_20260726`.
2. Applied stub `users`, then `20260725110500_create_courses_table.up.sql` and
   `20260725134707_create_course_nodes_table.up.sql`.
3. Seeded one Course and one CourseNode. Course checksum:
   `0892314c35e08ca9f709102a7b813910`. CourseNode count: `1`.
4. Applied `20260726083000_create_learning_items_table.up.sql`.
5. Inspection confirmed columns, defaults, composite FK
   `(course_id, course_node_id) → course_nodes(course_id, id) ON DELETE RESTRICT`,
   unique `(course_node_id, position)`, non-negative position, nonblank title,
   typed `item_type`, and index `(course_node_id, position)`.
6. Constraint proofs rejected duplicate position, blank title, invalid type,
   negative position, and cross-course FK attachment.
7. Applied only the T01 down SQL; `learning_items` was absent afterward.
   Course checksum and CourseNode count were unchanged.
8. Reapplied the T01 up SQL successfully. Course checksum remained
   `0892314c35e08ca9f709102a7b813910`.
9. Dropped the temporary database and removed temporary SQL overlays.

Database migration status: created and verified locally; not run in production.
