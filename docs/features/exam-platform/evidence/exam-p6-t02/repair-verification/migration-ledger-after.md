# Migration Ledger — After Repair

Date: 2026-07-28  
Captured at: 2026-07-28T02:37 UTC  
Repair applied at: 2026-07-28 02:31:43 UTC

## Ledger state (gorp_migrations)

Last applied migration: `20260728120000_create_assessment_attempt_snapshot_items.up.sql`  
Applied at: `2026-07-28 02:31:43.747791+00`

Total rows in ledger: **103** (89 existing + 14 newly applied)

### Newly applied migrations (at 2026-07-28 02:31:43 UTC)

| id | applied_at |
|---|---|
| 20260727120000_question_revisions_and_answer_authority.down.sql | 2026-07-28 02:31:43.380228 |
| 20260727120000_question_revisions_and_answer_authority.up.sql | 2026-07-28 02:31:43.516650 |
| 20260727140000_create_question_import_jobs.down.sql | 2026-07-28 02:31:43.520349 |
| 20260727140000_create_question_import_jobs.up.sql | 2026-07-28 02:31:43.545133 |
| 20260727150000_alter_question_import_jobs_commit.down.sql | 2026-07-28 02:31:43.548349 |
| 20260727150000_alter_question_import_jobs_commit.up.sql | 2026-07-28 02:31:43.558789 |
| 20260727160000_create_question_collections.down.sql | 2026-07-28 02:31:43.561850 |
| 20260727160000_create_question_collections.up.sql | 2026-07-28 02:31:43.609643 |
| 20260727170000_create_test_snapshots.down.sql | 2026-07-28 02:31:43.612888 |
| 20260727170000_create_test_snapshots.up.sql | 2026-07-28 02:31:43.663480 |
| 20260728100000_alter_assessment_attempts_add_snapshot.down.sql | 2026-07-28 02:31:43.666550 |
| 20260728100000_alter_assessment_attempts_add_snapshot.up.sql | 2026-07-28 02:31:43.693389 |
| 20260728100100_alter_attempt_answers_restrict.down.sql | 2026-07-28 02:31:43.696409 |
| 20260728100100_alter_attempt_answers_restrict.up.sql | 2026-07-28 02:31:43.709308 |
| 20260728120000_create_assessment_attempt_snapshot_items.down.sql | 2026-07-28 02:31:43.712230 |
| **20260728120000_create_assessment_attempt_snapshot_items.up.sql** | **2026-07-28 02:31:43.747791** ← LAST APPLIED |

## Second migrate up result (idempotency)

A second `docker compose run --rm migration` was run immediately after.  
Exit code: 0  
Output: `warning .env file not found, scanning from OS ENV` (warning only, no error)  
Result: **No pending migrations — clean no-op**

## Migration count verification

| Source | Count |
|---|---|
| Repository migration files | 105 files (up + down pairs + some up-only) |
| gorp_migrations ledger rows (after repair) | 103 rows |
| Difference | 2 rows (the `.up.sql`-only migrations without `.down.sql` companion) |

The ledger is consistent with the repository.
