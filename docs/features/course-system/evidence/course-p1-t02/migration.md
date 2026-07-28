# Course Migration Verification

## Migration

- Up: `api/database/migrations/20260725110500_create_courses_table.up.sql`
- Down: `api/database/migrations/20260725110500_create_courses_table.down.sql`
- Runner: repository `gk-circle migrate up|down` command using sql-migrate
- Database: temporary local PostgreSQL 15.2 database in Docker
- Production/NUC: not accessed

The migration creates only the `courses` table and
`courses_owner_id_idx`. It uses the existing `users.id` ownership type and
foreign-key convention. The down migration drops only `courses`.

## Verified schema

| Column | PostgreSQL type | Nullable |
|---|---|---|
| `id` | `uuid` | No |
| `owner_id` | `bpchar(20)` | No |
| `title` | `text` | No |
| `short_description` | `text` | Yes |
| `language` | `text` | Yes |
| `difficulty` | `text` | Yes |
| `visibility` | `text` | Yes |
| `status` | `varchar(20)` | No |
| `created_at` | `timestamp` | No |
| `updated_at` | `timestamp` | No |

Verified constraints:

- `courses_pkey`
- `courses_owner_id_fkey`
- `courses_title_not_blank`
- `courses_status_check`

## Apply, rollback, and reapply

The full clean migration chain applied successfully to
`course_p1_t02_verify`. Repository-wide rollback then removed the T02 table and
both T02 migration records, but continued into historical migrations and
encountered the pre-existing invalid SQL:

`20240513190411_rename_column_score_to_points.down.sql`

That file uses `RENAME COLUMN IF EXISTS`, which PostgreSQL rejects. The
application also incorrectly reports process exit code 0 after printing the
error. A subsequent full-chain reapply was blocked by the partially reverted
historical state. These results are classified `FAILED_BASELINE`, supported by
the direct comparison that T02's table was absent and its migration-record
count was zero after rollback.

To verify T02 independently, a second temporary database,
`course_p1_t02_isolated_verify`, was created with only the existing `users`
table prerequisite. The repository runner was pointed at an isolated directory
containing only the T02 up/down files:

1. Apply: passed; `courses` present and two split-file migration records present.
2. Rollback: passed; `courses` absent and zero T02 migration records present.
3. Reapply: passed; `courses` and all expected constraints present.

Both temporary databases and the temporary migration directory were deleted
after preserving this evidence.
